package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

/*
 * The unit tests in this file simulate command line invocation.
 */
func Test_ExecuteCommand_NoInputURL(t *testing.T) {
	cmd := RootCmd
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	// log.SetOutput(outbuf)
	// log.SetOutput(errbuf)
	// defer func() {
	// 	log.SetOutput(os.Stderr)
	// }()
	// cmd.SetArgs([]string{"--help"})
	cmd.SetArgs([]string{"--input-url", "none"})
	exError := RootCmd.Execute()
	if exError == nil {
		t.Fatalf("expected Execute() to generated an error")
	}
	stderr, err := ioutil.ReadAll(errbuf)
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := ioutil.ReadAll(outbuf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("stderr:", string(stderr))
	fmt.Println("stdout:", string(stdout))
	if !strings.Contains(string(stderr), "Check the input-url parameter") {
		t.Fatalf("expected input-url parameter error")
	}
}

func Test_ExecuteCommand_Help(t *testing.T) {
	cmd := RootCmd
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	cmd.SetArgs([]string{"--help"})
	RootCmd.Execute()

	stdout, err := ioutil.ReadAll(outbuf)
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println("stdout:", string(stdout))
	if !strings.Contains(string(stdout), "Available Commands") {
		t.Fatalf("expected help text")
	}
}
