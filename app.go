package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ton-blockchain/adnl-tunnel/config"
	"github.com/ton-blockchain/adnl-tunnel/tunnel"
	"github.com/tonutils/torrent-client/core/api"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/tonutils/torrent-client/core/gostorage"
	"github.com/tonutils/torrent-client/core/upnp"
	"github.com/tonutils/torrent-client/oshook"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xssnick/ton-payment-network/tonpayments/chain"
	"github.com/xssnick/tonutils-go/adnl"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-storage/storage"
	"log"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx                   context.Context
	api                   *api.API
	daemonProcess         *os.Process
	rootPath              string
	config                *Config
	loaded                bool
	frontMounted          bool
	tunnelSettingsUpdated bool

	openFileData []byte
	openFileHash string

	lastCreateProgressReport time.Time
	creationCtx              context.Context
	cancelCreation           context.CancelFunc

	closerCtx context.Context
	closeCtx  context.CancelFunc

	tunnelCtx context.Context

	mx sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	oshook.HookStartup(a.openFile, a.openHash)
	adnl.Logger = func(v ...any) {}
	storage.Logger = log.Println
	storage.DownloadThreads = runtime.NumCPU() * 2
	storage.DownloadPrefetch = storage.DownloadThreads * 5

	tunnel.ChannelCapacityForNumPayments = 50
	tunnel.ChannelPacketsToPrepay = 20000

	var err error
	a.rootPath, err = PrepareRootPath()
	if err != nil {
		a.Throw(err)
	}

	cfg, err := LoadConfig(a.rootPath)
	if err != nil {
		a.Throw(err)
	}

	a.config = cfg

	return a
}

func (a *App) exit(ctx context.Context) {
	log.Println("Exiting...")

	a.closeCtx()
	if a.daemonProcess != nil {
		_ = a.daemonProcess.Kill()
	}

	if a.tunnelCtx != nil {
		log.Println("Stopping tunnel...")
		<-a.tunnelCtx.Done()
		log.Println("Tunnel stopped")
	}
	log.Println("Graceful exit completed")
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.closerCtx, a.closeCtx = context.WithCancel(a.ctx)
}

func (a *App) Throw(err error) {
	msg := err.Error()
	if len(msg) > 800 {
		msg = msg[:800]
	}
	runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
		Type:          runtime2.ErrorDialog,
		Title:         "Fatal error",
		Message:       msg,
		DefaultButton: "Exit",
	})
	a.exit(a.ctx)
	panic(err.Error())
}

func (a *App) ShowMsg(text string) {
	_, _ = runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
		Type:          runtime2.InfoDialog,
		Title:         "Info",
		Message:       text,
		DefaultButton: "OK",
	})
}

func (a *App) ShowWarnMsg(text string) {
	_, _ = runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
		Type:          runtime2.WarningDialog,
		Title:         "Warning",
		Message:       text,
		DefaultButton: "OK",
	})
}

func (a *App) prepare() {
	oshook.HookStartup(a.openFile, a.openHash)

	if (!a.config.PortsChecked && !a.config.SeedMode) || a.config.FetchIPOnStartup {
		log.Println("Trying to forward ports using UPnP")

		up, err := upnp.NewUPnP()
		if err != nil {
			log.Println("UPnP init failed", err.Error())
		} else {
			if err = up.ForwardPortTCP(18889); err != nil {
				log.Println("Port 18889 TCP forwarding failed:", err.Error())
			}
			if err = up.ForwardPortUDP(13333); err != nil {
				log.Println("Port 13333 UDP forwarding failed:", err.Error())
			}
		}

		ip, seed := CheckCanSeed()
		if seed {
			a.config.SeedMode = true
			a.config.ListenAddr = ip + ":13333"
			log.Println("Static seed mode is enabled, ports are open.")
		} else {
			log.Println("Static seed mode was not activated, ports are closed.")
		}
		a.config.PortsChecked = true
		_ = a.config.SaveConfig(a.rootPath)
	}
}

var oncePrepare sync.Once

type SectionInfo struct {
	Name  string
	Outer bool
}

func (a *App) DummySec() []SectionInfo {
	return []SectionInfo{}
}

