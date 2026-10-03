import assert from "node:assert/strict"
import test from "node:test"
import { barraMovil, repartirModulos } from "./navegacion.ts"

test("la barra del teléfono deja cuatro módulos y manda el resto a Más", () => {
  assert.deepEqual(repartirModulos(["a", "b", "c", "d", "e"]), { barra: ["a", "b", "c", "d", "e"], mas: [] })
  const nueve = repartirModulos(["1", "2", "3", "4", "5", "6", "7", "8", "9"])
  assert.deepEqual(nueve.barra, ["1", "2", "3", "4"])
  assert.deepEqual(nueve.mas, ["5", "6", "7", "8", "9"])
})

test("el capataz en el teléfono no abre Más y la oficina sí agrupa el resto", () => {
  const campo = [
    { id: "hoy" },
    { id: "mapa" },
    { id: "labores" },
    { id: "registros" },
    { id: "bitacora" },
    { id: "ejemplares" },
  ]
  const capataz = barraMovil("capataz", campo)
  assert.deepEqual(capataz.barra.map((item) => item.id), ["hoy", "mapa", "labores", "registros"])
  assert.deepEqual(capataz.mas, [])

  const oficina = barraMovil("coordinacion", [
    { id: "resumen" },
    { id: "mapa" },
    { id: "labores" },
    { id: "solicitudes" },
    { id: "reportes" },
    { id: "historial" },
  ])
  assert.deepEqual(oficina.barra.map((item) => item.id), ["resumen", "mapa", "labores", "solicitudes"])
  assert.deepEqual(oficina.mas.map((item) => item.id), ["reportes", "historial"])
})
