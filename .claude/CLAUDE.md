# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`validate` is a Go CLI tool that validates JSONL files for Senzing. It verifies each line is valid JSON and contains required `RECORD_ID` and `DATA_SOURCE` fields. Part of the [senzing-tools](https://github.com/senzing-garage/senzing-tools) suite, located in Senzing Garage (experimental, not production-ready).

## Common Commands

```bash
# Build
make build                    # Build for current OS/architecture
make build-all                # Build for all platforms (darwin/linux/windows, amd64/arm64)

# Test
make test                     # Run all tests
make check-coverage           # Run tests and verify 70% coverage threshold

# Lint
make lint                     # Run all linters (golangci-lint, govulncheck, cspell)
make fix                      # Auto-fix many linting issues

# Docker
make docker-build             # Build Docker image
make docker-test              # Run tests via docker-compose

# Development setup
make dependencies-for-development  # Install dev tools
make dependencies                  # Update Go module dependencies
```

### Running a Single Test

```bash
go test -run TestFunctionName ./validate/
go test -run TestFunctionName ./cmd/
```

## Architecture

```console
main.go           → Entry point, calls cmd.Execute()
├── cmd/          → CLI layer using Cobra/Viper
│   ├── root.go   → Main command setup with flags
│   └── context_*.go → Platform-specific context (linux/darwin/windows)
└── validate/     → Core business logic
    ├── main.go   → Validate interface and BasicValidate struct
    └── validate.go → Read() method with line-by-line JSONL validation
```

**Key Parameters:**

- `--input-url` / `SENZING_TOOLS_INPUT_URL`: File URL (file://, http://, https://)
- `--input-file-type` / `SENZING_TOOLS_INPUT_FILE_TYPE`: Override file type detection
- `--json-output` / `SENZING_TOOLS_JSON_OUTPUT`: JSON output mode
- `--log-level` / `SENZING_TOOLS_LOG_LEVEL`: Logging level

**Component ID:** 6203 (for Senzing message identification)

## Code Standards

**Linting:** 100+ linters configured in `.github/linters/.golangci.yaml`

- Max line length: 120 characters
- Max cyclomatic complexity: 15
- Do not use deprecated `io/ioutil` package (use `io` and `os` instead)
- Formatters: gofumpt, goimports, gci, golines

**Coverage:** 70% minimum for files, packages, and total (configured in `.github/coverage/testcoverage.yaml`)

**Testing:** Uses `testify` for assertions. Platform-specific tests use `_linux_test.go` / `_windows_test.go` suffixes.

## Dependencies

Primary Senzing libraries:

- `github.com/senzing-garage/go-cmdhelping` - Command helper utilities
- `github.com/senzing-garage/go-helpers` - Helper utilities
- `github.com/senzing-garage/go-logging` - Structured logging

CLI framework: `spf13/cobra` and `spf13/viper`
