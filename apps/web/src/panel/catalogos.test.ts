import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { CatalogosPanel, ListaCatalogo } from "./Catalogos.tsx"
import { aplicarEstadosCatalogo, etiquetaEstado, ESTADOS } from "../operacion.ts"

test("el panel ofrece las clases que faltaban y no borra al desactivar", () => {
  const html = renderToStaticMarkup(createElement(CatalogosPanel, { editable: true }))
  for (const texto of [
    "Clases de actividad",
    "Plagas",
    "Productos fitosanitarios",
    "Frecuencias",
    "Sedes",
    "Cuarteles (histórico)",
    "Sectores de capataz",
    "provisional",
  ]) {
    assert.ok(html.includes(texto), texto)
  }
})

test("se puede corregir el nombre de un ítem", () => {
  const html = renderToStaticMarkup(
    createElement(ListaCatalogo, {
      editable: true,
      items: [
        {
          id: 4,
          clase: "estado",
          codigo: "pendiente",
          nombre: "Por iniciar",
          activo: true,
          orden: 1,
          provisional: false,
        },
      ],
      onRenombrar: async () => {},
      onDesactivar: async () => {},
    }),
  )
  assert.match(html, /Corregir nombre/)
  assert.match(html, /Por iniciar/)
  assert.match(html, /<small>pendiente<\/small>/)
})

test("las etiquetas de estado salen del catálogo y no del slug", () => {
  assert.equal(etiquetaEstado("pendiente"), "Por iniciar")
  assert.equal(etiquetaEstado("ejecutado"), "Ejecutado")
  assert.equal(etiquetaEstado("bloqueada"), "Bloqueada (provisional)")
  assert.equal(etiquetaEstado("cerrada"), "Cerrado")
  assert.equal(etiquetaEstado("cancelada"), "Cancelado")
  assert.equal(etiquetaEstado("archivada"), "Archivado")
  aplicarEstadosCatalogo([
    { codigo: "ejecutado", nombre: "Ejecutado en campo", activo: true, orden: 3 },
    { codigo: "bloqueada", nombre: "Bloqueada", activo: false, orden: 90, provisional: true },
  ])
  assert.equal(etiquetaEstado("ejecutado"), "Ejecutado en campo")
  assert.equal(ESTADOS.some((item) => item.id === "bloqueada"), false)
  assert.equal(etiquetaEstado("bloqueada"), "Bloqueada (provisional)")
  aplicarEstadosCatalogo([
    { codigo: "pendiente", nombre: "Por iniciar", activo: true, orden: 1 },
    { codigo: "en_proceso", nombre: "En proceso", activo: true, orden: 2 },
    { codigo: "ejecutado", nombre: "Ejecutado", activo: true, orden: 3 },
    { codigo: "cerrada", nombre: "Cerrado", activo: true, orden: 4 },
    { codigo: "cancelada", nombre: "Cancelado", activo: true, orden: 5 },
    { codigo: "archivada", nombre: "Archivado", activo: true, orden: 6 },
    { codigo: "sin_estado", nombre: "Sin estado", activo: true, orden: 0 },
  ])
})
