package gostorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tonutils/torrent-client/core/client"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"github.com/xssnick/tonutils-storage-provider/pkg/contract"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"
)

func (c *Client) FetchProviderContract(ctx context.Context, torrentHash []byte, owner *address.Address) (*client.ProviderContractData, error) {
	t := c.storage.GetTorrent(torrentHash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}
	if t.Info == nil {
		return nil, fmt.Errorf("info is not downloaded")
	}

	addr, _, _, err := contract.PrepareV1DeployData(torrentHash, t.Info.RootHash, t.Info.FileSize, t.Info.PieceSize, owner, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to calc contract addr: %w", err)
	}

	master, err := c.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch master block: %w", err)
	}

	list, balance, err := contract.GetProvidersV1(ctx, c.api, master, addr)
	if err != nil {
		if errors.Is(err, contract.ErrNotDeployed) {
			return nil, contract.ErrNotDeployed
		}
		return nil, fmt.Errorf("failed to fetch providers list: %w", err)
	}

	return &client.ProviderContractData{
		Size:      t.Info.FileSize,
		Address:   addr,
		Providers: list,
		Balance:   balance,
	}, nil
}

func (c *Client) BuildAddProviderTransaction(ctx context.Context, torrentHash []byte, owner *address.Address, providers []client.NewProviderData) (addr *address.Address, bodyData, stateInit []byte, err error) {
	t := c.storage.GetTorrent(torrentHash)
	if t == nil {
		return nil, nil, nil, fmt.Errorf("torrent is not found")
	}
	if t.Info == nil {
		return nil, nil, nil, fmt.Errorf("info is not downloaded")
	}

	var prs []contract.ProviderV1
	for _, p := range providers {
		prs = append(prs, contract.ProviderV1{
			Address:       p.Address,
			MaxSpan:       p.MaxSpan,
			PricePerMBDay: p.PricePerMBDay,
		})
	}

	addr, si, body, err := contract.PrepareV1DeployData(torrentHash, t.Info.RootHash, t.Info.FileSize, t.Info.PieceSize, owner, prs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to prepare contract data: %w", err)
	}

	siCell, err := tlb.ToCell(si)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("serialize state init: %w", err)
	}
	return addr, body.ToBOC(), siCell.ToBOC(), nil
}

func (c *Client) BuildWithdrawalTransaction(torrentHash []byte, owner *address.Address) (addr *address.Address, bodyData []byte, err error) {
	t := c.storage.GetTorrent(torrentHash)
	if t == nil {
		return nil, nil, fmt.Errorf("torrent is not found")
	}
	if t.Info == nil {
		return nil, nil, fmt.Errorf("info is not downloaded")
	}

	addr, body, err := contract.PrepareWithdrawalRequest(torrentHash, t.Info.RootHash, t.Info.FileSize, t.Info.PieceSize, owner)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare contract data: %w", err)
	}

	return addr, body.ToBOC(), nil
}

func (c *Client) FetchProviderRates(ctx context.Context, torrentHash, providerKey []byte) (*client.ProviderRates, error) {
	t := c.storage.GetTorrent(torrentHash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}
	if t.Info == nil {
		return nil, fmt.Errorf("info is not downloaded")
	}

	rates, err := c.provider.GetStorageRates(ctx, providerKey, t.Info.FileSize)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "value is not found"):
			return nil, errors.New("provider is not found")
		case strings.Contains(err.Error(), "context deadline exceeded"):
			return nil, errors.New("provider is not respond in a given time")
		}
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	return &client.ProviderRates{
		Available:        rates.Available,
		RatePerMBDay:     tlb.FromNanoTON(new(big.Int).SetBytes(rates.RatePerMBDay)),
		MinBounty:        tlb.FromNanoTON(new(big.Int).SetBytes(rates.MinBounty)),
		SpaceAvailableMB: rates.SpaceAvailableMB,
		MinSpan:          rates.MinSpan,
		MaxSpan:          rates.MaxSpan,
		Size:             t.Info.FileSize,
	}, nil
}