func (a *App) ready(ctx context.Context) {
	oncePrepare.Do(func() {
		a.prepare()

		go func() {
			var err error
			var cl api.StorageClient
			if a.config.UseDaemon {
				cl, err = client.ConnectToStorageDaemon(a.config.DaemonControlAddr, a.config.DaemonDBPath)
				if err != nil {
					a.ShowWarnMsg("Failed to connect to storage daemon, falling back to tonutils-storage.\n\n" + err.Error())
				}
			}

			if !a.config.UseDaemon || err != nil {
				addr := strings.Split(a.config.ListenAddr, ":")

				lAddr := "0.0.0.0"

				if len(addr) < 2 {
					a.Throw(fmt.Errorf("ListenAddr in config.json is not valid"))
					return
				}

				cfg := gostorage.Config{
					Key:               ed25519.NewKeyFromSeed(a.config.Key),
					ListenAddr:        lAddr + ":" + addr[1],
					ExternalIP:        addr[0],
					DownloadsPath:     a.config.DownloadsPath,
					NetworkConfigPath: a.config.NetworkConfigPath,
				}
				if !a.config.SeedMode {
					cfg.ExternalIP = ""
				}

				if cfg.ExternalIP == "0.0.0.0" {
					a.ShowWarnMsg("external ip cannot be 0.0.0.0, disabling seed mode, " +
						"change ip in settings to your real external ip")
					cfg.ExternalIP = ""
				}

				var stopTunnel context.CancelFunc
				if a.config.TunnelConfig != nil && a.config.TunnelConfig.NodesPoolConfigPath != "" {
					a.tunnelCtx, stopTunnel = context.WithCancel(context.Background())
				}

				tunCfg := a.config.TunnelConfig

			retry:
				var err error
				cl, err = gostorage.NewClient(a.closerCtx, a.rootPath+"/tonutils-storage-db", cfg, tunCfg, func(addr string) {
					go func() {
						// wait till frontend init, to display event
						for !a.loaded {
							time.Sleep(50 * time.Millisecond)
						}
						runtime2.EventsEmit(a.ctx, "tunnel_assigned", addr)
						log.Println("TUNNEL ASSIGNED:", addr)
					}()
				}, func() {
					stopTunnel()
				}, func(to, from []*tunnel.SectionInfo) bool {
					if tunCfg == nil {
						// skip all routes
						return false
					}

					var priceIn, priceOut = big.NewInt(0), big.NewInt(0)
					var sect []SectionInfo
					for i, n := range append(to, from...) {
						sect = append(sect, SectionInfo{
							Name:  base64.StdEncoding.EncodeToString(n.Keys.ReceiverPubKey)[:8],
							Outer: i == len(to)-1,
						})

						if n.PaymentInfo != nil {
							if n.PaymentInfo.ExtraCurrencyID != 0 || n.PaymentInfo.JettonMaster != nil {
								a.ShowWarnMsg("Route has node with payment in currency other than TON, it is not yet supported in Torrent, rerouting")
								return false
							}

							// consider 1 packet = 512 bytes, actually more, but this is avg payload
							var packetsPerMB int64 = 2048

							amt := new(big.Int).SetUint64(n.PaymentInfo.PricePerPacket)
							amt.Mul(amt, big.NewInt(packetsPerMB))

							vcFee := big.NewInt(0)
							for _, section := range n.PaymentInfo.PaymentTunnel {
								vcFee.Add(vcFee, section.MinFee)
							}

							packetsPerChannel := tunnel.ChannelCapacityForNumPayments * tunnel.ChannelPacketsToPrepay
							// channel fee per 1 mb
							feeDiv := new(big.Float).Quo(new(big.Float).SetInt64(packetsPerMB), new(big.Float).SetInt64(packetsPerChannel))

							feePer1MB, _ := feeDiv.Mul(new(big.Float).SetInt(vcFee), feeDiv).Int(vcFee)
							amt.Add(amt, feePer1MB)

							if i < len(to)-1 {
								priceOut.Add(priceOut, amt)
							} else if i == len(to)-1 {
								priceOut.Add(priceOut, amt)
								priceIn.Add(priceOut, amt)
							} else {
								priceIn.Add(priceOut, amt)
							}
						}
					}

					for !a.frontMounted {
						time.Sleep(10 * time.Millisecond)
					}
					runtime2.EventsEmit(a.ctx, "tunnel_check", sect, tlb.FromNanoTON(priceIn).String(), tlb.FromNanoTON(priceOut).String())

					ch := make(chan bool, 1)
					runtime2.EventsOn(a.ctx, "tunnel_check_result", func(optionalData ...interface{}) {
						runtime2.EventsOff(a.ctx, "tunnel_check_result")
						if len(optionalData) == 0 {
							// cancel tunnel, start without it
							tunCfg = nil
							ch <- false
							return
						}

						ch <- optionalData[0].(bool)
					})
					return <-ch
				}, func() bool {
					if tunCfg == nil {
						return false
					}

					for !a.frontMounted {
						time.Sleep(10 * time.Millisecond)
					}
					runtime2.EventsEmit(a.ctx, "tunnel_reinit_ask")

					ch := make(chan bool, 1)
					runtime2.EventsOn(a.ctx, "tunnel_reinit_ask_result", func(optionalData ...interface{}) {
						runtime2.EventsOff(a.ctx, "tunnel_reinit_ask_result")
						ch <- optionalData[0].(bool)
					})
					return <-ch
				}, func(s string) {
					for !a.frontMounted {
						time.Sleep(10 * time.Millisecond)
					}

					runtime2.EventsEmit(a.ctx, "report_state", s)
				}, func(coins tlb.Coins) {
					runtime2.EventsEmit(a.ctx, "tunnel_paid_updated", coins.String())
				})
				if err != nil {
					if strings.HasPrefix(err.Error(), "tunnel preparation failed:") {
						if tunCfg != nil {
							a.ShowWarnMsg("Failed to prepare tunnel, will start without it\n\nError: " + err.Error())
							tunCfg = nil
						}
						goto retry
					}
					a.Throw(fmt.Errorf("failed to init go storage: %w", err))
					return
				}
			}

			// loading done, hook again to steal it from webview
			a.api = api.NewAPI(ctx, cl)
			a.api.SetOnListRefresh(func() {
				runtime2.EventsEmit(a.ctx, "update")
				runtime2.EventsEmit(a.ctx, "update_peers")
				runtime2.EventsEmit(a.ctx, "update_files")
				runtime2.EventsEmit(a.ctx, "update_info")
			})
			a.api.SetSpeedRefresh(func(speed api.Speed) {
				runtime2.EventsEmit(a.ctx, "speed", speed)
			})
			a.loaded = true

			runtime2.EventsOn(a.ctx, "refresh", func(optionalData ...interface{}) {
				_ = a.api.SyncTorrents()
			})

			for range cl.GetNotifier() {
				_ = a.api.SyncTorrents()

				// to not refresh too often
				time.Sleep(70 * time.Millisecond)
			}
		}()
	})
}

