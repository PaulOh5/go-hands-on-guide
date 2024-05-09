package cmd

import (
	"errors"
	"fmt"
	"io"
)

type pkgData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: mync http [get|post] -h")
	HandleGetHttp(w, []string{"-h"})
	HandlePostHttp(w, []string{"-h"})
}

func HandleHttp(w io.Writer, args []string) error {
	var err error
	if len(args) < 1 {
		err = ErrorInvalidHttpMethod
	} else {
		switch args[0] {
		case "get":
			err = HandleGetHttp(w, args[1:])
		case "post":
			err = HandlePostHttp(w, args[1:])
		case "-h":
			printUsage(w)
		case "--help":
			printUsage(w)
		default:
			err = ErrorInvalidHttpMethod
		}
	}

	if errors.Is(err, ErrorNoServerSpecified) ||
		errors.Is(err, ErrorInvalidHttpMethod) ||
		errors.Is(err, ErrorInvalidHTTPPostOption) {
		fmt.Fprintln(w, err)
		fmt.Fprintln(w, "For direction of command, Run \"mync http -h\"")
	}

	return err
}
