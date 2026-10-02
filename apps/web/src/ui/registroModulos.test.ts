import assert from "node:assert/strict"
import test from "node:test"
import { modulosDe } from "./registroModulos.ts"

test("la jefatura administra cuentas y el capataz no ve esa pestaña", () => {
  assert.equal(modulosDe("jefatura").includes("admin"), true)
  assert.equal(modulosDe("admin").includes("admin"), true)
  assert.equal(modulosDe("capataz").includes("admin"), false)
  assert.equal(modulosDe("coordinacion").includes("admin"), false)
})
