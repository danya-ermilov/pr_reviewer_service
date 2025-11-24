.PHONY: build run up

build:
	docker build -t prreview:local .

up: build
	docker compose up -d --build

run:
	go run ./cmd/server
