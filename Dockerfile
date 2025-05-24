FROM golang:1.24 AS builder
WORKDIR /app

COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build pkg/cmd/main.go



FROM alpine:edge AS release

WORKDIR /

COPY --from=builder /app/main /main

EXPOSE 8080



ENTRYPOINT ["/main"]
