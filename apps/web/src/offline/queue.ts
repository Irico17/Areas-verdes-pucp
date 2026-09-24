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

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB, 2)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE)) db.createObjectStore(STORE, { keyPath: "id" })
      if (!db.objectStoreNames.contains(ESTADOS)) db.createObjectStore(ESTADOS, { keyPath: "id" })
      if (!db.objectStoreNames.contains(CACHE)) db.createObjectStore(CACHE, { keyPath: "id" })
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
