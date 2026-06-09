include .env
export

MIGRATE = migrate -path ./migrations -database "$(DATABASE_URL)"

.PHONY: run test lint swag migrate-up migrate-down docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

# Regenerate Swagger docs from annotations — run after any handler/DTO change.
swag:
	swag init -g cmd/api/main.go --parseDependency --output docs

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down
