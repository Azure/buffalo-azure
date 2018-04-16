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
	"fmt"

	"github.com/spf13/cobra"
)

// eventgridCmd represents the eventgrid command
var eventgridCmd = &cobra.Command{
	Use:   "eventgrid",
	Short: "Generates new action(s) for handling Azure Event Grid events.",
	Long: `More documenation can be found at:
https://azure.microsoft.com/en-us/services/event-grid/`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("eventgrid called")
	},
}

// Flavor indicates which set of templates and HTTP Muxes to use
// to host your Event Grid application.
type Flavor string

// All of the well-known HTTP Mux implementations that are either supported
// or plan to be supported by the eventgrid generator.
const (
	FlavorBuffalo Flavor = "buffalo"
	FlavorStdlib  Flavor = "stdlib"
	FlavorGorilla Flavor = "gorilla"
	FlavorEcho    Flavor = "echo"
)

func init() {
	rootCmd.AddCommand(eventgridCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventgridCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventgridCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	eventgridCmd.PersistentFlags().StringP("flavor", "f", string(FlavorBuffalo), "The HTTP framework that should be used for receiving and dispatching Event Grid events.")
}
