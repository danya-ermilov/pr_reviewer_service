BINARY := prreview
IMAGE := pr_reviewer_service-app
COMPOSE := docker compose

.PHONY: all build test fmt lint docker-build docker-up docker-down docker-logs clean run migrate

all: build

# локальная сборка бинарника
build:
	go build -o $(BINARY) ./cmd/prreview

# запуск unit-тестов
test:
	go test ./... -v

fmt:
	go fmt ./...

lint:
	go fmt ./...
	go vet ./...
	golangci-lint run --fix --timeout=5m ./...

# собрать docker image
docker-build:
	docker build -t $(IMAGE) .

# поднять стек (DB + app). Использует docker compose v2 (docker compose ...)
docker-up:
	$(COMPOSE) up -d --build

# опустить стек и удалить volume с БД
docker-down:
	$(COMPOSE) down -v

docker-logs:
	$(COMPOSE) logs -f

# чистка артефактов
clean:
	-rm -f $(BINARY)

# запуск миграций (локально, если бинарь умеет migrate)
migrate:
	./$(BINARY) migrate

# локальный запуск (использует переменные окружения)
run:
	DATABASE_URL=$${DATABASE_URL:-postgres://pruser:prpass@localhost:5432/pr_review?sslmode=disable} PORT=$${PORT:-8080} ./$(BINARY)
