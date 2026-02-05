package requests

import (
	"encoding/json"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
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

func NewListTransfersRequest(r *http.Request) (TransferRequest, error) {
	var request TransferRequest

	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return request, errors.Wrap(err, "failed to decode request body")
		}
	}

	if request.Page.Limit == 0 {
		request.Page.Limit = 15
	}
	if request.Page.Order == "" {
		request.Page.Order = "desc"
	}

	return request, request.validate()
}

func (r *TransferRequest) validate() error {
	return validation.Errors{
		"page/limit": validation.Validate(&r.Page.Limit, validation.Max(uint64(100))),
		"page/order": validation.Validate(&r.Page.Order, validation.In("asc", "desc")), 
	}.Filter()
}