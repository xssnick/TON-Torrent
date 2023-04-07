package api

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/liteclient"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type File struct {
	Name  string
	Size  string
	Child []*File
	Path  string

	rawSz int64
}

type PlainFile struct {
	Path       string
	Name       string
	Size       string
	Downloaded string
	Progress   float64
}

type Torrent struct {
	ID             string
	Name           string
	Size           string
	DownloadedSize string
	Progress       float64
	State          string
	Upload         string
	Download       string
	Path           string

	rawDowSpeed    int64
	rawDownloaded  int64
	rawSize        int64
	rawDescription string
}

type TorrentInfo struct {
	Description string
	Size        string
	Downloaded  string
	TimeLeft    string
	Progress    float64
	State       string
	Upload      string
	Download    string
	Path        string
	Peers       int
	AddedAt     string
}

type SpeedLimits struct {
	Download int64
	Upload   int64
}

type Peer struct {
	IP       string
	ADNL     string
	Upload   string
	Download string
}

type Speed struct {
	Upload   string
	Download string
}

type API struct {
	daemon   *client.StorageClient
	torrents []*Torrent

	onSpeedsRefresh func(Speed)
	onListRefresh   func()
	onCompleted     func(hash []byte)
	globalCtx       context.Context
	mx              sync.RWMutex
}

func NewAPI(globalCtx context.Context, addr string, rootPath string) *API {
	api := &API{
		globalCtx: globalCtx,
	}

	var authKey ed25519.PrivateKey
	var serverKey string
	for {
		clientKey, err := os.ReadFile(rootPath + "/storage-db/cli-keys/client")
		if err != nil {
			log.Println("storage-db/client read err:", err.Error(), "retrying...")
			time.Sleep(300 * time.Millisecond)
			continue
		}
		authKey = ed25519.NewKeyFromSeed(clientKey[4:])
		break
	}

	for {
		key, err := os.ReadFile(rootPath + "/storage-db/cli-keys/server.pub")
		if err != nil {
			log.Println("storage-db/server.pub read err:", err.Error(), "retrying...")
			time.Sleep(300 * time.Millisecond)
			continue
		}
		serverKey = base64.StdEncoding.EncodeToString(key[4:])
		break
	}

	pool := liteclient.NewConnectionPoolWithAuth(authKey)
	for {
		// try to connect to daemon
		err := pool.AddConnection(context.Background(), addr, serverKey)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	api.daemon = client.NewStorageClient(pool)

	go func() {
		for {
			err := api.SyncTorrents()
			if err != nil {
				log.Println("SYNC ERR:", err.Error())
			}
			time.Sleep(150 * time.Millisecond)
		}
	}()

	return api
}

func (a *API) SetOnListRefresh(handler func()) {
	a.onListRefresh = handler
}

func (a *API) SetSpeedRefresh(handler func(Speed)) {
	a.onSpeedsRefresh = handler
}

func (a *API) SyncTorrents() error {
	torr, err := a.daemon.GetTorrents(a.globalCtx)
	if err != nil {
		log.Println("sync err", err.Error())
		return err
	}

	var download, upload float64
	var list []*Torrent
	for _, torrent := range torr.Torrents {
		full, err := a.daemon.GetTorrentFull(a.globalCtx, torrent.Hash)
		if err != nil {
			continue
		}
		download += full.Torrent.DownloadSpeed
		upload += full.Torrent.UploadSpeed

		tr := formatTorrent(full, true)
		if tr == nil {
			continue
		}
		list = append(list, tr)
	}

	a.mx.Lock()
	a.torrents = list
	a.mx.Unlock()

	if a.onListRefresh != nil {
		a.onListRefresh()
	}
	if a.onSpeedsRefresh != nil {
		a.onSpeedsRefresh(Speed{
			Upload:   toSpeed(int64(upload), false),
			Download: toSpeed(int64(download), false),
		})
	}

	return nil
}

func toSz(sz int64) string {
	switch {
	case sz < 1024:
		return fmt.Sprintf("%d Bytes", sz)
	case sz < 1024*1024:
		return fmt.Sprintf("%.2f KB", float64(sz)/1024)
	case sz < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB", float64(sz)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(sz)/(1024*1024*1024))
	}
}

func toSpeed(speed int64, zeroEmpty bool) string {
	if speed == 0 && zeroEmpty {
		return ""
	}

	switch {
	case speed < 1024:
		return fmt.Sprintf("%d Bytes/s", speed)
	case speed < 1024*1024:
		return fmt.Sprintf("%.2f KB/s", float64(speed)/1024)
	case speed < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB/s", float64(speed)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB/s", float64(speed)/(1024*1024*1024))
	}
}

