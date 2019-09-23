.PHONY: build test clean docker

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-bitflow
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/datenente/device-bitflow.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) install -tags=safe

cmd/device-bitflow:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

docker:
	docker build \
		-f cmd/device-bitflow/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t datenente/docker-device-bitflow:$(GIT_SHA) \
		-t datenente/docker-device-bitflow:$(VERSION)-dev \
		.

test:
	$(GO) vet ./...
	gofmt -l .
	$(GO) test -coverprofile=coverage.out ./...

clean:
	rm -f $(MICROSERVICES) $(MICROSERVICES).log
