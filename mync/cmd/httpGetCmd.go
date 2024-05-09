package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type getConfig struct {
	url    string
	output string
}

func HandleGetHttp(w io.Writer, args []string) error {
	var output string

	fs := flag.NewFlagSet("HTTP GET Method", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&output, "output", "", "Output file path")

	fs.Usage = func() {
		var usageString = `
http get: Send HTTP GET Request
http get: <options> server`

		fmt.Fprint(w, usageString)
		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options: ")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return ErrorNoServerSpecified
	}

	c := getConfig{output: output}
	c.url = fs.Arg(0)
	err = fetchRemoteResource(w, c)
	if err != nil {
		return err
	}
	return nil
}

func fetchRemoteResource(w io.Writer, config getConfig) error {
	r, err := http.Get(config.url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if config.output == "" {
		fmt.Fprintf(w, "%s", data)
	} else {
		err = createFile(string(data), config.output)
		if err != nil {
			return err
		}
	}
	return nil
}

func createFile(data, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}
