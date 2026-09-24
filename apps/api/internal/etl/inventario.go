package etl

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// InvRecord es un elemento de inventario ya sin HTML ni enlaces externos.
type InvRecord struct {
	Capa      string
	FeatureID string
	Nombre    string
	Subtipo   string
	Detalle   string
	Lugar     string
	Foto      string
	Geometry  json.RawMessage
}

type invFeature struct {
	Properties map[string]any  `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type invCollection struct {
	Features []invFeature `json:"features"`
}

var bebederoFiles = []struct{ file, subtipo string }{
	{"bebedero Nuevo.geojson", "nuevo"},
	{"bebedero Tipo fuente.geojson", "fuente"},
	{"bebedero Tipo llenador de botella.geojson", "llenador"},
	{"bebedero por deterioro.geojson", "deterioro"},
	{"bebedero por baja del equipo.geojson", "baja"},
}

// CollectInventario lee data/raw y omite lo que no esté. No guarda HTML ni URLs.
func CollectInventario(rawDir string) (map[string][]InvRecord, error) {
	out := map[string][]InvRecord{}
	fotos := fotoIndex(rawDir)

	if rows, err := readBebederos(rawDir, fotos); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["bebederos"] = rows
	}
	if rows, err := readNamedPoints(rawDir, "fauna.geojson", "fauna", "FA", "Animal", "Avistamiento"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["fauna"] = rows
	}
	if rows, err := readNamedPoints(rawDir, "puertas_entradas.geojson", "puertas", "PU", "", "Acceso"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["puertas"] = rows
	}
	if rows, err := readPolygons(rawDir, "playas_de_estacionamiento.geojson", "playas_estacionamiento", "PL", "Playa"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["playas_estacionamiento"] = rows
	}
	if rows, err := readPolygons(rawDir, "area_vereda_peligro.geojson", "area_vereda_peligro", "VD", "Vereda en riesgo"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["area_vereda_peligro"] = rows
	}
	if rows, err := readSheetPoints(rawDir, "flora.csv", "flora", "FL"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["flora"] = rows
	}
	if rows, err := readSheetPoints(rawDir, "cafetos.csv", "cafetos", "CF"); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["cafetos"] = rows
	}
	if rows, err := readTachos(rawDir); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		out["tachos"] = rows
	}
	return out, nil
}

// WriteInventario deja una copia normalizada en data/v1/inventario.
func WriteInventario(dir string, capas map[string][]InvRecord) error {
	sub := filepath.Join(dir, "inventario")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		return err
	}
	for capa, rows := range capas {
		features := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			features = append(features, map[string]any{
				"type":     "Feature",
				"id":       r.FeatureID,
				"geometry": json.RawMessage(r.Geometry),
				"properties": map[string]any{
					"feature_id": r.FeatureID,
					"capa":       capa,
					"nombre":     r.Nombre,
					"subtipo":    r.Subtipo,
					"detalle":    r.Detalle,
					"lugar":      r.Lugar,
					"foto":       r.Foto,
				},
			})
		}
		body, err := json.Marshal(map[string]any{
			"type":     "FeatureCollection",
			"name":     capa,
			"crs_nota": "EPSG:4326",
			"features": features,
		})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sub, capa+".geojson"), body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// LoadInventario reemplaza solo la tabla inventario.
func LoadInventario(db *gorm.DB, capas map[string][]InvRecord) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`TRUNCATE inventario RESTART IDENTITY`).Error; err != nil {
			return fmt.Errorf("truncar inventario: %w", err)
		}
		for capa, rows := range capas {
			for _, r := range rows {
				if len(r.Geometry) == 0 {
					continue
				}
				if err := tx.Exec(`
					INSERT INTO inventario (capa, feature_id, nombre, subtipo, detalle, lugar, foto, geom)
					VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), inventario_geom_4326($8))`,
					capa, r.FeatureID, r.Nombre, r.Subtipo, r.Detalle, r.Lugar, r.Foto, string(r.Geometry),
				).Error; err != nil {
					return fmt.Errorf("inventario %s %s: %w", capa, r.FeatureID, err)
				}
			}
		}
		return nil
	})
}

func readBebederos(rawDir string, fotos map[string]string) ([]InvRecord, error) {
	var rows []InvRecord
	seen := map[string]int{}
	any := false
	for _, spec := range bebederoFiles {
		path := filepath.Join(rawDir, spec.file)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("inventario ausente, se omite: %s\n", spec.file)
				continue
			}
			return nil, err
		}
		any = true
		fc, err := readFC(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", spec.file, err)
		}
		for i, f := range fc.Features {
			nombre := firstString(f.Properties, "Name", "name")
			if nombre == "" {
				nombre = fmt.Sprintf("Bebedero %s %d", spec.subtipo, i+1)
			}
			id := uniqueID(seen, "BB-"+sanitizeID(nombre))
			foto := matchFoto(fotos, nombre)
			rows = append(rows, InvRecord{
				Capa:      "bebederos",
				FeatureID: id,
				Nombre:    nombre,
				Subtipo:   spec.subtipo,
				Detalle:   "Tipo " + spec.subtipo,
				Lugar:     plainPlace(f.Properties, nombre),
				Foto:      foto,
				Geometry:  f.Geometry,
			})
		}
	}
	if !any {
		return nil, nil
	}
	return rows, nil
}

func readNamedPoints(rawDir, file, capa, prefix, nameKey, fallback string) ([]InvRecord, error) {
	path := filepath.Join(rawDir, file)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("inventario ausente, se omite: %s\n", file)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	fc, err := readFC(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	rows := make([]InvRecord, 0, len(fc.Features))
	for i, f := range fc.Features {
		nombre := firstString(f.Properties, nameKey, "Name", "name")
		if nombre == "" {
			nombre = fmt.Sprintf("%s %d", fallback, i+1)
		}
		rows = append(rows, InvRecord{
			Capa:      capa,
			FeatureID: fmt.Sprintf("%s-%04d", prefix, i+1),
			Nombre:    nombre,
			Detalle:   fallback,
			Geometry:  f.Geometry,
		})
	}
	return rows, nil
}

func readPolygons(rawDir, file, capa, prefix, fallback string) ([]InvRecord, error) {
	return readNamedPoints(rawDir, file, capa, prefix, "nombre", fallback)
}

func readTachos(rawDir string) ([]InvRecord, error) {
	path := filepath.Join(rawDir, "sheets", "tachos.csv")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("inventario ausente, se omite: sheets/tachos.csv\n")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	records, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}
	header := indexHeader(records[0])
	seen := map[string]int{}
	var rows []InvRecord
	for _, rec := range records[1:] {
		lat, lon, ok := latLonFrom(rec)
		if !ok {
			continue
		}
		idRaw := cell(rec, header, "id")
		if idRaw == "" {
			continue
		}
		rows = append(rows, InvRecord{
			Capa:      "tachos",
			FeatureID: uniqueID(seen, "TA-"+sanitizeID(idRaw)),
			Nombre:    idRaw,
			Lugar:     cell(rec, header, "lugar"),
			Detalle:   cell(rec, header, "note"),
			Geometry:  pointGeom(lon, lat),
		})
	}
	return rows, nil
}

func readSheetPoints(rawDir, file, capa, prefix string) ([]InvRecord, error) {
	path := filepath.Join(rawDir, "sheets", file)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("inventario ausente, se omite: sheets/%s\n", file)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	records, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	var rows []InvRecord
	n := 0
	for _, rec := range records {
		lat, lon, ok := latLonFrom(rec)
		if !ok {
			continue
		}
		n++
		lugar := ""
		comun := ""
		cientifico := ""
		if len(rec) > 1 {
			lugar = strings.TrimSpace(rec[1])
		}
		if len(rec) > 6 {
			comun = strings.TrimSpace(rec[6])
		}
		if len(rec) > 7 {
			cientifico = strings.TrimSpace(rec[7])
		}
		nombre := comun
		if nombre == "" {
			nombre = cientifico
		}
		if nombre == "" {
			nombre = fmt.Sprintf("%s %d", capa, n)
		}
		detalle := cientifico
		if comun != "" && cientifico != "" {
			detalle = comun + " · " + cientifico
		}
		rows = append(rows, InvRecord{
			Capa:      capa,
			FeatureID: fmt.Sprintf("%s-%04d", prefix, n),
			Nombre:    nombre,
			Detalle:   detalle,
			Lugar:     lugar,
			Geometry:  pointGeom(lon, lat),
		})
	}
	return rows, nil
}

func readFC(path string) (invCollection, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return invCollection{}, err
	}
	var fc invCollection
	if err := json.Unmarshal(body, &fc); err != nil {
		return invCollection{}, err
	}
	return fc, nil
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	var rows [][]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		rows = append(rows, rec)
	}
	return rows, nil
}

func indexHeader(header []string) map[string]int {
	out := map[string]int{}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if key != "" {
			out[key] = i
		}
	}
	return out
}

func cell(rec []string, header map[string]int, name string) string {
	i, ok := header[name]
	if !ok || i >= len(rec) {
		return ""
	}
	v := strings.TrimSpace(rec[i])
	if strings.EqualFold(v, "null") {
		return ""
	}
	return v
}

func latLonFrom(rec []string) (lat, lon float64, ok bool) {
	for _, raw := range rec {
		v, good := commaFloat(raw)
		if !good {
			continue
		}
		if v <= -11.9 && v >= -12.2 {
			lat = v
		}
		if v <= -76.9 && v >= -77.3 {
			lon = v
		}
	}
	return lat, lon, lat != 0 && lon != 0
}

func commaFloat(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"`)
	if s == "" || strings.EqualFold(s, "null") || strings.Contains(s, "/") {
		return 0, false
	}
	if strings.Count(s, ",") == 1 && !strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func pointGeom(lon, lat float64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"type":"Point","coordinates":[%.7f,%.7f]}`, lon, lat))
}

func firstString(props map[string]any, keys ...string) string {
	for _, key := range keys {
		if props == nil {
			return ""
		}
		v, ok := props[key]
		if !ok || v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

func plainPlace(props map[string]any, nombre string) string {
	best := ""
	for key, value := range props {
		if value == nil || strings.EqualFold(key, "descriptio") || strings.EqualFold(key, "description") {
			continue
		}
		s, ok := value.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if len(s) < 12 || len(s) > 240 || strings.Contains(s, "<") || strings.Contains(s, "http") || strings.Contains(s, "drive.google") {
			continue
		}
		if strings.EqualFold(s, nombre) {
			continue
		}
		if strings.Count(s, " ") > strings.Count(best, " ") {
			best = s
		}
	}
	return best
}

func sanitizeID(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == '-' || r == '_' || r == ' ' {
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "x"
	}
	return out
}

func uniqueID(seen map[string]int, id string) string {
	seen[id]++
	if seen[id] == 1 {
		return id
	}
	return fmt.Sprintf("%s-%d", id, seen[id])
}

func fotoIndex(rawDir string) map[string]string {
	dir := filepath.Join(rawDir, "drive_fotos")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		lower := strings.ToLower(name)
		if e.IsDir() || (!strings.HasSuffix(lower, ".jpg") && !strings.HasSuffix(lower, ".jpeg")) {
			continue
		}
		out[strings.TrimSuffix(lower, filepath.Ext(lower))] = name
	}
	return out
}

func matchFoto(fotos map[string]string, nombre string) string {
	if len(fotos) == 0 {
		return ""
	}
	key := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(nombre), "PT_"))
	key = strings.TrimPrefix(key, "pt_")
	if name, ok := fotos[key]; ok {
		return name
	}
	for stem, name := range fotos {
		if stem != "" && strings.Contains(key, stem) {
			return name
		}
	}
	return ""
}
