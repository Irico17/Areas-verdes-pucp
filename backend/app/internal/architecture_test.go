package internal_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const (
	baseModule = "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal"
	gormModule = "gorm.io/gorm"
)

func TestArchitectureLayers(t *testing.T) {
	internalDir := "."

	fset := token.NewFileSet()

	var (
		violations  []string
		parsedCount int
	)

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if filepath.Base(path) == "architecture_test.go" {
			return nil
		}

		parsedCount++

		rel, err := filepath.Rel(internalDir, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 2 {
			return nil
		}
		layer := parts[0]

		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("error parsing %s: %v", path, err)
			return nil
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			switch layer {
			case "domain":
				// domain imports application/persistence/infrastructure/presentation
				if isSubpackage(importPath, baseModule+"/application") ||
					isSubpackage(importPath, baseModule+"/persistence") ||
					isSubpackage(importPath, baseModule+"/infrastructure") ||
					isSubpackage(importPath, baseModule+"/presentation") {
					violations = append(violations, path+": layer domain imports "+importPath)
				}
			case "application":
				// application imports persistence/infrastructure/presentation or gorm
				if isSubpackage(importPath, baseModule+"/persistence") ||
					isSubpackage(importPath, baseModule+"/infrastructure") ||
					isSubpackage(importPath, baseModule+"/presentation") ||
					isSubpackage(importPath, gormModule) {
					violations = append(violations, path+": layer application imports "+importPath)
				}
			case "persistence":
				// persistence imports application/dto or presentation
				if isSubpackage(importPath, baseModule+"/application/dto") ||
					isSubpackage(importPath, baseModule+"/presentation") {
					violations = append(violations, path+": layer persistence imports "+importPath)
				}
			case "presentation":
				// presentation imports persistence or gorm directly
				if isSubpackage(importPath, baseModule+"/persistence") ||
					isSubpackage(importPath, gormModule) {
					violations = append(violations, path+": layer presentation imports "+importPath)
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("error walking internal directory: %v", err)
	}

	t.Logf("Parsed and checked %d Go files for architectural layer boundaries", parsedCount)

	if len(violations) > 0 {
		for _, v := range violations {
			t.Errorf("Architecture violation: %s", v)
		}
	}
}

func isSubpackage(importPath, target string) bool {
	return importPath == target || strings.HasPrefix(importPath, target+"/")
}
