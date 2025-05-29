package gostorage

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	tunnelConfig "github.com/ton-blockchain/adnl-tunnel/config"
	"github.com/ton-blockchain/adnl-tunnel/tunnel"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/adnl"
	adnlAddress "github.com/xssnick/tonutils-go/adnl/address"
	"github.com/xssnick/tonutils-go/adnl/dht"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-storage-provider/pkg/transport"
	"github.com/xssnick/tonutils-storage/config"
	"github.com/xssnick/tonutils-storage/db"
	"github.com/xssnick/tonutils-storage/provider"
	"github.com/xssnick/tonutils-storage/storage"
	"net"
	"net/netip"
	"os"
	"runtime"
	"sort"
	"time"
)

type Config struct {
	Key           ed25519.PrivateKey
	ListenAddr    string
	ExternalIP    string
	DownloadsPath string

	NetworkConfigPath string
}

type Client struct {
	storage   *db.Storage
	srv       *storage.Server
	connector storage.NetConnector
	provider  *provider.Client

	notify chan bool
}

func NewClient(globalCtx context.Context, dbPath string, cfg Config, tunCfg *tunnelConfig.ClientConfig, onTunnel func(addr string), onStopped func(), tunAcceptor func(to, from []*tunnel.SectionInfo) int, reRouter func() bool, reportLoadingState func(string), onPaidUpdate func(coins tlb.Coins)) (*Client, error) {
	c := &Client{
		notify: make(chan bool, 1), // to refresh fast a bit after
	}

	closerCtx, closerCancel := context.WithCancel(globalCtx)

	tunStop := make(chan bool, 1)

	var success bool
	var destroyed bool
	var tunnelInitialized bool

	var toClose []func()
	destroy := func(final bool) {
		if !final && success {
			return
		}
		if destroyed {
			return
		}
		destroyed = true

		closerCancel()

		if tunnelInitialized {
			<-tunStop
		}

		for _, f := range toClose {
			f()
		}

		onStopped()
	}
	defer destroy(false)

	reportLoadingState("Checking config...")

	var ip net.IP
	var port uint16
	if cfg.ExternalIP != "" {
		ip = net.ParseIP(cfg.ExternalIP)
		if ip == nil {
			pterm.Error.Println("External ip is invalid")
			return nil, fmt.Errorf("invalid external ip")
		}
	}

	addr, err := netip.ParseAddrPort(cfg.ListenAddr)
	if err != nil {
		pterm.Error.Println("Listen addr is invalid")
		return nil, fmt.Errorf("invalid listen addr")
	}
	port = addr.Port()

	reportLoadingState("Fetching ton network config...")

	var lsCfg *liteclient.GlobalConfig
	if cfg.NetworkConfigPath != "" {
		lsCfg, err = liteclient.GetConfigFromFile(cfg.NetworkConfigPath)
		if err != nil {
			pterm.Error.Println("Failed to load ton network config from file:", err.Error())
			return nil, fmt.Errorf("failed to load ton network config from file: %w", err)
		}
	} else {
		lsCfg, err = liteclient.GetConfigFromUrl(closerCtx, "https://ton-blockchain.github.io/global.config.json")
		if err != nil {
			pterm.Warning.Println("Failed to download ton config:", err.Error(), "; We will take it from static cache")
			lsCfg = &liteclient.GlobalConfig{}
			if err = json.NewDecoder(bytes.NewBufferString(config.FallbackNetworkConfig)).Decode(lsCfg); err != nil {
				pterm.Error.Println("Failed to parse fallback ton config:", err.Error())
				return nil, fmt.Errorf("failed to parse fallback ton config: %w", err)
			}
		}
	}

	reportLoadingState("Initializing liteclient...")

	lsPool := liteclient.NewConnectionPool()
	apiClient := ton.NewAPIClient(lsPool, ton.ProofCheckPolicyFast).WithRetry(3)
	toClose = append(toClose, lsPool.Stop)

	// connect async to not slow down main processes
	go func() {
		for {
			if err := lsPool.AddConnectionsFromConfig(closerCtx, lsCfg); err != nil {
				pterm.Warning.Println("Failed to add connections from ton config:", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
	}()

	var gate *adnl.Gateway
	var netMgr adnl.NetManager
	if tunCfg != nil && tunCfg.NodesPoolConfigPath != "" {
		reportLoadingState("Preparing ADNL tunnel...")

		data, err := os.ReadFile(tunCfg.NodesPoolConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load tunnel nodes pool config: %w", err)
		}

		var tunNodesCfg tunnelConfig.SharedConfig
		if err = json.Unmarshal(data, &tunNodesCfg); err != nil {
			return nil, fmt.Errorf("failed to parse tunnel nodes pool config: %w", err)
		}

		tunnel.AskReroute = reRouter
		tunnel.Acceptor = tunAcceptor
		events := make(chan any, 1)
		go tunnel.RunTunnel(closerCtx, tunCfg, &tunNodesCfg, lsCfg, log.Logger, events)
		tunnelInitialized = true

		initUpd := make(chan any, 1)
		inited := false
		go func() {
			atm := &tunnel.AtomicSwitchableRegularTunnel{}
			for event := range events {
				switch e := event.(type) {
				case tunnel.StoppedEvent:
					close(tunStop)
					return
				case tunnel.MsgEvent:
					reportLoadingState(e.Msg)
				case tunnel.UpdatedEvent:
					log.Info().Msg("tunnel updated")

					e.Tunnel.SetOutAddressChangedHandler(func(addr *net.UDPAddr) {
						gate.SetAddressList([]*adnlAddress.UDP{
							{
								IP:   addr.IP,
								Port: int32(addr.Port),
							},
						})
						onTunnel(addr.String())
					})
					onTunnel(fmt.Sprintf("%s:%d", e.ExtIP.String(), e.ExtPort))

					go func() {
						for {
							select {
							case <-e.Tunnel.AliveCtx().Done():
								return
							case <-time.After(5 * time.Second):
								onPaidUpdate(e.Tunnel.CalcPaidAmount()["TON"])
							}
						}
					}()

					atm.SwitchTo(e.Tunnel)
					if !inited {
						inited = true
						netMgr = adnl.NewMultiNetReader(atm)
						gate = adnl.NewGatewayWithNetManager(cfg.Key, netMgr)

						select {
						case initUpd <- e:
						default:
						}
					} else {
						gate.SetAddressList([]*adnlAddress.UDP{
							{
								IP:   e.ExtIP,
								Port: int32(e.ExtPort),
							},
						})

						log.Info().Msg("connection switched to new tunnel")
					}
				case tunnel.ConfigurationErrorEvent:
					log.Err(e.Err).Msg("tunnel configuration error, will retry...")
				case error:
					select {
					case initUpd <- e:
					default:
					}
				}
			}
		}()

		switch x := (<-initUpd).(type) {
		case tunnel.UpdatedEvent:
			ip = x.ExtIP
			port = x.ExtPort

			pterm.Info.Println("Using tunnel - IP:", x.ExtIP.String(), " Port:", x.ExtPort)
		case error:
			return nil, fmt.Errorf("tunnel preparation failed: %w", x)
		}
	} else {
		reportLoadingState("Binding UDP port...")

		dl, err := adnl.DefaultListener(cfg.ListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create default listener: %w", err)
		}
		netMgr = adnl.NewMultiNetReader(dl)
		gate = adnl.NewGatewayWithNetManager(cfg.Key, netMgr)
	}
	toClose = append(toClose, func() {
		gate.Close()
		netMgr.Close()
	})

	reportLoadingState("Starting ADNL server...")

	listenThreads := runtime.NumCPU()
	if listenThreads > 32 {
		listenThreads = 32
	}

	serverMode := ip != nil
	if serverMode {
		gate.SetAddressList([]*adnlAddress.UDP{
			{
				IP:   ip.To4(),
				Port: int32(port),
			},
		})
		if err = gate.StartServer(cfg.ListenAddr, listenThreads); err != nil {
			return nil, fmt.Errorf("failed to start adnl gateway in server mode: %w", err)
		}
	} else {
		if err = gate.StartClient(listenThreads); err != nil {
			return nil, fmt.Errorf("failed to start adnl gateway in client mode: %w", err)
		}
	}

	_, dhtAdnlKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 dht adnl key: %w", err)
	}

	dhtGate := adnl.NewGatewayWithNetManager(dhtAdnlKey, netMgr)
	if err = dhtGate.StartClient(); err != nil {
		return nil, fmt.Errorf("failed to init dht adnl gateway: %w", err)
	}
	toClose = append(toClose, func() {
		dhtGate.Close()
	})

	dhtClient, err := dht.NewClientFromConfig(dhtGate, lsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init dht client: %w", err)
	}
	toClose = append(toClose, func() {
		dhtClient.Close()
	})

	providerGateSeed := sha256.Sum256(cfg.Key.Seed())
	gateKey := ed25519.NewKeyFromSeed(providerGateSeed[:])
	gateProvider := adnl.NewGatewayWithNetManager(gateKey, netMgr)
	if err = gateProvider.StartClient(); err != nil {
		return nil, fmt.Errorf("failed to start adnl gateway for provider: %w", err)
	}
	toClose = append(toClose, func() {
		gateProvider.Close()
	})

	reportLoadingState("Loading storage...")

	ldb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load db: %w", err)
	}
	toClose = append(toClose, func() {
		ldb.Close()
	})

	c.srv = storage.NewServer(dhtClient, gate, cfg.Key, serverMode, 8)
	toClose = append(toClose, func() {
		c.srv.Stop()
	})

	ch := make(chan db.Event, 1)
	c.connector = storage.NewConnector(c.srv)
	c.storage, err = db.NewStorage(ldb, c.connector, 0, false, false, false, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}
	toClose = append(toClose, func() {
		c.storage.Close()
	})
	c.srv.SetStorage(c.storage)

	prvClient := transport.NewClient(gateProvider, dhtClient)
	c.provider = provider.NewClient(c.storage, apiClient, prvClient)

	d, u, err := c.storage.GetSpeedLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load speed limits: %w", err)
	}

	c.connector.SetDownloadLimit(d)
	c.connector.SetUploadLimit(u)

	go func() {
		defer destroy(true)

		ticker := time.Tick(1000 * time.Millisecond)
		for {
			select {
			case <-closerCtx.Done():
				return
			case <-ch:
			case <-ticker:
			}

			select {
			case c.notify <- true:
			default:
			}
		}
	}()
	success = true

	reportLoadingState("Initialized")

	return c, nil
}

