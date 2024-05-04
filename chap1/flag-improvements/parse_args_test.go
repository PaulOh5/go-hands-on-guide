package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		args []string
		config
		output string
		err    error
	}{
		{
			args:   []string{"-n", "10"},
			config: config{numTimes: 10},
		},
		{
			args:   []string{"-n", "abc"},
			err:    errors.New("invalid value \"abc\" for flag -n: parse error"),
			config: config{numTimes: 0},
		},
		{
			args:   []string{"-n", "1", "Paul"},
			err:    nil,
			config: config{numTimes: 1, name: "Paul"},
		},
		{
			args:   []string{"-n", "1", "Jone", "Paul"},
			err:    errors.New("more than one positional argument specifed"),
			config: config{numTimes: 1},
		},
	}
	byteBuf := new(bytes.Buffer)
	for _, tc := range tests {
		c, err := parseArgs(byteBuf, tc.args)
		if tc.err == nil && err != nil {
			t.Fatalf("Expected nil error, got: %v\n", err)
		}
		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Fatalf("Expected error to be: %v, got: %v\n", tc.err, err)
		}
		if c.numTimes != tc.numTimes {
			t.Errorf("Expected numTimes to be: %v, got: %v\n", tc.numTimes, c.numTimes)
		}
		gotMsg := byteBuf.String()
		if len(tc.output) > 0 && gotMsg != tc.output {
			t.Errorf("Expected output to be: %v, Got: %v\n", tc.output, gotMsg)
		}
		byteBuf.Reset()
	}
}
