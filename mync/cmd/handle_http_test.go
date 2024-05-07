package cmd

import (
	"bytes"
	"errors"
	"testing"
)

func TestHandleHttp(t *testing.T) {
	usageMessage := "\nhttp: A HTTP client.\n\nhttp: <options> server\n\nOptions: \n  -output string\n    \tOutput file path\n  -verb string\n    \tHTTP method (default \"GET\")\n"
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
