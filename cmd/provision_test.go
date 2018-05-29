package cmd

import (
	"testing"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Test_getDeploymentTemplate_links(t *testing.T) {
	testCases := []string{
		"https://aka.ms/buffalo-template",
		"http://aka.ms/buffalo-template",
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result, err := getDeploymentTemplate(tc)
			if err != nil {
				t.Error(err)
			}

			if result.Template != nil {
				t.Log("unexpected value present in template")
				t.Fail()
			}

			if result.TemplateLink == nil {
				t.Log("unexpected nil template link")
				t.Fail()
				return
			}

			if result.TemplateLink.URI == nil {
				t.Log("unexpected nil uri")
				t.Fail()
				return
			}

			if got := *result.TemplateLink.URI; !strings.EqualFold(got, tc) {
				t.Logf("got:\n\t%q\nwant:\n\t%q", got, tc)
				t.Fail()
				return
			}
		})
	}
}

func Test_getDeploymentTemplate_localFiles(t *testing.T) {
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

			result, err := getDeploymentTemplate(tc)
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
