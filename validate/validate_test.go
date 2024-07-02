//go:build !windows
// +build !windows

package validate

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// test Read method
// ----------------------------------------------------------------------------

// read jsonl file successfully, no record validation errors
func TestBasicValidate_Read(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// read jsonl file successully, but with record validation errors
func TestBasicValidate_Read_with_bad_records(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 16 lines, 4 were bad"
	if !strings.Contains(got, want) && result == true {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// read jsonl file successfully, but attept to set a bad log level
// falls back to INFO
func TestBasicValidate_Read_bad_loglevel(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
		LogLevel: "BAD",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Unable to set log level to BAD"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// attempt to read a jsonl file, but the input url is bad
func TestBasicValidate_Read_bad_url(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{
		InputURL: "BAD",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error, Check the input-url parameter: BAD"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the input url isn't parsable
func TestBasicValidate_Read_bad_url_parse(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{
		InputURL: "http://bad:bad{BAD=bad@example.com",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error parsing input-url"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the input url is not understood
func TestBasicValidate_Read_url_drop_through(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{
		InputURL: "BAD,Really bad",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error unable to handle"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the file doesn't exist
func TestBasicValidate_Read_file_doesnt_exist(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{
		InputURL: "file:///badfile.jsonl",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error opening input file: /badfile.jsonl"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

func TestBasicValidate_Read_stdin_unpipe_error(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		test.Fatal(err)
	}
	os.Stdin = file
	defer func() {
		if err := file.Close(); err != nil {
			test.Fatal(err)
		}
	}()

	validator := &BasicValidate{}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a file, but it has a file type that is not known
func TestBasicValidate_Read_bad_file_type(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "txt")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type)"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a file type that is not known, but override with input file type
func TestBasicValidate_Read_override_file_type(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "txt")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputFileType: "JSONL",
		InputURL:      fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// ----------------------------------------------------------------------------
// test Read with .gz files
// ----------------------------------------------------------------------------

// read a gz file successfully, with no record validation errors
func TestBasicValidate_Read_gz(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempGZIPDataFile(test, testGoodData)
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// read a gz file successfully, but with record validation errors
func TestBasicValidate_Read_gz_bad(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempGZIPDataFile(test, testBadData)
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 16 lines, 4 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}
}

// ----------------------------------------------------------------------------
// test Read resources
// ----------------------------------------------------------------------------

func TestBasicValidate_Read_resource_jsonl(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, cleanUpTempFile := createTempDataFile(test, testGoodData, "jsonl")
	defer cleanUpTempFile()
	server, listener, port := serveResource(test, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		test.Error(err)
	}
}

func TestBasicValidate_Read_resource_unknown_extension(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, cleanUpTempFile := createTempDataFile(test, testGoodData, "bad")
	defer cleanUpTempFile()
	server, listener, port := serveResource(test, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "If this is a valid JSONL resource"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		test.Error(err)
	}
}

func TestBasicValidate_Read_resource_bad_url(tes *testing.T) {

	r, w, cleanUp := mockStdout(tes)
	defer cleanUp()

	filename, cleanUpTempFile := createTempDataFile(tes, testGoodData, "jsonl")
	defer cleanUpTempFile()
	server, listener, _ := serveResource(tes, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:4444444/%s", "bad.jsonl"),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error retrieving input-url"
	if !strings.Contains(got, want) {
		tes.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		tes.Errorf("ValidateImpl.Read() = %v, want false", result)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		tes.Error(err)
	}
}

func TestBasicValidate_Read_resource_gzip(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, cleanUpTempFile := createTempGZIPDataFile(test, testGoodData)
	defer cleanUpTempFile()
	server, listener, port := serveResource(test, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.Read() = %v, want true", result)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		test.Error(err)
	}
}

func TestBasicValidate_Read_resource_gzip_bad_url(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, cleanUpTempFile := createTempGZIPDataFile(test, testGoodData)
	defer cleanUpTempFile()
	server, listener, _ := serveResource(test, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:44444444/%s", "bad.gz"),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error retrieving GZIPped input-url"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.Read() = %v, want false", result)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		test.Error(err)
	}
}

func TestBasicValidate_Read_resource_gzip_not_gzipped(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, cleanUpTempFile := createTempDataFile(test, testGoodData, "gz")
	defer cleanUpTempFile()
	server, listener, port := serveResource(test, filename)
	go func() {
		if err := server.Serve(*listener); err != http.ErrServerClosed {
			log.Fatalf("server.Serve(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &BasicValidate{
		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error reading GZIPped input-url"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result {
		test.Errorf("ValidateImpl.Read() = %v, want %v", result, false)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		test.Error(err)
	}
}

// ----------------------------------------------------------------------------
// test Read with Json output
// ----------------------------------------------------------------------------

// read a json file successfully, with no record validation errors
// func TestBasicValidate_Read_jsonOutput(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &BasicValidate{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JSONOutput: true,
// 		LogLevel:   "WARN",
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		test.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}

// }

// read a json file successfully, but with record validation errors
// func TestBasicValidate_Read_jsonOutput_bad(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(test, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &BasicValidate{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JSONOutput: true,
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		test.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// ----------------------------------------------------------------------------
// test jsonl file read with json output
// ----------------------------------------------------------------------------

// read a json file successfully, with no record validation errors
func TestBasicValidate_readJsonlFile(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readJSONLFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readJSONLFile() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.readJSONLFile() = %v, want true", result)
	}

}

// read a json file successfully, but with record validation errors
func TestBasicValidate_readJsonlFile_bad(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readJSONLFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 16 lines, 4 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readJSONLFile() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.readJSONLFile() = %v, want true", result)
	}
}

// read a json file successfully, with no record validation errors
// func TestBasicValidate_readJsonlFile_jsonOutput(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &BasicValidate{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JSONOutput: true,
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		test.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// read a json file successfully, but with record validation errors
// func TestBasicValidate_readJsonlFile_jsonOutput_bad(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(test, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &BasicValidate{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JSONOutput: true,
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		test.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// ----------------------------------------------------------------------------
// test gzip file read
// ----------------------------------------------------------------------------

// read a gzip file successfully, no record validation errors
func TestBasicValidate_readGzipFile(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempGZIPDataFile(test, testGoodData)
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZIPFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want true", result)
	}
}

// read a gzip file successfully, but with record validation errors
func TestBasicValidate_readGzipFile_bad(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempGZIPDataFile(test, testBadData)
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZIPFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 16 lines, 4 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
	}
	if result != true {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want true", result)
	}
}

// attempt to read a gzip file that doesn't exist
func TestBasicValidate_readGzipFile_file_does_not_exist(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename := "/bad.gz"

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZIPFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "no such file or directory"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want false", result)
	}
}

// attempt to read a gzip file that isn't a gzip file
func TestBasicValidate_readGzipFile_not_a_gzip_file(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testBadData, "gz")
	defer moreCleanUp()

	validator := &BasicValidate{
		InputURL: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZIPFile(filename)

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "invalid header"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.readGZIPFile() = %v, want false", result)
	}
}

// ----------------------------------------------------------------------------
// test read stdin
// ----------------------------------------------------------------------------

func TestBasicValidate_readStdin(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		test.Fatal(err)
	}

	os.Stdin = file

	defer func() {
		if err := file.Close(); err != nil {
			test.Fatal(err)
		}
	}()

	validator := &BasicValidate{}
	result := validator.readStdin()

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readStdin() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.readStdin() = %v, want false", result)
	}
}
func TestBasicValidate_readStdin_unpipe_error(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(test, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		test.Fatal(err)
	}
	os.Stdin = file

	defer func() {
		if err := file.Close(); err != nil {
			test.Fatal(err)
		}
	}()

	validator := &BasicValidate{}
	result := validator.readStdin()

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.readStdin() = %v, want %v", got, want)
	}
	if result == true {
		test.Errorf("ValidateImpl.readStdin() = %v, want false", result)
	}
}

// ----------------------------------------------------------------------------
// test validateLines
// ----------------------------------------------------------------------------

// validate lines with no record validation errors
func TestBasicValidate_validateLines(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{}
	validator.validateLines(strings.NewReader(testGoodData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		test.Errorf("ValidateImpl.validateLines() = %v, want %v", got, want)
	}
}

// validate lines, but with record validation errors
func TestBasicValidate_validateLines_with_validation_errors(test *testing.T) {

	r, w, cleanUp := mockStdout(test)
	defer cleanUp()

	validator := &BasicValidate{}
	validator.validateLines(strings.NewReader(testBadData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(test, got, msg)
}

// validate lines with no record validation errors, json output
// func TestBasicValidate_validateLines_jsonOutput(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	validator := &BasicValidate{
// 		JSONOutput: true,
// 	}
// 	validator.validateLines(strings.NewReader(testGoodData))

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// }

// validate lines, but with record validation errors and json output
// func TestBasicValidate_validateLines_with_validation_errors_jsonOutput(test *testing.T) {

// 	r, w, cleanUp := mockStderr(test)
// 	defer cleanUp()

// 	validator := &BasicValidate{
// 		JSONOutput: true,
// 	}
// 	validator.validateLines(strings.NewReader(testBadData))

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		test.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// }

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

// create a tempdata file with the given content and extension
func createTempDataFile(test *testing.T, content string, fileextension string) (filename string, cleanUp func()) {
	test.Helper()
	tmpfile, err := os.CreateTemp(test.TempDir(), "test.*."+fileextension)
	require.NoError(test, err)

	_, err = tmpfile.WriteString(content)
	require.NoError(test, err)

	filename = tmpfile.Name()

	err = tmpfile.Close()
	require.NoError(test, err)

	return filename,
		func() {
			err := os.Remove(filename)
			require.NoError(test, err)
		}
}

// create a temp gzipped datafile with the given content
func createTempGZIPDataFile(test *testing.T, content string) (filename string, cleanUp func()) {
	test.Helper()
	tmpfile, err := os.CreateTemp("", "test.*.jsonl.gz")
	require.NoError(test, err)
	defer tmpfile.Close()
	gf := gzip.NewWriter(tmpfile)
	defer gf.Close()
	fw := bufio.NewWriter(gf)
	_, err = fw.WriteString(content)
	require.NoError(test, err)
	fw.Flush()
	filename = tmpfile.Name()
	return filename,
		func() {
			err := os.Remove(filename)
			require.NoError(test, err)
		}
}

// serve the requested resource on a random port
func serveResource(test *testing.T, filename string) (*http.Server, *net.Listener, int) {
	test.Helper()
	listener, err := net.Listen("tcp", ":0") // #nosec:G102
	require.NoError(test, err)
	port := listener.Addr().(*net.TCPAddr).Port
	idx := strings.LastIndex(filename, string(os.PathSeparator))
	fs := http.FileServer(http.Dir(filename[:idx]))
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           fs,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &server, &listener, port

}

// capture stdout for testing
func mockStdout(test *testing.T) (reader *os.File, writer *os.File, cleanUp func()) {
	test.Helper()

	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoErrorf(test, err, "couldn't get os Pipe: %v", err)
	os.Stdout = writer

	return reader,
		writer,
		func() {
			// clean-up
			os.Stdout = origStdout
		}
}

// capture stderr for testing
// func mockStderr(test *testing.T) (reader *os.File, writer *os.File, cleanUp func()) {
// 	test.Helper()
// 	origStderr := os.Stderr
// 	reader, writer, err := os.Pipe()
// 	if err != nil {
// 		assert.Fail(test, "couldn't get os Pipe: %v", err)
// 	}
// 	os.Stderr = writer

// 	return reader,
// 		writer,
// 		func() {
// 			// clean-up
// 			os.Stderr = origStderr
// 		}
// }

var testGoodData = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000002", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "SUITE E-2,UNION COURT BUILDING, P.O. BOX N-8188, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000002"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000003", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, LYFORD CAY, P.O. BOX N-7785, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000003"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000004", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "P.O. BOX N-3708 BAHAMAS FINANCIAL CENTRE, P.O. BOX N-3708 SHIRLEY & CHARLOTTE STS, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000004"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000006", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "303 SHIRLEY STREET, P.O. BOX N-492, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000006"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000007", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "OCEAN CENTRE, MONTAGU FORESHORE, P.O. BOX SS-19084 EAST BAY STREET, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000007"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000008", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "PROVIDENCE HOUSE, EAST WING EAST HILL ST, P.O. BOX CB-12399, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000008"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000009", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "BAYSIDE EXECUTIVE PARK, WEST BAY & BLAKE, P.O. BOX N-4875, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000009"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"SOCIAL_HANDLE": "shuddersv", "DATE_OF_BIRTH": "16/7/1974", "ADDR_STATE": "NC", "ADDR_POSTAL_CODE": "257609", "ENTITY_TYPE": "TEST", "GENDER": "F", "srccode": "MDMPER", "RECORD_ID": "151110080", "DSRC_ACTION": "A", "ADDR_CITY": "Raleigh", "DRIVERS_LICENSE_NUMBER": "95", "PHONE_NUMBER": "984-881-8384", "NAME_LAST": "OBERMOELLER", "entityid": "151110080", "ADDR_LINE1": "3802 eBllevue RD", "DATA_SOURCE": "TEST"}
{"SOCIAL_HANDLE": "battlesa", "ADDR_STATE": "LA", "ADDR_POSTAL_CODE": "70706", "NAME_FIRST": "DEVIN", "ENTITY_TYPE": "TEST", "GENDER": "M", "srccode": "MDMPER", "CC_ACCOUNT_NUMBER": "5018608175414044187", "RECORD_ID": "151267101", "DSRC_ACTION": "A", "ADDR_CITY": "Denham Springs", "DRIVERS_LICENSE_NUMBER": "614557601", "PHONE_NUMBER": "318-398-0649", "NAME_LAST": "LOVELL", "entityid": "151267101", "ADDR_LINE1": "8487 Ashley ", "DATA_SOURCE": "TEST"}
`
var testBadData = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000002", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "SUITE E-2,UNION COURT BUILDING, P.O. BOX N-8188, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000002"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000003", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, LYFORD CAY, P.O. BOX N-7785, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000003"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000004", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "P.O. BOX N-3708 BAHAMAS FINANCIAL CENTRE, P.O. BOX N-3708 SHIRLEY & CHARLOTTE STS, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000004"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005B" "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000006", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "303 SHIRLEY STREET, P.O. BOX N-492, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000006"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000007", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "OCEAN CENTRE, MONTAGU FORESHORE, P.O. BOX SS-19084 EAST BAY STREET, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000007"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000008", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "PROVIDENCE HOUSE, EAST WING EAST HILL ST, P.O. BOX CB-12399, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000008"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000009", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "BAYSIDE EXECUTIVE PARK, WEST BAY & BLAKE, P.O. BOX N-4875, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000009"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010B" "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"SOCIAL_HANDLE": "shuddersv", "DATE_OF_BIRTH": "16/7/1974", "ADDR_STATE": "NC", "ADDR_POSTAL_CODE": "257609", "ENTITY_TYPE": "TEST", "GENDER": "F", "srccode": "MDMPER", "RECORD_ID": "151110080", "DSRC_ACTION": "A", "ADDR_CITY": "Raleigh", "DRIVERS_LICENSE_NUMBER": "95", "PHONE_NUMBER": "984-881-8384", "NAME_LAST": "OBERMOELLER", "entityid": "151110080", "ADDR_LINE1": "3802 eBllevue RD", "DATA_SOURCE": "TEST"}
{"SOCIAL_HANDLE": "battlesa", "ADDR_STATE": "LA", "ADDR_POSTAL_CODE": "70706", "NAME_FIRST": "DEVIN", "ENTITY_TYPE": "TEST", "GENDER": "M", "srccode": "MDMPER", "CC_ACCOUNT_NUMBER": "5018608175414044187", "RECORD_ID": "151267101", "DSRC_ACTION": "A", "ADDR_CITY": "Denham Springs", "DRIVERS_LICENSE_NUMBER": "614557601", "PHONE_NUMBER": "318-398-0649", "NAME_LAST": "LOVELL", "entityid": "151267101", "ADDR_LINE1": "8487 Ashley ", "DATA_SOURCE": "TEST"}
`
