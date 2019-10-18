# CHANGED BY CHRISTIAN ALEXANDER BAHRDT
# This file is derivative from the following file 
# https://github.com/edgexfoundry/device-sdk-go/blob/edinburgh/Makefile

.PHONY: run test clean
.PHONY: build build-engine build-service
.PHONY: docker docker-device-bitflow-stream docker-engine

GO=CGO_ENABLED=0 GO111MODULE=on go

SERVICE=cmd/device-bitflow-stream/device-bitflow-stream
ENGINE=cmd/engine/engine
.PHONY: $(SERVICE)
.PHONY: $(ENGINE)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/datenente/device-bitflow-stream.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(SERVICE) $(ENGINE)
	$(GO) install -tags=safe

build-engine: $(ENGINE)
	$(GO) install -tags=safe

build-service: $(SERVICE)
	$(GO) install -tags=safe

cmd/device-bitflow-stream/device-bitflow-stream:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/device-bitflow-stream

cmd/engine/engine:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/engine

docker: docker-device-bitflow-stream docker-engine
	echo "docker..."

docker-device-bitflow-stream:
	docker build \
		-f ./Dockerfile_Device_Service \
		--label "git_sha=$(GIT_SHA)" \
#		-t datenente/device-bitflow-stream:$(GIT_SHA) \
		-t datenente/device-bitflow-stream:$(VERSION) \
		.

docker-engine:
	docker build \
		-f ./Dockerfile_Script_Execution_Engine \
		--label "git_sha=$(GIT_SHA)" \
#		-t datenente/device-bitflow-stream-engine:$(GIT_SHA) \
		-t datenente/device-bitflow-stream-engine:$(VERSION) \
        .

run:
	cd cmd/device-bitflow-stream ; ./device-bitflow-stream --registry=consul://localhost:8500 --confdir=./res/

test:
	$(GO) vet ./...
	gofmt -l .
	$(GO) test -tags unit -coverprofile=coverage.out ./...
	$(GO) test -tags package -coverprofile=coverage.out ./...

clean:
	rm -f $(SERVICE) $(SERVICE).log
	rm -f $(ENGINE)