var onceCheck = sync.Once{}

func (a *App) WaitReady() {
	onceCheck.Do(func() {
		a.frontMounted = true
		go func() {
			// wait for daemon ready
			for !a.loaded {
				time.Sleep(50 * time.Millisecond)
			}

			runtime2.EventsEmit(a.ctx, "daemon_ready")
			runtime2.OnFileDrop(a.ctx, func(x, y int, paths []string) {
				if len(paths) == 0 {
					return
				}
				if !strings.HasSuffix(paths[0], ".tonbag") {
					return
				}
				a.openFilePath(paths[0])
			})

			if a.openFileData != nil {
				a.openFile(a.openFileData)
				a.openFileData = nil
			} else if a.openFileHash != "" {
				a.openHash(a.openFileHash)
				a.openFileHash = ""
			}
		}()
	})
}

func (a *App) SetSpeedLimit(down, up int64) string {
	err := a.api.SetSpeedLimits(&api.SpeedLimits{
		Download: down,
		Upload:   up,
	})
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return ""
}

func (a *App) GetSpeedLimit() *api.SpeedLimits {
	limits, err := a.api.GetSpeedLimits()
	if err != nil {
		log.Println(err.Error())
		return &api.SpeedLimits{}
	}
	return limits
}

func (a *App) GetProviderContract(hash, owner string) api.ProviderContract {
	return a.api.GetProviderContract(hash, owner)
}

