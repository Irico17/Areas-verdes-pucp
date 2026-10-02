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
// to appear in swagger.json, and validates that swagger.json has at least 62 business operations.
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
	var businessOpsCount int

	for path, methods := range swaggerDoc.Paths {
		ginPath := reParam.ReplaceAllString(path, ":$1")
		fullPath := swaggerDoc.BasePath + ginPath
		for method := range methods {
			key := fmt.Sprintf("%s %s", strings.ToUpper(method), fullPath)
			swaggerRoutes[key] = true
			businessOpsCount++
		}
	}

	if businessOpsCount < 62 {
		t.Fatalf("expected >= 62 business operations in swagger.json, got %d", businessOpsCount)
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

	t.Logf("Swagger coverage test passed: %d routes checked, %d business operations in swagger.json", matchedRoutes, businessOpsCount)
}
