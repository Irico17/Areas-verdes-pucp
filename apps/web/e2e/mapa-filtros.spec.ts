import { expect, test } from "@playwright/test"

const PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFCSjAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
  "base64",
)

test("el pin abre el detalle sin cambiar de pestaña y a 390 px se ve la lista", async ({ page }) => {
  await page.route("https://tile.openstreetmap.org/**", (route) =>
    route.fulfill({ status: 200, contentType: "image/png", body: PNG }),
  )
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.locator(".guard").getByRole("button", { name: "Mapa", exact: true }).click()
  await expect(page.locator(".guard").getByRole("button", { name: "Mapa", exact: true })).toHaveAttribute("aria-current", "page")
  await page.waitForFunction(() => {
    const host = document.querySelector(".map-host") as { mapa?: { getLayer: (id: string) => unknown; querySourceFeatures: (id: string) => { geometry?: { type?: string; coordinates?: number[] } }[] } } | null
    const mapa = host?.mapa
    if (!mapa?.getLayer("actividades-circle")) return false
    return mapa.querySourceFeatures("actividades").some((feature) => feature.geometry?.type === "Point")
  })

  const punto = await page.evaluate(() => {
    const host = document.querySelector(".map-host") as HTMLElement & {
      mapa?: { project: (ll: [number, number]) => { x: number; y: number }; querySourceFeatures: (id: string) => { geometry?: { type?: string; coordinates?: number[] } }[] }
    }
    const mapa = host.mapa
    const feature = mapa?.querySourceFeatures("actividades").find((item) => item.geometry?.type === "Point")
    const coords = feature?.geometry?.coordinates
    if (!mapa || !coords) return null
    const proyectado = mapa.project([coords[0], coords[1]])
    return { x: proyectado.x, y: proyectado.y }
  })
  if (!punto) throw new Error("no hay un punto de actividad en el mapa")
  await page.locator(".maplibregl-canvas").click({ position: punto, force: true })

  await expect(page.locator(".cv-popup-title")).toBeVisible()
  await expect(page.locator("#detalle-actividad")).toBeVisible()
  await expect(page.locator(".guard").getByRole("button", { name: "Mapa", exact: true })).toHaveAttribute("aria-current", "page")
  await expect(page.locator(".guard").getByRole("button", { name: "Actividades", exact: true })).not.toHaveAttribute("aria-current", "page")
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/mapa-pin-sin-cambiar-pestana.png" })

  await page.setViewportSize({ width: 390, height: 844 })
  const lista = page.getByRole("group", { name: "Vista de actividades" }).getByRole("button", { name: "Lista", exact: true })
  await expect(lista).toBeVisible()
  const caja = await lista.boundingBox()
  if (!caja || caja.y < 0 || caja.y > 844 || caja.x < 0 || caja.x > 390) {
    throw new Error(`el conmutador quedó fuera de la vista: ${JSON.stringify(caja)}`)
  }
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/mapa-conmutador-390.png" })
  await lista.click()
  await expect(page.locator(".lista-sobre-mapa")).toBeVisible()
  await expect(page.locator(".guard").getByRole("button", { name: "Mapa", exact: true })).toHaveAttribute("aria-current", "page")
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/mapa-lista-390.png" })
})

test("sin teselas la lista queda forzada y dice por qué", async ({ page }) => {
  await page.route("https://tile.openstreetmap.org/**", (route) => route.abort("failed"))
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  await expect(page.getByText("Sin teselas. La lista queda a la vista porque el plano no cargó.")).toBeVisible()
  await expect(page.locator(".lista-sobre-mapa")).toBeVisible()
  await expect(page.getByRole("group", { name: "Vista de actividades" }).getByRole("button", { name: "Mapa", exact: true })).toBeDisabled()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/mapa-sin-teselas.png" })
})
