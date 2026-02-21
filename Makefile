.PHONY: dev build docker up down

dev:
	go run cmd/main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/tg-manager ./cmd

docker:
	@echo "Building Docker image..."
	@docker build -t tg-manager:latest .

up:
	docker compose up -d

down:
	docker compose down
