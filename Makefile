GO ?= GO111MODULE=on CGO_ENABLED=0 go

run:
	$(GO) run main.go

update:
	git pull
	docker compose up -d --build
	docker compose logs -f backend

