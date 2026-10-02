import { expect, test } from "@playwright/test"

test("se abre un ejemplar, cambia la salud y queda el código anterior", async ({ page }) => {
  const codigo = `E2E-FLORA-${Date.now()}`
  const nuevo = `${codigo}-B`

  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  await expect(page.getByRole("button", { name: "Ejemplares" })).toBeVisible()

  const alta = await page.request.post("/areas-verdes/v1/catastro/ejemplares", {
    data: {
      codigo,
      nombre_comun: "Tipuana de ensayo",
      tipo_vegetacion: "Árbol",
      cantidad: 1,
      lat: -12.07,
      lon: -77.08,
    },
  })
  expect(alta.status()).toBe(201)

  await page.getByRole("button", { name: "Ejemplares" }).click()
  await page.getByLabel("Buscar por código, nombre común o N°").fill(codigo)
  await page.getByRole("button", { name: new RegExp(codigo) }).click()
  await expect(page.getByRole("heading", { name: codigo })).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/ejemplares-ficha.png", fullPage: true })

  await page.getByLabel("Salud").selectOption("bueno")
  await page.getByRole("button", { name: "Guardar ficha" }).click()
  await expect(page.getByText("Ficha guardada.")).toBeVisible()
  await expect(page.getByRole("button", { name: new RegExp(codigo) }).getByText("Bueno")).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/ejemplares-salud.png", fullPage: true })

  await page.getByLabel("Código nuevo").fill(nuevo)
  await page.getByRole("button", { name: "Recodificar" }).click()
  await expect(page.getByText(`${codigo} pasó a ${nuevo}`)).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/ejemplares-codigo-anterior.png", fullPage: true })
})
