package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alwayswannafeed/eth-ind/internal/data"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

type TransferRequest struct {
	Filters struct {
		Sender      *string    `json:"sender"`
		Receiver    *string    `json:"receiver"`
		Participant *string    `json:"participant"`
		TimeFrom    *time.Time `json:"time_from"`
		TimeTo      *time.Time `json:"time_to"`
	} `json:"filters"`
	Page struct {
		Limit  uint64 `json:"limit"`
		Cursor uint64 `json:"cursor"`
		Order  string `json:"order"`
	} `json:"page"`
}

func GetTransfers(w http.ResponseWriter, r *http.Request) {
	req := TransferRequest{}
	req.Page.Limit = 15
	req.Page.Order = "desc"

	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Log(r).WithError(err).Error("failed to decode request body")
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
	}

	selector := data.TransferSelector{
		Sender:      req.Filters.Sender,
		Receiver:    req.Filters.Receiver,
		Participant: req.Filters.Participant,
		TimeFrom:    req.Filters.TimeFrom,
		TimeTo:      req.Filters.TimeTo,
		PageParams: data.PageParams{
			Limit:  req.Page.Limit,
			Cursor: req.Page.Cursor,
			Order:  req.Page.Order,
		},
	}

	transfers, err := Storage(r).Transfers().Select(selector)
	if err != nil {
		Log(r).WithError(err).Error("failed to select transfers")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	ape.Render(w, data.NewTransferListResponse(transfers))
}