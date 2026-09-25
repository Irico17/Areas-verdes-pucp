package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// UnirContrato lee openapi.yaml y mezcla los paths de openapi/<tag>.yaml.
// Cada frente añade el suyo en su archivo; este directorio se descubre por nombre.
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
