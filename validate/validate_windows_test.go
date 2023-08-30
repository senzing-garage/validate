//go:build windows
// +build windows

package validate

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------------------
// test Read method
// ----------------------------------------------------------------------------

// // read jsonl file successfully, no record validation errors
// func TestValidateImpl_Read(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // read jsonl file successully, but with record validation errors
// func TestValidateImpl_Read_with_bad_records(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) && result == true {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // read jsonl file successfully, but attept to set a bad log level
// // falls back to INFO
// func TestValidateImpl_Read_bad_loglevel(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 		LogLevel: "BAD",
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Unable to set log level to BAD"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// attempt to read a jsonl file, but the input url is bad
func TestValidateImpl_Read_bad_url(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputURL: "BAD",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error, Check the input-url parameter: BAD"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the input url isn't parsable
func TestValidateImpl_Read_bad_url_parse(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputURL: "http://bad:bad{BAD=bad@example.com",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error parsing input-url"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the input url is not understood
func TestValidateImpl_Read_url_drop_through(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputURL: "BAD,Really bad",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error unable to handle"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// attempt to read a jsonl file, but the file doesn't exist
func TestValidateImpl_Read_file_doesnt_exist(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputURL: "file:///badfile.jsonl",
	}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error opening input file: /badfile.jsonl"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

func TestValidateImpl_Read_stdin_unpipe_error(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = file
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	validator := &ValidateImpl{}
	result := validator.Read(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.Read() = %v, want false", result)
	}
}

// // // attempt to read a file, but it has a file type that is not known
// // func TestValidateImpl_Read_bad_file_type(t *testing.T) {

// // 	r, w, cleanUp := mockStdout(t)
// // 	defer cleanUp()

// // 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "txt")
// // 	defer moreCleanUp()

// // 	validator := &ValidateImpl{
// // 		InputURL: fmt.Sprintf("file://%s", filename),
// // 	}
// // 	result := validator.Read(context.Background())

// // 	w.Close()
// // 	out, _ := io.ReadAll(r)
// // 	got := string(out)

// // 	want := "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type)"
// // 	if !strings.Contains(got, want) {
// // 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// // 	}
// // 	if result == true {
// // 		t.Errorf("ValidateImpl.Read() = %v, want false", result)
// // 	}
// // }

// // attempt to read a file type that is not known, but override with input file type
// func TestValidateImpl_Read_override_file_type(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "txt")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputFileType: "JSONL",
// 		InputURL:      fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // ----------------------------------------------------------------------------
// // test Read with .gz files
// // ----------------------------------------------------------------------------

// // read a gz file successfully, with no record validation errors
// func TestValidateImpl_Read_gz(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempGZIPDataFile(t, testGoodData)
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // read a gz file successfully, but with record validation errors
// func TestValidateImpl_Read_gz_bad(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempGZIPDataFile(t, testBadData)
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // ----------------------------------------------------------------------------
// // test Read resources
// // ----------------------------------------------------------------------------

// func TestValidateImpl_Read_resource_jsonl(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "jsonl")
// 	defer cleanUpTempFile()
// 	server, listener, port := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	idx := strings.LastIndex(filename, "/")
// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestValidateImpl_Read_resource_unknown_extension(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "bad")
// 	defer cleanUpTempFile()
// 	server, listener, port := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	idx := strings.LastIndex(filename, "/")
// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "If this is a valid JSONL resource"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result == true {
// 		t.Errorf("ValidateImpl.Read() = %v, want false", result)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestValidateImpl_Read_resource_bad_url(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "jsonl")
// 	defer cleanUpTempFile()
// 	server, listener, _ := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:4444444/%s", "bad.jsonl"),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Fatal error retrieving input-url"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result == true {
// 		t.Errorf("ValidateImpl.Read() = %v, want false", result)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestValidateImpl_Read_resource_gzip(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempGZIPDataFile(t, testGoodData)
// 	defer cleanUpTempFile()
// 	server, listener, port := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	idx := strings.LastIndex(filename, "/")
// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestValidateImpl_Read_resource_gzip_bad_url(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempGZIPDataFile(t, testGoodData)
// 	defer cleanUpTempFile()
// 	server, listener, _ := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:44444444/%s", "bad.gz"),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Fatal error retrieving GZIPped input-url"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result == true {
// 		t.Errorf("ValidateImpl.Read() = %v, want false", result)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestValidateImpl_Read_resource_gzip_not_gzipped(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "gz")
// 	defer cleanUpTempFile()
// 	server, listener, port := serveResource(t, filename)
// 	go func() {
// 		if err := server.Serve(*listener); err != http.ErrServerClosed {
// 			log.Fatalf("server.Serve(): %v", err)
// 		}
// 	}()

// 	idx := strings.LastIndex(filename, "/")
// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("http://localhost:%d/%s", port, filename[(idx+1):]),
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Fatal error reading GZIPped input-url"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", result, false)
// 	}

// 	if err := server.Shutdown(context.Background()); err != nil {
// 		t.Error(err)
// 	}
// }

// // ----------------------------------------------------------------------------
// // test Read with Json output
// // ----------------------------------------------------------------------------

// // read a json file successfully, with no record validation errors
// func TestValidateImpl_Read_jsonOutput(t *testing.T) {

// 	r, w, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}

// }

// // read a json file successfully, but with record validation errors
// func TestValidateImpl_Read_jsonOutput_bad(t *testing.T) {

// 	r, w, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	result := validator.Read(context.Background())

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // ----------------------------------------------------------------------------
// // test jsonl file read with json output
// // ----------------------------------------------------------------------------

// // read a json file successfully, with no record validation errors
// func TestValidateImpl_readJsonlFile(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readJSONLFile() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.readJSONLFile() = %v, want true", result)
// 	}

// }

// // read a json file successfully, but with record validation errors
// func TestValidateImpl_readJsonlFile_bad(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readJSONLFile() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.readJSONLFile() = %v, want true", result)
// 	}
// }

// // read a json file successfully, with no record validation errors
// func TestValidateImpl_readJsonlFile_jsonOutput(t *testing.T) {

// 	r, w, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // read a json file successfully, but with record validation errors
// func TestValidateImpl_readJsonlFile_jsonOutput_bad(t *testing.T) {

// 	r, w, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	result := validator.readJSONLFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.Read() = %v, want true", result)
// 	}
// }

// // ----------------------------------------------------------------------------
// // test gzip file read
// // ----------------------------------------------------------------------------

// // read a gzip file successfully, no record validation errors
// func TestValidateImpl_readGzipFile(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempGZIPDataFile(t, testGoodData)
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readGZIPFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 12 lines, 0 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want true", result)
// 	}
// }

// // read a gzip file successfully, but with record validation errors
// func TestValidateImpl_readGzipFile_bad(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempGZIPDataFile(t, testBadData)
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readGZIPFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "Validated 16 lines, 4 were bad"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
// 	}
// 	if result != true {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want true", result)
// 	}
// }

// // attempt to read a gzip file that doesn't exist
// func TestValidateImpl_readGzipFile_file_does_not_exist(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename := "/bad.gz"

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readGZIPFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "no such file or directory"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
// 	}
// 	if result == true {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want false", result)
// 	}
// }

// // attempt to read a gzip file that isn't a gzip file
// func TestValidateImpl_readGzipFile_not_a_gzip_file(t *testing.T) {

// 	r, w, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "gz")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputURL: fmt.Sprintf("file://%s", filename),
// 	}
// 	result := validator.readGZIPFile(filename)

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	got := string(out)

// 	want := "invalid header"
// 	if !strings.Contains(got, want) {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want %v", got, want)
// 	}
// 	if result == true {
// 		t.Errorf("ValidateImpl.readGZIPFile() = %v, want false", result)
// 	}
// }

// ----------------------------------------------------------------------------
// test read stdin
// ----------------------------------------------------------------------------

func TestValidateImpl_readStdin(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = file

	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	validator := &ValidateImpl{}
	result := validator.readStdin()

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.readStdin() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.readStdin() = %v, want false", result)
	}
}
func TestValidateImpl_readStdin_unpipe_error(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = file

	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	validator := &ValidateImpl{}
	result := validator.readStdin()

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Fatal error stdin not piped"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.readStdin() = %v, want %v", got, want)
	}
	if result == true {
		t.Errorf("ValidateImpl.readStdin() = %v, want false", result)
	}
}

