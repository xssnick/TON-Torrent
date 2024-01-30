package gostorage

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/adnl"
	"github.com/xssnick/tonutils-go/adnl/dht"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-storage-provider/pkg/transport"
	"github.com/xssnick/tonutils-storage/config"
	"github.com/xssnick/tonutils-storage/db"
	"github.com/xssnick/tonutils-storage/storage"
	"log"
	"net"
	"os"
	"sort"
	"sync"
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
	connector storage.NetConnector
	api       ton.APIClientWrapped
	provider  *transport.Client

	infoCache map[string]*client.ProviderStorageInfo
	mx        sync.RWMutex

	notify chan bool
}

func NewClient(dbPath string, cfg Config) (*Client, error) {
	c := &Client{
		infoCache: map[string]*client.ProviderStorageInfo{},
		notify:    make(chan bool, 10), // to refresh fast a bit after
	}

	ldb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load db: %w", err)
	}

	var ip net.IP
	if cfg.ExternalIP != "" {
		ip = net.ParseIP(cfg.ExternalIP)
		if ip == nil {
			return nil, fmt.Errorf("external ip is invalid")
		}
	}

	var lsCfg *liteclient.GlobalConfig
	if cfg.NetworkConfigPath != "" {
		lsCfg, err = liteclient.GetConfigFromFile(cfg.NetworkConfigPath)
		if err != nil {
			pterm.Error.Println("Failed to load ton network config from file:", err.Error())
			os.Exit(1)
		}
	} else {
		lsCfg, err = liteclient.GetConfigFromUrl(context.Background(), "https://ton.org/global.config.json")
		if err != nil {
			pterm.Warning.Println("Failed to download ton config:", err.Error(), "; We will take it from static cache")
			lsCfg = &liteclient.GlobalConfig{}
			if err = json.NewDecoder(bytes.NewBufferString(config.FallbackNetworkConfig)).Decode(lsCfg); err != nil {
				pterm.Error.Println("Failed to parse fallback ton config:", err.Error())
				os.Exit(1)
			}
		}
	}

	lsPool := liteclient.NewConnectionPool()
	c.api = ton.NewAPIClient(lsPool, ton.ProofCheckPolicyFast)

	// connect async to not slow down main processes
	go func() {
		for {
			if err := lsPool.AddConnectionsFromConfig(context.Background(), lsCfg); err != nil {
				pterm.Warning.Println("Failed to add connections from ton config:", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
	}()

	gate := adnl.NewGateway(cfg.Key)

	serverMode := ip != nil
	if serverMode {
		gate.SetExternalIP(ip)
		if err = gate.StartServer(cfg.ListenAddr); err != nil {
			return nil, fmt.Errorf("failed to start adnl gateway in server mode: %w", err)
		}
	} else {
		if err = gate.StartClient(); err != nil {
			return nil, fmt.Errorf("failed to start adnl gateway in client mode: %w", err)
		}
	}

	dhtGate := adnl.NewGateway(cfg.Key)
	if err = dhtGate.StartClient(); err != nil {
		return nil, fmt.Errorf("failed to init dht adnl gateway: %w", err)
	}

	dhtClient, err := dht.NewClientFromConfig(dhtGate, lsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init dht client: %w", err)
	}

	providerGateSeed := sha256.Sum256(cfg.Key.Seed())
	gateKey := ed25519.NewKeyFromSeed(providerGateSeed[:])
	gateProvider := adnl.NewGateway(gateKey)
	if err = gateProvider.StartClient(); err != nil {
		return nil, fmt.Errorf("failed to start adnl gateway for provider: %w", err)
	}
	c.provider = transport.NewClient(gateProvider, dhtClient)

	downloadGate := adnl.NewGateway(cfg.Key)
	if err = downloadGate.StartClient(); err != nil {
		return nil, fmt.Errorf("failed to init downloader gateway: %w", err)
	}

	srv := storage.NewServer(dhtClient, gate, cfg.Key, serverMode)

	ch := make(chan db.Event, 1)
	c.connector = storage.NewConnector(srv)
	c.storage, err = db.NewStorage(ldb, c.connector, false, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}
	srv.SetStorage(c.storage)

	d, u, err := c.storage.GetSpeedLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to load speed limits: %w", err)
	}

	c.connector.SetDownloadLimit(d)
	c.connector.SetUploadLimit(u)

	go func() {
		ticker := time.Tick(3 * time.Second)
		for {
			select {
			case <-ch:
			case <-ticker:
			}

			select {
			case c.notify <- true:
			default:
			}
		}
	}()

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
	torrent := client.Torrent{
		Hash:           t.BagID,
		AddedAt:        uint32(t.CreatedAt.Unix()),
		RootDir:        t.Path,
		ActiveDownload: activeDownload,
		ActiveUpload:   activeUpload,
		Completed:      false,
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
	return mf.Serialize()
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
	list.TotalParts = int64(t.PiecesNum())

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
		log.Println("UI SET LIMITS ERR:", err)
	}

	return nil
}

func (c *Client) GetNotifier() <-chan bool {
	return c.notify
}
