package cmd

import "testing"

func TestParseEventArg(t *testing.T) {
	testCases := []struct {
		input             string
		expectedEventType string
		expectedGoType    string
		expectedErr       error
	}{
		{"a:b", "a", "b", nil},
		{"a:b:c", "a:b", "c", nil},
		{
			"Microsoft.Storage.BlobCreated:github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.StorageBlobCreatedEventData",
			"Microsoft.Storage.BlobCreated",
			"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.StorageBlobCreatedEventData",
			nil,
		},
		{
			"Microsoft.Storage.BlobCreated",
			"Microsoft.Storage.BlobCreated",
			wellKnownEvents["Microsoft.Storage.BlobCreated"],
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			eventType, goType, err := parseEventArg(tc.input)
			if err != tc.expectedErr {
				t.Logf("\ngot:\n\t%v\nwant:\n\t%v", err, tc.expectedErr)
				t.Fail()
			}

			if eventType != tc.expectedEventType {
				t.Logf("\ngot:\n\t%s\nwant:\n\t%s", eventType, tc.expectedEventType)
				t.Fail()
			}

			if goType != tc.expectedGoType {
				t.Logf("\ngot:\n\t%s\nwant:\n\t%s", eventType, tc.expectedGoType)
				t.Fail()
			}
		})
	}
}
