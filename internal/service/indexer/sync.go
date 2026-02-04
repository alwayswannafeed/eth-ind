package indexer

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"gitlab.com/distributed_lab/logan/v3"
	
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

const syncBatchSize = 9 //RPC Free-limit

func (i *Indexer) runSync(ctx context.Context) {
	cfg := i.cfg.Indexer()
	ticker := time.NewTicker(cfg.SyncInterval)
	defer ticker.Stop()

	i.log.Info("Starting sync")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i.processSync(ctx)
		}
	}
}

func (i *Indexer) processSync(ctx context.Context) {
	lastBlockDB, lastHashDB, err := i.getSyncState()
	if err != nil {
		i.log.WithError(err).Error("Failed to get sync state")
		return
	}

	headHeader, err := i.rpcSync.HeaderByNumber(ctx, nil)
	if err != nil {
		i.log.WithError(err).Error("Failed to get network head")
		return
	}
	headBlock := headHeader.Number.Uint64()

	if lastBlockDB >= headBlock {
		return
	}

	// reorg check
	lastBlockRPC, err := i.rpcSync.HeaderByNumber(ctx, new(big.Int).SetUint64(lastBlockDB))
	if err != nil {
		i.log.WithError(err).Error("Failed to verify last block hash")
		return
	}

	if lastBlockRPC.Hash().String() != lastHashDB {
		i.log.WithFields(logan.F{
			"db_block": lastBlockDB,
			"db_hash":  lastHashDB,
			"rpc_hash": lastBlockRPC.Hash().String(),
		}).Warn("REORG DETECTED! Rolling back...")

		if err := i.rollbackOneBlock(lastBlockDB); err != nil {
			i.log.WithError(err).Error("Failed to rollback")
		}
		return
	}

	from := lastBlockDB + 1
	to := from + syncBatchSize - 1
	if to > headBlock {
		to = headBlock
	}

	headerTo, err := i.rpcSync.HeaderByNumber(ctx, new(big.Int).SetUint64(to))
	if err != nil {
		i.log.WithError(err).Error("Failed to get header for target block")
		return
	}
	targetHash := headerTo.Hash().String()

	newLastBlock, newLastHash, err := i.parseAndSaveWithState(ctx, i.rpcSync, from, to, data.StateSyncKey, to, targetHash)
	if err != nil {
		i.log.WithError(err).Error("Failed to parse and save batch")
		return
	}

	i.log.WithFields(logan.F{
		"from": from,
		"to":   newLastBlock,
		"hash": newLastHash,
	}).Info("Synced new blocks")
}

func (i *Indexer) getSyncState() (uint64, string, error) {
	blockItem, err := i.storage.State().Get(data.StateSyncKey)
	if err != nil {
		return 0, "", err
	}

	if blockItem == nil {
		return 0, "", nil 
	}
	
	blockNum, _ := strconv.ParseUint(blockItem.Value, 10, 64)
	hashItem, err := i.storage.State().Get(data.StateSyncHashKey)

	if err != nil {
		return 0, "", err
	}
	
	hashVal := ""
	if hashItem != nil {
		hashVal = hashItem.Value
	} else {
	}

	return blockNum, hashVal, nil
}

func (i *Indexer) rollbackOneBlock(badBlock uint64) error {
	prevBlock := badBlock - 1
	prevHeader, err := i.rpcSync.HeaderByNumber(context.Background(), new(big.Int).SetUint64(prevBlock))

	if err != nil {
		return err
	}
	prevHash := prevHeader.Hash().String()

	return i.storage.Transaction(func(s data.Storage) error {
		if err := s.Transfers().DeleteFromBlock(badBlock); err != nil {
			return err
		}

		if err := s.State().Upsert(data.StateSyncKey, strconv.FormatUint(prevBlock, 10)); err != nil {
			return err
		}
		
		if err := s.State().Upsert(data.StateSyncHashKey, prevHash); err != nil {
			return err
		}
		
		return nil
	})
}