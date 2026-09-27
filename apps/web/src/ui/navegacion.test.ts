import assert from "node:assert/strict"
import test from "node:test"
import { repartirModulos } from "./navegacion.ts"

test("la barra del teléfono deja cuatro módulos y manda el resto a Más", () => {
  assert.deepEqual(repartirModulos(["a", "b", "c", "d", "e"]), { barra: ["a", "b", "c", "d", "e"], mas: [] })
  const nueve = repartirModulos(["1", "2", "3", "4", "5", "6", "7", "8", "9"])
  assert.deepEqual(nueve.barra, ["1", "2", "3", "4"])
  assert.deepEqual(nueve.mas, ["5", "6", "7", "8", "9"])
})
