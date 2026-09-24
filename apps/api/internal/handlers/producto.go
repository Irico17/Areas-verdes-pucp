package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/atencion"
	"campusverde/api/internal/catalogos"
	"campusverde/api/internal/catastro"

	"github.com/gin-gonic/gin"
)

func usuarioEn(c *gin.Context) (accesos.Usuario, bool) {
	v, ok := c.Get("usuario")
	if !ok {
		return accesos.Usuario{}, false
	}
	u, ok := v.(accesos.Usuario)
	return u, ok
}

func sesionOCuerpo(c *gin.Context, rol, capataz string) (string, string) {
	u, ok := usuarioEn(c)
	if !ok {
		return rol, capataz
	}
	if u.Rol == "capataz" && u.CapatazID != "" {
		capataz = u.CapatazID
	}
	return u.Rol, capataz
}

func exige(c *gin.Context, accion string) (accesos.Usuario, bool) {
	u, ok := usuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return u, false
	}
	if !accesos.Permite(u.Rol, accion) {
		c.JSON(http.StatusForbidden, gin.H{"error": "su rol no tiene ese permiso"})
		return u, false
	}
	return u, true
}

func setCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accesos.Cookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

// Sesion expone el ingreso local. No es el SSO de la PUCP.
type Sesion struct {
	Store *accesos.Store
}

func (h Sesion) Entrar(c *gin.Context) {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	token, user, err := h.Store.Login(c.Request.Context(), body.Usuario, body.Clave)
	if err != nil {
		c.JSON(401, gin.H{"error": "usuario o clave incorrectos"})
		return
	}
	setCookie(c, token, 12*60*60)
	c.JSON(200, gin.H{"usuario": user})
}

func (h Sesion) Actual(c *gin.Context) {
	u, ok := usuarioEn(c)
	if !ok {
		c.JSON(401, gin.H{"error": "sin sesión"})
		return
	}
	c.JSON(200, gin.H{"usuario": u})
}

func (h Sesion) Salir(c *gin.Context) {
	if h.Store != nil {
		h.Store.Logout(c.Request.Context(), accesos.Token(c.Request))
	}
	setCookie(c, "", -1)
	c.JSON(200, gin.H{"ok": true})
}

func (h Sesion) Usuarios(c *gin.Context) {
	u, ok := usuarioEn(c)
	if !ok || u.Rol != "admin" {
		c.JSON(403, gin.H{"error": "solo administración ve las cuentas"})
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	rows, err := h.Store.Usuarios(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer las cuentas"})
		return
	}
	perms, err := h.Store.Permisos(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer los permisos"})
		return
	}
	c.JSON(200, gin.H{
		"usuarios": rows,
		"permisos": perms,
		"aviso":    "Cuentas locales de desarrollo. El SSO de la PUCP sigue pendiente de validación.",
	})
}

// Catalogo administra valores con baja lógica.
type Catalogo struct{ Store *catalogos.Store }

func (h Catalogo) List(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	rows, err := h.Store.List(c.Request.Context(), c.Query("clase"), c.Query("activos") == "1")
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer el catálogo"})
		return
	}
	c.JSON(200, gin.H{"items": rows, "clases": catalogos.Clases})
}

