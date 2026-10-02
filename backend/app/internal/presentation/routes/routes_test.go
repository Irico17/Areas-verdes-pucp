package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainEntities "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/ratelimit"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type mockSaludUC struct{}

func (mockSaludUC) VerificarSalud(_ context.Context) (*dto.EstadoSaludDTO, error) {
	return &dto.EstadoSaludDTO{Database: "up", PostGIS: "3.5"}, nil
}

type mockSesionRoutesUC struct{}

func (mockSesionRoutesUC) Login(_ context.Context, _, _ string) (string, *dto.UsuarioSesionDTO, error) {
	return "token", &dto.UsuarioSesionDTO{}, nil
}

func (mockSesionRoutesUC) Logout(_ context.Context, _ string) {
}

func (mockSesionRoutesUC) Resolver(_ context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	switch token {
	case "token-admin":
		return &dto.UsuarioSesionDTO{ID: 1, Usuario: "admin", Rol: enums.RolAdmin.String(), RolNombre: "Administrador"}, nil
	case "token-coordinacion":
		return &dto.UsuarioSesionDTO{ID: 2, Usuario: "coordinacion", Rol: enums.RolCoordinacion.String(), RolNombre: "Ingeniería/Coordinación"}, nil
	case "token-jefatura":
		return &dto.UsuarioSesionDTO{ID: 3, Usuario: "jefatura", Rol: enums.RolJefatura.String(), RolNombre: "Jefatura"}, nil
	case "token-norte":
		return &dto.UsuarioSesionDTO{ID: 4, Usuario: "norte", Rol: enums.RolCapataz.String(), RolNombre: "Capataz", CapatazID: "cap-norte"}, nil
	default:
		return nil, errors.New("sin sesion")
	}
}

type mockUsuarioRoutesUC struct{}

func (mockUsuarioRoutesUC) ListarUsuarios(_ context.Context) (*dto.UsuariosResponseDTO, error) {
	return &dto.UsuariosResponseDTO{
		Usuarios: []dto.CuentaDTO{
			{ID: 1, Usuario: "admin", Rol: enums.RolAdmin.String(), Activo: true},
		},
		Permisos: []dto.PermisoDTO{},
		Aviso:    "La jefatura de sección administra las cuentas.",
	}, nil
}

func (mockUsuarioRoutesUC) Crear(_ context.Context, _ int64, in dto.CrearCuentaDTO) (*dto.CuentaDTO, error) {
	return &dto.CuentaDTO{ID: 9, Usuario: in.Usuario, Nombre: in.Nombre, Rol: in.Rol, Activo: true, DebeCambiarPassword: true}, nil
}

func (mockUsuarioRoutesUC) Actualizar(_ context.Context, _ int64, _, usuario string, _ dto.ActualizarCuentaDTO) (*dto.CuentaDTO, error) {
	return &dto.CuentaDTO{ID: 9, Usuario: usuario, Activo: true}, nil
}

func (mockUsuarioRoutesUC) CambiarClavePropia(_ context.Context, usuario, _, _ string) (*dto.UsuarioSesionDTO, error) {
	return &dto.UsuarioSesionDTO{Usuario: usuario}, nil
}

func (mockUsuarioRoutesUC) ActualizarPermiso(_ context.Context, _ int64, _, _ string, _ bool) error {
	return nil
}

func (mockUsuarioRoutesUC) CrearRol(_ context.Context, _ int64, codigo, nombre string) (*dto.RolDTO, error) {
	return &dto.RolDTO{Codigo: codigo, Nombre: nombre, Activo: true}, nil
}

func (mockUsuarioRoutesUC) ActualizarRol(_ context.Context, _ int64, _ string, _ bool) error {
	return nil
}

type mockContratoOpenAPI struct{}

func (mockContratoOpenAPI) ObtenerContrato() ([]byte, error) {
	return []byte("openapi: 3.0.0"), nil
}

type mockCatalogoRoutesUC struct{}

func (mockCatalogoRoutesUC) Listar(_ context.Context, _ dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
	return &dto.CatalogoListResponseDTO{
		Items: []dto.CatalogoItemDTO{
			{ID: 1, Clase: "estado", Codigo: "pendiente", Nombre: "Pendiente", Activo: true, Orden: 1},
		},
		Clases: enums.ClasesCatalogoValidas(),
	}, nil
}

func (mockCatalogoRoutesUC) Crear(_ context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	return &dto.CatalogoItemDTO{
		ID:     10,
		Clase:  in.Clase,
		Codigo: in.Codigo,
		Nombre: in.Nombre,
		Activo: true,
		Orden:  0,
	}, nil
}

func (mockCatalogoRoutesUC) Desactivar(_ context.Context, id int64, _ int64) (*dto.DesactivarCatalogoResponseDTO, error) {
	return &dto.DesactivarCatalogoResponseDTO{
		Activo: false,
		ID:     id,
	}, nil
}

func (mockCatalogoRoutesUC) Renombrar(_ context.Context, in dto.RenombrarCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	return &dto.CatalogoItemDTO{
		ID:     in.ID,
		Nombre: in.Nombre,
		Activo: true,
	}, nil
}

type mockGeoRoutesUC struct{}

func (mockGeoRoutesUC) Areas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection("areas_verdes"), nil
}

func (mockGeoRoutesUC) Zonas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection("zonas"), nil
}

func (mockGeoRoutesUC) Capa(_ context.Context, capa string, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection(capa), nil
}

func (mockGeoRoutesUC) Capas(_ context.Context) (dto.CapasIndexDTO, error) {
	return dto.CapasIndexDTO{CapasConocidas: []string{"jardines_reserva", "xerofitica"}, Cargadas: []dto.CapaCountDTO{}}, nil
}

func (mockGeoRoutesUC) Resumen(_ context.Context) (dto.ResumenDTO, error) {
	return dto.ResumenDTO{CRS: "EPSG:4326", Areas: 521, Zonas: 534}, nil
}

