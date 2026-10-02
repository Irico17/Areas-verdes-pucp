package etl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
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

type cargaLista struct {
	Vista     VistaPrevia
	persistir func(tx *gorm.DB, loteID int64) error
}

// Previsualizar llama al lector de la entidad y no abre una transacción de escritura.
func Previsualizar(entidad, nombre string, body []byte) (VistaPrevia, error) {
	carga, err := leerEntidad(entidad, nombre, body, nil)
	if err != nil {
		return VistaPrevia{}, err
	}
	carga.Vista.Escrito = false
	return carga.Vista, nil
}

// Confirmar es el segundo paso: escribe las filas válidas y deja el lote listo para revertir.
func Confirmar(db *gorm.DB, loteID, usuarioID int64) (int64, int, error) {
	if db == nil || loteID < 1 || usuarioID < 1 {
		return 0, 0, ErrLote
	}
	var entidad, estado, nombre string
	var contenido []byte
	err := db.Raw(`
		SELECT entidad, estado, COALESCE(nombre_archivo, ''), contenido
		FROM lotes_importacion WHERE id = $1`, loteID).Row().Scan(&entidad, &estado, &nombre, &contenido)
	if err != nil {
		return 0, 0, err
	}
	if estado != "vista_previa" {
		return 0, 0, ErrLote
	}
	var lugares map[string]string
	_ = db.Transaction(func(tx *gorm.DB) error {
		lugares, _ = mapaLugares(tx)
		return nil
	})
	carga, err := leerEntidad(entidad, nombre, contenido, lugares)
	if err != nil {
		return 0, 0, err
	}
	if carga.Vista.Validas == 0 {
		return loteID, 0, ErrSinValidas
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		var estadoTx string
		if err := tx.Raw(`SELECT estado FROM lotes_importacion WHERE id = $1 FOR UPDATE`, loteID).Row().Scan(&estadoTx); err != nil {
			return err
		}
		if estadoTx != "vista_previa" {
			return ErrLote
		}
		if err := carga.persistir(tx, loteID); err != nil {
			return err
		}
		if err := tx.Exec(`
			UPDATE cambios SET usuario_id = $2
			WHERE lote_id = $1 AND usuario_id IS NULL`, loteID, usuarioID).Error; err != nil {
			return err
		}
		return tx.Exec(`
			UPDATE lotes_importacion
			SET estado = 'confirmado', filas = $2
			WHERE id = $1`, loteID, carga.Vista.Validas).Error
	})
	if err != nil {
		return loteID, 0, err
	}
	return loteID, carga.Vista.Validas, nil
}

func entidadConocida(entidad string) bool {
	for _, item := range EntidadesImportables {
		if item == entidad {
			return true
		}
	}
	return false
}

