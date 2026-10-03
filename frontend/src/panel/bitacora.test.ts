import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { Bitacora } from "./Bitacora.tsx"
import { etiquetaHito } from "../ui/nomenclatura.ts"

test("la bitácora muestra la reasignación con las dos cuadrillas y la miniatura", () => {
  const html = renderToStaticMarkup(
    createElement(Bitacora, {
      embebida: true,
      actividadId: "11111111-1111-4111-8111-111111111111",
      eventos: [
        {
          id: 7,
          tipo: "reasignada",
          actor_rol: "coordinacion",
          usuario_nombre: "Inés Calderón",
          cuadrilla_anterior: "Cuadrilla Norte",
          equipo: "Cuadrilla Sur",
          nota: "Reasignación",
          created_at: "2026-10-02T15:00:00Z",
          evidencias: [{ id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", nombre: "aspersor.jpg", mime: "image/jpeg" }],
        },
      ],
    }),
  )
  assert.match(html, /Reasignada/)
  assert.match(html, /Cuadrilla anterior/)
  assert.match(html, /Cuadrilla Norte/)
  assert.match(html, /Cuadrilla nueva/)
  assert.match(html, /Cuadrilla Sur/)
  assert.match(html, /aspersor\.jpg/)
  assert.match(html, /<img /)
  assert.equal(etiquetaHito("inicio"), "Inicio")
  assert.equal(etiquetaHito("supervision"), "Supervisión")
  assert.equal(etiquetaHito("derivacion"), "Derivación")
  assert.equal(etiquetaHito("observacion"), "Observación")
  assert.equal(etiquetaHito("conformidad"), "Conformidad")
  assert.equal(etiquetaHito("avance"), "Avance")
  assert.equal(etiquetaHito("creada"), "Creada")
})
