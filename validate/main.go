package validate

import (
	"context"
	"errors"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type Validate interface {
	Read(ctx context.Context, inputURL string)
	SetLogLevel(ctx context.Context, logLevelName string) error
}

// ----------------------------------------------------------------------------
// Constants
// ----------------------------------------------------------------------------

// Identfier of the  package found messages having the format "senzing-6203xxxx".
const ComponentID = 6203

// Log message prefix.
const Prefix = "validate: "

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

// Message templates for szconfig implementations.
var IDMessages = map[int]string{
	2200: Prefix + "Validating URL string: %s",
	2201: Prefix + "Validating as a JSONL file.",
	2203: Prefix + "Validating a GZIP file.",
	2204: Prefix + "Validating as a JSONL resource.",
	2205: Prefix + "Validating a GZIP resource.",
	2210: Prefix + "Validated %d lines, %d were bad.",
	3001: Prefix + "%d line(s) had no RECORD_ID field.",
	3002: Prefix + "%d line(s) had no DATA_SOURCE field.",
	3003: Prefix + "%d line(s) are not well formed JSON-lines.",
	3004: Prefix + "%d line(s) did not validate for an unknown reason.",
	3005: Prefix + "Line %d: a RECORD_ID field is required",
	3006: Prefix + "Line %d: a DATA_SOURCE field is required",
	3007: Prefix + "Line %d: JSON-line not well formed",
	3008: Prefix + "Line %d: did not validate for an unknown reason",
	3009: Prefix + "Warning: Unable to set log level to %s, defaulting to INFO",
	5000: Prefix + "Fatal error, Check the input-url parameter: %s",
	5001: Prefix + "Fatal error parsing input-url.",
	5002: Prefix + "Fatal error unable to handle %s input URLs.",
	5003: Prefix + "Fatal error retrieving input-url: %s",
	5004: Prefix + "Fatal error opening input file: %s",
	5005: Prefix + "Fatal error opening stdin.",
	5006: Prefix + "Fatal error stdin not piped.",
	5007: Prefix + "Fatal error opening GZIPped file: %s",
	5008: Prefix + "Fatal error reading GZIPped file: %s",
	5009: Prefix + "Fatal error retrieving GZIPped input-url: %s",
	5010: Prefix + "Fatal error reading GZIPped input-url: %s",
	5011: Prefix + "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type).",
	5012: Prefix + "If this is a valid JSONL resource, please rename with the .jsonl extension or use the file type override (--file-type).",
}

// Status strings for specific messages.
var IDStatuses = map[int]string{}

var errForPackage = errors.New("validate")
