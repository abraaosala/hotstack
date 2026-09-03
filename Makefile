.PHONY: build run clean test install build-all release-snapshot

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack.exe ./cli/hotstack

run: build
	./hotstack.exe

clean:
	rm -f hotstack.exe hotstack-*

test:
	go test ./...

vet:
	go vet ./...

install:
	go install -trimpath -ldflags="$(LDFLAGS)" ./cli/hotstack

build-all:
	GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-linux-amd64    ./cli/hotstack
	GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-linux-arm64    ./cli/hotstack
	GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-darwin-amd64   ./cli/hotstack
	GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-darwin-arm64   ./cli/hotstack
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-windows-amd64.exe ./cli/hotstack
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o hotstack-windows-arm64.exe ./cli/hotstack
