package pg

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

type stateQ struct {
	db  *pgdb.DB
	sql squirrel.StatementBuilderType
}

func NewStateQ(db *pgdb.DB) data.StateQ {
	return &stateQ{
		db:  db,
		sql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (q *stateQ) New() data.StateQ {
	return NewStateQ(q.db)
}

func (q *stateQ) Get(key string) (*data.State, error) {
	var result data.State
	
	stmt := q.sql.Select("key", "value").
		From("state").
		Where(squirrel.Eq{"key": key})

	err := q.db.Get(&result, stmt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (q *stateQ) Upsert(key string, value string) error {
	stmt := q.sql.Insert("state").
		Columns("key", "value").
		Values(key, value).
		Suffix("ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value")

	return q.db.Exec(stmt)
}