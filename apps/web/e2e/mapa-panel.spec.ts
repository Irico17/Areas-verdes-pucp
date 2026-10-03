import { expect, test, type Locator, type Page } from "./fixtures"

async function entrar(page: Page, usuario: string) {
  await page.goto("/")
  await page.getByLabel("Usuario").fill(usuario)
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
}

function solapan(a: { x: number; y: number; width: number; height: number }, b: { x: number; y: number; width: number; height: number }) {
  return a.x < b.x + b.width && a.x + a.width > b.x && a.y < b.y + b.height && a.y + a.height > b.y
}

async function caja(locator: Locator) {
  const box = await locator.boundingBox()
  if (!box) throw new Error("sin caja")
  return box
}

async function captura(page: Page, ancho: number, alto: number, estado: "colapsado" | "abierto") {
  await page.setViewportSize({ width: ancho, height: alto })
  await page.screenshot({ path: `e2e-artifacts/mapa-panel/${ancho}-${estado}.png` })
}

async function cerrarRecorrido(page: Page) {
  const omitir = page.getByRole("button", { name: "Omitir" })
  try {
    await omitir.waitFor({ state: "visible", timeout: 3000 })
    await omitir.click()
  } catch {
    /* el recorrido ya se había cerrado */
  }
}

test("en el teléfono el panel de capas arranca cerrado y no tapa la barra ni el zoom", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "coordinacion")
  await page.getByRole("navigation", { name: "Módulos" }).getByRole("button", { name: "Mapa", exact: true }).click()
  await cerrarRecorrido(page)
  const abrir = page.getByRole("button", { name: "Mostrar capas" })
  await expect(abrir).toBeVisible()
  await expect(abrir).toHaveAttribute("aria-expanded", "false")
  await expect(page.getByRole("button", { name: "Plano" })).toHaveCount(0)
  const lista = page.getByRole("group", { name: "Vista de actividades" }).getByRole("button", { name: "Lista", exact: true })
  await expect(lista).toBeVisible()

  const zoom = page.locator(".maplibregl-ctrl-bottom-right")
  await expect(zoom).toBeVisible()
  const nav = page.locator(".guard")
  expect(solapan(await caja(abrir), await caja(zoom))).toBe(false)
  expect(solapan(await caja(abrir), await caja(nav))).toBe(false)

  await page.getByRole("button", { name: "Cuenta" }).click()
  const salir = page.getByRole("button", { name: "Salir" })
  await expect(salir).toBeVisible()
  expect(solapan(await caja(abrir), await caja(salir))).toBe(false)
  await page.getByRole("button", { name: "Cuenta" }).click()
  await expect(salir).toBeHidden()

  await captura(page, 360, 800, "colapsado")
  await captura(page, 390, 844, "colapsado")
  await captura(page, 430, 844, "colapsado")

  await page.setViewportSize({ width: 390, height: 844 })
  await abrir.click()
  const cerrar = page.getByRole("button", { name: "Cerrar capas" })
  await expect(cerrar).toBeVisible()
  await expect(cerrar).toHaveAttribute("aria-expanded", "true")
  const hoja = page.locator(".marco-mapa[data-abierto='si']")
  await expect(hoja.getByRole("button", { name: "Plano" })).toBeVisible()
  await expect(hoja.getByRole("button", { name: "Relieve" })).toBeVisible()
  await expect(hoja.getByRole("button", { name: "Mapa", exact: true })).toBeVisible()
  const hojaCaja = await caja(hoja)
  expect(hojaCaja.height).toBeLessThanOrEqual(844 * 0.56)
  expect(solapan(hojaCaja, await caja(nav))).toBe(false)
  expect(solapan(hojaCaja, await caja(zoom))).toBe(false)
  await captura(page, 390, 844, "abierto")
  await captura(page, 360, 800, "abierto")
  await captura(page, 430, 844, "abierto")

  await page.setViewportSize({ width: 390, height: 844 })
  await page.locator(".maplibregl-canvas").click({ position: { x: 30, y: 180 }, force: true })
  await expect(abrir).toBeVisible()
  await abrir.click()
  await cerrar.click()
  await expect(abrir).toBeVisible()
  await abrir.click()
  await page.keyboard.press("Escape")
  await expect(abrir).toBeFocused()
})

test("el chip de sin conexión no tapa el botón de capas", async ({ page, context }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  await cerrarRecorrido(page)
  await context.setOffline(true)
  const chip = page.locator(".chip.chip-alerta")
  await expect(chip).toBeVisible()
  const abrir = page.getByRole("button", { name: "Mostrar capas" })
  await expect(abrir).toBeVisible()
  expect(solapan(await caja(abrir), await caja(chip))).toBe(false)
  await context.setOffline(false)
})

test("en escritorio el panel se oculta y se recuerda al recargar", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("navigation", { name: "Módulos" }).getByRole("button", { name: "Mapa", exact: true }).click()
  await cerrarRecorrido(page)
  await expect(page.getByRole("button", { name: "Plano" })).toBeVisible()
  await captura(page, 1280, 800, "abierto")
  await page.getByRole("button", { name: "Ocultar capas" }).click()
  const abrir = page.getByRole("button", { name: "Mostrar capas" })
  await expect(abrir).toBeVisible()
  await expect(page.getByRole("button", { name: "Plano" })).toHaveCount(0)
  const zoom = page.locator(".maplibregl-ctrl-bottom-right")
  await expect(zoom).toBeVisible()
  expect(solapan(await caja(abrir), await caja(zoom))).toBe(false)
  await captura(page, 1280, 800, "colapsado")

  await page.reload()
  await expect(page.getByRole("button", { name: "Mostrar capas" })).toBeVisible()
  await expect(page.getByRole("button", { name: "Plano" })).toHaveCount(0)
  const guardado = await page.evaluate(() => localStorage.getItem("cv:panel-mapa:coordinacion:ancho"))
  expect(guardado).toBe("0")

  await page.getByRole("button", { name: "Mostrar capas" }).click()
  await expect(page.getByRole("button", { name: "Plano" })).toBeVisible()
  await page.reload()
  await expect(page.getByRole("button", { name: "Plano" })).toBeVisible()
})
