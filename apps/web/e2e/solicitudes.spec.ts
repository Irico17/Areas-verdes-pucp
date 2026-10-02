import { expect, test } from "@playwright/test"

test("la solicitud no es solo un texto de lugar y la orden elige empresa", async ({ page }) => {
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.getByRole("button", { name: "Solicitudes", exact: true }).click()

  const solicitud = page.locator("form.solicitud-alta")
  await expect(solicitud).toBeVisible()
  await expect(solicitud.locator('select[name="lugar_id"]')).toBeVisible()
  await expect(solicitud.locator('input[name="lugar"]')).toHaveCount(0)
  await expect(solicitud.getByLabel("Latitud")).toBeVisible()
  await expect(solicitud.getByLabel("Longitud")).toBeVisible()
  await expect(solicitud.getByLabel("Cantidad solicitada")).toBeVisible()
  await expect(solicitud.getByLabel("Cantidad ejecutada")).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/solicitud-ubicacion.png", fullPage: true })

  const orden = page.locator("form.orden-alta")
  const empresa = orden.getByLabel("Empresa")
  await expect(empresa).toBeVisible()
  await expect(orden.locator('input[name="empresa"]')).toHaveCount(0)
  await expect(empresa.locator("option", { hasText: "Taller Verde Andino (ficticia)" })).toHaveCount(1)
  await empresa.selectOption({ label: "Taller Verde Andino (ficticia)" })
  await expect(orden.getByLabel("Frecuencia")).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/orden-empresa.png", fullPage: true })
})