func (mockGeoRoutesUC) Edificios(_ context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`), nil
}

type mockAreaVerdeRoutesUC struct{}

func (mockAreaVerdeRoutesUC) Fichas(_ context.Context, _ string) ([]dto.FichaDTO, error) {
	return []dto.FichaDTO{
		{FeatureID: "AV-0001", Nombre: "Área 1", ConGeom: true},
	}, nil
}

func (mockAreaVerdeRoutesUC) ActualizarFicha(_ context.Context, id string, req dto.ActualizarFichaDTO, _ *int64) (*dto.FichaDTO, error) {
	return &dto.FichaDTO{FeatureID: id, Nombre: req.Nombre, Uso: req.Uso, ConGeom: true}, nil
}

func (mockAreaVerdeRoutesUC) CrearSinGeom(_ context.Context, req dto.CrearAreaSinGeomDTO, _ *int64) (*dto.FichaDTO, error) {
	return &dto.FichaDTO{FeatureID: req.FeatureID, Nombre: req.Nombre, Uso: req.Uso, ConGeom: false}, nil
}

func (mockAreaVerdeRoutesUC) Baja(_ context.Context, _ string, _ *int64) error {
	return nil
}

type mockZonaRoutesUC struct{}

func (mockZonaRoutesUC) Listar(_ context.Context) ([]dto.ZonaSupervisionDTO, error) {
	return []dto.ZonaSupervisionDTO{{ID: 1, Codigo: "Z1", Nombre: "Zona 1", ConGeom: true, Activo: true}}, nil
}

func (mockZonaRoutesUC) Crear(_ context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error) {
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: req.Codigo, Nombre: req.Nombre, ConGeom: true, Activo: true}, nil
}

func (mockZonaRoutesUC) Actualizar(_ context.Context, codigo string, req dto.ActualizarZonaSupervisionDTO, _ *int64) (dto.ZonaSupervisionDTO, error) {
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: codigo, Nombre: req.Nombre, ConGeom: req.GeoJSON != "", Activo: true}, nil
}

func (mockZonaRoutesUC) Baja(_ context.Context, codigo string, _ *int64) (dto.ZonaSupervisionDTO, error) {
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: codigo, Nombre: "Zona 1", Activo: false, ConGeom: true}, nil
}

type mockCuadrillaRoutesUC struct{}

func (mockCuadrillaRoutesUC) Listar(_ context.Context) ([]dto.CuadrillaDTO, error) {
	return []dto.CuadrillaDTO{{ID: "C1", NombreFicticio: "Equipo 1", Turno: "manana", Activo: true}}, nil
}

func (mockCuadrillaRoutesUC) Crear(_ context.Context, req dto.CrearCuadrillaDTO) (dto.CuadrillaDTO, error) {
	return dto.CuadrillaDTO{ID: req.ID, NombreFicticio: req.Nombre, Turno: req.Turno, Activo: true}, nil
}

type mockLugarRoutesUC struct{}

func (mockLugarRoutesUC) Listar(_ context.Context) ([]dto.LugarDTO, error) {
	return []dto.LugarDTO{{ID: 1, Nombre: "Lugar 1", Lat: -12.07, Lon: -77.08, Activo: true}}, nil
}

func (mockLugarRoutesUC) Crear(_ context.Context, req dto.CrearLugarDTO) (dto.LugarDTO, error) {
	return dto.LugarDTO{ID: 1, Nombre: req.Nombre, Lat: req.Lat, Lon: req.Lon, Activo: true}, nil
}

type mockEspecieRoutesUC struct{}

func (mockEspecieRoutesUC) Listar(_ context.Context) ([]dto.EspecieDTO, error) {
	return []dto.EspecieDTO{{ID: 1, NombreCientifico: "Species 1", NombreComun: "Comun 1", Activo: true}}, nil
}

func (mockEspecieRoutesUC) Crear(_ context.Context, req dto.CrearEspecieDTO) (dto.EspecieDTO, error) {
	return dto.EspecieDTO{ID: 1, NombreCientifico: req.Cientifico, NombreComun: req.Comun, Activo: true}, nil
}

type mockEjemplarRoutesUC struct{}

func (mockEjemplarRoutesUC) Listar(_ context.Context, limit, offset int, _ string) (dto.EjemplaresPaginadosDTO, error) {
	return dto.EjemplaresPaginadosDTO{
		Ejemplares: []dto.EjemplarDTO{{ID: 1, Codigo: "EJ-1", Activo: true}},
		Total:      1,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (mockEjemplarRoutesUC) Crear(_ context.Context, req dto.EjemplarDTO) (dto.EjemplarDTO, error) {
	req.ID = 10
	req.Activo = true
	return req, nil
}

func (mockEjemplarRoutesUC) Actualizar(_ context.Context, id int64, req dto.ActualizarEjemplarDTO, _ *int64) (dto.EjemplarDTO, error) {
	return dto.EjemplarDTO{ID: id, Codigo: "EJ-1", Salud: req.Salud, Activo: true}, nil
}

func (mockEjemplarRoutesUC) Recodificar(_ context.Context, id int64, req dto.RecodificarDTO, _ *int64) (dto.CodigoHistoricoDTO, error) {
	return dto.CodigoHistoricoDTO{ID: 1, EjemplarID: id, CodigoAnterior: "OLD", CodigoNuevo: req.Codigo}, nil
}

func (mockEjemplarRoutesUC) ListarCodigos(_ context.Context, id int64) ([]dto.CodigoHistoricoDTO, error) {
	return []dto.CodigoHistoricoDTO{{ID: 1, EjemplarID: id, CodigoAnterior: "OLD", CodigoNuevo: "EJ-1"}}, nil
}

type mockCatastroRefRoutesUC struct{}

func (mockCatastroRefRoutesUC) ListarPoligonos(_ context.Context) ([]dto.PoligonoCuadrillaDTO, error) {
	return []dto.PoligonoCuadrillaDTO{{ID: 1, FeatureID: "POL-1", Codigo: "P1", Nombre: "Poligono 1", Activo: true}}, nil
}

func (mockCatastroRefRoutesUC) ListarCapa(_ context.Context, _ string) ([]dto.CapaFichaDTO, error) {
	return []dto.CapaFichaDTO{{ID: 1, FeatureID: "F-1", Activo: true}}, nil
}

type mockInventarioRoutesUC struct{}

func (mockInventarioRoutesUC) Index(_ context.Context) (dto.IndiceInventarioDTO, error) {
	return dto.IndiceInventarioDTO{Capas: domainEntities.CapasConocidas, Cargadas: []dto.CapaCountDTO{}}, nil
}

func (mockInventarioRoutesUC) Capa(_ context.Context, capa string) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection(capa), nil
}

func (mockInventarioRoutesUC) Foto(_ context.Context, name string) (string, error) {
	return "", nil
}

type mockReservasMockRoutesUC struct{}

func (mockReservasMockRoutesUC) ObtenerAgenda(_ context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
	return dto.ReservasMockResponseDTO{Fake: true, Total: 0, Reservas: []dto.ReservaItemDTO{}}, []byte(`{"fake":true,"total":0,"reservas":[]}`), nil
}

type mockInventarioCampoRoutesUC struct{}

func (mockInventarioCampoRoutesUC) ListarTachos(_ context.Context) (dto.ListarTachosResponseDTO, error) {
	return dto.ListarTachosResponseDTO{Tachos: []dto.TachoDTO{}}, nil
}
func (mockInventarioCampoRoutesUC) GuardarTacho(_ context.Context, req dto.TachoDTO) (*dto.TachoDTO, error) {
	req.ID = 1
	return &req, nil
}
func (mockInventarioCampoRoutesUC) ActualizarTacho(_ context.Context, id int64, req dto.TachoDTO, _ []byte) (*dto.TachoDTO, error) {
	req.ID = id
	return &req, nil
}
func (mockInventarioCampoRoutesUC) CSVTachos(_ context.Context) (string, error) {
	return "codigo,lugar\n", nil
}
func (mockInventarioCampoRoutesUC) BajaTacho(_ context.Context, _ int64) error {
	return nil
}
func (mockInventarioCampoRoutesUC) ListarBebederos(_ context.Context) (dto.ListarBebederosResponseDTO, error) {
	return dto.ListarBebederosResponseDTO{Bebederos: []dto.BebederoDTO{}}, nil
}
func (mockInventarioCampoRoutesUC) GuardarBebedero(_ context.Context, req dto.BebederoDTO) (*dto.BebederoDTO, error) {
	req.ID = 1
	return &req, nil
}
func (mockInventarioCampoRoutesUC) ActualizarBebedero(_ context.Context, id int64, req dto.BebederoDTO, _ []byte) (*dto.BebederoDTO, error) {
	req.ID = id
	return &req, nil
}
func (mockInventarioCampoRoutesUC) BajaBebedero(_ context.Context, _ int64) error {
	return nil
}
func (mockInventarioCampoRoutesUC) ListarPuntos(_ context.Context, _ string) (dto.ListarPuntosResponseDTO, error) {
	return dto.ListarPuntosResponseDTO{Puntos: []dto.PuntoDTO{}}, nil
}
func (mockInventarioCampoRoutesUC) GuardarPunto(_ context.Context, req dto.PuntoDTO, _ []byte) (*dto.PuntoDTO, error) {
	req.ID = 1
	return &req, nil
}
func (mockInventarioCampoRoutesUC) ActualizarPunto(_ context.Context, id int64, req dto.PuntoDTO, _ []byte) (*dto.PuntoDTO, error) {
	req.ID = id
	return &req, nil
}
func (mockInventarioCampoRoutesUC) BajaPunto(_ context.Context, _ int64) error {
	return nil
}
func (mockInventarioCampoRoutesUC) FormatoPuntos(_ context.Context, _ []byte) (dto.FormatoPuntosResponseDTO, error) {
	return dto.FormatoPuntosResponseDTO{Filas: 0, Rechazados: []dto.RechazoDTO{}, ColumnasOmitidas: []string{}, Nota: "columnas omitidas por datos personales"}, nil
}
func (mockInventarioCampoRoutesUC) ListarReservas(_ context.Context, _, _ string) (dto.ListarReservasResponseDTO, error) {
	return dto.ListarReservasResponseDTO{Origen: "ficticio", Aviso: "Agenda ficticia", Reservas: []dto.ReservaDTO{}}, nil
}
func (mockInventarioCampoRoutesUC) GuardarReserva(_ context.Context, req dto.ReservaDTO) (*dto.ReservaDTO, error) {
	req.ID = 1
	return &req, nil
}
func (mockInventarioCampoRoutesUC) ActualizarReserva(_ context.Context, id int64, req dto.ReservaDTO, _ []byte) (*dto.ReservaDTO, error) {
	req.ID = id
	return &req, nil
}
func (mockInventarioCampoRoutesUC) BajaReserva(_ context.Context, _ int64) error {
	return nil
}
func (mockInventarioCampoRoutesUC) ListarCapa(_ context.Context, _ string) (dto.ListarFichasCapaResponseDTO, error) {
	return dto.ListarFichasCapaResponseDTO{Filas: []dto.FichaCapaDTO{}}, nil
}
func (mockInventarioCampoRoutesUC) GuardarCapa(_ context.Context, _ string, req dto.FichaCapaDTO) (*dto.FichaCapaDTO, error) {
	req.ID = 1
	return &req, nil
}
func (mockInventarioCampoRoutesUC) ActualizarCapa(_ context.Context, _ string, id int64, req dto.FichaCapaDTO, _ []byte) (*dto.FichaCapaDTO, error) {
	req.ID = id
	return &req, nil
}
func (mockInventarioCampoRoutesUC) CSVCapa(_ context.Context, _ string) (string, error) {
	return "feature_id,nombre\n", nil
}
func (mockInventarioCampoRoutesUC) BajaCapa(_ context.Context, _ string, _ int64) error {
	return nil
}

type mockOperacionRoutesUC struct{}

func (mockOperacionRoutesUC) ListarCapataces(_ context.Context) ([]dto.CapatazDTO, error) {
	return []dto.CapatazDTO{
		{ID: "cap-norte", Equipo: "Equipo Norte", Turno: "mañana"},
	}, nil
}

func (mockOperacionRoutesUC) ListarActividades(_ context.Context, _ dto.FiltroIntervencionesDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection("actividades"), nil
}

func (mockOperacionRoutesUC) CrearActividad(_ context.Context, in dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error) {
	return dto.CrearIntervencionResponseDTO{
		Creada:  true,
		Feature: domainEntities.Feature{ID: in.ID},
	}, nil
}

func (mockOperacionRoutesUC) AsignarActividad(_ context.Context, in dto.AsignarIntervencionDTO) (domainEntities.Feature, error) {
	return domainEntities.Feature{ID: in.ID}, nil
}

func (mockOperacionRoutesUC) CambiarEstado(_ context.Context, in dto.CambiarEstadoDTO) (domainEntities.Feature, error) {
	return domainEntities.Feature{ID: in.ID}, nil
}

func (mockOperacionRoutesUC) ArchivarActividad(_ context.Context, in dto.ArchivarIntervencionDTO) (dto.ArchivarIntervencionResponseDTO, error) {
	return dto.ArchivarIntervencionResponseDTO{
		Archivada: true,
		ID:        in.ID,
	}, nil
}

func (mockOperacionRoutesUC) Timeline(_ context.Context, id string) (dto.TimelineResponseDTO, error) {
	return dto.TimelineResponseDTO{
		ActividadID: id,
		Eventos:     []dto.EventoTimelineDTO{},
	}, nil
}

func (mockOperacionRoutesUC) GuardarFicha(_ context.Context, _ dto.FichaIntervencionDTO) error {
	return nil
}

func (mockOperacionRoutesUC) CrearAvance(_ context.Context, _ dto.CrearAvanceDTO) error {
	return nil
}

type mockSolicitudRoutesUC struct{}

func (mockSolicitudRoutesUC) Listar(_ context.Context) (*dto.SolicitudesResponseDTO, error) {
	return &dto.SolicitudesResponseDTO{Solicitudes: []dto.SolicitudDTO{}}, nil
}

func (mockSolicitudRoutesUC) Crear(_ context.Context, in dto.CrearSolicitudDTO) (*dto.SolicitudDTO, error) {
	return &dto.SolicitudDTO{ID: in.ID, Titulo: in.Titulo}, nil
}

func (mockSolicitudRoutesUC) Editar(_ context.Context, in dto.EditarSolicitudDTO) (*dto.SolicitudDTO, error) {
	return &dto.SolicitudDTO{ID: in.ID, Titulo: in.Titulo}, nil
}

type mockServicioTercerizadoRoutesUC struct{}

func (mockServicioTercerizadoRoutesUC) Listar(_ context.Context) (*dto.OrdenesResponseDTO, error) {
	return &dto.OrdenesResponseDTO{Ordenes: []dto.OrdenDTO{}}, nil
}

func (mockServicioTercerizadoRoutesUC) Crear(_ context.Context, in dto.CrearOrdenDTO, _, _ string) (*dto.OrdenDTO, error) {
	return &dto.OrdenDTO{ID: in.ID, ActividadID: in.ActividadID}, nil
}

func (mockServicioTercerizadoRoutesUC) Editar(_ context.Context, in dto.EditarOrdenDTO, _, _ string) (*dto.EditarOrdenResponseDTO, error) {
	return &dto.EditarOrdenResponseDTO{Orden: dto.OrdenDTO{ID: in.ID}}, nil
}

type mockRiegoRoutesUC struct{}

func (mockRiegoRoutesUC) Listar(_ context.Context, _ string) (*dto.RiegoResponseDTO, error) {
	return &dto.RiegoResponseDTO{Registros: []dto.RiegoDTO{}}, nil
}

func (mockRiegoRoutesUC) Crear(_ context.Context, in dto.CrearRiegoDTO, _, _ string) (*dto.CrearRiegoResponseDTO, error) {
	return &dto.CrearRiegoResponseDTO{ID: in.ID}, nil
}

type mockPodaRoutesUC struct{}

func (mockPodaRoutesUC) Listar(_ context.Context) (*dto.PodasResponseDTO, error) {
	return &dto.PodasResponseDTO{Podas: []dto.PodaDTO{}}, nil
}

func (mockPodaRoutesUC) Crear(_ context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error) {
	return &dto.PodaDTO{ID: in.ID, Codigo: in.Codigo}, nil
}

func (mockPodaRoutesUC) Editar(_ context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error) {
	return &dto.PodaDTO{ID: in.ID, Codigo: in.Codigo}, nil
}

func (mockPodaRoutesUC) Archivar(_ context.Context, _ string) error {
	return nil
}

type mockViveroRoutesUC struct{}

func (mockViveroRoutesUC) Listar(_ context.Context, _ string) (*dto.ViveroResponseDTO, error) {
	return &dto.ViveroResponseDTO{Registros: []dto.ViveroDTO{}}, nil
}

func (mockViveroRoutesUC) Crear(_ context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error) {
	return &dto.ViveroDTO{ID: in.ID, Area: in.Area}, nil
}

func (mockViveroRoutesUC) Editar(_ context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error) {
	return &dto.ViveroDTO{ID: in.ID, Area: in.Area}, nil
}

func (mockViveroRoutesUC) Archivar(_ context.Context, _ string) error {
	return nil
}

type mockEvidenciaRoutesUC struct{}

func (mockEvidenciaRoutesUC) Listar(_ context.Context, _ string) (*dto.ListarEvidenciasResponseDTO, error) {
	return &dto.ListarEvidenciasResponseDTO{Evidencias: []dto.EvidenciaDTO{}}, nil
}

func (mockEvidenciaRoutesUC) Subir(_ context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
	return &dto.SubirEvidenciaResponseDTO{ID: in.ID, Idempotente: false}, nil
}

func (mockEvidenciaRoutesUC) Abrir(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return io.NopCloser(bytes.NewReader([]byte("test"))), "image/jpeg", nil
}

type mockReporteRoutesUC struct{}

func (mockReporteRoutesUC) ObtenerReporte(_ context.Context, _ dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error) {
	return &dto.ReporteResponseDTO{
		Aviso:      "aviso",
		PorEstado:  []dto.ConteoReporteDTO{},
		Filas:      []dto.FilaReporteDTO{},
		Pendientes: []dto.HuecoIndicadorDTO{},
	}, nil
}

func (mockReporteRoutesUC) Exportar(_ context.Context, _ dto.FiltroReporteDTO, _ string, w io.Writer) error {
	_, err := io.WriteString(w, "export")
	return err
}

type mockIARoutesUC struct{}

func (mockIARoutesUC) Sugerir(_ context.Context, _ string) dto.SugerenciaIADTO {
	return dto.SugerenciaIADTO{
		Codigo:         "riego",
		Etiqueta:       "Riego",
		Explicacion:    "regla",
		Confianza:      "baja",
		RequiereHumano: true,
	}
}

type mockLoteRoutesUC struct{}

func (mockLoteRoutesUC) Importar(_ context.Context, _ int64, in dto.ImportarLoteDTO) (*dto.ImportarLoteResponseDTO, error) {
	return &dto.ImportarLoteResponseDTO{LoteID: 1, Filas: len(in.Filas)}, nil
}

func (mockLoteRoutesUC) Revertir(_ context.Context, loteID, _ int64, _ bool) (*dto.ReporteReversionDTO, error) {
	return &dto.ReporteReversionDTO{LoteID: loteID, Revertidas: []string{"1"}, Excluidas: []dto.ExcluidaDTO{}}, nil
}

func (mockLoteRoutesUC) Editar(_ context.Context, _ int64, in dto.EditarAuditoriaDTO) (*dto.EditarAuditoriaResponseDTO, error) {
	return &dto.EditarAuditoriaResponseDTO{Editada: true, EntidadID: in.EntidadID}, nil
}

func (mockLoteRoutesUC) Timeline(_ context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	if f.Entidad == "" || f.EntidadID == "" {
		return nil, apperrors.InputError{Reason: "entidad y entidad_id son obligatorios"}
	}
	return []dto.EventoAuditoriaDTO{
		{ID: 1, Entidad: f.Entidad, EntidadID: f.EntidadID, Accion: "edicion", UsuarioID: 1, Usuario: "admin", Nombre: "Admin", CreatedAt: "2026-10-01T00:00:00Z"},
	}, nil
}

func (mockLoteRoutesUC) Historial(_ context.Context, _ dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	return []dto.EventoAuditoriaDTO{
		{ID: 1, Entidad: "catalogos", EntidadID: "1", Accion: "edicion", UsuarioID: 1, Usuario: "admin", Nombre: "Admin", CreatedAt: "2026-10-01T00:00:00Z"},
	}, nil
}

type mockImportacionRoutesUC struct{}

func (mockImportacionRoutesUC) Entidades(_ context.Context) []string {
	return []string{"areas_verdes", "lugares"}
}

func (mockImportacionRoutesUC) Previsualizar(_ context.Context, entidad, nombre string, body []byte, usuarioID int64) (*dto.VistaPreviaResponseDTO, error) {
	return &dto.VistaPreviaResponseDTO{
		ID:      1,
		LoteID:  1,
		Entidad: entidad,
		Validas: 1,
	}, nil
}

func (mockImportacionRoutesUC) Confirmar(_ context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
	return &dto.ConfirmarImportacionResponseDTO{
		LoteID:  loteID,
		Validas: 1,
		Escrito: true,
	}, nil
}

type mockZonificacionRoutesUC struct{}

func (mockZonificacionRoutesUC) ListarSectores(context.Context, bool) (*dto.SectorListDTO, error) {
	return &dto.SectorListDTO{Sectores: []dto.SectorCapatazDTO{}}, nil
}
func (mockZonificacionRoutesUC) CrearSector(context.Context, dto.CrearSectorDTO) (dto.SectorCapatazDTO, error) {
	return dto.SectorCapatazDTO{ID: 1, Codigo: "sector-prueba", Nombre: "Sector de prueba", Color: "#3f73b0", Activo: true}, nil
}
func (mockZonificacionRoutesUC) ActualizarSector(context.Context, dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error) {
	return dto.SectorCapatazDTO{ID: 1, Codigo: "sector-prueba", Nombre: "Sector de prueba", Color: "#3f73b0", Activo: true}, nil
}
func (mockZonificacionRoutesUC) DesactivarSector(context.Context, string, int64) error { return nil }
func (mockZonificacionRoutesUC) ImportarSectores(context.Context, []dto.CrearSectorDTO, int64) (dto.ImportacionSectorDTO, error) {
	return dto.ImportacionSectorDTO{}, nil
}
func (mockZonificacionRoutesUC) ListarLugares(context.Context) ([]dto.LugarCatalogoDTO, error) {
	return []dto.LugarCatalogoDTO{}, nil
}
func (mockZonificacionRoutesUC) ResolverLugar(context.Context, dto.ResolverLugarDTO) (dto.LugarResueltoDTO, error) {
	return dto.LugarResueltoDTO{}, nil
}
func (mockZonificacionRoutesUC) Vias(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","name":"vias","features":[]}`), nil
}
func (mockZonificacionRoutesUC) ImportarVias(context.Context, []byte, int64) (dto.ImportacionViaDTO, error) {
	return dto.ImportacionViaDTO{}, nil
}
func (mockZonificacionRoutesUC) Cuarteles(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","name":"cuarteles","aviso":"sin archivo de cuarteles","features":[]}`), nil
}
func (mockZonificacionRoutesUC) Edificios(context.Context) ([]dto.EdificioRefDTO, error) {
	return []dto.EdificioRefDTO{}, nil
}
func (mockZonificacionRoutesUC) CrearReferente(context.Context, dto.CrearReferenteDTO) (dto.ReferenteDTO, error) {
	return dto.ReferenteDTO{}, nil
}

