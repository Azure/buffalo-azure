package eventgrid_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Azure/buffalo-azure/sdk/eventgrid"
	"github.com/gobuffalo/buffalo"
)

func TestSubscriptionValidationMiddleware_SubscriptionEvent(t *testing.T) {
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

func TestSubscriptionValidationMiddleware_OtherEvent(t *testing.T) {
	var called bool
	setCalled := func(c buffalo.Context) error {
		called = true
		return nil
	}

	subject := eventgrid.SubscriptionValidationMiddleware(setCalled)

	req, err := http.NewRequest(http.MethodPost, "localhost", bytes.NewReader([]byte(`[{
		"topic": "/subscriptions/{subscription-id}/resourceGroups/Storage/providers/Microsoft.Storage/storageAccounts/xstoretestaccount",
		"subject": "/blobServices/default/containers/oc2d2817345i200097container/blobs/oc2d2817345i20002296blob",
		"eventType": "Microsoft.Storage.BlobCreated",
		"eventTime": "2017-06-26T18:41:00.9584103Z",
		"id": "831e1650-001e-001b-66ab-eeb76e069631",
		"data": {
			"api": "PutBlockList",
			"clientRequestId": "6d79dbfb-0e37-4fc4-981f-442c9ca65760",
			"requestId": "831e1650-001e-001b-66ab-eeb76e000000",
			"eTag": "0x8D4BCC2E4835CD0",
			"contentType": "application/octet-stream",
			"contentLength": 524288,
			"blobType": "BlockBlob",
			"url": "https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/oc2d2817345i20002296blob",
			"sequencer": "00000000000004420000000000028963",
			"storageDiagnostics": {
			"batchId": "b68529f3-68cd-4744-baa4-3c0498ec19f0"
			}
		},
		"dataVersion": "",
		"metadataVersion": "1"
	}]`)))

	req.Header.Add("Content-Type", "application/json")

	ctx := NewMockContext(req)

	err = subject(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	respBodyLength, err := io.Copy(ioutil.Discard, ctx.Response().(*MockResponseWriter).Body())
	if err != nil {
		t.Error(err)
		return
	}

	if respBodyLength != 0 {
		t.Logf("expected an empty body")
		t.Fail()
	}

	if called == false {
		t.Logf("handler pass through never occured")
		t.Fail()
	}
}