func (c *Client) GetTorrents(ctx context.Context) (*client.TorrentsList, error) {
	all := c.storage.GetAll()

	var list client.TorrentsList
	for _, t := range all {
		full, err := c.getTorrent(t.BagID, false)
		if err != nil {
			return nil, err
		}
		list.Torrents = append(list.Torrents, full.Torrent)
	}
	return &list, nil
}

func (c *Client) AddByHash(ctx context.Context, hash []byte, dir string) (*client.TorrentFull, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("invalid bag id: should be 32 bytes len")
	}

	tor := c.storage.GetTorrent(hash)
	if tor == nil {
		tor = storage.NewTorrent(dir, c.storage, c.connector)
		tor.BagID = hash
	}

	if err := tor.Start(true, false, false); err != nil {
		return nil, fmt.Errorf("download error: %w", err)
	}

	if err := c.storage.SetTorrent(tor); err != nil {
		return nil, fmt.Errorf("failed to add bag: %w", err)
	}
	return c.GetTorrentFull(ctx, tor.BagID)
}

func (c *Client) AddByMeta(ctx context.Context, meta []byte, dir string) (*client.TorrentFull, error) {
	if len(meta) < 8 {
		return nil, fmt.Errorf("too short meta")
	}
	if binary.LittleEndian.Uint32(meta) == 0x6a7181e0 {
		// skip id, for compatibility with boxed and not boxed
		meta = meta[4:]
	}

	var ti client.MetaFile
	_, err := tl.Parse(&ti, meta, false)
	if err != nil {
		return nil, err
	}

	tor := c.storage.GetTorrent(ti.Hash)
	if tor == nil {
		tor = storage.NewTorrent(dir, c.storage, c.connector)
		tor.BagID = ti.Hash
	}

	if tor.Info == nil {
		tor.Info = &ti.Info
	}
	if tor.Header == nil && ti.Header != nil {
		tor.Header = ti.Header
		tor.InitMask()
	}

	if err = tor.Start(true, false, false); err != nil {
		return nil, fmt.Errorf("download error: %w", err)
	}

	if err = c.storage.SetTorrent(tor); err != nil {
		return nil, fmt.Errorf("failed to add bag: %w", err)
	}
	return c.GetTorrentFull(ctx, tor.BagID)
}

