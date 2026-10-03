import { expect, test } from "./fixtures"

test("el historial muestra un cambio de la semilla y dice que no hay retención", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("jefatura")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  await page.getByRole("button", { name: "Historial", exact: true }).click()
  await expect(page.getByRole("heading", { name: "Historial" })).toBeVisible()
  await expect(page.getByText("No hay un plazo de retención.")).toBeVisible()
  await expect(page.getByText("Este historial no se borra por antigüedad.")).toBeVisible()
  await expect(page.locator(".historial-lista").getByText("Equipo Norte").first()).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/capturas/historial-semilla.png" })
})
