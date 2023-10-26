# validate examples

## Examples of CLI

1. :pencil2: Specify JSONL file URL using command line option.
   Example:

    ```console
    senzing-tools validate \
        --input-url https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl
    ```

1. :pencil2: Specify JSONL file URL using command line option.
   Example:

    ```console
    senzing-tools validate \
        --input-url file:///path/to/json/lines/file.jsonl
    ```

1. :pencil2: Specify JSONL file URL using command line option.  Notice this file doesn't end with `.jsonl` so we need to specify the file type.
   Example:

    ```console
    senzing-tools validate \
        --input-url file:///path/to/json/lines/file.txt \
        --input-file-type JSONL
    ```

1. :pencil2: Change the log level using command line option.
   Example:

    ```console
    senzing-tools validate \
        --input-url file:///path/to/json/lines/file.jsonl \
        --log-level DEBUG
    ```


### Using environment variables

1. :pencil2: Specify JSONL file URL using environment variable.
   Example:

    ```console
    export SENZING_TOOLS_INPUT_URL=https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl
    senzing-tools validate
    ```

1. :pencil2: Specify JSONL file URL using environment variable.
   Example:

    ```console
    export SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.jsonl
    senzing-tools validate
    ```

1. :pencil2: Specify JSONL file URL using environment variable. Notice this file doesn't end with `.jsonl` so we need to specify the file type.
   Example:

    ```console
    export SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.txt
    export SENZING_TOOLS_INPUT_FILE_TYPE=JSONL
    senzing-tools validate
    ```

1. :pencil2: Specify file URL using environment variable.
   Example:

    ```console
    export SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.jsonl
    export SENZING_TOOLS_LOG_LEVEL=DEBUG
    senzing-tools validate
    ```



## Examples of Docker

1. :pencil2: Run `senzing/senzing-tools` to validate a file with a Docker container.
   Example:

    ```console
    docker run \
        --env SENZING_TOOLS_COMMAND=validate \
        --env SENZING_TOOLS_INPUT_URL=https://public-read-access.s3.amazonaws.com/TestDataSets/SenzingTruthSet/truth-set-3.0.0.jsonl \
        --rm \
        senzing/senzing-tools
    ```

1. :pencil2: Run `senzing/senzing-tools` to validate a file with a Docker container.
   Example:

    ```console
    docker run \
        --env SENZING_TOOLS_COMMAND=validate \
        --env SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.jsonl \
        --rm \
        senzing/senzing-tools
    ```

1. :pencil2: Run `senzing/senzing-tools` to validate a file with a Docker container.
   Example:

    ```console
    docker run \
        --env SENZING_TOOLS_COMMAND=validate \
        --env SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.txt \
        --env SENZING_TOOLS_INPUT_FILE_TYPE=JSONL
        --rm \
        senzing/senzing-tools
    ```

1. :pencil2: Run `senzing/senzing-tools` to validate a file with a Docker container.
   Example:

    ```console
    docker run \
        --env SENZING_TOOLS_COMMAND=validate \
        --env SENZING_TOOLS_INPUT_URL=file:///path/to/json/lines/file.jsonl \
        --env SENZING_TOOLS_LOG_LEVEL=DEBUG \
        --rm \
        senzing/senzing-tools
    ```