func (c *Client) CreateTorrent(ctx context.Context, path, description string, progressCallback func(done uint64, max uint64)) (*client.TorrentFull, error) {
	rootPath, dir, files, err := c.storage.DetectFileRefs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read files: %w", err)
	}

	it, err := storage.CreateTorrent(ctx, rootPath, dir, description, c.storage, c.connector, files, progressCallback)
	if err != nil {
		return nil, fmt.Errorf("failed to create bag: %w", err)
	}
	it.Start(true, false, false)

	err = c.storage.SetTorrent(it)
	if err != nil {
		return nil, fmt.Errorf("failed to save bag: %w", err)
	}
	return c.GetTorrentFull(ctx, it.BagID)
}

func (c *Client) GetTorrentFull(ctx context.Context, hash []byte) (*client.TorrentFull, error) {
	return c.getTorrent(hash, true)
}

func (c *Client) GetUploadStats(ctx context.Context, hash []byte) (uint64, error) {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return 0, fmt.Errorf("torrent is not found")
	}
	return t.GetUploadStats(), nil
}

func (c *Client) getTorrent(hash []byte, withFiles bool) (*client.TorrentFull, error) {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}

	var files []client.FileInfo
	activeDownload, activeUpload := t.IsActive()
	verificationInProgress, _ := t.GetLastVerifiedAt()
	torrent := client.Torrent{
		Hash:           t.BagID,
		AddedAt:        uint32(t.CreatedAt.Unix()),
		RootDir:        t.Path,
		ActiveDownload: activeDownload,
		ActiveUpload:   activeUpload,
		Completed:      false,
		Verified:       !verificationInProgress,
		FatalError:     nil,
	}
	if t.Info != nil {
		torrent.Flags |= 1
		incSize := t.Info.FileSize
		torrent.TotalSize = &incSize
		torrent.Description = &t.Info.Description.Value
	}
	if t.Header != nil {
		torrent.Flags |= 2
		dir := string(t.Header.DirName)
		torrent.DirName = &dir
		filesCount := uint64(t.Header.FilesCount)
		torrent.FilesCount = &filesCount
		incSize := t.Info.FileSize - t.Info.HeaderSize
		torrent.IncludedSize = &incSize

		for _, p := range t.GetPeers() {
			torrent.DownloadSpeed += float64(p.GetDownloadSpeed())
			torrent.UploadSpeed += float64(p.GetUploadSpeed())
		}

		files = make([]client.FileInfo, t.Header.FilesCount)
		for i := uint32(0); i < t.Header.FilesCount; i++ {
			fi, err := t.GetFileOffsetsByID(i)
			if err != nil {
				return nil, fmt.Errorf("failed to get offset for file %d: %w", i, err)
			}
			files[i].Name = fi.Name
			files[i].Size = int64(fi.Size)
		}

		completed := true
		mask := t.PiecesMask()
		for _, u := range t.GetActiveFilesIDs() {
			fi, err := t.GetFileOffsetsByID(u)
			if err != nil {
				return nil, fmt.Errorf("failed to get offset for file %d: %w", u, err)
			}

			var sz uint64
			for y := fi.FromPiece; y <= fi.ToPiece; y++ {
				has := mask[y/8] & (1 << (y % 8))
				if has == 0 {
					// not all needed pieces downloaded
					completed = false
				} else {
					sz += uint64(t.Info.PieceSize)
				}
			}

			if sz > fi.Size {
				sz = fi.Size
			}
			torrent.DownloadedSize = sz

			if withFiles {
				files[u].Priority = 1
				files[u].DownloadedSize = int64(sz)
			}
		}
		torrent.Completed = completed
	}

	return &client.TorrentFull{
		Torrent: torrent,
		Files:   files,
	}, nil
}

