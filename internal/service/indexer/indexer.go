package indexer

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	
	"github.com/alwayswannafeed/eth-ind/internal/config"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

type Indexer struct {
	cfg     config.Config
	log     *logan.Entry
	storage data.Storage

	rpcSync *ethclient.Client
	rpcHist *ethclient.Client
}

func NewIndexer(cfg config.Config, storage data.Storage) *Indexer {
	return &Indexer{
		cfg:     cfg,
		log:     cfg.Log().WithField("service", "indexer"),
		storage: storage,
	}
}

func (i *Indexer) Run(ctx context.Context) error {
	i.log.Info("Starting indexer service...")
	var err error
	cfg := i.cfg.Indexer()
	
	i.rpcSync, err = ethclient.Dial(cfg.RPCSyncUrl)
	if err != nil {
		return errors.Wrap(err, "failed to dial sync rpc")
	}
	defer i.rpcSync.Close()

	i.rpcHist, err = ethclient.Dial(cfg.RPCHistUrl)
	if err != nil {
		return errors.Wrap(err, "failed to dial hist rpc")
	}
	defer i.rpcHist.Close()

	if err := i.ensureStartState(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure start state")
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		i.runSync(ctx) 
	}()

	go func() {
		defer wg.Done()
		i.runHist(ctx) 
	}()

	wg.Wait()
	i.log.Info("Indexer service stopped")
	return nil
}

func (i *Indexer) ensureStartState(ctx context.Context) error {
	state, err := i.storage.State().Get(data.StateSyncKey)
	if err != nil {
		return errors.Wrap(err, "failed to get state from db")
	}

	if state != nil {
		i.log.WithFields(logan.F{
			"sync_point": state.Value,
		}).Info("Resuming indexing from existing state")
		return nil
	}

	i.log.Info("No state found. Initializing fresh start...")

	header, err := i.rpcSync.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get latest block header")
	}
	currentHead := header.Number.Uint64()

	startDepth := i.cfg.Indexer().StartDepth
	var startBlock uint64
	if currentHead > startDepth {
		startBlock = currentHead - startDepth
	} else {
		startBlock = 0
	}

	i.log.WithFields(logan.F{
		"network_head": currentHead,
		"start_block":  startBlock,
		"depth":        startDepth,
	}).Info("Calculated start point")

	err = i.storage.Transaction(func(s data.Storage) error {
		blockStr := new(big.Int).SetUint64(startBlock).String()
		
		if err := s.State().Upsert(data.StateSyncKey, blockStr); err != nil {
			return err
		}
		
		if err := s.State().Upsert(data.StateHistKey, blockStr); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "failed to initialize state in db")
	}

	i.log.Info("State initialized successfully")
	return nil
}