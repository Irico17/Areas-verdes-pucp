import assert from "node:assert/strict"
import { test } from "node:test"
import { clavePanelMapa, guardarPreferenciaPanel, leerPreferenciaPanel, panelAbiertoInicial } from "./panelMapa.ts"

function memoria(): Storage {
  const datos = new Map<string, string>()
  return {
    get length() {
      return datos.size
    },
    clear() {
      datos.clear()
    },
    getItem(clave) {
      return datos.get(clave) ?? null
    },
    key(indice) {
      return [...datos.keys()][indice] ?? null
    },
    removeItem(clave) {
      datos.delete(clave)
    },
    setItem(clave, valor) {
      datos.set(clave, String(valor))
    },
  }
}

test("la preferencia del panel se guarda por persona y por ancho", () => {
  const caja = memoria()
  const anterior = globalThis.localStorage
  Object.defineProperty(globalThis, "localStorage", { configurable: true, value: caja })
  try {
    assert.equal(panelAbiertoInicial("coordinacion", true), false)
    assert.equal(panelAbiertoInicial("coordinacion", false), true)
    guardarPreferenciaPanel("coordinacion", false, false)
    guardarPreferenciaPanel("norte", false, true)
    assert.equal(leerPreferenciaPanel("coordinacion", false), false)
    assert.equal(leerPreferenciaPanel("norte", false), true)
    assert.equal(leerPreferenciaPanel("coordinacion", true), null)
    assert.equal(panelAbiertoInicial("coordinacion", true), false)
    assert.equal(clavePanelMapa(" coordinacion ", true), "cv:panel-mapa:coordinacion:estrecho")
  } finally {
    Object.defineProperty(globalThis, "localStorage", { configurable: true, value: anterior })
  }
})

test("sin almacenamiento el panel usa el valor por ancho y no lanza", () => {
  const anterior = Object.getOwnPropertyDescriptor(globalThis, "localStorage")
  Object.defineProperty(globalThis, "localStorage", {
    configurable: true,
    get() {
      throw new Error("sin almacenamiento")
    },
  })
  try {
    assert.equal(panelAbiertoInicial("coordinacion", true), false)
    assert.equal(panelAbiertoInicial("coordinacion", false), true)
    assert.doesNotThrow(() => guardarPreferenciaPanel("coordinacion", false, false))
    assert.equal(leerPreferenciaPanel("coordinacion", false), null)
  } finally {
    if (anterior) Object.defineProperty(globalThis, "localStorage", anterior)
    else delete (globalThis as { localStorage?: Storage }).localStorage
  }
})
