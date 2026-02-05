package handlers

import (
	"net/http"

	"github.com/alwayswannafeed/eth-ind/internal/service/requests"
	"github.com/alwayswannafeed/eth-ind/internal/data"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetTransfers(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewListTransfersRequest(r)
	if err != nil {
		Log(r).WithError(err).Warn("failed to parse request") 
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
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