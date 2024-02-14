.PHONY: ckeck install upload

BUILD_VERSION ?= $(shell cat VERSION)
CORE_BUILD_OUTPUT ?= fun-telegram
GO ?= GO111MODULE=on CGO_ENABLED=0 go
GOOS ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f2)
CORE_DOCKER_IMAGE ?= ghcr.io/teadove/fun-telegram-core:$(BUILD_VERSION)
DS_DOCKER_IMAGE ?= ghcr.io/teadove/fun-telegram-ds:$(BUILD_VERSION)
PYTHON ?= .venv/bin/python
PYTHON_PRE ?= ../.venv/bin/python

install:
	python3.11 -m venv .venv
	cd ds && $(PYTHON_PRE) -m pip install poetry
	cd ds && $(PYTHON_PRE) -m poetry update

run-ds:
	cd ds && $(PYTHON_PRE) main.py

docker-login:
	docker login ghcr.io

docker-build: docker-login
	docker buildx build --platform linux/arm64,linux/amd64 DockerfileCore --tag $(CORE_DOCKER_IMAGE) --no-cache --push
	docker buildx build --platform linux/arm64,linux/amd64 DockerfileDS --tag $(DS_DOCKER_IMAGE) --no-cache --push

test-integration:
	go test ./... --run 'TestIntegration_*' -cover -count=1 -p=100

test-unit:
	go test ./... --run 'TestUnit_*' -cover -count=1 -p=100

lint:
	golangci-lint run ./...
	golines --base-formatter=gofmt --max-len=120 --no-reformat-tags -w .

test: test-unit lint test-integration

run-core:
	@$(GO) run main.go

run-docker-rebuild:
	docker-compose -f docker-compose-local.yaml up -d --build

run-infra:
	docker-compose -f docker-compose-infra.yaml up -d

update:
	git pull
	docker-compose up -d
	docker-compose logs -f client

logs:
	docker-compose logs -f client

core-build:
	@$(GO) build -o $(CORE_BUILD_OUTPUT) main.go

core-clean:
	@echo -n ">> CLEAN"
	@$(GO) clean -i ./...
	@rm -f $(CORE_BUILD_OUTPUT)*
	@rm -rf dist/*
	@printf '%s\n' '$(OK)'


core-crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-linux-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-windows-amd64
	@printf '%s\n' '$(OK)'

	@echo -n ">> CROSSCOMPILE linux/arm64"
	@GOOS=linux GOARCH=arm64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-linux-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/arm64"
	@GOOS=darwin GOARCH=arm64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/arm64"
	@GOOS=windows GOARCH=arm64 $(GO) build -o $(CORE_BUILD_OUTPUT)-$(BUILD_VERSION)-windows-arm64
	@printf '%s\n' '$(OK)'
