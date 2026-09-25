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
  createdAt: string
}

export type ColaEvidencias = {
  all: () => Promise<QueuedEvidencia[]>
  put: (item: QueuedEvidencia) => Promise<void>
  del: (id: string) => Promise<void>
}

export type ResultadoEnvio = "ok" | "conflicto" | "despues"

export async function drenarEvidencias(
  cola: ColaEvidencias,
  post: (item: QueuedEvidencia) => Promise<ResultadoEnvio>,
  curso: Set<string> = new Set(),
): Promise<{ enviadas: string[]; conflictos: string[] }> {
  const items = [...(await cola.all())].sort((a, b) => a.createdAt.localeCompare(b.createdAt))
  const enviadas: string[] = []
  const conflictos: string[] = []
  for (const item of items) {
    if (curso.has(item.id)) continue
    curso.add(item.id)
    try {
      const resultado = await post(item)
      if (resultado === "despues") continue
      await cola.del(item.id)
      if (resultado === "ok") enviadas.push(item.id)
      else conflictos.push(item.id)
    } catch {
      /* sin red: el registro sigue en la cola */
    } finally {
      curso.delete(item.id)
    }
  }
  return { enviadas, conflictos }
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
): Promise<{ enviadas: string[]; conflictos: string[] }> {
  return drenarEvidencias(colaIndexed(), post, envioEnCurso)
}

