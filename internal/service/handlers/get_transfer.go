package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"github.com/alwayswannafeed/eth-ind/internal/service/requests"
	"github.com/alwayswannafeed/eth-ind/internal/service/resources"
)

func GetTransferByID(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewGetTransferRequest(r)
	if err != nil {
		Log(r).WithError(err).Warn("failed to parse transfer id")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	transfer, err := Storage(r).Transfers().GetByID(req.ID)
	if err != nil {
		Log(r).WithError(err).Error("failed to get transfer by id")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if transfer == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, resources.NewTransferResponse(*transfer))
}