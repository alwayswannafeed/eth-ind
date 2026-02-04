package indexer

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"github.com/alwayswannafeed/eth-ind/internal/data"
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
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{cfg.Contract},
		Topics: [][]common.Hash{
			{transferEventSignature}, 
		},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to filter logs")
	}

	blockTimestamps, err := i.getBlockTimestamps(ctx, client, logs)
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to get block timestamps")
	}

	var transfers []data.Transfer
	
	for _, vLog := range logs {
		if len(vLog.Topics) < 3 {
			continue
		}

		fromAddr := common.BytesToAddress(vLog.Topics[1].Bytes()).Bytes()
		toAddr := common.BytesToAddress(vLog.Topics[2].Bytes()).Bytes()
		amount := new(big.Int).SetBytes(vLog.Data)
		ts, ok := blockTimestamps[vLog.BlockNumber]

		if !ok {
			i.log.WithField("block", vLog.BlockNumber).Warn("Timestamp missing for block, using Now")
			ts = time.Now()
		}

		transfers = append(transfers, data.Transfer{
			ID:             uuid.New(),
			TxHash:         vLog.TxHash.Bytes(),
			BlockNumber:    vLog.BlockNumber,
			LogIndex:       uint32(vLog.Index),
			BlockHash:      vLog.BlockHash.Bytes(),
			BlockTimestamp: ts,
			FromAddr:       fromAddr,
			ToAddr:         toAddr,
			Amount:         amount.String(),
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
			i.log.WithError(err).Warn("Failed to fetch final block hash")
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