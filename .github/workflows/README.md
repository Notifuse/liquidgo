# GitHub Actions Workflows

This directory contains GitHub Actions workflows for continuous integration and testing.

## Workflows

### 🧪 Tests (`test.yml`)

**Triggers:** Push and PR to `main`. Every job uses the Go version of `go.mod`.

**Jobs:**

1. **test** - Linux amd64 and arm64: vet, build, `go test ./...` and `go test -race ./...`;
   on amd64, `go test -short ./...` for `386` and the build for `arm`.
2. **windows** - `go test -short ./...`, for the paths of `LocalFileSystem`.
3. **reference** - Checks that `integration/testdata/differential/expected.json` is what
   Shopify Liquid 5.11.0 renders (`scripts/differential_reference.rb --check`), then runs the
   conformance and differential tests.
4. **fuzz** - `FuzzRender` for 2 minutes; the input it fails on is kept as an artifact.
5. **lint** - `golangci-lint` with `.golangci.yml`.

Coverage is uploaded to Codecov by `coverage.yml` only.

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

The workflows read the version from `go.mod`.

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