func setupTestRouter(swaggerEnabled bool) *gin.Engine {
	return setupTestRouterWithEnv(swaggerEnabled, "")
}

func setupTestRouterWithEnv(swaggerEnabled bool, appEnv string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	_ = engine.SetTrustedProxies([]string{})

	cfg := &config.Config{
		AppEnv: appEnv,
		Server: config.ServerConfig{
			Port:    "8080",
			GinMode: "release",
		},
		Swagger: config.SwaggerConfig{
			Enabled: swaggerEnabled,
		},
		Seguridad: config.SeguridadConfig{
			CORSOrigins:    []string{"*"},
			CookieSecure:   false,
			CookieSameSite: "lax",
		},
	}

	healthCtrl := controller.NewHealthController(mockSaludUC{})
	metaCtrl := controller.NewMetaController(mockContratoOpenAPI{}, zerolog.Nop())
	sesionCtrl := controller.NewSesionController(mockSesionRoutesUC{})
	usuarioCtrl := controller.NewUsuarioController(mockUsuarioRoutesUC{}, services.NewPermisosMemoria(), zerolog.Nop())
	catalogoCtrl := controller.NewCatalogoController(mockCatalogoRoutesUC{}, zerolog.Nop())
	geoCtrl := controller.NewGeoController(mockGeoRoutesUC{}, zerolog.Nop())
	areaVerdeCtrl := controller.NewAreaVerdeController(mockAreaVerdeRoutesUC{}, zerolog.Nop())
	catastroCtrl := controller.NewCatastroController(
		mockZonaRoutesUC{},
		mockCuadrillaRoutesUC{},
		mockLugarRoutesUC{},
		mockEspecieRoutesUC{},
		mockEjemplarRoutesUC{},
		mockCatastroRefRoutesUC{},
		zerolog.Nop(),
	)
	inventarioCtrl := controller.NewInventarioController(mockInventarioRoutesUC{}, zerolog.Nop())
	reservasMockCtrl := controller.NewReservasMockController(mockReservasMockRoutesUC{}, zerolog.Nop())
	inventarioCampoCtrl := controller.NewInventarioCampoController(mockInventarioCampoRoutesUC{}, zerolog.Nop())
	operacionCtrl := controller.NewIntervencionController(mockOperacionRoutesUC{}, zerolog.Nop())
	solicitudCtrl := controller.NewSolicitudController(mockSolicitudRoutesUC{}, zerolog.Nop())
	ordenCtrl := controller.NewServicioTercerizadoController(mockServicioTercerizadoRoutesUC{}, zerolog.Nop())
	riegoCtrl := controller.NewRiegoController(mockRiegoRoutesUC{}, zerolog.Nop())
	podaCtrl := controller.NewPodaController(mockPodaRoutesUC{}, zerolog.Nop())
	viveroCtrl := controller.NewViveroController(mockViveroRoutesUC{}, zerolog.Nop())
	evidenciaCtrl := controller.NewEvidenciaController(mockEvidenciaRoutesUC{}, zerolog.Nop())
	reporteCtrl := controller.NewReporteController(mockReporteRoutesUC{}, zerolog.Nop())
	iaCtrl := controller.NewIAController(mockIARoutesUC{}, zerolog.Nop())
	auditoriaCtrl := controller.NewAuditoriaController(mockLoteRoutesUC{}, zerolog.Nop())
	importacionCtrl := controller.NewImportacionController(mockImportacionRoutesUC{}, zerolog.Nop())
	zonificacionCtrl := controller.NewZonificacionController(mockZonificacionRoutesUC{}, zerolog.Nop())

	permisosSvc := services.NewPermisosMemoria()
	limitador := ratelimit.NewMemoriaLimitador(100, time.Minute)

	healthGrp := groups.NewHealthGroup(healthCtrl)
	metaGrp := groups.NewMetaGroup(metaCtrl)
	legadoGrp := groups.NewLegadoGroup(healthCtrl)
	swaggerGrp := groups.NewSwaggerGroup()
	sesionGrp := groups.NewSesionGroup(sesionCtrl)
	accesosGrp := groups.NewAccesosGroup(usuarioCtrl)
	catalogoGrp := groups.NewCatalogoGroup(catalogoCtrl, permisosSvc)
	geoGrp := groups.NewGeoGroup(geoCtrl, permisosSvc)
	catastroGrp := groups.NewCatastroGroup(areaVerdeCtrl, catastroCtrl, permisosSvc)
	inventarioGrp := groups.NewInventarioGroup(inventarioCtrl, permisosSvc)
	reservasMockGrp := groups.NewReservasMockGroup(reservasMockCtrl, permisosSvc)
	inventarioCampoGrp := groups.NewInventarioCampoGroup(inventarioCampoCtrl, permisosSvc)
	operacionGrp := groups.NewOperacionGroup(operacionCtrl, permisosSvc)
	atencionGrp := groups.NewAtencionGroup(solicitudCtrl, ordenCtrl, riegoCtrl, permisosSvc)
	podaGrp := groups.NewPodaGroup(podaCtrl, permisosSvc)
	viveroGrp := groups.NewViveroGroup(viveroCtrl, permisosSvc)
	evidenciaGrp := groups.NewEvidenciaGroup(evidenciaCtrl, permisosSvc)
	reporteGrp := groups.NewReporteGroup(reporteCtrl, permisosSvc)
	iaGrp := groups.NewIAGroup(iaCtrl, permisosSvc)
	auditoriaGrp := groups.NewAuditoriaGroup(auditoriaCtrl, permisosSvc)
	importacionGrp := groups.NewImportacionGroup(importacionCtrl, permisosSvc)
	zonificacionGrp := groups.NewZonificacionGroup(zonificacionCtrl, permisosSvc)

	r := routes.NewRouter(routes.RouterParams{
		Engine:               engine,
		Config:               cfg,
		Logger:               zerolog.Nop(),
		Limitador:            limitador,
		SesionUC:             mockSesionRoutesUC{},
		HealthGroup:          healthGrp,
		MetaGroup:            metaGrp,
		LegadoGroup:          legadoGrp,
		SwaggerGroup:         swaggerGrp,
		SesionGroup:          sesionGrp,
		AccesosGroup:         accesosGrp,
		CatalogoGroup:        catalogoGrp,
		GeoGroup:             geoGrp,
		CatastroGroup:        catastroGrp,
		InventarioGroup:      inventarioGrp,
		ReservasMockGroup:    reservasMockGrp,
		InventarioCampoGroup: inventarioCampoGrp,
		OperacionGroup:       operacionGrp,
		AtencionGroup:        atencionGrp,
		PodaGroup:            podaGrp,
		ViveroGroup:          viveroGrp,
		EvidenciaGroup:       evidenciaGrp,
		ReporteGroup:         reporteGrp,
		IAGroup:              iaGrp,
		AuditoriaGroup:       auditoriaGrp,
		ImportacionGroup:     importacionGrp,
		ZonificacionGroup:    zonificacionGrp,
	})
	r.Setup()
	return engine
}

