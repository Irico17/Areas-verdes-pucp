import assert from "node:assert/strict"
import test from "node:test"
import { altoTapado, tecladoAbierto } from "./teclado.ts"

test("el teclado cuenta si tapa más de 120 px, sin zoom y con un campo enfocado", () => {
  assert.equal(altoTapado(844, { height: 500, offsetTop: 0 }), 344)
  assert.equal(altoTapado(844, { height: 844, offsetTop: 0 }), 0)
  assert.equal(tecladoAbierto(344, 1, true), true)
  assert.equal(tecladoAbierto(344, 1, false), false)
  assert.equal(tecladoAbierto(80, 1, true), false)
  assert.equal(tecladoAbierto(344, 1.6, true), false)
})
