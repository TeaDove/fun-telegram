.PHONY: ckeck install upload

CORE_BUILD_VERSION ?= $(shell cat CORE_VERSION)
CORE_DOCKER_IMAGE ?= ghcr.io/teadove/fun-telegram-core:$(CORE_BUILD_VERSION)

docker-login:
	docker login ghcr.io

core-docker-buildx-local:
	rm -f bootstrap-amd64
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o=bootstrap-amd64
	rm -f bootstrap-arm64
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -o=bootstrap-arm64
	docker buildx build --platform linux/amd64,linux/arm64 -f=DockerfileCoreLocal . --tag $(CORE_DOCKER_IMAGE) --push
	rm -f bootstrap-amd64 bootstrap-arm64

core-docker-buildx: docker-login
	docker buildx build --platform linux/arm64,linux/amd64 -f=DockerfileCore . --tag $(CORE_DOCKER_IMAGE) --push

lint:
	golangci-lint run ./...
	golines --base-formatter=gofmt --max-len=120 --no-reformat-tags -w .

test:
	go test ./... -cover -count=1 -p=100

core-run:
	CGO_ENABLED=0 go run main.go

core-run-docker-rebuild:
	docker-compose -f docker-compose-local.yaml up -d --build

update-local:
	git pull
	docker-compose -f docker-compose-local.yaml up -d --build
	docker-compose logs -f client ds

infra-run:
	docker-compose -f docker-compose-infra.yaml up -d

update:
	git pull
	docker-compose down
	docker-compose up -d
	docker-compose logs -f ds core

logs:
	docker-compose logs -f ds core
