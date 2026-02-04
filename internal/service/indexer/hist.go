package indexer

import (
	"context"
	"strconv"
	"time"

	"gitlab.com/distributed_lab/logan/v3"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

const histBatchSize = 9 //RPC free-limit

func (i *Indexer) runHist(ctx context.Context) {
	i.log.Info("Starting hist")
	for {
		if ctx.Err() != nil {
			return
		}

		finished, err := i.processHist(ctx)
		if err != nil {
			i.log.WithError(err).Error("Hist error, retrying in 5s...")
			time.Sleep(5 * time.Second)
			continue
		}

		if finished {
			i.log.Info("History fully synced. Hist stopping.")
			return
		}
		
	}
}

func (i *Indexer) processHist(ctx context.Context) (bool, error) {
	stateItem, err := i.storage.State().Get(data.StateHistKey)
	if err != nil {
		return false, err
	}
	if stateItem == nil {
		return false, nil
	}

	currentBlock, _ := strconv.ParseUint(stateItem.Value, 10, 64)

	if currentBlock <= 0 {
		return true, nil
	}

	to := currentBlock
	var from uint64
	if to > histBatchSize {
		from = to - histBatchSize
	} else {
		from = 0
	}

	nextStateVal := uint64(0)
	if from > 0 {
		nextStateVal = from - 1
	}

	_, _, err = i.parseAndSaveWithState(ctx, i.rpcHist, from, to, data.StateHistKey, nextStateVal, "") 
	if err != nil {
		return false, err
	}

	i.log.WithFields(logan.F{
		"range_from": from,
		"range_to":   to,
		"progress":   currentBlock,
	}).Info("History batch indexed")

	if from == 0 {
		return true, nil
	}

	return false, nil
}