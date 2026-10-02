package routes_test

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestSwaggerRouteCoverage walks engine.Routes() and requires every /areas-verdes/v1 route
// to appear in swagger.json. The documented operations must match the routes actually registered.
func TestSwaggerRouteCoverage(t *testing.T) {
	engine := setupTestRouter(true)

	// Try relative paths from internal/presentation/routes
	candidates := []string{
		"../../docs/swagger.json",
		"../../../docs/swagger.json",
		"docs/swagger.json",
	}
	var data []byte
	var err error
	for _, cand := range candidates {
		data, err = os.ReadFile(cand)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("could not read swagger.json: %v", err)
	}

	var swaggerDoc struct {
		BasePath string                            `json:"basePath"`
		Paths    map[string]map[string]interface{} `json:"paths"`
	}
	if err := json.Unmarshal(data, &swaggerDoc); err != nil {
		t.Fatalf("could not parse swagger.json: %v", err)
	}

	reParam := regexp.MustCompile(`\{([^}]+)\}`)
	swaggerRoutes := make(map[string]bool)
	httpMethods := map[string]bool{
		"get": true, "post": true, "put": true, "patch": true, "delete": true, "head": true, "options": true,
	}
	var businessOpsCount int

	for path, methods := range swaggerDoc.Paths {
		ginPath := reParam.ReplaceAllString(path, ":$1")
		fullPath := swaggerDoc.BasePath + ginPath
		for method := range methods {
			if !httpMethods[strings.ToLower(method)] {
				continue
			}
			key := fmt.Sprintf("%s %s", strings.ToUpper(method), fullPath)
			swaggerRoutes[key] = true
			businessOpsCount++
		}
	}

	ginRoutes := make(map[string]bool)
	for _, r := range engine.Routes() {
		ginRoutes[fmt.Sprintf("%s %s", r.Method, r.Path)] = true
	}
	// Reverse direction: every documented operation must exist as a real route.
	for key := range swaggerRoutes {
		if !ginRoutes[key] {
			t.Errorf("swagger.json documents a route that does not exist in Gin: %s", key)
		}
	}

	var matchedRoutes int
	for _, r := range engine.Routes() {
		if !strings.HasPrefix(r.Path, "/areas-verdes/v1") {
			continue
		}
		if strings.HasPrefix(r.Path, "/areas-verdes/v1/swagger") {
			continue // Swagger UI documentation endpoint
		}
		key := fmt.Sprintf("%s %s", r.Method, r.Path)
		if !swaggerRoutes[key] {
			t.Errorf("Gin route missing from swagger.json: %s", key)
		} else {
			matchedRoutes++
		}
	}

	if matchedRoutes == 0 {
		t.Fatalf("no /areas-verdes/v1 routes were matched")
	}
	if businessOpsCount != matchedRoutes {
		t.Fatalf("swagger documenta %d operaciones y el router registra %d rutas /areas-verdes/v1 (sin la UI de swagger)", businessOpsCount, matchedRoutes)
	}

	t.Logf("Swagger coverage test passed: %d routes checked, %d business operations in swagger.json", matchedRoutes, businessOpsCount)
}
