import assert from "node:assert/strict"
import test from "node:test"
import { modulosDe, modulosPorPermisos, moduloPermitido } from "./registroModulos.ts"

test("las pestañas salen de los permisos y el capataz no ve el editor de oficina", () => {
  const capataz = modulosDe("capataz")
  assert.deepEqual(capataz.slice(0, 4), ["hoy", "mapa", "labores", "registros"])
  assert.equal(capataz.includes("catastro"), false)
  assert.equal(capataz.includes("inventario"), false)
  assert.equal(capataz.includes("historial"), false)
  assert.equal(capataz.includes("admin"), false)
  assert.equal(capataz.includes("bitacora"), true)
  assert.equal(capataz.includes("ejemplares"), true)

  const jefatura = modulosDe("jefatura")
  assert.equal(jefatura[0], "resumen")
  assert.equal(jefatura.includes("registros"), false)
  assert.equal(jefatura.includes("inventario"), false)
  assert.equal(jefatura.includes("catastro"), true)
  assert.equal(jefatura.includes("historial"), true)
  assert.equal(jefatura.includes("admin"), true)
  assert.equal(jefatura.includes("importaciones"), true)
  assert.equal(jefatura.includes("catalogos"), false)

  const coordinacion = modulosDe("coordinacion")
  assert.equal(coordinacion.includes("admin"), false)
  assert.equal(coordinacion.includes("registros"), true)
  assert.equal(coordinacion.includes("catalogos"), true)
  assert.equal(coordinacion.includes("historial"), true)
  assert.equal(coordinacion.at(-1), "historial")

  const admin = modulosDe("admin")
  assert.equal(admin.includes("admin"), true)
  assert.equal(admin.includes("historial"), true)
  assert.equal(admin.includes("registros"), true)
})

test("sin el permiso registrar no hay registros de campo ni inventario", () => {
  const soloConsulta = modulosPorPermisos(["consultar", "validar", "reportes"], "jefatura")
  assert.equal(moduloPermitido("registros", ["consultar", "validar"]), false)
  assert.equal(soloConsulta.includes("registros"), false)
  assert.equal(soloConsulta.includes("inventario"), false)
  assert.equal(soloConsulta.includes("historial"), true)
})
