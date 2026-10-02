import assert from "node:assert/strict"
import test from "node:test"
import type { Evidencia } from "../producto.ts"
import type { QueuedEvidencia } from "./queue.ts"
import { apiUrl } from "../api"
import {
  accionTrasFallo,
  avanzar,
  crearCacheUrls,
  enviarConProgreso,
  esImagen,
  ocupado,
  textoEstado,
  unirEvidencias,
  urlArchivo,
  type EstadoSubida,
} from "./subida.ts"

function pendiente(id: string, createdAt: string, lat: number | null = null): QueuedEvidencia {
  return {
    id,
    actividadId: "11111111-1111-4111-8111-111111111111",
    nota: "",
    nombre: `foto-${id}.jpg`,
    mime: "image/jpeg",
    bytes: new ArrayBuffer(0),
    sha256: "a".repeat(64),
    lat,
    lon: lat == null ? null : -77,
    exif: {},
    createdAt,
  }
}

function enviada(id: string, createdAt: string): Evidencia {
  return { id, nombre: `foto-${id}.jpg`, mime: "image/jpeg", bytes: 10, created_at: createdAt }
}

test("avanzar recorre inactivo -> preparando -> subiendo -> guardada", () => {
  let e: EstadoSubida = { fase: "inactivo" }
  e = avanzar(e, { tipo: "preparar" })
  assert.deepEqual(e, { fase: "preparando" })
  e = avanzar(e, { tipo: "progreso", cargado: 0, total: 100 })
  assert.deepEqual(e, { fase: "subiendo", pct: 0 })
  e = avanzar(e, { tipo: "progreso", cargado: 42, total: 100 })
  assert.deepEqual(e, { fase: "subiendo", pct: 42 })
  e = avanzar(e, { tipo: "progreso", cargado: 100, total: 100 })
  assert.deepEqual(e, { fase: "subiendo", pct: 100 })
  e = avanzar(e, { tipo: "ok", sinPunto: false })
  assert.deepEqual(e, { fase: "guardada", sinPunto: false })
})

test("progreso sin total (o total 0) deja pct en null", () => {
  const a = avanzar({ fase: "inactivo" }, { tipo: "progreso", cargado: 10, total: null })
  assert.deepEqual(a, { fase: "subiendo", pct: null })
  const b = avanzar({ fase: "inactivo" }, { tipo: "progreso", cargado: 10, total: 0 })
  assert.deepEqual(b, { fase: "subiendo", pct: null })
})

test("el porcentaje se limita a 0-100 y se redondea", () => {
  const bajo = avanzar({ fase: "inactivo" }, { tipo: "progreso", cargado: -5, total: 100 })
  assert.deepEqual(bajo, { fase: "subiendo", pct: 0 })
  const alto = avanzar({ fase: "inactivo" }, { tipo: "progreso", cargado: 130, total: 100 })
  assert.deepEqual(alto, { fase: "subiendo", pct: 100 })
  const redondeo = avanzar({ fase: "inactivo" }, { tipo: "progreso", cargado: 1, total: 3 })
  assert.deepEqual(redondeo, { fase: "subiendo", pct: 33 })
})

test("fallo reintentable y no reintentable", () => {
  const r = avanzar({ fase: "inactivo" }, { tipo: "fallo", texto: "problema de red", reintentable: true })
  assert.deepEqual(r, { fase: "error", texto: "problema de red", reintentable: true })
  const nr = avanzar({ fase: "inactivo" }, { tipo: "fallo", texto: "duplicado", reintentable: false })
  assert.deepEqual(nr, { fase: "error", texto: "duplicado", reintentable: false })
})

test("limpiar vuelve a inactivo", () => {
  assert.deepEqual(avanzar({ fase: "error", texto: "x", reintentable: true }, { tipo: "limpiar" }), { fase: "inactivo" })
})

test("textoEstado para cada fase", () => {
  assert.equal(textoEstado({ fase: "inactivo" }), "")
  assert.equal(textoEstado({ fase: "preparando" }), "Preparando foto…")
  assert.equal(textoEstado({ fase: "subiendo", pct: null }), "Subiendo…")
  assert.equal(textoEstado({ fase: "subiendo", pct: 42 }), "Subiendo…")
  assert.equal(textoEstado({ fase: "guardada", sinPunto: false }), "Foto enviada.")
  assert.equal(textoEstado({ fase: "guardada", sinPunto: true }), "Foto enviada, sin ubicación.")
  assert.equal(textoEstado({ fase: "en-cola", texto: "Guardado en este equipo." }), "Guardado en este equipo.")
  assert.equal(textoEstado({ fase: "error", texto: "No se pudo enviar.", reintentable: true }), "No se pudo enviar.")
})

test("ocupado solo en preparando y subiendo", () => {
  assert.equal(ocupado({ fase: "inactivo" }), false)
  assert.equal(ocupado({ fase: "preparando" }), true)
  assert.equal(ocupado({ fase: "subiendo", pct: 10 }), true)
  assert.equal(ocupado({ fase: "guardada", sinPunto: false }), false)
  assert.equal(ocupado({ fase: "en-cola", texto: "" }), false)
  assert.equal(ocupado({ fase: "error", texto: "", reintentable: false }), false)
})

type Handler = (() => void) | null

class FakeXHR {
  upload: { onprogress: ((event: { loaded: number; lengthComputable: boolean; total: number }) => void) | null } = {
    onprogress: null,
  }
  onload: Handler = null
  onerror: Handler = null
  ontimeout: Handler = null
  onabort: Handler = null
  status = 0
  withCredentials = false
  timeout = 0
  metodo = ""
  url = ""
  enviado: FormData | null = null
  open(metodo: string, url: string) {
    this.metodo = metodo
    this.url = url
  }
  send(data: FormData) {
    this.enviado = data
  }
}

