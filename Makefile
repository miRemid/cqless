API_VERSION := v1

GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
TARGET_OS=linux
USE_CGO=0
BINARY_NAME=cqless
VERSION?=0.0.0
SERVICE_PORT?=3000
DOCKER_REGISTRY?=kamir3mid/ #if set it should finished by /
EXPORT_RESULT?=false # for CI please set EXPORT_RESULT to true

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

GEN_DIR := $(API_VERSION)/pkg
PROTO_FILES := $(wildcard $(API_VERSION)/pb/*.proto)
define exec-command
$(1)

endef

.PHONY: all test build vendor gen

all: help

## Build:
build-test:
	mkdir -p build/bin
	$(GOCMD) build -o build/bin/$(BINARY_NAME) $(API_VERSION)/*.go
build: ## Build your project and put the output binary in out/bin/
	mkdir -p build/bin
	CGO_ENABLED=$(USE_CGO) $(GOCMD) build -a -ldflags '-s' -o build/bin/$(BINARY_NAME) $(API_VERSION)/*.go
build-vendor:
	mkdir -p build/bin
	CGO_ENABLED=$(USE_CGO) $(GOCMD) build -a -ldflags '-s' -mod vendor -o build/bin/$(BINARY_NAME) $(API_VERSION)/*.go

clean: ## Remove build related file
	rm -fr ./build
	rm -rf ${HOME}/.local/share/cqless

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	$(GOCMD) mod vendor

watch: ## Run the code with cosmtrek/air to have automatic reload on changes
	$(eval PACKAGE_NAME=$(shell head -n 1 go.mod | cut -d ' ' -f2))
	docker run -it --rm -w /go/src/$(PACKAGE_NAME) -v $(shell pwd):/go/src/$(PACKAGE_NAME) -p $(SERVICE_PORT):$(SERVICE_PORT) cosmtrek/air

## Test:
test: ## Run the tests of the project
ifeq ($(EXPORT_RESULT), true)
	go get -u github.com/jstemmer/go-junit-report
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | go-junit-report -set-exit-code > junit-report.xml)
endif
	$(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)

coverage: ## Run the tests of the project and export the coverage
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
ifeq ($(EXPORT_RESULT), true)
	go get -u github.com/AlekSi/gocov-xml
	go get -u github.com/axw/gocov/gocov
	gocov convert profile.cov | gocov-xml > coverage.xml
endif

## Lint:
lint: lint-go lint-dockerfile lint-yaml ## Run all available linters

lint-dockerfile: ## Lint your Dockerfile
# If dockerfile is present we lint it.
ifeq ($(shell test -e ./Dockerfile && echo -n yes),yes)
	$(eval CONFIG_OPTION = $(shell [ -e $(shell pwd)/.hadolint.yaml ] && echo "-v $(shell pwd)/.hadolint.yaml:/root/.config/hadolint.yaml" || echo "" ))
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--format checkstyle" || echo "" ))
	$(eval OUTPUT_FILE = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "| tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -i $(CONFIG_OPTION) hadolint/hadolint hadolint $(OUTPUT_OPTIONS) - < ./Dockerfile $(OUTPUT_FILE)
endif

lint-go: ## Use golintci-lint on your project
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--out-format checkstyle ./... | tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run --deadline=65s $(OUTPUT_OPTIONS)

lint-yaml: ## Use yamllint on the yaml file of your projects
ifeq ($(EXPORT_RESULT), true)
	go get -u github.com/thomaspoignant/yamllint-checkstyle
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | yamllint-checkstyle > yamllint-checkstyle.xml)
endif
	docker run --rm -it -v $(shell pwd):/data cytopia/yamllint -f parsable $(shell git ls-files '*.yml' '*.yaml') $(OUTPUT_OPTIONS)

## Docker:
docker-build: ## Use the dockerfile to build the container
	docker build --rm --tag $(BINARY_NAME) -f docker/Dockerfile .

docker-release: ## Release the container with tag latest and version
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

.PHONY: gen
gen:
	rm -rf ./$(GEN_DIR)/pb	
	mkdir -p ./$(GEN_DIR)/pb
	$(foreach file,$(PROTO_FILES),$(call exec-command, protoc --go_out ./$(GEN_DIR) --go-grpc_out ./$(GEN_DIR) $(file)))
	$(foreach file,$(PROTO_FILES),$(call exec-command, protoc --grpc-gateway_out ./$(GEN_DIR) --grpc-gateway_opt logtostderr=true $(file)))