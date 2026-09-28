// Package archivos provides file and document reading adapters.
package archivos

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type contratoOpenAPIAdapter struct {
	openAPIPath string
}

// NewContratoOpenAPIAdapter creates an adapter to load and merge OpenAPI YAML contracts.
func NewContratoOpenAPIAdapter(cfg *config.Config) contracts.IContratoOpenAPI {
	path := ""
	if cfg != nil {
		path = cfg.Datos.OpenAPIPath
	}
	return &contratoOpenAPIAdapter{openAPIPath: path}
}

// ObtenerContrato returns the merged OpenAPI contract bytes.
func (a *contratoOpenAPIAdapter) ObtenerContrato() ([]byte, error) {
	if a.openAPIPath == "" {
		return nil, os.ErrNotExist
	}
	if _, err := os.Stat(a.openAPIPath); err != nil {
		return nil, err
	}
	return UnirContrato(a.openAPIPath)
}

// UnirContrato reads openapi.yaml and merges paths from openapi/<tag>.yaml.
func UnirContrato(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("leer contrato: %w", err)
	}
	dir := filepath.Join(filepath.Dir(path), "openapi")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return body, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	paths, _ := root["paths"].(map[string]any)
	if paths == nil {
		paths = map[string]any{}
	}
	visto := map[string]string{}
	for key := range paths {
		if strings.HasPrefix(key, "/") {
			visto[key] = filepath.Base(path)
		}
	}
	for _, name := range names {
		partBody, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var part map[string]any
		if err := yaml.Unmarshal(partBody, &part); err != nil {
			return nil, fmt.Errorf("leer %s: %w", name, err)
		}
		for key, value := range part {
			if !strings.HasPrefix(key, "/") {
				continue
			}
			if prev, ok := visto[key]; ok {
				return nil, fmt.Errorf("path %s repetido en %s y %s", key, prev, name)
			}
			visto[key] = name
			paths[key] = value
		}
	}
	root["paths"] = paths
	out, err := yaml.Marshal(root)
	if err != nil {
		return nil, err
	}
	return out, nil
}
