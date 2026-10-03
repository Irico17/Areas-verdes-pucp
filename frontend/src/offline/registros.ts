import { apiUrl } from "../api"
import {
  drenarRegistros,
  encolarRegistro,
  listRegistros,
  quitarRegistro,
  clasificarEstado,
  type ConflictoCola,
  type QueuedRegistro,
  type ResultadoEnvio,
  type TipoRegistro,
} from "./queue"

export const COLA_VACIADA = "campus-cola-vaciada"

export type DetalleCola = { conflictos: ConflictoCola[] }

export async function enviarJson(path: string, method: string, body: unknown): Promise<ResultadoEnvio> {
  try {
    const res = await fetch(path, {
      method,
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    })
    return clasificarEstado(res.status)
  } catch {
    return { tipo: "reintento", status: 0 }
  }
}

export type AltaRegistro = Omit<QueuedRegistro, "createdAt">

/** Publica ya. Si no hay red, deja el cuerpo en la cola y no lo descarta. */
export async function publicarOEncolar(item: AltaRegistro): Promise<"ok" | "conflicto" | "encolado" | "error"> {
  const resultado = await enviarJson(item.path, item.method, item.body)
  if (resultado.tipo === "ok") return "ok"
  if (resultado.tipo === "conflicto") return "conflicto"
  if (resultado.tipo === "reintento" && resultado.status === 0) {
    await encolarRegistro({ ...item, createdAt: new Date().toISOString() })
    return "encolado"
  }
  return "error"
}

export function rutaAvance(actividadId: string): string {
  return apiUrl(`/operacion/actividades/${actividadId}/avances`)
}

export function rutaFicha(actividadId: string): string {
  return apiUrl(`/operacion/actividades/${actividadId}/ficha`)
}

export function rutaRiego(): string {
  return apiUrl("/riego")
}

export function rutaPoda(id: string, nueva: boolean): string {
  return nueva ? apiUrl("/podas") : apiUrl(`/podas/${id}`)
}

export function rutaVivero(): string {
  return apiUrl("/vivero")
}

function avisar(conflictos: ConflictoCola[]): void {
  if (typeof window === "undefined") return
  window.dispatchEvent(new CustomEvent<DetalleCola>(COLA_VACIADA, { detail: { conflictos } }))
}

let cadena: Promise<void> = Promise.resolve()

async function correr(): Promise<void> {
  let conflictos: ConflictoCola[] = []
  try {
    const items = await listRegistros()
    const resultado = await drenarRegistros(items, (item) => enviarJson(item.path, item.method, item.body), quitarRegistro)
    conflictos = resultado.conflictos
  } catch {
    conflictos = []
  }
  avisar(conflictos)
}

/** Vacía avances, riego, poda, vivero y fichas. Un 409 sale de la cola y se avisa. */
export function vaciarRegistros(): Promise<void> {
  const siguiente = cadena.then(correr, correr)
  cadena = siguiente.then(
    () => {},
    () => {},
  )
  return siguiente
}

export async function pendientesDe(tipo: TipoRegistro): Promise<QueuedRegistro[]> {
  const items = await listRegistros()
  return items.filter((item) => item.tipo === tipo)
}
