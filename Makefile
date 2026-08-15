BIN      := bin/blx
PKG      := ./...

VERSION  ?= dev
LDFLAGS  := -s -w -X main.version=$(VERSION)

BIN_WIN  := bin/blx.exe
WINVER   ?= 0.2.0
WINRES   := $(shell go env GOPATH)/bin/go-winres
WINRES_JSON := winres/winres.json

.PHONY: all build build-windows winres build-icon build-installer run test test-race vet fmt clean

all: build

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/server

$(WINRES):
	go install github.com/tc-hib/go-winres@v0.3.3

winres: $(WINRES)
	$(WINRES) make --in $(WINRES_JSON) --out cmd/server/rsrc --arch amd64 \
		--product-version=$(WINVER) --file-version=$(WINVER)

build-windows: winres
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_WIN) ./cmd/server

build-icon:
	$(RM) cmd/server/rsrc_windows_*.syso
	go run ./cmd/genicon
	$(WINRES) make --in $(WINRES_JSON) --out cmd/server/rsrc --arch amd64 \
		--product-version=$(WINVER) --file-version=$(WINVER)

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
	rm -f cmd/server/rsrc_windows_*.syso