func TestNoRouteRetorna404ConFormatoEsperado(t *testing.T) {
	engine := setupTestRouter(true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ruta-que-no-existe", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, se obtuvo %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("error leyendo respuesta JSON: %v", err)
	}
	if body["error"] != "ruta no encontrada" {
		t.Fatalf("mensaje esperado 'ruta no encontrada', obtenido %q", body["error"])
	}
}

func TestSwaggerCondicionadoPorConfiguracion(t *testing.T) {
	// 1. Con SWAGGER_ENABLED = false, swagger no debe estar montado -> 404
	engineSinSwagger := setupTestRouter(false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/swagger/index.html", nil)
	engineSinSwagger.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 con swagger deshabilitado, se obtuvo %d", w.Code)
	}
	var bodySinSwagger map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &bodySinSwagger)
	if bodySinSwagger["error"] != "ruta no encontrada" {
		t.Fatalf("se esperaba error 'ruta no encontrada', obtenido %q", bodySinSwagger["error"])
	}

	// 2. Con SWAGGER_ENABLED = true, swagger debe responder (200 o redirect)
	engineConSwagger := setupTestRouter(true)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/swagger/index.html", nil)
	engineConSwagger.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("con swagger habilitado no debe dar 404")
	}
}

func TestOpenAPICondicionadoPorConfiguracion(t *testing.T) {
	rutas := []string{"/api/v1/openapi.yaml", "/areas-verdes/v1/openapi.yaml"}

	// 1. Con AppEnv = "produccion", openapi.yaml responde 404 en ambos prefijos aunque Swagger.Enabled sea true o false
	for _, swaggerEnabled := range []bool{true, false} {
		engineProd := setupTestRouterWithEnv(swaggerEnabled, "produccion")
		for _, ruta := range rutas {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ruta, nil)
			engineProd.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("se esperaba 404 en %s con AppEnv=produccion y swaggerEnabled=%v, se obtuvo %d", ruta, swaggerEnabled, w.Code)
			}
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != "ruta no encontrada" {
				t.Fatalf("se esperaba error 'ruta no encontrada' en %s, obtenido %q", ruta, body["error"])
			}
		}
	}

	// 2. Con AppEnv develop, qa o vacío (""), openapi.yaml responde 200 en ambos prefijos aunque Swagger.Enabled sea false
	for _, env := range []string{"develop", "qa", ""} {
		engine := setupTestRouterWithEnv(false, env)
		for _, ruta := range rutas {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ruta, nil)
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("se esperaba 200 en %s con AppEnv=%q y swagger deshabilitado, se obtuvo %d", ruta, env, w.Code)
			}
		}
	}
}

func TestMontajeDoblePrefijos(t *testing.T) {
	engine := setupTestRouter(true)

	for _, ruta := range []string{
		"/areas-verdes/v1/health",
		"/areas-verdes/v1",
		"/api/v1",
		"/health",
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("se esperaba 200 en %s, obtenido %d", ruta, w.Code)
		}
	}

	// /api/v1/health debe dar 404 para paridad con la API anterior (decisión 17)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("se esperaba 404 en /api/v1/health, obtenido %d", w.Code)
	}
}

