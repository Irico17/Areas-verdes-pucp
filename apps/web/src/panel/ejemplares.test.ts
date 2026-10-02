import assert from "node:assert/strict"
import test from "node:test"
import { cuerpoFicha, etiquetaSalud, lineaCodigo, puedeEditarEjemplar, tituloLista } from "./ejemplares.ts"

test("el capataz consulta y coordinación edita", () => {
  assert.equal(puedeEditarEjemplar("capataz"), false)
  assert.equal(puedeEditarEjemplar("jefatura"), false)
  assert.equal(puedeEditarEjemplar("coordinacion"), true)
  assert.equal(puedeEditarEjemplar("admin"), true)
})

test("el historial nombra el código anterior", () => {
  assert.equal(lineaCodigo("FL-100", "FL-200"), "FL-100 pasó a FL-200")
  assert.equal(etiquetaSalud("bueno"), "Bueno")
  assert.equal(etiquetaSalud(null), "Sin dato")
  assert.equal(tituloLista({ id: 1, codigo: "FL-1", nombre_comun: "", tipo_vegetacion: "Árbol", cantidad: 1, activo: true }), "FL-1")
})

test("la ficha manda salud y deja el sector vacío si no hay selección", () => {
  const listo = cuerpoFicha({
    salud: "regular",
    especieId: "",
    lugarId: "",
    sectorId: "4",
    lat: "-12.07",
    lon: "-77.08",
  })
  assert.equal(listo.ok, true)
  if (!listo.ok) return
  assert.equal(listo.cuerpo.salud, "regular")
  assert.equal(listo.cuerpo.sector_cuartel_id, 4)
  assert.equal(listo.cuerpo.especie_id, null)

  const roto = cuerpoFicha({ salud: "", especieId: "", lugarId: "", sectorId: "", lat: "-12.07", lon: "" })
  assert.equal(roto.ok, false)
})