func formatTorrent(full *client.TorrentFull, hide0Speed bool) *Torrent {
	torrent := full.Torrent // newer object
	var dataSz, downloadedSz int64
	for _, file := range full.Files {
		if file.Priority > 0 {
			dataSz += file.Size
			downloadedSz += file.DownloadedSize
		}
	}

	if dataSz == 0 { // only header loaded (just added)
		return nil
	}

	progress := (float64(downloadedSz) / float64(dataSz)) * 100

	// round 2 decimals
	progress = float64(int64(progress*10)) / 10

	var rawDesc string
	// choose something as a name
	name := *torrent.Description
	if name == "" {
		if torrent.DirName != nil {
			name = *torrent.DirName
		}

		if name == "/" || name == "" {
			name = hex.EncodeToString(torrent.Hash)
		}
	} else {
		rawDesc = name
	}

	var dowSpeed, uplSpeed string
	if !torrent.Completed && torrent.ActiveDownload {
		// display speed only if in progress
		dowSpeed = toSpeed(int64(torrent.DownloadSpeed), hide0Speed)
	}

	if torrent.ActiveUpload {
		uplSpeed = toSpeed(int64(torrent.UploadSpeed), hide0Speed)
	}

	state := "fail"
	if torrent.ActiveDownload && !torrent.Completed {
		state = "downloading"
	} else if torrent.ActiveUpload && torrent.Completed {
		state = "seeding"
	} else {
		state = "inactive"
	}

	path := torrent.RootDir
	if torrent.DirName != nil {
		path += "/" + *torrent.DirName
	}

	return &Torrent{
		ID:             strings.ToUpper(hex.EncodeToString(torrent.Hash)),
		Name:           name,
		Size:           toSz(dataSz),
		DownloadedSize: toSz(downloadedSz),
		Progress:       progress,
		State:          state,
		Upload:         uplSpeed,
		Download:       dowSpeed,
		Path:           path,
		rawDowSpeed:    int64(torrent.DownloadSpeed),
		rawDownloaded:  downloadedSz,
		rawSize:        dataSz,
		rawDescription: rawDesc,
	}
}

func (a *API) GetTorrents() []*Torrent {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.torrents
}

func (a *API) CheckTorrentHeader(hash string) (bool, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return false, err
	}

	t, err := a.daemon.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return false, err
	}

	// if we have size then we have a header
	return t.Torrent.IncludedSize != nil, nil
}

func (a *API) AddTorrentByHash(hash, rootDir string) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
	}

	_, err = a.daemon.AddByHash(a.globalCtx, hashBytes, rootDir)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate hash") {
			// if we already have it - still ok, maybe use wants to load more files,
			// or something went wrong on kill switch
			return nil
		}
		return err
	}

	return nil
}

func (a *API) AddTorrentByMeta(meta []byte, rootDir string) error {
	_, err := a.daemon.AddByMeta(a.globalCtx, meta, rootDir)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate hash") {
			// if we already have it - still ok, maybe use wants to load more files,
			// or something went wrong on kill switch
			return nil
		}
		return err
	}
	return nil
}

func (a *API) CreateTorrent(dir, description string) (string, error) {
	t, err := a.daemon.CreateTorrent(a.globalCtx, dir, description)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(t.Torrent.Hash)), nil
}

