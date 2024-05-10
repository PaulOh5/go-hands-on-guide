package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type getConfig struct {
	url             string
	output          string
	disableRedirect bool
	timeout         int
	header          Header
	auth            string
	report          bool
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

type ReportClient struct {
	log *log.Logger
}

func (c ReportClient) RoundTrip(r *http.Request) (*http.Response, error) {
	startTime := time.Now()
	resp, err := http.DefaultTransport.RoundTrip(r)
	elapsedTime := time.Since(startTime)
	c.log.Printf("Execution time: %s\n", elapsedTime)
	return resp, err
}

func HandleGetHttp(w io.Writer, args []string) error {
	var output string
	var disableRedirect bool
	var timeout int
	var header Header
	var auth string
	var report bool

	fs := flag.NewFlagSet("HTTP GET Method", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&output, "output", "", "Output file path")
	fs.BoolVar(&disableRedirect, "disable-redirect", false, "Disable redirection")
	fs.IntVar(&timeout, "timeout", 1000, "Time out, unit is ms (default=1000ms)")
	fs.Var(&header, "header", "Header value (key=value)")
	fs.StringVar(&auth, "basicauth", "", "Atuh value (user:password)")
	fs.BoolVar(&report, "report", false, "Latency report (default=false)")

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
		timeout:         timeout,
		header:          header,
		auth:            auth,
		report:          report,
	}
	c.url = fs.Arg(0)
	httpClient := createHTTPClient(c)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()
	request, err := createHTTPGetRequest(ctx, c)
	if err != nil {
		return err
	}
	err = fetchRemoteResource(w, httpClient, request, c)
	if err != nil {
		return err
	}
	return nil
}

func createHTTPClient(config getConfig) *http.Client {
	var client *http.Client

	if config.disableRedirect {
		client = &http.Client{
			Timeout:       time.Duration(config.timeout) * time.Millisecond,
			CheckRedirect: redirectPolicyFunc,
		}
	} else {
		client = &http.Client{}
	}

	if config.report {
		reportTransport := ReportClient{}
		client.Transport = &reportTransport
	}

	return client
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if len(via) >= 1 {
		return errors.New("stopped after 1 redirect")
	}
	return nil
}

func fetchRemoteResource(w io.Writer, client *http.Client, request *http.Request, config getConfig) error {
	r, err := client.Do(request)
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

func createHTTPGetRequest(ctx context.Context, config getConfig) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", config.url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range config.header {
		req.Header.Add(k, v)
	}

	authParts := strings.Split(config.auth, ":")
	if len(authParts) == 2 {
		username := authParts[0]
		password := authParts[1]
		req.SetBasicAuth(username, password)
	} else {
		return nil, errors.New("invalid auth string. auth string must be a \"username:password\"")
	}

	return req, err
}
