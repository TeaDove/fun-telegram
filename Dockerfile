# Start by building the application.
FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY . .

ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o bootstrap

## Now copy it into our base image.
FROM debian:trixie

RUN rm -rf /var/lib/apt/lists/* && apt-get update && apt-get install -y --no-install-recommends ca-certificates curl
RUN update-ca-certificates
RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]