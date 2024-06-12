package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
 * The unit tests in this file simulate command line invocation.
 */
func Test_ExecuteCommand_NoInputURL(test *testing.T) {
	cmd := RootCmd
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	cmd.SetArgs([]string{"--input-url", "none"})
	exError := RootCmd.Execute()
	if exError == nil {
		test.Fatalf("expected Execute() to generated an error")
	}
	stderr, err := io.ReadAll(errbuf)
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(string(stderr), "validation failed") {
		test.Fatalf("expected input-url parameter error")
	}
}

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
	if err != nil {
		test.Fatal(err)
	}
	// fmt.Println("stdout:", string(stdout))
	if !strings.Contains(string(stdout), "Available Commands") {
		test.Fatalf("expected help text")
	}
}
