# Start by building the application.
FROM golang:1.20-bullseye as build

WORKDIR /src
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11
COPY --from=build /src/app /
CMD ["/app"]