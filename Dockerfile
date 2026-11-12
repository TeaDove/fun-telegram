FROM golang:1.27.4-trixie AS build

WORKDIR /src

ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY main.go ./
COPY internal/ ./internal/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -o=bootstrap .

FROM gcr.io/distroless/static-debian13:latest

WORKDIR /
COPY --from=build /src/bootstrap /bootstrap

CMD ["/bootstrap"]
