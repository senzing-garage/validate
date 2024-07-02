//go:build linux

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PreRun_Linux(test *testing.T) {
	_ = test
	args := []string{"command-name"}
	PreRun(RootCmd, args)
}

func Test_RunE_Linux(test *testing.T) {
	tempDir := test.TempDir()
	inputFile := filepath.Join(tempDir, "move-cmd-input.jsonl")
	err := touchFile(inputFile)
	require.NoError(test, err)
	os.Setenv("SENZING_TOOLS_INPUT_URL", fmt.Sprintf("file://%s", inputFile))
	err = RunE(RootCmd, []string{})
	require.NoError(test, err)
}

// ----------------------------------------------------------------------------
// Test private functions
// ----------------------------------------------------------------------------

func Test_docsAction(test *testing.T) {
	var buffer bytes.Buffer
	err := docsAction(&buffer, "/tmp")
	require.NoError(test, err)
}
