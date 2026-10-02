import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { FILTRO_ACTIVIDADES } from "../operacion.ts"
import { FiltrosActividad } from "./FiltrosActividad.tsx"

const comunes = {
  cuadrillas: [{ id: "cap-norte", label: "Cuadrilla Norte" }],
  tipos: [{ id: "riego_manual", label: "Riego manual" }],
  riesgos: [{ id: "alto", label: "Alto" }],
  origenes: [{ id: "interna", label: "Interna" }],
}

test("la oficina ve los ocho criterios", () => {
  const html = renderToStaticMarkup(createElement(FiltrosActividad, { rol: "coordinacion", ...comunes }))
  for (const etiqueta of ["Estado", "Tipo de actividad", "Cuadrilla", "Sector de capataz", "Quién ejecuta", "Origen", "Nivel de riesgo", "Desde", "Hasta", "Ejemplar", "Personal de la actividad"]) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
  assert.match(html, /name="cuadrilla_id"/)
  assert.match(html, /name="nivel_riesgo"/)
  assert.match(html, /name="desde"/)
})

test("los contadores de estado y Quitar filtros usan el número visible", () => {
  const html = renderToStaticMarkup(
    createElement(FiltrosActividad, {
      rol: "coordinacion",
      ...comunes,
      conteos: { pendiente: 3, en_proceso: 1, bloqueada: 1 },
      valor: { ...FILTRO_ACTIVIDADES, estado: "pendiente", tipo: "riego_manual" },
    }),
  )
  assert.match(html, /Por iniciar 3/)
  assert.match(html, /En proceso 1/)
  assert.match(html, /Bloqueada \(provisional\) 1/)
  assert.match(html, /Quitar filtros/)
  assert.match(html, /aria-pressed="true"/)
  assert.match(html, /Riego manual/)
})

test("el capataz no ve el filtro de cuadrilla ajena", () => {
  const html = renderToStaticMarkup(createElement(FiltrosActividad, { rol: "capataz", ...comunes }))
  assert.equal(html.includes('name="cuadrilla_id"'), false)
  assert.equal(html.includes("Cuadrilla Norte"), false)
  assert.match(html, /name="sector"/)
  assert.match(html, /name="ejemplar_id"/)
  assert.match(html, /name="responsable"/)
  assert.match(html, /Incluir cerradas/)
})
