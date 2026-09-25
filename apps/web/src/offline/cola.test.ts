import assert from "node:assert/strict"
import test from "node:test"
import { clasificarEstado, drenarEvidencias, mensajeReintento, type ColaEvidencias, type QueuedEvidencia } from "./queue.ts"

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
      return { tipo: "ok" }
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
    return { tipo: "conflicto" as const }
  }
  const r = await drenarEvidencias(cola, post, new Set())
  assert.deepEqual(r.conflictos, ["9"])
  assert.equal(intentos, 1)
  await drenarEvidencias(cola, post, new Set())
  assert.equal(intentos, 1)
  assert.deepEqual(cola.ids(), [])
})

test("sin red el registro permanece y se reintenta después del backoff", async () => {
  const cola = memoria([item("3")])
  const inicio = 1_700_000_000_000
  await drenarEvidencias(cola, async () => ({ tipo: "reintento", status: 0 }), new Set(), () => inicio)
  assert.deepEqual(cola.ids(), ["3"])
  const pronto = await drenarEvidencias(cola, async () => ({ tipo: "ok" }), new Set(), () => inicio)
  assert.deepEqual(pronto.enviadas, [])
  const r = await drenarEvidencias(cola, async () => ({ tipo: "ok" }), new Set(), () => inicio + 5_000)
  assert.deepEqual(r.enviadas, ["3"])
  assert.deepEqual(cola.ids(), [])
})

test("401, 403 y 404 no se descartan y esperan el backoff", async () => {
  for (const status of [400, 401, 403, 404, 500]) {
    const cola = memoria([item(String(status))])
    let intentos = 0
    const ahora = 1_700_000_000_000
    const r = await drenarEvidencias(
      cola,
      async () => {
        intentos += 1
        return { tipo: "reintento", status }
      },
      new Set(),
      () => ahora,
    )
    assert.deepEqual(r.reintentos, [{ id: String(status), status }])
    assert.deepEqual(cola.ids(), [String(status)])
    const guardado = (await cola.all())[0]
    assert.ok(guardado.proximoIntento && guardado.proximoIntento > new Date(ahora).toISOString())
    await drenarEvidencias(cola, async () => {
      intentos += 1
      return { tipo: "ok" }
    }, new Set(), () => ahora)
    assert.equal(intentos, 1, `status ${status} se reintentó antes de tiempo`)
    assert.equal(clasificarEstado(status).tipo, "reintento")
  }
  assert.match(mensajeReintento(401), /sesión venció/)
  assert.match(mensajeReintento(403), /Sin permiso/)
  assert.match(mensajeReintento(404), /aún no está sincronizada/)
  assert.equal(clasificarEstado(409).tipo, "conflicto")
  assert.equal(clasificarEstado(201).tipo, "ok")
})
