//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
 * The unit tests in this file simulate command line invocation.
 */
func TestMain(test *testing.T) {
	tempDir := test.TempDir()
	inputFile := filepath.Join(tempDir, "move-main-input.jsonl")
	err := touchFile(inputFile)
	require.NoError(test, err)
	os.Setenv("SENZING_TOOLS_INPUT_URL", fmt.Sprintf("file://%s", inputFile))
	main()
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