func (c *Client) GetTorrentMeta(ctx context.Context, hash []byte) ([]byte, error) {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}

	if t.Info == nil || t.Header == nil {
		return nil, fmt.Errorf("torrent header is not initialized")
	}

	mf := client.MetaFile{
		Info:   *t.Info,
		Header: t.Header,
	}

	return tl.Serialize(mf, true)
}

func (c *Client) GetPeers(ctx context.Context, hash []byte) (*client.PeersList, error) {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}

	if t.Info == nil {
		return &client.PeersList{}, nil
	}

	var list client.PeersList
	list.TotalParts = int64(t.Info.PiecesNum())

	for s, p := range t.GetPeers() {
		adnlAddr, _ := hex.DecodeString(s)
		list.Peers = append(list.Peers, client.Peer{
			ADNL: adnlAddr,
			IP:   p.Addr,
			DownloadSpeed: client.Double{
				Value: float64(p.GetDownloadSpeed()),
			},
			UploadSpeed: client.Double{
				Value: float64(p.GetUploadSpeed()),
			},
			ReadyParts: list.TotalParts, //TODO: real parts ready
		})
		list.DownloadSpeed.Value += float64(p.GetDownloadSpeed())
		list.UploadSpeed.Value += float64(p.GetUploadSpeed())
	}
	sort.Slice(list.Peers, func(i, j int) bool {
		a := list.Peers[i].DownloadSpeed.Value + list.Peers[i].UploadSpeed.Value
		b := list.Peers[j].DownloadSpeed.Value + list.Peers[j].UploadSpeed.Value
		if a != b {
			return a > b
		}
		return list.Peers[i].IP > list.Peers[j].IP
	})
	return &list, nil
}

