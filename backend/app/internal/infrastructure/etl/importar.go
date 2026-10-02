package etl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EntidadesImportables son las de docs/MAPA-DATOS-Y-EDICION.md secciones 4.1 a 4.15.
var EntidadesImportables = []string{
	"areas_verdes",
	"zonas_supervision",
	"poligonos_cuadrilla",
	"cuadrillas",
	"lugares",
	"ejemplares",
	"palmeras",
	"cafetos",
	"catalogo_actividades",
	"labores",
	"poda",
	"vivero",
	"tachos",
	"bebederos",
	"fauna",
	"puertas",
	"playas",
	"vereda",
	"xerofitica",
	"jardines_reserva",
	"reservas",
	"puntos_pucp",
}

// ErrEntidad es una entidad fuera del mapa de datos.
var ErrEntidad = errors.New("entidad no importable")

// ErrSinValidas: no hay filas que escribir.
var ErrSinValidas = errors.New("ninguna fila válida")

// ErrLote: el lote no está en vista previa.
var ErrLote = errors.New("lote no confirmable")

// VistaPrevia es el primer paso. No escribe en las tablas de la entidad.
type VistaPrevia struct {
	Entidad          string           `json:"entidad"`
	Formato          string           `json:"formato"`
	Validas          int              `json:"validas"`
	Errores          []ErrorFila      `json:"errores"`
	Filas            []map[string]any `json:"filas"`
	ColumnasOmitidas []string         `json:"columnas_omitidas,omitempty"`
	AvisoOmitidas    string           `json:"aviso_omitidas,omitempty"`
	Avisos           []string         `json:"avisos,omitempty"`
	Escrito          bool             `json:"escrito"`
}

// Previsualizar llama al lector de la entidad y no abre una transacción de escritura.
func Previsualizar(entidad, nombre string, body []byte) (VistaPrevia, error) {
	vista, err := leerVistaPrevia(entidad, nombre, body, nil)
	if err != nil {
		return VistaPrevia{}, err
	}
	vista.Escrito = false
	return vista, nil
}

func entidadConocida(entidad string) bool {
	for _, item := range EntidadesImportables {
		if item == entidad {
			return true
		}
	}
	return false
}

