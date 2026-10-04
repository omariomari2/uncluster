# Uncluster

Uncluster converts HTML into editable web projects.
It is written in Go and provides a command-line tool, an API, and a browser interface.

It can separate CSS and JavaScript, collect page assets, and generate React or Express/EJS projects.
The conversion uses fixed rules. It does not use a language model.

## Run locally

Use Go 1.21 or later.
Run these commands from the repository directory:

```sh
go run .
```

Open `http://localhost:3000`.

To use the command-line tool:

```sh
go run ./cmd/uncluster --help
go run ./cmd/uncluster ./page.html -to nodejs -out ./react-site
```

Replace `page.html` with an HTML file you can use.
Use Node.js 18 or later to run the generated Vite project.

## Run the checks

```sh
go build ./...
go test ./...
```

## Current limits

Conversion preserves page structure instead of inferring application behavior.
Review generated components, dependencies, and asset rights before deployment.

Remote requests have destination checks and download limits.
These controls do not establish permission to copy a website.

## Details

- [Commands, API, and output formats](docs/usage.md)
- [Conversion intent](docs/intent/tsx-conversion.md)
- [Command-line source](cmd/uncluster/main.go)

