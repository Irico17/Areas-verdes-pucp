import { expect, test } from "@playwright/test"

test("el riego elige el sector de capataz y muestra la cobertura provisional", async ({ page }) => {
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.getByRole("button", { name: "Registros de campo" }).click()
  await page.getByRole("button", { name: "Nuevo registro" }).click()
  const riego = page.locator("#riego")
  await riego.scrollIntoViewIfNeeded()
  await expect(riego.getByRole("heading", { name: "Riego" })).toBeVisible()
  await expect(riego.getByRole("status")).toHaveText(
    /Cobertura provisional: \d+ %\. Pendiente de validar con la jefatura de sección\./,
  )
  await expect(riego.getByText("definición pendiente")).toHaveCount(0)
  await expect(riego.locator('select[name="sector_id"]')).toBeVisible()
  await expect(riego.locator('input[name="sector"]')).toHaveCount(0)
  await expect(riego.locator('select[name="sector_id"] option')).toContainText(["Bosque húmedo"])
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/riego-cobertura-provisional.png", fullPage: true })
  await riego.screenshot({ path: "/opt/cursor/artifacts/screenshots/riego-panel-cobertura.png" })
})