func leerVistaPrevia(entidad, nombre string, body []byte, lugares map[string]string) (VistaPrevia, error) {
	if !entidadConocida(entidad) {
		return VistaPrevia{}, fmt.Errorf("%w: %s", ErrEntidad, entidad)
	}
	arch, err := prepararArchivo(entidad, nombre, body)
	if err != nil {
		return VistaPrevia{}, err
	}
	vista := VistaPrevia{Entidad: entidad, Formato: arch.Formato, Errores: []ErrorFila{}, Filas: []map[string]any{}}
	switch entidad {
	case "areas_verdes":
		rows, err := NormalizeAreas(arch.Geo)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
	case "zonas_supervision":
		rows, rech, err := leerZonasSupervision(arch.Geo)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "poligonos_cuadrilla":
		rows, rech, err := leerPoligonos(arch.Geo)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "cuadrillas":
		rows, rech, err := leerCuadrillasFicticias(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "lugares":
		rows, rech, err := leerLugares(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "ejemplares":
		gviz := arch.JSON
		if len(gviz) == 0 {
			var err error
			gviz, err = csvAGviz(arch.CSV)
			if err != nil {
				return VistaPrevia{}, err
			}
		}
		rows, rech, err := leerGvizFlora(gviz)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "palmeras":
		rows, rech, err := leerMedidasPalmera(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "cafetos":
		rows, _, rech, err := leerCafetos(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "catalogo_actividades":
		rows, rech, err := leerActividades(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "labores":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		defer limpiar()
		rows, informe, err := ImportarMonitoreo(path, NuevaTabla())
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Errores = append(vista.Errores, informe.Errores...)
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		for _, aviso := range informe.Avisos {
			vista.Avisos = append(vista.Avisos, fmt.Sprintf("fila %d %s: %s", aviso.Fila, aviso.Campo, aviso.Motivo))
		}
	case "poda":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		defer limpiar()
		rows, errores, err := ImportarPoda(path, NuevaTabla(), lugares)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Errores = errores
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
	case "vivero":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		defer limpiar()
		rows, informe, err := ImportarVivero(path, NuevaTabla(), lugares)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Errores = informe.Errores
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		if informe.ViveroDelta != "" {
			vista.Avisos = append(vista.Avisos, informe.ViveroDelta)
		}
	case "tachos":
		rows, rech, _, err := LeerTachos(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "bebederos":
		dir, limpiar, err := tempBebederos(nombre, arch.Geo)
		if err != nil {
			return VistaPrevia{}, err
		}
		defer limpiar()
		rows, rech, err := LeerBebederos(dir)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "fauna", "puertas", "playas", "vereda":
		spec := capaDe(entidad)
		dir, limpiar, err := tempArchivo(spec.file, arch.Geo)
		if err != nil {
			return VistaPrevia{}, err
		}
		defer limpiar()
		rows, err := leerCapaPuntos(dir, spec.file, spec.prefix, spec.nameKey)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
	case "xerofitica":
		rows, err := NormalizeCapa(arch.Geo, "xerofitica", "XE")
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
	case "jardines_reserva":
		rows, err := NormalizeCapa(arch.Geo, "jardines_reserva", "JR")
		if err != nil {
			return VistaPrevia{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
	case "reservas":
		raw := body
		if arch.Formato != "json" {
			var err error
			raw, err = reservasDesdeTabular(arch.CSV)
			if err != nil {
				return VistaPrevia{}, err
			}
		}
		rows, rech, err := LeerReservasFicticias(raw)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
	case "puntos_pucp":
		rows, rech, omitidas, err := LeerPuntosPUCP(arch.CSV)
		if err != nil {
			return VistaPrevia{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		vista.ColumnasOmitidas = omitidas
		if len(omitidas) > 0 {
			vista.AvisoOmitidas = "columnas omitidas por datos personales"
		}
	default:
		return VistaPrevia{}, fmt.Errorf("%w: %s", ErrEntidad, entidad)
	}
	if vista.Errores == nil {
		vista.Errores = []ErrorFila{}
	}
	if vista.Filas == nil {
		vista.Filas = []map[string]any{}
	}
	return vista, nil
}

func llenar(vista VistaPrevia, n int, rech []Rechazo, filas any) VistaPrevia {
	vista.Validas = n
	vista.Errores = make([]ErrorFila, 0, len(rech))
	for _, r := range rech {
		vista.Errores = append(vista.Errores, ErrorFila{Fila: r.Fila, Campo: r.Campo, Motivo: r.Motivo})
	}
	vista.Filas = primeras(filas)
	return vista
}

func primeras(v any) []map[string]any {
	raw, err := json.Marshal(v)
	if err != nil {
		return []map[string]any{}
	}
	var lista []map[string]any
	if err := json.Unmarshal(raw, &lista); err != nil {
		return []map[string]any{}
	}
	if len(lista) > 20 {
		lista = lista[:20]
	}
	if lista == nil {
		return []map[string]any{}
	}
	return lista
}

type capaSpec struct {
	file, prefix, nameKey, tabla string
}

func capaDe(entidad string) capaSpec {
	switch entidad {
	case "fauna":
		return capaSpec{"fauna.geojson", "FA", "Animal", "fauna"}
	case "puertas":
		return capaSpec{"puertas_entradas.geojson", "PU", "Name", "puertas"}
	case "playas":
		return capaSpec{"playas_de_estacionamiento.geojson", "PL", "Name", "playas_estacionamiento"}
	default:
		return capaSpec{"area_vereda_peligro.geojson", "VD", "Name", "veredas_riesgo"}
	}
}

func tempCSV(body []byte) (string, func(), error) {
	dir, err := os.MkdirTemp("", "importacion")
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(dir, "datos.csv")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}
	return path, func() { os.RemoveAll(dir) }, nil
}

func tempArchivo(nombre string, body []byte) (string, func(), error) {
	dir, err := os.MkdirTemp("", "importacion")
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, nombre), body, 0o644); err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

func tempBebederos(nombre string, body []byte) (string, func(), error) {
	elegido := ""
	base := filepath.Base(nombre)
	for _, spec := range bebederoFiles {
		if spec.file == base || strings.Contains(strings.ToLower(base), spec.subtipo) {
			elegido = spec.file
			break
		}
	}
	if elegido == "" {
		elegido = bebederoFiles[0].file
	}
	return tempArchivo(elegido, body)
}

type cuadrillaImport struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre_ficticio"`
}

// leerCuadrillasFicticias no persiste el nombre real: usa el anonimizador ya mergeado.
func leerCuadrillasFicticias(body []byte) ([]cuadrillaImport, []Rechazo, error) {
	filas, err := readCSVBytes(body)
	if err != nil {
		return nil, nil, err
	}
	if len(filas) < 2 {
		return nil, nil, fmt.Errorf("cuadrillas: csv vacío")
	}
	header := indexHeader(filas[0])
	var nombres []string
	for _, rec := range filas[1:] {
		if n := cell(rec, header, "nombre"); n != "" {
			nombres = append(nombres, n)
		}
	}
	tabla := NuevaTabla()
	tabla.Aplicar(nombres)
	var out []cuadrillaImport
	var rech []Rechazo
	vistos := map[string]bool{}
	for i, rec := range filas[1:] {
		nombre := cell(rec, header, "nombre")
		if nombre == "" {
			continue
		}
		ficticio := tabla.Ficticio(nombre)
		if ficticio == "" {
			rech = append(rech, Rechazo{Fuente: "cuadrillas", Fila: i + 2, Campo: "nombre", Motivo: "no se importa el nombre real"})
			continue
		}
		id := "cf-" + slug(ficticio)
		if vistos[id] {
			continue
		}
		vistos[id] = true
		out = append(out, cuadrillaImport{ID: id, Nombre: ficticio})
	}
	return out, rech, nil
}
