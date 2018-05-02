package eventgrid_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
	"github.com/gobuffalo/buffalo"
)

func TestSubscriptionValidationMiddleware(t *testing.T) {
	setFailure := func(c buffalo.Context) error {
		t.Fail()
		return c.Error(http.StatusInternalServerError, errors.New("`SubscriptionValidationMiddleware` did not detect request for validation"))
	}

	subject := eventgrid.SubscriptionValidationMiddleware(setFailure)

	req, err := http.NewRequest(http.MethodPost, "localhost", bytes.NewReader([]byte(`[{
	"id": "2d1781af-3a4c-4d7c-bd0c-e34b19da4e66",
	"topic": "/subscriptions/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
	"subject": "",
	"data": {
		"validationCode": "512d38b6-c7b8-40c8-89fe-f46f9e9622b6"
	},
	"eventType": "Microsoft.EventGrid.SubscriptionValidationEvent",
	"eventTime": "2018-01-25T22:12:19.4556811Z",
	"metadataVersion": "1",
	"dataVersion": "1"
}]`)))

	req.Header.Add("Aeg-Event-Type", "SubscriptionValidation")
	req.Header.Add("Content-Type", "application/json")

	ctx := NewMockContext(req)

	err = subject(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	type SubscriptionValidationResponse struct {
		ResponseCode string `json:"validationResponse,omitempty"`
	}

	var seenResponse SubscriptionValidationResponse
	dec := json.NewDecoder(ctx.Response().(*MockResponseWriter).Body())

	err = dec.Decode(&seenResponse)
	if err != nil {
		t.Error(err)
		return
	}

	if want := "512d38b6-c7b8-40c8-89fe-f46f9e9622b6"; seenResponse.ResponseCode != want {
		t.Logf("got: %s want: %s", seenResponse.ResponseCode, want)
		t.Fail()
	}
}
