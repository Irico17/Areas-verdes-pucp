# Campus Verde — API Go + PostGIS
# Uso: make bootstrap && make api

# Guardar si el llamador definió puertos en el entorno o línea de comandos
CALLER_HAS_POSTGRES_PORT := $(filter environment command line,$(origin POSTGRES_PORT))
CALLER_HAS_API_PORT := $(filter environment command line,$(origin API_PORT))
CALLER_HAS_WEB_PORT := $(filter environment command line,$(origin WEB_PORT))
CALLER_POSTGRES_PORT := $(POSTGRES_PORT)
CALLER_API_PORT := $(API_PORT)
CALLER_WEB_PORT := $(WEB_PORT)

ifneq (,$(wildcard .env))
include .env
export
endif

# Restaurar valores del llamador sobre .env si existían
ifneq ($(CALLER_HAS_POSTGRES_PORT),)
POSTGRES_PORT := $(CALLER_POSTGRES_PORT)
endif
ifneq ($(CALLER_HAS_API_PORT),)
API_PORT := $(CALLER_API_PORT)
endif
ifneq ($(CALLER_HAS_WEB_PORT),)
WEB_PORT := $(CALLER_WEB_PORT)
endif

DEPLOY_PORT_OVERRIDES :=
ifneq ($(CALLER_HAS_POSTGRES_PORT),)
DEPLOY_PORT_OVERRIDES += POSTGRES_PORT="$(CALLER_POSTGRES_PORT)"
else
DEPLOY_PORT_OVERRIDES += POSTGRES_PORT=
endif
ifneq ($(CALLER_HAS_API_PORT),)
DEPLOY_PORT_OVERRIDES += API_PORT="$(CALLER_API_PORT)"
else
DEPLOY_PORT_OVERRIDES += API_PORT=
endif
ifneq ($(CALLER_HAS_WEB_PORT),)
DEPLOY_PORT_OVERRIDES += WEB_PORT="$(CALLER_WEB_PORT)"
else
DEPLOY_PORT_OVERRIDES += WEB_PORT=
endif

.PHONY: help env setup up docker-up down docker-down wait migrate etl etl-lote sectores api web test backend-test swagger bootstrap counts stack smoke rollback

help:
	@echo "make bootstrap   # compose + migraciones + ETL"
	@echo "make setup       # crea .env desde .env.example sin sobrescribir"
	@echo "make up          # solo PostGIS del compose histórico (sin ENV)"
	@echo "make up ENV=develop|qa|produccion  # stack de ese ambiente"
	@echo "make down ENV=develop|qa|produccion"
	@echo "make smoke ENV=develop|qa|produccion"
	@echo "make rollback ENV=develop|qa|produccion [TAG=...]"
	@echo "make docker-up   # alias docker compose up -d (compose histórico)"
	@echo "make docker-down # alias docker compose down"
	@echo "make stack       # postgis + api + web en compose"
	@echo "make wait        # espera a que Postgres acepte conexiones"
	@echo "make migrate     # aplica SQL de db/migrations (backend/app)"
	@echo "make etl         # data/raw → data/v1 → PostGIS"
	@echo "make etl-lote    # upsert del frente 1E (sin TRUNCATE)"
	@echo "make sectores    # data/raw/lote/jefe_de_grupo.json → data/v1/zonas_sector.json"
	@echo "make api         # API nueva en backend/app"
	@echo "make web         # visor MapLibre en http://127.0.0.1:4317"
	@echo "make test        # go test de backend/app"
	@echo "make backend-test # alias de make test"
	@echo "make swagger     # regenera docs de swag"
	@echo "make counts      # conteos en PostGIS"
	@echo "make down        # detiene el compose histórico, o ENV=... un ambiente"

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
	@if [ -n "$(ENV)" ]; then \
		$(DEPLOY_PORT_OVERRIDES) bash deploy/deploy.sh "$(ENV)" --local; \
	else \
		docker compose up -d db; \
	fi

docker-up:
	docker compose up -d

stack: env
	docker compose up -d --build

down:
	@if [ -n "$(ENV)" ]; then \
		$(DEPLOY_PORT_OVERRIDES) bash deploy/deploy.sh "$(ENV)" --local --down; \
	else \
		docker compose down; \
	fi

docker-down:
	docker compose down

smoke:
	@test -n "$(ENV)" || { echo "make smoke necesita ENV=develop|qa|produccion"; exit 1; }
	$(DEPLOY_PORT_OVERRIDES) bash deploy/deploy.sh "$(ENV)" --local --smoke

rollback:
	@test -n "$(ENV)" || { echo "make rollback necesita ENV=develop|qa|produccion"; exit 1; }
	$(DEPLOY_PORT_OVERRIDES) bash deploy/deploy.sh "$(ENV)" --local --rollback $(TAG)

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
	cd frontend && npm run dev

test:
	$(MAKE) -C backend test

backend-test: test

swagger:
	$(MAKE) -C backend swagger

bootstrap:
	docker compose up -d db
	$(MAKE) wait migrate etl
	@echo "Listo. Arranca la API con: make api"

counts: env
	@if [ -n "$(ENV)" ]; then \
		$(DEPLOY_PORT_OVERRIDES) bash deploy/conteos.sh "$(ENV)"; \
	else \
		./scripts/counts.sh; \
	fi
