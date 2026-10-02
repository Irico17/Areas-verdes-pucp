import { expect, test } from "@playwright/test"

test("la jefatura da de alta una cuenta y la persona cambia su clave", async ({ page }) => {
  await page.goto("/")
  await expect(page.getByRole("heading", { name: "Entrar al turno" })).toBeVisible()
  await expect(page.getByText(/universidad/i)).toHaveCount(0)
  await expect(page.getByText(/SSO/)).toHaveCount(0)
  await expect(page.getByText("Inicia sesión con tu cuenta de VerdePUCP.")).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/login-sin-sso.png", fullPage: true })

  await page.getByLabel("Usuario").fill("jefatura")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  await page.getByRole("button", { name: "Admin" }).click()
  await expect(page.getByRole("heading", { name: "Nueva cuenta" })).toBeVisible()
  await expect(page.getByText("La jefatura de sección administra las cuentas.")).toBeVisible()

  const rol = page.getByLabel("Rol").first()
  await expect(rol.locator("option")).toContainText([
    "Capataz",
    "Ingeniería / Coordinación",
    "Jefatura de sección",
    "Administrador del sistema",
  ])
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/admin-alta-cuentas.png", fullPage: true })

  const alta = page.locator("form").filter({ has: page.getByRole("button", { name: "Dar de alta" }) })
  await alta.getByLabel("Usuario", { exact: true }).fill("lucia.paredes")
  await alta.getByLabel("Nombre", { exact: true }).fill("Lucía Paredes")
  await rol.selectOption("capataz")
  await alta.getByLabel("Cuadrilla").fill("cap-norte")
  await alta.getByLabel("Clave inicial").fill("clave-lucia-01")
  await page.getByRole("button", { name: "Dar de alta" }).click()
  await expect(page.getByText("Cuenta creada. La persona debe cambiar la clave al entrar.")).toBeVisible()
  await expect(page.getByText("lucia.paredes")).toBeVisible()

  await page.getByRole("button", { name: "Salir" }).click()
  await expect(page.getByRole("heading", { name: "Entrar al turno" })).toBeVisible()
  await page.getByLabel("Usuario").fill("lucia.paredes")
  await page.getByLabel("Clave").fill("clave-lucia-01")
  await page.getByRole("button", { name: "Entrar" }).click()
  await expect(page.getByRole("heading", { name: "Elija su clave" })).toBeVisible()
  await expect(page.getByText(/universidad/i)).toHaveCount(0)
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/cambio-clave.png", fullPage: true })
})
