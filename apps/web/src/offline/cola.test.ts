import assert from "node:assert/strict"
import test from "node:test"
import { drenarEvidencias, type ColaEvidencias, type QueuedEvidencia } from "./queue.ts"

function item(id: string, marca = "a"): QueuedEvidencia {
  return {
    id,
    actividadId: "11111111-1111-4111-8111-111111111111",
    nota: "",
    nombre: "foto.jpg",
    mime: "image/jpeg",
    bytes: new TextEncoder().encode(marca).buffer,
    sha256: marca.repeat(64).slice(0, 64),
    lat: null,
    lon: null,
    exif: {},
    createdAt: `2026-09-25T10:00:0${id}.000Z`,
  }
}

function memoria(inicial: QueuedEvidencia[]): ColaEvidencias & { ids: () => string[] } {
  const rows = new Map(inicial.map((row) => [row.id, row]))
  return {
    all: async () => [...rows.values()],
    put: async (row) => {
      rows.set(row.id, row)
    },
    del: async (id) => {
      rows.delete(id)
    },
    ids: () => [...rows.keys()],
  }
}

test("la cola envía cada evidencia una sola vez y la quita", async () => {
  const cola = memoria([item("1"), item("2")])
  const vistos: string[] = []
  const curso = new Set<string>()
  const primero = await drenarEvidencias(
    cola,
    async (row) => {
      vistos.push(row.id)
      return "ok"
    },
    curso,
  )
  assert.deepEqual(primero.enviadas, ["1", "2"])
  assert.deepEqual(cola.ids(), [])
  const segundo = await drenarEvidencias(
    cola,
    async () => {
      throw new Error("no debía reenviar")
    },
    curso,
  )
  assert.deepEqual(vistos, ["1", "2"])
  assert.deepEqual(segundo.enviadas, [])
})

test("un 409 saca el UUID de la cola y no lo reintenta", async () => {
  const cola = memoria([item("9", "b")])
  let intentos = 0
  const post = async () => {
    intentos += 1
    return "conflicto" as const
  }
  const r = await drenarEvidencias(cola, post, new Set())
  assert.deepEqual(r.conflictos, ["9"])
  assert.equal(intentos, 1)
  await drenarEvidencias(cola, post, new Set())
  assert.equal(intentos, 1)
  assert.deepEqual(cola.ids(), [])
})

test("sin red el registro permanece para el siguiente intento", async () => {
  const cola = memoria([item("3")])
  await drenarEvidencias(cola, async () => "despues", new Set())
  assert.deepEqual(cola.ids(), ["3"])
  const r = await drenarEvidencias(cola, async () => "ok", new Set())
  assert.deepEqual(r.enviadas, ["3"])
  assert.deepEqual(cola.ids(), [])
})
