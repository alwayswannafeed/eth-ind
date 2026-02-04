package data

const (
	StateSyncKey     = "sync_block"
	StateHistKey     = "hist_block"
	StateSyncHashKey = "sync_block_hash"
)

type State struct {
	Key   string `db:"key" struct:"key"`
	Value string `db:"value" struct:"value"`
}

type StateQ interface {
	New() StateQ
	
	Get(key string) (*State, error)
	Upsert(key string, value string) error
}