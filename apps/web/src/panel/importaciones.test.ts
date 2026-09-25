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
  for (const texto of ["Vista previa", "Confirmar escritura", "Revertir lote", "Lugares", "Puntos PUCP"]) {
    assert.ok(html.includes(texto), texto)
  }
})
