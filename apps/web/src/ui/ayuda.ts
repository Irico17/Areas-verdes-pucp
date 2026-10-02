export type PasoAyuda = { titulo: string; texto: string }

const CAPATAZ: PasoAyuda[] = [
  { titulo: "Hoy", texto: "Aquí están las actividades de su cuadrilla y lo que falta por enviar." },
  { titulo: "En el mapa", texto: "El color es el estado. La letra marca la clase de actividad." },
  { titulo: "Registros", texto: "Riego, poda y vivero se anotan en Registros de campo, no en la lista de actividades." },
]

const OFICINA: PasoAyuda[] = [
  { titulo: "Resumen", texto: "La bandeja muestra lo que pide una decisión: bloqueadas, sin cuadrilla, solicitudes sueltas." },
  { titulo: "Mapa", texto: "Las capas y la leyenda flotan sobre el plano. El panel lateral se abre cuando hace falta." },
  { titulo: "Actividades", texto: "Una actividad se ubica en el mapa y se clasifica con la clase y el tipo del catálogo." },
]

const JEFATURA: PasoAyuda[] = [
  { titulo: "Resumen", texto: "El conteo por estado sale del reporte. No es un indicador oficial." },
  { titulo: "Lectura", texto: "Donde su rol no registra, la ficha se ve y no aparece Guardar." },
  { titulo: "Historial", texto: "Los cambios quedan en Historial, dentro de Configuración. No se borran por antigüedad." },
]

const ADMIN: PasoAyuda[] = [
  { titulo: "Cuentas", texto: "Administración, en Configuración, abre cuentas y permisos de esta instalación." },
  { titulo: "Catálogos", texto: "Los nombres visibles salen del catálogo. El código interno no se renombra." },
  { titulo: "Importar", texto: "Importar datos revisa el archivo antes de escribir. Se puede revertir esa importación." },
]

export function pasosDe(rol: string): PasoAyuda[] {
  if (rol === "capataz") return CAPATAZ
  if (rol === "jefatura") return JEFATURA
  if (rol === "admin") return ADMIN
  return OFICINA
}

export function claveRecorrido(usuario: string): string {
  return `cv:recorrido:${usuario}`
}
