/**
 * Etiquetas visibles de VerdePUCP.
 * Los códigos (rol, estado, id de capa, ruta) no se renombran aquí.
 * Fuente: docs/GLOSARIO-NOMENCLATURA.md.
 */

export const MODULO = {
  hoy: "Hoy",
  resumen: "Resumen",
  mapa: "Mapa",
  labores: "Actividades",
  laboresCapataz: "Mis actividades",
  registros: "Registros de campo",
  registrosCorto: "Registros",
  catastro: "Áreas y sectores",
  inventario: "Inventario",
  solicitudes: "Solicitudes y servicios",
  solicitudesCorto: "Solicitudes",
  reportes: "Reportes",
  catalogos: "Catálogos",
  importaciones: "Importar datos",
  admin: "Administración",
  ejemplares: "Ejemplares",
  bitacora: "Bitácora",
  historial: "Historial",
  configuracion: "Configuración",
  mas: "Más",
  operacion: "Operación",
  datos: "Datos",
} as const

export const EJEMPLAR = {
  titulo: "Ejemplares",
  lede: "Ficha de un individuo del inventario de flora. El código anterior queda consultable.",
  buscar: "Buscar por código, nombre común o N°",
  vacio: "Ningún ejemplar coincide con esa búsqueda.",
  elegir: "Elija un ejemplar para ver la ficha.",
  sinCodigo: "Sin código",
  sinNombre: "Sin nombre común",
  salud: "Salud",
  saludBueno: "Bueno",
  saludRegular: "Regular",
  saludDeteriorado: "Deteriorado",
  sinSalud: "Sin dato",
  especie: "Especie",
  sinEspecie: "Sin especie",
  lugar: "Lugar",
  sinLugar: "Sin lugar",
  sector: "Sector de capataz o cuartel",
  sinSector: "Sin sector de capataz ni cuartel",
  cuartel: "Cuartel histórico",
  coordenadas: "Coordenadas",
  lat: "Latitud",
  lon: "Longitud",
  codigo: "Código vigente",
  codigoNuevo: "Código nuevo",
  recodificar: "Recodificar",
  guardar: "Guardar ficha",
  guardando: "Guardando…",
  historial: "Historial de códigos",
  sinHistorial: "Este ejemplar no tiene códigos anteriores.",
  pasoA: "pasó a",
  soloConsulta: "Consulta. La corrección de la ficha corresponde a Ingeniería / Coordinación.",
  guardada: "Ficha guardada.",
  recodificado: "Código actualizado. El anterior quedó en el historial.",
  error: "No se pudo guardar la ficha.",
  errorLista: "No se pudieron leer los ejemplares.",
  errorCoord: "Escriba latitud y longitud juntas, dentro del campus.",
  errorCodigo: "El código nuevo no puede quedar vacío.",
  mas: "Cargar más",
} as const

export const CAPA = {
  areas: { label: "Áreas verdes", hint: "Catastro principal" },
  zonas: { label: "Sectores de capataz", hint: "Polígono de la cuadrilla, sin nombres reales" },
  jardinesReserva: { label: "Jardines de reserva", hint: "Capa auxiliar" },
  xerofitica: { label: "Xerofítica", hint: "Capa auxiliar" },
  edificios: { label: "Edificios", hint: "Solo en relieve" },
  vias: { label: "Vías", hint: "Referente lineal. Apagada hasta importar un GeoJSON." },
  cuarteles: { label: "Cuarteles (histórico)", hint: "Solo lectura, cuando exista el archivo." },
  puertas: { label: "Puertas y entradas", hint: "Entradas del campus" },
  playas: { label: "Playas de estacionamiento", hint: "Polígonos" },
  puntos: { label: "Puntos del campus", hint: "Título y pin, sin teléfono ni placeId" },
  fauna: { label: "Fauna", hint: "Avistamientos" },
  tachos: { label: "Tachos", hint: "Puntos de residuos" },
  bebederos: { label: "Bebederos", hint: "Tipo fuente, tipo llenador de botella, deterioro y baja" },
  flora: { label: "Flora", hint: "Ejemplares del relevamiento" },
  cafetos: { label: "Cafetos", hint: "Cafetos del relevamiento" },
  vereda: { label: "Vereda en riesgo", hint: "Polígono" },
  reservas: "Reservas de jardín",
} as const