func (a *API) GetTorrentMeta(hash string) ([]byte, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	m, err := a.daemon.GetTorrentMeta(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (a *API) GetPeers(hash string) ([]Peer, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	m, err := a.daemon.GetPeers(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	var peers []Peer
	for _, p := range m.Peers {
		peers = append(peers, Peer{
			IP:       p.IP,
			ADNL:     strings.ToUpper(hex.EncodeToString(p.ADNL)),
			Upload:   toSpeed(int64(p.UploadSpeed.Value), true),
			Download: toSpeed(int64(p.DownloadSpeed.Value), true),
		})
	}
	return peers, nil
}

func (a *API) GetSpeedLimits() (*SpeedLimits, error) {
	limits, err := a.daemon.GetSpeedLimits(a.globalCtx)
	if err != nil {
		return nil, err
	}

	return &SpeedLimits{
		Download: int64(limits.Download.Value) / 1024, // to KB
		Upload:   int64(limits.Upload.Value) / 1024,
	}, nil
}

func (a *API) SetSpeedLimits(limits *SpeedLimits) error {
	dow, up := int64(-1), int64(-1)
	if limits.Download >= 0 {
		dow = limits.Download * 1024
	}
	if limits.Upload >= 0 {
		up = limits.Upload * 1024
	}

	return a.daemon.SetSpeedLimits(a.globalCtx, dow, up)
}

func (a *API) GetTorrentFiles(hash string) ([]*File, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	t, err := a.daemon.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	sort.Slice(t.Files, func(i, j int) bool {
		return t.Files[i].Name < t.Files[j].Name
	})

	root := &File{}

	for _, file := range t.Files {
		path := strings.Split(file.Name, "/")

		cur := root
	next:
		for i, s := range path { // create dir structure
			cur.rawSz += file.Size

			if i == len(path)-1 {
				cur.Child = append(cur.Child, &File{
					Path:  file.Name,
					Name:  s,
					rawSz: file.Size,
				})
				continue
			}

			// check if we already have this dir
			for _, c := range cur.Child {
				if c.Name == s {
					cur = c
					continue next
				}
			}

			// we don't have a dir yet, create new
			add := &File{
				Path: strings.Join(path[:i+1], "/"),
				Name: s,
			}
			cur.Child = append(cur.Child, add)
			// dive into next dir
			cur = add
		}
	}
	root.calcSize()

	return root.Child, nil
}

func (a *API) GetPlainFiles(hash string) ([]PlainFile, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	t, err := a.daemon.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	var files []PlainFile
	for _, file := range t.Files {
		dir := ""
		if t.Torrent.DirName != nil {
			dir = *t.Torrent.DirName
		}

		var progress float64
		if file.Size > 0 {
			progress = float64(int64(float64(file.DownloadedSize)/float64(file.Size)*1000)) / 10
		}

		files = append(files, PlainFile{
			Path:       t.Torrent.RootDir + "/" + dir + file.Name,
			Name:       file.Name,
			Size:       toSz(file.Size),
			Downloaded: toSz(file.DownloadedSize),
			Progress:   progress,
		})
	}

	return files, nil
}

func (a *API) GetInfo(hash string) (*TorrentInfo, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	peers, err := a.daemon.GetPeers(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	t, err := a.daemon.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	tr := formatTorrent(t, false)
	if tr == nil {
		return nil, fmt.Errorf("not initialized torrent")
	}

	leftSz := tr.rawSize - tr.rawDownloaded

	var left string
	if tr.rawDowSpeed > 0 {
		secsLeft := leftSz / tr.rawDowSpeed
		minutesLeft := secsLeft / 60
		hoursLeft := minutesLeft / 60
		if hoursLeft > 0 {
			left = fmt.Sprintf("%dh ", hoursLeft)
		}
		if minutesLeft > 0 {
			left += fmt.Sprintf("%dm ", minutesLeft%60)
		}
		if hoursLeft == 0 {
			left += fmt.Sprintf("%ds ", secsLeft%60)
		}
	} else {
		left = "∞"
	}

	return &TorrentInfo{
		Description: tr.Name,
		Size:        tr.Size,
		Downloaded:  tr.DownloadedSize,
		TimeLeft:    left,
		Progress:    tr.Progress,
		State:       tr.State,
		Upload:      tr.Upload,
		Download:    tr.Download,
		Path:        tr.Path,
		Peers:       len(peers.Peers),
		AddedAt:     time.Unix(int64(t.Torrent.AddedAt), 0).Format("02 Jan 2006 15:04:05"),
	}, nil
}

func (a *API) RemoveTorrent(hash string, withFiles, onlyNotInitiated bool) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
	}

	if onlyNotInitiated {
		full, err := a.daemon.GetTorrentFull(a.globalCtx, hashBytes)
		if err != nil {
			return err
		}

		initiated := false
		for _, file := range full.Files {
			if file.Priority > 0 {
				initiated = true
				break
			}
		}
		if initiated {
			return nil
		}
	}

	err = a.daemon.RemoveTorrent(a.globalCtx, hashBytes, withFiles)
	if err != nil {
		return err
	}
	return nil
}

func (a *API) SetActive(hash string, active bool) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
	}

	err = a.daemon.SetActive(a.globalCtx, hashBytes, active)
	if err != nil {
		return err
	}
	return nil
}

func (a *API) SetPriorities(hash string, list []string, priority int) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
	}

	for _, f := range list {
		err = a.daemon.SetFilePriority(a.globalCtx, hashBytes, f, int32(priority))
		if err != nil {
			return err
		}
	}
	return nil
}

func toHashBytes(hash string) ([]byte, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid hash hex string")
	}

	if len(hashBytes) != 32 {
		return nil, fmt.Errorf("invalid hash size, length should be 64 symbols")
	}
	return hashBytes, nil
}

func (f *File) calcSize() {
	f.Size = toSz(f.rawSz)
	for _, file := range f.Child {
		file.calcSize()
	}
}