func TestRutasActualesRespondenIgual(t *testing.T) {
	engine := setupTestRouter(true)

	rutasCongeladas := []struct {
		metodo string
		ruta   string
		codigo int
	}{
		{http.MethodGet, "/health", 200},
		{http.MethodGet, "/api/v1", 200},
		{http.MethodGet, "/areas-verdes/v1", 200},
		{http.MethodGet, "/api/v1/openapi.yaml", 200},
		{http.MethodGet, "/areas-verdes/v1/openapi.yaml", 200},
		{http.MethodGet, "/areas-verdes/v1/health", 200},
		{http.MethodGet, "/api/v1/health", 404},
		{http.MethodGet, "/api/v1/sesion", 401},
		{http.MethodGet, "/areas-verdes/v1/sesion", 401},
		{http.MethodDelete, "/api/v1/sesion", 200},
		{http.MethodDelete, "/areas-verdes/v1/sesion", 200},
		// Accesos (sin sesión da 403 por paridad con API anterior)
		{http.MethodGet, "/api/v1/accesos/usuarios", 403},
		{http.MethodGet, "/areas-verdes/v1/accesos/usuarios", 403},
		{http.MethodGet, "/api/v1/catalogos", 401},
		{http.MethodGet, "/areas-verdes/v1/catalogos", 401},
		{http.MethodPost, "/api/v1/catalogos", 401},
		{http.MethodPost, "/areas-verdes/v1/catalogos", 401},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", 401},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", 401},
		{http.MethodPatch, "/api/v1/catalogos/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/catalogos/1", 401},
		// Geo de lectura (Lote 8)
		{http.MethodGet, "/api/v1/geo/resumen", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen", 401},
		{http.MethodGet, "/api/v1/geo/areas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/areas", 401},
		{http.MethodGet, "/api/v1/geo/zonas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas", 401},
		{http.MethodGet, "/api/v1/geo/capas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/capas", 401},
		{http.MethodGet, "/api/v1/geo/capas/jardines_reserva", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/capas/jardines_reserva", 401},
		{http.MethodGet, "/api/v1/geo/edificios", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios", 401},
		// Fichas de área (Lote 8)
		{http.MethodGet, "/api/v1/catastro/areas", 401},
		{http.MethodGet, "/areas-verdes/v1/catastro/areas", 401},
		{http.MethodPost, "/api/v1/catastro/areas", 401},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas", 401},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", 401},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001", 401},
		{http.MethodPost, "/api/v1/catastro/areas/AV-0001/baja", 401},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas/AV-0001/baja", 401},
		{http.MethodPatch, "/api/v1/catastro/zonas-supervision/Z1", 401},
		{http.MethodPatch, "/areas-verdes/v1/catastro/zonas-supervision/Z1", 401},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision/Z1/baja", 401},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja", 401},
		// Inventario heredado y reservas mock (Lote 10)
		{http.MethodGet, "/api/v1/geo/inventario", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario", 401},
		{http.MethodGet, "/api/v1/geo/inventario/fotos/x.jpg", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario/fotos/x.jpg", 401},
		{http.MethodGet, "/api/v1/geo/inventario/bebederos", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario/bebederos", 401},
		{http.MethodGet, "/api/v1/geo/reservas-mock", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/reservas-mock", 401},
		// Inventario campo frente 2B (Lote 11)
		{http.MethodGet, "/api/v1/inventario/tachos", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos", 401},
		{http.MethodPost, "/api/v1/inventario/tachos", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/tachos", 401},
		{http.MethodPatch, "/api/v1/inventario/tachos/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/inventario/tachos/1", 401},
		{http.MethodGet, "/api/v1/inventario/tachos.csv", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos.csv", 401},
		{http.MethodDelete, "/api/v1/inventario/tachos/1", 401},
		{http.MethodDelete, "/areas-verdes/v1/inventario/tachos/1", 401},
		{http.MethodGet, "/api/v1/inventario/bebederos", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/bebederos", 401},
		{http.MethodPost, "/api/v1/inventario/bebederos", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/bebederos", 401},
		{http.MethodPatch, "/api/v1/inventario/bebederos/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/inventario/bebederos/1", 401},
		{http.MethodDelete, "/api/v1/inventario/bebederos/1", 401},
		{http.MethodDelete, "/areas-verdes/v1/inventario/bebederos/1", 401},
		{http.MethodGet, "/api/v1/inventario/puntos", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/puntos", 401},
		{http.MethodPost, "/api/v1/inventario/puntos", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/puntos", 401},
		{http.MethodPatch, "/api/v1/inventario/puntos/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/inventario/puntos/1", 401},
		{http.MethodDelete, "/api/v1/inventario/puntos/1", 401},
		{http.MethodDelete, "/areas-verdes/v1/inventario/puntos/1", 401},
		{http.MethodPost, "/api/v1/inventario/formato/puntos", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/formato/puntos", 401},
		{http.MethodGet, "/api/v1/inventario/reservas", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/reservas", 401},
		{http.MethodPost, "/api/v1/inventario/reservas", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/reservas", 401},
		{http.MethodPatch, "/api/v1/inventario/reservas/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/inventario/reservas/1", 401},
		{http.MethodDelete, "/api/v1/inventario/reservas/1", 401},
		{http.MethodDelete, "/areas-verdes/v1/inventario/reservas/1", 401},
		{http.MethodGet, "/api/v1/inventario/capas/fauna", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/capas/fauna", 401},
		{http.MethodPost, "/api/v1/inventario/capas/fauna", 401},
		{http.MethodPost, "/areas-verdes/v1/inventario/capas/fauna", 401},
		{http.MethodPatch, "/api/v1/inventario/capas/fauna/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/inventario/capas/fauna/1", 401},
		{http.MethodGet, "/api/v1/inventario/export/fauna", 401},
		{http.MethodGet, "/areas-verdes/v1/inventario/export/fauna", 401},
		{http.MethodDelete, "/api/v1/inventario/capas/fauna/1", 401},
		{http.MethodDelete, "/areas-verdes/v1/inventario/capas/fauna/1", 401},
		// Solicitudes, órdenes y riego (Lote 13)
		{http.MethodGet, "/api/v1/solicitudes", 401},
		{http.MethodGet, "/areas-verdes/v1/solicitudes", 401},
		{http.MethodPost, "/api/v1/solicitudes", 401},
		{http.MethodPost, "/areas-verdes/v1/solicitudes", 401},
		{http.MethodPatch, "/api/v1/solicitudes/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/solicitudes/1", 401},
		{http.MethodGet, "/api/v1/ordenes", 401},
		{http.MethodGet, "/areas-verdes/v1/ordenes", 401},
		{http.MethodPost, "/api/v1/ordenes", 401},
		{http.MethodPost, "/areas-verdes/v1/ordenes", 401},
		{http.MethodPatch, "/api/v1/ordenes/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/ordenes/1", 401},
		{http.MethodGet, "/api/v1/riego", 401},
		{http.MethodGet, "/areas-verdes/v1/riego", 401},
		{http.MethodPost, "/api/v1/riego", 401},
		{http.MethodPost, "/areas-verdes/v1/riego", 401},
		// Poda y vivero (Lote 14)
		{http.MethodGet, "/api/v1/podas", 401},
		{http.MethodGet, "/areas-verdes/v1/podas", 401},
		{http.MethodPost, "/api/v1/podas", 401},
		{http.MethodPost, "/areas-verdes/v1/podas", 401},
		{http.MethodPatch, "/api/v1/podas/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/podas/1", 401},
		{http.MethodPost, "/api/v1/podas/1/archivar", 401},
		{http.MethodPost, "/areas-verdes/v1/podas/1/archivar", 401},
		{http.MethodGet, "/api/v1/vivero", 401},
		{http.MethodGet, "/areas-verdes/v1/vivero", 401},
		{http.MethodPost, "/api/v1/vivero", 401},
		{http.MethodPost, "/areas-verdes/v1/vivero", 401},
		{http.MethodPatch, "/api/v1/vivero/1", 401},
		{http.MethodPatch, "/areas-verdes/v1/vivero/1", 401},
		{http.MethodPost, "/api/v1/vivero/1/archivar", 401},
		{http.MethodPost, "/areas-verdes/v1/vivero/1/archivar", 401},
		// Reportes e IA (Lote 16)
		{http.MethodGet, "/api/v1/reportes/labores", 401},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores", 401},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo", 401},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo", 401},
		// Importaciones (Lote 19)
		{http.MethodGet, "/api/v1/importaciones/entidades", 401},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades", 401},
		{http.MethodPost, "/api/v1/importaciones", 401},
		{http.MethodPost, "/areas-verdes/v1/importaciones", 401},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar", 401},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar", 401},
		{http.MethodGet, "/api/v1/no-existe", 404},
		{http.MethodGet, "/areas-verdes/v1/no-existe", 404},
	}

	vistas := map[string]bool{}
	for _, rt := range engine.Routes() {
		vistas[rt.Method+" "+rt.Path] = true
	}

	for _, want := range rutasCongeladas {
		if want.codigo == 404 {
			continue
		}
		path := want.ruta
		switch path {
		case "/api/v1/importaciones/1/confirmar":
			path = "/api/v1/importaciones/:id/confirmar"
		case "/areas-verdes/v1/importaciones/1/confirmar":
			path = "/areas-verdes/v1/importaciones/:id/confirmar"
		case "/api/v1/catalogos/1/desactivar":
			path = "/api/v1/catalogos/:id/desactivar"
		case "/areas-verdes/v1/catalogos/1/desactivar":
			path = "/areas-verdes/v1/catalogos/:id/desactivar"
		case "/api/v1/catalogos/1":
			path = "/api/v1/catalogos/:id"
		case "/areas-verdes/v1/catalogos/1":
			path = "/areas-verdes/v1/catalogos/:id"
		case "/api/v1/geo/capas/jardines_reserva":
			path = "/api/v1/geo/capas/:capa"
		case "/areas-verdes/v1/geo/capas/jardines_reserva":
			path = "/areas-verdes/v1/geo/capas/:capa"
		case "/api/v1/catastro/areas/AV-0001":
			path = "/api/v1/catastro/areas/:id"
		case "/areas-verdes/v1/catastro/areas/AV-0001":
			path = "/areas-verdes/v1/catastro/areas/:id"
		case "/api/v1/catastro/areas/AV-0001/baja":
			path = "/api/v1/catastro/areas/:id/baja"
		case "/areas-verdes/v1/catastro/areas/AV-0001/baja":
			path = "/areas-verdes/v1/catastro/areas/:id/baja"
		case "/api/v1/catastro/zonas-supervision/Z1":
			path = "/api/v1/catastro/zonas-supervision/:codigo"
		case "/areas-verdes/v1/catastro/zonas-supervision/Z1":
			path = "/areas-verdes/v1/catastro/zonas-supervision/:codigo"
		case "/api/v1/catastro/zonas-supervision/Z1/baja":
			path = "/api/v1/catastro/zonas-supervision/:codigo/baja"
		case "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja":
			path = "/areas-verdes/v1/catastro/zonas-supervision/:codigo/baja"
		case "/api/v1/geo/inventario/fotos/x.jpg":
			path = "/api/v1/geo/inventario/fotos/:name"
		case "/areas-verdes/v1/geo/inventario/fotos/x.jpg":
			path = "/areas-verdes/v1/geo/inventario/fotos/:name"
		case "/api/v1/geo/inventario/bebederos":
			path = "/api/v1/geo/inventario/:capa"
		case "/areas-verdes/v1/geo/inventario/bebederos":
			path = "/areas-verdes/v1/geo/inventario/:capa"
		case "/api/v1/inventario/tachos/1":
			path = "/api/v1/inventario/tachos/:id"
		case "/areas-verdes/v1/inventario/tachos/1":
			path = "/areas-verdes/v1/inventario/tachos/:id"
		case "/api/v1/inventario/bebederos/1":
			path = "/api/v1/inventario/bebederos/:id"
		case "/areas-verdes/v1/inventario/bebederos/1":
			path = "/areas-verdes/v1/inventario/bebederos/:id"
		case "/api/v1/inventario/puntos/1":
			path = "/api/v1/inventario/puntos/:id"
		case "/areas-verdes/v1/inventario/puntos/1":
			path = "/areas-verdes/v1/inventario/puntos/:id"
		case "/api/v1/inventario/reservas/1":
			path = "/api/v1/inventario/reservas/:id"
		case "/areas-verdes/v1/inventario/reservas/1":
			path = "/areas-verdes/v1/inventario/reservas/:id"
		case "/api/v1/inventario/capas/fauna":
			path = "/api/v1/inventario/capas/:capa"
		case "/areas-verdes/v1/inventario/capas/fauna":
			path = "/areas-verdes/v1/inventario/capas/:capa"
		case "/api/v1/inventario/capas/fauna/1":
			path = "/api/v1/inventario/capas/:capa/:id"
		case "/areas-verdes/v1/inventario/capas/fauna/1":
			path = "/areas-verdes/v1/inventario/capas/:capa/:id"
		case "/api/v1/inventario/export/fauna":
			path = "/api/v1/inventario/export/:capa"
		case "/areas-verdes/v1/inventario/export/fauna":
			path = "/areas-verdes/v1/inventario/export/:capa"
		case "/api/v1/solicitudes/1":
			path = "/api/v1/solicitudes/:id"
		case "/areas-verdes/v1/solicitudes/1":
			path = "/areas-verdes/v1/solicitudes/:id"
		case "/api/v1/ordenes/1":
			path = "/api/v1/ordenes/:id"
		case "/areas-verdes/v1/ordenes/1":
			path = "/areas-verdes/v1/ordenes/:id"
		case "/api/v1/podas/1":
			path = "/api/v1/podas/:id"
		case "/areas-verdes/v1/podas/1":
			path = "/areas-verdes/v1/podas/:id"
		case "/api/v1/podas/1/archivar":
			path = "/api/v1/podas/:id/archivar"
		case "/areas-verdes/v1/podas/1/archivar":
			path = "/areas-verdes/v1/podas/:id/archivar"
		case "/api/v1/vivero/1":
			path = "/api/v1/vivero/:id"
		case "/areas-verdes/v1/vivero/1":
			path = "/areas-verdes/v1/vivero/:id"
		case "/api/v1/vivero/1/archivar":
			path = "/api/v1/vivero/:id/archivar"
		case "/areas-verdes/v1/vivero/1/archivar":
			path = "/areas-verdes/v1/vivero/:id/archivar"
		}
		key := want.metodo + " " + path
		if !vistas[key] {
			t.Errorf("falta la ruta registrada %s", key)
		}
	}

	for _, want := range rutasCongeladas {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(want.metodo, want.ruta, nil))
		if w.Code != want.codigo {
			t.Errorf("%s %s -> %d, se esperaba %d (%s)", want.metodo, want.ruta, w.Code, want.codigo, w.Body.String())
		}
	}
}

func TestRutasCatalogos_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/catalogos"},
		{http.MethodPost, "/api/v1/catalogos"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar"},
		{http.MethodPatch, "/api/v1/catalogos/1"},
		{http.MethodGet, "/areas-verdes/v1/catalogos"},
		{http.MethodPost, "/areas-verdes/v1/catalogos"},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar"},
		{http.MethodPatch, "/areas-verdes/v1/catalogos/1"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasCatalogos_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catalogos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-norte", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catalogos", "token-norte", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catalogos/1", "token-norte", `{"nombre":"Otro"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/catalogos/1", "token-norte", `{"nombre":"Otro"}`, 403, "su rol no tiene ese permiso"},

		// Coordinación: consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-coordinacion", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-coordinacion", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catalogos/1", "token-coordinacion", `{"nombre":"Otro"}`, 403, "su rol no tiene ese permiso"},

		// Jefatura: consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-jefatura", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-jefatura", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catalogos/1", "token-jefatura", `{"nombre":"Otro"}`, 403, "su rol no tiene ese permiso"},

		// Admin: consultar=true, catalogos=true
		{http.MethodGet, "/api/v1/catalogos", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catalogos", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-admin", `{"clase":"estado","codigo":"nuevo","nombre":"Nuevo"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/catalogos", "token-admin", `{"clase":"estado","codigo":"nuevo","nombre":"Nuevo"}`, 201, ""},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-admin", "", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", "token-admin", "", 200, ""},
		{http.MethodPatch, "/api/v1/catalogos/1", "token-admin", `{"nombre":"Nombre corregido"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/catalogos/1", "token-admin", `{"nombre":"Nombre corregido"}`, 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasGeoYCatastro_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/geo/resumen"},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen"},
		{http.MethodGet, "/api/v1/geo/areas"},
		{http.MethodGet, "/areas-verdes/v1/geo/areas"},
		{http.MethodGet, "/api/v1/geo/zonas"},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas"},
		{http.MethodGet, "/api/v1/geo/capas"},
		{http.MethodGet, "/areas-verdes/v1/geo/capas"},
		{http.MethodGet, "/api/v1/geo/capas/jardines_reserva"},
		{http.MethodGet, "/areas-verdes/v1/geo/capas/jardines_reserva"},
		{http.MethodGet, "/api/v1/geo/edificios"},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios"},
		{http.MethodGet, "/api/v1/catastro/areas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/areas"},
		{http.MethodPost, "/api/v1/catastro/areas"},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas"},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001"},
		{http.MethodPost, "/api/v1/catastro/areas/AV-0001/baja"},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas/AV-0001/baja"},
		{http.MethodPatch, "/api/v1/catastro/zonas-supervision/Z1"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/zonas-supervision/Z1"},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision/Z1/baja"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasGeoYCatastro_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/resumen", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/geo/areas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/areas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-norte", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-norte", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas/AV-0001/baja", "token-norte", `{}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas/AV-0001/baja", "token-norte", `{}`, 200, ""},
		{http.MethodPatch, "/api/v1/catastro/zonas-supervision/Z1", "token-norte", `{"nombre":"Norte","geom":{"type":"MultiPolygon","coordinates":[]}}`, 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision/Z1/baja", "token-norte", `{}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja", "token-norte", `{}`, 200, ""},

		// Coordinación: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/zonas", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-coordinacion", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-coordinacion", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},

		// Jefatura: consultar=true, registrar=false
		{http.MethodGet, "/api/v1/geo/capas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/capas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/areas", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-jefatura", `{"nombre":"Nueva","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas", "token-jefatura", `{"nombre":"Nueva","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-jefatura", `{"nombre":"Modif","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001", "token-jefatura", `{"nombre":"Modif","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/areas/AV-0001/baja", "token-jefatura", `{}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catastro/zonas-supervision/Z1", "token-jefatura", `{"nombre":"Norte"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja", "token-jefatura", `{}`, 403, "su rol no tiene ese permiso"},

		// Admin: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/edificios", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-admin", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-admin", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasCatastroMaestro_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/catastro/zonas-supervision"},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision"},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision"},

		{http.MethodGet, "/api/v1/catastro/poligonos"},
		{http.MethodGet, "/areas-verdes/v1/catastro/poligonos"},

		{http.MethodGet, "/api/v1/catastro/cuadrillas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/cuadrillas"},
		{http.MethodPost, "/api/v1/catastro/cuadrillas"},
		{http.MethodPost, "/areas-verdes/v1/catastro/cuadrillas"},

		{http.MethodGet, "/api/v1/catastro/lugares"},
		{http.MethodGet, "/areas-verdes/v1/catastro/lugares"},
		{http.MethodPost, "/api/v1/catastro/lugares"},
		{http.MethodPost, "/areas-verdes/v1/catastro/lugares"},

		{http.MethodGet, "/api/v1/catastro/especies"},
		{http.MethodGet, "/areas-verdes/v1/catastro/especies"},
		{http.MethodPost, "/api/v1/catastro/especies"},
		{http.MethodPost, "/areas-verdes/v1/catastro/especies"},

		{http.MethodGet, "/api/v1/catastro/ejemplares"},
		{http.MethodGet, "/areas-verdes/v1/catastro/ejemplares"},
		{http.MethodPost, "/api/v1/catastro/ejemplares"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares"},
		{http.MethodPatch, "/api/v1/catastro/ejemplares/1"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/ejemplares/1"},

		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos"},
		{http.MethodGet, "/areas-verdes/v1/catastro/ejemplares/1/codigos"},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares/1/codigos"},

		{http.MethodGet, "/api/v1/catastro/fauna"},
		{http.MethodGet, "/areas-verdes/v1/catastro/fauna"},
		{http.MethodGet, "/api/v1/catastro/puertas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/puertas"},
		{http.MethodGet, "/api/v1/catastro/playas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/playas"},
		{http.MethodGet, "/api/v1/catastro/veredas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/veredas"},
		{http.MethodGet, "/api/v1/catastro/xerofiticas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/xerofiticas"},
		{http.MethodGet, "/api/v1/catastro/jardines-reserva"},
		{http.MethodGet, "/areas-verdes/v1/catastro/jardines-reserva"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasCatastroMaestro_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// 1. Capataz (norte): consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-norte", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/poligonos", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/cuadrillas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-norte", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/lugares", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-norte", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/especies", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/especies", "token-norte", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-norte", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/ejemplares/1", "token-norte", `{"salud":"bueno"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/ejemplares/1", "token-norte", `{"salud":"bueno"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-norte", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/fauna", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/puertas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/playas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/veredas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/xerofiticas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/jardines-reserva", "token-norte", "", 200, ""},

		// 2. Coordinación: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-coordinacion", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-coordinacion", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/ejemplares/1", "token-coordinacion", `{"salud":"bueno"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/catastro/ejemplares/1", "token-coordinacion", `{"salud":"bueno"}`, 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-coordinacion", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},

		// 3. Admin: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-admin", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-admin", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-admin", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/especies", "token-admin", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-admin", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-admin", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},

		// 4. Jefatura: consultar=true, registrar=false
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/poligonos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/cuadrillas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/lugares", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/especies", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/fauna", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-jefatura", `{"codigo":"Z1","nombre":"Zona 1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision", "token-jefatura", `{"codigo":"Z1","nombre":"Zona 1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-jefatura", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-jefatura", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/especies", "token-jefatura", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-jefatura", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catastro/ejemplares/1", "token-jefatura", `{"salud":"bueno"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-jefatura", `{"codigo_nuevo":"AV-NEW"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares/1/codigos", "token-jefatura", `{"codigo_nuevo":"AV-NEW"}`, 403, "su rol no tiene ese permiso"},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasInventarioYReservasMock_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/geo/inventario"},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario"},
		{http.MethodGet, "/api/v1/geo/inventario/fotos/x.jpg"},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario/fotos/x.jpg"},
		{http.MethodGet, "/api/v1/geo/inventario/bebederos"},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario/bebederos"},
		{http.MethodGet, "/api/v1/geo/reservas-mock"},
		{http.MethodGet, "/areas-verdes/v1/geo/reservas-mock"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasInventarioYReservasMock_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		statusEsperado int
	}

	casos := []caso{
		// Capataz (norte): consultar=true
		{http.MethodGet, "/api/v1/geo/inventario", "token-norte", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario", "token-norte", 200},
		{http.MethodGet, "/api/v1/geo/inventario/bebederos", "token-norte", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario/bebederos", "token-norte", 200},
		{http.MethodGet, "/api/v1/geo/reservas-mock", "token-norte", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/reservas-mock", "token-norte", 200},

		// Coordinación: consultar=true
		{http.MethodGet, "/api/v1/geo/inventario", "token-coordinacion", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario", "token-coordinacion", 200},
		{http.MethodGet, "/api/v1/geo/inventario/bebederos", "token-coordinacion", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/reservas-mock", "token-coordinacion", 200},

		// Jefatura: consultar=true
		{http.MethodGet, "/api/v1/geo/inventario", "token-jefatura", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario", "token-jefatura", 200},
		{http.MethodGet, "/api/v1/geo/reservas-mock", "token-jefatura", 200},

		// Admin: consultar=true
		{http.MethodGet, "/api/v1/geo/inventario", "token-admin", 200},
		{http.MethodGet, "/areas-verdes/v1/geo/inventario", "token-admin", 200},
		{http.MethodGet, "/api/v1/geo/reservas-mock", "token-admin", 200},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(c.metodo, c.ruta, nil)
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
	}
}

func TestRutasInventarioCampo_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/inventario/tachos"},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos"},
		{http.MethodPost, "/api/v1/inventario/tachos"},
		{http.MethodPost, "/areas-verdes/v1/inventario/tachos"},
		{http.MethodPatch, "/api/v1/inventario/tachos/1"},
		{http.MethodPatch, "/areas-verdes/v1/inventario/tachos/1"},
		{http.MethodGet, "/api/v1/inventario/tachos.csv"},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos.csv"},
		{http.MethodDelete, "/api/v1/inventario/tachos/1"},
		{http.MethodDelete, "/areas-verdes/v1/inventario/tachos/1"},

		{http.MethodGet, "/api/v1/inventario/bebederos"},
		{http.MethodGet, "/areas-verdes/v1/inventario/bebederos"},
		{http.MethodPost, "/api/v1/inventario/bebederos"},
		{http.MethodPost, "/areas-verdes/v1/inventario/bebederos"},
		{http.MethodPatch, "/api/v1/inventario/bebederos/1"},
		{http.MethodPatch, "/areas-verdes/v1/inventario/bebederos/1"},
		{http.MethodDelete, "/api/v1/inventario/bebederos/1"},
		{http.MethodDelete, "/areas-verdes/v1/inventario/bebederos/1"},

		{http.MethodGet, "/api/v1/inventario/puntos"},
		{http.MethodGet, "/areas-verdes/v1/inventario/puntos"},
		{http.MethodPost, "/api/v1/inventario/puntos"},
		{http.MethodPost, "/areas-verdes/v1/inventario/puntos"},
		{http.MethodPatch, "/api/v1/inventario/puntos/1"},
		{http.MethodPatch, "/areas-verdes/v1/inventario/puntos/1"},
		{http.MethodDelete, "/api/v1/inventario/puntos/1"},
		{http.MethodDelete, "/areas-verdes/v1/inventario/puntos/1"},
		{http.MethodPost, "/api/v1/inventario/formato/puntos"},
		{http.MethodPost, "/areas-verdes/v1/inventario/formato/puntos"},

		{http.MethodGet, "/api/v1/inventario/reservas"},
		{http.MethodGet, "/areas-verdes/v1/inventario/reservas"},
		{http.MethodPost, "/api/v1/inventario/reservas"},
		{http.MethodPost, "/areas-verdes/v1/inventario/reservas"},
		{http.MethodPatch, "/api/v1/inventario/reservas/1"},
		{http.MethodPatch, "/areas-verdes/v1/inventario/reservas/1"},
		{http.MethodDelete, "/api/v1/inventario/reservas/1"},
		{http.MethodDelete, "/areas-verdes/v1/inventario/reservas/1"},

		{http.MethodGet, "/api/v1/inventario/capas/fauna"},
		{http.MethodGet, "/areas-verdes/v1/inventario/capas/fauna"},
		{http.MethodPost, "/api/v1/inventario/capas/fauna"},
		{http.MethodPost, "/areas-verdes/v1/inventario/capas/fauna"},
		{http.MethodPatch, "/api/v1/inventario/capas/fauna/1"},
		{http.MethodPatch, "/areas-verdes/v1/inventario/capas/fauna/1"},
		{http.MethodGet, "/api/v1/inventario/export/fauna"},
		{http.MethodGet, "/areas-verdes/v1/inventario/export/fauna"},
		{http.MethodDelete, "/api/v1/inventario/capas/fauna/1"},
		{http.MethodDelete, "/areas-verdes/v1/inventario/capas/fauna/1"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasInventarioCampo_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, registrar=true
		{http.MethodGet, "/api/v1/inventario/tachos", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/tachos", "token-norte", `{"lat":-12.0,"lon":-77.0,"tipo":"General"}`, 201, ""},
		{http.MethodPatch, "/api/v1/inventario/tachos/1", "token-norte", `{"tipo":"Plástico"}`, 200, ""},
		{http.MethodGet, "/api/v1/inventario/tachos.csv", "token-norte", "", 200, ""},
		{http.MethodDelete, "/api/v1/inventario/tachos/1", "token-norte", "", 200, ""},

		{http.MethodGet, "/api/v1/inventario/bebederos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/bebederos", "token-norte", `{"lat":-12.0,"lon":-77.0}`, 201, ""},
		{http.MethodPatch, "/api/v1/inventario/bebederos/1", "token-norte", `{"operativo":true}`, 200, ""},
		{http.MethodDelete, "/api/v1/inventario/bebederos/1", "token-norte", "", 200, ""},

		{http.MethodGet, "/api/v1/inventario/puntos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/puntos", "token-norte", `{"title":"P1","lat":-12.0,"lon":-77.0}`, 201, ""},
		{http.MethodPatch, "/api/v1/inventario/puntos/1", "token-norte", `{"title":"P1-mod"}`, 200, ""},
		{http.MethodDelete, "/api/v1/inventario/puntos/1", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/formato/puntos", "token-norte", `{"csv":"Title,Latitude,Longitude\nP1,-12.0,-77.0"}`, 200, ""},

		{http.MethodGet, "/api/v1/inventario/reservas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/reservas", "token-norte", `{"id_espacio":"E1","fecha":"2026-09-01"}`, 201, ""},
		{http.MethodPatch, "/api/v1/inventario/reservas/1", "token-norte", `{"solicitante":"Modif"}`, 200, ""},
		{http.MethodDelete, "/api/v1/inventario/reservas/1", "token-norte", "", 200, ""},

		{http.MethodGet, "/api/v1/inventario/capas/fauna", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/capas/fauna", "token-norte", `{"nombre":"F1","lat":-12.0,"lon":-77.0}`, 201, ""},
		{http.MethodPatch, "/api/v1/inventario/capas/fauna/1", "token-norte", `{"nombre":"F1-mod"}`, 200, ""},
		{http.MethodGet, "/api/v1/inventario/export/fauna", "token-norte", "", 200, ""},
		{http.MethodDelete, "/api/v1/inventario/capas/fauna/1", "token-norte", "", 200, ""},

		// Coordinación: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/inventario/tachos", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/tachos", "token-coordinacion", `{"lat":-12.0,"lon":-77.0,"tipo":"General"}`, 201, ""},
		{http.MethodGet, "/api/v1/inventario/reservas", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/reservas", "token-coordinacion", `{"id_espacio":"E1","fecha":"2026-09-01"}`, 201, ""},

		// Admin: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/inventario/tachos", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/tachos", "token-admin", `{"lat":-12.0,"lon":-77.0,"tipo":"General"}`, 201, ""},
		{http.MethodGet, "/api/v1/inventario/capas/fauna", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/capas/fauna", "token-admin", `{"nombre":"F1","lat":-12.0,"lon":-77.0}`, 201, ""},

		// Jefatura: consultar=true, registrar=false
		{http.MethodGet, "/api/v1/inventario/tachos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/inventario/tachos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/inventario/bebederos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/inventario/puntos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/inventario/reservas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/inventario/capas/fauna", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/inventario/formato/puntos", "token-jefatura", `{"csv":"Title,Latitude,Longitude\nP1,-12.0,-77.0"}`, 200, ""}, // consultar=true permite formato/puntos

		// Jefatura writes: forbidden 403
		{http.MethodPost, "/api/v1/inventario/tachos", "token-jefatura", `{"lat":-12.0,"lon":-77.0}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/inventario/tachos", "token-jefatura", `{"lat":-12.0,"lon":-77.0}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/inventario/tachos/1", "token-jefatura", `{"tipo":"General"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodDelete, "/api/v1/inventario/tachos/1", "token-jefatura", "", 403, "su rol no tiene ese permiso"},

		{http.MethodPost, "/api/v1/inventario/bebederos", "token-jefatura", `{"lat":-12.0,"lon":-77.0}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/inventario/bebederos/1", "token-jefatura", `{"operativo":true}`, 403, "su rol no tiene ese permiso"},
		{http.MethodDelete, "/api/v1/inventario/bebederos/1", "token-jefatura", "", 403, "su rol no tiene ese permiso"},

		{http.MethodPost, "/api/v1/inventario/puntos", "token-jefatura", `{"title":"P1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/inventario/puntos/1", "token-jefatura", `{"title":"P1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodDelete, "/api/v1/inventario/puntos/1", "token-jefatura", "", 403, "su rol no tiene ese permiso"},

		{http.MethodPost, "/api/v1/inventario/reservas", "token-jefatura", `{"id_espacio":"E1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/inventario/reservas/1", "token-jefatura", `{"solicitante":"S"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodDelete, "/api/v1/inventario/reservas/1", "token-jefatura", "", 403, "su rol no tiene ese permiso"},

		{http.MethodPost, "/api/v1/inventario/capas/fauna", "token-jefatura", `{"nombre":"F1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/inventario/capas/fauna/1", "token-jefatura", `{"nombre":"F1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodDelete, "/api/v1/inventario/capas/fauna/1", "token-jefatura", "", 403, "su rol no tiene ese permiso"},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasOperacion_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/operacion/capataces"},
		{http.MethodGet, "/areas-verdes/v1/operacion/capataces"},
		{http.MethodGet, "/api/v1/operacion/actividades"},
		{http.MethodGet, "/areas-verdes/v1/operacion/actividades"},
		{http.MethodPost, "/api/v1/operacion/actividades"},
		{http.MethodPost, "/areas-verdes/v1/operacion/actividades"},
		{http.MethodPatch, "/api/v1/operacion/actividades/1/asignacion"},
		{http.MethodPatch, "/areas-verdes/v1/operacion/actividades/1/asignacion"},
		{http.MethodPatch, "/api/v1/operacion/actividades/1/estado"},
		{http.MethodPatch, "/areas-verdes/v1/operacion/actividades/1/estado"},
		{http.MethodPost, "/api/v1/operacion/actividades/1/archivar"},
		{http.MethodPost, "/areas-verdes/v1/operacion/actividades/1/archivar"},
		{http.MethodGet, "/api/v1/operacion/actividades/1/timeline"},
		{http.MethodGet, "/areas-verdes/v1/operacion/actividades/1/timeline"},
		{http.MethodPatch, "/api/v1/operacion/actividades/1/ficha"},
		{http.MethodPatch, "/areas-verdes/v1/operacion/actividades/1/ficha"},
		{http.MethodPost, "/api/v1/operacion/actividades/1/avances"},
		{http.MethodPost, "/areas-verdes/v1/operacion/actividades/1/avances"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasOperacion_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, registrar=true, validar=false
		{http.MethodGet, "/api/v1/operacion/capataces", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/operacion/capataces", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/operacion/actividades", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/operacion/actividades", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/timeline", "token-norte", "", 200, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/estado", "token-norte", `{"estado":"en_proceso"}`, 200, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/ficha", "token-norte", `{"comentario":"test"}`, 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/avances", "token-norte", `{"id":"22222222-2222-4222-8222-222222222222","fecha":"2026-09-28","area_feature_id":"AV-0001"}`, 201, ""},

		// Capataz no puede crear, asignar ni archivar (requieren validar)
		{http.MethodPost, "/api/v1/operacion/actividades", "token-norte", `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/operacion/actividades", "token-norte", `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/asignacion", "token-norte", `{"capataz_id":"cap-sur"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/archivar", "token-norte", `{"motivo":"duplicada"}`, 403, "su rol no tiene ese permiso"},

		// Coordinación: consultar=true, registrar=true, validar=true
		{http.MethodGet, "/api/v1/operacion/actividades", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades", "token-coordinacion", `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego"}`, 201, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/asignacion", "token-coordinacion", `{"capataz_id":"cap-sur"}`, 200, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/estado", "token-coordinacion", `{"estado":"en_proceso"}`, 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/archivar", "token-coordinacion", `{"motivo":"duplicada"}`, 200, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/ficha", "token-coordinacion", `{"comentario":"test"}`, 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/avances", "token-coordinacion", `{"id":"22222222-2222-4222-8222-222222222222","fecha":"2026-09-28","area_feature_id":"AV-0001"}`, 201, ""},

		// Admin: todos los permisos
		{http.MethodGet, "/api/v1/operacion/actividades", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades", "token-admin", `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego"}`, 201, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/asignacion", "token-admin", `{"capataz_id":"cap-sur"}`, 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/archivar", "token-admin", `{"motivo":"duplicada"}`, 200, ""},

		// Jefatura: consultar=true, validar=true, registrar=false
		{http.MethodGet, "/api/v1/operacion/actividades", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades", "token-jefatura", `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego"}`, 201, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/asignacion", "token-jefatura", `{"capataz_id":"cap-sur"}`, 200, ""},
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/estado", "token-jefatura", `{"estado":"en_proceso"}`, 200, ""},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/archivar", "token-jefatura", `{"motivo":"duplicada"}`, 200, ""},
		// Jefatura no tiene registrar:
		{http.MethodPatch, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/ficha", "token-jefatura", `{"comentario":"test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/avances", "token-jefatura", `{"id":"22222222-2222-4222-8222-222222222222","fecha":"2026-09-28","area_feature_id":"AV-0001"}`, 403, "su rol no tiene ese permiso"},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasAtencion_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/solicitudes"},
		{http.MethodGet, "/areas-verdes/v1/solicitudes"},
		{http.MethodPost, "/api/v1/solicitudes"},
		{http.MethodPost, "/areas-verdes/v1/solicitudes"},
		{http.MethodPatch, "/api/v1/solicitudes/1"},
		{http.MethodPatch, "/areas-verdes/v1/solicitudes/1"},

		{http.MethodGet, "/api/v1/ordenes"},
		{http.MethodGet, "/areas-verdes/v1/ordenes"},
		{http.MethodPost, "/api/v1/ordenes"},
		{http.MethodPost, "/areas-verdes/v1/ordenes"},
		{http.MethodPatch, "/api/v1/ordenes/1"},
		{http.MethodPatch, "/areas-verdes/v1/ordenes/1"},

		{http.MethodGet, "/api/v1/riego"},
		{http.MethodGet, "/areas-verdes/v1/riego"},
		{http.MethodPost, "/api/v1/riego"},
		{http.MethodPost, "/areas-verdes/v1/riego"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasAtencion_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// 1. Capataz (norte):
		// Solicitudes requiere permiso "solicitudes" (capataz no lo tiene)
		{http.MethodGet, "/api/v1/solicitudes", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodGet, "/areas-verdes/v1/solicitudes", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/solicitudes", "token-norte", `{"fuente":"interna","titulo":"Solicitud"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/solicitudes", "token-norte", `{"fuente":"interna","titulo":"Solicitud"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/solicitudes/1", "token-norte", `{"titulo":"Modif"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/solicitudes/1", "token-norte", `{"titulo":"Modif"}`, 403, "su rol no tiene ese permiso"},

		// Ordenes: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/ordenes", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/ordenes", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/ordenes", "token-norte", `{"actividad_id":"act-1","empresa":"Contratista"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/ordenes", "token-norte", `{"actividad_id":"act-1","empresa":"Contratista"}`, 201, ""},
		{http.MethodPatch, "/api/v1/ordenes/1", "token-norte", `{"estado":"ejecutada"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/ordenes/1", "token-norte", `{"estado":"ejecutada"}`, 200, ""},

		// Riego: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/riego", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/riego", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/riego", "token-norte", `{"fecha":"2026-10-01","turno":"manana"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/riego", "token-norte", `{"fecha":"2026-10-01","turno":"manana"}`, 201, ""},

		// 2. Coordinación:
		// Solicitudes: solicitudes=true
		{http.MethodGet, "/api/v1/solicitudes", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/solicitudes", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/solicitudes", "token-coordinacion", `{"fuente":"interna","titulo":"Solicitud"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/solicitudes", "token-coordinacion", `{"fuente":"interna","titulo":"Solicitud"}`, 201, ""},
		{http.MethodPatch, "/api/v1/solicitudes/1", "token-coordinacion", `{"titulo":"Modif"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/solicitudes/1", "token-coordinacion", `{"titulo":"Modif"}`, 200, ""},

		// Ordenes: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/ordenes", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/ordenes", "token-coordinacion", `{"actividad_id":"act-1","empresa":"Contratista"}`, 201, ""},
		{http.MethodPatch, "/api/v1/ordenes/1", "token-coordinacion", `{"estado":"ejecutada"}`, 200, ""},

		// Riego: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/riego", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/riego", "token-coordinacion", `{"fecha":"2026-10-01","turno":"manana"}`, 201, ""},

		// 3. Jefatura:
		// Solicitudes: solicitudes=true
		{http.MethodGet, "/api/v1/solicitudes", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/solicitudes", "token-jefatura", `{"fuente":"interna","titulo":"Solicitud"}`, 201, ""},
		{http.MethodPatch, "/api/v1/solicitudes/1", "token-jefatura", `{"titulo":"Modif"}`, 200, ""},

		// Ordenes: consultar=true, registrar=false (403 para POST/PATCH)
		{http.MethodGet, "/api/v1/ordenes", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/ordenes", "token-jefatura", `{"actividad_id":"act-1","empresa":"Contratista"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/ordenes", "token-jefatura", `{"actividad_id":"act-1","empresa":"Contratista"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/ordenes/1", "token-jefatura", `{"estado":"ejecutada"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/ordenes/1", "token-jefatura", `{"estado":"ejecutada"}`, 403, "su rol no tiene ese permiso"},

		// Riego: consultar=true, registrar=false (403 para POST)
		{http.MethodGet, "/api/v1/riego", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/riego", "token-jefatura", `{"fecha":"2026-10-01","turno":"manana"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/riego", "token-jefatura", `{"fecha":"2026-10-01","turno":"manana"}`, 403, "su rol no tiene ese permiso"},

		// 4. Admin: todos los permisos
		{http.MethodGet, "/api/v1/solicitudes", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/solicitudes", "token-admin", `{"fuente":"interna","titulo":"Solicitud"}`, 201, ""},
		{http.MethodPatch, "/api/v1/solicitudes/1", "token-admin", `{"titulo":"Modif"}`, 200, ""},
		{http.MethodGet, "/api/v1/ordenes", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/ordenes", "token-admin", `{"actividad_id":"act-1","empresa":"Contratista"}`, 201, ""},
		{http.MethodPatch, "/api/v1/ordenes/1", "token-admin", `{"estado":"ejecutada"}`, 200, ""},
		{http.MethodGet, "/api/v1/riego", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/riego", "token-admin", `{"fecha":"2026-10-01","turno":"manana"}`, 201, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasPodaVivero_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/podas"},
		{http.MethodGet, "/areas-verdes/v1/podas"},
		{http.MethodPost, "/api/v1/podas"},
		{http.MethodPost, "/areas-verdes/v1/podas"},
		{http.MethodPatch, "/api/v1/podas/1"},
		{http.MethodPatch, "/areas-verdes/v1/podas/1"},
		{http.MethodPost, "/api/v1/podas/1/archivar"},
		{http.MethodPost, "/areas-verdes/v1/podas/1/archivar"},

		{http.MethodGet, "/api/v1/vivero"},
		{http.MethodGet, "/areas-verdes/v1/vivero"},
		{http.MethodPost, "/api/v1/vivero"},
		{http.MethodPost, "/areas-verdes/v1/vivero"},
		{http.MethodPatch, "/api/v1/vivero/1"},
		{http.MethodPatch, "/areas-verdes/v1/vivero/1"},
		{http.MethodPost, "/api/v1/vivero/1/archivar"},
		{http.MethodPost, "/areas-verdes/v1/vivero/1/archivar"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasPodaVivero_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// 1. Capataz (norte):
		// Poda: consultar=true, registrar=true, validar=false (archivar -> 403)
		{http.MethodGet, "/api/v1/podas", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/podas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/podas", "token-norte", `{"codigo":"PO-001"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/podas", "token-norte", `{"codigo":"PO-001"}`, 201, ""},
		{http.MethodPatch, "/api/v1/podas/1", "token-norte", `{"prioridad":"alta"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/podas/1", "token-norte", `{"prioridad":"alta"}`, 200, ""},
		{http.MethodPost, "/api/v1/podas/1/archivar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/podas/1/archivar", "token-norte", "", 403, "su rol no tiene ese permiso"},

		// Vivero: consultar=true, registrar=true, validar=false (archivar -> 403)
		{http.MethodGet, "/api/v1/vivero", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/vivero", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/vivero", "token-norte", `{"area":"Fauna"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/vivero", "token-norte", `{"area":"Fauna"}`, 201, ""},
		{http.MethodPatch, "/api/v1/vivero/1", "token-norte", `{"area":"Fauna"}`, 200, ""},
		{http.MethodPatch, "/areas-verdes/v1/vivero/1", "token-norte", `{"area":"Fauna"}`, 200, ""},
		{http.MethodPost, "/api/v1/vivero/1/archivar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/vivero/1/archivar", "token-norte", "", 403, "su rol no tiene ese permiso"},

		// 2. Coordinación: tiene consultar, registrar, validar
		{http.MethodGet, "/api/v1/podas", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/podas", "token-coordinacion", `{"codigo":"PO-001"}`, 201, ""},
		{http.MethodPatch, "/api/v1/podas/1", "token-coordinacion", `{"prioridad":"alta"}`, 200, ""},
		{http.MethodPost, "/api/v1/podas/1/archivar", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/api/v1/vivero", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/vivero", "token-coordinacion", `{"area":"Fauna"}`, 201, ""},
		{http.MethodPatch, "/api/v1/vivero/1", "token-coordinacion", `{"area":"Fauna"}`, 200, ""},
		{http.MethodPost, "/api/v1/vivero/1/archivar", "token-coordinacion", "", 200, ""},

		// 3. Jefatura: consultar=true, validar=true, registrar=false (POST/PATCH -> 403)
		{http.MethodGet, "/api/v1/podas", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/podas", "token-jefatura", `{"codigo":"PO-001"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/podas", "token-jefatura", `{"codigo":"PO-001"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/podas/1", "token-jefatura", `{"prioridad":"alta"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/podas/1", "token-jefatura", `{"prioridad":"alta"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/podas/1/archivar", "token-jefatura", "", 200, ""},

		{http.MethodGet, "/api/v1/vivero", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/vivero", "token-jefatura", `{"area":"Fauna"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/vivero", "token-jefatura", `{"area":"Fauna"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/vivero/1", "token-jefatura", `{"area":"Fauna"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/vivero/1", "token-jefatura", `{"area":"Fauna"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/vivero/1/archivar", "token-jefatura", "", 200, ""},

		// 4. Admin: todos los permisos
		{http.MethodGet, "/api/v1/podas", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/podas", "token-admin", `{"codigo":"PO-001"}`, 201, ""},
		{http.MethodPatch, "/api/v1/podas/1", "token-admin", `{"prioridad":"alta"}`, 200, ""},
		{http.MethodPost, "/api/v1/podas/1/archivar", "token-admin", "", 200, ""},
		{http.MethodGet, "/api/v1/vivero", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/vivero", "token-admin", `{"area":"Fauna"}`, 201, ""},
		{http.MethodPatch, "/api/v1/vivero/1", "token-admin", `{"area":"Fauna"}`, 200, ""},
		{http.MethodPost, "/api/v1/vivero/1/archivar", "token-admin", "", 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasEvidencias_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/evidencias"},
		{http.MethodGet, "/areas-verdes/v1/evidencias"},
		{http.MethodPost, "/api/v1/evidencias"},
		{http.MethodPost, "/areas-verdes/v1/evidencias"},
		{http.MethodGet, "/api/v1/evidencias/1/archivo"},
		{http.MethodGet, "/areas-verdes/v1/evidencias/1/archivo"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasEvidencias_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	casos := []struct {
		metodo         string
		ruta           string
		token          string
		statusEsperado int
	}{
		// 1. Capataz: consultar=true -> GET 200
		{http.MethodGet, "/api/v1/evidencias", "token-norte", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias", "token-norte", 200},
		{http.MethodGet, "/api/v1/evidencias/1/archivo", "token-norte", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias/1/archivo", "token-norte", 200},

		// 2. Coordinación: consultar=true -> GET 200
		{http.MethodGet, "/api/v1/evidencias", "token-coordinacion", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias", "token-coordinacion", 200},
		{http.MethodGet, "/api/v1/evidencias/1/archivo", "token-coordinacion", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias/1/archivo", "token-coordinacion", 200},

		// 3. Jefatura: consultar=true -> GET 200
		{http.MethodGet, "/api/v1/evidencias", "token-jefatura", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias", "token-jefatura", 200},
		{http.MethodGet, "/api/v1/evidencias/1/archivo", "token-jefatura", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias/1/archivo", "token-jefatura", 200},

		// 4. Admin: consultar=true -> GET 200
		{http.MethodGet, "/api/v1/evidencias", "token-admin", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias", "token-admin", 200},
		{http.MethodGet, "/api/v1/evidencias/1/archivo", "token-admin", 200},
		{http.MethodGet, "/areas-verdes/v1/evidencias/1/archivo", "token-admin", 200},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(c.metodo, c.ruta, nil)
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
	}
}

func TestRutasReportesEIA_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/reportes/labores"},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores"},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo"},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasReportesEIA_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	casos := []struct {
		metodo         string
		ruta           string
		token          string
		statusEsperado int
		errorEsperado  string
	}{
		// 1. Capataz: reportes=false -> 403; consultar=true -> 200
		{http.MethodGet, "/api/v1/reportes/labores", "token-norte", 403, "su rol no tiene ese permiso"},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores", "token-norte", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo", "token-norte", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo", "token-norte", 200, ""},

		// 2. Coordinación: reportes=true -> 200; consultar=true -> 200
		{http.MethodGet, "/api/v1/reportes/labores", "token-coordinacion", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores", "token-coordinacion", 200, ""},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo", "token-coordinacion", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo", "token-coordinacion", 200, ""},

		// 3. Jefatura: reportes=true -> 200; consultar=true -> 200
		{http.MethodGet, "/api/v1/reportes/labores", "token-jefatura", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores", "token-jefatura", 200, ""},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo", "token-jefatura", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo", "token-jefatura", 200, ""},

		// 4. Admin: reportes=true -> 200; consultar=true -> 200
		{http.MethodGet, "/api/v1/reportes/labores", "token-admin", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/reportes/labores", "token-admin", 200, ""},
		{http.MethodPost, "/api/v1/ia/sugerir-tipo", "token-admin", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/ia/sugerir-tipo", "token-admin", 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var body io.Reader
		if c.metodo == http.MethodPost {
			body = strings.NewReader(`{"titulo":"Revisar aspersores del eje"}`)
		}
		req := httptest.NewRequest(c.metodo, c.ruta, body)
		if c.metodo == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, resp["error"])
			}
		}
	}
}

func TestRutasAuditoriaELotes_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodPost, "/api/v1/lotes"},
		{http.MethodPost, "/areas-verdes/v1/lotes"},
		{http.MethodPost, "/api/v1/lotes/1/revertir"},
		{http.MethodPost, "/areas-verdes/v1/lotes/1/revertir"},
		{http.MethodPost, "/api/v1/auditoria/ediciones"},
		{http.MethodPost, "/areas-verdes/v1/auditoria/ediciones"},
		{http.MethodGet, "/api/v1/auditoria/cambios"},
		{http.MethodGet, "/areas-verdes/v1/auditoria/cambios"},
		{http.MethodGet, "/api/v1/auditoria/timeline"},
		{http.MethodGet, "/areas-verdes/v1/auditoria/timeline"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasAuditoriaELotes_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	casos := []struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}{
		// 1. Capataz: validar=false -> 403; consultar=true -> 200
		{http.MethodPost, "/api/v1/lotes", "token-norte", `{"entidad":"catalogos","filas":[]}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/lotes", "token-norte", `{"entidad":"catalogos","filas":[]}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/lotes/1/revertir", "token-norte", `{"confirmar":false}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/lotes/1/revertir", "token-norte", `{"confirmar":false}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/auditoria/ediciones", "token-norte", `{"entidad":"catalogos","entidad_id":"1","despues":{}}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/auditoria/ediciones", "token-norte", `{"entidad":"catalogos","entidad_id":"1","despues":{}}`, 403, "su rol no tiene ese permiso"},
		{http.MethodGet, "/api/v1/auditoria/cambios", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/cambios", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-norte", "", 200, ""},

		// 2. Coordinación: validar=true, consultar=true
		{http.MethodPost, "/api/v1/lotes", "token-coordinacion", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes", "token-coordinacion", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/api/v1/lotes/1/revertir", "token-coordinacion", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes/1/revertir", "token-coordinacion", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/api/v1/auditoria/ediciones", "token-coordinacion", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/auditoria/ediciones", "token-coordinacion", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodGet, "/api/v1/auditoria/cambios", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/cambios", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/api/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-coordinacion", "", 200, ""},

		// 3. Jefatura: validar=true, consultar=true
		{http.MethodPost, "/api/v1/lotes", "token-jefatura", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes", "token-jefatura", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/api/v1/lotes/1/revertir", "token-jefatura", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes/1/revertir", "token-jefatura", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/api/v1/auditoria/ediciones", "token-jefatura", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/auditoria/ediciones", "token-jefatura", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodGet, "/api/v1/auditoria/cambios", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/cambios", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-jefatura", "", 200, ""},

		// 4. Admin: validar=true, consultar=true
		{http.MethodPost, "/api/v1/lotes", "token-admin", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes", "token-admin", `{"entidad":"catalogos","filas":[]}`, 201, ""},
		{http.MethodPost, "/api/v1/lotes/1/revertir", "token-admin", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/lotes/1/revertir", "token-admin", `{"confirmar":false}`, 200, ""},
		{http.MethodPost, "/api/v1/auditoria/ediciones", "token-admin", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodPost, "/areas-verdes/v1/auditoria/ediciones", "token-admin", `{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Nuevo"}}`, 200, ""},
		{http.MethodGet, "/api/v1/auditoria/cambios", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/cambios", "token-admin", "", 200, ""},
		{http.MethodGet, "/api/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/auditoria/timeline?entidad=catalogos&entidad_id=1", "token-admin", "", 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var body io.Reader
		if c.body != "" {
			body = strings.NewReader(c.body)
		}
		req := httptest.NewRequest(c.metodo, c.ruta, body)
		if c.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, resp["error"])
			}
		}
	}
}

func TestRutasImportaciones_SinAutenticacionRetorna401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/importaciones/entidades"},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades"},
		{http.MethodPost, "/api/v1/importaciones"},
		{http.MethodPost, "/areas-verdes/v1/importaciones"},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar"},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasImportaciones_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	casos := []struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}{
		// 1. Capataz: consultar=true -> 200 para GET; validar=false -> 403 para POST
		{http.MethodGet, "/api/v1/importaciones/entidades", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/importaciones", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/importaciones", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar", "token-norte", "", 403, "su rol no tiene ese permiso"},

		// 2. Coordinación: consultar=true -> 200; validar=true -> 200/201
		{http.MethodGet, "/api/v1/importaciones/entidades", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar", "token-coordinacion", "", 201, ""},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar", "token-coordinacion", "", 201, ""},

		// 3. Jefatura: consultar=true -> 200; validar=true -> 200/201
		{http.MethodGet, "/api/v1/importaciones/entidades", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar", "token-jefatura", "", 201, ""},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar", "token-jefatura", "", 201, ""},

		// 4. Admin: consultar=true -> 200; validar=true -> 200/201
		{http.MethodGet, "/api/v1/importaciones/entidades", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/importaciones/entidades", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/importaciones/1/confirmar", "token-admin", "", 201, ""},
		{http.MethodPost, "/areas-verdes/v1/importaciones/1/confirmar", "token-admin", "", 201, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var body io.Reader
		if c.body != "" {
			body = strings.NewReader(c.body)
		}
		req := httptest.NewRequest(c.metodo, c.ruta, body)
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var resp map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, resp["error"])
			}
		}
	}
}
