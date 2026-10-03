import assert from "node:assert/strict"
import { test } from "node:test"
import {
  UMBRAL_TESELAS,
  cargaInicial,
  clasificarErrorMapa,
  marcarEstiloListo,
  marcarEstiloRoto,
  planoNoArranco,
  registrarFalloTesela,
  registrarTeselaOk,
  textoDeError,
} from "./cargaPlano.ts"

const TESELA = (z: number, x: number, y: number) =>
  `AJAXError: Failed to fetch (0): https://tile.openstreetmap.org/${z}/${x}/${y}.png`

test("una tesela que falla no da el plano por perdido", () => {
  let estado = marcarEstiloListo(cargaInicial())
  estado = registrarTeselaOk(estado)
  estado = registrarFalloTesela(estado, TESELA(15, 9646, 15432))
  estado = registrarFalloTesela(estado, TESELA(15, 9646, 15432))
  assert.equal(planoNoArranco(estado), false)
  assert.equal(planoNoArranco(estado, true), false)
})

test("varias teselas distintas sin ninguna buena sí tumban el plano", () => {
  let estado = marcarEstiloListo(cargaInicial())
  for (let i = 0; i < UMBRAL_TESELAS; i += 1) {
    assert.equal(planoNoArranco(estado), false)
    estado = registrarFalloTesela(estado, TESELA(15, 100 + i, 200))
  }
  assert.equal(planoNoArranco(estado), true)
})

test("al quedar en reposo sin teselas buenas el plano no arrancó", () => {
  let estado = marcarEstiloListo(cargaInicial())
  estado = registrarFalloTesela(estado, TESELA(15, 1, 1))
  assert.equal(planoNoArranco(estado), false)
  assert.equal(planoNoArranco(estado, true), true)
})

test("un clic o un elemento sin id no se clasifica como fallo de teselas", () => {
  assert.equal(clasificarErrorMapa("The feature id does not exist in the source areas"), "ignorar")
  assert.equal(clasificarErrorMapa("Style is not done loading."), "ignorar")
  assert.equal(clasificarErrorMapa("The layer actividades-circle does not exist in the map's style and cannot be queried"), "ignorar")
  assert.equal(clasificarErrorMapa(TESELA(15, 1, 2)), "tesela")
  assert.equal(clasificarErrorMapa("Failed to fetch", "osm"), "tesela")
  assert.equal(clasificarErrorMapa("Failed to load style"), "estilo")
  const ajax = Object.assign(new Error("Failed to fetch"), { url: "https://tile.openstreetmap.org/15/3/4.png" })
  assert.match(textoDeError(ajax), /tile\.openstreetmap\.org\/15\/3\/4/)
  assert.equal(clasificarErrorMapa(textoDeError(ajax)), "tesela")
})

test("el estilo roto antes de cargar también impide arrancar", () => {
  const estado = marcarEstiloRoto(cargaInicial())
  assert.equal(planoNoArranco(estado), true)
  assert.equal(planoNoArranco(marcarEstiloListo(estado)), false)
})
