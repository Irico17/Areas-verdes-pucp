import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { iconoDe } from "./trazosClase.ts"
import { IconoClase } from "./iconoClase.tsx"

test("el icono sale de la clase y, si falta, del tipo, nunca de una letra", () => {
  assert.equal(iconoDe("riego", "limpieza"), "riego")
  assert.equal(iconoDe("", "limpieza"), "mantenimiento")
  assert.equal(iconoDe("", "inspeccion"), "inspeccion_monitoreo")
  assert.equal(iconoDe("", ""), "generica")
  const html = renderToStaticMarkup(createElement(IconoClase, { tipo: "poda" }))
  assert.match(html, /data-icono="poda"/)
  assert.match(html, /<path /)
  assert.equal(html.includes(">P<"), false)
  assert.equal(html.includes(">R<"), false)
})
