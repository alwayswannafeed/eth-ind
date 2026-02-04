package config

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

// IndexerConfig - структура, яку ми будемо використовувати в коді
type IndexerConfig struct {
	RPCSyncUrl   string
	RPCHistUrl   string
	Contract     common.Address
	StartDepth   uint64
	SyncInterval time.Duration
}

// Indexerer - інтерфейс, який каже: "я вмію віддавати конфіг індексатора"
type Indexerer interface {
	Indexer() IndexerConfig
}

// NewIndexerer повертає реалізацію логіки читання конфігу
func NewIndexerer(getter kv.Getter) Indexerer {
	return &indexer{
		getter: getter,
	}
}

type indexer struct {
	getter kv.Getter
	once   comfig.Once
}

func (i *indexer) Indexer() IndexerConfig {
	return i.once.Do(func() interface{} {
		var raw struct {
			RPCSyncUrl   string        `fig:"rpc_sync_url,required"`
			RPCHistUrl   string        `fig:"rpc_hist_url,required"`
			Contract     string        `fig:"contract_address,required"`
			StartDepth   uint64        `fig:"start_depth"`
			SyncInterval time.Duration `fig:"sync_interval"`
		}

		err := figure.
			Out(&raw).
			From(kv.MustGetStringMap(i.getter, "indexer")).
			Please()
		if err != nil {
			panic(err)
		}

		if raw.StartDepth == 0 {
			raw.StartDepth = 10000
		}
		if raw.SyncInterval == 0 {
			raw.SyncInterval = 10 * time.Second
		}

		return IndexerConfig{
			RPCSyncUrl:   raw.RPCSyncUrl,
			RPCHistUrl:   raw.RPCHistUrl,
			Contract:     common.HexToAddress(raw.Contract),
			StartDepth:   raw.StartDepth,
			SyncInterval: raw.SyncInterval,
		}
	}).(IndexerConfig)
}