/** Textos de la propiedad Uso del GeoJSON. Los dos minoritarios no se funden. */
export const USO_FUENTE = {
  institucional: "Uso Institucional",
  administrativo: "Áreas de uso administrativo",
  recreativo: "Áreas de uso recreativo/descanso",
  sostenible: "Áreas de manejo sostenible y reducción de consumo de agua",
  deportivas: "Áreas deportivas y recreación activa",
  conservacion: "Áreas de conservación",
  "sin-uso": "Sin uso",
} as const

export const SECTOR_CAPATAZ = {
  "cua-valeria": "Sector de capataz — Valeria Quispe (ficticio)",
  "cua-mateo": "Sector de capataz — Mateo Salazar (ficticio)",
  "cua-renato": "Sector de capataz — Renato Cárdenas (ficticio)",
  "campo-deportivo": "Campo deportivo",
  "bosque-humedo": "Bosque húmedo",
  "sin-sector": "Sin sector de capataz",
} as const

export const ACTIVIDAD = {
  titulo: "Actividades",
  lede: "Puntos abiertos. El color es el estado y el icono, la clase de actividad.",
  filtros: "Filtros",
  cualquierEstado: "Cualquier estado",
  cualquierTipo: "Cualquier tipo",
  cualquierSector: "Cualquier sector de capataz",
  cualquierEjecutor: "Cualquier ejecutor",
  cualquierOrigen: "Cualquier origen",
  cualquierRiesgo: "Cualquier nivel de riesgo",
  desde: "Desde",
  hasta: "Hasta",
  historico: "Incluir cerradas",
  soloCuadrilla: "Solo ve las actividades de su cuadrilla.",
  cuadrilla: "Cuadrilla",
  todas: "Todas las cuadrillas",
  sinAsignar: "Sin asignar",
  sinCuadrilla: "Sin cuadrilla",
  tipo: "Tipo de actividad",
  todosTipos: "Todos",
  marcar: "+ Nueva actividad",
  cancelarMarca: "Cancelar marca",
  ubicar: "Toque el mapa donde está el trabajo.",
  quienEjecuta: "Quién ejecuta",
  personalPropio: "Personal propio",
  servicioTercerizado: "Servicio tercerizado",
  regla: "Regla local, en proceso. No sale de esta cuadrilla. Confirme el tipo de actividad antes de guardar.",
  sugerir: "Sugerir tipo",
  detalle: "Detalle",
  tituloCampo: "Título",
  crear: "Crear actividad",
  guardando: "Guardando…",
  reintentar: "Reintentar envío",
  estado: "Estado",
  guardarEstado: "Cambiar estado",
  foto: "Foto",
  masAcciones: "Más acciones",
  reasignarA: "Reasignar a",
  motivo: "Motivo de archivo",
  elegir: "Elegir",
  reasignar: "Reasignar",
  archivar: "Archivar",
  cancelar: "Cancelar",
  faltaMotivo: "Elija el motivo de archivo.",
  que: "Qué",
  quien: "Quién",
  confirmarArchivo: "Confirmar archivo",
  bitacora: "Bitácora",
  ficha: "Ficha de la actividad",
  clase: "Clase de actividad",
  fechaSolicitud: "Fecha de solicitud",
  fechaAtencion: "Fecha de atención",
  lugar: "Lugar",
  lugarPlaceholder: "Si no hay pin, el lugar ubica la actividad",
  origen: "Origen",
  codigoExterno: "Código externo",
  codigoExternoHint: "Solo si ya viene. No se inventa.",
  unidad: "Unidad solicitante",
  riesgo: "Nivel de riesgo",
  fechaProgramada: "Fecha programada",
  cantidad: "Cantidad",
  personal: "Personal de la actividad",
  personalLede: "Nombres ficticios. No son cuentas.",
  elegirClase: "Elegir clase",
  elegirTipo: "Elegir tipo",
  elegirRiesgo: "Elegir nivel",
  elegirOrigen: "Elegir origen",
  sinTipos: "Esta clase no tiene tipos en el catálogo.",
  faltaClase: "Elija la clase de actividad.",
  faltaTipo: "Elija el tipo de actividad.",
  clasificacion: "Clasificación",
  pedido: "Datos del pedido",
  zonaSupervision: "Zona de supervisión",
  sinZona: "Sin zona de supervisión",
  punto: "Punto",
  comentario: "Comentario",
  guardarFicha: "Guardar ficha",
  creePrimero: "Cree la actividad en el mapa para poder guardar la ficha.",
  sinTitulo: "Actividad",
  creada: "Actividad creada. Marque el siguiente punto.",
  enCola: "Sin señal. La actividad quedó guardada en este teléfono y se enviará sola.",
  archivada: "Actividad archivada. Ya no aparece en el mapa abierto.",
  noCierra: "El rol capataz no puede cerrar ni cancelar una actividad.",
  duplicada: "Una actividad pendiente ya existía con otro contenido y se descartó.",
  noLeer: "No se pudieron leer las actividades",
  vacio: "No hay actividades con este filtro. Quite un filtro o ubique una nueva en el mapa.",
  colaLocal: (n: number) => `${n} por enviar.`,
  archivarPregunta: (titulo: string) => `¿Archivar "${titulo}"? Deja de verse en el mapa; no se borra.`,
  quitarFiltros: "Quitar filtros",
  avance: "Avance",
  avanceLede: "Queda en la actividad que ya es de esta cuadrilla.",
  guardarAvance: "Guardar avance",
  avanceGuardado: "Avance guardado.",
  avanceError: "No se pudo guardar el avance.",
  faltaLugarAvance: "Indique un área o un ejemplar.",
  area: "Área",
  areaPlaceholder: "Área ya registrada",
  ejemplar: "Ejemplar",
  ejemplarPlaceholder: "Código de ejemplar",
  fechaAvance: "Fecha",
  servicioMarca: "servicio tercerizado",
  propioMarca: "personal propio",
} as const

