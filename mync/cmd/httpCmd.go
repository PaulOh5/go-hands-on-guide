package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type httpConfig struct {
	url          string
	verb         string
	output       string
	body         string
	bodyFilePath string
}

type pkgData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type pkgRegisterResult struct {
	Id string `json:"id"`
}

func HandleHttp(w io.Writer, args []string) error {
	var v string
	var output string
	var b string
	var bfp string
	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&v, "verb", "GET", "HTTP method")
	fs.StringVar(&output, "output", "", "Output file path")
	fs.StringVar(&b, "body", "", "Body of request (only json format string)")
	fs.StringVar(&bfp, "body-file", "", "File path of body of request (only json file)")

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

	c := httpConfig{
		verb:         v,
		output:       output,
		body:         b,
		bodyFilePath: bfp,
	}
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
			err = createFile(string(data), c.output)
			if err != nil {
				return err
			}
		}
	case "POST":
		if c.bodyFilePath != "" {
			data, err := os.ReadFile(c.bodyFilePath)
			if err != nil {
				return err
			}
			result, err := registerPackage(c.url, data)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "Package registered with id: %s\n", result.Id)
		} else if c.body != "" {
			result, err := registerPackage(c.url, []byte(c.body))
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "Package registered with id: %s\n", result.Id)
		} else {
			return ErrorInvalidHTTPPostOption
		}
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

func registerPackage(url string, data []byte) (pkgRegisterResult, error) {
	p := pkgRegisterResult{}
	reader := bytes.NewReader(data)
	r, err := http.Post(url, "application/json", reader)
	if err != nil {
		return p, err
	}
	defer r.Body.Close()
	respData, err := io.ReadAll(r.Body)
	if err != nil {
		return p, err
	}
	if r.StatusCode != http.StatusOK {
		return p, errors.New(string(respData))
	}
	err = json.Unmarshal(respData, &p)
	return p, err
}
