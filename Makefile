.PHONY: install uninstall build clean

# Install the CLI to $GOPATH/bin (available system-wide)
install:
	go install ./cmd/uncluster/

# Remove the installed binary
uninstall:
	rm -f "$$(go env GOPATH)/bin/uncluster" "$$(go env GOPATH)/bin/uncluster.exe"

# Build the binary in the current directory
build:
	go build -o uncluster ./cmd/uncluster/

# Build the API server
build-server:
	go build -o uncluster-server .

# Remove local build artifacts
clean:
	rm -f uncluster uncluster.exe uncluster-server uncluster-server.exe
