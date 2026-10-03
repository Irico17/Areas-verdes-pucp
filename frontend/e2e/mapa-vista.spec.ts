import { expect, test } from "./fixtures"

const PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "base64",
)

async function entrar(page: import("@playwright/test").Page) {
  await page.setViewportSize({ width: 1280, height: 800 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  await page.getByRole("navigation", { name: "Módulos" }).getByRole("button", { name: "Mapa", exact: true }).click()
  await page.waitForFunction(() => {
    const host = document.querySelector(".map-host") as { mapa?: { loaded: () => boolean; isSourceLoaded: (id: string) => boolean } } | null
    const mapa = host?.mapa
    return Boolean(mapa?.loaded() && mapa.isSourceLoaded("osm"))
  })
}

const POLIGONO_VACIO = {
  type: "FeatureCollection",
  features: [
    {
      type: "Feature",
      id: "sin-datos",
      properties: {},
      geometry: {
        type: "Polygon",
        coordinates: [
          [
            [-77.0832, -12.07415],
            [-77.07795, -12.07415],
            [-77.07795, -12.0644],
            [-77.0832, -12.0644],
            [-77.0832, -12.07415],
          ],
        ],
      },
    },
  ],
}

test("un clic en un elemento sin datos no pasa a Lista", async ({ page }) => {
  await entrar(page)
  const mapaBtn = page.getByRole("group", { name: "Vista de actividades" }).getByRole("button", { name: "Mapa", exact: true })
  await expect(mapaBtn).toHaveAttribute("aria-pressed", "true")

  await page.evaluate(() => {
    const host = document.querySelector(".map-host") as {
      mapa?: { fire: (tipo: string, evento: unknown) => void }
    }
    const mapa = host.mapa
    if (!mapa) throw new Error("sin mapa")
    for (let i = 0; i < 6; i += 1) {
      mapa.fire("error", {
        error: new Error(`AJAXError: Failed to fetch (0): https://tile.openstreetmap.org/15/${9640 + i}/15430.png`),
        sourceId: "osm",
      })
    }
    mapa.fire("error", { error: new Error("The feature id does not exist in the source areas") })
  })
  await expect(page.getByText("Sin teselas. La lista queda a la vista porque el plano no cargó.")).toHaveCount(0)
  await expect(mapaBtn).toHaveAttribute("aria-pressed", "true")

  await page.evaluate((coleccion) => {
    const host = document.querySelector(".map-host") as {
      mapa?: {
        getStyle: () => { sources: Record<string, unknown> }
        getSource: (id: string) => { setData?: (data: unknown) => void } | undefined
        setFilter: (capa: string, filtro: null) => void
        triggerRepaint: () => void
      }
    }
    const mapa = host.mapa
    if (!mapa) throw new Error("sin mapa")
    for (const id of Object.keys(mapa.getStyle().sources)) {
      const fuente = mapa.getSource(id)
      if (!fuente?.setData || id === "osm") continue
      const fijar = fuente.setData.bind(fuente)
      fuente.setData = () => {}
      fijar(id === "areas" ? coleccion : { type: "FeatureCollection", features: [] })
    }
    mapa.setFilter("areas-fill", null)
    mapa.triggerRepaint()
  }, POLIGONO_VACIO)

  await page.waitForFunction(() => {
    const host = document.querySelector(".map-host") as {
      mapa?: {
        project: (ll: [number, number]) => { x: number; y: number }
        queryRenderedFeatures: (punto: { x: number; y: number }, opciones: { layers: string[] }) => { source?: string }[]
      }
    }
    const mapa = host.mapa
    if (!mapa) return false
    const punto = mapa.project([-77.0796, -12.0696])
    return mapa.queryRenderedFeatures(punto, { layers: ["areas-fill"] }).some((hit) => hit.source === "areas")
  })

  const punto = await page.evaluate(() => {
    const host = document.querySelector(".map-host") as { mapa?: { project: (ll: [number, number]) => { x: number; y: number } } }
    const proyectado = host.mapa?.project([-77.0796, -12.0696])
    if (!proyectado) return null
    return { x: proyectado.x, y: proyectado.y }
  })
  if (!punto) throw new Error("no se pudo proyectar el clic")
  await page.locator(".maplibregl-canvas").click({ position: punto, force: true })

  await expect(page.getByText("Este elemento no tiene datos para mostrar.")).toBeVisible()
  await expect(page.getByText("Sin teselas. La lista queda a la vista porque el plano no cargó.")).toHaveCount(0)
  await expect(mapaBtn).toHaveAttribute("aria-pressed", "true")
  await expect(page.locator(".lista-sobre-mapa")).toHaveCount(0)
})

test("si ninguna tesela carga se puede volver a Mapa", async ({ page }) => {
  await page.route("https://tile.openstreetmap.org/**", (route) => route.abort("failed"))
  await page.setViewportSize({ width: 1280, height: 800 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
  const aviso = page.getByText("Sin teselas. La lista queda a la vista porque el plano no cargó.")
  await expect(aviso).toBeVisible()
  const mapa = page.getByRole("group", { name: "Vista de actividades" }).getByRole("button", { name: "Mapa", exact: true })
  await expect(mapa).toBeEnabled()
  await page.unroute("https://tile.openstreetmap.org/**")
  await page.route("https://tile.openstreetmap.org/**", (route) => route.fulfill({ status: 200, contentType: "image/png", body: PNG }))
  await mapa.click()
  await expect(aviso).toBeHidden()
  await expect(mapa).toHaveAttribute("aria-pressed", "true")
  await expect(page.locator(".maplibregl-canvas")).toBeVisible()
})
