FROM golang:1.23-alpine AS builder

ENV GO111MODULE=on
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ./bin/zportfolio-service ./cmd

FROM alpine:latest

ENV LOGLEVEL=debug
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/zportfolio-service .
COPY --from=builder /app/config/config.yml ./config/config.yml

EXPOSE 8081

RUN ls -R /app
ENTRYPOINT ["./zportfolio-service"]