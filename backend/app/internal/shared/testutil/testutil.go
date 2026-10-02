// Package testutil provides helper utilities for tests across packages.
package testutil

import (
	"os"
	"path/filepath"
)

// FindRepoRoot locates the repository root containing apps/api/openapi.yaml.
func FindRepoRoot() string {
	wd, err := os.Getwd()
	if err == nil {
		for dir := wd; ; {
			candidate := filepath.Join(dir, "apps", "api", "openapi.yaml")
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "."
}

// FindMigrationsDir locates the migrations directory for tests without loading .env.
func FindMigrationsDir() string {
	if env := os.Getenv("MIGRATIONS_DIR"); env != "" {
		if fi, err := os.Stat(env); err == nil && fi.IsDir() {
			return env
		}
	}
	root := FindRepoRoot()
	candidate := filepath.Join(root, "db", "migrations")
	if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
		return candidate
	}
	return filepath.Join("..", "..", "..", "..", "..", "db", "migrations")
}