func (h Catalogo) Create(c *gin.Context) {
	if _, ok := exige(c, "catalogos"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body struct {
		Clase  string `json:"clase"`
		Codigo string `json:"codigo"`
		Nombre string `json:"nombre"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.Create(c.Request.Context(), body.Clase, body.Codigo, body.Nombre)
	if err != nil {
		if msg, ok := catalogos.EsEntrada(err); ok {
			c.JSON(400, gin.H{"error": msg})
			return
		}
		c.JSON(500, gin.H{"error": "no se pudo guardar el ítem"})
		return
	}
	c.JSON(201, item)
}

func (h Catalogo) Off(c *gin.Context) {
	if _, ok := exige(c, "catalogos"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(400, gin.H{"error": "id inválido"})
		return
	}
	if err := h.Store.Deactivate(c.Request.Context(), id); err != nil {
		if msg, ok := catalogos.EsEntrada(err); ok {
			c.JSON(404, gin.H{"error": msg})
			return
		}
		c.JSON(500, gin.H{"error": "no se pudo desactivar"})
		return
	}
	c.JSON(200, gin.H{"activo": false, "id": id})
}

func parseID(raw string, dest *int64) (int64, error) {
	var n int64
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, io.EOF
		}
		n = n*10 + int64(r-'0')
	}
	if n == 0 {
		return 0, io.EOF
	}
	*dest = n
	return n, nil
}

// Fichas edita metadatos de áreas. La geometría puede seguir vacía.
type Fichas struct{ Store *catastro.Store }

func (h Fichas) List(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	rows, err := h.Store.Fichas(c.Request.Context(), c.Query("q"))
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(200, gin.H{"areas": rows})
}

func (h Fichas) Patch(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body struct {
		Nombre     string `json:"nombre"`
		Uso        string `json:"uso"`
		RiegoAct   string `json:"riego_act"`
		Referencia string `json:"referencia"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.ActualizarFicha(c.Request.Context(), c.Param("id"), body.Nombre, body.Uso, body.RiegoAct, body.Referencia)
	if err != nil {
		if err == catastro.ErrFichaNoEncontrada {
			c.JSON(404, gin.H{"error": "área no encontrada"})
			return
		}
		c.JSON(400, gin.H{"error": "no se pudo guardar la ficha"})
		return
	}
	c.JSON(200, item)
}

func (h Fichas) Create(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body struct {
		FeatureID string `json:"feature_id"`
		Nombre    string `json:"nombre"`
		Uso       string `json:"uso"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearSinGeom(c.Request.Context(), body.FeatureID, body.Nombre, body.Uso)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear el área sin geometría"})
		return
	}
	c.JSON(201, item)
}

// Atencion agrupa solicitudes, órdenes, riego, evidencias y el reporte.
type Atencion struct {
	Store *atencion.Store
	Dir   string
}

func (h Atencion) ready(c *gin.Context) bool {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return false
	}
	return true
}

func (h Atencion) Solicitudes(c *gin.Context) {
	if _, ok := exige(c, "solicitudes"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarSolicitudes(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer las solicitudes"})
		return
	}
	c.JSON(200, gin.H{"solicitudes": rows})
}

func (h Atencion) CrearSolicitud(c *gin.Context) {
	if _, ok := exige(c, "solicitudes"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.SolicitudInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearSolicitud(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, item)
}

func (h Atencion) Ordenes(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarOrdenes(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer las órdenes"})
		return
	}
	c.JSON(200, gin.H{"ordenes": rows})
}

func (h Atencion) CrearOrden(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body struct {
		ID          string `json:"id"`
		ActividadID string `json:"actividad_id"`
		Empresa     string `json:"empresa"`
		Referencia  string `json:"referencia"`
		Frecuencia  string `json:"frecuencia"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearOrden(c.Request.Context(), body.ID, body.ActividadID, body.Empresa, body.Referencia, body.Frecuencia)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, item)
}

func (h Atencion) Riego(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarRiego(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer el riego"})
		return
	}
	c.JSON(200, gin.H{
		"aviso":     "Registro por sector, turno y equipo. La cobertura oficial no está definida.",
		"registros": rows,
	})
}

func (h Atencion) CrearRiego(c *gin.Context) {
	u, ok := exige(c, "registrar")
	if !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body struct {
		ID        string `json:"id"`
		Sector    string `json:"sector"`
		Turno     string `json:"turno"`
		CapatazID string `json:"capataz_id"`
		Fecha     string `json:"fecha"`
		Nota      string `json:"nota"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if u.Rol == "capataz" {
		body.CapatazID = u.CapatazID
	}
	if err := h.Store.CrearRiego(c.Request.Context(), body.ID, body.Sector, body.Turno, body.CapatazID, body.Fecha, body.Nota); err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, gin.H{"id": body.ID})
}

func (h Atencion) Evidencias(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarEvidencias(c.Request.Context(), c.Query("actividad_id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer las evidencias"})
		return
	}
	c.JSON(200, gin.H{"evidencias": rows})
}

func (h Atencion) SubirEvidencia(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	actividadID := c.PostForm("actividad_id")
	nota := c.PostForm("nota")
	file, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(400, gin.H{"error": "falta el archivo"})
		return
	}
	if file.Size > 8<<20 {
		c.JSON(400, gin.H{"error": "el archivo supera 8 MB"})
		return
	}
	mime := file.Header.Get("Content-Type")
	ext := extDe(mime, file.Filename)
	if ext == "" {
		c.JSON(400, gin.H{"error": "se admite jpg, png, webp o pdf"})
		return
	}
	id := c.PostForm("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "falta el id"})
		return
	}
	if err := os.MkdirAll(h.Dir, 0o755); err != nil {
		c.JSON(500, gin.H{"error": "no se pudo abrir la carpeta de evidencias"})
		return
	}
	dest := filepath.Join(h.Dir, id+ext)
	if err := c.SaveUploadedFile(file, dest); err != nil {
		c.JSON(500, gin.H{"error": "no se pudo guardar el archivo"})
		return
	}
	if err := h.Store.GuardarEvidencia(c.Request.Context(), id, actividadID, filepath.Base(file.Filename), mime, dest, nota, int(file.Size)); err != nil {
		_ = os.Remove(dest)
		writeAtencion(c, err)
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func (h Atencion) Archivo(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	ruta, mime, err := h.Store.RutaEvidencia(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeAtencion(c, err)
		return
	}
	clean := filepath.Clean(ruta)
	base := filepath.Clean(h.Dir)
	if base == "" || !strings.HasPrefix(clean, base+string(os.PathSeparator)) {
		c.JSON(404, gin.H{"error": "archivo no disponible"})
		return
	}
	c.Header("Content-Type", mime)
	c.File(clean)
}

func (h Atencion) Reporte(c *gin.Context) {
	if _, ok := exige(c, "reportes"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rep, err := h.Store.Reporte(c.Request.Context(), c.Query("estado"), c.Query("desde"), c.Query("hasta"))
	if err != nil {
		writeAtencion(c, err)
		return
	}
	switch c.Query("formato") {
	case "csv":
		c.Header("Content-Disposition", `attachment; filename="labores.csv"`)
		c.Data(200, "text/csv; charset=utf-8", []byte(atencion.CSV(rep.Filas)))
	case "xls":
		c.Header("Content-Disposition", `attachment; filename="labores.xls"`)
		c.Data(200, "application/vnd.ms-excel", []byte(atencion.ExcelXML(rep.Filas)))
	default:
		c.JSON(200, rep)
	}
}

func (h Atencion) Sugerir(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	var body struct {
		Titulo string `json:"titulo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	c.JSON(200, atencion.SugerirTipo(body.Titulo))
}

func extDe(mime, name string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return ".jpg"
	case strings.HasSuffix(lower, ".png"):
		return ".png"
	case strings.HasSuffix(lower, ".webp"):
		return ".webp"
	case strings.HasSuffix(lower, ".pdf"):
		return ".pdf"
	default:
		return ""
	}
}

func writeAtencion(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case err == catastro.ErrFichaNoEncontrada:
		c.JSON(404, gin.H{"error": "no encontrado"})
	default:
		writeOperacionErr(c, err)
	}
}
