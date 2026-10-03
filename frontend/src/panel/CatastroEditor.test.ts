import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { CatastroEditor } from "./CatastroEditor.tsx"
import { areasDeFixture, zonasDeFixture } from "./catastro.ts"

const etiquetas = [
  "Identificador",
  "Código",
  "Nombre",
  "Uso",
  "Proyecto de riego",
  "Riego actual",
  "Referencia",
  "Perímetro (m)",
  "Área (m²)",
  "Zona de supervisión",
  "Geometría",
  "Editar geometría",
]

test("la ficha de área lista los campos de 4.1", async () => {
  const html = renderToStaticMarkup(
    createElement(CatastroEditor, {
      areasIniciales: areasDeFixture(),
      zonasIniciales: zonasDeFixture(),
    }),
  )
  assert.match(html, /Bosque de prueba/)
  assert.match(html, /AV-0004/)
  for (const etiqueta of etiquetas) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
})

test("la ficha de zona lista código, nombre, área y geometría", async () => {
  const html = renderToStaticMarkup(
    createElement(CatastroEditor, {
      areasIniciales: areasDeFixture(),
      zonasIniciales: zonasDeFixture(),
      entidadInicial: "zona",
    }),
  )
  assert.match(html, /Zonas/)
  assert.match(html, />Z1</)
})