func (a *App) FetchProviderRates(hash, provider string) api.ProviderRates {
	return a.api.FetchProviderRates(hash, provider)
}

func (a *App) RequestProviderStorageInfo(hash, provider, owner string) api.ProviderStorageInfo {
	return a.api.RequestProviderStorageInfo(hash, provider, owner)
}

func (a *App) BuildProviderContractData(hash, ownerAddr, amount string, providers []api.NewProviderData) *api.Transaction {
	t, err := a.api.BuildProviderContractData(hash, ownerAddr, amount, providers)
	if err != nil {
		a.ShowWarnMsg(err.Error())
		return nil
	}
	return t
}

func (a *App) BuildWithdrawalContractData(hash, ownerAddr string) *api.Transaction {
	t, err := a.api.BuildWithdrawalContractData(hash, ownerAddr)
	if err != nil {
		a.ShowWarnMsg(err.Error())
		return nil
	}
	return t
}

func (a *App) GetPaymentNetworkWalletAddr() string {
	w, err := chain.InitWallet(ton.NewAPIClient(liteclient.NewOfflineClient()), ed25519.NewKeyFromSeed(a.config.TunnelConfig.Payments.WalletPrivateKey))
	if err != nil {
		log.Println(err.Error())
		return "{ERROR}"
	}
	return w.WalletAddress().String()
}

func (a *App) openFile(data []byte) {
	if a.loaded {
		res := a.addByMeta(data)
		if res.Err == "" {
			runtime2.EventsEmit(a.ctx, "open_torrent", res.Hash)
		} else {
			a.ShowMsg("Error while parsing meta file: " + res.Err + "")
		}
	} else {
		// wait for loading
		a.openFileData = data
	}
}

func (a *App) openFilePath(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
		return
	}
	a.openFile(data)
}

func (a *App) openHash(hash string) {
	if a.loaded {
		res := a.AddTorrentByHash(hash)
		if res == "" {
			runtime2.EventsEmit(a.ctx, "open_torrent", hash)
		} else {
			a.ShowMsg("Error while parsing hash '" + hash + "': " + res + "")
		}
	} else {
		// wait for loading
		a.openFileHash = hash
	}
}

func (a *App) OpenDir() string {
	str, err := runtime2.OpenDirectoryDialog(a.ctx, runtime2.OpenDialogOptions{})
	if err != nil {
		log.Println(err.Error())
	}
	return str
}

func (a *App) OpenFile() string {
	str, err := runtime2.OpenFileDialog(a.ctx, runtime2.OpenDialogOptions{})
	if err != nil {
		log.Println(err.Error())
	}
	return str
}

type TunnelConfigInfo struct {
	Max     int
	MaxFree int
	Path    string
}

func (a *App) OpenTunnelConfig() *TunnelConfigInfo {
	path, err := runtime2.OpenFileDialog(a.ctx, runtime2.OpenDialogOptions{
		DefaultDirectory: "",
		DefaultFilename:  "nodes-pool.json",
		Title:            "Open Nodes Pool Config",
		Filters: []runtime2.FileFilter{
			{
				DisplayName: "nodes-pool.json",
				Pattern:     "*.json",
			},
		},
		ShowHiddenFiles:            false,
		CanCreateDirectories:       false,
		ResolvesAliases:            false,
		TreatPackagesAsDirectories: false,
	})
	if err != nil {
		println(err.Error())
		return nil
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			_, _ = runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
				Type:          runtime2.ErrorDialog,
				Title:         "Failed to read tunnel config",
				Message:       err.Error(),
				DefaultButton: "Ok",
			})
			return nil
		}

		var sharedCfg config.SharedConfig
		if err = json.Unmarshal(data, &sharedCfg); err != nil {
			_, _ = runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
				Type:          runtime2.ErrorDialog,
				Title:         "Failed to parse tunnel config",
				Message:       err.Error(),
				DefaultButton: "Ok",
			})
			return nil
		}

		if len(sharedCfg.NodesPool) == 0 {
			_, _ = runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
				Type:    runtime2.ErrorDialog,
				Title:   "Failed to parse nodes pool config",
				Message: "Invalid nodes pool config format",
			})
			return nil
		}

		maxFree := 0
		for _, node := range sharedCfg.NodesPool {
			if node.Payment == nil {
				maxFree++
			}
		}

		return &TunnelConfigInfo{Path: path, Max: len(sharedCfg.NodesPool), MaxFree: maxFree}
	}

	return &TunnelConfigInfo{Path: ""}
}

