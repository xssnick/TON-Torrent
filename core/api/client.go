package api

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/liteclient"
	"log"
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

type Torrent struct {
	ID       string
	Name     string
	Size     string
	Progress float64
	State    string
	Upload   string
	Download string
	Path     string
}

type API struct {
	daemon   *client.StorageClient
	torrents []Torrent

	onListRefresh func()
	onCompleted   func(hash []byte)
	globalCtx     context.Context
	mx            sync.RWMutex
}

func NewAPI(globalCtx context.Context, addr string, authKey ed25519.PrivateKey, serverKey string) *API {
	api := &API{
		globalCtx: globalCtx,
	}

	pool := liteclient.NewConnectionPoolWithAuth(authKey)
	err := pool.AddConnection(context.Background(), addr, serverKey)
	if err != nil {
		panic(err)
	}

	api.daemon = client.NewStorageClient(pool)

	go func() {
		for {
			err := api.SyncTorrents()
			if err != nil {
				log.Println("SYNC ERR:", err.Error())
			}
			time.Sleep(250 * time.Millisecond)
		}
	}()

	return api
}

func (a *API) SetOnListRefresh(handler func()) {
	a.onListRefresh = handler
}

func (a *API) SetOnCompleted(handler func()) {
	a.onListRefresh = handler
}

func (a *API) SyncTorrents() error {
	torr, err := a.daemon.GetTorrents(a.globalCtx)
	if err != nil {
		log.Println("sync err", err.Error())
		return err
	}

	var list []Torrent
	for _, torrent := range torr.Torrents {
		if torrent.IncludedSize == nil { // only header loaded (just added)
			continue
		}

		dataSz := *torrent.TotalSize - *torrent.IncludedSize
		var progress float64 = 0
		if torrent.DownloadedSize > *torrent.IncludedSize {
			// header downloaded
			progress = (float64(torrent.DownloadedSize-*torrent.IncludedSize) / float64(dataSz)) * 100
		}

		// round 2 decimals
		progress = float64(int64(progress*10)) / 10

		// choose something as a name
		name := *torrent.Description
		if name == "" {
			if torrent.DirName != nil {
				name = *torrent.DirName
			}

			if name == "/" || name == "" {
				name = hex.EncodeToString(torrent.Hash)
			}
		}

		var dowSpeed, uplSpeed string
		if !torrent.Completed && torrent.ActiveDownload {
			// display speed only if in progress
			dowSpeed = toSpeed(int64(torrent.DownloadSpeed))
		}

		if torrent.ActiveUpload {
			uplSpeed = toSpeed(int64(torrent.UploadSpeed))
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

		list = append(list, Torrent{
			ID:       hex.EncodeToString(torrent.Hash),
			Name:     name,
			Size:     toSz(int64(dataSz)),
			Progress: progress,
			State:    state,
			Upload:   uplSpeed,
			Download: dowSpeed,
			Path:     path,
		})
	}

	a.mx.Lock()
	a.torrents = list
	a.mx.Unlock()

	if a.onListRefresh != nil {
		a.onListRefresh()
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

func toSpeed(speed int64) string {
	if speed == 0 {
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

func (a *API) GetTorrents() []Torrent {
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
		return err
	}

	return nil
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

func (a *API) RemoveTorrent(hash string, withFiles bool) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
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
