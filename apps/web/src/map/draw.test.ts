import assert from "node:assert/strict"
import { test } from "node:test"
import { cerrarTrazo, erroresGeometria, focoEnCampo, moverVertice, type MultiPolygon } from "./draw.ts"
import { areaDesdeFeature, areasDeFixture, payloadArea, validarArea } from "../panel/catastro.ts"

const cuadrado: MultiPolygon = {
  type: "MultiPolygon",
  coordinates: [
    [
      [
        [-77.08, -12.07],
        [-77.079, -12.07],
        [-77.079, -12.0692],
        [-77.08, -12.0692],
        [-77.08, -12.07],
      ],
    ],
  ],
}

test("mover un vértice deja el anillo cerrado y entra al payload", () => {
  const area = areasDeFixture()[0]
  assert.equal(erroresGeometria(area.geom, false).length, 0)
  const movida = moverVertice(area.geom!, { polygon: 0, ring: 0, index: 1 }, [-77.0785, -12.07])
  assert.deepEqual(movida.coordinates[0][0][1], [-77.0785, -12.07])
  assert.deepEqual(movida.coordinates[0][0][0], movida.coordinates[0][0][4])
  const payload = payloadArea({ ...area, geom: movida })
  assert.equal(payload.feature_id, "AV-0004")
  assert.equal(payload.codigo, "G 13")
  assert.equal(payload.nombre, "Bosque de prueba")
  assert.equal(payload.uso, "Uso Institucional")
  assert.equal(payload.proy_riego, "Cuenta con aspersión")
  assert.equal(payload.riego_act, "Riego por aspersión")
  assert.equal(payload.referencia, "Junto al eje central")
  assert.equal(payload.perimetro_m, 180.4)
  assert.equal(payload.area_m2, 980.2)
  assert.equal(payload.zona_supervision_id, "Z1")
  assert.equal(payload.geom?.type, "MultiPolygon")
  assert.deepEqual(payload.geom?.coordinates[0][0][1], [-77.0785, -12.07])
})

test("un polígono abierto o fuera del campus no es válido", () => {
  const abierto = cerrarTrazo([
    [-77.08, -12.07],
    [-77.079, -12.07],
  ])
  assert.equal(abierto, null)
  const fuera: MultiPolygon = {
    type: "MultiPolygon",
    coordinates: [
      [
        [
          [-70, 0],
          [-70.001, 0],
          [-70.001, 0.001],
          [-70, 0],
        ],
      ],
    ],
  }
  assert.ok(erroresGeometria(fuera, true).length > 0)
  assert.deepEqual(cuadrado.coordinates[0][0][0], cuadrado.coordinates[0][0][4])
})

test("un anillo cruzado dentro del campus se rechaza", () => {
  const cruzado: MultiPolygon = {
    type: "MultiPolygon",
    coordinates: [
      [
        [
          [-77.08, -12.07],
          [-77.079, -12.0692],
          [-77.079, -12.07],
          [-77.08, -12.0692],
          [-77.08, -12.07],
        ],
      ],
    ],
  }
  const fallos = erroresGeometria(cruzado, true)
  assert.ok(fallos.some((item) => item.includes("se cruza")))
})

test("un anillo de cuatro posiciones sin cerrar se rechaza", () => {
  const abierto: MultiPolygon = {
    type: "MultiPolygon",
    coordinates: [
      [
        [
          [-77.08, -12.07],
          [-77.079, -12.07],
          [-77.079, -12.0692],
          [-77.0802, -12.0694],
        ],
      ],
    ],
  }
  const fallos = erroresGeometria(abierto, true)
  assert.ok(fallos.some((item) => item.includes("no está cerrado")))
})

test("las flechas no mueven el vértice si el foco está en un campo", () => {
  const input = { tagName: "INPUT", parentElement: null }
  const textarea = { tagName: "TEXTAREA", parentElement: null }
  const select = { tagName: "SELECT", parentElement: null }
  const editable = { tagName: "DIV", isContentEditable: true, parentElement: null }
  const mapa = { tagName: "CANVAS", parentElement: null }
  assert.equal(focoEnCampo(input as unknown as EventTarget), true)
  assert.equal(focoEnCampo(textarea as unknown as EventTarget), true)
  assert.equal(focoEnCampo(select as unknown as EventTarget), true)
  assert.equal(focoEnCampo(editable as unknown as EventTarget), true)
  assert.equal(focoEnCampo(mapa as unknown as EventTarget), false)
  assert.equal(focoEnCampo(null), false)
})

test("el feature de la fuente se lee con los nombres del contrato", () => {
  const area = areaDesdeFeature({
    properties: { Nombre: "Jardín", código: "B 4", Uso: "Uso Institucional", Perimetro: 10, Área: 20 },
    geometry: cuadrado,
  })
  const fallos = validarArea(area)
  assert.equal(fallos.length, 0)
  assert.equal(payloadArea(area).feature_id, "AV-0001")
})