type TorrentCreateResult struct {
	Hash string
	Err  string
}

func (a *App) CreateTorrent(dir, description string) TorrentCreateResult {
	a.creationCtx, a.cancelCreation = context.WithCancel(a.ctx)
	hash, err := a.api.CreateTorrent(a.creationCtx, dir, description, a.reportCreationProgress)
	if err != nil {
		log.Println(err.Error())
		return TorrentCreateResult{Err: err.Error()}
	}
	return TorrentCreateResult{Hash: hash}
}

func (a *App) CancelCreateTorrent() {
	log.Println("CANCEL CREATION")
	if a.cancelCreation != nil {
		a.cancelCreation()
	}
}

func (a *App) reportCreationProgress(done, max uint64) {
	now := time.Now()
	if a.lastCreateProgressReport.Add(50 * time.Millisecond).After(now) {
		// not refresh too often
		return
	}
	a.lastCreateProgressReport = now

	runtime2.EventsEmit(a.ctx, "update-create-progress", fmt.Sprintf("%.2f", (float64(done)/float64(max))*100))
}

func (a *App) ExportMeta(hash string) string {
	m, err := a.api.GetTorrentMeta(hash)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	info, err := a.api.GetInfo(hash)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	name := info.Description
	if name == "" {
		name = hash
	}

	path, err := runtime2.SaveFileDialog(a.ctx, runtime2.SaveDialogOptions{
		DefaultFilename: name + ".tonbag",
		Title:           "Save .tonbag",
		Filters: []runtime2.FileFilter{{
			DisplayName: "TON Torrent (*.tonbag)",
			Pattern:     "*.tonbag",
		}},
	})
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	err = os.WriteFile(path, m, 0666)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	return path
}

func (a *App) GetTorrents() []*api.Torrent {
	list := a.api.GetTorrents()
	if list == nil {
		list = []*api.Torrent{}
	}
	return list
}

func (a *App) GetFiles(hash string) []*api.File {
	list, err := a.api.GetTorrentFiles(hash)
	if err != nil {
		log.Println(err.Error())
	}
	if list == nil {
		return []*api.File{}
	}
	return list
}

func (a *App) GetInfo(hash string) *api.TorrentInfo {
	info, err := a.api.GetInfo(hash)
	if err != nil {
		log.Println(err.Error())
		return &api.TorrentInfo{}
	}
	return info
}

func (a *App) GetPlainFiles(hash string) []api.PlainFile {
	list, err := a.api.GetPlainFiles(hash)
	if err != nil {
		log.Println(err.Error())
	}
	if list == nil {
		return []api.PlainFile{}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].RawSize > list[j].RawSize
	})

	if len(list) > 1000 {
		list = list[:1000]
	}

	return list
}

func (a *App) GetPeers(hash string) []api.Peer {
	list, err := a.api.GetPeers(hash)
	if err != nil {
		log.Println(err.Error())
	}
	if list == nil {
		return []api.Peer{}
	}
	return list
}

func (a *App) StartDownload(hash string, files []string) {
	err := a.api.SetPriorities(hash, files, 1)
	if err != nil {
		log.Println(err.Error())
	}
}

func (a *App) AddTorrentByHash(hash string) string {
	err := a.api.AddTorrentByHash(hash, a.config.DownloadsPath+"/"+strings.ToUpper(hash))
	if err != nil {
		return err.Error()
	}

	return ""
}

type TorrentAddResult struct {
	Hash string
	Err  string
}

func (a *App) AddTorrentByMeta(meta string) TorrentAddResult {
	metaBytes, err := base64.StdEncoding.DecodeString(meta)
	if err != nil {
		return TorrentAddResult{Err: err.Error()}
	}
	return a.addByMeta(metaBytes)
}

