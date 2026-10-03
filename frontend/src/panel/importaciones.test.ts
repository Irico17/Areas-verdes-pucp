import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { ImportacionesPanel } from "./Importaciones.tsx"
import { ENTIDADES } from "./importaciones.ts"

test("el importador nombra las entidades del mapa y pide confirmación", () => {
  assert.equal(ENTIDADES.some((item) => item.id === "lugares"), true)
  assert.equal(ENTIDADES.some((item) => item.id === "indicadores"), false)
  const html = renderToStaticMarkup(createElement(ImportacionesPanel))
  for (const texto of ["Revisar archivo", "Confirmar importación", "Revertir esta importación", "Territorio", "Lugares", "Puntos del campus", "Actividades", "Puertas y entradas", "Playas de estacionamiento"]) {
    assert.ok(html.includes(texto), texto)
  }
})
