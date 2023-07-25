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
		return false
	}

	v.log(2200, v.InputURL)
	u, err := url.Parse(v.InputURL)
	if err != nil {
		v.log(4001, err)
		return false
	}
	if u.Scheme == "file" {
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			v.log(2201, nil)
			return v.readJSONLFile(u.Path)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			v.log(2203, nil)
			return v.readGZFile(u.Path)
		} else {
			v.log(2003, nil)
		}
	} else if u.Scheme == "http" || u.Scheme == "https" {
		if strings.HasSuffix(u.Path, "jsonl") || strings.ToUpper(v.InputFileType) == "JSONL" {
			v.log(2204, nil)
			return v.readJSONLResource(v.InputURL)
		} else if strings.HasSuffix(u.Path, "gz") || strings.ToUpper(v.InputFileType) == "GZ" {
			v.log(2205, nil)
			return v.readGZResource(v.InputURL)
		} else {
			v.log(2004, nil)
		}
	} else {
		v.log(4002, u.Scheme)
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
		v.log(4003, jsonURL, err)
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
		v.log(4004, jsonFile, err)
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
		v.log(4005, err)
		return false
	}
	//printFileInfo(info)

	if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {

		reader := bufio.NewReader(os.Stdin)
		v.validateLines(reader)
		return true
	}
	v.log(4006, err)
	return false
}

// ----------------------------------------------------------------------------
func (v *ValidateImpl) readGZResource(gzURL string) bool {
	response, err := http.Get(gzURL)
	if err != nil {
		v.log(4009, gzURL, err)
		return false
	}
	defer response.Body.Close()
	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		v.log(4010, gzURL, err)
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
		v.log(4007, gzFile, err)
		return false
	}
	defer gzipfile.Close()

	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		v.log(4008, gzFile, err)
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
				// fmt.Println("Line", totalLines, err)
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
		v.log(2206, noRecordId)
	}
	if noDataSource > 0 {
		v.log(2207, noDataSource)
	}
	if malformed > 0 {
		v.log(2208, malformed)
	}
	if badRecord > 0 {
		v.log(2209, badRecord)
	}
	v.log(2210, totalLines, noRecordId+noDataSource+malformed+badRecord)
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
