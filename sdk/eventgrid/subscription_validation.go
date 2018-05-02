package eventgrid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
)

// SubscriptionValidationRequest allows for easy unmarshaling of the first
// event sent by an Event Grid Topic.
type SubscriptionValidationRequest struct {
	ValidationCode uuid.UUID `json:"validationCode,omitempty"`
}

// SubscriptionValidationMiddleware provides a `buffalo.Handler` which will triage all incoming requests
// to either submit it for event processing, or echo back the response the server expects to validate a
// subscription.
func SubscriptionValidationMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if typeHeader := c.Request().Header.Get("Aeg-Event-Type"); strings.EqualFold(typeHeader, "SubscriptionValidation") {
			var events []Event
			if err := c.Bind(&events); err != nil {
				return c.Error(http.StatusBadRequest, err)
			}

			if numEvents := len(events); numEvents != 1 {
				return c.Error(http.StatusBadRequest, fmt.Errorf("expected exactly 1 event, got %d", numEvents))
			}

			return ReceiveSubscriptionValidationRequest(c, events[0])
		}
		return next(c)
	}
}

// ReceiveSubscriptionValidationRequest will echo the ValidateCode sent in the request
// back to the Event Grid Topic seeking subscription validation.
func ReceiveSubscriptionValidationRequest(c buffalo.Context, e Event) error {
	var svr SubscriptionValidationRequest
	err := json.Unmarshal(e.Data, &svr)
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	type SubscriptionValidationResponse struct {
		ValidationCode uuid.UUID `json:"validationResponse,omitempty"`
	}

	if logger := c.Logger(); logger != nil {
		logger.Info("received validation request from: ", c.Request().RemoteAddr)
	}

	enc := json.NewEncoder(c.Response())

	err = enc.Encode(&SubscriptionValidationResponse{svr.ValidationCode})
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	c.Response().WriteHeader(http.StatusOK)
	return nil
}
