package resources

import (
	"encoding/hex"
	"time"

	"github.com/alwayswannafeed/eth-ind/internal/data"
	//"gitlab.com/distributed_lab/ape/problems"
)

const ResourceTypeTransfer = "transfers"

type Key struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type TransferAttributes struct {
	TxHash         string    `json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number"`
	LogIndex       uint32    `json:"log_index"`
	BlockHash      string    `json:"block_hash"`
	BlockTimestamp time.Time `json:"block_timestamp"`
	FromAddr       string    `json:"from_addr"`
	ToAddr         string    `json:"to_addr"`
	Amount         string    `json:"amount"`
}

type TransferResponse struct {
	Key
	Attributes TransferAttributes `json:"attributes"`
}

type TransferListResponse struct {
	Data  []TransferResponse `json:"data"`
}

func toHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func NewTransferResponse(t data.Transfer) TransferResponse {
	return TransferResponse{
		Key: Key{
			ID:   t.ID.String(),
			Type: ResourceTypeTransfer,
		},
		Attributes: TransferAttributes{
			TxHash:         toHex(t.TxHash),
			BlockNumber:    t.BlockNumber,
			LogIndex:       t.LogIndex,
			BlockHash:      toHex(t.BlockHash),
			BlockTimestamp: t.BlockTimestamp,
			FromAddr:       toHex(t.FromAddr),
			ToAddr:         toHex(t.ToAddr),
			Amount:         t.Amount,
		},
	}
}

func NewTransferListResponse(transfers []data.Transfer) TransferListResponse {
	list := make([]TransferResponse, len(transfers))
	for i, t := range transfers {
		list[i] = NewTransferResponse(t)
	}
	return TransferListResponse{
		Data: list,
	}
}