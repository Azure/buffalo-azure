package eventgrid

import (
	"encoding/json"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
)

// SubscriptionValidationRequest allows for easy unmarshaling of the first
// event sent by an Event Grid Topic.
type SubscriptionValidationRequest struct {
	ValidationCode uuid.UUID `json:"validationCode,omitempty"`
}

// ReceiveSubscriptionValidationRequest will
func ReceiveSubscriptionValidationRequest(c buffalo.Context, e Event) error {
	var svr SubscriptionValidationRequest
	err := json.Unmarshal(e.Data, &svr)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	type SubscriptionValidationResponse struct {
		ValidationCode uuid.UUID `json:"validationResponse,omitempty"`
	}

	c.Logger().Info("received validation request from: ", c.Request().RemoteAddr)

	enc := json.NewEncoder(c.Response())

	err = enc.Encode(&SubscriptionValidationResponse{svr.ValidationCode})
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
