package handlers

import (
	//"encoding/hex"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"github.com/alwayswannafeed/eth-ind/internal/data"
)

func GetTransferByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		Log(r).WithError(err).Warn("failed to parse transfer id")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	transfer, err := Storage(r).Transfers().GetByID(id)
	if err != nil {
		Log(r).WithError(err).Error("failed to get transfer by id")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if transfer == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, data.NewTransferResponse(*transfer))
}