export const HITO = {
  creada: "Creada",
  asignada: "Asignada",
  reasignada: "Reasignada",
  estado: "Estado",
  cancelada: "Cancelada",
  archivada: "Archivada",
  evidencia: "Evidencia",
  inicio: "Inicio",
  supervision: "Supervisión",
  derivacion: "Derivación",
  observacion: "Observación",
  conformidad: "Conformidad",
  avance: "Avance",
} as const

export function etiquetaHito(tipo: string): string {
  if (Object.prototype.hasOwnProperty.call(HITO, tipo)) {
    return HITO[tipo as keyof typeof HITO]
  }
  return tipo
}

export const BITACORA = {
  titulo: "Bitácora",
  lede: "Quién hizo qué, a qué hora, y la evidencia que quedó bajo ese hito.",
  vacio: "Esta actividad todavía no tiene hitos.",
  sinActividad: "Elija una actividad para leer su cadena.",
  anterior: "Cuadrilla anterior",
  nueva: "Cuadrilla nueva",
  sinCuadrilla: "Sin cuadrilla",
  quien: "Registró",
  texto: "Texto del hito",
  tipo: "Tipo de hito",
  anotar: "Anotar hito",
  anotando: "Anotando…",
  adjuntar: "Adjuntar evidencia a este hito",
  subiendo: "Subiendo…",
  evidenciaDe: "Evidencia del hito",
  elegirTipo: "Elegir tipo",
  error: "No se pudo actualizar la bitácora.",
  cargando: "Leyendo la cadena…",
  actividades: "Actividades abiertas",
} as const

export const CLASE_ACTIVIDAD = {
  habilitacion: "Habilitación de jardines",
  rehabilitacion: "Rehabilitación y rediseño de jardines",
  mantenimiento: "Mantenimiento de jardines",
  poda: "Poda",
  propagacion: "Propagación y plantación",
  riego: "Riego",
  residuos: "Manejo de residuos vegetales",
  fitosanitario: "Manejo fitosanitario",
  inspeccion_monitoreo: "Inspección y monitoreo",
} as const

export function etiquetaClase(clase: string): string {
  const tabla = CLASE_ACTIVIDAD as Record<string, string>
  return tabla[clase] ?? ""
}

export const CUADRILLA_DEMO = {
  "cap-norte": "Cuadrilla Norte",
  "cap-sur": "Cuadrilla Sur",
  "cap-riego": "Cuadrilla Riego",
} as const

export function etiquetaCuadrilla(id: string, nombreApi: string): string {
  const tabla = CUADRILLA_DEMO as Record<string, string>
  return tabla[id] ?? nombreApi
}

