FROM --platform=linux/amd64 golang:1.23-bullseye AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ./bin/zportfolio-service ./cmd

RUN GOOS=linux go build -o ./bin/zportfolio-service ./cmd

FROM --platform=linux/amd64 debian:bullseye-slim 

ENV LOGLEVEL=debug

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates && \
	rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/config
COPY --from=builder /app/bin/zportfolio-service .
COPY --from=builder /app/config/config.yml ./config/config.yml
RUN chmod +x /app/zportfolio-service
EXPOSE 8081

ENTRYPOINT ["./zportfolio-service"]

