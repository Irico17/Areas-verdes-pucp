import type { CreateBody } from "../operacion"

export type QueuedLabor = {
  id: string
  body: CreateBody
  createdAt: string
}

const DB = "campus-verde"
const STORE = "cola-labores"

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB, 1)
    req.onupgradeneeded = () => {
      if (!req.result.objectStoreNames.contains(STORE)) {
        req.result.createObjectStore(STORE, { keyPath: "id" })
      }
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