export const VISTA_ACTIVIDAD = {
  mapa: "Mapa",
  lista: "Lista",
  grupo: "Vista de actividades",
  sinRed: "Sin red. La lista queda a la vista porque el mapa no puede cargar.",
  sinTeselas: "Sin teselas. La lista queda a la vista porque el plano no cargó.",
  detalle: "Detalle de la actividad",
  vacio: "No hay actividades con este filtro.",
} as const

export const MAPA = {
  capas: "Capas",
  mostrarCapas: "Mostrar capas",
  ocultarPanel: "Ocultar capas",
  cerrarPanel: "Cerrar capas",
  catastro: "Catastro",
  mostrarCatastro: "Mostrar catastro",
  colorear: "Colorear el catastro por",
  uso: "Uso del área",
  sector: "Sector de capataz",
  leyendaUso: "Leyenda por uso del área",
  leyendaSector: "Leyenda por sector de capataz",
  cuadrillasFicticias: "Cuadrillas con nombre ficticio. La asignación real no se publica.",
  inventario: "Inventario",
  inventarioLede: "Capas opcionales. Apagadas hasta que se necesiten.",
  sinDetalle: "Este elemento no tiene datos para mostrar.",
  leyendo: "Leyendo catastro…",
} as const

export const RIEGO = {
  titulo: "Riego",
  lede: "Turno por sector de capataz y cuadrilla. El porcentaje no es la fórmula oficial.",
  cobertura: (n: number) => `Cobertura provisional: ${n} %. Pendiente de validar con la jefatura de sección.`,
  turnos: (n: number) => (n === 1 ? "1 turno registrado." : `${n} turnos registrados.`),
  zona: "Zona de supervisión",
  sector: "Sector de capataz",
  elegirSector: "Elija un sector de capataz",
  sinSectores: "No hay sectores de capataz activos.",
  turno: "Turno",
  fecha: "Fecha",
  nota: "Nota",
  registrar: "Registrar turno",
  enCola: "Sin señal. El turno quedó guardado en este teléfono y se enviará solo.",
  sinCuadrilla: "sin cuadrilla",
  vacio: "Todavía no hay turnos registrados.",
  manana: "Mañana",
  tarde: "Tarde",
  errorLeer: "No se pudo leer el riego",
  errorFecha: "Corrija la fecha antes de registrar.",
  errorRegistrar: "No se pudo registrar",
  errorSectores: "No se pudieron leer los sectores de capataz",
} as const

export const SOLICITUD = {
  titulo: "Solicitudes",
  lede: "Captura manual. El código externo se conserva si viene de Centuria u OSG; el sistema no lo inventa.",
  unaOrden: "orden registrada",
  variasOrdenes: "órdenes registradas",
  sinCantidad: "sin dato",
  tituloCampo: "Título",
  codigo: "Código externo",
  codigoPlaceholder: "Opcional",
  prioridad: "Prioridad",
  prioridadBaja: "Baja",
  prioridadMedia: "Media",
  prioridadAlta: "Alta",
  estado: "Estado",
  elegirEstado: "Por iniciar",
  ubicacion: "Ubicación",
  ubicacionHint: "Elija un lugar del catálogo o indique un punto dentro del campus. Puede usar los dos.",
  puntoJunto: "La latitud y la longitud van juntas.",
  numeroInvalido: "Escriba un número entero en las cantidades, y un decimal en el punto.",
  latitud: "Latitud",
  longitud: "Longitud",
  pedida: "Cantidad solicitada",
  ejecutada: "Cantidad ejecutada",
  registrar: "Registrar solicitud",
  vacio: "No hay solicitudes registradas.",
  errorLectura: "No se pudieron leer las solicitudes",
  errorCrear: "No se pudo crear",
  errorOrden: "No se pudo crear la orden",
  lugarCargado: "Lugar (dato ya cargado)",
  ordenes: "Órdenes de servicio",
  empresa: "Empresa",
  elegirEmpresa: "Elegir empresa",
  empresaCargada: "Empresa (dato ya cargado)",
  referencia: "Referencia de contratación",
  frecuencia: "Frecuencia",
  elegirFrecuencia: "Elegir frecuencia",
  frecuenciaCargada: "Frecuencia (dato ya cargado)",
  conformidad: "Conformidad",
  conformidadHint: "La conformidad queda en la orden. No cierra la solicitud.",
  registrarOrden: "Registrar orden",
  evidencias: "Evidencias de la orden",
  sinEvidencias: "Esta orden no tiene evidencias.",
  vincula: "Solo se vinculan a una actividad de servicio tercerizado. Sin orden de servicio, esa actividad no se cierra.",
  vacioOrden: "Todavía no hay órdenes. Una actividad de servicio tercerizado no se cierra hasta que exista una.",
  elegir: "Elija una actividad en el mapa antes de crear la orden.",
  seleccionada: (id: string) => `Actividad seleccionada ${id}`,
  ninguna: "Ninguna actividad seleccionada.",
  metricas: "Métricas del servicio tercerizado: definición pendiente.",
  origen: "Origen",
} as const

