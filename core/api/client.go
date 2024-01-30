package api

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-storage-provider/pkg/contract"
	"log"
	"math/big"
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
	RawSize    int64
}

type NewProviderData struct {
	Key           string
	MaxSpan       uint32
	PricePerMBDay string
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
	PeersNum       int
	Uploaded       string
	Ratio          string

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
	Uploaded    string
	Ratio       string
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

type ProviderContract struct {
	Success   bool
	Deployed  bool
	Address   string
	Providers []Provider
	Balance   string
}

type ProviderRates struct {
	Success  bool
	Reason   string
	Provider Provider
}

type Provider struct {
	Key         string
	LastProof   string
	PricePerDay string
	Span        string
	Status      string
	Reason      string
	Progress    float64
	Data        NewProviderData
}

type ProviderStorageInfo struct {
	Status     string
	Reason     string
	Downloaded float64
}

type Transaction struct {
	Body      string
	StateInit string
	Address   string
	Amount    string
}

type StorageClient interface {
	GetTorrents(ctx context.Context) (*client.TorrentsList, error)
	AddByHash(ctx context.Context, hash []byte, dir string) (*client.TorrentFull, error)
	AddByMeta(ctx context.Context, meta []byte, dir string) (*client.TorrentFull, error)
	CreateTorrent(ctx context.Context, dir, description string, progressCallback func(done uint64, max uint64)) (*client.TorrentFull, error)
	GetTorrentFull(ctx context.Context, hash []byte) (*client.TorrentFull, error)
	GetTorrentMeta(ctx context.Context, hash []byte) ([]byte, error)
	GetPeers(ctx context.Context, hash []byte) (*client.PeersList, error)
	RemoveTorrent(ctx context.Context, hash []byte, withFiles bool) error
	SetActive(ctx context.Context, hash []byte, active bool) error
	SetFilesPriority(ctx context.Context, hash []byte, names []string, priority int32) error
	GetSpeedLimits(ctx context.Context) (*client.SpeedLimits, error)
	SetSpeedLimits(ctx context.Context, download, upload int64) error
	GetUploadStats(ctx context.Context, hash []byte) (uint64, error)
	FetchProviderContract(ctx context.Context, torrentHash []byte, owner *address.Address) (*client.ProviderContractData, error)
	FetchProviderRates(ctx context.Context, torrentHash, providerKey []byte) (*client.ProviderRates, error)
	RequestProviderStorageInfo(ctx context.Context, torrentHash, providerKey []byte, owner *address.Address) (*client.ProviderStorageInfo, error)
	BuildAddProviderTransaction(ctx context.Context, torrentHash []byte, owner *address.Address, providers []client.NewProviderData) (addr *address.Address, bodyData, stateInit []byte, err error)
	BuildWithdrawalTransaction(torrentHash []byte, owner *address.Address) (addr *address.Address, bodyData []byte, err error)
	GetNotifier() <-chan bool
}

type API struct {
	client   StorageClient
	torrents []*Torrent

	onSpeedsRefresh func(Speed)
	onListRefresh   func()
	onCompleted     func(hash []byte)
	globalCtx       context.Context
	mx              sync.RWMutex
}

func NewAPI(globalCtx context.Context, client StorageClient) *API {
	api := &API{
		globalCtx: globalCtx,
		client:    client,
	}

	return api
}

func (a *API) SetOnListRefresh(handler func()) {
	a.onListRefresh = handler
}

func (a *API) SetSpeedRefresh(handler func(Speed)) {
	a.onSpeedsRefresh = handler
}

func (a *API) SyncTorrents() error {
	a.mx.Lock()
	defer a.mx.Unlock()

	torr, err := a.client.GetTorrents(a.globalCtx)
	if err != nil {
		log.Println("sync err", err.Error())
		return err
	}

	var download, upload float64
	var list []*Torrent
iter:
	for _, torrent := range torr.Torrents {
		// optimization for inactive torrents, to fetch just once if inactive
		if !torrent.ActiveUpload && !torrent.ActiveDownload {
			for _, t := range a.torrents {
				if t.State == "inactive" && t.ID == hex.EncodeToString(torrent.Hash) {
					// nothing changed, just add it again
					list = append(list, t)
					continue iter
				}
			}
		}

		full, err := a.client.GetTorrentFull(a.globalCtx, torrent.Hash)
		if err != nil {
			continue
		}
		download += full.Torrent.DownloadSpeed
		upload += full.Torrent.UploadSpeed

		var lnPeers = 0
		if torrent.ActiveUpload || torrent.ActiveDownload {
			peers, err := a.client.GetPeers(a.globalCtx, torrent.Hash)
			if err != nil {
				continue
			}
			lnPeers = len(peers.Peers)
		}

		uploaded, err := a.client.GetUploadStats(a.globalCtx, torrent.Hash)
		if err != nil {
			continue
		}

		tr := formatTorrent(full, lnPeers, true, uploaded)
		if tr == nil {
			continue
		}
		list = append(list, tr)
	}

	a.torrents = list
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

func toRatio(uploaded, size uint64) string {
	if size == 0 || uploaded == 0 {
		return "0"
	}
	return new(big.Float).Quo(new(big.Float).SetUint64(uploaded), new(big.Float).SetUint64(size)).Text('f', 2)
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

func formatTorrent(full *client.TorrentFull, peersNum int, hide0Speed bool, uploaded uint64) *Torrent {
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
		PeersNum:       peersNum,
		Uploaded:       toSz(int64(uploaded)),
		Ratio:          toRatio(uploaded, uint64(dataSz)),
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

	t, err := a.client.GetTorrentFull(a.globalCtx, hashBytes)
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

	_, err = a.client.AddByHash(a.globalCtx, hashBytes, rootDir)
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
	_, err := a.client.AddByMeta(a.globalCtx, meta, rootDir)
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

func (a *API) CreateTorrent(ctx context.Context, dir, description string, progressCallback func(done uint64, max uint64)) (string, error) {
	t, err := a.client.CreateTorrent(ctx, dir, description, progressCallback)
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

	m, err := a.client.GetTorrentMeta(a.globalCtx, hashBytes)
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

	m, err := a.client.GetPeers(a.globalCtx, hashBytes)
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
	limits, err := a.client.GetSpeedLimits(a.globalCtx)
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

	return a.client.SetSpeedLimits(a.globalCtx, dow, up)
}

func (a *API) GetTorrentFiles(hash string) ([]*File, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	t, err := a.client.GetTorrentFull(a.globalCtx, hashBytes)
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

	t, err := a.client.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	var files []PlainFile
	for _, file := range t.Files {
		dir := ""
		if t.Torrent.DirName != nil {
			dir = *t.Torrent.DirName
		}

		var progress float64 = 100
		if file.Size > 0 {
			progress = float64(int64(float64(file.DownloadedSize)/float64(file.Size)*1000)) / 10
		}

		files = append(files, PlainFile{
			Path:       t.Torrent.RootDir + "/" + dir + file.Name,
			Name:       file.Name,
			Size:       toSz(file.Size),
			Downloaded: toSz(file.DownloadedSize),
			Progress:   progress,
			RawSize:    file.Size,
		})
	}

	return files, nil
}

func (a *API) GetInfo(hash string) (*TorrentInfo, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	peers, err := a.client.GetPeers(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	t, err := a.client.GetTorrentFull(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	uploaded, err := a.client.GetUploadStats(a.globalCtx, hashBytes)
	if err != nil {
		return nil, err
	}

	tr := formatTorrent(t, len(peers.Peers), false, uploaded)
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
		Uploaded:    tr.Uploaded,
		Ratio:       tr.Ratio,
	}, nil
}

func (a *API) RemoveTorrent(hash string, withFiles, onlyNotInitiated bool) error {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return err
	}

	if onlyNotInitiated {
		full, err := a.client.GetTorrentFull(a.globalCtx, hashBytes)
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

	err = a.client.RemoveTorrent(a.globalCtx, hashBytes, withFiles)
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

	err = a.client.SetActive(a.globalCtx, hashBytes, active)
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

	err = a.client.SetFilesPriority(a.globalCtx, hashBytes, list, int32(priority))
	if err != nil {
		return err
	}
	return nil
}

func (a *API) GetProviderContract(hash, ownerAddr string) ProviderContract {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return ProviderContract{Success: false}
	}

	addr, err := address.ParseAddr(ownerAddr)
	if err != nil {
		log.Println("failed to get provider contract, parse addr error:", err.Error())
		return ProviderContract{Success: false}
	}

	data, err := a.client.FetchProviderContract(a.globalCtx, hashBytes, addr)
	if err != nil {
		if errors.Is(err, contract.ErrNotDeployed) {
			return ProviderContract{Success: true, Deployed: false}
		}
		log.Println("failed to get provider contract:", err.Error())
		return ProviderContract{Success: false}
	}

	var providers []Provider
	for _, p := range data.Providers {
		since := "Never"
		snc := time.Since(p.LastProofAt)
		if snc < 2*time.Minute {
			since = fmt.Sprint(int(snc.Seconds())) + " seconds ago"
		} else if snc < 2*time.Hour {
			since = fmt.Sprint(int(snc.Minutes())) + " minutes ago"
		} else if snc < 48*time.Hour {
			since = fmt.Sprint(int(snc.Hours())) + " hours ago"
		} else if snc < 1000*24*time.Hour {
			since = fmt.Sprint(int(snc.Hours())/24) + " days ago"
		}

		every := ""
		if p.MaxSpan < 3600 {
			every = fmt.Sprint(p.MaxSpan/60) + " Minutes"
		} else if p.MaxSpan < 100*3600 {
			every = fmt.Sprint(p.MaxSpan/3600) + " Hours"
		} else {
			every = fmt.Sprint(p.MaxSpan/86400) + " Days"
		}

		psi, err := a.client.RequestProviderStorageInfo(a.globalCtx, hashBytes, p.Key, addr)
		if err != nil {
			log.Println("failed to request provider info:", err.Error())
			return ProviderContract{Success: false}
		}

		providers = append(providers, Provider{
			Key:         strings.ToUpper(hex.EncodeToString(p.Key)),
			LastProof:   since,
			Span:        every,
			Progress:    psi.Progress,
			Status:      psi.Status,
			Reason:      psi.Reason,
			PricePerDay: tlb.FromNanoTON(new(big.Int).Mul(p.RatePerMB.Nano(), big.NewInt(int64(data.Size/1024/1024)))).String() + " TON",
			Data: NewProviderData{
				Key:           hex.EncodeToString(p.Key),
				MaxSpan:       p.MaxSpan,
				PricePerMBDay: p.RatePerMB.Nano().String(),
			},
		})

	}

	bal := data.Balance.String()
	if idx := strings.IndexByte(bal, '.'); idx != -1 {
		if len(bal) > idx+4 {
			// max 4 digits after comma
			bal = bal[:idx+4]
		}
	}
	return ProviderContract{
		Success:   true,
		Deployed:  true,
		Address:   data.Address.String(),
		Providers: providers,
		Balance:   bal + " TON",
	}
}

func (a *API) FetchProviderRates(hash, provider string) ProviderRates {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return ProviderRates{Success: false, Reason: "failed to parse torrent hash: " + err.Error()}
	}

	providerBytes, err := toHashBytes(provider)
	if err != nil {
		return ProviderRates{Success: false, Reason: "failed to parse provider hash: " + err.Error()}
	}

	rates, err := a.client.FetchProviderRates(a.globalCtx, hashBytes, providerBytes)
	if err != nil {
		return ProviderRates{Success: false, Reason: err.Error()}
	}

	if rates.SpaceAvailableMB < rates.Size {
		return ProviderRates{Success: false, Reason: "torrent is too big for this provider"}
	}

	if !rates.Available {
		return ProviderRates{Success: false, Reason: "provider is not available"}
	}

	span := uint32(86400)
	if span > rates.MaxSpan {
		span = rates.MaxSpan
	} else if span < rates.MinSpan {
		span = rates.MinSpan
	}

	every := ""
	if span < 3600 {
		every = fmt.Sprint(span/60) + " minutes"
	} else if span < 100*3600 {
		every = fmt.Sprint(span/3600) + " hours"
	} else {
		every = fmt.Sprint(span/86400) + " days"
	}

	ratePerMB := rates.RatePerMBDay.Nano()
	min := rates.MinBounty.Nano()
	perDay := new(big.Int).Mul(ratePerMB, big.NewInt(int64(rates.Size/1024/1024)))
	if perDay.Cmp(min) < 0 {
		// increase reward to fit min bounty
		coff := new(big.Float).Quo(new(big.Float).SetInt(min), new(big.Float).SetInt(perDay))
		coff = coff.Add(coff, big.NewFloat(0.01)) // increase a bit to not be less than needed
		ratePerMB, _ = new(big.Float).Mul(new(big.Float).SetInt(ratePerMB), coff).Int(ratePerMB)
		perDay = new(big.Int).Mul(ratePerMB, big.NewInt(int64(rates.Size/1024/1024)))
	}

	return ProviderRates{
		Success: true,
		Provider: Provider{
			Key:         strings.ToUpper(hex.EncodeToString(providerBytes)),
			PricePerDay: tlb.FromNanoTON(perDay).String() + " TON",
			Span:        every,
			Data: NewProviderData{
				Key:           hex.EncodeToString(providerBytes),
				MaxSpan:       span,
				PricePerMBDay: ratePerMB.String(),
			},
		},
	}
}

func (a *API) RequestProviderStorageInfo(hash, provider, ownerAddr string) ProviderStorageInfo {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return ProviderStorageInfo{Status: "internal"}
	}

	providerBytes, err := toHashBytes(provider)
	if err != nil {
		return ProviderStorageInfo{Status: "internal"}
	}

	addr, err := address.ParseAddr(ownerAddr)
	if err != nil {
		log.Println("failed to get provider contract, parse addr error:", err.Error())
		return ProviderStorageInfo{Status: "internal"}
	}

	info, err := a.client.RequestProviderStorageInfo(a.globalCtx, hashBytes, providerBytes, addr)
	if err != nil {
		return ProviderStorageInfo{Status: "not_connected"}
	}

	return ProviderStorageInfo{
		Status:     info.Status,
		Reason:     info.Reason,
		Downloaded: info.Progress,
	}
}

func (a *API) BuildProviderContractData(hash, ownerAddr, amount string, providers []NewProviderData) (*Transaction, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	owner, err := address.ParseAddr(ownerAddr)
	if err != nil {
		return nil, err
	}

	amt, err := tlb.FromTON(amount)
	if err != nil {
		return nil, err
	}

	var prs []client.NewProviderData
	for _, p := range providers {
		keyBytes, err := toHashBytes(p.Key)
		if err != nil {
			return nil, fmt.Errorf("provider key: %w", err)
		}

		price, ok := new(big.Int).SetString(p.PricePerMBDay, 10)
		if !ok {
			return nil, fmt.Errorf("incorrect amount format")
		}

		prs = append(prs, client.NewProviderData{
			Address:       address.NewAddress(0, 0, keyBytes),
			MaxSpan:       p.MaxSpan,
			PricePerMBDay: tlb.FromNanoTON(price),
		})
	}

	addr, body, si, err := a.client.BuildAddProviderTransaction(a.globalCtx, hashBytes, owner, prs)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Body:      base64.StdEncoding.EncodeToString(body),
		StateInit: base64.StdEncoding.EncodeToString(si),
		Address:   addr.Bounce(false).String(),
		Amount:    amt.Nano().String(),
	}, nil
}

func (a *API) BuildWithdrawalContractData(hash, ownerAddr string) (*Transaction, error) {
	hashBytes, err := toHashBytes(hash)
	if err != nil {
		return nil, err
	}

	owner, err := address.ParseAddr(ownerAddr)
	if err != nil {
		return nil, err
	}

	addr, body, err := a.client.BuildWithdrawalTransaction(hashBytes, owner)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Body:      base64.StdEncoding.EncodeToString(body),
		StateInit: "",
		Address:   addr.Bounce(true).String(),
		Amount:    tlb.MustFromTON("0.03").Nano().String(),
	}, nil
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
