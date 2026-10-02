import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { FiltrosActividad } from "./FiltrosActividad.tsx"

const comunes = {
  cuadrillas: [{ id: "cap-norte", label: "Cuadrilla Norte" }],
  tipos: [{ id: "riego_manual", label: "Riego manual" }],
  riesgos: [{ id: "alto", label: "Alto" }],
  origenes: [{ id: "interna", label: "Interna" }],
}

test("la oficina ve los ocho criterios", () => {
  const html = renderToStaticMarkup(createElement(FiltrosActividad, { rol: "coordinacion", ...comunes }))
  for (const etiqueta of ["Estado", "Tipo de actividad", "Cuadrilla", "Sector de capataz", "Quién ejecuta", "Origen", "Nivel de riesgo", "Desde", "Hasta"]) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
  assert.match(html, /name="cuadrilla_id"/)
  assert.match(html, /name="nivel_riesgo"/)
  assert.match(html, /name="desde"/)
})

test("el capataz no ve el filtro de cuadrilla ajena", () => {
  const html = renderToStaticMarkup(createElement(FiltrosActividad, { rol: "capataz", ...comunes }))
  assert.equal(html.includes('name="cuadrilla_id"'), false)
  assert.equal(html.includes("Cuadrilla Norte"), false)
  assert.match(html, /name="sector"/)
  assert.match(html, /Incluir cerradas/)
})
