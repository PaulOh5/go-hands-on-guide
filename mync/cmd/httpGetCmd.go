package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type getConfig struct {
	url             string
	output          string
	disableRedirect bool
	header          Header
	auth            string
}

type Header map[string]string

func (h *Header) Set(value string) error {
	kv := strings.Split(value, "=")
	(*h)[kv[0]] = kv[1]
	return nil
}

func (h *Header) String() string {
	return fmt.Sprint(*h)
}

func HandleGetHttp(w io.Writer, args []string) error {
	var output string
	var disableRedirect bool
	var header Header
	var auth string

	fs := flag.NewFlagSet("HTTP GET Method", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&output, "output", "", "Output file path")
	fs.BoolVar(&disableRedirect, "disable-redirect", false, "Disable redirection")
	fs.Var(&header, "header", "Header value (key=value)")
	fs.StringVar(&auth, "basicauth", "", "Atuh value (user:password)")

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

	c := getConfig{
		output:          output,
		disableRedirect: disableRedirect,
		header:          header,
		auth:            auth,
	}
	c.url = fs.Arg(0)
	httpClient := createHTTPClient(c)
	err = fetchRemoteResource(w, httpClient, c)
	if err != nil {
		return err
	}
	return nil
}

func createHTTPClient(config getConfig) *http.Client {
	if config.disableRedirect {
		return &http.Client{CheckRedirect: redirectPolicyFunc}
	} else {
		return &http.Client{}
	}
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if len(via) >= 1 {
		return errors.New("stopped after 1 redirect")
	}
	return nil
}

func fetchRemoteResource(w io.Writer, client *http.Client, config getConfig) error {
	r, err := client.Get(config.url)
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

// 여기서부터 시작
func createHTTPGetRequest(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req, err
}
