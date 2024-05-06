package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/PaulOh5/mync/cmd"
)

var errInvalidSubCommand = errors.New("invalid sub-command specified")

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: mync [http|grpc] -h")
	cmd.HandleHttp(w, []string{"-h"})
	cmd.HandleGrpc(w, []string{"-h"})
}

func handleCommand(w io.Writer, args []string) error {
	var err error
	if len(args) < 1 {
		err = errInvalidSubCommand
	} else {
		switch args[0] {
		case "http":
			err = cmd.HandleHttp(w, args[1:])
		case "grpc":
			err = cmd.HandleGrpc(w, args[1:])
		case "-h":
			printUsage(w)
		case "--help":
			printUsage(w)
		default:
			err = errInvalidSubCommand
		}
	}

	if errors.Is(err, cmd.ErrorNoServerSpecified) ||
		errors.Is(err, errInvalidSubCommand) ||
		errors.Is(err, cmd.ErrorInvalidHttpMethod) {
		fmt.Fprintln(w, err)
		printUsage(w)
	}
	return err
}

func main() {
	err := handleCommand(os.Stdout, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
