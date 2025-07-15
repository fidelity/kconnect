# Build information
build_date := $(shell date +%Y-%m-%dT%H:%M:%SZ)
git_commit := $(shell git rev-parse --short HEAD)
version_pkg := github.com/fidelity/kconnect/internal/version
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
UNAME := $(shell uname -s)

# Directories
gopath := $(shell go env GOPATH)
GOBIN ?= $(gopath)/bin
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
TOOLS_SHARE_DIR := $(TOOLS_DIR)/share
BIN_DIR := bin
SHARE_DIR := share
PATH := $(abspath $(TOOLS_BIN_DIR)):$(PATH)
export PATH

$(TOOLS_BIN_DIR):
	mkdir -p $@

$(TOOLS_SHARE_DIR):
	mkdir -p $@

$(BIN_DIR):
	mkdir -p $@

# Docs
MDBOOK_VERSION := v0.4.3
BOOKS_DIR := docs/book
RUST_TARGET := unknown-$(OS)-gnu
MDBOOK_EXTRACT_COMMAND := tar xfvz $(TOOLS_SHARE_DIR)/mdbook.tar.gz -C $(TOOLS_BIN_DIR)
MDBOOK_ARCHIVE_EXT := .tar.gz
ifeq ($(OS), windows)
	RUST_TARGET := pc-windows-msvc
	MDBOOK_ARCHIVE_EXT := .zip
	MDBOOK_EXTRACT_COMMAND := unzip -d /tmp
endif

ifeq ($(OS), darwin)
	RUST_TARGET := apple-darwin
endif

# Binaries
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
DEFAULTER_GEN := $(TOOLS_BIN_DIR)/defaulter-gen
CONVERSION_GEN := $(TOOLS_BIN_DIR)/conversion-gen
MOCKGEN := $(TOOLS_BIN_DIR)/mockgen
MDBOOK := $(TOOLS_BIN_DIR)/mdbook
MDBOOK_EMBED := $(TOOLS_BIN_DIR)/mdbook-embed
MDBOOK_RELEASELINK := $(TOOLS_BIN_DIR)/mdbook-releaselink
MDBOOK_TABULATE := $(TOOLS_BIN_DIR)/mdbook-tabulate

.DEFAULT_GOAL := help

##@ Build

.PHONY: build
build: # Build the CLI binary
	CGO_ENABLED=0 go build -ldflags "-X $(version_pkg).commitHash=$(git_commit) -X $(version_pkg).buildDate=$(build_date)" ./cmd/kconnect

.PHONY: build-cross
build-cross: # Build the CLI binary for linux/mac/windows
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X $(version_pkg).commitHash=$(git_commit) -X $(version_pkg).buildDate=$(build_date)" -o out/kconnect_osx ./cmd/kconnect
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X $(version_pkg).commitHash=$(git_commit) -X $(version_pkg).buildDate=$(build_date)" -o out/kconnect_linux ./cmd/kconnect
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X $(version_pkg).commitHash=$(git_commit) -X $(version_pkg).buildDate=$(build_date)" -o out/kconnect_windows.exe ./cmd/kconnect

.PHONY: generate
generate: $(MOCKGEN) $(CONTROLLER_GEN) $(CONVERSION_GEN)  # Generate code for the api definitions
	go generate ./...
	$(CONTROLLER_GEN) \
		paths=./api/... \
		object:headerFile=./hack/boilerplate.generatego.txt

	$(CONVERSION_GEN) \
		./api/v1alpha1 \
		--output-file=zz_generated.conversion \
		--go-header-file=./hack/boilerplate.generatego.txt

##@ Release

.PHONY: release
release: # Builds a release
	goreleaser

.PHONY: release-local
release-local: # Builds a release locally
	goreleaser --snapshot --skip=publish --clean

##@ Test & CI

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: $(GOLANGCI_LINT) # Run the linter across the codebase
	$(GOLANGCI_LINT) run -v

.PHONY: ci
ci: tidy fmt vet test build # Target for CI


##@ Utility

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Get and build golangci-lint
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Get and build controller-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) sigs.k8s.io/controller-tools/cmd/controller-gen

$(DEFAULTER_GEN): $(TOOLS_DIR)/go.mod # Get and build defaulter-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) k8s.io/code-generator/cmd/defaulter-gen

$(CONVERSION_GEN): $(TOOLS_DIR)/go.mod # Get and build conversion-gen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) k8s.io/code-generator/cmd/conversion-gen

$(MOCKGEN): $(TOOLS_DIR)/go.mod # Get and build mockgen
	cd $(TOOLS_DIR); go build -tags=tools -o $(subst hack/tools/,,$@) github.com/golang/mock/mockgen

CMDDOCSGEN:= $(TOOLS_BIN_DIR)/cmddocsgen
$(CMDDOCSGEN):
	go build -tags=tools  -o $(TOOLS_BIN_DIR)/cmddocsgen ./tools/cmddocsgen


##@ Docs

MDBOOK_SHARE := $(TOOLS_SHARE_DIR)/mdbook$(MDBOOK_ARCHIVE_EXT)
$(MDBOOK_SHARE): $(TOOLS_SHARE_DIR)
	curl -sL -o $(MDBOOK_SHARE) "https://github.com/rust-lang/mdBook/releases/download/$(MDBOOK_VERSION)/mdBook-$(MDBOOK_VERSION)-x86_64-$(RUST_TARGET)$(MDBOOK_ARCHIVE_EXT)"

MDBOOK := $(TOOLS_BIN_DIR)/mdbook
$(MDBOOK): $(TOOLS_BIN_DIR) $(MDBOOK_SHARE)
	$(MDBOOK_EXTRACT_COMMAND)
	chmod +x $@
	touch -m $@


.PHONY: docs-build
docs-build: docs-generate $(MDBOOK) ## Build the kconnect book
	$(MDBOOK) build $(BOOKS_DIR)

.PHONY: docs-generate
docs-generate: $(CMDDOCSGEN) ## Generate the cmd line docs
	$(CMDDOCSGEN) $(BOOKS_DIR)/src/commands
	rm -f $(TOOLS_BIN_DIR)/cmddocsgen

.PHONY: docs-verify
docs-verify: docs-generate ## Verify the generated docs are up to date
	cd $(BOOKS_DIR)/src/commands
	@if !(git diff --quiet HEAD ); then \
		git diff; \
		echo "generated command docs are out of date, run make docs-generate"; exit 1; \
	fi

.PHONY: docs-serve
docs-serve: $(MDBOOK) ## Run a local webserver with the compiled book
	$(MDBOOK) serve $(BOOKS_DIR)

.PHONY: docs-clean
docs-clean:
	rm -rf $(BOOKS_DIR)/book


.PHONY: help
help:  ## Display this help. Thanks to https://suva.sh/posts/well-documented-makefiles/
ifeq ($(OS),Windows_NT)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif
