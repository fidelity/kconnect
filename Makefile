build_date := $(shell date +%Y-%m-%dT%H:%M:%SZ)
git_commit := $(shell git rev-parse --short HEAD)

version_pkg := github.com/fidelity/kconnect/internal/version

# Directories
gopath := $(shell go env GOPATH)
GOBIN ?= $(gopath)/bin
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin

# Binaries
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/golangci-lint)
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
DEFAULTER_GEN := $(TOOLS_BIN_DIR)/defaulter-gen
CONVERSION_GEN := $(TOOLS_BIN_DIR)/conversion-gen
MOCKGEN := $(TOOLS_BIN_DIR)/mockgen

.DEFAULT_GOAL := help

##@ Build

.PHONY: build
build: # Build the CLI binary
	CGO_ENABLED=0 go build -ldflags "-X $(version_pkg).commitHash=$(git_commit) -X $(version_pkg).buildDate=$(build_date)" .

.PHONY: generate
generate: $(MOCKGEN) $(CONTROLLER_GEN) $(CONVERSION_GEN)  # Generate code for the api definitions
	go generate
	$(CONTROLLER_GEN) \
		paths=./pkg/history/api/... \
		object:headerFile=./hack/boilerplate.generatego.txt

	$(CONVERSION_GEN) \
		--input-dirs=./pkg/history/api/v1alpha1 \
		--output-file-base=zz_generated.conversion \
		--go-header-file=./hack/boilerplate.generatego.txt

##@ Release

.PHONY: release
release: # Builds a release
	goreleaser

.PHONY: release-local
release-local: # Builds a relase locally
	goreleaser --snapshot --skip-publish --rm-dist

##@ Test & CI

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: $(GOLANGCI_LINT) # Run the linter across the codebase
	$(GOLANGCI_LINT) run -v

.PHONY: ci
ci: build test lint # Target for CI


##@ Utility

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Get and build golangci-lint
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) github.com/golangci/golangci-lint/cmd/golangci-lint

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Get and build controller-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) sigs.k8s.io/controller-tools/cmd/controller-gen

$(DEFAULTER_GEN): $(TOOLS_DIR)/go.mod # Get and build defaulter-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) k8s.io/code-generator/cmd/defaulter-gen

$(CONVERSION_GEN): $(TOOLS_DIR)/go.mod # Get and build conversion-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) k8s.io/code-generator/cmd/conversion-gen

$(MOCKGEN): $(TOOLS_DIR)/go.mod # Get and build mockgen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) github.com/golang/mock/mockgen


.PHONY: help
help:  ## Display this help. Thanks to https://suva.sh/posts/well-documented-makefiles/
ifeq ($(OS),Windows_NT)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif
