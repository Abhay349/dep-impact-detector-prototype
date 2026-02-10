# Behavioral Impact Detector Prototype

This tool prototype demonstrates how to detect breaking changes in downstream consumers of a Go library by running their test suites against your local changes.

## Demo Video

https://github.com/user-attachments/assets/d6ca1165-191c-4b9c-a847-dd12243b7020


## Documentation

- [Architecture & Design](./ARCHITECTURE.md)
- [Future Work](./FUTURE_WORK.md)

## Features

- **Consumer Discovery**: Configurable list of dependent repositories via `config.json`.
- **Sandbox Execution**: Clones consumers to a temporary directory.
- **Baseline Analysis**: Runs tests against the currently published version of the target library.
- **Impact Analysis**: Injects your local version of the library using `go mod replace`, runs tests again, and compares results.
- **Reporting**: Outputs a difference report highlighting regressions (Pass -> Fail) and fixes (Fail -> Pass).

## Prerequisites

- Go 1.21+
- Git

## Quick Start

1.  **Dependencies**: Ensure you have Go 1.21+ and Git installed.
2.  **Configuration**: Edit `config.json` to add the GitHub repositories of the consumers you want to test.

    ```json
    {
      "target_module": "go.opentelemetry.io/otel/trace",
      "consumers": [
        {
          "name": "opentelemetry-go-contrib",
          "repo_url": "https://github.com/open-telemetry/opentelemetry-go-contrib",
          "branch": "main",
          "module_dir": "instrumentation/net/http/otelhttp",
          "packages": ["./..."]
        }
      ]
    }
    ```

3.  **Run**:
    ```bash
    go run main.go -local /absolute/path/to/your/local/library
    ```

    - `-local`: Absolute path to the local generic/modified version of the library you are testing against consumers.
    - `-config`: Path to config file (default `config.json`).
    - `-workdir`: Temporary directory for cloning (default `./temp_work`).

## How to Verify

1.  Run the tool once to ensure a clean run ("0 Regressions").
2.  Modify your local library code to introduce a breaking change (e.g., change a public function signature).
3.  Run the tool again. You should see "Regressions" or "Test Missing" reported for the consumer.

## How it Works

1.  **Clone**: The tool clones the consumer repository.
2.  **Baseline**: It runs `go test -json ./...` to establish a baseline using the released dependencies defined in `go.mod`.
3.  **Inject**: It runs `go mod edit -replace <target_module>=<local_path>` and `go mod tidy` to force usage of your local code.
4.  **Experiment**: It runs `go test -json ./...` again with your changes.
5.  **Compare**: It parses both JSON outputs and reports tests that changed status.

## Directory Structure

- `main.go`: Entry point and orchestration.
- `runner/`: Handles Git and Go command execution.
- `analyzer/`: Parses test output and computes diffs.
- `config.json`: Configuration for consumers.
