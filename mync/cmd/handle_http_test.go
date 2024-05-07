package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func packageRegHTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		p := pkgData{}
		d := pkgRegisterResult{}
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(data, &p)
		if err != nil || len(p.Name) == 0 || len(p.Version) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		d.Id = p.Name + "-" + p.Version
		jsonData, err := json.Marshal(d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	} else {
		http.Error(w, "Invalid HTTP method specified", http.StatusMethodNotAllowed)
		return
	}
}

func StartTestPackageServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(packageRegHTTPHandler))
	return ts
}

func TestHandleHttp(t *testing.T) {
	usageMessage := "\nhttp: A HTTP client.\n\nhttp: <options> server\n\nOptions: \n  -body string\n    \tBody of request (only json format string)\n  -bodyFilePath string\n    \tFile path of body of request (only json file)\n  -output string\n    \tOutput file path\n  -verb string\n    \tHTTP method (default \"GET\")\n"
	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		// 인수가 지정되지 않은 경우
		{
			args: []string{},
			err:  ErrorNoServerSpecified,
		},
		// 인수가 -h로 지정된 경우
		{
			args:   []string{"-h"},
			err:    errors.New("flag: help requested"),
			output: usageMessage,
		},
		{
			args: []string{"-verb", "PUT", "http://localhost"},
			err:  ErrorInvalidHttpMethod,
		},
	}

	byteBuf := new(bytes.Buffer)

	for _, tc := range testConfigs {
		err := HandleHttp(byteBuf, tc.args)
		if tc.err == nil && err != nil {
			t.Fatalf("Expected nil error, but got %v", err)
		}
		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Fatalf("Expected error %v, but got %v", tc.err, err)
		}
		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("Expected output %q, but got %q", tc.output, gotOutput)
			}
		}
		byteBuf.Reset()
	}
}

func TestPostMethodWithString(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	args := []string{"-verb", "POST", "-body", `{"name":"test","version":"1.0"}`, ts.URL}
	byteBuf := new(bytes.Buffer)
	err := HandleHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	gotOutput := byteBuf.String()
	expectedOutput := "Package registered with id: test-1.0\n"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}

func TestPostMethodWithFile(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(`{"name":"test","version":"1.0"}`)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	args := []string{"-verb", "POST", "-body-file", tmpFile.Name(), ts.URL}
	byteBuf := new(bytes.Buffer)
	err = HandleHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	gotOutput := byteBuf.String()
	expectedOutput := "Package registered with id: test-1.0\n"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}
