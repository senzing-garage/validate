/*
 */
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/senzing/go-cmdhelping/cmdhelper"
	"github.com/senzing/go-cmdhelping/option"
	"github.com/senzing/validate/examplepackage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Short string = "Validates a JSON-lines file."
	Use   string = "validate"
	Long  string = `
Welcome to validate!
Validate that each line of a JSON-lines (JSONL) file conforms to the Generic Entity Specification.

Usage example:

validate --input-url "file:///path/to/json/lines/file.jsonl"
validate --input-url "https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl"
`
)

// ----------------------------------------------------------------------------
// Context variables
// ----------------------------------------------------------------------------

var ContextVariables = []option.ContextVariable{
	option.EngineModuleName.SetDefault(fmt.Sprintf("validate-%d", time.Now().Unix())),
	option.LogLevel,
	option.InputURL,
}

// ----------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------

// Since init() is always invoked, define command line parameters.
func init() {
	cmdhelper.Init(RootCmd, ContextVariables)
}

// ----------------------------------------------------------------------------
// Public functions
// ----------------------------------------------------------------------------

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Used in construction of cobra.Command
func PreRun(cobraCommand *cobra.Command, args []string) {
	cmdhelper.PreRun(cobraCommand, args, Use, ContextVariables)
}

// Used in construction of cobra.Command
func RunE(_ *cobra.Command, _ []string) error {
	var err error = nil
	ctx := context.Background()

	inputURL := viper.GetString(option.InputURL.Arg)
	inputURLLen := len(inputURL)

	// if inputURLLen == 0 {
	// 	//assume stdin
	// 	return readStdin()
	// }

	//This assumes the URL includes a schema and path so, minimally:
	//  "s://p" where the schema is 's' and 'p' is the complete path
	if inputURLLen < 5 {
		// logger.LogMessage(MessageIdFormat, 2002, fmt.Sprintf("Check the inputURL parameter: %s", inputURL))
		return fmt.Errorf("Check the inputURL parameter: [%s]", inputURL)
	}
	examplePackage := &examplepackage.ExamplePackageImpl{
		Something: "Main says 'Hi!'",
	}
	err = examplePackage.SaySomething(ctx)
	return err
}

// Used in construction of cobra.Command
func Version() string {
	return cmdhelper.Version(githubVersion, githubIteration)
}

// ----------------------------------------------------------------------------
// Command
// ----------------------------------------------------------------------------

// RootCmd represents the command.
var RootCmd = &cobra.Command{
	Use:     Use,
	Short:   Short,
	Long:    Long,
	PreRun:  PreRun,
	RunE:    RunE,
	Version: Version(),
}
