package validate

import "context"

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type Initializer interface {
	Read(ctx context.Context, inputURL string)
	SetLogLevel(ctx context.Context, logLevelName string) error
}

// ----------------------------------------------------------------------------
// Constants
// ----------------------------------------------------------------------------

// Identfier of the  package found messages having the format "senzing-6501xxxx".
const ComponentId = 6203

// Log message prefix.
const Prefix = "validate: "

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

// Message templates for g2config implementations.
var IdMessages = map[int]string{
	2002: Prefix + "Check the input-url parameter: %s",
	2003: Prefix + "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type).",
	2004: Prefix + "If this is a valid JSONL resource, please rename with the .jsonl extension or use the file type override (--file-type).",
	2200: Prefix + "Validating URL string: %s",
	2201: Prefix + "Validating as a JSONL file.",
	2203: Prefix + "Validating a GZ file.",
	2204: Prefix + "Validating as a JSONL resource.",
	2205: Prefix + "Validating a GZ resource.",
	2210: Prefix + "Validated %d lines, %d were bad.",
	3001: Prefix + "%d line(s) had no RECORD_ID field.",
	3002: Prefix + "%d line(s) had no DATA_SOURCE field.",
	3003: Prefix + "%d line(s) are not well formed JSON-lines.",
	3004: Prefix + "%d line(s) did not validate for an unknown reason.",
	4001: Prefix + "Error: Unable to set log level to: %s",
	5001: Prefix + "Fatal error parsing input-url.",
	5002: Prefix + "Fatal error unable to handle %s input URLs.",
	5003: Prefix + "Fatal error retrieving input-url: %s",
	5004: Prefix + "Fatal error opening input file: %s",
	5005: Prefix + "Fatal error opening stdin.",
	5006: Prefix + "Fatal error stdin not piped.",
	5007: Prefix + "Fatal error opening gzipped file: %s",
	5008: Prefix + "Fatal error reading gzipped file: %s",
	5009: Prefix + "Fatal error retrieving gzipped input-url: %s",
	5010: Prefix + "Fatal error reading gzipped input-url: %s",

	// 10:  "Enter " + Prefix + "Initialize().",
	// 11:  "Exit  " + Prefix + "Initialize(); json.Marshal failed; returned (%v).",
	// 12:  "Exit  " + Prefix + "Initialize(); initializerImpl.InitializeSpecificDatabase failed; returned (%v).",
	// 13:  "Exit  " + Prefix + "Initialize(); senzingSchema.SetLogLevel failed; returned (%v).",
	// 14:  "Exit  " + Prefix + "Initialize(); senzingSchema.InitializeSenzing failed; returned (%v).",
	// 15:  "Exit  " + Prefix + "Initialize(); senzingConfig.SetLogLevel failed; returned (%v).",
	// 16:  "Exit  " + Prefix + "Initialize(); senzingConfig.InitializeSenzing; returned (%v).",
	// 17:  "Exit  " + Prefix + "Initialize(); initializerImpl.observers.RegisterObserver; returned (%v).",
	// 18:  "Exit  " + Prefix + "Initialize(); initializerImpl.createGrpcObserver; returned (%v).",
	// 19:  "Exit  " + Prefix + "Initialize(); initializerImpl.registerObserverSenzingSchema; returned (%v).",
	// 20:  "Exit  " + Prefix + "Initialize(); initializerImpl.registerObserverSenzingConfig; returned (%v).",
	// 21:  "Exit  " + Prefix + "Initialize(); os.Stat failed; returned (%v).",
	// 29:  "Exit  " + Prefix + "Initialize() returned (%v).",
	// 40:  "Enter " + Prefix + "InitializeSpecificDatabase().",
	// 41:  "Exit  " + Prefix + "InitializeSpecificDatabase(); json.Marshal failed; returned (%v).",
	// 42:  "Exit  " + Prefix + "InitializeSpecificDatabase(); engineconfigurationjsonparser.New failed; returned (%v).",
	// 43:  "Exit  " + Prefix + "InitializeSpecificDatabase(); parser.GetDatabaseUrls failed; returned (%v).",
	// 44:  "Exit  " + Prefix + "InitializeSpecificDatabase(); url.Parse failed; returned (%v).",
	// 45:  "Exit  " + Prefix + "InitializeSpecificDatabase(); initializerImpl.initializeSpecificDatabaseSqlite; returned (%v).",
	// 49:  "Exit  " + Prefix + "InitializeSpecificDatabase() returned (%v).",
	// 50:  "Enter " + Prefix + "RegisterObserver(%s).",
	// 51:  "Exit  " + Prefix + "RegisterObserver(%s); json.Marshal failed; returned (%v).",
	// 52:  "Exit  " + Prefix + "RegisterObserver(%s); initializerImpl.observers.RegisterObserver failed; returned (%v).",
	// 53:  "Exit  " + Prefix + "RegisterObserver(%s); initializerImpl.getSenzingConfig().RegisterObserver failed; returned (%v).",
	// 54:  "Exit  " + Prefix + "RegisterObserver(%s); initializerImpl.getSenzingSchema().RegisterObserver; returned (%v).",
	// 59:  "Exit  " + Prefix + "RegisterObserver(%s) returned (%v).",
	// 60:  "Enter " + Prefix + "SetLogLevel(%s).",
	// 61:  "Exit  " + Prefix + "SetLogLevel(%s); json.Marshal failed; returned (%v).",
	// 62:  "Exit  " + Prefix + "SetLogLevel(%s); logging.IsValidLogLevelName failed; returned (%v).",
	// 63:  "Exit  " + Prefix + "SetLogLevel(%s); initializerImpl.getLogger().SetLogLevel failed; returned (%v).",
	// 64:  "Exit  " + Prefix + "SetLogLevel(%s); initializerImpl.senzingConfigSingleton.SetLogLevel failed; returned (%v).",
	// 65:  "Exit  " + Prefix + "SetLogLevel(%s); initializerImpl.getSenzingSchema().SetLogLevel failed; returned (%v).",
	// 69:  "Exit  " + Prefix + "SetLogLevel(%s) returned (%v).",
	// 70:  "Enter " + Prefix + "UnregisterObserver(%s).",
	// 71:  "Exit  " + Prefix + "UnregisterObserver(%s); json.Marshal failed; returned (%v).",
	// 72:  "Exit  " + Prefix + "UnregisterObserver(%s); initializerImpl.getSenzingConfig().UnregisterObserver failed; returned (%v).",
	// 73:  "Exit  " + Prefix + "UnregisterObserver(%s); initializerImpl.getSenzingSchema().UnregisterObserver failed; returned (%v).",
	// 74:  "Exit  " + Prefix + "UnregisterObserver(%s); initializerImpl.observers.UnregisterObserver failed; returned (%v).",
	// 79:  "Exit  " + Prefix + "UnregisterObserver(%s) returned (%v).",
	// 80:  "Enter " + Prefix + "SetObserverOrigin(%s).",
	// 81:  "Exit  " + Prefix + "SetObserverOrigin(%s); json.Marshal failed; returned (%v).",
	// 89:  "Exit  " + Prefix + "SetObserverOrigin(%s).",
	// 100: "Enter " + Prefix + "initializeSpecificDatabaseSqlite(%v).",
	// 101: "Exit  " + Prefix + "initializeSpecificDatabaseSqlite(%v); os.Stat failed; returned (%v).",
	// 102: "Exit  " + Prefix + "initializeSpecificDatabaseSqlite(%v); os.MkdirAll failed; returned (%v).",
	// 103: "Exit  " + Prefix + "initializeSpecificDatabaseSqlite(%v); os.Create failed; returned (%v).",
	// 109: "Exit  " + Prefix + "initializeSpecificDatabaseSqlite(%v) returned (%v).",
}

// Status strings for specific messages.
var IdStatuses = map[int]string{}
