.PHONY: build run test tidy swagger gen-models

build:
	@echo "Building project..."
	cd app && go build -o bin/api ./cmd

run:
	@echo "Starting Areas-Verdes service..."
	cd app && go run ./cmd

test:
	@echo "Running tests..."
	cd app && go test ./...

tidy:
	@echo "Tidying modules..."
	cd app && go mod tidy

# Genera modelos GORM y ejecuta control de deriva contra una base de datos desechable
# construida a partir de nuestras migraciones (apps/api/migrations).
# Requiere una base de datos desechable (prefijo vp_ o modelgen_) migrada con cmd/migrate.
# La salida va a un directorio temporal o MODELGEN_OUT. Aborta si apunta a campus_verde o bases con datos.
gen-models:
	@echo "Ejecutando modelgen y control de deriva contra PostgreSQL..."
	cd app && go run ./cmd/modelgen

swagger:
	@echo "Generating Swagger documentation..."
	cd app && swag init -g cmd/main.go -o docs/
	swagger2openapi app/docs/swagger.json -o openapi/openapi.base.json
	node openapi/sanitize-openapi.js openapi/openapi.base.json
	@echo "Swagger documentation generated in app/docs/"
