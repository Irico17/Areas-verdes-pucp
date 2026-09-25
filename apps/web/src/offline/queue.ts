import type { CreateBody } from "../operacion"

export type QueuedLabor = {
  id: string
  body: CreateBody
  createdAt: string
}

const DB = "campus-verde"
const STORE = "cola-labores"
const ESTADOS = "cola-estados"
const CACHE = "cache-labores"
const EVIDENCIAS = "cola-evidencias"

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB, 3)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE)) db.createObjectStore(STORE, { keyPath: "id" })
      if (!db.objectStoreNames.contains(ESTADOS)) db.createObjectStore(ESTADOS, { keyPath: "id" })
      if (!db.objectStoreNames.contains(CACHE)) db.createObjectStore(CACHE, { keyPath: "id" })
      if (!db.objectStoreNames.contains(EVIDENCIAS)) db.createObjectStore(EVIDENCIAS, { keyPath: "id" })
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error ?? new Error("IndexedDB no disponible"))
  })
}

function request<T>(req: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error ?? new Error("falló IndexedDB"))
  })
}

export async function listQueue(): Promise<QueuedLabor[]> {
  const db = await openDb()
  const rows = await request(db.transaction(STORE, "readonly").objectStore(STORE).getAll())
  db.close()
  return (rows as QueuedLabor[]).sort((a, b) => a.createdAt.localeCompare(b.createdAt))
}

export async function enqueue(item: QueuedLabor): Promise<void> {
  const db = await openDb()
  await request(db.transaction(STORE, "readwrite").objectStore(STORE).put(item))
  db.close()
}

export async function removeQueued(id: string): Promise<void> {
  const db = await openDb()
  await request(db.transaction(STORE, "readwrite").objectStore(STORE).delete(id))
  db.close()
}

export type QueuedEstado = {
  id: string
  actividadId: string
  estado: string
  createdAt: string
}

export async function listEstados(): Promise<QueuedEstado[]> {
  const db = await openDb()
  const rows = await request(db.transaction(ESTADOS, "readonly").objectStore(ESTADOS).getAll())
  db.close()
  return (rows as QueuedEstado[]).sort((a, b) => a.createdAt.localeCompare(b.createdAt))
}

export async function enqueueEstado(item: QueuedEstado): Promise<void> {
  const db = await openDb()
  await request(db.transaction(ESTADOS, "readwrite").objectStore(ESTADOS).put(item))
  db.close()
}

export async function removeEstado(id: string): Promise<void> {
  const db = await openDb()
  await request(db.transaction(ESTADOS, "readwrite").objectStore(ESTADOS).delete(id))
  db.close()
}

export async function saveLabores(body: unknown): Promise<void> {
  const db = await openDb()
  await request(db.transaction(CACHE, "readwrite").objectStore(CACHE).put({ id: "abiertas", body }))
  db.close()
}

export async function loadLabores<T>(): Promise<T | null> {
  const db = await openDb()
  const row = await request(db.transaction(CACHE, "readonly").objectStore(CACHE).get("abiertas"))
  db.close()
  if (!row || typeof row !== "object" || !("body" in row)) return null
  return (row as { body: T }).body
}

export type QueuedEvidencia = {
  id: string
  actividadId: string
  nota: string
  nombre: string
  mime: string
  bytes: ArrayBuffer
  sha256: string
  lat: number | null
  lon: number | null
  exif: Record<string, string | number | null>
  ordenId?: string
  createdAt: string
  intentos?: number
  proximoIntento?: string
}

export type ColaEvidencias = {
  all: () => Promise<QueuedEvidencia[]>
  put: (item: QueuedEvidencia) => Promise<void>
  del: (id: string) => Promise<void>
}

export type ResultadoEnvio =
  | { tipo: "ok" }
  | { tipo: "conflicto" }
  | { tipo: "reintento"; status: number }

