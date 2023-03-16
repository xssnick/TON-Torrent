package client

import (
	"context"
	"fmt"
	"github.com/xssnick/tonutils-go/tl"
)

type ADNL interface {
	QueryADNL(ctx context.Context, payload, response tl.Serializable) error
}

type StorageClient struct {
	client ADNL
}

func NewStorageClient(client ADNL) *StorageClient {
	return &StorageClient{
		client: client,
	}
}

func (s *StorageClient) GetTorrents(ctx context.Context) (*TorrentsList, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetTorrents{}, &res)
	if err != nil {
		return nil, fmt.Errorf("faled to query torrents list: %w", err)
	}

	switch t := res.(type) {
	case TorrentsList:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("daemon err: %s", t.Message)
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
		return nil, fmt.Errorf("faled to query add by hash: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("daemon err: %s", t.Message)
	}
	return nil, fmt.Errorf("unexpected response")
}

func (s *StorageClient) GetTorrentFull(ctx context.Context, hash []byte) (*TorrentFull, error) {
	var res tl.Serializable
	err := s.client.QueryADNL(ctx, GetTorrentFull{
		Hash: hash,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("faled to query get torrent full: %w", err)
	}

	switch t := res.(type) {
	case TorrentFull:
		return &t, nil
	case DaemonError:
		return nil, fmt.Errorf("daemon err: %s", t.Message)
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
		return fmt.Errorf("faled to query remove torrent: %w", err)
	}

	switch t := res.(type) {
	case Success:
		return nil
	case DaemonError:
		return fmt.Errorf("daemon err: %s", t.Message)
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
		return fmt.Errorf("faled to query set active download torrent: %w", err)
	}

	switch t := res.(type) {
	case Success:
		var res tl.Serializable
		err = s.client.QueryADNL(ctx, SetActiveUpload{
			Hash:   hash,
			Active: active,
		}, &res)
		if err != nil {
			return fmt.Errorf("faled to query set active upload torrent: %w", err)
		}

		switch t := res.(type) {
		case Success:
			return nil
		case DaemonError:
			return fmt.Errorf("daemon err: %s", t.Message)
		}
		return fmt.Errorf("unexpected response")
	case DaemonError:
		return fmt.Errorf("daemon err: %s", t.Message)
	}
	return fmt.Errorf("unexpected response")
}