export const REPORTE = {
  lede: "Reporte básico de actividades, filtrable por zona de supervisión, cuadrilla, origen y fechas. Los conteos no son indicadores oficiales.",
  zona: "Zona de supervisión",
  vacio: "No hay actividades en ese rango.",
  metricas: "Métricas del servicio tercerizado",
  notaJefatura: "Jefatura de sección no ha acordado la fórmula. No hay portal del servicio tercerizado.",
} as const

export const EVIDENCIA = {
  titulo: "Evidencia",
  aria: "Evidencias de la actividad",
  fotos: "Fotos de la actividad",
  sinFotos: "Esta actividad no tiene fotos.",
  esperaServidor: "La foto se guarda cuando la actividad ya está en el servidor.",
  enNavegador: "Guardado en este teléfono.",
  enNavegadorSinPunto: "Guardado en este teléfono, sin ubicación.",
  sesion: "La sesión venció. La foto sigue en este teléfono.",
  sinPermiso: "Sin permiso para esta actividad. La foto sigue en este teléfono.",
  noSincronizada: "La actividad aún no está sincronizada. La foto sigue en este teléfono.",
} as const

export const PODA = {
  lede: "Registro de poda sobre un ejemplar. El código externo solo se conserva si ya viene como OSG.",
  enCola: "Sin señal. La poda quedó guardada en este teléfono y se enviará sola.",
  guardada: "Poda guardada. No se creó un código OSG.",
  error: "No se pudo guardar la poda.",
} as const

export const VIVERO = {
  enCola: "Sin señal. El registro de vivero quedó guardado en este teléfono y se enviará solo.",
  guardado: "Registro de vivero guardado.",
  error: "No se pudo guardar el registro de vivero.",
} as const

export const COLA = {
  marca: "por enviar",
  local: (n: number) => `${n} por enviar.`,
  reintentar: "Reintentar envío",
  conflicto: "Un registro pendiente ya existía con otro contenido y se descartó.",
  avance: "Sin señal. El avance quedó guardado en este teléfono y se enviará solo.",
  ficha: "Sin señal. La ficha quedó guardada en este teléfono y se enviará sola.",
  fichaGuardada: "Ficha guardada. El pin del mapa no cambia.",
  fichaError: "No se pudo guardar la ficha.",
  sinConexion: "Sin conexión. Intente de nuevo.",
} as const

export const FALLO = {
  sinRed: "Sin conexión. Intente de nuevo.",
  poda: "Su rol no puede registrar podas. Pida a Coordinación que lo haga.",
  vivero: "Su rol no puede registrar vivero. Pida a Coordinación que lo haga.",
  inventario: "Su rol no puede registrar inventario. Pida a Coordinación que lo haga.",
  ficha: "Su rol no puede guardar la ficha. Pida a Coordinación que lo haga.",
  riego: "Su rol no puede registrar riego. Pida a Coordinación que lo haga.",
} as const

export const DEMO = {
  aviso: "Datos de demostración. Esta fuente sigue pendiente de validar.",
} as const

export const HOY = {
  actividades: "Mis actividades de hoy",
  enLinea: "En línea",
  sinConexion: (n: number) => `Sin conexión · ${n} por enviar`,
  porIniciar: "Por iniciar",
  enProceso: "En proceso",
  bloqueadas: "Bloqueadas",
  bloqueadaNota: "Bloqueada se conserva como estado provisional hasta que el cliente decida.",
  iniciar: "Iniciar",
  terminar: "Terminar y Foto",
  verMotivo: "Ver motivo",
  lista: "Lista",
  mapa: "Mapa",
  riego: "Registrar riego",
  poda: "Registrar poda",
  vivero: "Registrar vivero",
  mas: "Poda y vivero",
  vacio: "No hay actividades abiertas en su cuadrilla. Cuando Coordinación asigne una, aparecerá aquí.",
  forzada: "Sin conexión. El mapa queda oculto y se muestra la lista.",
  bitacora: "Bitácora",
  ejemplares: "Ejemplares",
} as const