func (a *App) addByMeta(meta []byte) TorrentAddResult {
	if len(meta) < 8 {
		return TorrentAddResult{Err: "too short meta"}
	}
	if binary.LittleEndian.Uint32(meta) == 0x6a7181e0 {
		// skip id, for compatibility with boxed and not boxed
		meta = meta[4:]
	}

	var ti client.MetaFile
	_, err := tl.Parse(&ti, meta, false)
	if err != nil {
		return TorrentAddResult{Err: err.Error()}
	}
	hash := hex.EncodeToString(ti.Hash)

	err = a.api.AddTorrentByMeta(meta, a.config.DownloadsPath+"/"+strings.ToUpper(hash))
	if err != nil {
		return TorrentAddResult{Err: err.Error()}
	}
	return TorrentAddResult{Hash: hash}
}

func (a *App) CheckHeader(hash string) bool {
	hasHeader, err := a.api.CheckTorrentHeader(hash)
	if err != nil {
		log.Println(hash, err.Error())
	}
	return hasHeader
}

func (a *App) WantRemoveTorrent(hashes []string) {
	runtime2.EventsEmit(a.ctx, "want_remove_torrent", hashes)
}

func (a *App) RemoveTorrent(hash string, withFiles, onlyNotInitiated bool) string {
	err := a.api.RemoveTorrent(hash, withFiles, onlyNotInitiated)
	if err != nil {
		log.Println(hash, err.Error())
		return err.Error()
	}
	return ""
}

func (a *App) IsDarkTheme() bool {
	return a.config.IsDarkTheme
}

func (a *App) SwitchTheme() {
	a.config.IsDarkTheme = !a.config.IsDarkTheme
	a.config.SaveConfig(a.rootPath)
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) SaveTunnelConfig(num uint, payments bool) string {
	a.config.TunnelConfig.TunnelSectionsNum = num
	a.config.TunnelConfig.PaymentsEnabled = payments

	err := a.config.SaveConfig(a.rootPath)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	a.tunnelSettingsUpdated = true

	return ""
}

func (a *App) SaveConfig(downloads string, useTonutilsStorage, seedMode bool, storageExtIP, daemonControlAddr, daemonDB, tunnelConfigPath string) string {
	notify := false
	a.config.DownloadsPath = downloads

	if useTonutilsStorage != !a.config.UseDaemon {
		a.config.UseDaemon = !useTonutilsStorage
		notify = true
	}

	if a.tunnelSettingsUpdated {
		notify = true
		a.tunnelSettingsUpdated = false
	}

	if tunnelConfigPath != a.config.TunnelConfig.NodesPoolConfigPath {
		a.config.TunnelConfig.NodesPoolConfigPath = tunnelConfigPath
		notify = true
	}

	if useTonutilsStorage {
		if a.config.SeedMode != seedMode {
			notify = true
			a.config.SeedMode = seedMode
		}

		if a.config.SeedMode && a.config.ListenAddr != storageExtIP {
			a.config.ListenAddr = storageExtIP
			notify = true
		}
	} else {
		if a.config.DaemonDBPath != daemonDB {
			a.config.DaemonDBPath = daemonDB
			notify = true
		}
		if a.config.DaemonControlAddr != daemonControlAddr {
			a.config.DaemonControlAddr = daemonControlAddr
			notify = true
		}
	}

	err := a.config.SaveConfig(a.rootPath)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}

	if notify {
		a.ShowMsg("Some settings will be changed on the next launch.\nPlease restart the app to apply them.")
	}

	return ""
}

func (a *App) OpenFolder(path string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		path = strings.ReplaceAll(path, "/", "\\")
		cmd = "explorer"
	}

	exec.Command(cmd, path).Start()
}

func (a *App) OpenFolderSelectFile(path string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", "-R", path).Start()
	case "linux":
		// TODO: select file somehow
		exec.Command("xdg-open", path).Start()
	case "windows":
		path = strings.ReplaceAll(path, "/", "\\")
		exec.Command("explorer", "/select,"+path).Start()
	}
}

func (a *App) SetActive(hash string, active bool) string {
	err := a.api.SetActive(hash, active)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return ""
}
