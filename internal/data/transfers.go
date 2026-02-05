package data

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type Transfer struct {
	ID             uuid.UUID `db:"id"`
	TxHash         []byte    `db:"tx_hash"`
	BlockNumber    uint64    `db:"block_number"`
	LogIndex       uint32    `db:"log_index"`
	BlockHash      []byte    `db:"block_hash"`
	BlockTimestamp time.Time `db:"block_timestamp"`
	FromAddr       []byte    `db:"from_addr"`
	ToAddr         []byte    `db:"to_addr"`
	Amount         string    `db:"amount"`
}

type TransferSelector struct {
	Sender      *string
	Receiver    *string
	Participant *string
	TimeFrom    *time.Time
	TimeTo      *time.Time
	
	PageParams  PageParams
}

type PageParams struct {
	Limit  uint64
	Cursor uint64
	Order  string
}

func (p PageParams) ApplyTo(stmt squirrel.SelectBuilder, cursorColumn string) squirrel.SelectBuilder {
	if p.Limit == 0 {
		p.Limit = 15
	}
	stmt = stmt.Limit(p.Limit)

	if p.Order == "asc" {
		stmt = stmt.OrderBy(cursorColumn + " ASC")
	} else {
		stmt = stmt.OrderBy(cursorColumn + " DESC")
	}
	
	if p.Cursor != 0 {
		if p.Order == "asc" {
			stmt = stmt.Where(squirrel.Gt{cursorColumn: p.Cursor})
		} else {
			stmt = stmt.Where(squirrel.Lt{cursorColumn: p.Cursor})
		}
	}

	return stmt
}

type TransfersQ interface {
	New() TransfersQ
	Insert(transfers ...Transfer) error
	Select(selector TransferSelector) ([]Transfer, error)
	GetByID(id uuid.UUID) (*Transfer, error)
	DeleteFromBlock(blockNumber uint64) error
	GetEarliestBlockTime() (*time.Time, error)
}