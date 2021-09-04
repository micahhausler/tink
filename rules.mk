# Only use the recipes defined in these makefiles
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:
# Delete target files if there's an error
# This avoids a failure to then skip building on next run if the output is created by shell redirection for example
# Not really necessary for now, but just good to have already if it becomes necessary later.
.DELETE_ON_ERROR:
# Treat the whole recipe as a one shell script/invocation instead of one-per-line
.ONESHELL:
# Use bash instead of plain sh
SHELL := bash
.SHELLFLAGS := -o pipefail -euc

binaries := cmd/tink-cli/tink-cli cmd/virtual-worker/virtual-worker cmd/tink-server/tink-server cmd/tink-worker/tink-worker cmd/tink-controller/tink-controller
version := $(shell git rev-parse --short HEAD)
tag := $(shell git tag --points-at HEAD)
ifneq (,$(tag))
version := $(tag)-$(version)
endif
LDFLAGS := -ldflags "-X main.version=$(version)"
export CGO_ENABLED := 0

cli: cmd/tink-cli/tink-cli
virtual-worker: cmd/virtual-worker/virtual-worker
server: cmd/tink-server/tink-server
worker : cmd/tink-worker/tink-worker
controller: cmd/tink-controller/tink-controller

.PHONY: server cli worker test virtual-worker $(binaries)
crossbinaries := $(addsuffix -linux-,$(binaries))
crossbinaries := $(crossbinaries:=386) $(crossbinaries:=amd64) $(crossbinaries:=arm64) $(crossbinaries:=armv6) $(crossbinaries:=armv7)

.PHONY: crosscompile $(crossbinaries)
%-386:   FLAGS=GOOS=linux GOARCH=386
%-amd64: FLAGS=GOOS=linux GOARCH=amd64
%-arm64: FLAGS=GOOS=linux GOARCH=arm64
%-armv6: FLAGS=GOOS=linux GOARCH=arm GOARM=6
%-armv7: FLAGS=GOOS=linux GOARCH=arm GOARM=7
$(binaries) $(crossbinaries):
	$(FLAGS) go build $(LDFLAGS) -o $@ ./$(@D)

.PHONY: images tink-cli-image tink-server-image tink-worker-image tink-controller-image virtual-worker-image
tink-cli-image: cmd/tink-cli/tink-cli-linux-amd64
	docker build -t tink-cli cmd/tink-cli/
tink-server-image: cmd/tink-server/tink-server-linux-amd64
	docker build -t tink-server cmd/tink-server/
tink-worker-image: cmd/tink-worker/tink-worker-linux-amd64
	docker build -t tink-worker cmd/tink-worker/
tink-controller-image: cmd/tink-controller/tink-controller-linux-amd64
	docker build -t tink-controller cmd/tink-controller/
virtual-worker-image: cmd/virtual-worker/virtual-worker-linux-amd64
	docker build -t virtual-worker cmd/virtual-worker/

run-stack:
	docker-compose up --build

.PHONY: moq
moq:
	go install github.com/matryer/moq

protos/gen_mock: moq
	rm ./protos/*/mock.go
	go generate ./protos/...
	goimports -w ./protos/**/mock.go

grpc/gen_doc:
	protoc \
		-I./protos \
		-I./protos/third_party \
		--doc_out=./doc \
		--doc_opt=html,index.html \
		protos/hardware/*.proto protos/template/*.proto protos/workflow/*.proto
