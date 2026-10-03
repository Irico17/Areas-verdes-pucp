import type { Rol } from "../types"
import { MODULO } from "./nomenclatura"
import { permisosDeRol, tienePermiso, type Accion } from "./permisos"

export type Modulo =
  | "hoy"
  | "resumen"
  | "mapa"
  | "labores"
  | "registros"
  | "catastro"
  | "inventario"
  | "solicitudes"
  | "reportes"
  | "catalogos"
  | "importaciones"
  | "admin"
  | "ejemplares"
  | "bitacora"
  | "historial"

export type GrupoNav = "operacion" | "datos" | "configuracion"

/** Orden de presentación. La visibilidad la deciden los permisos, no esta lista sola. */
const ORDEN: Record<Rol, readonly Modulo[]> = {
  capataz: ["hoy", "mapa", "labores", "registros", "bitacora", "ejemplares"],
  coordinacion: [
    "resumen",
    "mapa",
    "labores",
    "solicitudes",
    "registros",
    "catastro",
    "inventario",
    "reportes",
    "ejemplares",
    "bitacora",
    "importaciones",
    "catalogos",
    "historial",
  ],
  jefatura: [
    "resumen",
    "mapa",
    "labores",
    "reportes",
    "solicitudes",
    "catastro",
    "ejemplares",
    "bitacora",
    "importaciones",
    "historial",
    "admin",
  ],
  admin: [
    "resumen",
    "mapa",
    "labores",
    "solicitudes",
    "registros",
    "catastro",
    "inventario",
    "reportes",
    "ejemplares",
    "bitacora",
    "admin",
    "catalogos",
    "importaciones",
    "historial",
  ],
}

const ORDEN_OFICINA = ORDEN.coordinacion

const IDS: Modulo[] = [
  "hoy",
  "resumen",
  "mapa",
  "labores",
  "registros",
  "catastro",
  "inventario",
  "solicitudes",
  "reportes",
  "catalogos",
  "importaciones",
  "admin",
  "ejemplares",
  "bitacora",
  "historial",
]

export const MODULOS: { id: Modulo; label: string }[] = IDS.map((id) => ({ id, label: MODULO[id] }))

function esCampo(permisos: readonly string[]): boolean {
  return (
    tienePermiso(permisos, "consultar") &&
    tienePermiso(permisos, "registrar") &&
    !tienePermiso(permisos, "validar") &&
    !tienePermiso(permisos, "reportes") &&
    !tienePermiso(permisos, "solicitudes")
  )
}

/** Una pestaña existe solo si el permiso que la API exige está concedido. */
export function moduloPermitido(id: Modulo, permisos: readonly string[]): boolean {
  const p = (accion: Accion) => tienePermiso(permisos, accion)
  const campo = esCampo(permisos)
  switch (id) {
    case "hoy":
      return campo && p("consultar")
    case "resumen":
      return p("consultar") && !campo
    case "mapa":
    case "labores":
    case "ejemplares":
    case "bitacora":
      return p("consultar")
    case "registros":
      return p("registrar")
    case "solicitudes":
      return p("solicitudes")
    case "catastro":
      return p("consultar") && !campo
    case "inventario":
      return p("registrar") && !campo
    case "reportes":
      return p("reportes")
    case "importaciones":
    case "historial":
      return p("validar")
    case "catalogos":
      return p("catalogos") || (p("registrar") && p("validar"))
    case "admin":
      return p("usuarios")
    default:
      return false
  }
}

export function modulosPorPermisos(permisos: readonly string[], rol: string): Modulo[] {
  const orden = rol === "capataz" || rol === "coordinacion" || rol === "jefatura" || rol === "admin" ? ORDEN[rol] : ORDEN_OFICINA
  const vistos = new Set<Modulo>()
  const salida: Modulo[] = []
  for (const id of orden) {
    if (moduloPermitido(id, permisos)) {
      salida.push(id)
      vistos.add(id)
    }
  }
  for (const id of ORDEN_OFICINA) {
    if (!vistos.has(id) && moduloPermitido(id, permisos)) salida.push(id)
  }
  return salida
}

export function modulosDe(rol: Rol): Modulo[] {
  return modulosPorPermisos(permisosDeRol(rol), rol)
}

export function entradaDe(rol: string): Modulo {
  return rol === "capataz" ? "hoy" : "resumen"
}

export function grupoDe(id: Modulo): GrupoNav {
  if (id === "importaciones" || id === "catalogos" || id === "admin" || id === "historial") return "configuracion"
  if (id === "catastro" || id === "inventario" || id === "ejemplares" || id === "bitacora") return "datos"
  return "operacion"
}

export function etiquetaModulo(id: Modulo, rol: string): string {
  if (id === "labores" && rol === "capataz") return MODULO.laboresCapataz
  return MODULO[id]
}

export function etiquetaCorta(id: Modulo, rol: string): string {
  if (id === "registros") return MODULO.registrosCorto
  if (id === "solicitudes") return MODULO.solicitudesCorto
  return etiquetaModulo(id, rol)
}
