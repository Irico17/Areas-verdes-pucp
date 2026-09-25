import assert from "node:assert/strict"
import test from "node:test"
import { ANCLAJES, alturaPx, anclajeCercano, siguienteAnclaje } from "./bottomSheet.ts"

test("el panel se ancla en mínima, media y casi completa, y se cierra si baja del todo", () => {
  assert.equal(anclajeCercano(0.1), null)
  assert.equal(anclajeCercano(0.3), ANCLAJES[0])
  assert.equal(anclajeCercano(0.5), ANCLAJES[1])
  assert.equal(anclajeCercano(0.9), ANCLAJES[2])
  assert.equal(siguienteAnclaje(ANCLAJES[1], 1), ANCLAJES[2])
  assert.equal(siguienteAnclaje(ANCLAJES[0], -1), null)
  assert.equal(alturaPx(0.55, 800), 440)
})
