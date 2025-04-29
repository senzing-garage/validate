package cmd_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/senzing-garage/go-helpers/wraperror"
	"github.com/senzing-garage/validate/cmd"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Test public functions
// ----------------------------------------------------------------------------

func Test_CompletionCmd(test *testing.T) {
	test.Parallel()

	err := cmd.CompletionCmd.Execute()
	require.NoError(test, err)
	err = cmd.CompletionCmd.RunE(cmd.CompletionCmd, []string{})
	require.NoError(test, err)
}

func Test_DocsCmd(test *testing.T) {
	test.Parallel()

	err := cmd.DocsCmd.Execute()
	require.NoError(test, err)
	err = cmd.DocsCmd.RunE(cmd.DocsCmd, []string{})
	require.NoError(test, err)
}

func Test_Execute(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "--help"}

	cmd.Execute()
}

func Test_Execute_completion(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "completion"}

	cmd.Execute()
}

func Test_Execute_docs(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "docs"}

	cmd.Execute()
}

func Test_Execute_help(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "--help"}

	cmd.Execute()
}

func Test_ExecuteCommand_Help(test *testing.T) {
	cmd := cmd.RootCmd
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")

	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(test, err)
	stdout, err := io.ReadAll(outbuf)
	require.NoError(test, err)
	// fmt.Println("stdout:", string(stdout))
	if !strings.Contains(string(stdout), "Available Commands") {
		test.Fatalf("expected help text")
	}
}

func Test_PreRun(test *testing.T) {
	_ = test
	args := []string{"command-name", "--help"}
	cmd.PreRun(cmd.RootCmd, args)
}

// func Test_RunE(test *testing.T) {
// 	test.Setenv("SENZING_TOOLS_AVOID_SERVING", "true")
// 	err := RunE(RootCmd, []string{})
// 	require.NoError(test, err)
// }

// func Test_RootCmd(test *testing.T) {
// 	_ = test
// 	err := RootCmd.Execute()
// 	require.NoError(test, err)
// 	err = RootCmd.RunE(RootCmd, []string{})
// 	require.NoError(test, err)
// }

// ----------------------------------------------------------------------------
// Utility functions
// ----------------------------------------------------------------------------

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		return wraperror.Errorf(err, "touchFile.os.OpenFile error: %w", err)
	}

	err = file.Close()

	return wraperror.Errorf(err, "touchFile error: %w", err)
}
