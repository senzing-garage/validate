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
	"strings"

	"github.com/senzing/go-common/record"
	"github.com/senzing/go-logging/logging"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type ValidateImpl struct {
	InputFileType string
	InputUrl      string
	JsonOutput    bool
	logger        logging.LoggingInterface
	LogLevel      string
}

// ----------------------------------------------------------------------------
// Public methods
// ----------------------------------------------------------------------------

// using the information in the ValidateImpl object read and validate that
// the records are valid
func (v *ValidateImpl) Read(ctx context.Context) bool {

	// Initialize logging.

	logLevel := v.LogLevel
	if logLevel == "" {
		logLevel = "INFO"
	}
	err := v.SetLogLevel(ctx, logLevel)
	if err != nil {
		v.log(3009, logLevel, err)
	}

	InputUrlLen := len(v.InputUrl)

	if InputUrlLen == 0 {
		//assume stdin
		return v.readStdin()
	}

	//This assumes the URL includes a schema and path so, minimally:
	//  "s://p" where the schema is 's' and 'p' is the complete path
	if InputUrlLen < 5 {
		v.log(5000, v.InputUrl)
		return false
	}

	v.log(2200, v.InputUrl)
	u, err := url.Parse(v.InputUrl)
	if err != nil {
		v.log(5001, err)
		return false
	}
	if u.Scheme == "file" {
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			v.log(2201)
			return v.readJSONLFile(u.Path)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			v.log(2203)
			return v.readGZFile(u.Path)
		} else {
			v.log(5011)
		}
	} else if u.Scheme == "http" || u.Scheme == "https" {
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			v.log(2204)
			return v.readJSONLResource(v.InputUrl)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			v.log(2205)
			return v.readGZResource(v.InputUrl)
		} else {
			v.log(5012)
		}
	} else {
		v.log(5002, u.Scheme)
	}
	return false
}

/*
The SetLogLevel method sets the level of logging.

Input
  - ctx: A context to control lifecycle.
  - logLevel: The desired log level. TRACE, DEBUG, INFO, WARN, ERROR, FATAL or PANIC.
*/
func (v *ValidateImpl) SetLogLevel(ctx context.Context, logLevelName string) error {
	var err error = nil

	// Verify value of logLevelName.

	if !logging.IsValidLogLevelName(logLevelName) {
		return fmt.Errorf("invalid error level: %s", logLevelName)
	}

	// Set ValidateImpl log level.

	err = v.getLogger().SetLogLevel(logLevelName)
	return err
}

// ----------------------------------------------------------------------------
// Internal methods
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------

// opens and reads a JSONL resource
func (v *ValidateImpl) readJSONLResource(jsonURL string) bool {
	response, err := http.Get(jsonURL)

	if err != nil {
		v.log(5003, jsonURL, err)
		return false
	}
	defer response.Body.Close()
	v.validateLines(response.Body)
	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL file
func (v *ValidateImpl) readJSONLFile(jsonFile string) bool {
	file, err := os.Open(jsonFile)
	if err != nil {
		v.log(5004, jsonFile, err)
		return false
	}
	defer file.Close()
	v.validateLines(file)
	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL that has been piped to stdin
func (v *ValidateImpl) readStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		v.log(5005, err)
		return false
	}
	//printFileInfo(info)

	if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {

		reader := bufio.NewReader(os.Stdin)
		v.validateLines(reader)
		return true
	}
	v.log(5006, err)
	return false
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL resource that has been Gzipped
func (v *ValidateImpl) readGZResource(gzURL string) bool {
	response, err := http.Get(gzURL)
	if err != nil {
		v.log(5009, gzURL, err)
		return false
	}
	defer response.Body.Close()
	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		v.log(5010, gzURL, err)
		return false
	}
	defer reader.Close()
	v.validateLines(reader)
	return true
}

// ----------------------------------------------------------------------------

// opens and reads a JSONL file that has been Gzipped
func (v *ValidateImpl) readGZFile(gzFile string) bool {
	gzipfile, err := os.Open(gzFile)
	if err != nil {
		v.log(5007, gzFile, err)
		return false
	}
	defer gzipfile.Close()

	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		v.log(5008, gzFile, err)
		return false
	}
	defer reader.Close()
	v.validateLines(reader)
	return true
}

// ----------------------------------------------------------------------------

// validate that each line read from the reader is a valid record
func (v *ValidateImpl) validateLines(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	totalLines := 0
	noRecordId := 0
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
					if strings.Contains(err.Error(), "RECORD_ID") {
						v.log(3005, totalLines)
						noRecordId++
					} else if strings.Contains(err.Error(), "DATA_SOURCE") {
						v.log(3006, totalLines)
						noDataSource++
					} else if strings.Contains(err.Error(), "not well formed") {
						v.log(3007, totalLines)
						malformed++
					} else {
						v.log(3008, totalLines)
						badRecord++
					}
				}
			}
		}
	}
	if noRecordId > 0 {
		v.log(3001, noRecordId)
	}
	if noDataSource > 0 {
		v.log(3002, noDataSource)
	}
	if malformed > 0 {
		v.log(3003, malformed)
	}
	if badRecord > 0 {
		v.log(3004, badRecord)
	}
	v.log(2210, totalLines, noRecordId+noDataSource+malformed+badRecord)
}

// ----------------------------------------------------------------------------
// Logging --------------------------------------------------------------------

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

var debugOptions []interface{} = []interface{}{
	&logging.OptionCallerSkip{Value: 5},
}

var traceOptions []interface{} = []interface{}{
	&logging.OptionCallerSkip{Value: 5},
}

// ----------------------------------------------------------------------------
// logger methods

// Get the Logger singleton.
func (v *ValidateImpl) getLogger() logging.LoggingInterface {
	var err error = nil
	if v.logger == nil {
		options := []interface{}{
			&logging.OptionCallerSkip{Value: 4},
		}
		v.logger, err = logging.NewSenzingToolsLogger(ComponentId, IdMessages, options...)
		if err != nil {
			panic(err)
		}
	}
	return v.logger
}

// Log message.
func (v *ValidateImpl) log(messageNumber int, details ...interface{}) {
	if v.JSONOutput {
		v.getLogger().Log(messageNumber, details...)
	} else {
		fmt.Println(fmt.Sprintf(IdMessages[messageNumber], details...))
	}
}

// Debug.
func (v *ValidateImpl) debug(messageNumber int, details ...interface{}) {
	details = append(details, debugOptions...)
	v.getLogger().Log(messageNumber, details...)
}

// Trace method entry.
func (v *ValidateImpl) traceEntry(messageNumber int, details ...interface{}) {
	details = append(details, traceOptions...)
	v.getLogger().Log(messageNumber, details...)
}

// Trace method exit.
func (v *ValidateImpl) traceExit(messageNumber int, details ...interface{}) {
	details = append(details, traceOptions...)
	v.getLogger().Log(messageNumber, details...)
}
