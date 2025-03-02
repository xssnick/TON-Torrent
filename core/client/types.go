package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"github.com/xssnick/tonutils-storage/storage"
	"math"
)

func init() {
	tl.Register(AddByMeta{}, "storage.daemon.addByMeta meta:bytes root_dir:string start_download:Bool allow_upload:Bool priorities:(vector storage.PriorityAction) flags:# = storage.daemon.TorrentFull")
	tl.Register(AddByHash{}, "storage.daemon.addByHash hash:int256 root_dir:string start_download:Bool allow_upload:Bool priorities:(vector storage.PriorityAction) flags:# = storage.daemon.TorrentFull")
	tl.Register(GetTorrents{}, "storage.daemon.getTorrents flags:# = storage.daemon.TorrentList")
	tl.Register(RemoveTorrent{}, "storage.daemon.removeTorrent hash:int256 remove_files:Bool = storage.daemon.Success")
	tl.Register(SetActiveUpload{}, "storage.daemon.setActiveUpload hash:int256 active:Bool = storage.daemon.Success")
	tl.Register(SetActiveDownload{}, "storage.daemon.setActiveDownload hash:int256 active:Bool = storage.daemon.Success")
	tl.Register(GetTorrentFull{}, "storage.daemon.getTorrentFull hash:int256 flags:# = storage.daemon.TorrentFull")
	tl.Register(SetFilePriorityByName{}, "storage.daemon.setFilePriorityByName hash:int256 name:string priority:int = storage.daemon.SetPriorityStatus")
	tl.Register(CreateTorrent{}, "storage.daemon.createTorrent path:string description:string allow_upload:Bool copy_inside:Bool flags:# = storage.daemon.TorrentFull")
	tl.Register(GetTorrentMeta{}, "storage.daemon.getTorrentMeta hash:int256 flags:# = storage.daemon.TorrentMeta")
	tl.Register(GetPeers{}, "storage.daemon.getTorrentPeers hash:int256 flags:# = storage.daemon.PeerList")
	tl.Register(GetSpeedLimits{}, "storage.daemon.getSpeedLimits flags:# = storage.daemon.SpeedLimits")
	tl.Register(SetSpeedLimits{}, "storage.daemon.setSpeedLimits flags:# download:flags.0?double upload:flags.1?double = storage.daemon.Success")

	tl.Register(TorrentsList{}, "storage.daemon.torrentList torrents:(vector storage.daemon.torrent) = storage.daemon.TorrentList")
	tl.Register(Torrent{}, "storage.daemon.torrent hash:int256 flags:# total_size:flags.0?long description:flags.0?string files_count:flags.1?long included_size:flags.1?long dir_name:flags.1?string downloaded_size:long added_at:int root_dir:string active_download:Bool active_upload:Bool completed:Bool download_speed:double upload_speed:double fatal_error:flags.2?string = storage.daemon.Torrent")
	tl.Register(TorrentFull{}, "storage.daemon.torrentFull torrent:storage.daemon.torrent files:(vector storage.daemon.fileInfo) = storage.daemon.TorrentFull")
	tl.Register(TorrentMeta{}, "storage.daemon.torrentMeta meta:bytes = storage.daemon.TorrentMeta")
	tl.Register(PriorityActionAll{}, "storage.priorityAction.all priority:int = storage.PriorityAction")
	tl.Register(PriorityActionIndex{}, "storage.priorityAction.idx idx:long priority:int = storage.PriorityAction")
	tl.Register(PriorityActionName{}, "storage.priorityAction.name name:string priority:int = storage.PriorityAction")
	tl.Register(FileInfo{}, "storage.daemon.fileInfo name:string size:long flags:# priority:int downloaded_size:long = storage.daemon.FileInfo")
	tl.Register(PriorityStatusSet{}, "storage.daemon.prioritySet = storage.daemon.SetPriorityStatus")
	tl.Register(PriorityStatusPending{}, "storage.daemon.priorityPending = storage.daemon.SetPriorityStatus")
	tl.Register(Peer{}, "storage.daemon.peer adnl_id:int256 ip_str:string download_speed:double upload_speed:double ready_parts:long = storage.daemon.Peer")
	tl.Register(PeersList{}, "storage.daemon.peerList peers:(vector storage.daemon.peer) download_speed:double upload_speed:double total_parts:long = storage.daemon.PeerList")
	tl.Register(SpeedLimits{}, "storage.daemon.speedLimits download:double upload:double = storage.daemon.SpeedLimits")

	tl.Register(DaemonError{}, "storage.daemon.queryError message:string = storage.daemon.QueryError")
	tl.Register(Success{}, "storage.daemon.success = storage.daemon.Success")
}

type DaemonError struct {
	Message string `tl:"string"`
}

type Success struct{}

type PriorityActionAll struct {
	Priority int32 `tl:"int"`
}

type PriorityActionIndex struct {
	Index    int64 `tl:"long"`
	Priority int32 `tl:"int"`
}

