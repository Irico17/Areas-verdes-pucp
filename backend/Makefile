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

# Introspects the configured PostgreSQL schema without running migrations.
# Requires the database to be available and DATABASE_* to be set in the
# environment or backend/app/.env.
# Output goes to a temp directory or MODELGEN_OUT. Aborts if pointed at campus_verde.
gen-models:
	@echo "Generating GORM models from PostgreSQL..."
	cd app && go run ./cmd/modelgen

swagger:
	@echo "Generating Swagger documentation..."
	cd app && swag init -g cmd/main.go -o docs/
	swagger2openapi app/docs/swagger.json -o openapi/openapi.base.json
	node openapi/sanitize-openapi.js openapi/openapi.base.json
	@echo "Swagger documentation generated in app/docs/"
