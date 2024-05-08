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

type httpConfig struct {
	url          string
	verb         string
	output       string
	body         string
	bodyFilePath string
	upload       string
	formdata     FormData
}

type pkgData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type pkgRegisterResult struct {
	Id   string `json:"id"`
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

func HandleHttp(w io.Writer, args []string) error {
	var v string
	var output string
	var b string
	var bfp string
	var u string
	var formdata FormData

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&v, "verb", "GET", "HTTP method")
	fs.StringVar(&output, "output", "", "Output file path")
	fs.StringVar(&b, "body", "", "Body of request (only json format string)")
	fs.StringVar(&bfp, "body-file", "", "File path of body of request (only json file)")
	fs.StringVar(&u, "upload", "", "Upload file path")
	fs.Var(&formdata, "formdata", "Form data (key=value)")

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
		upload:       u,
		formdata:     formdata,
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
		if c.upload != "" && len(c.formdata) > 0 {
			result, err := registerPackageData(c.url, c.upload, c.formdata)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "Package registered with id: %s\n", result.Id)
			fmt.Fprintf(w, "Filename: %s\n", result.Name)
			fmt.Fprintf(w, "Size: %d\n", result.Size)
		} else if c.bodyFilePath != "" {
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

func registerPackageData(url string, path string, formData FormData) (pkgRegisterResult, error) {
	p := pkgRegisterResult{}
	var b bytes.Buffer
	var fw io.Writer
	var err error

	mw := multipart.NewWriter(&b)
	for _, data := range formData {
		kv := strings.Split(data, "=")
		fw, err = mw.CreateFormField(kv[0])
		if err != nil {
			return p, err
		}
		fmt.Fprint(fw, kv[1])
	}
	fw, err = mw.CreateFormFile("filedata", path)
	if err != nil {
		return p, err
	}
	file, err := os.Open(path)
	if err != nil {
		return p, err
	}
	defer file.Close()
	_, err = io.Copy(fw, file)
	if err != nil {
		return p, err
	}
	mw.Close()
	contentType := mw.FormDataContentType()
	r, err := http.Post(url, contentType, &b)
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
