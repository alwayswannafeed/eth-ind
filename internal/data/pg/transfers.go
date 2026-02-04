package pg

import (
	"database/sql"
	"time"
	"encoding/hex"
    "strings"
    "github.com/google/uuid"
	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

type transfersQ struct {
	db  *pgdb.DB
	sql squirrel.StatementBuilderType
}

func NewTransfersQ(db *pgdb.DB) data.TransfersQ {
	return &transfersQ{
		db:  db,
		sql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (q *transfersQ) New() data.TransfersQ {
	return NewTransfersQ(q.db)
}

func (q *transfersQ) Insert(transfers ...data.Transfer) error {
	if len(transfers) == 0 {
		return nil
	}

	stmt := q.sql.Insert("transfers").Columns(
		"id", "tx_hash", "block_number", "log_index", 
		"block_hash", "block_timestamp", 
		"from_addr", "to_addr", "amount",
	)

	for _, t := range transfers {
		stmt = stmt.Values(
			t.ID, t.TxHash, t.BlockNumber, t.LogIndex, 
			t.BlockHash, t.BlockTimestamp, 
			t.FromAddr, t.ToAddr, t.Amount,
		)
	}

	stmt = stmt.Suffix("ON CONFLICT (tx_hash, log_index) DO NOTHING")
	return q.db.Exec(stmt)
}

func (q *transfersQ) DeleteFromBlock(blockNumber uint64) error {
	stmt := q.sql.Delete("transfers").
		Where(squirrel.GtOrEq{"block_number": blockNumber})

	return q.db.Exec(stmt)
}

func (q *transfersQ) GetByID(id uuid.UUID) (*data.Transfer, error) {
	var result data.Transfer

	stmt := q.sql.Select("*").From("transfers").Where(squirrel.Eq{"id": id})

	err := q.db.Get(&result, stmt)

	if err == sql.ErrNoRows {
		return nil, nil 
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (q *transfersQ) Select(selector data.TransferSelector) ([]data.Transfer, error) {
	stmt := q.sql.Select("*").From("transfers")
	if selector.Sender != nil {
		addrBytes, err := decodeHexAddress(*selector.Sender)
		if err != nil {
			return nil, err
		}
		stmt = stmt.Where(squirrel.Eq{"from_addr": addrBytes})
	}

	if selector.Receiver != nil {
		addrBytes, err := decodeHexAddress(*selector.Receiver)
		if err != nil {
			return nil, err
		}
		stmt = stmt.Where(squirrel.Eq{"to_addr": addrBytes})
	}

	if selector.Participant != nil {
		addrBytes, err := decodeHexAddress(*selector.Participant)
		if err != nil {
			return nil, err
		}
		stmt = stmt.Where(squirrel.Or{
			squirrel.Eq{"from_addr": addrBytes},
			squirrel.Eq{"to_addr": addrBytes},
		})
	}

	if selector.TimeFrom != nil {
		stmt = stmt.Where(squirrel.GtOrEq{"block_timestamp": *selector.TimeFrom})
	}
	if selector.TimeTo != nil {
		stmt = stmt.Where(squirrel.LtOrEq{"block_timestamp": *selector.TimeTo})
	}

	stmt = selector.PageParams.ApplyTo(stmt, "block_timestamp")

	sqlQuery, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	var dest []data.Transfer
	err = q.db.Select(&dest, squirrel.Expr(sqlQuery, args...))
	
	return dest, err
}

func (q *transfersQ) GetEarliestBlockTime() (*time.Time, error) {
	stmt := q.sql.Select("MIN(block_timestamp)").From("transfers")
	var t *time.Time
	err := q.db.Get(&t, stmt)
	return t, err
}

func decodeHexAddress(input string) ([]byte, error) {
	clean := strings.TrimPrefix(input, "0x")
	return hex.DecodeString(clean)
}