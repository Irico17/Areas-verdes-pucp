import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { CatastroEditor } from "./CatastroEditor.tsx"
import { areasDeFixture, zonasDeFixture } from "./catastro.ts"
import { SelectorLugar } from "./SelectorLugar.tsx"

test("el alta de área elige el lugar en un select y no ofrece texto libre", () => {
  const html = renderToStaticMarkup(
    createElement(CatastroEditor, {
      areasIniciales: areasDeFixture(),
      zonasIniciales: zonasDeFixture(),
    }),
  )
  assert.match(html, /name="lugar_id"/)
  assert.equal(html.includes('name="lugar"'), false)
  assert.equal(html.includes("lugar_libre"), false)
  assert.equal(html.includes('name="lugar_libre"'), false)
  assert.match(html, /Sectores de capataz/)
  assert.match(html, /sin archivo de cuarteles/)
})

test("un lugar ya cargado se muestra y no se convierte en una opción nueva", () => {
  const html = renderToStaticMarkup(
    createElement(SelectorLugar, {
      id: "lugar-prueba",
      lugares: [{ id: 4, nombre: "Jardín del lago" }],
      lugarId: "",
      lugarLibre: "Bancas del norte (dato viejo)",
      onChange: () => {},
    }),
  )
  assert.match(html, /<select[^>]*name="lugar_id"/)
  assert.equal((html.match(/<option/g) ?? []).length, 2)
  assert.match(html, /Lugar \(dato ya cargado\): Bancas del norte/)
  assert.equal(html.includes("<input"), false)
})
