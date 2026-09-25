import assert from "node:assert/strict"
import { test } from "node:test"
import {
  CAMPOS_AREA,
  CAMPOS_ZONA,
  validarArea,
  validarZona,
  zonasDeFixture,
  type AreaVerde,
} from "./catastro.ts"

const area: AreaVerde = {
  feature_id: "AV-0008",
  codigo: null,
  nombre: null,
  uso: null,
  proy_riego: null,
  riego_act: null,
  referencia: null,
  perimetro_m: null,
  area_m2: null,
  geom: null,
  zona_supervision_id: null,
}

test("el contrato de área y de zona no inventa campos", () => {
  assert.deepEqual(CAMPOS_AREA, [
    "feature_id",
    "codigo",
    "nombre",
    "uso",
    "proy_riego",
    "riego_act",
    "referencia",
    "perimetro_m",
    "area_m2",
    "geom",
    "zona_supervision_id",
  ])
  assert.deepEqual(CAMPOS_ZONA, ["codigo", "nombre", "area_m2", "geom"])
})

test("un área sin identificador y una zona sin nombre no pasan", () => {
  assert.ok(validarArea({ ...area, feature_id: "zona-1" }).some((item) => item.campo === "feature_id"))
  const zona = { ...zonasDeFixture()[0], nombre: "  ", geom: null }
  const fallos = validarZona(zona)
  assert.ok(fallos.some((item) => item.campo === "nombre"))
  assert.ok(fallos.some((item) => item.campo === "geom"))
})

test("la referencia no pasa de 500 caracteres y el perímetro no es negativo", () => {
  const larga = validarArea({ ...area, referencia: "a".repeat(501) })
  assert.ok(larga.some((item) => item.campo === "referencia"))
  const negativa = validarArea({ ...area, perimetro_m: -1 })
  assert.ok(negativa.some((item) => item.campo === "perimetro_m"))
})
