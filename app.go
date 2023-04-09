package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/tonutils/torrent-client/core/api"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/tonutils/torrent-client/core/daemon"
	"github.com/tonutils/torrent-client/oshook"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xssnick/tonutils-go/tl"
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
	ctx           context.Context
	api           *api.API
	daemonProcess *os.Process
	rootPath      string
	config        *Config
	loaded        bool

	openFileData []byte
	openFileHash string

	mx sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	oshook.HookStartup(a.openFile, a.openHash, a.ShowMsg)
	return a
}

func (a *App) exit(ctx context.Context) {
	if a.daemonProcess != nil {
		_ = a.daemonProcess.Kill()
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
	runtime2.MessageDialog(a.ctx, runtime2.MessageDialogOptions{
		Type:          runtime2.InfoDialog,
		Title:         "Info",
		Message:       text,
		DefaultButton: "OK",
	})
}

func (a *App) prepare() {
	oshook.HookStartup(a.openFile, a.openHash, a.ShowMsg)

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

	var exPath = CustomRoot // if we have defined root, use it for daemon path too
	if exPath == "" {
		ex, err := os.Executable()
		if err != nil {
			a.Throw(err)
		}
		exPath = filepath.Dir(ex)
	}

	a.daemonProcess, err = daemon.Run(a.ctx, a.rootPath, exPath, a.config.ListenAddr,
		strings.SplitN(a.config.DaemonControlAddr, ":", 2)[1], func(err error) {
			if err != nil {
				a.Throw(fmt.Errorf("storage daemon crashed with error:\n%w", err))
			}
		})
	if err != nil {
		a.Throw(err)
	}

	/*go func() {
		// TODO: forward port
		up, err := upnp.NewUPnP()
		if err != nil {
			return
			// a.Throw(err)
		}
	}()*/
}

func (a *App) ready(ctx context.Context) {
	a.prepare()

	// loading done, hook again to steal it from webview
	a.api = api.NewAPI(ctx, a.config.DaemonControlAddr, a.rootPath)
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

	runtime2.EventsEmit(a.ctx, "ready")
}

var onceCheck = sync.Once{}

func (a *App) CheckOpenedFile() {
	go func() {
		// TODO: on windows it may require some time, need to find exact reason
		//if runtime.GOOS == "windows" {
		///	time.Sleep(1000 * time.Millisecond)
		//}

		// wait for daemon ready
		for !a.loaded {
			time.Sleep(50 * time.Millisecond)
		}

		onceCheck.Do(func() {
			if a.openFileData != nil {
				a.openFile(a.openFileData)
				a.openFileData = nil
			} else if a.openFileHash != "" {
				a.openHash(a.openFileHash)
				a.openFileHash = ""
			}
		})
	}()
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
	str, _ := runtime2.OpenDirectoryDialog(a.ctx, runtime2.OpenDialogOptions{
		Title: "Select directory to convert to torrent",
	})
	return str
}

type TorrentCreateResult struct {
	Hash string
	Err  string
}

func (a *App) CreateTorrent(dir, description string) TorrentCreateResult {
	hash, err := a.api.CreateTorrent(dir, description)
	if err != nil {
		log.Println(err.Error())
		return TorrentCreateResult{Err: err.Error()}
	}
	return TorrentCreateResult{Hash: hash}
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
	})
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	err = os.WriteFile(path, m, 0766)
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

func (a *App) WantRemoveTorrent(hash string) {
	runtime2.EventsEmit(a.ctx, "want_remove_torrent", hash)
}

func (a *App) RemoveTorrent(hash string, withFiles, onlyNotInitiated bool) string {
	err := a.api.RemoveTorrent(hash, withFiles, onlyNotInitiated)
	if err != nil {
		log.Println(hash, err.Error())
		return err.Error()
	}
	return ""
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) SaveConfig(downloads, extIP string) string {
	a.config.DownloadsPath = downloads
	if a.config.ListenAddr != extIP {
		a.config.ListenAddr = extIP
		a.ShowMsg("Listen address will be changed on the next launch.")
	}
	err := a.config.SaveConfig(a.rootPath)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return ""
}

func (a *App) OpenFolder(path string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = "open"
	case "windows":
		cmd = "explorer"
	}

	exec.Command(cmd, path).Start()
}

func (a *App) OpenFolderSelectFile(path string) {
	switch runtime.GOOS {
	case "darwin", "linux":
		exec.Command("open", "-R", path).Start()
	case "windows":
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
