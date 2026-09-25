import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { ReportesPanel } from "./Modulos.tsx"
import { reporteQuery } from "../producto.ts"

test("el reporte básico arma zona, cuadrilla, origen y fechas", () => {
  const q = reporteQuery(
    { zona: "Z1", cuadrilla: "Nora Beltrán", origen: "monitoreo", desde: "2026-05-01", hasta: "2026-05-31", estado: "cerrada" },
    "csv",
  )
  const params = new URLSearchParams(q)
  assert.equal(params.get("formato"), "csv")
  assert.equal(params.get("zona"), "Z1")
  assert.equal(params.get("cuadrilla"), "Nora Beltrán")
  assert.equal(params.get("origen"), "monitoreo")
  assert.equal(params.get("desde"), "2026-05-01")
  assert.equal(params.get("hasta"), "2026-05-31")
  assert.equal(params.has("pdf"), false)
})

test("el panel de reportes ofrece zona, cuadrilla, origen y fechas", () => {
  const html = renderToStaticMarkup(createElement(ReportesPanel))
  for (const texto of ["Zona", "Cuadrilla", "Origen", "Desde", "Hasta"]) {
    assert.ok(html.includes(texto), texto)
  }
  assert.equal(html.includes("PDF"), false)
  for (const texto of ["Cobertura", "Rendimiento", "Métricas de proveedor", "definición pendiente"]) {
    assert.ok(html.includes(texto), texto)
  }
  assert.equal(html.includes("%"), false)
})
