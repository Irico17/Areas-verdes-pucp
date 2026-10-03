import assert from "node:assert/strict"
import { test } from "node:test"
import { seleccionarActividad } from "./seleccionActividad.ts"

test("elegir un marcador no cambia la pestaña de trabajo", () => {
  const vistos: string[] = []
  seleccionarActividad("abc", {
    elegir: () => vistos.push("elegir"),
    irALabores: () => vistos.push("pestaña"),
    abrirPanel: () => vistos.push("panel"),
    limpiar: () => vistos.push("limpiar"),
  })
  assert.deepEqual(vistos, ["elegir", "panel"])
})

test("sin id limpia la selección", () => {
  let limpio = false
  seleccionarActividad(null, {
    elegir: () => {},
    irALabores: () => {},
    abrirPanel: () => {},
    limpiar: () => {
      limpio = true
    },
  })
  assert.equal(limpio, true)
})
