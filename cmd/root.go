/*
 */
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/senzing-garage/go-cmdhelping/cmdhelper"
	"github.com/senzing-garage/go-cmdhelping/option"
	"github.com/senzing-garage/go-helpers/wraperror"
	"github.com/senzing-garage/validate/validate"
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

var ContextVariablesForMultiPlatform = []option.ContextVariable{
	option.EngineInstanceName.SetDefault(fmt.Sprintf("validate-%d", time.Now().Unix())),
	option.InputFileType,
	option.InputURL,
	option.JSONOutput,
	option.LogLevel,
}

var ContextVariables = append(ContextVariablesForMultiPlatform, ContextVariablesForOsArch...)

var errPackage = errors.New("cmd")

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

// Used in construction of cobra.Command.
func PreRun(cobraCommand *cobra.Command, args []string) {
	cmdhelper.PreRun(cobraCommand, args, Use, ContextVariables)
}

// Used in construction of cobra.Command.
func RunE(_ *cobra.Command, _ []string) error {
	var err error

	ctx := context.Background()

	validator := &validate.BasicValidate{
		InputFileType: viper.GetString(option.InputFileType.Arg),
		InputURL:      viper.GetString(option.InputURL.Arg),
		JSONOutput:    viper.GetBool(option.JSONOutput.Arg),
		LogLevel:      viper.GetString(option.LogLevel.Arg),
	}

	if !validator.Read(ctx) {
		err = wraperror.Errorf(errPackage, "validation failed. error: %w", errPackage)
	}

	return err
}

// Used in construction of cobra.Command.
func Version() string {
	return cmdhelper.Version(githubVersion, githubIteration)
}

// ----------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------

// Since init() is always invoked, define command line parameters.
func init() {
	cmdhelper.Init(RootCmd, ContextVariables)
}
