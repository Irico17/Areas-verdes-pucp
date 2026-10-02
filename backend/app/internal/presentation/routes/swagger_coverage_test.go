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
// (except the Swagger UI) to appear in swagger.json, and the reverse.
// The expected operation count is the number of those mounted routes. A fixed floor
// such as ">= 62" stays green after endpoints are removed, as long as the leftover
// file still clears the old snapshot.
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
	httpVerb := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
	}
	swaggerRoutes := make(map[string]bool)

	for path, methods := range swaggerDoc.Paths {
		ginPath := reParam.ReplaceAllString(path, ":$1")
		fullPath := swaggerDoc.BasePath + ginPath
		for method := range methods {
			verb := strings.ToUpper(method)
			if !httpVerb[verb] {
				continue
			}
			key := fmt.Sprintf("%s %s", verb, fullPath)
			swaggerRoutes[key] = true
		}
	}

	ginRoutes := make(map[string]bool)
	businessRoutes := make(map[string]bool)
	for _, r := range engine.Routes() {
		key := fmt.Sprintf("%s %s", r.Method, r.Path)
		ginRoutes[key] = true
		if !strings.HasPrefix(r.Path, "/areas-verdes/v1") {
			continue
		}
		if strings.HasPrefix(r.Path, "/areas-verdes/v1/swagger") {
			continue // Swagger UI documentation endpoint
		}
		businessRoutes[key] = true
	}

	if len(businessRoutes) == 0 {
		t.Fatal("no /areas-verdes/v1 routes were matched")
	}

	// Reverse direction: every documented operation must exist as a real route.
	for key := range swaggerRoutes {
		if !ginRoutes[key] {
			t.Errorf("swagger.json documents a route that does not exist in Gin: %s", key)
		}
	}

	for key := range businessRoutes {
		if !swaggerRoutes[key] {
			t.Errorf("Gin route missing from swagger.json: %s", key)
		}
	}

	if len(swaggerRoutes) != len(businessRoutes) {
		t.Errorf("swagger.json documents %d operations; router mounts %d /areas-verdes/v1 routes outside /swagger", len(swaggerRoutes), len(businessRoutes))
	}

	t.Logf("Swagger coverage test passed: %d routes checked", len(businessRoutes))
}
