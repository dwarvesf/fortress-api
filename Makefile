APP_NAME=fortress-api
DEFAULT_PORT=8200
POSTGRES_TEST_CONTAINER?=fortress_local_test

.PHONY: setup init build dev test migrate-up migrate-down

setup:
	go install github.com/rubenv/sql-migrate/...@latest
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/vektra/mockery/v2@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/cosmtrek/air@latest

init: setup
	go install github.com/rubenv/sql-migrate/...@latest
	make remove-infras
	docker-compose up -d
	@echo "Waiting for database connection..."
	@while ! docker exec fortress_local pg_isready > /dev/null; do \
		sleep 1; \
	done
	@while ! docker exec $(POSTGRES_TEST_CONTAINER) pg_isready > /dev/null; do \
		sleep 1; \
	done
	make migrate-up
	make migrate-test
	make seed-db

remove-infras:
	docker-compose down --remove-orphans

build:
	env GOOS=darwin GOARCH=amd64 go build -o bin ./...

dev:
	go run ./cmd/server/main.go

air:
	air -c .air.toml

test:
	@PROJECT_PATH=$(shell pwd) go test -cover ./...

migrate-test:
	sql-migrate up -env=test

migrate-new:
	sql-migrate new -env=local ${name}

migrate-up:
	sql-migrate up -env=local

migrate-down:
	sql-migrate down -env=local

docker-build:
	docker build \
	--build-arg DEFAULT_PORT="${DEFAULT_PORT}" \
	-t ${APP_NAME}:latest .

seed-db:
	@docker exec -t fortress-postgres sh -c "mkdir -p /seed"
	@docker exec -t fortress-postgres sh -c "rm -rf /seed/*"
	@docker cp migrations/seed fortress-postgres:/
	@docker exec -t fortress-postgres sh -c "PGPASSWORD=postgres psql -U postgres -d fortress_local -f /seed/seed.sql"

gen-mock:
	echo "add later"

gen-swagger:
	swag init  --parseDependency -g ./cmd/server/main.go

