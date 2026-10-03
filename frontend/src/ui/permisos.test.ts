import assert from "node:assert/strict"
import test from "node:test"
import { MATRIZ_SEMILLA, permisosDeRol, tienePermiso } from "./permisos.ts"

test("la matriz del front reproduce la semilla del backend", () => {
  assert.deepEqual(permisosDeRol("capataz"), ["consultar", "registrar"])
  assert.deepEqual(permisosDeRol("coordinacion"), ["consultar", "registrar", "validar", "solicitudes", "reportes"])
  assert.deepEqual(permisosDeRol("jefatura"), ["consultar", "validar", "reportes", "solicitudes", "evidencias", "usuarios"])
  assert.deepEqual(permisosDeRol("admin"), ["consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes", "usuarios"])
  assert.deepEqual(MATRIZ_SEMILLA.jefatura.includes("usuarios"), true)
  assert.equal(tienePermiso(permisosDeRol("capataz"), "validar"), false)
  assert.equal(tienePermiso(permisosDeRol("jefatura"), "registrar"), false)
  assert.equal(tienePermiso(permisosDeRol("admin"), "catalogos"), true)
  assert.deepEqual(permisosDeRol("otro"), [])
})
