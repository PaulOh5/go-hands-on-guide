package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

type postConfig struct {
	url          string
	body         string
	bodyFilePath string
	upload       string
	formData     FormData
}

type pkgRegisterResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type FormData []string

func (f *FormData) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *FormData) String() string {
	return fmt.Sprint(*f)
}

func HandlePostHttp(w io.Writer, args []string) error {
	var body string
	var bodyFilePath string
	var upload string
	var formData FormData

	fs := flag.NewFlagSet("HTTP POST Method", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&body, "body", "", "Body of request (only json fromat string)")
	fs.StringVar(&bodyFilePath, "body-file", "", "File path of body for reuqest (only json file)")
	fs.StringVar(&upload, "upload", "", "Upload file path")
	fs.Var(&formData, "formdata", "Form data (key=value)")

	fs.Usage = func() {
		var usageString = `
http post: Send HTTP POST Request
http post: <options> server`

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

	c := postConfig{
		body:         body,
		bodyFilePath: bodyFilePath,
		upload:       upload,
		formData:     formData,
	}
	c.url = fs.Arg(0)
	requestBody, contentType, err := createBody(c)
	if err != nil {
		return err
	}
	result, err := registerPakcage(requestBody, contentType, c.url)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Package registered with id: %s\n", result.ID)
	if result.Name != "" {
		fmt.Fprintf(w, "Filename: %s\n", result.Name)
	}
	if result.Size != 0 {
		fmt.Fprintf(w, "Size: %d\n", result.Size)
	}
	return nil
}

func createBody(config postConfig) (*bytes.Buffer, string, error) {
	var b *bytes.Buffer
	if config.upload != "" && len(config.formData) > 0 {
		return createMultiPartMessage(config.upload, config.formData)
	} else if config.bodyFilePath != "" {
		data, err := os.ReadFile(config.bodyFilePath)
		if err != nil {
			return b, "", err
		}
		b = bytes.NewBuffer(data)
		return b, "application/json", nil
	} else if config.body != "" {
		b = bytes.NewBuffer([]byte(config.body))
		return b, "application/json", nil
	} else {
		return b, "", ErrorInvalidHTTPPostOption
	}
}

func createMultiPartMessage(upload string, formData []string) (*bytes.Buffer, string, error) {
	var b bytes.Buffer
	var err error
	var fw io.Writer

	mw := multipart.NewWriter(&b)
	for _, data := range formData {
		kv := strings.Split(data, "=")
		fw, err = mw.CreateFormField(kv[0])
		if err != nil {
			return &b, "", err
		}
		fmt.Fprint(fw, kv[1])
	}
	fw, err = mw.CreateFormFile("filedata", upload)
	if err != nil {
		return &b, "", err
	}
	file, err := os.Open(upload)
	if err != nil {
		return &b, "", err
	}
	defer file.Close()
	_, err = io.Copy(fw, file)
	if err != nil {
		return &b, "", err
	}
	mw.Close()
	contentType := mw.FormDataContentType()
	return &b, contentType, nil
}

func registerPakcage(body *bytes.Buffer, contentType string, url string) (pkgRegisterResult, error) {
	p := pkgRegisterResult{}
	r, err := http.Post(url, contentType, body)
	if err != nil {
		return p, err
	}
	defer r.Body.Close()
	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		return p, err
	}
	if r.StatusCode != http.StatusOK {
		return p, errors.New(string(responseData))
	}
	err = json.Unmarshal(responseData, &p)
	return p, err
}
