import { apiUrl } from "../api"

export type SectorCapataz = {
  id: number
  codigo: string
  nombre: string
  color: string
  activo: boolean
}

export type LugarCatalogo = { id: number; nombre: string }
export type EdificioRef = { id: string; nombre: string }

async function leer(path: string): Promise<unknown> {
  const res = await fetch(apiUrl(path), { credentials: "include" })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json()
}

async function enviar(path: string, method: string, body: unknown): Promise<unknown> {
  const res = await fetch(apiUrl(path), {
    method,
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json()
}

async function mensaje(res: Response): Promise<string> {
  const data = (await res.json().catch(() => null)) as { error?: string } | null
  return data?.error || `La solicitud respondió ${res.status}`
}

export async function listarSectores(soloActivos = false): Promise<SectorCapataz[]> {
  const body = (await leer(`/zonificacion/sectores${soloActivos ? "?activos=1" : ""}`)) as { sectores?: SectorCapataz[] }
  return body.sectores ?? []
}

export async function crearSector(sector: { codigo: string; nombre: string; color: string }): Promise<SectorCapataz> {
  return (await enviar("/zonificacion/sectores", "POST", sector)) as SectorCapataz
}

export async function desactivarSector(codigo: string): Promise<void> {
  await enviar(`/zonificacion/sectores/${encodeURIComponent(codigo)}/desactivar`, "POST", {})
}

export async function importarSectores(sectores: { codigo: string; nombre: string; color: string }[]): Promise<void> {
  await enviar("/zonificacion/sectores/importar", "POST", { sectores })
}

export async function listarLugaresCatalogo(): Promise<LugarCatalogo[]> {
  const body = (await leer("/zonificacion/lugares")) as { lugares?: LugarCatalogo[] }
  return body.lugares ?? []
}

export async function listarEdificiosReferente(): Promise<EdificioRef[]> {
  const body = (await leer("/zonificacion/edificios")) as { edificios?: EdificioRef[] }
  return body.edificios ?? []
}

export async function guardarReferente(lugarId: number, edificioId: string): Promise<void> {
  await enviar("/zonificacion/referentes", "POST", { lugar_id: lugarId, edificio_id: edificioId })
}

export async function importarVias(archivo: unknown): Promise<void> {
  const res = await fetch(apiUrl("/zonificacion/vias/importar"), {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(archivo),
  })
  if (!res.ok) throw new Error(await mensaje(res))
}
