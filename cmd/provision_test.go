package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
)

func init() {
	var err error
	environment, err = azure.EnvironmentFromName(provisionConfig.GetString(EnvironmentName))
	if err != nil {
		environment = azure.PublicCloud
	}

	log.SetOutput(ioutil.Discard)
}

func Test_getAuthorizer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subscriptionID := provisionConfig.GetString(SubscriptionName)
	clientID := provisionConfig.GetString(ClientSecretName)
	clientSecret := provisionConfig.GetString(ClientSecretName)
	tenantID := provisionConfig.GetString(TenantIDName)

	if tenantID == "" || subscriptionID == "" {
		// If you don't want to tinker with the environment, you can pass these in as command-line arguments
		// to the `go test` command, the same way you would have to call the azure provision command.
		t.Skip("test environment not configured with a tenant or subscription")
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go t.Run("service principal, no tenant inference", func(t *testing.T) {
		defer wg.Done()
		if clientID == "" || clientSecret == "" {
			t.Skip("test environment not configured with a service principal")
			return
		}

		if auth, err := getAuthorizer(ctx, subscriptionID, clientID, clientSecret, tenantID); err != nil {
			t.Error(err)
		} else if auth == nil {
			t.Log("auth unexpected nil in non error case")
			t.Fail()
		}
	})

	go t.Run("service principal, tenant inference", func(t *testing.T) {
		defer wg.Done()

		if clientID == "" || clientSecret == "" {
			t.Skip("test environment not configured with a service principal")
			return
		}

		if _, err := getAuthorizer(ctx, subscriptionID, clientID, clientSecret, ""); err == nil {
			// Is this failing because you've found a work around and implemented Service Principal tenant inference?
			// Awesome, change this test.
			// Otherwise, something is wrong that could cause us to mislead customers into thinking they can do tenant
			// inference.
			t.Log("tenant inference should fail when using a service principal")
			t.Fail()
		}
	})

	wg.Wait()
}

func Test_getDeploymentTemplate_links(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCases := []string{
		"https://aka.ms/buffalo-template",
		"http://aka.ms/buffalo-template",
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result, err := getDeploymentTemplate(ctx, tc)
			if err != nil {
				t.Error(err)
			}

			if result.Template == nil {
				t.Log("unexpected nil present in template")
				t.Fail()
			}

			if result.TemplateLink != nil {
				t.Log("unexpected value template link")
				t.Fail()
				return
			}
		})
	}
}

func Test_getDeploymentTemplate_localFiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCases := []string{
		"./testdata/template1.json",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			handle, err := os.Open(tc)
			if err != nil {
				t.Error(err)
				return
			}

			result, err := getDeploymentTemplate(ctx, tc)
			if err != nil {
				t.Error(err)
				return
			}

			if result.TemplateLink != nil {
				t.Log("unexpected value present in template link")
				t.Fail()
			}

			if result.Template == nil {
				t.Log("unexpected nil template")
				t.Fail()
				return
			}

			want, err := ioutil.ReadAll(handle)
			if err != nil {
				t.Error(err)
			}
			minimized := bytes.NewBuffer([]byte{})
			enc := json.NewEncoder(minimized)
			err = enc.Encode(json.RawMessage(want))
			if err != nil {
				t.Error(err)
				return
			}
			want = minimized.Bytes()
			want = []byte(strings.TrimSpace(string(want)))

			got, err := json.Marshal(result.Template)

			report := func(got, want []byte) string {
				shrink := func(target []byte, maxLength int) (retval []byte) {
					if len(target) > maxLength {
						retval = append(target[:maxLength/2], []byte("...")...)
						retval = append(retval, target[len(target)-maxLength/2:]...)
					} else {
						retval = target
					}
					return
				}

				const maxLength = 30

				gotLength := len(got)
				got = shrink(got, maxLength)

				wantLength := len(want)
				want = shrink(want, maxLength)

				return fmt.Sprintf("\ngot (len %d):\n\t%q\nwant (len %d):\n\t%q", gotLength, got, wantLength, want)
			}

			if len(want) == len(got) {
				for i, current := range want {
					if got[i] != current {
						t.Log(report(got, want))
						t.Fail()
						break
					}
				}
			} else {
				t.Log(report(got, want))
				t.Fail()
			}
		})
	}
}
