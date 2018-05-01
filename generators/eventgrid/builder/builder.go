// Stuff
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	path     string
	contents []byte
}

func main() {
	exitStatus := 1
	defer func() {
		os.Exit(exitStatus)
	}()
	var err error

	var output io.Writer = os.Stdout
	for i, arg := range os.Args {
		if arg == "-o" {
			output, err = os.Create(os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2])
			if err != nil {
				return
			}
			break
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	discoveredFiles := make(chan file)

	stagingHandle, err := ioutil.TempFile("", "buffalo_staging")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to create a temporary file to stage the results.")
		return
	}
	defer os.Remove(stagingHandle.Name())

	fmt.Fprintln(stagingHandle, `package eventgrid

var staticTemplates = make(TemplateCache)

func init(){`)

	writeErr := make(chan error)
	go func(result chan<- error) {
		result <- writeFileEntry(ctx, stagingHandle, discoveredFiles)
		close(result)
	}(writeErr)

	readErr := make(chan error)
	go func(result chan<- error) {
		for _, arg := range os.Args {
			if err := readFiles(ctx, arg, discoveredFiles); err != nil {
				result <- err
				break
			}
		}
		close(result)
		close(discoveredFiles)
	}(readErr)

	var readDone, writeDone bool
	for {
		if readDone && writeDone {
			break
		}
		//time.Sleep(200 * time.Millisecond)

		select {
		case err, ok := <-readErr:
			if ok && err != context.Canceled && err != nil {
				fmt.Fprintln(os.Stderr, "unable to read files: ", err)
				cancel()
			}
			readDone = true
		case err, ok := <-writeErr:
			if ok && err != context.Canceled && err != nil {
				fmt.Fprintln(os.Stderr, "unable to write files: ", err)
				cancel()
			}
			writeDone = true
		}
	}

	fmt.Fprintln(stagingHandle, "}")

	stagingHandle.Close()

	stagingReader, err := os.Open(stagingHandle.Name())
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read staging file: ", err)
		return
	}

	_, err = io.Copy(output, stagingReader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to copy staging file to Stdout: ", err)
		return
	}

	exitStatus = 0
}

func writeFileEntry(ctx context.Context, output io.Writer, input <-chan file) error {
	lineItem := bytes.NewBuffer([]byte{})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry, ok := <-input:
			if !ok {
				return nil
			}

			if len(entry.contents) == 0 {
				return errors.New("cannot process an empty file")
			}

			fmt.Fprintf(lineItem, "\tstaticTemplates[%q] = []byte{ ", strings.Replace(entry.path, `\`, "/", -1))

			const terminator = ", "
			for _, item := range entry.contents {
				fmt.Fprintf(lineItem, "%d%s", item, terminator)
			}
			lineItem.Truncate(lineItem.Len() - len(terminator))

			fmt.Fprint(lineItem, "}\n")

			if _, err := io.Copy(output, lineItem); err != nil {
				return err
			}
			lineItem.Reset()
		}
	}
}

func readFiles(ctx context.Context, root string, output chan<- file) error {
	addFile := func(path string, info os.FileInfo, outerError error) error {
		if outerError != nil {
			return outerError
		}

		// Skip files that aren't Go source code.
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		handle, err := os.Open(path)
		if err != nil {
			return err
		}
		defer handle.Close()

		result := file{
			path: path,
		}
		result.contents, err = ioutil.ReadAll(handle)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case output <- result:
			// Intentionally Left Blank
		}

		return nil
	}

	retChan := make(chan error, 1)
	go func() {
		retChan <- filepath.Walk(root, addFile)
	}()

	return <-retChan
}
