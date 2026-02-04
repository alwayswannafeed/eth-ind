package pg

import (

	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)
type storage struct {
	db  *pgdb.DB
	sql squirrel.StatementBuilderType
}

func New(db *pgdb.DB) data.Storage {
	return &storage{
		db:  db,
		sql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *storage) Transfers() data.TransfersQ {
	return NewTransfersQ(s.db)
}

func (s *storage) State() data.StateQ {
	return NewStateQ(s.db)
}

func (s *storage) Transaction(fn func(data.Storage) error) error {
    return s.db.Transaction(func() error {
        clone := *s
        clone.db = s.db.Clone()
        return fn(&clone)
    })
}