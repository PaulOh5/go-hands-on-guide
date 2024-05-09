package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestHandleHttpError(t *testing.T) {
	helperMessage := "\nFor direction of command, Run \"mync http -h\"\n"
	testConfigs := []struct {
		args   []string
		err    error
		output string
	}{
		// 인수가 지정되지 않은 경우
		{
			args:   []string{},
			err:    ErrorInvalidHttpMethod,
			output: "invalid HTTP method" + helperMessage,
		},
		// 지원되지 않은 HTTP Method
		{
			args:   []string{"put", "http://localhost"},
			err:    ErrorInvalidHttpMethod,
			output: "invalid HTTP method" + helperMessage,
		},
		// GET Method에서 URL이 없는 경우
		{
			args:   []string{"get", "-output", "/tmp/test.txt"},
			err:    ErrorNoServerSpecified,
			output: "you have to specify the remote server" + helperMessage,
		},
		// POST Method에서 URL이 없는 경우
		{
			args:   []string{"post"},
			err:    ErrorNoServerSpecified,
			output: "you have to specify the remote server" + helperMessage,
		},
		// POST Option이 잘못된 경우
	}

	byteBuf := new(bytes.Buffer)

	for _, tc := range testConfigs {
		err := HandleHttp(byteBuf, tc.args)
		if err == nil {
			t.Fatal("Expected error, but got nil error")
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

func TestGetMethod(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	args := []string{ts.URL}
	byteBuf := new(bytes.Buffer)
	err := HandleGetHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	gotOutput := byteBuf.String()
	expectedOutput := "package1-0.1"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}

func TestGetMethodWithOutput(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()

	file, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary file")
	}
	defer os.Remove(file.Name())

	args := []string{"-output", file.Name(), ts.URL}
	byteBuf := new(bytes.Buffer)
	err = HandleGetHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	fileContent, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("Failed to read temporary file")
	}
	gotOutput := string(fileContent)
	expectedOutput := "package1-0.1"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}

func TestPostMethodWithStringBody(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	args := []string{"-body", `{"name":"test","version":"1.0"}`, ts.URL}
	byteBuf := new(bytes.Buffer)
	err := HandlePostHttp(byteBuf, args)
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
	args := []string{"-body-file", tmpFile.Name(), ts.URL}
	byteBuf := new(bytes.Buffer)
	err = HandlePostHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	gotOutput := byteBuf.String()
	expectedOutput := "Package registered with id: test-1.0\n"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}

func TestPostMethodWithFormData(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("some data")
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	args := []string{
		"-upload",
		tmpFile.Name(),
		"-formdata",
		"name=test",
		"-formdata",
		"version=1.0",
		ts.URL,
	}
	byteBuf := new(bytes.Buffer)
	err = HandlePostHttp(byteBuf, args)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
	gotOutput := byteBuf.String()
	expectedOutput := "Package registered with id: test-1.0\n"
	fileName := strings.Split(tmpFile.Name(), "/")
	expectedOutput += fmt.Sprintf("Filename: %s\n", fileName[len(fileName)-1])
	expectedOutput += "Size: 9\n"
	if expectedOutput != gotOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, gotOutput)
	}
}
