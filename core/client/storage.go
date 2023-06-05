package client

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tl"
	"log"
	"os"
)

type ADNL interface {
	QueryADNL(ctx context.Context, payload, response tl.Serializable) error
}

type StorageClient struct {
	client ADNL
}

func ConnectToStorageDaemon(addr string, dbPath string) (*StorageClient, error) {
	clientKey, err := os.ReadFile(dbPath + "/cli-keys/client")
	if err != nil {
		log.Println(dbPath+"/client read err:", err.Error())
		return nil, fmt.Errorf(dbPath+"/cli-keys/client read err: %w", err)
	}
	authKey := ed25519.NewKeyFromSeed(clientKey[4:])

	key, err := os.ReadFile(dbPath + "/cli-keys/server.pub")
	if err != nil {
		log.Println(dbPath+"/cli-keys/server.pub read err:", err.Error())
		return nil, fmt.Errorf(dbPath+"/cli-keys/server.pub read err: %w", err)

	}
	serverKey := base64.StdEncoding.EncodeToString(key[4:])

	pool := liteclient.NewConnectionPoolWithAuth(authKey)
	err = pool.AddConnection(context.Background(), addr, serverKey)
	if err != nil {
		log.Println("connect to daemon err:", err.Error())
		return nil, fmt.Errorf("connect to daemon err: %w", err)
	}

	return &StorageClient{
		client: pool,
	}, nil
}

func (s *StorageClient) GetTorrents(ctx context.Context) (*TorrentsList, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetTorrents{}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query torrents list: %w", err)
	}

	switch t := res.(type) {
	case TorrentsList:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) AddByHash(ctx context.Context, hash []byte, dir string) (*TorrentFull, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, AddByHash{
		Hash:          hash,
		RootDir:       dir,
		StartDownload: true,
		AllowUpload:   true,
		Priorities:    []any{PriorityActionAll{0}}, // download only header
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query add by hash: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) AddByMeta(ctx context.Context, meta []byte, dir string) (*TorrentFull, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, AddByMeta{
		Meta:          meta,
		RootDir:       dir,
		StartDownload: true,
		AllowUpload:   true,
		Priorities:    []any{PriorityActionAll{0}}, // download only header
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query add by meta: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) CreateTorrent(ctx context.Context, dir, description string) (*TorrentFull, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, CreateTorrent{
		Path:        dir,
		Description: description,
		AllowUpload: true,
		CopyInside:  false,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query create torrent: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) GetTorrentFull(ctx context.Context, hash []byte) (*TorrentFull, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetTorrentFull{
		Hash: hash,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query get torrent full: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) GetTorrentMeta(ctx context.Context, hash []byte) ([]byte, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetTorrentMeta{
		Hash: hash,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query get torrent meta: %w", err)
	}

	switch t := res.(type) {
	case TorrentMeta:
		return t.Meta, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) GetPeers(ctx context.Context, hash []byte) (*PeersList, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetPeers{
		Hash: hash,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query get peers: %w", err)
	}

	switch t := res.(type) {
	case PeersList:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) RemoveTorrent(ctx context.Context, hash []byte, withFiles bool) error {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, RemoveTorrent{
		Hash:        hash,
		RemoveFiles: withFiles,
	}, &res)
	if err != nil {
		return fmt.Errorf("failed to query remove torrent: %w", err)
	}

	switch t := res.(type) {
	case Success:
		return nil
	case DaemonError:
		return fmt.Errorf("%s", t.Message)
	}
	return fmt.Errorf("unexpected response")
}

func (s *StorageClient) SetActive(ctx context.Context, hash []byte, active bool) error {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, SetActiveDownload{
		Hash:   hash,
		Active: active,
	}, &res)
	if err != nil {
		return fmt.Errorf("failed to query set active download torrent: %w", err)
	}

	switch t := res.(type) {
	case Success:
		var res tl.Serializable
		err = s.client.QueryADNL(ctx, SetActiveUpload{
			Hash:   hash,
			Active: active,
		}, &res)
		if err != nil {
			return fmt.Errorf("failed to query set active upload torrent: %w", err)
		}

		switch t := res.(type) {
		case Success:
			return nil
		case DaemonError:
			return fmt.Errorf("%s", t.Message)
		}
		return fmt.Errorf("unexpected response")
	case DaemonError:
		return fmt.Errorf("%s", t.Message)
	}
	return fmt.Errorf("unexpected response")
}

func (s *StorageClient) SetFilesPriority(ctx context.Context, hash []byte, names []string, priority int32) error {
	for _, name := range names {
		err := s.SetFilePriority(ctx, hash, name, priority)
		if err != nil {
			return fmt.Errorf("failed to set priority for %s: %w", name, err)
		}
	}
	return nil
}

func (s *StorageClient) SetFilePriority(ctx context.Context, hash []byte, name string, priority int32) error {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, SetFilePriorityByName{
		Hash:     hash,
		Name:     name,
		Priority: priority,
	}, &res)
	if err != nil {
		return fmt.Errorf("failed to query set file priority by name: %w", err)
	}

	switch t := res.(type) {
	case PriorityStatusPending, PriorityStatusSet:
		return nil
	case DaemonError:
		log.Println("priority set err:", t.Message)
		return fmt.Errorf("%s", t.Message)
	}
	return fmt.Errorf("unexpected response")
}

func (s *StorageClient) GetSpeedLimits(ctx context.Context) (*SpeedLimits, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetSpeedLimits{
		Flags: 0b11,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to query get speed limits: %w", err)
	}

	switch t := res.(type) {
	case SpeedLimits:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("%s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) SetSpeedLimits(ctx context.Context, download, upload int64) error {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, SetSpeedLimits{
		Flags: 0b11,
		Download: &Double{
			Value: float64(download),
		},
		Upload: &Double{
			Value: float64(upload),
		},
	}, &res)
	if err != nil {
		return fmt.Errorf("failed to query get speed limits: %w", err)
	}

	switch t := res.(type) {
	case Success:
		return nil
	case DaemonError:
		return fmt.Errorf("%s", t.Message)
	}
	return fmt.Errorf("unexpected response")
}
