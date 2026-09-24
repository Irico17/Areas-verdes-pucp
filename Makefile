# Campus Verde — API Go + PostGIS
# Uso: make bootstrap && make api

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help env up down wait migrate etl api web test bootstrap counts

help:
	@echo "make bootstrap   # compose + migraciones + ETL"
	@echo "make up          # docker compose up -d (PostGIS)"
	@echo "make wait        # espera a que Postgres acepte conexiones"
	@echo "make migrate     # aplica SQL de apps/api/migrations"
	@echo "make etl         # data/raw → data/v1 → PostGIS"
	@echo "make api         # GET /health y /api/v1/geo/..."
	@echo "make web         # visor MapLibre en http://127.0.0.1:4317"
	@echo "make test        # go test ./..."
	@echo "make counts      # conteos en PostGIS"
	@echo "make down        # detiene compose"

env:
	@test -f .env || cp .env.example .env

up: env
	docker compose up -d

down:
	docker compose down

wait: env
	./scripts/wait-db.sh

migrate: env
	cd apps/api && go run ./cmd/migrate

etl: env
	cd apps/api && go run ./cmd/etl

api: env
	cd apps/api && go run ./cmd/api

web:
	cd apps/web && npm run dev

test:
	cd apps/api && go test ./...

bootstrap: up wait migrate etl
	@echo "Listo. Arranca la API con: make api"

counts: env
	./scripts/counts.sh
