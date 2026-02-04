package handlers

import (
	"net/http"
	"time"

	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	earliestTime, err := Storage(r).Transfers().GetEarliestBlockTime()
	if err != nil {
		Log(r).WithError(err).Error("failed to get earliest block time")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := StatusResponse{
		HistoryDate: earliestTime,
		CurrentTime: time.Now().UTC(),
	}

	ape.Render(w, response)
}