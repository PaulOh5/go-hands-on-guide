package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type httpConfig struct {
	url    string
	verb   string
	output string
}

func HandleHttp(w io.Writer, args []string) error {
	var v string
	var output string
	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&v, "verb", "GET", "HTTP method")
	fs.StringVar(&output, "output", "", "Output file path")

	fs.Usage = func() {
		var usageString = `
http: A HTTP client.

http: <options> server`

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

	c := httpConfig{verb: v, output: output}
	c.url = fs.Arg(0)

	switch c.verb {
	case "GET":
		data, err := fetchRemoteResource(c.url)
		if err != nil {
			return err
		}

		if c.output == "" {
			fmt.Fprintf(w, "%s\n", data)
		} else {
			f, err := os.Create(c.output)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.Write(data)
			if err != nil {
				return err
			}
		}
	case "POST":
		fmt.Fprintln(w, "Executing http POST command")
	case "HEAD":
		fmt.Fprintln(w, "Executing http HEAD command")
	default:
		return ErrorInvalidHttpMethod
	}
	return nil
}

func fetchRemoteResource(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
