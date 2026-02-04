package data

import (
	"encoding/hex"
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

type TransferResource struct {
	ID             string    `json:"id"`
	TxHash         string    `json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number"`
	LogIndex       uint32    `json:"log_index"`
	BlockHash      string    `json:"block_hash"`
	BlockTimestamp time.Time `json:"block_timestamp"`
	FromAddr       string    `json:"from_addr"`
	ToAddr         string    `json:"to_addr"`
	Amount         string    `json:"amount"`
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

type TransferListResponse struct {
	Data []TransferResource `json:"data"`
}

func NewTransferListResponse(list []Transfer) TransferListResponse {
	resources := make([]TransferResource, 0, len(list))

	for _, t := range list {
		resources = append(resources, TransferResource{
			ID:             t.ID.String(),
			TxHash:         toHex(t.TxHash),
			BlockNumber:    t.BlockNumber,
			LogIndex:       t.LogIndex,
			BlockHash:      toHex(t.BlockHash),
			BlockTimestamp: t.BlockTimestamp,
			FromAddr:       toHex(t.FromAddr),
			ToAddr:         toHex(t.ToAddr),
			Amount:         t.Amount,
		})
	}

	return TransferListResponse{
		Data: resources,
	}
}

func toHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

type TransfersQ interface {
	New() TransfersQ
	Insert(transfers ...Transfer) error
	Select(selector TransferSelector) ([]Transfer, error)
	GetByID(id uuid.UUID) (*Transfer, error)
	DeleteFromBlock(blockNumber uint64) error
	GetEarliestBlockTime() (*time.Time, error)
}