all: cli server worker ## Build all binaries for host OS and CPU

-include rules.mk

crosscompile: $(crossbinaries) ## Build all binaries for Linux and all supported CPU arches
images: tink-cli-image tink-server-image tink-worker-image tink-controller-image ## Build all docker images
run: crosscompile run-stack ## Builds and runs the Tink stack (tink, db, cli) via docker-compose

test: ## Run tests
	go clean -testcache
	go test ./... -v

verify: ## Run lint like checkers
	goimports -d .
	golint ./...

help: ## Print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120

rebuild: tink-server-image
	cd ~/go/src/github.com/tinkerbell/sandbox/deploy && docker-compose kill tink-server && docker-compose rm -f tink-server && docker-compose up -d
	sleep 5
	./cmd/tink-cli/tink-cli  workflow get  wf1


GO_INSTALL = ./scripts/go_install.sh
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)

CONTROLLER_GEN_VER := v0.2.9
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)

ENVSUBST_BIN := envsubst
ENVSUBST := $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)-drone

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(CONTROLLER_GEN): ## Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)


.PHONY: generate
generate: ## Generate code, manifests etc.
	$(MAKE) generate-go
	$(MAKE) generate-manifests

.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) # Generate Go code.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./k8s/..."

.PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) # Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./k8s/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
