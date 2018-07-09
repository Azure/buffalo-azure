package common_test

import (
	"testing"

	. "github.com/Azure/buffalo-azure/generators/common"
)

func TestImportBag_AddImport_Unique(t *testing.T) {
	testCases := [][]PackagePath{
		{
			"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid",
		},
		{
			"github.com/Azure/buffalo-azure/generators/common",
			"github.com/Azure/buffalo-azure/sdk/eventgrid",
		},
		{
			"fmt",
			"github.com/Azure/buffalo-azure/generators/common",
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {

			seen := map[PackageSpecifier]struct{}{}
			subject := NewImportBag()

			for _, pkgPath := range tc {
				spec := subject.AddImport(pkgPath)
				if _, ok := seen[spec]; ok {
					t.Logf("duplicate spec returned")
					t.Fail()
				}
			}
		})
	}
}

func TestImportBag_List(t *testing.T) {
	testCases := []struct {
		inputs   []PackagePath
		expected []string
	}{
		{
			inputs: []PackagePath{
				"github.com/Azure/buffalo-azure/generators/common",
				"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid",
			},
			expected: []string{
				`"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"`,
				`"github.com/Azure/buffalo-azure/generators/common"`,
			},
		},
		{
			inputs: []PackagePath{
				"github.com/Azure/buffalo-azure/sdk/eventgrid",
				"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid",
			},
			expected: []string{
				`eventgrid1 "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"`,
				`"github.com/Azure/buffalo-azure/sdk/eventgrid"`,
			},
		},
		{
			inputs: []PackagePath{
				"github.com/Azure/buffalo-azure/sdk/eventgrid",
				"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid",
				"github.com/Azure/buffalo-azure/generators/eventgrid",
			},
			expected: []string{
				`eventgrid1 "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"`,
				`eventgrid2 "github.com/Azure/buffalo-azure/generators/eventgrid"`,
				`"github.com/Azure/buffalo-azure/sdk/eventgrid"`,
			},
		},
		{
			inputs: []PackagePath{
				"github.com/Azure/buffalo-azure/sdk/eventgrid",
				"github.com/Azure/buffalo-azure/sdk/eventgrid",
			},
			expected: []string{
				`"github.com/Azure/buffalo-azure/sdk/eventgrid"`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			subject := NewImportBag()
			for _, i := range tc.inputs {
				subject.AddImport(i)
			}

			got := subject.List()

			if len(got) != len(tc.expected) {
				t.Logf("got: %d imports want: %d imports", len(got), len(tc.expected))
				t.Fail()
			}

			for i, val := range tc.expected {
				if got[i] != val {
					t.Logf("at %d\n\tgot:  %s\n\twant: %s", i, got[i], val)
					t.Fail()
				}
			}
		})
	}
}

func TestImportBag_NewImportBagFromFile(t *testing.T) {
	testCases := []struct {
		inputFile string
		expected  []string
	}{
		{"./testdata/main.go", []string{`"fmt"`, `_ "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"`}},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			subject, err := NewImportBagFromFile(tc.inputFile)
			if err != nil {
				t.Error(err)
				return
			}

			got := subject.List()

			for i, imp := range tc.expected {
				if imp != got[i] {
					t.Logf("at: %d\n\tgot:  %s\n\twant: %s", i, got[i], imp)
					t.Fail()
				}
			}

			if len(got) != len(tc.expected) {
				t.Logf("got: %d imports want: %d imports", len(got), len(tc.expected))
				t.Fail()
			}
		})
	}
}