type PriorityActionName struct {
	Name     string `tl:"string"`
	Priority int32  `tl:"int"`
}

type PriorityStatusSet struct{}
type PriorityStatusPending struct{}

type SetFilePriorityByName struct {
	Hash     []byte `tl:"int256"`
	Name     string `tl:"string"`
	Priority int32  `tl:"int"`
}

type GetTorrentFull struct {
	Hash  []byte `tl:"int256"`
	Flags uint32 `tl:"int"`
}

type RemoveTorrent struct {
	Hash        []byte `tl:"int256"`
	RemoveFiles bool   `tl:"bool"`
}

type GetPeers struct {
	Hash  []byte `tl:"int256"`
	Flags uint32 `tl:"int"`
}

type CreateTorrent struct {
	Path        string `tl:"string"`
	Description string `tl:"string"`
	AllowUpload bool   `tl:"bool"`
	CopyInside  bool   `tl:"bool"`
	Flags       uint32 `tl:"int"`
}

type AddByMeta struct {
	Meta          []byte `tl:"bytes"`
	RootDir       string `tl:"string"`
	StartDownload bool   `tl:"bool"`
	AllowUpload   bool   `tl:"bool"`
	Priorities    []any  `tl:"vector struct boxed [storage.priorityAction.all,storage.priorityAction.idx,storage.priorityAction.name]"`
	Flags         uint32 `tl:"int"`
}

type AddByHash struct {
	Hash          []byte `tl:"int256"`
	RootDir       string `tl:"string"`
	StartDownload bool   `tl:"bool"`
	AllowUpload   bool   `tl:"bool"`
	Priorities    []any  `tl:"vector struct boxed [storage.priorityAction.all,storage.priorityAction.idx,storage.priorityAction.name]"`
	Flags         uint32 `tl:"int"`
}

type TorrentFull struct {
	Torrent Torrent    `tl:"struct"`
	Files   []FileInfo `tl:"vector struct"`
}

type GetTorrents struct {
	Flags uint32 `tl:"int"`
}

type GetTorrentMeta struct {
	Hash  []byte `tl:"int256"`
	Flags uint32 `tl:"int"`
}

type TorrentMeta struct {
	Meta []byte `tl:"bytes"`
}

type Double struct {
	Value float64
}

type Peer struct {
	ADNL          []byte `tl:"int256"`
	IP            string `tl:"string"`
	DownloadSpeed Double `tl:"struct"`
	UploadSpeed   Double `tl:"struct"`
	ReadyParts    int64  `tl:"long"`
}

type PeersList struct {
	Peers         []Peer `tl:"vector struct"`
	DownloadSpeed Double `tl:"struct"`
	UploadSpeed   Double `tl:"struct"`
	TotalParts    int64  `tl:"long"`
}

type FileInfo struct {
	Name           string `tl:"string"`
	Size           int64  `tl:"long"`
	Flags          uint32 `tl:"int"`
	Priority       int32  `tl:"int"`
	DownloadedSize int64  `tl:"long"`
}

type SetActiveDownload struct {
	Hash   []byte `tl:"int256"`
	Active bool   `tl:"bool"`
}

type SetActiveUpload struct {
	Hash   []byte `tl:"int256"`
	Active bool   `tl:"bool"`
}

type SpeedLimits struct {
	Download Double `tl:"struct"`
	Upload   Double `tl:"struct"`
}

type GetSpeedLimits struct {
	Flags uint32 `tl:"int"`
}

type SetSpeedLimits struct {
	Flags    uint32  `tl:"int"`
	Download *Double `tl:"struct"`
	Upload   *Double `tl:"struct"`
}

type Torrent struct {
	Hash           []byte
	Flags          uint32
	TotalSize      *uint64 // 0
	Description    *string // 0
	FilesCount     *uint64 // 1
	IncludedSize   *uint64 // 1
	DirName        *string // 1
	AddedAt        uint32
	DownloadedSize uint64
	RootDir        string
	ActiveDownload bool
	ActiveUpload   bool
	Completed      bool
	DownloadSpeed  float64
	UploadSpeed    float64
	FatalError     *string // 2

	Verified bool
}

type TorrentsList struct {
	Torrents []Torrent `tl:"vector struct"`
}

type MetaFile struct {
	Hash      []byte
	Info      storage.TorrentInfo
	RootProof *cell.Cell
	Header    *storage.TorrentHeader
}

var BoolTrue = tl.CRC("boolTrue = Bool")

