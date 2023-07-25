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
	InputURL      string
	logger        logging.LoggingInterface
	LogLevel      string
}

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
func (v *ValidateImpl) Read(ctx context.Context) bool {

	inputURLLen := len(v.InputURL)

	if inputURLLen == 0 {
		//assume stdin
		return v.readStdin()
	}

	//This assumes the URL includes a schema and path so, minimally:
	//  "s://p" where the schema is 's' and 'p' is the complete path
	if inputURLLen < 5 {
		v.log(2002, v.InputURL)
		// logger.LogMessage(MessageIdFormat, 2002, fmt.Sprintf("Check the inputURL parameter: %s", v.InputURL))
		return false
	}

	// logger.LogMessage(MessageIdFormat, 1200, fmt.Sprintf("Validating URL string: %s", v.InputURL))
	v.log(1200, v.InputURL)
	u, err := url.Parse(v.InputURL)
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9001, "Fatal error parsing inputURL.", err)
		v.log(9001, err)
		return false
	}
	if u.Scheme == "file" {
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			// logger.LogMessage(MessageIdFormat, 1201, "Validating as a JSONL file.")
			v.log(1201, nil)
			return v.readJSONLFile(u.Path)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			// logger.LogMessage(MessageIdFormat, 1202, "Validating a GZ file.")
			v.log(1202, nil)
			return v.readGZFile(u.Path)
		} else {
			// logger.LogMessage(MessageIdFormat, 2003, "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--fileType).")
			v.log(2003, nil)
		}
	} else if u.Scheme == "http" || u.Scheme == "https" {
		fmt.Println("scheme:", u.Scheme)
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			// logger.LogMessage(MessageIdFormat, 1204, "Validating as a JSONL resource.")
			v.log(1204, nil)
			return v.readJSONLResource(v.InputURL)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			// logger.LogMessage(MessageIdFormat, 1205, "Validating a GZ resource.")
			v.log(1205, nil)
			return v.readGZResource(v.InputURL)
		} else {
			// logger.LogMessage(MessageIdFormat, 2004, "If this is a valid JSONL resource, please rename with the .jsonl extension or use the file type override (--file-type).")
			v.log(2004, nil)
		}
	} else {
		// logger.LogMessage(MessageIdFormat, 9002, fmt.Sprintf("We don't handle %s input URLs.", u.Scheme))
		v.log(9002, u.Scheme)
	}
	return false
}

// ----------------------------------------------------------------------------
// Internal methods
// ----------------------------------------------------------------------------

// --- Logging ----------------------------------------------------------------

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
	v.getLogger().Log(messageNumber, details...)
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

// ----------------------------------------------------------------------------
func (v *ValidateImpl) readJSONLResource(jsonURL string) bool {
	response, err := http.Get(jsonURL)

	if err != nil {
		fmt.Println("unable to get:", jsonURL)
		// logger.LogMessageFromError(MessageIdFormat, 9003, "Fatal error retrieving inputURL.", err)
		v.log(9003, jsonURL, err)
		return false
	}
	defer response.Body.Close()
	v.validateLines(response.Body)
	return true
}

// ----------------------------------------------------------------------------
func (v *ValidateImpl) readJSONLFile(jsonFile string) bool {
	file, err := os.Open(jsonFile)
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9004, "Fatal error opening inputURL.", err)
		v.log(9004, jsonFile, err)
		return false
	}
	defer file.Close()
	v.validateLines(file)
	return true
}

// ----------------------------------------------------------------------------
func (v *ValidateImpl) readStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9005, "Fatal error opening stdin.", err)
		v.log(9005, err)
		return false
	}
	//printFileInfo(info)

	if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {

		reader := bufio.NewReader(os.Stdin)
		v.validateLines(reader)
		return true
	}
	// logger.LogMessageFromError(MessageIdFormat, 9006, "Fatal error stdin not piped.", err)
	v.log(9006, err)
	return false
}

// ----------------------------------------------------------------------------
func (v *ValidateImpl) readGZResource(gzURL string) bool {
	response, err := http.Get(gzURL)
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9009, "Fatal error retrieving inputURL.", err)
		v.log(9009, gzURL, err)
		return false
	}
	defer response.Body.Close()
	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9010, "Fatal error reading inputURL.", err)
		v.log(9010, gzURL, err)
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
		// logger.LogMessageFromError(MessageIdFormat, 9007, "Fatal error opening inputURL.", err)
		v.log(9007, gzFile, err)
		return false
	}
	defer gzipfile.Close()

	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		// logger.LogMessageFromError(MessageIdFormat, 9008, "Fatal error reading inputURL.", err)
		v.log(9008, gzFile, err)
		return false
	}
	defer reader.Close()
	v.validateLines(reader)
	return true
}

// ----------------------------------------------------------------------------
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
				fmt.Println("Line", totalLines, err)
				if err != nil {
					if strings.Contains(err.Error(), "RECORD_ID") {
						noRecordId++
					} else if strings.Contains(err.Error(), "DATA_SOURCE") {
						noDataSource++
					} else if strings.Contains(err.Error(), "not well formed") {
						malformed++
					} else {
						badRecord++
					}
				}
			}
		}
	}
	if noRecordId > 0 {
		// logger.LogMessage(MessageIdFormat, 1206, fmt.Sprintf("%d line(s) had no RECORD_ID field.", noRecordId))
		v.log(1206, noRecordId)
	}
	if noDataSource > 0 {
		// logger.LogMessage(MessageIdFormat, 1207, fmt.Sprintf("%d line(s) had no DATA_SOURCE field.", noDataSource))
		v.log(1207, noDataSource)
	}
	if malformed > 0 {
		// logger.LogMessage(MessageIdFormat, 1208, fmt.Sprintf("%d line(s) are not well formed JSON-lines.", malformed))
		v.log(1208, malformed)
	}
	if badRecord > 0 {
		// logger.LogMessage(MessageIdFormat, 1209, fmt.Sprintf("%d line(s) did not validate for an unknown reason.", badRecord))
		v.log(1209, badRecord)
	}
	// logger.LogMessage(MessageIdFormat, 1210, fmt.Sprintf("Validated %d lines, %d were bad.", totalLines, noRecordId+noDataSource+malformed+badRecord))
	v.log(1210, totalLines, noRecordId+noDataSource+malformed+badRecord)
}

// ----------------------------------------------------------------------------
func (v *ValidateImpl) printFileInfo(info os.FileInfo) {
	fmt.Println("name: ", info.Name())
	fmt.Println("size: ", info.Size())
	fmt.Println("mode: ", info.Mode())
	fmt.Println("mod time: ", info.ModTime())
	fmt.Println("is dir: ", info.IsDir())
	if info.Mode()&os.ModeDevice == os.ModeDevice {
		fmt.Println("detected device: ", os.ModeDevice)
	}
	if info.Mode()&os.ModeCharDevice == os.ModeCharDevice {
		fmt.Println("detected char device: ", os.ModeCharDevice)
	}
	if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		fmt.Println("detected named pipe: ", os.ModeNamedPipe)
	}
	fmt.Printf("\n\n")
}
