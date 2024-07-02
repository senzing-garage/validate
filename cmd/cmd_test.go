package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Test public functions
// ----------------------------------------------------------------------------

/*
 * The unit tests in this file simulate command line invocation.
 */
func Test_Execute(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "--help"}
	Execute()
}

func Test_Execute_completion(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "completion"}
	Execute()
}

func Test_Execute_docs(test *testing.T) {
	_ = test
	os.Args = []string{"command-name", "docs"}
	Execute()
}

// func Test_ExecuteCommand_NoInputURL(test *testing.T) {
// 	cmd := RootCmd
// 	outbuf := bytes.NewBufferString("")
// 	errbuf := bytes.NewBufferString("")
// 	cmd.SetOut(outbuf)
// 	cmd.SetErr(errbuf)
// 	cmd.SetArgs([]string{"--input-url", "none"})
// 	err := RootCmd.Execute()
// 	require.Error(test, err, "expected Execute() to generated an error")
// 	stderr, err := io.ReadAll(errbuf)
// 	require.NoError(test, err)
// 	if !strings.Contains(string(stderr), "validation failed") {
// 		test.Fatalf("expected input-url parameter error")
// 	}
// }

func Test_ExecuteCommand_Help(test *testing.T) {
	cmd := RootCmd
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	cmd.SetArgs([]string{"--help"})
	err := RootCmd.Execute()
	require.NoError(test, err)
	stdout, err := io.ReadAll(outbuf)
	require.NoError(test, err)
	// fmt.Println("stdout:", string(stdout))
	if !strings.Contains(string(stdout), "Available Commands") {
		test.Fatalf("expected help text")
	}
}

// ----------------------------------------------------------------------------
// Test private functions
// ----------------------------------------------------------------------------

func Test_completionAction(test *testing.T) {
	var buffer bytes.Buffer
	err := completionAction(&buffer)
	require.NoError(test, err)
}

func Test_docsAction_badDir(test *testing.T) {
	var buffer bytes.Buffer
	badDir := "/tmp/no/directory/exists"
	err := docsAction(&buffer, badDir)
	require.Error(test, err)
}

// ----------------------------------------------------------------------------
// Utiity functions
// ----------------------------------------------------------------------------

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
