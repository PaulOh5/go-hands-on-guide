package main

import (
	"errors"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		c   config
		err error
	}{
		{
			c:   config{},
			err: errors.New("must specify a number greater than 0"),
		},
		{
			c:   config{numTimes: -1},
			err: errors.New("must specify a number greater than 0"),
		},
		{
			c:   config{numTimes: 10},
			err: nil,
		},
	}

	for _, tc := range tests {
		err := validateArgs(tc.c)
		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("Expected error to be: %v, but got: %v\n", tc.err, err)
		}
		if tc.err == nil && err != nil {
			t.Errorf("Expected no error, but got: %v\n", err)
		}
	}
}
