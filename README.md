# validate

## Synopsis

`validate` is a command in the
[senzing-tools](https://github.com/Senzing/senzing-tools)
suite of tools.
This command validates that a JSONL file is properly formatted and each line
contains sufficient key-value pairs for Senzing to each as a record.  It is
highly recommend that this code be taken and extended to validate JSONL records
to meet your needs.

[![Go Reference](https://pkg.go.dev/badge/github.com/senzing/validate.svg)](https://pkg.go.dev/github.com/senzing/validate)
[![Go Report Card](https://goreportcard.com/badge/github.com/senzing/validate)](https://goreportcard.com/report/github.com/senzing/validate)
[![go-test.yaml](https://github.com/Senzing/validate/actions/workflows/go-test.yaml/badge.svg)](https://github.com/Senzing/validate/actions/workflows/go-test.yaml)
[![License](https://img.shields.io/badge/License-Apache2-brightgreen.svg)](https://github.com/Senzing/validate/blob/main/LICENSE)

## Overview

`validate` tests each line of a give JSONL file to ensure that it is valid
JSON and contains two necessary key-value pairs:  `RECORD_ID` and `DATA_SOURCE`.

The file is given to `validate` with the command-line parameter `input-url` or
as the environment variable `SENZING_TOOLS_INPUT_URL`.  Note this is a URL so
local files will need `file://` and remote files `http://` or `https://`. If
the given file has the `.gz` extension, it will be treated as a compressed file
JSONL file.  If the file has a `.jsonl` extension it will be treated
accordingly. If the file has another extension it will be rejected, unless the
`input-file-type` or `SENZING_TOOLS_INPUT_FILE_TYPE` is set to `JSONL`.

`validate` is intended as a starting point for other validation needs.  It
should be fairly straight forward to extend it to test other JSON objects or
extend it to other file types.

## Install

1. The `validate` command is installed with the
   [senzing-tools](https://github.com/Senzing/senzing-tools)
   suite of tools.
   See senzing-tools [install](https://github.com/Senzing/senzing-tools#install).

## Use

```console
senzing-tools validate [flags]
```

1. For options and flags:
    1. [Online documentation](https://hub.senzing.com/senzing-tools/senzing-tools_validate.html)
    1. Runtime documentation:

        ```console
        senzing-tools validate --help
        ```

1. In addition to the following simple usage examples, there are additional [Examples](docs/examples.md).

### Using command line options

1. :pencil2: Specify database using command line option.
   Example:

    ```console
    senzing-tools validate \
        --input-url https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl
    ```

1. See [Parameters](#parameters) for additional parameters.

### Using environment variables

1. :pencil2: Specify database using environment variable.
   Example:

    ```console
    export SENZING_TOOLS_INPUT_URL=https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl
    senzing-tools validate
    ```

1. See [Parameters](#parameters) for additional parameters.

### Using Docker

This usage shows how to initialze a database with a Docker container.

1. :pencil2: Run `senzing/senzing-tools`.
   Example:

    ```console
    docker run \
        --env SENZING_TOOLS_COMMAND=validate \
        --env SENZING_TOOLS_INPUT_URL=https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl \
        --rm \
        senzing/senzing-tools
    ```

1. See [Parameters](#parameters) for additional parameters.

### Parameters

- **[SENZING_TOOLS_INPUT_FILE_TYPE](https://github.com/Senzing/knowledge-base/blob/main/lists/environment-variables.md#senzing_tools_input_file_type)**
- **[SENZING_TOOLS_INPUT_URL](https://github.com/Senzing/knowledge-base/blob/main/lists/environment-variables.md#senzing_tools_input_url)**
- **[SENZING_TOOLS_JSON_OUTPUT](https://github.com/Senzing/knowledge-base/blob/main/lists/environment-variables.md#senzing_tools_json_output)**
- **[SENZING_TOOLS_ENGINE_LOG_LEVEL](https://github.com/Senzing/knowledge-base/blob/main/lists/environment-variables.md#senzing_tools_engine_log_level)**

## References

- [Command reference](https://hub.senzing.com/senzing-tools/senzing-tools_validate.html)
- [Development](docs/development.md)
- [Errors](docs/errors.md)
- [Examples](docs/examples.md)