func leerEntidad(entidad, nombre string, body []byte, lugares map[string]string) (cargaLista, error) {
	if !entidadConocida(entidad) {
		return cargaLista{}, fmt.Errorf("%w: %s", ErrEntidad, entidad)
	}
	arch, err := prepararArchivo(entidad, nombre, body)
	if err != nil {
		return cargaLista{}, err
	}
	vista := VistaPrevia{Entidad: entidad, Formato: arch.Formato, Errores: []ErrorFila{}, Filas: []map[string]any{}}
	var persistir func(tx *gorm.DB, loteID int64) error
	switch entidad {
	case "areas_verdes":
		rows, err := NormalizeAreas(arch.Geo)
		if err != nil {
			return cargaLista{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			for _, r := range rows {
				if err := tx.Exec(`
					INSERT INTO areas_verdes (
					  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
					  perimetro_m, area_m2, geom, origen_ref, activo
					) VALUES (
					  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, catastro_geom_4326($11), $1, TRUE
					)
					ON CONFLICT (feature_id) DO UPDATE SET
					  codigo = EXCLUDED.codigo, nombre = EXCLUDED.nombre, uso = EXCLUDED.uso,
					  proy_riego = EXCLUDED.proy_riego, riego_act = EXCLUDED.riego_act,
					  referencia = EXCLUDED.referencia, perimetro_m = EXCLUDED.perimetro_m,
					  area_m2 = EXCLUDED.area_m2, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref,
					  updated_at = now()`,
					r.FeatureID, r.SourceIndex, r.Codigo, r.Nombre, r.Uso, r.ProyRiego, r.RiegoAct, r.Referencia,
					r.PerimetroM, r.AreaM2, geomArg(r),
				).Error; err != nil {
					return err
				}
				despues, _ := json.Marshal(map[string]any{"feature_id": r.FeatureID, "nombre": r.Nombre})
				if err := auditar(tx, "areas_verdes", r.FeatureID, nil, despues, loteID); err != nil {
					return err
				}
			}
			return nil
		}
	case "zonas_supervision":
		rows, rech, err := leerZonasSupervision(arch.Geo)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarZonas(tx, rows, loteID, rep)
		}
	case "poligonos_cuadrilla":
		rows, rech, err := leerPoligonos(arch.Geo)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarCuadrillasYPoligonos(tx, rows, loteID, rep)
		}
	case "cuadrillas":
		rows, rech, err := leerCuadrillasFicticias(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			for _, c := range rows {
				if err := tx.Exec(`
					INSERT INTO cuadrillas (id, nombre_ficticio, turno, activo)
					VALUES ($1, $2, 'manana', TRUE)
					ON CONFLICT (id) DO UPDATE SET nombre_ficticio = EXCLUDED.nombre_ficticio`,
					c.ID, c.Nombre).Error; err != nil {
					return err
				}
				despues, _ := json.Marshal(map[string]any{"id": c.ID, "nombre_ficticio": c.Nombre})
				if err := auditar(tx, "cuadrillas", c.ID, nil, despues, loteID); err != nil {
					return err
				}
			}
			return nil
		}
	case "lugares":
		rows, rech, err := leerLugares(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarLugares(tx, rows, loteID, rep)
		}
	case "ejemplares":
		gviz := arch.JSON
		if len(gviz) == 0 {
			var err error
			gviz, err = csvAGviz(arch.CSV)
			if err != nil {
				return cargaLista{}, err
			}
		}
		rows, rech, err := leerGvizFlora(gviz)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarEjemplares(tx, rows, loteID, rep)
		}
	case "palmeras":
		rows, rech, err := leerMedidasPalmera(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarMedidas(tx, rows, nil, loteID, rep)
		}
	case "cafetos":
		rows, _, rech, err := leerCafetos(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarCafetos(tx, rows, loteID, rep)
		}
	case "catalogo_actividades":
		rows, rech, err := leerActividades(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return cargarTipos(tx, rows, loteID, rep)
		}
	case "labores":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		defer limpiar()
		rows, informe, err := ImportarMonitoreo(path, NuevaTabla())
		if err != nil {
			return cargaLista{}, err
		}
		vista.Errores = append(vista.Errores, informe.Errores...)
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		for _, aviso := range informe.Avisos {
			vista.Avisos = append(vista.Avisos, fmt.Sprintf("fila %d %s: %s", aviso.Fila, aviso.Campo, aviso.Motivo))
		}
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return upsertLabores(tx, rows, loteID, rep)
		}
	case "poda":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		defer limpiar()
		rows, errores, err := ImportarPoda(path, NuevaTabla(), lugares)
		if err != nil {
			return cargaLista{}, err
		}
		vista.Errores = errores
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return upsertPodas(tx, rows, loteID, rep)
		}
	case "vivero":
		path, limpiar, err := tempCSV(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		defer limpiar()
		rows, informe, err := ImportarVivero(path, NuevaTabla(), lugares)
		if err != nil {
			return cargaLista{}, err
		}
		vista.Errores = informe.Errores
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		if informe.ViveroDelta != "" {
			vista.Avisos = append(vista.Avisos, informe.ViveroDelta)
		}
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteLote{Cargados: map[string]int{}}
			return upsertVivero(tx, rows, loteID, rep)
		}
	case "tachos":
		rows, rech, _, err := LeerTachos(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error { return guardarTachos(tx, rows, loteID) }
	case "bebederos":
		dir, limpiar, err := tempBebederos(nombre, arch.Geo)
		if err != nil {
			return cargaLista{}, err
		}
		defer limpiar()
		rows, rech, err := LeerBebederos(dir)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error { return guardarBebederos(tx, rows, loteID) }
	case "fauna", "puertas", "playas", "vereda":
		spec := capaDe(entidad)
		dir, limpiar, err := tempArchivo(spec.file, arch.Geo)
		if err != nil {
			return cargaLista{}, err
		}
		defer limpiar()
		rows, err := leerCapaPuntos(dir, spec.file, spec.prefix, spec.nameKey)
		if err != nil {
			return cargaLista{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		tabla := spec.tabla
		persistir = func(tx *gorm.DB, loteID int64) error {
			for _, row := range rows {
				if err := upsertCapaPunto(tx, tabla, row, loteID); err != nil {
					return err
				}
			}
			return nil
		}
	case "xerofitica":
		rows, err := NormalizeCapa(arch.Geo, "xerofitica", "XE")
		if err != nil {
			return cargaLista{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			for _, r := range rows {
				if err := tx.Exec(`
					INSERT INTO xerofiticas (feature_id, clase, riego, area_m2, perimetro_m, geom, origen_ref)
					VALUES ($1,$2,$3,$4,$5, catastro_geom_4326($6), $1)
					ON CONFLICT (feature_id) DO UPDATE SET
					  clase = EXCLUDED.clase, riego = EXCLUDED.riego, area_m2 = EXCLUDED.area_m2,
					  perimetro_m = EXCLUDED.perimetro_m, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref,
					  updated_at = now()`,
					r.FeatureID, r.Clase, r.RiegoAct, r.AreaM2, r.PerimetroM, geomArg(r),
				).Error; err != nil {
					return err
				}
				despues, _ := json.Marshal(map[string]any{"feature_id": r.FeatureID})
				if err := auditar(tx, "xerofiticas", r.FeatureID, nil, despues, loteID); err != nil {
					return err
				}
			}
			return nil
		}
	case "jardines_reserva":
		rows, err := NormalizeCapa(arch.Geo, "jardines_reserva", "JR")
		if err != nil {
			return cargaLista{}, err
		}
		vista.Validas = len(rows)
		vista.Filas = primeras(rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			for _, r := range rows {
				if err := tx.Exec(`
					INSERT INTO jardines_reserva (
					  feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia, pertenecen,
					  perimetro_m, area_m2, geom, origen_ref
					) VALUES (
					  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10, catastro_geom_4326($11), $1
					)
					ON CONFLICT (feature_id) DO UPDATE SET
					  codigo = EXCLUDED.codigo, nombre = EXCLUDED.nombre, uso = EXCLUDED.uso,
					  proy_riego = EXCLUDED.proy_riego, riego_act = EXCLUDED.riego_act,
					  referencia = EXCLUDED.referencia, pertenecen = EXCLUDED.pertenecen,
					  perimetro_m = EXCLUDED.perimetro_m, area_m2 = EXCLUDED.area_m2,
					  geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`,
					r.FeatureID, r.Codigo, r.Nombre, r.Uso, r.ProyRiego, r.RiegoAct, r.Referencia, r.Pertenecen,
					r.PerimetroM, r.AreaM2, geomArg(r),
				).Error; err != nil {
					return err
				}
				despues, _ := json.Marshal(map[string]any{"feature_id": r.FeatureID})
				if err := auditar(tx, "jardines_reserva", r.FeatureID, nil, despues, loteID); err != nil {
					return err
				}
			}
			return nil
		}
	case "reservas":
		raw := body
		if arch.Formato != "json" {
			var err error
			raw, err = reservasDesdeTabular(arch.CSV)
			if err != nil {
				return cargaLista{}, err
			}
		}
		rows, rech, err := LeerReservasFicticias(raw)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		persistir = func(tx *gorm.DB, loteID int64) error {
			rep := &ReporteCapas{Cargados: map[string]int{}}
			return guardarReservas(tx, rows, loteID, rep)
		}
	case "puntos_pucp":
		rows, rech, omitidas, err := LeerPuntosPUCP(arch.CSV)
		if err != nil {
			return cargaLista{}, err
		}
		vista = llenar(vista, len(rows), rech, rows)
		vista.ColumnasOmitidas = omitidas
		if len(omitidas) > 0 {
			vista.AvisoOmitidas = "columnas omitidas por datos personales"
		}
		persistir = func(tx *gorm.DB, loteID int64) error { return guardarPuntos(tx, rows, loteID) }
	default:
		return cargaLista{}, fmt.Errorf("%w: %s", ErrEntidad, entidad)
	}
	if vista.Errores == nil {
		vista.Errores = []ErrorFila{}
	}
	if vista.Filas == nil {
		vista.Filas = []map[string]any{}
	}
	return cargaLista{Vista: vista, persistir: persistir}, nil
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