func (c *Client) RemoveTorrent(ctx context.Context, hash []byte, withFiles bool) error {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return fmt.Errorf("torrent is not found")
	}
	return c.storage.RemoveTorrent(t, withFiles)
}

func (c *Client) SetActive(ctx context.Context, hash []byte, active bool) error {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return fmt.Errorf("torrent is not found")
	}

	if !active {
		t.Stop()
	} else {
		err := t.Start(true, false, false)
		if err != nil {
			return err
		}
	}
	return c.storage.SetTorrent(t)
}

func (c *Client) SetFilesPriority(ctx context.Context, hash []byte, names []string, priority int32) error {
	t := c.storage.GetTorrent(hash)
	if t == nil {
		return fmt.Errorf("torrent is not found")
	}

	list := t.GetActiveFilesIDs()

	for _, name := range names {
		fileInfo, err := t.GetFileOffsets(name)
		if err != nil {
			return err
		}

		if priority == 0 {
			// remove from list if exists
			for y, u := range list {
				if u == fileInfo.Index {
					list[y] = list[len(list)-1]
					break
				}
			}
			continue
		}

		found := false
		for _, u := range list {
			if u == fileInfo.Index {
				found = true
				break
			}
		}
		if !found {
			list = append(list, fileInfo.Index)
		}
	}
	return t.SetActiveFilesIDs(list)
}

func (c *Client) GetSpeedLimits(ctx context.Context) (*client.SpeedLimits, error) {
	return &client.SpeedLimits{
		Download: client.Double{
			Value: float64(c.connector.GetDownloadLimit()),
		},
		Upload: client.Double{
			Value: float64(c.connector.GetUploadLimit()),
		},
	}, nil
}

func (c *Client) SetSpeedLimits(ctx context.Context, download, upload int64) error {
	if download < 0 {
		download = 0
	}
	if upload < 0 {
		upload = 0
	}

	c.connector.SetDownloadLimit(uint64(download))
	c.connector.SetUploadLimit(uint64(upload))
	err := c.storage.SetSpeedLimits(uint64(download), uint64(upload))
	if err != nil {
		log.Error().Err(err).Msg("UI SET LIMITS ERR")
	}

	return nil
}

func (c *Client) GetNotifier() <-chan bool {
	return c.notify
}

func (c *Client) FetchProviderContract(ctx context.Context, torrentHash []byte, owner *address.Address) (*provider.ProviderContractData, error) {
	return c.provider.FetchProviderContract(ctx, torrentHash, owner)
}

func (c *Client) FetchProviderRates(ctx context.Context, torrentHash, providerKey []byte) (*provider.ProviderRates, error) {
	return c.provider.FetchProviderRates(ctx, torrentHash, providerKey)
}

func (c *Client) RequestProviderStorageInfo(ctx context.Context, torrentHash, providerKey []byte, owner *address.Address) (*provider.ProviderStorageInfo, error) {
	return c.provider.RequestProviderStorageInfo(ctx, torrentHash, providerKey, owner)
}

func (c *Client) BuildAddProviderTransaction(ctx context.Context, torrentHash []byte, owner *address.Address, providers []provider.NewProviderData) (addr *address.Address, bodyData, stateInit []byte, err error) {
	return c.provider.BuildAddProviderTransaction(ctx, torrentHash, owner, providers)
}

func (c *Client) BuildWithdrawalTransaction(torrentHash []byte, owner *address.Address) (addr *address.Address, bodyData []byte, err error) {
	return c.provider.BuildWithdrawalTransaction(torrentHash, owner)
}
