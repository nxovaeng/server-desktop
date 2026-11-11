VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
GOPATH_BIN := $(shell go env GOPATH 2>/dev/null)/bin
export PATH := $(GOPATH_BIN):$(PATH)

LINUX_TAGS := $(shell if pkg-config --exists webkit2gtk-4.1 2>/dev/null; then echo "-tags webkit2_41"; fi)

.PHONY: all frontend test vet build package clean

all: build

frontend:
	cd frontend && npm run build

vet:
	go vet $(LINUX_TAGS) ./...

test:
	go test $(LINUX_TAGS) -count=1 ./...

build:
	./scripts/build.sh --skip-test --version "$(VERSION)"

package:
	./scripts/build.sh --version "$(VERSION)" --out dist --package

clean:
	rm -rf build/bin/ dist/ release/ frontend/dist/*
	touch frontend/dist/.gitkeep
