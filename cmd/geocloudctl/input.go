package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const fromStdin = "-"

func getInput(cmd *cobra.Command) (io.Reader, string, error) {
	var (
		path = cmd.Flag("file").Value.String()
		r    io.Reader
		err  error
	)
	switch path {
	case "":
		return nil, "", fmt.Errorf("--file or -f is required")
	case fromStdin:
		r = os.Stdin
	default:
		r, err = os.Open(path)
		if err != nil {
			return nil, "", err
		}
	}

	contentType := cmd.Flag("content-type").Value.String()
	if contentType == "" {
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, "", err
		}

		// if Content-Type is not supplied, try to detect it. If its not detected as zip,
		// then its either 'application/json' or not. In either case, set it to 'application/json'
		if contentType = http.DetectContentType(b); contentType != "application/zip" {
			contentType = "application/json"
		}

		r = bytes.NewReader(b)
	} else if contentType != "application/json" && contentType != "application/zip" {
		return nil, "", fmt.Errorf("unknown Content-Type '%s'", contentType)
	}

	return r, contentType, nil
}
