# Campus Verde — API Go + PostGIS
# Uso: make bootstrap && make api

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help env setup up docker-up down docker-down wait migrate etl etl-lote sectores api web test backend-test swagger bootstrap counts stack

help:
	@echo "make bootstrap   # compose + migraciones + ETL"
	@echo "make setup       # crea .env desde .env.example sin sobrescribir"
	@echo "make up          # solo PostGIS (el desarrollo local)"
	@echo "make docker-up   # alias docker compose up -d"
	@echo "make docker-down # alias docker compose down"
	@echo "make stack       # postgis + api + web en compose"
	@echo "make wait        # espera a que Postgres acepte conexiones"
	@echo "make migrate     # aplica SQL de db/migrations (backend/app)"
	@echo "make etl         # data/raw → data/v1 → PostGIS"
	@echo "make etl-lote    # upsert del frente 1E (sin TRUNCATE)"
	@echo "make sectores    # data/raw/lote/jefe_de_grupo.json → data/v1/zonas_sector.json"
	@echo "make api         # API nueva en backend/app"
	@echo "make web         # visor MapLibre en http://127.0.0.1:4317"
	@echo "make test        # go test de apps/api"
	@echo "make backend-test # go test de backend/app"
	@echo "make swagger     # regenera docs de swag"
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
	cd backend/app && go run ./cmd/migrate

etl: env
	cd backend/app && go run ./cmd/etl

etl-lote: env
	cd backend/app && go run ./cmd/etl-lote

sectores: env
	cd backend/app && go run ./cmd/sectores

api: env
	cd backend/app && go run ./cmd

web:
	cd apps/web && npm run dev

test:
	cd apps/api && go test ./...

backend-test:
	$(MAKE) -C backend test

swagger:
	$(MAKE) -C backend swagger

bootstrap: up wait migrate etl
	@echo "Listo. Arranca la API con: make api"

counts: env
	./scripts/counts.sh
