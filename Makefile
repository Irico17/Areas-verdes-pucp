# Campus Verde — API Go + PostGIS
# Uso: make bootstrap && make api

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help env setup up docker-up down docker-down wait migrate etl etl-lote sectores api web test bootstrap counts stack

help:
	@echo "make bootstrap   # compose + migraciones + ETL"
	@echo "make setup       # crea .env desde .env.example sin sobrescribir"
	@echo "make up          # solo PostGIS (el desarrollo local)"
	@echo "make docker-up   # alias docker compose up -d"
	@echo "make docker-down # alias docker compose down"
	@echo "make stack       # postgis + api + web en compose"
	@echo "make wait        # espera a que Postgres acepte conexiones"
	@echo "make migrate     # aplica SQL de apps/api/migrations"
	@echo "make etl         # data/raw → data/v1 → PostGIS"
	@echo "make etl-lote    # upsert del frente 1E (sin TRUNCATE)"
	@echo "make sectores    # data/raw/lote/jefe_de_grupo.json → data/v1/zonas_sector.json"
	@echo "make api         # GET /health y /api/v1/geo/..."
	@echo "make web         # visor MapLibre en http://127.0.0.1:4317"
	@echo "make test        # go test ./..."
	@echo "make counts      # conteos en PostGIS"
	@echo "make down        # detiene compose"

env:
	@test -f .env || cp .env.example .env

setup:
	@if [ -f .env ]; then \
		echo ".env ya existe, no se sobrescribe"; \
	elif [ -f .env.example ]; then \
		cp .env.example .env && echo ".env creado desde .env.example"; \
	elif [ -f backend/app/.env.example ]; then \
		cp backend/app/.env.example .env && echo ".env creado desde backend/app/.env.example"; \
	fi

up: env
	docker compose up -d db

docker-up:
	docker compose up -d

stack: env
	docker compose up -d --build

down:
	docker compose down

docker-down:
	docker compose down

wait: env
	./scripts/wait-db.sh

migrate: env
	cd apps/api && go run ./cmd/migrate

etl: env
	cd apps/api && go run ./cmd/etl

etl-lote: env
	cd apps/api && go run ./cmd/etl-lote

sectores: env
	cd apps/api && go run ./cmd/sectores

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
