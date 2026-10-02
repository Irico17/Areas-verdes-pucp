/**
 * Etiquetas visibles de VerdePUCP.
 * Los códigos (rol, estado, id de capa, ruta) no se renombran aquí.
 * Fuente: docs/GLOSARIO-NOMENCLATURA.md.
 */

export const MODULO = {
  mapa: "Mapa",
  labores: "Actividades",
  catastro: "Catastro",
  inventario: "Inventario",
  solicitudes: "Solicitudes",
  reportes: "Reportes",
  catalogos: "Catálogos",
  importaciones: "Importar",
  admin: "Administración",
} as const

export const CAPA = {
  areas: { label: "Áreas verdes", hint: "Catastro principal" },
  zonas: { label: "Sectores de capataz", hint: "Polígono de la cuadrilla, sin nombres reales" },
  jardinesReserva: { label: "Jardines de reserva", hint: "Capa auxiliar" },
  xerofitica: { label: "Xerofítica", hint: "Capa auxiliar" },
  edificios: { label: "Edificios", hint: "Solo en relieve" },
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
  lede: "Puntos abiertos. El color es el estado y la letra, el tipo.",
  soloCuadrilla: "Solo ve las actividades de su cuadrilla.",
  cuadrilla: "Cuadrilla",
  todas: "Todas las cuadrillas",
  sinAsignar: "Sin asignar",
  sinCuadrilla: "Sin cuadrilla",
  tipo: "Tipo de actividad",
  todosTipos: "Todos",
  marcar: "Marcar actividad",
  cancelarMarca: "Cancelar marca",
  ubicar: "Haga clic en el mapa para ubicar la actividad.",
  quienEjecuta: "Quién ejecuta",
  personalPropio: "Personal propio",
  servicioTercerizado: "Servicio tercerizado",
  regla: "Regla local, en proceso. No sale de esta cuadrilla. Confirme el tipo de actividad antes de guardar.",
  sugerir: "Sugerir tipo",
  detalle: "Detalle",
  tituloCampo: "Título",
  crear: "Crear actividad",
  guardando: "Guardando…",
  vacio: "No hay actividades con este filtro.",
  reintentar: "Reintentar envío",
  estado: "Estado",
  guardarEstado: "Guardar estado",
  reasignarA: "Reasignar a",
  motivo: "Motivo de archivo",
  elegir: "Elegir",
  reasignar: "Reasignar",
  archivar: "Archivar",
  confirmarArchivo: "Confirmar archivo",
  bitacora: "Bitácora",
  ficha: "Ficha de la actividad",
  clase: "Clase de actividad",
  fechaSolicitud: "Fecha de solicitud",
  fechaAtencion: "Fecha de atención",
  lugar: "Lugar",
  lugarPlaceholder: "Si no hay pin, el lugar ubica la actividad",
  comentario: "Comentario",
  guardarFicha: "Guardar ficha",
  creePrimero: "Cree la actividad en el mapa para poder guardar la ficha.",
  sinTitulo: "Actividad",
  creada: "Actividad creada. Marque el siguiente punto.",
  enCola: "Sin conexión: la actividad quedó en la cola de este navegador.",
  archivada: "Actividad archivada. Ya no aparece en el mapa abierto.",
  noCierra: "El rol capataz no puede cerrar ni cancelar una actividad.",
  duplicada: "Una actividad en cola ya existía con otro contenido y se descartó.",
  noLeer: "No se pudieron leer las actividades",
  colaLocal: (n: number) => `${n} en cola local.`,
  servicioMarca: "servicio tercerizado",
  propioMarca: "personal propio",
} as const

export const MAPA = {
  capas: "Capas",
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
  leyendo: "Leyendo catastro…",
} as const

export const RIEGO = {
  titulo: "Riego",
  lede: "Sector de capataz, turno y cuadrilla. Cobertura: definición pendiente.",
  zona: "Zona de supervisión",
  sector: "Sector de capataz",
  sectorPlaceholder: "Nombre del sector de capataz",
  turno: "Turno",
  fecha: "Fecha",
  nota: "Nota",
  registrar: "Registrar turno",
  sinCuadrilla: "sin cuadrilla",
  vacio: "Todavía no hay turnos registrados.",
  manana: "Mañana",
  tarde: "Tarde",
} as const

export const SOLICITUD = {
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
  aria: "Evidencias de la actividad",
  fotos: "Fotos de la actividad",
  sinFotos: "Esta actividad no tiene fotos.",
  esperaServidor: "La foto se guarda cuando la actividad ya está en el servidor.",
  enNavegador: "Guardado en este navegador.",
  enNavegadorSinPunto: "Guardado en este navegador, sin ubicación.",
  sesion: "La sesión venció. La foto sigue en este navegador.",
  sinPermiso: "Sin permiso para esta actividad. La foto sigue en este navegador.",
  noSincronizada: "La actividad aún no está sincronizada. La foto sigue en este navegador.",
} as const

export const PODA = {
  lede: "Registro de poda sobre un ejemplar. El código externo solo se conserva si ya viene como OSG.",
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
  clase_actividad: "Clases de actividad",
  plaga: "Plagas",
  producto_fitosanitario: "Productos fitosanitarios",
  frecuencia: "Frecuencias",
  sede: "Sedes",
  cuartel: "Cuarteles (histórico)",
  sector_capataz: "Sectores de capataz",
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
  reservas: "Reservas de jardín",
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
