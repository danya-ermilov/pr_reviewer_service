# PR Reviewer Service

Репозиторий содержит сервис PR-review (Go), Dockerfile и docker-compose для быстрого старта, а также инструкции по тестированию.

---

## Быстрый старт (Docker)

Требования:
- Docker Desktop / Docker Engine
- (в WSL) включена интеграция WSL в Docker Desktop или запускать из Linux/macOS

Запуск:

```bash
# 1) Сброс старых контейнеров/томов
docker compose down -v

# 2) Поднять стек
docker compose up -d --build
```

Сервис будет доступен на:
```
http://localhost:8080
```

Swagger UI:
```
http://localhost:8080/docs/#/
```

Миграции применяются автоматически — запуском управляет entrypoint.sh.

## Makefile

Основные команды:
```bash
make build          # локальная сборка бинаря
make test           # запуск unit-тестов
make fmt            # форматирование
make lint           # запуск линтера (golangci-lint)
make docker-build   # сборка Docker image
make docker-up      # поднять сервис (docker compose up -d --build)
make docker-down    # остановить и очистить (docker compose down -v)
make run            # локальный запуск
make migrate        # применение миграций
```

## Структура проекта
```bash
├── Dockerfile
├── Makefile
├── docker-compose.yaml
├── entrypoint.sh
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   └── prreview/
│       └── main.go
├── internal/
│   ├── app/
│   ├── config/
│   ├── handlers/
│   ├── models/
│   ├── repo/
│   ├── server/
│   └── services/
├── migrations/
│   └── 0001_init.sql
└── swagger-ui/
```
---

## Принятые решения и допущения

1. Миграции

Применяются автоматически в entrypoint.sh после проверки готовности PostgreSQL.

2. Локальная БД

Если в системе работает локальный PostgreSQL на порту 5432 — он мешает контейнеру.
Поэтому Docker-сервис экспортируется на 5433:

```bash
ports:
  - "5433:5432"
```

3. Повторение id пользователя

Если создается новая команда с id пользователя, которое уже существует, то пользователь удалится из старой команды и существует только в новой.
