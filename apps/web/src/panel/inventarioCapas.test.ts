import assert from "node:assert/strict"
import test from "node:test"
import { CAPAS_EDITABLES, SUMAS_PUBLICADAS, csvFichas, sumarConteos, validarBebedero, validarPunto, validarReserva } from "./inventarioCapas.ts"

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
