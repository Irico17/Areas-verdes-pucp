import type { Rol } from "./types"

const KEY = "campus-verde-rol"

export function readRol(): Rol {
  const raw = localStorage.getItem(KEY)
  if (raw === "jefatura" || raw === "coordinacion" || raw === "capataz") return raw
  return "coordinacion"
}

export function writeRol(rol: Rol) {
  localStorage.setItem(KEY, rol)
}