// ----------------------------------------------------------------------------
// test validateLines
// ----------------------------------------------------------------------------

// validate lines with no record validation errors
func TestValidateImpl_validateLines(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testGoodData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.validateLines() = %v, want %v", got, want)
	}
}

// validate lines, but with record validation errors
func TestValidateImpl_validateLines_with_validation_errors(t *testing.T) {

	r, w, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testBadData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

// validate lines with no record validation errors, json output
func TestValidateImpl_validateLines_jsonOutput(t *testing.T) {

	r, w, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testGoodData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 12 lines, 0 were bad"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
}

// validate lines, but with record validation errors and json output
func TestValidateImpl_validateLines_with_validation_errors_jsonOutput(t *testing.T) {

	r, w, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testBadData))

	w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)

	want := "Validated 16 lines, 4 were bad"
	if !strings.Contains(got, want) {
		t.Errorf("ValidateImpl.Read() = %v, want %v", got, want)
	}
}

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

// create a tempdata file with the given content and extension
func createTempDataFile(t *testing.T, content string, fileextension string) (filename string, cleanUp func()) {
	t.Helper()
	tmpfile, err := os.CreateTemp(t.TempDir(), "test.*."+fileextension)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}

	filename = tmpfile.Name()

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return filename,
		func() {
			os.Remove(filename)
		}
}

// create a temp gzipped datafile with the given content
func createTempGZIPDataFile(t *testing.T, content string) (filename string, cleanUp func()) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "test.*.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()
	gf := gzip.NewWriter(tmpfile)
	defer gf.Close()
	fw := bufio.NewWriter(gf)
	if _, err := fw.WriteString(content); err != nil {
		t.Fatal(err)
	}
	fw.Flush()
	filename = tmpfile.Name()
	return filename,
		func() {
			os.Remove(filename)
		}
}

// serve the requested resource on a random port
func serveResource(t *testing.T, filename string) (*http.Server, *net.Listener, int) {
	t.Helper()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	idx := strings.LastIndex(filename, string(os.PathSeparator))
	fs := http.FileServer(http.Dir(filename[:idx]))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: fs,
	}
	return &server, &listener, port

}

// capture stdout for testing
func mockStdout(t *testing.T) (reader *os.File, writer *os.File, cleanUp func()) {
	t.Helper()

	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stdout = writer

	return reader,
		writer,
		func() {
			//clean-up
			os.Stdout = origStdout
		}
}

// capture stderr for testing
func mockStderr(t *testing.T) (reader *os.File, writer *os.File, cleanUp func()) {
	t.Helper()
	origStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stderr = writer

	return reader,
		writer,
		func() {
			//clean-up
			os.Stderr = origStderr
		}
}

var testGoodData string = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
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
var testBadData string = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
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