export function clasificarEstado(status: number): ResultadoEnvio {
  if (status >= 200 && status < 300) return { tipo: "ok" }
  if (status === 409) return { tipo: "conflicto" }
  return { tipo: "reintento", status }
}

export function mensajeReintento(status: number): string {
  if (status === 401) return "La sesión venció. La foto sigue en este equipo."
  if (status === 403) return "Sin permiso para esta labor. La foto sigue en este equipo."
  if (status === 404) return "La labor aún no está sincronizada. La foto sigue en este equipo."
  if (status === 400) return "No se pudo enviar la foto. Se reintentará."
  return "No se pudo enviar la foto. Se reintentará."
}

export function conBackoff(item: QueuedEvidencia, ahora = Date.now()): QueuedEvidencia {
  const intentos = (item.intentos ?? 0) + 1
  const espera = Math.min(60_000, 2_000 * 2 ** Math.min(intentos - 1, 5))
  return { ...item, intentos, proximoIntento: new Date(ahora + espera).toISOString() }
}

export async function drenarEvidencias(
  cola: ColaEvidencias,
  post: (item: QueuedEvidencia) => Promise<ResultadoEnvio>,
  curso: Set<string> = new Set(),
  ahora: () => number = Date.now,
): Promise<{ enviadas: string[]; conflictos: string[]; reintentos: { id: string; status: number }[] }> {
  const items = [...(await cola.all())].sort((a, b) => a.createdAt.localeCompare(b.createdAt))
  const enviadas: string[] = []
  const conflictos: string[] = []
  const reintentos: { id: string; status: number }[] = []
  const reloj = new Date(ahora()).toISOString()
  for (const item of items) {
    if (curso.has(item.id)) continue
    if (item.proximoIntento && item.proximoIntento > reloj) continue
    curso.add(item.id)
    try {
      const resultado = await post(item)
      if (resultado.tipo === "reintento") {
        await cola.put(conBackoff(item, ahora()))
        reintentos.push({ id: item.id, status: resultado.status })
        continue
      }
      await cola.del(item.id)
      if (resultado.tipo === "ok") enviadas.push(item.id)
      else conflictos.push(item.id)
    } catch {
      await cola.put(conBackoff(item, ahora()))
      reintentos.push({ id: item.id, status: 0 })
    } finally {
      curso.delete(item.id)
    }
  }
  return { enviadas, conflictos, reintentos }
}

function colaIndexed(): ColaEvidencias {
  return {
    async all() {
      const db = await openDb()
      const rows = await request(db.transaction(EVIDENCIAS, "readonly").objectStore(EVIDENCIAS).getAll())
      db.close()
      return rows as QueuedEvidencia[]
    },
    async put(item) {
      const db = await openDb()
      await request(db.transaction(EVIDENCIAS, "readwrite").objectStore(EVIDENCIAS).put(item))
      db.close()
    },
    async del(id) {
      const db = await openDb()
      await request(db.transaction(EVIDENCIAS, "readwrite").objectStore(EVIDENCIAS).delete(id))
      db.close()
    },
  }
}

const envioEnCurso = new Set<string>()

export async function listarEvidencias(): Promise<QueuedEvidencia[]> {
  const rows = await colaIndexed().all()
  return rows.sort((a, b) => a.createdAt.localeCompare(b.createdAt))
}

export async function encolarEvidencia(item: QueuedEvidencia): Promise<void> {
  await colaIndexed().put(item)
}

export async function quitarEvidencia(id: string): Promise<void> {
  await colaIndexed().del(id)
}

export function vaciarEnvioEnCurso(): void {
  envioEnCurso.clear()
}

export async function enviarColaEvidencias(
  post: (item: QueuedEvidencia) => Promise<ResultadoEnvio>,
): Promise<{ enviadas: string[]; conflictos: string[]; reintentos: { id: string; status: number }[] }> {
  return drenarEvidencias(colaIndexed(), post, envioEnCurso)
}

