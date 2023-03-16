package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"github.com/tonutils/torrent-client/core/api"
	"github.com/tonutils/torrent-client/core/daemon"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx context.Context
	api *api.API
	mx  sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go daemon.Run()
	time.Sleep(1 * time.Second)

	/*go func() {
		// TODO: forward port
		up, err := upnp.NewUPnP()
		if err != nil {
			return
			// panic(err)
		}
	}()*/

	ourKey, err := base64.StdEncoding.DecodeString("EwP8Ano8Fn+a8lQTPkYHuKdXUuUyt1kK1ooH2Uf9DIM=")
	if err != nil {
		panic(err)
	}
	pk := ed25519.NewKeyFromSeed(ourKey)

	a.api = api.NewAPI(ctx, "127.0.0.1:5555", pk, "MLQ71gfZoJW10MNKNEFNm19qlCk+XZ6WRKEg4QKzEDU=")
	a.api.SetOnListRefresh(func() {
		runtime2.EventsEmit(a.ctx, "update")
	})

	runtime2.EventsOn(a.ctx, "refresh", func(optionalData ...interface{}) {
		_ = a.api.SyncTorrents()
	})
}

func (a *App) GetTorrents() []api.Torrent {
	list := a.api.GetTorrents()
	if list == nil {
		list = []api.Torrent{}
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

func (a *App) AddTorrentByHash(hash string) string {
	err := a.api.AddTorrentByHash(hash, "/Users/xssnick/Downloads/"+strings.ToUpper(hash))
	// "/usr/bin/ton/storage/storage-daemon/storage-db/torrent/torrent-files"
	if err != nil {
		return err.Error()
	}

	return ""
}

func (a *App) CheckHeader(hash string) bool {
	hasHeader, _ := a.api.CheckTorrentHeader(hash)
	return hasHeader
}

func (a *App) RemoveTorrent(hash string, withFiles bool) string {
	err := a.api.RemoveTorrent(hash, withFiles)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) OpenFolder(path string) {
	println(path)
	var cmd string
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = "open"
	case "windows":
		cmd = "explorer"
	}

	exec.Command(cmd, path).Start()
}

func (a *App) SetActive(hash string, active bool) string {
	err := a.api.SetActive(hash, active)
	if err != nil {
		return err.Error()
	}
	return ""
}

func downloadsPath() string {
	var path string
	switch runtime.GOOS {
	case "darwin", "linux":
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, "Downloads")
	case "windows":
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, "Downloads")
	}
	return path
}
