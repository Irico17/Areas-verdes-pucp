import type { Evidencia } from "../producto"
import type { QueuedEvidencia } from "./queue"

export type EstadoSubida =
  | { fase: "inactivo" }
  | { fase: "preparando" }
  | { fase: "subiendo"; pct: number | null }
  | { fase: "guardada"; sinPunto: boolean }
  | { fase: "en-cola"; texto: string }
  | { fase: "error"; texto: string; reintentable: boolean }

export type EventoSubida =
  | { tipo: "preparar" }
  | { tipo: "progreso"; cargado: number; total: number | null }
  | { tipo: "ok"; sinPunto: boolean }
  | { tipo: "cola"; texto: string }
  | { tipo: "fallo"; texto: string; reintentable: boolean }
  | { tipo: "limpiar" }

export function avanzar(_e: EstadoSubida, ev: EventoSubida): EstadoSubida {
  switch (ev.tipo) {
    case "preparar":
      return { fase: "preparando" }
    case "progreso": {
      if (ev.total == null || ev.total <= 0) return { fase: "subiendo", pct: null }
      const pct = Math.min(100, Math.max(0, Math.round((ev.cargado / ev.total) * 100)))
      return { fase: "subiendo", pct }
    }
    case "ok":
      return { fase: "guardada", sinPunto: ev.sinPunto }
    case "cola":
      return { fase: "en-cola", texto: ev.texto }
    case "fallo":
      return { fase: "error", texto: ev.texto, reintentable: ev.reintentable }
    case "limpiar":
      return { fase: "inactivo" }
  }
}

export function textoEstado(e: EstadoSubida): string {
  switch (e.fase) {
    case "inactivo":
      return ""
    case "preparando":
      return "Preparando foto…"
    case "subiendo":
      return e.pct == null ? "Subiendo…" : `Subiendo ${e.pct} %`
    case "guardada":
      return e.sinPunto ? "Foto enviada, sin ubicación." : "Foto enviada."
    case "en-cola":
      return e.texto
    case "error":
      return e.texto
  }
}

export function ocupado(e: EstadoSubida): boolean {
  return e.fase === "preparando" || e.fase === "subiendo"
}

export function enviarConProgreso(
  url: string,
  data: FormData,
  alProgreso: (cargado: number, total: number | null) => void,
  xhrFactory: () => XMLHttpRequest = () => new XMLHttpRequest(),
): Promise<number> {
  return new Promise((resolve) => {
    const xhr = xhrFactory()
    xhr.open("POST", url)
    xhr.withCredentials = true
    xhr.timeout = 60_000
    xhr.upload.onprogress = (event) => {
      alProgreso(event.loaded, event.lengthComputable ? event.total : null)
    }
    xhr.onload = () => resolve(xhr.status)
    xhr.onerror = () => resolve(0)
    xhr.ontimeout = () => resolve(0)
    xhr.send(data)
  })
}

export type ItemEvidencia = {
  id: string
  nombre: string
  mime: string
  estado: "pendiente" | "enviada"
  fecha: string
  conPunto?: boolean
  local?: QueuedEvidencia
}

export function unirEvidencias(pendientes: QueuedEvidencia[], enviadas: Evidencia[]): ItemEvidencia[] {
  const idsEnviadas = new Set(enviadas.map((item) => item.id))
  const items: ItemEvidencia[] = []
  for (const p of pendientes) {
    if (idsEnviadas.has(p.id)) continue
    items.push({
      id: p.id,
      nombre: p.nombre,
      mime: p.mime,
      estado: "pendiente",
      fecha: p.createdAt,
      conPunto: p.lat != null && p.lon != null,
      local: p,
    })
  }
  for (const e of enviadas) {
    items.push({ id: e.id, nombre: e.nombre, mime: e.mime, estado: "enviada", fecha: e.created_at })
  }
  return items.sort((a, b) => b.fecha.localeCompare(a.fecha))
}

export function esImagen(mime: string): boolean {
  return mime.startsWith("image/")
}

export function urlArchivo(id: string): string {
  return `/api/v1/evidencias/${encodeURIComponent(id)}/archivo`
}

export function crearCacheUrls(
  crear: (blob: Blob) => string = URL.createObjectURL,
  revocar: (url: string) => void = URL.revokeObjectURL,
): { sincronizar(items: { id: string; blob: Blob }[]): Map<string, string>; liberar(): void } {
  const urls = new Map<string, string>()
  return {
    sincronizar(items) {
      const vivos = new Set(items.map((item) => item.id))
      for (const [id, url] of urls) {
        if (!vivos.has(id)) {
          revocar(url)
          urls.delete(id)
        }
      }
      for (const item of items) {
        if (!urls.has(item.id)) urls.set(item.id, crear(item.blob))
      }
      return urls
    },
    liberar() {
      for (const url of urls.values()) revocar(url)
      urls.clear()
    },
  }
}
