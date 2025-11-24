# PR Reviewer Service

Репозиторий содержит сервис PR-review (Go), Dockerfile и docker-compose для быстрого старта, а также инструкции по тестированию и нагрузочному тестированию.

## Что должно быть в `main`/`master`
- код сервиса
- `Makefile` (с командами сборки)
- `docker-compose.yaml`, `Dockerfile`, `entrypoint.sh`
- `README.md` с инструкцией запуска и описанием проблем/решений

---

## Быстрый старт (Docker)

Требования:
- Docker Desktop / Docker Engine
- (в WSL) включена интеграция WSL в Docker Desktop или запускать из Linux/macOS
- `docker compose` (вместо `docker-compose`) — в Makefile используется `docker compose`

Запуск:

```bash
# 1) Сброс старых контейнеров/томов (рекомендуется при первой сборке)
docker compose down -v

# 2) Поднять стек (Postgres + приложение)
docker compose up -d --build
