package handlers

import (
	"time"
	"github.com/google/uuid"
)

type StatusResponse struct {
	HistoryDate *time.Time `json:"history_reached_date"` 
	CurrentTime time.Time  `json:"current_server_time"`
}

type TransferResponse struct {
	ID             uuid.UUID `json:"id"`
	TxHash         string    `json:"tx_hash"` // hex string
	BlockNumber    uint64    `json:"block_number"`
	BlockTimestamp time.Time `json:"block_timestamp"`
	FromAddr       string    `json:"from_addr"` // hex string
	ToAddr         string    `json:"to_addr"`   // hex string
	Amount         string    `json:"amount"`
}