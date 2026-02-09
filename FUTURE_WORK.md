# Future Work & Improvements

The current prototype demonstrates the core concept of the impact detector. The following areas are planned for future development based on the GSoC proposal.

## 1. Automated Discovery

- **Global Index**: Integrate with pkg.go.dev (or similar index) to automatically discover _all_ downstream consumers of the target library.
- **Top N Dependents**: Use stars, forks, or download counts to prioritize testing the most critical dependents.

## 2. Advanced Execution Environment

- **Sandboxed Builds**: Run tests inside isolated Docker containers to prevent malicious code execution and ensure consistent environments.
- **Dependency Isolation**: Better handling of transitive dependencies and version conflicts when using `go mod replace`.

## 3. Intelligent Analysis

- **Code Coverage Impact**: Not just pass/fail, but analyze if the changed code paths were actually executed by the consumer tests.
- **Flakiness Detection**: Run tests multiple times to distinguish between real regressions and flaky tests.
- **Build Failures vs Test Failures**: Better categorization of compilation errors vs runtime test failures.

## 4. Enhanced Reporting

- **GitHub Action Integration**: Automatically comment on PRs with the impact report.
- **Visual Dashboard**: A web interface to view the impact matrix across hundreds of consumers.
- **Detail Drill-down**: Links to specific failing test logs and diffs.

## 5. Performance Optimization

- **Parallel Execution**: Run consumer tests concurrently.
- **Smart Caching**: Cache test results and clones more aggressively.
