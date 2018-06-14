// Copyright © 2018 Microsoft Corporation and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/spf13/cobra"

	"github.com/Azure/buffalo-azure/generators/eventgrid"
)

// eventgridCmd represents the eventgrid command
var eventgridCmd = &cobra.Command{
	Use:     "eventgrid <name> [<EventTypeString>:<type identifier>...]",
	Aliases: []string{"eg"},
	Short:   "Generates new action(s) for handling Azure Event Grid events.",
	Long: `Add actions for reacting to Event Grid Events to your Buffalo application.

An EventTypeString is arbitrary, but often resembles a namespace. These are not
Go-specific and are decided by the tool originating the Event. The ones provided
by Microsoft take the form "Microsoft.<Service>.<Event>". Some examples include:
  - Microsoft.EventGrid.SubscriptionValidationEvent
  - Microsoft.Resources.ResourceWriteSuccess
  - Microsoft.Storage.BlobCreated

A type identifier is a Go-specific detail of this Plugin. They take the form
"<package name>.<type name>". Some examples include:
  - github.com/Azure/buffalo-azure/sdk/eventgrid.Cache
  - github.com/markbates/cash.Money
For the purpose of the Event Grid Buffalo generator, it will make most sense
to specify a type which well accomodates the unmarshaling a JSON object into
your type.

All together, you may find yourself running a command like:

buffalo generate eventgrid blobs \
Microsoft.Storage.BlobCreated \
Microsoft.Storage.BlobDeleted

or

buffalo generate eventgrid github \
GitHub.PullRequest:github.com/google/go-github/github.PullRequestEvent \
GitHub.Label:github.com/google/go-github/github.LabelEvent

More documentation about Event Grid can be found at:
https://azure.microsoft.com/en-us/services/event-grid/`,
	Run: func(cmd *cobra.Command, args []string) {

		name := args[0]
		types := make(map[string]reflect.Type, len(args[1:]))

		for _, arg := range args[1:] {
			eventType, goType, err := parseEventArg(arg)
			if err != nil {
				return
			}

			types[eventType], err = eventgrid.NewTypeStubIdentifier(goType)
		}

		gen := eventgrid.Generator{}

		if err := gen.Run(meta.New("."), name, types); err != nil {
			fmt.Fprintln(os.Stderr, "unable to create subscriber file: ", err)
			os.Exit(1)
		}
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("missing required arguments")
		}

		for _, arg := range args[1:] {
			if _, _, err := parseEventArg(arg); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(eventgridCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventgridCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventgridCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func parseEventArg(arg string) (string, string, error) {
	if typeIdentifier, ok := wellKnownEvents[arg]; ok {
		return arg, typeIdentifier, nil
	}

	last := strings.LastIndex(arg, ":")
	if last < 0 {
		return "", "", errors.New("unexpected argument format")
	}
	return arg[:last], arg[last+1:], nil
}
