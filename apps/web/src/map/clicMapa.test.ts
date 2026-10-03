import assert from "node:assert/strict"
import { test } from "node:test"
import { decisionClic } from "./clicMapa.ts"

const POLIGONO = {
  type: "Polygon",
  coordinates: [
    [
      [-77.08, -12.07],
      [-77.07, -12.07],
      [-77.07, -12.06],
      [-77.08, -12.07],
    ],
  ],
}

test("un clic sin geometría se ignora", () => {
  assert.equal(decisionClic(null, { nombre: "Jardín Tinkuy" }), "ignorar")
  assert.equal(decisionClic({ type: "Polygon", coordinates: [] }, { nombre: "Jardín Tinkuy" }), "ignorar")
  assert.equal(decisionClic(undefined, null), "ignorar")
})

test("un clic con geometría y sin datos pide un detalle vacío", () => {
  assert.equal(decisionClic(POLIGONO, {}), "vacio")
  assert.equal(decisionClic(POLIGONO, null), "vacio")
  assert.equal(decisionClic(POLIGONO, { nombre: "  ", codigo: "" }), "vacio")
})

test("un clic con datos abre el detalle y no pide cambiar de vista", () => {
  assert.equal(decisionClic(POLIGONO, { nombre: "Jardín Tinkuy" }), "detalle")
  assert.equal(decisionClic({ type: "Point", coordinates: [-77.08, -12.07] }, { id: "act-1" }), "detalle")
})
