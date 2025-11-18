# GitHub Actions Workflows

This directory contains GitHub Actions workflows for continuous integration and testing.

## Workflows

### 🧪 Tests (`test.yml`)

**Triggers:** Push and PR to `main`, `master`, or `develop` branches

**Jobs:**

1. **test** - Runs on multiple OS and Go versions
   - **Matrix:** 
     - OS: Ubuntu, macOS, Windows
     - Go: 1.21, 1.22, 1.23
   - **Steps:**
     - Unit tests (`./liquid/...`)
     - Tag tests (`./liquid/tag/...`, `./liquid/tags/...`)
     - Integration tests (`./integration/...`)
   - **Coverage:** Uploads to Codecov (Ubuntu + Go 1.23 only)

2. **lint** - Code quality checks
   - Runs `golangci-lint` with configured linters
   - See `.golangci.yml` for configuration

3. **build** - Ensures code builds
   - Builds main package
   - Builds all examples

### 📊 Benchmarks (`benchmarks.yml`)

**Triggers:** Push/PR to `main`/`master`, or manual dispatch

**Jobs:**

1. **benchmark** - Performance testing
   - Runs all benchmarks in `performance/` directory
   - Stores results as artifacts
   - Tracks performance over time using `benchmark-action`
   - Alerts on performance regressions > 150%

## Configuration Files

### `.golangci.yml`

Linter configuration for code quality checks:

- **Enabled linters:** gofmt, govet, errcheck, staticcheck, unused, gosimple, ineffassign, typecheck, exportloopref, gocyclo, misspell
- **Complexity threshold:** 15
- **Exclusions:** Test files, reference-liquid directory

## Status Badges

Add these to your README:

```markdown
[![Tests](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml)
[![Benchmarks](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml)
[![codecov](https://codecov.io/gh/Notifuse/liquidgo/branch/main/graph/badge.svg)](https://codecov.io/gh/Notifuse/liquidgo)
```

## Local Testing

Run the same checks locally before pushing:

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...

# Run linter
golangci-lint run

# Run benchmarks
cd performance && go test -bench=. -benchmem
```

## Secrets Required

For full functionality, configure these secrets in your GitHub repository:

- `CODECOV_TOKEN` - For uploading coverage reports (optional)
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions

## Customization

### Changing Go Versions

Edit the matrix in `test.yml`:

```yaml
strategy:
  matrix:
    go: ['1.21', '1.22', '1.23']  # Add or remove versions
```

### Changing Test OS

Edit the matrix in `test.yml`:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]  # Modify as needed
```

### Adjusting Benchmark Threshold

Edit `benchmarks.yml`:

```yaml
with:
  alert-threshold: '150%'  # Change threshold percentage
```

### Adding More Linters

Edit `.golangci.yml`:

```yaml
linters:
  enable:
    - gofmt
    - govet
    # Add more linters here
```

## Troubleshooting

### Tests Failing on Specific OS

Check the workflow run details to see which OS/Go version combination is failing. You may need to add platform-specific handling in your code.

### Linter Errors

Run `golangci-lint run` locally to see the same errors. Fix them before pushing.

### Benchmark Action Failures

Ensure the benchmark output format is compatible with the `benchmark-action`. The action expects standard Go benchmark output format.

### Coverage Upload Failures

Verify that the `CODECOV_TOKEN` is set correctly if using a private repository. Public repositories don't require the token.

