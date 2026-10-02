import assert from "node:assert/strict"
import { test } from "node:test"
import {
  categoriaSector,
  categoriaUso,
  conCategorias,
  conteoPorCategoria,
  expresionColor,
  filtroCategorias,
  SECTORES,
  USOS,
} from "./categorias.ts"
import type { FeatureCollection, GeoFeature } from "../types.ts"

test("categoriaUso separa los seis textos de la propiedad Uso", () => {
  assert.equal(categoriaUso("Uso Institucional"), "institucional")
  assert.equal(categoriaUso("Áreas de uso administrativo"), "administrativo")
  assert.equal(categoriaUso("Áreas de uso recreativo/descanso"), "recreativo")
  assert.equal(categoriaUso("Áreas de manejo sostenible y reducción de consumo de agua"), "sostenible")
  assert.equal(categoriaUso("Áreas deportivas y recreación activa"), "deportivas")
  assert.equal(categoriaUso("Áreas de conservación"), "conservacion")
  for (const cat of USOS.filter((item) => item.id !== "sin-uso")) {
    assert.equal(cat.label.includes("Otros"), false, cat.id)
    assert.equal(categoriaUso(cat.label), cat.id)
  }
})

test("categoriaUso tolera nulos, vacíos, mayúsculas y tildes", () => {
  assert.equal(categoriaUso(null), "sin-uso")
  assert.equal(categoriaUso(""), "sin-uso")
  assert.equal(categoriaUso("INSTITUCIONAL"), "institucional")
  assert.equal(categoriaUso("sosténible"), "sostenible")
})

test("categoriaSector acepta los 5 slugs y cae a sin-sector", () => {
  for (const cat of SECTORES.filter((c) => c.id !== "sin-sector")) {
    assert.equal(categoriaSector(cat.id), cat.id)
  }
  assert.equal(categoriaSector("cua-desconocido"), "sin-sector")
  assert.equal(categoriaSector(null), "sin-sector")
  assert.equal(categoriaSector(undefined), "sin-sector")
})

function feature(props: Record<string, unknown>): GeoFeature {
  return { type: "Feature", geometry: null, properties: props }
}

test("conCategorias no muta la entrada y agrega el campo correcto", () => {
  const fc: FeatureCollection = {
    type: "FeatureCollection",
    features: [feature({ uso: "Áreas de uso administrativo" }), feature({ sector: "cua-mateo" })],
  }
  const original = JSON.parse(JSON.stringify(fc))
  const porUso = conCategorias(fc, "uso")
  assert.deepEqual(fc, original)
  assert.equal(porUso.features[0].properties?.cat_uso, "administrativo")
  assert.equal(porUso.features[0].properties?.cat_sector, undefined)

  const porSector = conCategorias(fc, "sector")
  assert.equal(porSector.features[1].properties?.cat_sector, "cua-mateo")
})

test("conCategorias sin datos devuelve una colección vacía", () => {
  assert.deepEqual(conCategorias(undefined, "uso"), { type: "FeatureCollection", features: [] })
})

test("conteoPorCategoria incluye ceros y suma el total", () => {
  const fc: FeatureCollection = {
    type: "FeatureCollection",
    features: [feature({ cat_uso: "deportivas" }), feature({ cat_uso: "conservacion" }), feature({ cat_uso: "institucional" })],
  }
  const conteo = conteoPorCategoria(fc, "cat_uso", USOS)
  assert.equal(conteo.deportivas, 1)
  assert.equal(conteo.conservacion, 1)
  assert.equal(conteo.institucional, 1)
  assert.equal(conteo.administrativo, 0)
  const total = Object.values(conteo).reduce((a, b) => a + b, 0)
  assert.equal(total, 3)
})

test("expresionColor cubre todos los ids y termina en un color de respaldo", () => {
  const expr = expresionColor("cat_uso", USOS, "fill") as unknown[]
  assert.equal(expr[0], "match")
  assert.deepEqual(expr[1], ["get", "cat_uso"])
  for (const cat of USOS) {
    assert.ok(expr.includes(cat.id), `falta ${cat.id}`)
    assert.ok(expr.includes(cat.fill), `falta el color de ${cat.id}`)
  }
  assert.equal(expr[expr.length - 1], USOS[USOS.length - 1].fill)
})

test("filtroCategorias devuelve null cuando están todas activas", () => {
  const todos = USOS.map((c) => c.id)
  assert.equal(filtroCategorias("cat_uso", todos, todos), null)
  assert.equal(filtroCategorias("cat_uso", [...todos].reverse(), todos), null)
})

test("filtroCategorias arma un filtro in cuando falta alguna", () => {
  const todos = USOS.map((c) => c.id)
  const filtro = filtroCategorias("cat_uso", ["deportivas"], todos)
  assert.deepEqual(filtro, ["in", ["get", "cat_uso"], ["literal", ["deportivas"]]])
})

function hexALab(hex: string): [number, number, number] {
  const n = Number.parseInt(hex.slice(1), 16)
  let [r, g, b] = [(n >> 16) & 255, (n >> 8) & 255, n & 255].map((v) => v / 255)
  ;[r, g, b] = [r, g, b].map((v) => (v > 0.04045 ? ((v + 0.055) / 1.055) ** 2.4 : v / 12.92))
  const x = (r * 0.4124 + g * 0.3576 + b * 0.1805) / 0.95047
  const y = r * 0.2126 + g * 0.7152 + b * 0.0722
  const z = (r * 0.0193 + g * 0.1192 + b * 0.9505) / 1.08883
  const f = (t: number) => (t > 0.008856 ? Math.cbrt(t) : 7.787 * t + 16 / 116)
  const [fx, fy, fz] = [f(x), f(y), f(z)]
  return [116 * fy - 16, 500 * (fx - fy), 200 * (fy - fz)]
}

function deltaE(hexA: string, hexB: string): number {
  const [l1, a1, b1] = hexALab(hexA)
  const [l2, a2, b2] = hexALab(hexB)
  return Math.sqrt((l1 - l2) ** 2 + (a1 - a2) ** 2 + (b1 - b2) ** 2)
}

test("la paleta de uso tiene colores separados entre sí y de la marca/selección", () => {
  for (let i = 0; i < USOS.length; i++) {
    for (let j = i + 1; j < USOS.length; j++) {
      assert.ok(deltaE(USOS[i].fill, USOS[j].fill) >= 20, `${USOS[i].id} vs ${USOS[j].id}`)
    }
    assert.ok(deltaE(USOS[i].fill, "#083465") >= 15, `${USOS[i].id} vs selección`)
    assert.ok(deltaE(USOS[i].fill, "#308046") >= 15, `${USOS[i].id} vs marca`)
  }
})

test("la paleta de sector tiene colores separados entre sí y de la marca/selección", () => {
  for (let i = 0; i < SECTORES.length; i++) {
    for (let j = i + 1; j < SECTORES.length; j++) {
      assert.ok(deltaE(SECTORES[i].fill, SECTORES[j].fill) >= 20, `${SECTORES[i].id} vs ${SECTORES[j].id}`)
    }
    assert.ok(deltaE(SECTORES[i].fill, "#083465") >= 15, `${SECTORES[i].id} vs selección`)
    assert.ok(deltaE(SECTORES[i].fill, "#308046") >= 15, `${SECTORES[i].id} vs marca`)
  }
})
