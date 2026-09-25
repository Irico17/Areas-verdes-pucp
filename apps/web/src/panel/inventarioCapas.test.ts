import assert from "node:assert/strict"
import test from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { InventarioCapas } from "./InventarioCapas.tsx"
import {
  CAPAS_EDITABLES,
  SUMAS_PUBLICADAS,
  csvFichas,
  enviarInventario,
  solicitudBaja,
  solicitudGuardar,
  sumarConteos,
  validarBebedero,
  validarPunto,
  validarReserva,
} from "./inventarioCapas.ts"

test("las seis capas de la sección 4.13 son editables", () => {
  assert.deepEqual(CAPAS_EDITABLES.map((capa) => capa.id), [
    "fauna",
    "puertas",
    "playas_estacionamiento",
    "veredas_riesgo",
    "xerofiticas",
    "jardines_reserva",
  ])
})

test("los 11 conteos publicados suman el mapa de datos", () => {
  const total = Object.values(SUMAS_PUBLICADAS).reduce((a, b) => a + b, 0)
  assert.equal(total, 294 + 142 + 260 + 242 + 56 + 0 + 12 + 2 + 38 + 26 + 4)
  const suma = sumarConteos([SUMAS_PUBLICADAS])
  assert.equal(suma.aniquem, 38)
  assert.equal(suma.peligrosos, 0)
})

test("bebedero exige estado y sede", () => {
  assert.equal(validarBebedero({ codigo: "PT_bb1", subtipo: "fuente", estado: "operativo", sede: "CAMPUS" }).length, 0)
  assert.ok(validarBebedero({ codigo: "PT_bb1", subtipo: "fuente", estado: "", sede: "CAMPUS" }).some((e) => e.campo === "estado"))
})

test("punto rechaza teléfono y placeId", () => {
  const errores = validarPunto({ titulo: "Biblioteca", lat: -12.07, lon: -77.08, url: "https://maps.example/?placeId=ChIJ" })
  assert.ok(errores.some((e) => e.campo === "url"))
  assert.equal(validarPunto({ titulo: "Biblioteca", lat: -12.07, lon: -77.08, url: "https://maps.example/?query=Biblioteca" }).length, 0)
})

test("reserva que no es ficticia se rechaza", () => {
  const base = { origen: "ficticio", fecha: "2026-09-22", hora_inicio: "09:00", hora_fin: "11:00", estado: "reservado", evento: "Taller" }
  assert.equal(validarReserva(base).length, 0)
  assert.ok(validarReserva({ ...base, origen: "hoja" }).some((e) => e.campo === "origen"))
})

test("el CSV de fichas no lleva contacto", () => {
  const csv = csvFichas([{ feature_id: "FA-0001", nombre: "Aves", codigo: "" }])
  assert.match(csv, /FA-0001,Aves,/)
  assert.equal(/phone|placeId/i.test(csv), false)
})

test("alta, edición y baja pegan a la API", () => {
  const alta = solicitudGuardar("tachos", null, { codigo: "PT1", lat: -12.07, lon: -77.08 })
  assert.equal(alta.method, "POST")
  assert.equal(alta.path, "/api/v1/inventario/tachos")
  const edicion = solicitudGuardar("jardines_reserva", 9, {
    feature_id: "JR-0001",
    codigo: "J1",
    nota: "norte",
    riego: "goteo",
    area_m2: 12,
    perimetro_m: 14,
    pertenecen: "EEGGLL",
    geojson: '{"type":"MultiPolygon","coordinates":[]}',
  })
  assert.equal(edicion.method, "PATCH")
  assert.equal(edicion.path, "/api/v1/inventario/capas/jardines_reserva/9")
  const baja = solicitudBaja("puntos", 3)
  assert.equal(baja.method, "DELETE")
  assert.equal(baja.path, "/api/v1/inventario/puntos/3")
})

test("enviarInventario usa la cookie y no manda contacto", async () => {
  let visto: { path: string; init: RequestInit } | undefined
  const cliente = (async (path: string, init?: RequestInit) => {
    visto = { path, init: init ?? {} }
    return new Response(JSON.stringify({ id: 4 }), { status: 201, headers: { "content-type": "application/json" } })
  }) as typeof fetch
  await enviarInventario(cliente, solicitudGuardar("reservas", null, { origen: "ficticio", evento: "Taller" }))
  assert.equal(visto?.path, "/api/v1/inventario/reservas")
  assert.equal(visto?.init.method, "POST")
  assert.equal(visto?.init.credentials, "include")
  assert.match(String(visto?.init.body), /ficticio/)
  assert.equal(/phone|placeId/i.test(String(visto?.init.body)), false)
})

test("la ficha de jardín muestra geometría y los campos 4.13", () => {
  const html = renderToStaticMarkup(createElement(InventarioCapas, { entidadInicial: "jardines_reserva", cliente: (() => Promise.reject(new Error("sin red"))) as typeof fetch }))
  for (const etiqueta of ["Código", "Nota", "Riego", "Área", "Perímetro", "Pertenecen", "Geometría"]) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
  assert.equal(html.includes("Listo para guardar"), false)
})
