package validate

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------------------
// test Read method
// ----------------------------------------------------------------------------

// read jsonl file successfully, no record validation errors
func TestRead(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read jsonl file successully, but with record validation errors
func TestRead_with_bad_records(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read jsonl file successfully, but attept to set a bad log level
// falls back to INFO
func TestRead_bad_loglevel(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
		LogLevel: "BAD",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Unable to set log level to BAD"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// attempt to read a jsonl file, but the input url is bad
func TestRead_bad_url(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "BAD",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 1; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error, Check the input-url parameter: BAD"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a jsonl file, but the input url isn't parsable
func TestRead_bad_url_parse(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "http://bad:bad{BAD=bad@example.com",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error parsing input-url"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a jsonl file, but the input url is not understood
func TestRead_url_drop_through(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "BAD,Really bad",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error unable to handle"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a jsonl file, but the file doesn't exist
func TestRead_file_doesnt_exist(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "file:///badfile.jsonl",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error opening input file: /badfile.jsonl"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

func TestRead_stdin_unpipe_error(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
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

	var got string = ""
	for i := 0; i < 1; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error stdin not piped"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a file, but it has a file type that is not known
func TestRead_bad_file_type(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "txt")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type)"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a file type that is not known, but override with input file type
func TestRead_override_file_type(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "txt")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputFileType: "JSONL",
		InputUrl:      fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// ----------------------------------------------------------------------------
// test Read with .gz files
// ----------------------------------------------------------------------------

// read a gz file successfully, with no record validation errors
func TestRead_gz(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempGzDataFile(t, testGoodData)
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a gz file successfully, but with record validation errors
func TestRead_gz_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempGzDataFile(t, testBadData)
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// ----------------------------------------------------------------------------
// test Read resources
// ----------------------------------------------------------------------------

func TestRead_resource_jsonl(t *testing.T) {

	scanner, cleanUpStdout := mockStdout(t)
	defer cleanUpStdout()

	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "jsonl")
	defer cleanUpTempFile()
	server := serveResource(t, 3000, filename)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("http://localhost:3000/%s", filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)

	if err := server.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestRead_resource_gzip(t *testing.T) {

	scanner, cleanUpStdout := mockStdout(t)
	defer cleanUpStdout()

	filename, cleanUpTempFile := createTempGzDataFile(t, testGoodData)
	defer cleanUpTempFile()
	server := serveResource(t, 3000, filename)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("http://localhost:3000/%s", filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)

	if err := server.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestRead_resource_unknown_extension(t *testing.T) {

	scanner, cleanUpStdout := mockStdout(t)
	defer cleanUpStdout()

	filename, cleanUpTempFile := createTempDataFile(t, testGoodData, "bad")
	defer cleanUpTempFile()
	server := serveResource(t, 3000, filename)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	idx := strings.LastIndex(filename, "/")
	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("http://localhost:3000/%s", filename[(idx+1):]),
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "If this is a valid JSONL resource"
	assert.Contains(t, got, msg)
	assert.False(t, result)

	if err := server.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

// ----------------------------------------------------------------------------
// test Read with Json output
// ----------------------------------------------------------------------------

// read a json file successfully, with no record validation errors
func TestRead_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a json file successfully, but with record validation errors
func TestRead_jsonOutput_bad(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// ----------------------------------------------------------------------------
// test jsonl file read with json output
// ----------------------------------------------------------------------------

// read a json file successfully, with no record validation errors
func TestReadJsonlFile(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readJSONLFile(filename)

	scanner.Scan()
	got := scanner.Text()

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a json file successfully, but with record validation errors
func TestReadJsonlFile_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readJSONLFile(filename)

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a json file successfully, with no record validation errors
func TestReadJsonlFile_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	result := validator.readJSONLFile(filename)

	scanner.Scan()
	got := scanner.Text()

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a json file successfully, but with record validation errors
func TestReadJsonlFile_jsonOutput_bad(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	result := validator.readJSONLFile(filename)

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// ----------------------------------------------------------------------------
// test gzip file read
// ----------------------------------------------------------------------------

// read a gzip file successfully, no record validation errors
func TestReadGzipFile(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempGzDataFile(t, testGoodData)
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZFile(filename)

	scanner.Scan()
	got := scanner.Text()

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// read a gzip file successfully, but with record validation errors
func TestReadGzipFile_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempGzDataFile(t, testBadData)
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZFile(filename)

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
	assert.True(t, result)
}

// attempt to read a gzip file that doesn't exist
func TestReadGzipFile_file_does_not_exist(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename := "/bad.gz"

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZFile(filename)

	scanner.Scan()
	got := scanner.Text()

	msg := "no such file or directory"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// attempt to read a gzip file that isn't a gzip file
func TestReadGzipFile_not_a_gzip_file(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "gz")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	result := validator.readGZFile(filename)

	scanner.Scan()
	got := scanner.Text()

	msg := "invalid header"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// ----------------------------------------------------------------------------
// test read stdin
// ----------------------------------------------------------------------------

func TestReadStdin(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
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

	var got string = ""
	for i := 0; i < 1; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error stdin not piped"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}
func TestReadStdin_unpipe_error(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
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

	var got string = ""
	for i := 0; i < 1; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error stdin not piped"
	assert.Contains(t, got, msg)
	assert.False(t, result)
}

// ----------------------------------------------------------------------------
// test validateLines
// ----------------------------------------------------------------------------

// validate lines with no record validation errors
func TestValidateLines(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testGoodData))

	scanner.Scan()
	got := scanner.Text()

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

// validate lines, but with record validation errors
func TestValidateLines_with_validation_errors(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testBadData))

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

// validate lines with no record validation errors, json output
func TestValidateLines_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testGoodData))

	scanner.Scan()
	got := scanner.Text()

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

// validate lines, but with record validation errors and json output
func TestValidateLines_with_validation_errors_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testBadData))

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
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
func createTempGzDataFile(t *testing.T, content string) (filename string, cleanUp func()) {
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

// serve the requested resource on the requested port
func serveResource(t *testing.T, port int, filename string) *http.Server {
	t.Helper()

	idx := strings.LastIndex(filename, "/")
	fs := http.FileServer(http.Dir(filename[:idx]))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: fs,
	}
	return &server

}

// capture stdout for testing
func mockStdout(t *testing.T) (buffer *bufio.Scanner, cleanUp func()) {
	t.Helper()
	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stdout = writer

	return bufio.NewScanner(reader),
		func() {
			//clean-up
			os.Stdout = origStdout
		}
}

// capture stderr for testing
func mockStderr(t *testing.T) (buffer *bufio.Scanner, cleanUp func()) {
	t.Helper()
	origStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stderr = writer

	return bufio.NewScanner(reader),
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
