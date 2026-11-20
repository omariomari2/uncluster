# Contributing to Uncluster

Thanks for taking the time to contribute. This document covers the local workflow for this repository.

## Requirements

- Go 1.21 or later
- Node.js 18 or later if you need to run a generated Vite project

## Build

From the repository root:

```sh
make build          # CLI binary: ./uncluster
make build-server   # API server: ./uncluster-server
make install        # install the CLI to $(go env GOPATH)/bin
make uninstall      # remove the installed CLI
make clean          # remove local build artifacts
```

You can also run without installing:

```sh
go run .                  # API and browser UI on http://localhost:3000
go run ./cmd/uncluster --help
```

The Makefile does not define a test target. Use the Go commands below.

## Test

```sh
go vet ./...
go test ./...
go test -race ./...
```

If you have [golangci-lint](https://golangci-lint.run/) installed, also run:

```sh
golangci-lint run ./...
```

## Pull requests

1. Fork the repository and create a branch from `main`.
2. Keep the change focused. Do not mix unrelated refactors with a bug fix or feature.
3. Make sure `go vet ./...` and `go test ./...` pass.
4. Open a pull request against `main` and fill in the pull request template.

Please do not include generated binaries, temporary output directories, or secrets in a commit.
