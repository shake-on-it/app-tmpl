ifneq (,$(wildcard ./.env))
	include .env
	export
endif

GOCMD=go
GOTEST=${GOCMD} test
GOVET=${GOCMD} vet
BINARY_NAME=backend
SERVICE_NAME=shakeonit
GIT_HASH?=${shell git rev-parse HEAD}
BUILD_TIME?=${shell date -u +'%Y-%m-%dT%H:%M:%SZ'}

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor

all: help

## Build:
build: ## Build project and place binary in public/bin
ifeq (${go_os},)
  override go_os=linux
endif
ifeq (${go_arch},)
  override go_arch=amd64
endif

build:
	mkdir -p public/bin
	GOOS=${go_os} GOARCH=${go_arch} GO111MODULE=on ${GOCMD} build -ldflags="\
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerEnv=${env}' \
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerGitHash=${GIT_HASH}' \
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerBuildTime=${BUILD_TIME}'" \
    -o public/bin/${BINARY_NAME} \
    cmd/service/service.go

start: ## Run project locally
ifeq (${env},)
  override env=local
endif

start:
	${GOCMD} run -ldflags="\
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerEnv=${env}' \
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerGitHash=${GIT_HASH}' \
        -X 'github.com/shake-on-it/app-tmpl/backend/common.ServerBuildTime=${BUILD_TIME}'" \
    cmd/service/service.go \
    --config etc/configs/${env}.json

clean: ## Remove build files
	rm -rf ./public

generate: ## Generate code files
	go generate ./...

transfer: ## Refresh systemd with latest binary
	sudo systemctl stop ${SERVICE_NAME}
	sudo cp public/bin/${BINARY_NAME} ~/go/bin/${SERVICE_NAME}
	sudo systemctl start ${SERVICE_NAME}

deploy: build transfer ## Build and transfer backend service

## Test:
test: ## Run all tests
	${GOTEST} -v -race ./...

## Lint:
lint: ## Lint all files
	echo 'linting...'

## Help
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
