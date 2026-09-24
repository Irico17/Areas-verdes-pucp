package etl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Options controla el ETL raw → v1 → PostGIS.
type Options struct {
	RawDir   string
	V1Dir    string
	DB       *gorm.DB
	SkipLoad bool
	Strict   bool
}

// Report son los conteos exactos escritos y, si aplica, cargados.
type Report struct {
	Areas    int
	Zonas    int
	Capas    map[string]int
	Manifest Manifest
}

type optionalCapa struct {
	file   string
	capa   string
	prefix string
	expect int
}

// Run normaliza el catastro y, salvo SkipLoad, lo carga en PostGIS.
func Run(opt Options) (Report, error) {
	areas, areasSum, err := readNormalize(opt.RawDir, "areas_verdes.geojson", NormalizeAreas)
	if err != nil {
		return Report{}, err
	}
	zonas, zonasSum, err := readNormalize(opt.RawDir, "jefe_de_grupo.json", NormalizeZonas)
	if err != nil {
		return Report{}, err
	}

	if opt.Strict {
		if len(areas) != ExpectedAreas {
			return Report{}, fmt.Errorf("áreas: %d, se esperaban %d", len(areas), ExpectedAreas)
		}
		if len(zonas) != ExpectedZonas {
			return Report{}, fmt.Errorf("zonas: %d, se esperaban %d", len(zonas), ExpectedZonas)
		}
	}
	if err := assertNoPII(zonas); err != nil {
		return Report{}, err
	}

	optionals := []optionalCapa{
		{"jardines_reserva.geojson", "jardines_reserva", "JR", ExpectedJardines},
		{"xerofitica.geojson", "xerofitica", "XE", ExpectedXerofitica},
	}
	capas := map[string][]Record{}
	capaSums := map[string]string{}
	for _, spec := range optionals {
		path := filepath.Join(opt.RawDir, spec.file)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("capa opcional ausente, se omite: %s\n", spec.file)
				continue
			}
			return Report{}, err
		}
		rows, sum, err := readNormalize(opt.RawDir, spec.file, func(b []byte) ([]Record, error) {
			return NormalizeCapa(b, spec.capa, spec.prefix)
		})
		if err != nil {
			return Report{}, err
		}
		if opt.Strict && len(rows) != spec.expect {
			return Report{}, fmt.Errorf("%s: %d, se esperaban %d", spec.capa, len(rows), spec.expect)
		}
		capas[spec.capa] = rows
		capaSums[spec.file] = sum
	}

	if err := WriteGeoJSON(opt.V1Dir, "areas_verdes", areas); err != nil {
		return Report{}, err
	}
	if err := WriteGeoJSON(opt.V1Dir, "zonas", zonas); err != nil {
		return Report{}, err
	}
	for name, rows := range capas {
		if err := WriteGeoJSON(opt.V1Dir, name, rows); err != nil {
			return Report{}, err
		}
	}

	manifest := Manifest{
		CRS:        "EPSG:4326",
		GeneradoEn: time.Now().UTC().Format(time.RFC3339),
		NotaPII:    "zonas derivan de jefe_de_grupo.json. Se omitió el campo jefes (nombres de personas). feature_id Z-NNNN no identifica a una persona.",
		Datasets: map[string]Dataset{
			"areas_verdes": {
				Archivo:      "areas_verdes.geojson",
				Fuente:       "areas_verdes.geojson",
				Features:     len(areas),
				SHA256Fuente: areasSum,
			},
			"zonas": {
				Archivo:      "zonas.geojson",
				Fuente:       "jefe_de_grupo.json",
				Features:     len(zonas),
				SHA256Fuente: zonasSum,
			},
		},
	}
	for name, rows := range capas {
		src := name + ".geojson"
		manifest.Datasets[name] = Dataset{
			Archivo:      src,
			Fuente:       src,
			Features:     len(rows),
			SHA256Fuente: capaSums[src],
		}
	}
	if err := WriteManifest(opt.V1Dir, manifest); err != nil {
		return Report{}, err
	}

	if !opt.SkipLoad {
		if opt.DB == nil {
			return Report{}, fmt.Errorf("no hay conexión a Postgres; usa --skip-load o define DATABASE_URL")
		}
		if err := Load(opt.DB, areas, zonas, capas); err != nil {
			return Report{}, err
		}
	}

	counts := map[string]int{}
	for name, rows := range capas {
		counts[name] = len(rows)
	}
	return Report{Areas: len(areas), Zonas: len(zonas), Capas: counts, Manifest: manifest}, nil
}

func readNormalize(dir, name string, fn func([]byte) ([]Record, error)) ([]Record, string, error) {
	path := filepath.Join(dir, name)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("leer %s: %w", path, err)
	}
	sum := sha256.Sum256(body)
	rows, err := fn(body)
	if err != nil {
		return nil, "", err
	}
	return rows, hex.EncodeToString(sum[:]), nil
}

func assertNoPII(zonas []Record) error {
	for _, z := range zonas {
		blob := strings.Join([]string{
			z.FeatureID, str(z.Codigo), str(z.Nombre), str(z.Uso), str(z.ProyRiego),
			str(z.RiegoAct), str(z.Referencia), string(z.Geometry),
		}, "\n")
		for _, frag := range PIIFragments {
			if strings.Contains(blob, frag) {
				return fmt.Errorf("PII o campo jefes presente en zona %s (%q)", z.FeatureID, frag)
			}
		}
	}
	return nil
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
