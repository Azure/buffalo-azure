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
	Use:     "eventgrid <name> [<EventTypeString>:<identifier>...]",
	Aliases: []string{"eg"},
	Short:   "Generates new action(s) for handling Azure Event Grid events.",
	Long: `More documenation can be found at:
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

		fmt.Println("eventgrid called")
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
	last := strings.LastIndex(arg, ":")
	if last < 0 {
		return "", "", errors.New("unexpected argument format")
	}
	return arg[:last], arg[last+1:], nil
}
