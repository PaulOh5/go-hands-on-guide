package main

import (
	"bytes"
	"testing"
)

func TestHandleCommand(t *testing.T) {
	usageMessage := "Usage: mync [http|grpc] -h\n\nhttp: A HTTP client.\n\nhttp: <options> server\n\nOptions: \n  -output string\n    \tOutput file path\n  -verb string\n    \tHTTP method (default \"GET\")\n\n\ngrpc: A gRPC client.\n\ngrpc: <options> server\n\nOptions: \n  -body string\n    \tBody of request\n  -method string\n    \tMethod to call\n"

	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		// 인수가 지정되지 않은 경우
		{
			args:   []string{},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
		// 인수가 -h로 지정된 경우
		{
			args:   []string{"-h"},
			err:    nil,
			output: usageMessage,
		},
		// 서브 커맨드가 잘못 지정된 경우
		{
			args:   []string{"foo"},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
	}

	byteBuf := new(bytes.Buffer)

	for _, tc := range testConfigs {
		err := handleCommand(byteBuf, tc.args)
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
