hash:=$(shell git describe --tags --always)
buildTime:=$(shell git log -1 --format="%cI")
tag:=$(shell git rev-parse --short HEAD)

.PHONY: build all clean test docker

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o logflow

docker:
	docker build -t logflow .

all:
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-windows-x64-$(tag).exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-windows-386-$(tag).exe
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-linux-x64-$(tag)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-linux-386-$(tag)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-linux-arm64-$(tag)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/logflow-darwin-x64-$(tag)

clean:
	rm -rf logflow

test:
	for _ in {1..5} ; do go test -v -failfast -count=1 -gcflags="all=-N -l" ./... && break ; done
