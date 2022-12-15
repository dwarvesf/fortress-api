APP_NAME=fortress-api
DEFAULT_PORT=8200
POSTGRES_TEST_CONTAINER?=fortress_local_test
POSTGRES_CONTAINER?=fortress_local
TOOLS_IMAGE=fortress-tools
APP_ENVIRONMENT=docker run --rm -v ${PWD}:/${APP_NAME} -w /${APP_NAME} --net=host ${TOOLS_IMAGE}

.PHONY: setup init build dev test migrate-up migrate-down

setup:
	docker build -f ./Dockerfile.tools -t ${TOOLS_IMAGE} .

init: setup
	make remove-infras
	docker-compose up -d
	@echo "Waiting for database connection..."
	@while ! docker exec ${POSTGRES_CONTAINER} pg_isready > /dev/null; do \
		sleep 1; \
	done
	@while ! docker exec $(POSTGRES_TEST_CONTAINER) pg_isready > /dev/null; do \
		sleep 1; \
	done
	make migrate-up
	make migrate-test
	make seed
	make seed-test

seed:
	@docker exec -t $(POSTGRES_CONTAINER) sh -c "mkdir -p /seed"
	@docker exec -t $(POSTGRES_CONTAINER) sh -c "rm -rf /seed/*"
	@docker cp migrations/seed $(POSTGRES_CONTAINER):/
	@docker exec -t $(POSTGRES_CONTAINER) sh -c "PGPASSWORD=postgres psql -U postgres -d $(POSTGRES_CONTAINER) -f /seed/seed.sql"

seed-test:
	@docker exec -t $(POSTGRES_TEST_CONTAINER) sh -c "mkdir -p /seed"
	@docker exec -t $(POSTGRES_TEST_CONTAINER) sh -c "rm -rf /seed/*"
	@docker cp migrations/test_seed $(POSTGRES_TEST_CONTAINER):/
	@docker exec -t $(POSTGRES_TEST_CONTAINER) sh -c "PGPASSWORD=postgres psql -U postgres -d $(POSTGRES_TEST_CONTAINER) -f /test_seed/seed.sql"

remove-infras:
	docker-compose down --remove-orphans

build:
	env GOOS=darwin GOARCH=amd64 go build -o bin ./...

dev:
	go run ./cmd/server/main.go

test: reset-test-db
	@PROJECT_PATH=$(shell pwd) go test -cover ./... -count=1 -p=1 

migrate-test:
	${APP_ENVIRONMENT} sql-migrate up -env=test

migrate-new:
	${APP_ENVIRONMENT} sql-migrate new -env=local ${name}

migrate-up:
	${APP_ENVIRONMENT} sql-migrate up -env=local

migrate-down:
	${APP_ENVIRONMENT} sql-migrate down -env=local

docker-build:
	docker build \
	--build-arg DEFAULT_PORT="${DEFAULT_PORT}" \
	-t ${APP_NAME}:latest .

reset-db:
	${APP_ENVIRONMENT} sql-migrate down -env=local -limit=0
	${APP_ENVIRONMENT} sql-migrate up -env=local
	make seed

reset-test-db:
	${APP_ENVIRONMENT} sql-migrate down -env=test -limit=0
	${APP_ENVIRONMENT} sql-migrate up -env=test
	make seed-test

gen-mock:
	echo "add later"

gen-swagger:
	${APP_ENVIRONMENT} swag init --parseDependency -g ./cmd/server/main.go
