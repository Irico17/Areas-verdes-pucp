import assert from "node:assert/strict"
import { test } from "node:test"
import { cerrarTrazo, erroresGeometria, moverVertice, type MultiPolygon } from "./draw.ts"
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

test("el feature de la fuente se lee con los nombres del contrato", () => {
  const area = areaDesdeFeature({
    properties: { Nombre: "Jardín", código: "B 4", Uso: "Uso Institucional", Perimetro: 10, Área: 20 },
    geometry: cuadrado,
  })
  const fallos = validarArea(area)
  assert.equal(fallos.length, 0)
  assert.equal(payloadArea(area).feature_id, "AV-0001")
})
