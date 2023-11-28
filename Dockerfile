# Start by building the application.
FROM golang:1.21-bullseye as build

WORKDIR /src
COPY . .

ENV CGO_ENABLED=1

RUN go mod download
RUN go build -o bootstrap

## Now copy it into our base image.
FROM debian:trixie-slim

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]