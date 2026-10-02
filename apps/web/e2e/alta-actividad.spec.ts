import { expect, test } from "@playwright/test"

test("el alta de una actividad muestra clase, tipo, riesgo y lugar de catálogo", async ({ page }) => {
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: "Nueva actividad" }).click()
  await expect(page.getByText("Toque el mapa donde está el trabajo.")).toBeVisible()

  const mapa = page.locator(".maplibregl-canvas")
  await mapa.waitFor({ state: "visible" })
  await page.waitForTimeout(800)
  await mapa.click({ position: { x: 72, y: 280 }, force: true })

  const alta = page.locator("form.alta-actividad")
  await expect(alta).toBeVisible()
  await expect(alta.getByLabel("Clase de actividad")).toBeVisible()
  await expect(alta.getByLabel("Tipo de actividad")).toBeVisible()
  await expect(alta.getByLabel("Origen")).toBeVisible()
  await expect(alta.getByLabel("Código externo")).toBeVisible()
  await expect(alta.getByLabel("Unidad solicitante")).toBeVisible()
  await expect(alta.getByLabel("Nivel de riesgo")).toBeVisible()
  await expect(alta.getByLabel("Fecha programada")).toBeVisible()
  await expect(alta.getByLabel("Cantidad")).toBeVisible()
  await expect(alta.getByText("Personal de la actividad")).toBeVisible()
  await expect(alta.getByText("Nombres ficticios. No son cuentas.")).toBeVisible()
  await expect(alta.getByLabel("Zona de supervisión")).toBeVisible()
  await expect(alta.locator('select[name="lugar_id"]')).toBeVisible()
  await expect(alta.locator('input[name="lugar"]')).toHaveCount(0)
  await expect(alta.locator('input[name="lugar_libre"]')).toHaveCount(0)

  const riesgo = alta.getByLabel("Nivel de riesgo")
  await expect(riesgo.locator("option")).toContainText(["Bajo", "Alto"])
  await expect(riesgo.locator("option", { hasText: "Medio" })).toHaveCount(0)

  await alta.getByLabel("Clase de actividad").selectOption("riego")
  await expect(alta.getByLabel("Tipo de actividad").locator("option", { hasText: "Riego manual" })).toHaveCount(1)
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/alta-actividad-campos.png" })
  await alta.getByText("Personal de la actividad").scrollIntoViewIfNeeded()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/alta-actividad-pedido.png" })
})
