import assert from "node:assert/strict"
import test from "node:test"
import { CALIDAD_INICIAL, LADO_LARGO, TOPE_CLIENTE, medida, rechazarPorTamano } from "./comprimir.ts"

test("el lado largo baja a 1600 px", () => {
  assert.deepEqual(medida(3200, 1600), { ancho: 1600, alto: 800 })
  assert.deepEqual(medida(800, 400), { ancho: 800, alto: 400 })
  assert.equal(LADO_LARGO, 1600)
  assert.equal(CALIDAD_INICIAL, 0.7)
  assert.equal(TOPE_CLIENTE, 1_572_864)
})

test("por encima de 8 MB se rechaza antes de salir", () => {
  assert.equal(rechazarPorTamano(8 * 1024 * 1024 + 1), true)
  assert.equal(rechazarPorTamano(TOPE_CLIENTE), false)
})
