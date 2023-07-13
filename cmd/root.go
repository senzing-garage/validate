/*
 */
package cmd

import (
	"context"
	"os"

	"github.com/senzing/senzing-tools/cmdhelper"
	"github.com/senzing/senzing-tools/envar"
	"github.com/senzing/senzing-tools/help"
	"github.com/senzing/senzing-tools/option"
	"github.com/senzing/validate/examplepackage"
	"github.com/spf13/cobra"
)

const (
	Short string = "Validates a JSON-lines file."
	Use   string = "validate"
	Long  string = `
Welcome to validate!
Validate the each line of a JSON-lines (JSONL) file conforms to the Generic Entity Specification.

Usage example:

validate --input-url "file:///path/to/json/lines/file.jsonl"
validate --input-url "https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl"
`
)

// ----------------------------------------------------------------------------
// Context variables
// ----------------------------------------------------------------------------

var ContextBools = []cmdhelper.ContextBool{
	// none defined
}

var ContextInts = []cmdhelper.ContextInt{
	// none defined
}

var ContextStrings = []cmdhelper.ContextString{
	{
		Default: cmdhelper.OsLookupEnvString(envar.Configuration, ""),
		Envar:   envar.Configuration,
		Help:    help.Configuration,
		Option:  option.Configuration,
	},
	{
		Default: cmdhelper.OsLookupEnvString(envar.EngineConfigurationJson, ""),
		Envar:   envar.EngineConfigurationJson,
		Help:    help.EngineConfigurationJson,
		Option:  option.EngineConfigurationJson,
	},
	{
		Default: cmdhelper.OsLookupEnvString(envar.LogLevel, "INFO"),
		Envar:   envar.LogLevel,
		Help:    help.LogLevel,
		Option:  option.LogLevel,
	},
	{
		Default: cmdhelper.OsLookupEnvString(envar.InputURL, ""),
		Envar:   envar.InputURL,
		Help:    help.InputURL,
		Option:  option.InputURL,
	}}

var ContextStringSlices = []cmdhelper.ContextStringSlice{
	// none defined
}

var ContextVariables = &cmdhelper.ContextVariables{
	Bools:        ContextBools,
	Ints:         ContextInts,
	Strings:      ContextStrings,
	StringSlices: ContextStringSlices,
}

// ----------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------

// Since init() is always invoked, define command line parameters.
func init() {
	cmdhelper.Init(RootCmd, *ContextVariables)
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
	cmdhelper.PreRun(cobraCommand, args, Use, *ContextVariables)
}

// Used in construction of cobra.Command
func RunE(_ *cobra.Command, _ []string) error {
	var err error = nil
	ctx := context.Background()
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
