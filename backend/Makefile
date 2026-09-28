.PHONY: build run test tidy swagger

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

swagger:
	@echo "Generating Swagger documentation..."
	cd app && swag init -g cmd/main.go -o docs/
	swagger2openapi app/docs/swagger.json -o openapi/openapi.base.json
	node openapi/sanitize-openapi.js openapi/openapi.base.json
	@echo "Swagger documentation generated in app/docs/"