export const RESUMEN = {
  titulo: "Resumen",
  lede: "Lo que pide una decisión hoy.",
  bloqueadas: "Bloqueadas",
  sinCuadrilla: "Sin cuadrilla asignada",
  tercerizadas: "Tercerizadas sin servicio",
  solicitudes: "Solicitudes sin actividad",
  sincronizar: "Pendientes de sincronizar",
  riego: "Riego de hoy",
  verTodas: "Ver todas",
  vacio: "Nada en este bloque.",
  nueva: "+ Nueva actividad",
  solicitud: "Registrar solicitud",
  operacion: "Estado de la operación",
  verReporte: "Ver reporte",
  excel: "Descargar Excel",
  indicadores: "Indicadores pendientes de validar",
  cuentas: "Cuentas por rol",
  cambios: "Últimos cambios",
  importar: "Importar datos",
  provisional: "Bloqueada sigue como estado provisional.",
} as const

export const REGISTRO_CAMPO = {
  titulo: "Registros de campo",
  lede: "La lista va primero. El formulario se abre con Nuevo registro.",
  nuevo: "Nuevo registro",
  cerrar: "Cerrar formulario",
  riego: "Riego",
  poda: "Poda",
  vivero: "Vivero",
  vacioRiego: "Todavía no hay turnos. Use Nuevo registro para anotar el de hoy.",
  vacioPoda: "Todavía no hay podas. Use Nuevo registro para anotar una.",
  vacioVivero: "Todavía no hay registros de vivero. Use Nuevo registro para anotar uno.",
} as const

export const IMPORTAR_UI = {
  titulo: "Importar datos",
  lede: "Cuatro pasos. Nada se escribe hasta confirmar.",
  paso1: "1 Elegir qué importar",
  paso2: "2 Subir archivo",
  paso3: "3 Revisar",
  paso4: "4 Confirmar",
  revisar: "Revisar archivo",
  confirmar: "Confirmar importación",
  revertir: "Revertir esta importación",
  territorio: "Territorio",
  operacion: "Operación",
  arbolado: "Arbolado",
  mobiliario: "Mobiliario y puntos",
  faltaArchivo: "Elija un archivo para poder revisarlo.",
  faltaRevision: "Revise el archivo antes de confirmar.",
  region: "Revisar archivo",
} as const

export const CUENTA = {
  menu: "Cuenta",
  salir: "Salir",
  ayuda: "¿Cómo se usa?",
  cerrarAyuda: "Cerrar ayuda",
  omitir: "Omitir",
  siguiente: "Siguiente",
  listo: "Listo",
} as const

export const AYUDA = {
  sector: "Polígono operativo de una cuadrilla. No es la zona de supervisión Z1–Z4.",
  bloqueada: "Estado provisional. El cliente todavía no decidió si se conserva.",
  cobertura: "El porcentaje no es la fórmula oficial. Sigue pendiente de validar.",
} as const

export const ADMIN = {
  titulo: "Administración",
} as const

export const CLASE_CATALOGO: Record<string, string> = {
  tipo_actividad: "Tipos de actividad",
  estado: "Estados",
  prioridad: "Prioridades",
  lugar: "Lugares",
  especie: "Especies",
  motivo_archivo: "Motivos de archivo",
  turno: "Turnos",
  fuente: "Origen de la solicitud",
  estado_solicitud: "Estados de la solicitud",
  empresa: "Empresas",
  clase_actividad: "Clases de actividad",
  plaga: "Plagas",
  producto_fitosanitario: "Productos fitosanitarios",
  frecuencia: "Frecuencias",
  sede: "Sedes",
  cuartel: "Cuarteles (histórico)",
  sector_capataz: "Sectores de capataz",
  nivel_riesgo: "Niveles de riesgo",
  subtipo_actividad: "Tipos de segundo nivel",
}

