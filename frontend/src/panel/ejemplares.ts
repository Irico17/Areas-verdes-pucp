import type { Rol } from "../types"
import { EJEMPLAR } from "../ui/nomenclatura"

export type Ejemplar = {
  id: number
  numero_origen?: number
  codigo: string
  especie_id?: number
  especie_cientifico?: string
  nombre_comun: string
  tipo_vegetacion: string
  cantidad: number
  ubicacion_lugar_id?: number
  lugar_nombre?: string
  lat?: number
  lon?: number
  salud?: string | null
  sector_cuartel_id?: number
  sector_cuartel_nombre?: string
  sector_cuartel_clase?: string
  activo: boolean
}

export type CodigoHistorico = {
  id: number
  ejemplar_id: number
  codigo_anterior: string
  codigo_nuevo: string
}

export type Opcion = { id: number; nombre: string; clase?: string }

const SALUDES = ["bueno", "regular", "deteriorado"] as const

export function puedeEditarEjemplar(rol: Rol): boolean {
  return rol === "coordinacion" || rol === "admin"
}

export function etiquetaSalud(codigo: string | null | undefined): string {
  if (codigo === "bueno") return EJEMPLAR.saludBueno
  if (codigo === "regular") return EJEMPLAR.saludRegular
  if (codigo === "deteriorado") return EJEMPLAR.saludDeteriorado
  return EJEMPLAR.sinSalud
}

export function lineaCodigo(anterior: string, nuevo: string): string {
  const viejo = anterior.trim() || EJEMPLAR.sinCodigo
  const actual = nuevo.trim() || EJEMPLAR.sinCodigo
  return `${viejo} ${EJEMPLAR.pasoA} ${actual}`
}

export function tituloLista(item: Ejemplar): string {
  const codigo = item.codigo.trim()
  if (codigo) return codigo
  if (item.numero_origen != null) return `${item.numero_origen}`
  return EJEMPLAR.sinCodigo
}

export function cuerpoFicha(ficha: {
  salud: string
  especieId: string
  lugarId: string
  sectorId: string
  lat: string
  lon: string
}): { ok: true; cuerpo: Record<string, unknown> } | { ok: false; aviso: string } {
  const latVacio = ficha.lat.trim() === ""
  const lonVacio = ficha.lon.trim() === ""
  if (latVacio !== lonVacio) return { ok: false, aviso: EJEMPLAR.errorCoord }
  let lat: number | null = null
  let lon: number | null = null
  if (!latVacio) {
    lat = Number(ficha.lat)
    lon = Number(ficha.lon)
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) return { ok: false, aviso: EJEMPLAR.errorCoord }
  }
  return {
    ok: true,
    cuerpo: {
      salud: ficha.salud || null,
      especie_id: ficha.especieId ? Number(ficha.especieId) : null,
      ubicacion_lugar_id: ficha.lugarId ? Number(ficha.lugarId) : null,
      sector_cuartel_id: ficha.sectorId ? Number(ficha.sectorId) : null,
      lat,
      lon,
    },
  }
}

export function saludValida(codigo: string): boolean {
  return codigo === "" || (SALUDES as readonly string[]).includes(codigo)
}