test("enviarConProgreso emite progreso y resuelve el status", async () => {
  const xhr = new FakeXHR()
  const progresos: { cargado: number; total: number | null }[] = []
  const promesa = enviarConProgreso(
    apiUrl("/evidencias"),
    new FormData(),
    (cargado, total) => progresos.push({ cargado, total }),
    () => xhr as unknown as XMLHttpRequest,
  )
  xhr.upload.onprogress?.({ loaded: 10, lengthComputable: true, total: 100 })
  xhr.upload.onprogress?.({ loaded: 100, lengthComputable: true, total: 100 })
  xhr.status = 201
  xhr.onload?.()
  const status = await promesa
  assert.equal(status, 201)
  assert.deepEqual(progresos, [
    { cargado: 10, total: 100 },
    { cargado: 100, total: 100 },
  ])
  assert.equal(xhr.metodo, "POST")
  assert.equal(xhr.withCredentials, true)
})

test("enviarConProgreso resuelve 0 en onerror y en ontimeout", async () => {
  const xhrError = new FakeXHR()
  const errorProm = enviarConProgreso("/x", new FormData(), () => {}, () => xhrError as unknown as XMLHttpRequest)
  xhrError.onerror?.()
  assert.equal(await errorProm, 0)

  const xhrTimeout = new FakeXHR()
  const timeoutProm = enviarConProgreso("/x", new FormData(), () => {}, () => xhrTimeout as unknown as XMLHttpRequest)
  xhrTimeout.ontimeout?.()
  assert.equal(await timeoutProm, 0)
})

test("enviarConProgreso resuelve 0 en onabort para no dejar la promesa colgada", async () => {
  const xhr = new FakeXHR()
  const promesa = enviarConProgreso("/x", new FormData(), () => {}, () => xhr as unknown as XMLHttpRequest)
  xhr.onabort?.()
  assert.equal(await promesa, 0)
})

test("unirEvidencias descarta la pendiente cuando ya llegó enviada y ordena por fecha descendente", () => {
  const items = unirEvidencias(
    [pendiente("a", "2026-09-25T10:00:00.000Z", -12), pendiente("b", "2026-09-25T09:00:00.000Z")],
    [enviada("a", "2026-09-25T10:00:00.000Z"), enviada("c", "2026-09-25T11:00:00.000Z")],
  )
  assert.deepEqual(
    items.map((item) => item.id),
    ["c", "a", "b"],
  )
  const a = items.find((item) => item.id === "a")
  assert.equal(a?.estado, "enviada")
  const b = items.find((item) => item.id === "b")
  assert.equal(b?.estado, "pendiente")
  assert.equal(b?.conPunto, false)
})

test("unirEvidencias conserva conPunto de la pendiente", () => {
  const items = unirEvidencias([pendiente("z", "2026-09-25T10:00:00.000Z", -12)], [])
  assert.equal(items[0].conPunto, true)
  assert.equal(items[0].local?.id, "z")
})

test("esImagen y urlArchivo", () => {
  assert.equal(esImagen("image/jpeg"), true)
  assert.equal(esImagen("application/pdf"), false)
  assert.equal(urlArchivo("a b/c"), apiUrl("/evidencias/a%20b%2Fc/archivo"))
})

test("crearCacheUrls crea una sola vez por id, revoca los que desaparecen y libera todo", () => {
  const creadas: Blob[] = []
  const revocadas: string[] = []
  let n = 0
  const cache = crearCacheUrls(
    (blob) => {
      creadas.push(blob)
      n += 1
      return `blob:${n}`
    },
    (url) => revocadas.push(url),
  )
  const blobA = new Blob(["a"])
  const blobB = new Blob(["b"])
  const primero = cache.sincronizar([
    { id: "a", blob: blobA },
    { id: "b", blob: blobB },
  ])
  assert.equal(primero.get("a"), "blob:1")
  assert.equal(primero.get("b"), "blob:2")
  assert.equal(creadas.length, 2)

  const segundo = cache.sincronizar([{ id: "a", blob: blobA }])
  assert.equal(segundo.get("a"), "blob:1")
  assert.equal(segundo.has("b"), false)
  assert.deepEqual(revocadas, ["blob:2"])
  assert.equal(creadas.length, 2)

  cache.liberar()
  assert.deepEqual(revocadas, ["blob:2", "blob:1"])
})

test("accionTrasFallo con estado 0 deja el ítem listo para la cola en vez de perderlo", () => {
  const item = pendiente("r1", "2026-09-25T10:00:00.000Z")
  const accion = accionTrasFallo({ tipo: "reintento", status: 0 }, item)
  assert.ok(accion)
  assert.equal(accion?.encolar.id, item.id)
  assert.equal(accion?.encolar.intentos, 1)
  assert.deepEqual(accion?.evento, {
    tipo: "fallo",
    texto: "No se pudo enviar la foto por un problema de red.",
    reintentable: true,
  })
})

test("accionTrasFallo con otros status también encola, con mensaje de cola", () => {
  const item = pendiente("r2", "2026-09-25T10:00:00.000Z")
  const accion = accionTrasFallo({ tipo: "reintento", status: 404 }, item)
  assert.ok(accion)
  assert.equal(accion?.encolar.id, item.id)
  assert.equal(accion?.evento.tipo, "cola")
})

test("accionTrasFallo no hace nada si el envío fue ok o hubo conflicto", () => {
  const item = pendiente("r3", "2026-09-25T10:00:00.000Z")
  assert.equal(accionTrasFallo({ tipo: "ok" }, item), null)
  assert.equal(accionTrasFallo({ tipo: "conflicto" }, item), null)
})
