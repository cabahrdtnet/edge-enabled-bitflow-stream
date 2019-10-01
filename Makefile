.PHONY: build run test clean docker

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-bitflow/device-bitflow
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/datenente/device-bitflow.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) install -tags=safe

cmd/device-bitflow/device-bitflow:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/device-bitflow

docker:
	docker build \
		-f ./Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t datenente/docker-device-bitflow:$(GIT_SHA) \
		-t datenente/docker-device-bitflow:$(VERSION)-dev \
		.

run:
	cd cmd/device-bitflow ; ./device-bitflow --registry=consul://localhost:8500 --confdir=./res/

test:
	$(GO) vet ./...
	gofmt -l .
	$(GO) test -coverprofile=coverage.out ./...

clean:
	rm -f $(MICROSERVICES) $(MICROSERVICES).log
