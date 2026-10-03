import type { Rol } from "../types"

/**
 * Copia de la matriz semilla del backend (`MatrizPermisos`).
 * GET /sesion no devuelve permisos: las pestañas se derivan de esta tabla.
 * Los códigos de acción no se renombran.
 */
export const ACCIONES = [
  "consultar",
  "registrar",
  "validar",
  "solicitudes",
  "reportes",
  "catalogos",
  "usuarios",
  "evidencias",
] as const

export type Accion = (typeof ACCIONES)[number]

export const MATRIZ_SEMILLA: Record<Rol, readonly Accion[]> = {
  capataz: ["consultar", "registrar"],
  coordinacion: ["consultar", "registrar", "validar", "solicitudes", "reportes"],
  jefatura: ["consultar", "validar", "reportes", "solicitudes", "evidencias", "usuarios"],
  admin: ["consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes", "usuarios"],
}

export function permisosDeRol(rol: string): readonly Accion[] {
  if (rol === "capataz" || rol === "coordinacion" || rol === "jefatura" || rol === "admin") {
    return MATRIZ_SEMILLA[rol]
  }
  return []
}

export function tienePermiso(permisos: readonly string[], accion: Accion): boolean {
  return permisos.includes(accion)
}
