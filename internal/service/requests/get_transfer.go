package requests

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GetTransferRequest struct {
	ID uuid.UUID
}

func NewGetTransferRequest(r *http.Request) (GetTransferRequest, error) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	
	if err != nil {
		return GetTransferRequest{}, errors.Wrap(err, "invalid transfer id")
	}

	return GetTransferRequest{ID: id}, nil
}