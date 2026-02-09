# Architecture & Design

This prototype implements a streamlined pipeline to detect breaking changes by executing downstream consumer tests against a proposed library change.

## High-Level Workflow

```mermaid
graph TD
    A[Start] --> B[Read Config (Consumers)]
    B --> C{For Each Consumer}
    C --> D[Clone Repository]
    D --> E[Run Baseline Tests (Published Version)]
    E --> F[Inject Local Library Version (go mod replace)]
    F --> G[Run Experiment Tests (Local Version)]
    G --> H[Compare Results]
    H --> I[Generate Impact Report]
    I --> C
    C --> J[End]
```

## detailed Components

### 1. Consumer Discovery (`config.json`)

The prototype currently uses a static configuration file to list downstream consumers.

- **Attributes**: Repository URL, Branch, Module Directory, Packages to Test.
- **Future**: Dynamic discovery via `pkg.go.dev` or BigQuery index.

### 2. Execution Engine (`runner/`)

Handles the sandboxed environment for running tests.

- **Cloning**: Uses `git clone --depth 1` for speed. Implements retry logic and caching for reliability.
- **Injection**: Uses `go mod edit -replace` to seamlessly swap the dependency graph to point to the local version of the library.
- **Test Execution**: Runs `go test -json` to capture structured output for analysis.

### 3. Analysis Engine (`analyzer/`)

Parses the JSON output from `go test` to perform a semantic comparison.

- **Baseline vs. Experiment**: Matches tests by Package and Name.
- **Regression Detection**: flags tests that `PASS` in Baseline but `FAIL` in Experiment.
- **Fix Detection**: flags tests that `FAIL` in Baseline but `PASS` in Experiment.

## Key Design Decisions

- **Local Execution**: The prototype runs locally to allow quick iteration for the developer before pushing code.
- **Standard Go Tooling**: Relies entirely on `go` commands (`mod`, `test`, `list`) rather than custom parsing, ensuring compatibility with the ecosystem.
- **JSON Output**: Parsing structured JSON is more robust than regex scraping of terminal output.