func (t *Torrent) Parse(data []byte) (_ []byte, err error) {
	// Manual parse because of not standard array definition
	if len(data) < 36 {
		return nil, fmt.Errorf("too short sizes data to parse")
	}
	t.Hash = data[:32]
	data = data[32:]
	t.Flags = binary.LittleEndian.Uint32(data)
	data = data[4:]
	if t.Flags&1 != 0 {
		sz := binary.LittleEndian.Uint64(data)
		data = data[8:]

		var desc []byte
		desc, data, err = tl.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse description: %w", err)
		}
		descStr := string(desc)

		t.TotalSize = &sz
		t.Description = &descStr
	}
	if t.Flags&2 != 0 {
		filesCnt := binary.LittleEndian.Uint64(data)
		data = data[8:]
		incSz := binary.LittleEndian.Uint64(data)
		data = data[8:]

		var dir []byte
		dir, data, err = tl.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dir name: %w", err)
		}
		dirStr := string(dir)

		t.FilesCount = &filesCnt
		t.IncludedSize = &incSz
		t.DirName = &dirStr
	}

	t.DownloadedSize = binary.LittleEndian.Uint64(data)
	data = data[8:]
	t.AddedAt = binary.LittleEndian.Uint32(data)
	data = data[4:]

	var rootDir []byte
	rootDir, data, err = tl.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse description: %w", err)
	}
	t.RootDir = string(rootDir)

	t.ActiveDownload = binary.LittleEndian.Uint32(data) == BoolTrue
	data = data[4:]
	t.ActiveUpload = binary.LittleEndian.Uint32(data) == BoolTrue
	data = data[4:]
	t.Completed = binary.LittleEndian.Uint32(data) == BoolTrue
	data = data[4:]

	t.DownloadSpeed = math.Float64frombits(binary.LittleEndian.Uint64(data))
	data = data[8:]
	t.UploadSpeed = math.Float64frombits(binary.LittleEndian.Uint64(data))
	data = data[8:]

	if t.Flags&4 != 0 {
		var fatalErr []byte
		fatalErr, data, err = tl.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dir name: %w", err)
		}
		fatalErrStr := string(fatalErr)

		t.FatalError = &fatalErrStr
	}

	return data, nil
}

func (t *Torrent) Serialize(buf *bytes.Buffer) error {
	return fmt.Errorf("not implemented")
}

func (d *Double) Parse(data []byte) (_ []byte, err error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("too short data")
	}
	d.Value = math.Float64frombits(binary.LittleEndian.Uint64(data))
	return data[8:], nil
}

func (d *Double) Serialize(buf *bytes.Buffer) error {
	var data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, math.Float64bits(d.Value))
	buf.Write(data)
	return nil
}

func (d *MetaFile) Parse(data []byte) (_ []byte, err error) {
	// torrent_file#6a7181e0 flags:(## 32) info_boc_size:uint32
	//   root_proof_boc_size:flags.0?uint32
	//   info_boc:(info_boc_size * [uint8])
	//   root_proof_boc:flags.0?(root_proof_boc_size * [uint8])
	//   header:flags.1?TorrentHeader = TorrentMeta;

	if len(data) < 8 {
		return nil, fmt.Errorf("too short data")
	}

	flags := binary.LittleEndian.Uint32(data)
	data = data[4:]
	infoSz := binary.LittleEndian.Uint32(data)
	data = data[4:]

	var rootProofSz uint32
	if flags&1 > 0 {
		// skip
		rootProofSz = binary.LittleEndian.Uint32(data)
		data = data[4:]
	}
	if uint32(len(data)) < infoSz {
		return nil, fmt.Errorf("too short info")
	}

	info, err := cell.FromBOC(data[:infoSz])
	if err != nil {
		return nil, fmt.Errorf("failed to load info cell: %w", err)
	}
	data = data[infoSz:]

	err = tlb.LoadFromCell(&d.Info, info.BeginParse())
	if err != nil {
		return nil, fmt.Errorf("failed to parse info: %w", err)
	}

	if flags&1 > 0 {
		// skip
		if len(data) < int(rootProofSz) {
			return nil, fmt.Errorf("invalid root proof")
		}

		proof, err := cell.FromBOC(data[:rootProofSz])
		if err != nil {
			return nil, fmt.Errorf("failed to load info cell: %w", err)
		}
		data = data[rootProofSz:]
		d.RootProof = proof
	}

	if flags&2 > 0 {
		var hdr storage.TorrentHeader
		_, err = tl.Parse(&hdr, data, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load header: %w", err)
		}
		d.Header = &hdr
	}

	d.Hash = info.Hash()

	return nil, nil
}

func (d *MetaFile) Serialize(buf *bytes.Buffer) (err error) {
	var flags uint32

	var proof, info, header []byte
	infoCell, err := tlb.ToCell(d.Info)
	if err != nil {
		return fmt.Errorf("failed to serialize info to cell: %w", err)
	}
	info = infoCell.ToBOC()

	if d.RootProof != nil {
		flags |= 1
		proof = d.RootProof.ToBOC()
	}
	if d.Header != nil {
		flags |= 2
		header, err = tl.Serialize(d.Header, false)
		if err != nil {
			return fmt.Errorf("failed to serialize header: %w", err)
		}
	}

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, flags)
	binary.LittleEndian.PutUint32(data[4:], uint32(len(info)))
	if d.RootProof != nil {
		tmp := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(len(proof)))
		data = append(data, tmp...)
	}
	data = append(data, info...)
	data = append(data, proof...)
	data = append(data, header...)
	buf.Write(data)
	return nil
}
