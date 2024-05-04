package main

// import (
// 	"errors"
// 	"os/exec"
// 	"testing"
// )

// func TestMain(t *testing.T) {
// 	tests := []struct {
// 		cmd      []string
// 		exitCode int
// 	}{
// 		{
// 			cmd:      []string{"./application", "-h"},
// 			exitCode: 0,
// 		},
// 		{
// 			cmd:      []string{"./application", "5"},
// 			exitCode: 0,
// 		},
// 		{
// 			cmd:      []string{"./application", "0"},
// 			exitCode: 1,
// 		},
// 		{
// 			cmd:      []string{"./application", "-1"},
// 			exitCode: 1,
// 		},
// 	}

// 	for _, tc := range tests {
// 		cmd := exec.Command(tc.cmd[0], tc.cmd[1:]...)
// 		err := cmd.Run()
// 		if err != nil {
// 			var exitErr *exec.ExitError
// 			if errors.As(err, &exitErr) {
// 				if exitErr.ExitCode() != tc.exitCode {
// 					t.Log(tc.cmd)
// 					t.Log(err)
// 					t.Errorf("Expected exit code: %d, got: %d\n", tc.exitCode, exitErr.ExitCode())
// 				}
// 			} else {
// 				t.Errorf("Expected exit error, got: %v\n", err)
// 			}
// 		}
// 	}
// }