export const IMPORTACION: Record<string, string> = {
  areas_verdes: "Áreas verdes",
  zonas_supervision: "Zonas de supervisión",
  poligonos_cuadrilla: "Polígonos de cuadrilla",
  cuadrillas: "Cuadrillas",
  lugares: "Lugares",
  ejemplares: "Ejemplares",
  palmeras: "Palmeras",
  cafetos: "Cafetos",
  catalogo_actividades: "Catálogo de actividades",
  labores: "Actividades",
  poda: "Poda",
  vivero: "Vivero",
  tachos: "Tachos",
  bebederos: "Bebederos",
  fauna: "Fauna",
  puertas: "Puertas y entradas",
  playas: "Playas de estacionamiento",
  vereda: "Vereda en riesgo",
  xerofitica: "Xerofítica",
  jardines_reserva: "Jardines de reserva",
  reservas: "Reservas de jardín",
  puntos_pucp: "Puntos del campus",
  sectores_capataz: "Sectores de capataz",
  vias: "Vías",
}

export const CONTEO_TACHO: Record<string, string> = {
  no_aprovechables: "No aprovechables",
  papel_carton: "Papel y cartón",
  plastico: "Plástico",
  vidrio: "Vidrio",
  pilas: "Pilas",
  peligrosos: "Peligrosos",
  raee: "RAEE",
  metales: "Metales",
  aniquem: "Aniquem",
  intermedios_plastico: "Intermedios plástico",
  intermedios_metal: "Intermedios metal",
}

export const SUBTIPO_BEBEDERO: Record<string, string> = {
  fuente: "Tipo fuente",
  llenador: "Tipo llenador de botella",
  nuevo: "Nuevo",
  deterioro: "Deterioro",
  baja: "Baja",
}

export const TIPO_RESPALDO = {
  incidencia: "Novedad de campo",
} as const

export const INVENTARIO_PESTANAS = {
  tachos: "Tachos",
  bebederos: "Bebederos",
  puntos: "Puntos del campus",
  otras: "Otras capas",
  reservas: "Reservas de jardín",
} as const

export const ZONIFICACION = {
  titulo: "Sectores de capataz",
  lede: "El color del mapa sale de este catálogo. Un sector se desactiva y la fila se conserva.",
  codigo: "Código",
  nombre: "Nombre",
  color: "Color en el mapa",
  alta: "Dar de alta",
  desactivar: "Desactivar",
  vacio: "Todavía no hay sectores de capataz activos.",
  importarSectores: "Importar sectores",
  importarVias: "Importar vías",
  vias: "Vías",
  viasLede: "La capa nace apagada y vacía. Solo entra un GeoJSON de líneas.",
  cuarteles: "Cuarteles (histórico)",
  sinCuarteles: "sin archivo de cuarteles",
  lugar: "Lugar",
  lugarCargado: "Lugar (dato ya cargado)",
  elegirLugar: "Elegir un lugar del catálogo",
  sinLugar: "Sin lugar",
  edificio: "Edificio referente",
  elegirEdificio: "Elegir el edificio por su id",
  sinEdificio: "Sin edificio",
  guardarReferente: "Guardar referente",
  duplicado: "Ese código ya está en el catálogo.",
  guardado: "Sector de capataz guardado.",
  inactivo: "Inactivo",
} as const

export const CATASTRO = {
  zonas: "Zonas de supervisión",
  nuevaZona: "Nueva zona de supervisión",
  zonaGuardada: "Zona de supervisión guardada.",
} as const

const NOMBRES_ZONA = ["", "Zona 1", "Zona 2", "Zona 3", "Zona 4"]

export function etiquetaZonaSupervision(codigo: string): string {
  const coincidencia = /^Z([1-4])$/i.exec(codigo.trim())
  if (!coincidencia) return codigo
  return NOMBRES_ZONA[Number(coincidencia[1])] ?? codigo
}

export function marcaEjecutor(codigo: string | undefined): string {
  if (codigo === "tercerizada") return ACTIVIDAD.servicioMarca
  if (codigo === "propia") return ACTIVIDAD.propioMarca
  return ""
}

export function resumenCatastro(areas: number, sectores: number, actividades: number): string {
  return `${areas} áreas · ${sectores} sectores de capataz · ${actividades} actividades`
}

export function conteoCatastro(areas: number | string, sectores: number | string): string {
  return `${areas} áreas · ${sectores} sectores de capataz`
}
