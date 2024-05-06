package cmd

import (
	"bytes"
	"testing"
)

func TestGrpcHttp(t *testing.T) {
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
		{
			args:   []string{"-method", "service.host.local/method", "-body", "{}", "http://localhost"},
			err:    nil,
			output: "Executing grpc command\n",
		},
	}

	byteBuf := new(bytes.Buffer)

	for _, tc := range testConfigs {
		err := HandleGrpc(byteBuf, tc.args)
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
