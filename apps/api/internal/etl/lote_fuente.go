package etl

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Fuentes del lote. La copia viva se guarda en data/raw/lote.
const (
	urlZonas    = "https://mapa-web-6.vercel.app/supervisoress.geojson"
	urlJefes    = "https://mapa-web-6.vercel.app/jefe_de_grupo.json"
	urlLugares  = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTakay8F0nlgM_t333fY-rOIG0auZ_vMh4-q0_d0_FJI9kS-bVv03Q0MxxgeX7XV9tz7jjPYJdMGI73/pub?gid=216669177&single=true&output=csv"
	urlGviz     = "https://docs.google.com/spreadsheets/d/1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI/gviz/tq?tqx=out:json"
	urlPalmeras = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTp1v1vtF9EI7zefB-PyT_vHaDfIPckJp5h9cskot0TmmZuEHFF1n6kpI9mxMju1Ay2hlaRqqZy3JHr/pub?gid=730733478&single=true&output=csv"
	urlCafetos  = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTp1v1vtF9EI7zefB-PyT_vHaDfIPckJp5h9cskot0TmmZuEHFF1n6kpI9mxMju1Ay2hlaRqqZy3JHr/pub?gid=746966399&single=true&output=csv"
	urlTipos    = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTakay8F0nlgM_t333fY-rOIG0auZ_vMh4-q0_d0_FJI9kS-bVv03Q0MxxgeX7XV9tz7jjPYJdMGI73/pub?gid=344271829&single=true&output=csv"
)

type fuentesLote struct {
	Zonas    []byte
	Jefes    []byte
	Lugares  []byte
	Gviz     []byte
	Palmeras []byte
	Cafetos  []byte
	Tipos    []byte
	Origen   map[string]string
}

// ResolverFuentes lee cada fuente en vivo una vez y la deja en data/raw/lote.
// Si el vivo falla, usa esa copia y, en último caso, el baseline de data/raw.
func ResolverFuentes(rawDir string, client *http.Client) (fuentesLote, error) {
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	cache := filepath.Join(rawDir, "lote")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return fuentesLote{}, err
	}
	out := fuentesLote{Origen: map[string]string{}}
	var err error
	out.Zonas, err = traer(client, urlZonas, filepath.Join(cache, "supervisoress.geojson"), filepath.Join(rawDir, "supervisoress.geojson"), out.Origen, "zonas")
	if err != nil {
		return fuentesLote{}, err
	}
	jefes, desde, err := traerSinGuardar(client, urlJefes, filepath.Join(cache, "jefe_de_grupo.json"), filepath.Join(rawDir, "jefe_de_grupo.json"))
	if err != nil {
		return fuentesLote{}, fmt.Errorf("poligonos: %w", err)
	}
	out.Origen["poligonos"] = desde
	out.Jefes, err = anonimizarJefes(jefes)
	if err != nil {
		return fuentesLote{}, err
	}
	if desde == "vivo" {
		if err := os.WriteFile(filepath.Join(cache, "jefe_de_grupo.json"), out.Jefes, 0o644); err != nil {
			return fuentesLote{}, err
		}
	}
	out.Lugares, err = traer(client, urlLugares, filepath.Join(cache, "lugares.csv"), filepath.Join(rawDir, "sheets", "lugares.csv"), out.Origen, "lugares")
	if err != nil {
		return fuentesLote{}, err
	}
	out.Gviz, err = traer(client, urlGviz, filepath.Join(cache, "flora_gviz.json"), filepath.Join(rawDir, "gviz", "flora_gviz_default.json"), out.Origen, "gviz")
	if err != nil {
		return fuentesLote{}, err
	}
	out.Palmeras, err = traer(client, urlPalmeras, filepath.Join(cache, "flora.csv"), filepath.Join(rawDir, "sheets", "flora.csv"), out.Origen, "palmeras")
	if err != nil {
		return fuentesLote{}, err
	}
	out.Cafetos, err = traer(client, urlCafetos, filepath.Join(cache, "cafetos.csv"), filepath.Join(rawDir, "sheets", "cafetos.csv"), out.Origen, "cafetos")
	if err != nil {
		return fuentesLote{}, err
	}
	tipos, err := traer(client, urlTipos, filepath.Join(cache, "actividades.csv"), filepath.Join(rawDir, "sheets", "actividades.csv"), out.Origen, "tipos")
	if err != nil {
		return fuentesLote{}, err
	}
	out.Tipos, err = recortarCatalogo(tipos)
	if err != nil {
		return fuentesLote{}, err
	}
	if out.Origen["tipos"] == "vivo" {
		if err := os.WriteFile(filepath.Join(cache, "actividades.csv"), out.Tipos, 0o644); err != nil {
			return fuentesLote{}, err
		}
	}
	return out, nil
}

// recortarCatalogo deja clase, tipo y descripción. El bloque derecho trae roles y no se guarda.
func recortarCatalogo(body []byte) ([]byte, error) {
	rows, err := readCSVBytes(body)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, rec := range rows {
		cols := []string{"", "", ""}
		for i := 0; i < 3 && i < len(rec); i++ {
			cols[i] = rec[i]
		}
		if err := w.Write(cols); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func traer(client *http.Client, url, cachePath, fallback string, origen map[string]string, clave string) ([]byte, error) {
	body, err := descargar(client, url)
	if err == nil && len(bytes.TrimSpace(body)) > 0 {
		if err := os.WriteFile(cachePath, body, 0o644); err != nil {
			return nil, err
		}
		origen[clave] = "vivo"
		return body, nil
	}
	if b, err2 := os.ReadFile(cachePath); err2 == nil && len(bytes.TrimSpace(b)) > 0 {
		origen[clave] = "cache"
		return b, nil
	}
	if b, err2 := os.ReadFile(fallback); err2 == nil && len(bytes.TrimSpace(b)) > 0 {
		origen[clave] = "baseline"
		return b, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", clave, err)
	}
	return nil, fmt.Errorf("%s: fuente vacía y sin copia local", clave)
}

func traerSinGuardar(client *http.Client, url, cachePath, fallback string) ([]byte, string, error) {
	body, err := descargar(client, url)
	if err == nil && len(bytes.TrimSpace(body)) > 0 {
		return body, "vivo", nil
	}
	if b, err2 := os.ReadFile(cachePath); err2 == nil && len(bytes.TrimSpace(b)) > 0 {
		return b, "cache", nil
	}
	if b, err2 := os.ReadFile(fallback); err2 == nil && len(bytes.TrimSpace(b)) > 0 {
		return b, "baseline", nil
	}
	if err != nil {
		return nil, "", err
	}
	return nil, "", fmt.Errorf("fuente vacía y sin copia local")
}

func descargar(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 40<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return body, nil
}

// anonimizarJefes sustituye el nombre de origen por su hash antes de persistir la copia.
// Las etiquetas de sector se conservan. El nombre real no se escribe.
func anonimizarJefes(body []byte) ([]byte, error) {
	var fc rawFC
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&fc); err != nil {
		return nil, fmt.Errorf("jefe_de_grupo: %w", err)
	}
	for i := range fc.Features {
		props := fc.Features[i].Properties
		if props == nil {
			continue
		}
		raw, _ := props["jefes"].(string)
		if raw == "" {
			continue
		}
		clave, etiqueta, visible := clavePersona(raw)
		if etiqueta {
			props["jefes"] = visible
			continue
		}
		if clave == "" {
			delete(props, "jefes")
			continue
		}
		props["jefes"] = clave
	}
	out, err := json.Marshal(fc)
	if err != nil {
		return nil, err
	}
	return out, nil
}
