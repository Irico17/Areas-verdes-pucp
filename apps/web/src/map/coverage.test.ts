import assert from "node:assert/strict"
import { test } from "node:test"
import { catastroVisible, conteoCapas } from "./coverage.ts"

test("el mapa pinta áreas y zonas cuando la API las trae", () => {
  const conteo = conteoCapas({
    areas: { features: Array.from({ length: 521 }, () => ({})) },
    zonas: { features: Array.from({ length: 534 }, () => ({})) },
  })
  assert.equal(conteo.areas, 521)
  assert.equal(conteo.zonas, 534)
  assert.equal(catastroVisible(conteo), true)
})

test("una colección vacía no cuenta como mapa cargado", () => {
  assert.equal(catastroVisible(conteoCapas({})), false)
})