func (c *Client) RequestProviderStorageInfo(ctx context.Context, torrentHash, providerKey []byte, owner *address.Address) (*client.ProviderStorageInfo, error) {
	t := c.storage.GetTorrent(torrentHash)
	if t == nil {
		return nil, fmt.Errorf("torrent is not found")
	}
	if t.Info == nil {
		return nil, fmt.Errorf("info is not downloaded")
	}

	addr, _, _, err := contract.PrepareV1DeployData(torrentHash, t.Info.RootHash, t.Info.FileSize, t.Info.PieceSize, owner, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to calc contract addr: %w", err)
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	var tm time.Time
	v := c.infoCache[addr.String()]
	if v != nil {
		tm = v.FetchedAt
	} else {
		v = &client.ProviderStorageInfo{
			Status: "connecting...",
		}
		c.infoCache[addr.String()] = v
	}

	// run job if result is older than 10 sec and no another active job
	if time.Since(tm) > 5*time.Second && (v.Context == nil || v.Context.Err() != nil) {
		var end func()
		v.Context, end = context.WithCancel(context.Background())

		go func() {
			defer end()

			proofByte := uint64(rand.Int63()) % t.Info.FileSize
			info, err := c.provider.RequestStorageInfo(ctx, providerKey, addr, proofByte)
			if err != nil {
				log.Println("failed to get storage info:", err)

				c.mx.Lock()
				c.infoCache[addr.String()] = &client.ProviderStorageInfo{
					Status:    "inactive",
					Reason:    err.Error(),
					FetchedAt: time.Now(),
				}
				c.mx.Unlock()
				return
			}

			json.NewEncoder(os.Stdout).Encode(info)

			progress, _ := new(big.Float).Quo(new(big.Float).SetUint64(info.Downloaded), new(big.Float).SetUint64(t.Info.FileSize)).Float64()

			if info.Status == "active" {
				proved := false
				// verify proof
				proof, err := cell.FromBOC(info.Proof)
				if err == nil {
					if proofData, err := cell.UnwrapProof(proof, t.Info.RootHash); err == nil {
						piece := uint32(proofByte / uint64(t.Info.PieceSize))
						pieces := uint32(t.Info.FileSize / uint64(t.Info.PieceSize))

						if err = checkProofBranch(proofData, piece, pieces); err == nil {
							info.Reason = fmt.Sprintf("Storage proof received just now")
							proved = true
						}
					} else {
						log.Println("failed to unwrap proof:", err)
					}
				}

				if !proved {
					info.Status = "untrusted"
					info.Reason = "Incorrect proof received"
				}
			} else if info.Status == "downloading" {
				info.Reason = fmt.Sprintf("Progress: %.2f", progress*100) + "%"
			} else if info.Status == "resolving" {
				info.Reason = fmt.Sprintf("Provider is trying to find source to download bag")
			} else if info.Status == "warning-balance" {
				info.Reason = fmt.Sprintf("Not enough balance to store bag, please topup or it will be deleted soon")
			}

			c.mx.Lock()
			c.infoCache[addr.String()] = &client.ProviderStorageInfo{
				Status:    info.Status,
				Reason:    info.Reason,
				Progress:  progress * 100,
				FetchedAt: time.Now(),
			}
			c.mx.Unlock()
		}()
	}

	return v, nil
}

func checkProofBranch(proof *cell.Cell, piece, piecesNum uint32) error {
	if piece >= piecesNum {
		return fmt.Errorf("piece is out of range %d/%d", piece, piecesNum)
	}

	tree := proof.BeginParse()

	// calc tree depth
	depth := int(math.Log2(float64(piecesNum)))
	if piecesNum > uint32(math.Pow(2, float64(depth))) {
		// add 1 if pieces num is not exact log2
		depth++
	}

	// check bits from left to right and load branches
	for i := depth - 1; i >= 0; i-- {
		isLeft := piece&(1<<i) == 0

		b, err := tree.LoadRef()
		if err != nil {
			return err
		}

		if isLeft {
			tree = b
			continue
		}

		// we need right branch
		tree, err = tree.LoadRef()
		if err != nil {
			return err
		}
	}

	if tree.BitsLeft() != 256 {
		return fmt.Errorf("incorrect branch")
	}
	return nil
}
