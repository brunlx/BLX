BIN      := bin/blx
PKG      := ./...

VERSION  ?= dev
LDFLAGS  := -s -w -X main.version=$(VERSION)

BIN_WIN  := bin/blx.exe
RSRC     := $(shell go env GOPATH)/bin/rsrc

.PHONY: all build build-windows build-icon build-installer run test test-race vet fmt clean

all: build

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_WIN) ./cmd/server

build-icon:
	$(RM) cmd/server/rsrc_windows_amd64.syso
	go run ./cmd/genicon
	$(RSRC) -ico assets/icon.ico -o cmd/server/rsrc_windows_amd64.syso

build-installer: build-windows
	wine "C:/Program Files/Inno Setup 7/ISCC.exe" /F"BLX-Setup" installer.iss

run: build
	PORT=8080 $(BIN)

test:
	go test $(PKG) -count=1

test-race:
	go test $(PKG) -count=1 -race

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

clean:
	rm -rf bin
