package validate

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/senzing-garage/go-helpers/record"
	"github.com/senzing-garage/go-helpers/wraperror"
	"github.com/senzing-garage/go-logging/logging"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type BasicValidate struct {
	InputFileType string
	InputURL      string
	JSONOutput    bool
	logger        logging.Logging
	LogLevel      string
}

// ----------------------------------------------------------------------------
// Public methods
// ----------------------------------------------------------------------------

// using the information in the BasicValidate object read and validate that
// the records are valid.
func (validate *BasicValidate) Read(ctx context.Context) bool {
	// Initialize logging.
	logLevel := validate.LogLevel
	if logLevel == "" {
		logLevel = "INFO"
	}

	err := validate.SetLogLevel(ctx, logLevel)
	if err != nil {
		validate.log(3009, logLevel, err)
	}

	inputURLLen := len(validate.InputURL)

	if inputURLLen == 0 {
		// assume stdin
		return validate.ReadStdin()
	}

	// This assumes the URL includes a schema and path so, minimally:
	//  "s://p" where the schema is 's' and 'p' is the complete path
	if inputURLLen < 5 {
		validate.log(5000, validate.InputURL)

		return false
	}

	result := validate.validateBasedOnURL()

	return result
}

/*
The SetLogLevel method sets the level of logging.

Input
  - ctx: A context to control lifecycle.
  - logLevel: The desired log level. TRACE, DEBUG, INFO, WARN, ERROR, FATAL or PANIC.
*/
func (validate *BasicValidate) SetLogLevel(ctx context.Context, logLevelName string) error {
	_ = ctx

	var err error

	// Verify value of logLevelName.

	if !logging.IsValidLogLevelName(logLevelName) {
		return wraperror.Errorf(errPackage, "invalid error level: %s error: %w", logLevelName, errPackage)
	}

	// Set ValidateImpl log level.

	err = validate.getLogger().SetLogLevel(logLevelName)

	return wraperror.Errorf(err, "validate.SetLogLevel error: %w", err)
}

// ----------------------------------------------------------------------------
// Private methods
//  	response, err := http.Get(jsonURL) // #nosec:G107
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------

// opens and reads a JSONL resource.
func (validate *BasicValidate) ReadJSONLResource(jsonURL string) bool {
	//nolint:noctx
	response, err := http.Get(jsonURL) //nolint:gosec
	if err != nil {
		validate.log(5003, jsonURL, err)

		return false
	}

	defer response.Body.Close()
	validate.ValidateLines(response.Body)

	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL file.
func (validate *BasicValidate) ReadJSONLFile(jsonFile string) bool {
	jsonFile = filepath.Clean(jsonFile)

	file, err := os.Open(jsonFile)
	if err != nil {
		validate.log(5004, jsonFile, err)

		return false
	}

	defer file.Close()
	validate.ValidateLines(file)

	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL that has been piped to stdin.
func (validate *BasicValidate) ReadStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		validate.log(5005, err)

		return false
	}

	if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		reader := bufio.NewReader(os.Stdin)
		validate.ValidateLines(reader)

		return true
	}

	validate.log(5006, err)

	return false
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL resource that has been GZIPped.
func (validate *BasicValidate) ReadGZIPResource(gzURL string) bool {
	//nolint:noctx
	response, err := http.Get(gzURL) //nolint:gosec
	if err != nil {
		validate.log(5009, gzURL, err)

		return false
	}

	defer response.Body.Close()

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		validate.log(5010, gzURL, err)

		return false
	}

	defer reader.Close()
	validate.ValidateLines(reader)

	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL file that has been GZIPped.
func (validate *BasicValidate) ReadGZIPFile(gzFile string) bool {
	gzFile = filepath.Clean(gzFile)

	gzipfile, err := os.Open(gzFile)
	if err != nil {
		validate.log(5007, gzFile, err)

		return false
	}

	defer gzipfile.Close()

	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		validate.log(5008, gzFile, err)

		return false
	}

	defer reader.Close()
	validate.ValidateLines(reader)

	return true
}

// ----------------------------------------------------------------------------

// validate that each line read from the reader is a valid record.
func (validate *BasicValidate) ValidateLines(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	totalLines := 0
	noRecordID := 0
	noDataSource := 0
	malformed := 0
	badRecord := 0

	for scanner.Scan() {
		totalLines++
		str := strings.TrimSpace(scanner.Text())
		// ignore blank lines
		if len(str) > 0 {
			valid, err := record.Validate(str)
			if !valid {
				if err != nil {
					switch {
					case strings.Contains(err.Error(), "RECORD_ID"):
						validate.log(3005, totalLines)

						noRecordID++
					case strings.Contains(err.Error(), "DATA_SOURCE"):
						validate.log(3006, totalLines)

						noDataSource++
					case strings.Contains(err.Error(), "not well formed"):
						validate.log(3007, totalLines)

						malformed++
					default:
						validate.log(3008, totalLines)

						badRecord++
					}
				}
			}
		}
	}

	if noRecordID > 0 {
		validate.log(3001, noRecordID)
	}

	if noDataSource > 0 {
		validate.log(3002, noDataSource)
	}

	if malformed > 0 {
		validate.log(3003, malformed)
	}

	if badRecord > 0 {
		validate.log(3004, badRecord)
	}

	validate.log(2210, totalLines, noRecordID+noDataSource+malformed+badRecord)
}

// ----------------------------------------------------------------------------
// Private methods
// ----------------------------------------------------------------------------

func (validate *BasicValidate) validateBasedOnURL() bool {
	validate.log(2200, validate.InputURL)

	parsedURL, err := url.Parse(validate.InputURL)
	if err != nil {
		validate.log(5001, err)

		return false
	}

	switch parsedURL.Scheme {
	case "file":
		switch {
		case strings.HasSuffix(parsedURL.Path, "jsonl"), strings.ToUpper(validate.InputFileType) == "JSONL":
			validate.log(2201)

			return validate.ReadJSONLFile(parsedURL.Path)
		case strings.HasSuffix(parsedURL.Path, "gz"), strings.ToUpper(validate.InputFileType) == "GZ":
			validate.log(2203)

			return validate.ReadGZIPFile(parsedURL.Path)
		default:
			validate.log(5011)
		}
	case "http", "https":
		switch {
		case strings.HasSuffix(parsedURL.Path, "jsonl"), strings.ToUpper(validate.InputFileType) == "JSONL":
			validate.log(2204)

			return validate.ReadJSONLResource(validate.InputURL)
		case strings.HasSuffix(parsedURL.Path, "gz"), strings.ToUpper(validate.InputFileType) == "GZ":
			validate.log(2205)

			return validate.ReadGZIPResource(validate.InputURL)
		default:
			validate.log(5012)
		}
	default:
		validate.log(5002, parsedURL.Scheme)
	}

	return false
}

// ----------------------------------------------------------------------------
// Logging
// ----------------------------------------------------------------------------

// Get the Logger singleton.
func (validate *BasicValidate) getLogger() logging.Logging { //nolint
	var err error

	if validate.logger == nil {
		options := []interface{}{
			&logging.OptionCallerSkip{Value: 4},
		}

		validate.logger, err = logging.NewSenzingLogger(ComponentID, IDMessages, options...)
		if err != nil {
			panic(err)
		}
	}

	return validate.logger
}

// Log message.
func (validate *BasicValidate) log(messageNumber int, details ...interface{}) {
	if validate.JSONOutput {
		validate.getLogger().Log(messageNumber, details...)
	} else {
		fmt.Println(fmt.Sprintf(IDMessages[messageNumber], details...)) //nolint
	}
}
