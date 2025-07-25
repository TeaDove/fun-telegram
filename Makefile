.PHONY: ckeck install upload

BUILD_VERSION ?= $(shell cat VERSION)
DOCKER_IMAGE ?= ghcr.io/teadove/fun-telegram:$(BUILD_VERSION)

docker-login:
	docker login ghcr.io

docker-buildx-local:
	rm -f bootstrap-amd64
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o=bootstrap-amd64
	rm -f bootstrap-arm64
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -o=bootstrap-arm64
	docker buildx build --platform linux/amd64,linux/arm64 -f=DockerfileLocal . --tag $(DOCKER_IMAGE) --push
	rm -f bootstrap-amd64 bootstrap-arm64

docker-buildx: docker-login
	docker buildx build --platform linux/arm64,linux/amd64 -f=Dockerfile . --tag $(DOCKER_IMAGE) --push

lint:
	gofumpt -w .
	golines --base-formatter=gofmt --max-len=120 --no-reformat-tags -w .
	#golangci-lint run ./...

test:
	go test ./... -cover -count=1 -p=100

run:
	CGO_ENABLED=1 go run main.go

run-docker-rebuild:
	docker-compose -f docker-compose-local.yaml up -d --build

infra-run:
	docker-compose -f docker-compose-infra.yaml up -d

update:
	git pull
	docker-compose up -d
	docker-compose logs -f ds core

logs:
	docker-compose logs -f ds core
