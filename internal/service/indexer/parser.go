package indexer

import (
	"context"
	"math/big"
	"strconv"
	"time"

	//"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"github.com/alwayswannafeed/eth-ind/internal/data"
	"github.com/alwayswannafeed/eth-ind/internal/erc20"
)

var transferEventSignature = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

func (i *Indexer) parseAndSaveWithState(
    ctx context.Context,
    client *ethclient.Client,
    from, to uint64,
    stateKey string,
    newStateVal uint64,
    newHashVal string,
) (uint64, string, error) {
    cfg := i.cfg.Indexer()
    contract, err := erc20.NewUSDC(cfg.Contract, client)

    if err != nil {
        return 0, "", errors.Wrap(err, "failed to create contract instance")
    }

    opts := &bind.FilterOpts{
        Start:   from,
        End:     &to, 
        Context: ctx,
    }

    iter, err := contract.FilterTransfer(opts, nil, nil)
    if err != nil {
        return 0, "", errors.Wrap(err, "failed to filter transfers")
    }

    var events []*erc20.USDCTransfer
    for iter.Next() {
        events = append(events, iter.Event)
    }

    if err := iter.Error(); err != nil {
        return 0, "", errors.Wrap(err, "iterator error")
    }

    var rawLogs []types.Log
    for _, e := range events {
        rawLogs = append(rawLogs, e.Raw)
    }

    blockTimestamps, err := i.getBlockTimestamps(ctx, client, rawLogs)
    if err != nil {
        return 0, "", errors.Wrap(err, "failed to get block timestamps")
    }

    var transfers []data.Transfer
    for _, event := range events {
        ts, ok := blockTimestamps[event.Raw.BlockNumber]
        if !ok {
            ts = time.Now() 
        }

        transfers = append(transfers, data.Transfer{
            ID:             uuid.New(),
            TxHash:         event.Raw.TxHash.Bytes(),
            BlockNumber:    event.Raw.BlockNumber,
            LogIndex:       uint32(event.Raw.Index),
            BlockHash:      event.Raw.BlockHash.Bytes(),
            BlockTimestamp: ts,
            FromAddr: event.From.Bytes(),
            ToAddr:   event.To.Bytes(),
            Amount:   event.Value.String(),
        })
    }

    err = i.storage.Transaction(func(s data.Storage) error {
        if len(transfers) > 0 {
            if err := s.Transfers().Insert(transfers...); err != nil {
                return errors.Wrap(err, "failed to insert transfers")
            }
        }
        if err := s.State().Upsert(stateKey, strconv.FormatUint(newStateVal, 10)); err != nil {
             return errors.Wrap(err, "failed to update state block")
        }
        if stateKey == data.StateSyncKey && newHashVal != "" {
            if err := s.State().Upsert(data.StateSyncHashKey, newHashVal); err != nil {
                return errors.Wrap(err, "failed to update state hash")
            }
        }
        return nil
    })
    
    if err != nil {
        return 0, "", err
    }

    finalHash := newHashVal
    if stateKey == data.StateSyncKey && finalHash == "" {
        header, err := client.HeaderByNumber(ctx, new(big.Int).SetUint64(to))
        if err != nil {
        } else {
            finalHash = header.Hash().String()
        }
    }

    return to, finalHash, nil
}

func (i *Indexer) getBlockTimestamps(ctx context.Context, client *ethclient.Client, logs []types.Log) (map[uint64]time.Time, error) {
	timestamps := make(map[uint64]time.Time)
	uniqueBlocks := make(map[uint64]struct{})

	for _, l := range logs {
		uniqueBlocks[l.BlockNumber] = struct{}{}
	}
	
	if len(uniqueBlocks) == 0 {
		return timestamps, nil
	}

	for blockNum := range uniqueBlocks {
		header, err := client.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			return nil, err
		}
		timestamps[blockNum] = time.Unix(int64(header.Time), 0)
	}

	return timestamps, nil
}