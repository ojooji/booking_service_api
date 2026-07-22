.PHONY: run build test migrate-up migrate-down swagger docker-up docker-down tidy

APP_NAME ?= booking-service-api
BUILD_DIR ?= ./bin

run:
	go run ./cmd/server/main.go

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

test:
	go test ./... -v -count=1

tidy:
	go mod tidy
	go mod verify

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

swagger:
	swag init -g cmd/server/main.go -o docs

docker-up:
	docker compose up -d

docker-down:
	docker compose down

coverage:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
