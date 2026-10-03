import { randomUUID } from "node:crypto"

import { expect, test, type APIRequestContext, type Page } from "./fixtures"

const API = "/areas-verdes/v1"
const CLAVE = "pando-local"
const AJENAS = ["Poda de setos sur", "Inspección bosque húmedo", "Fuga en línea de riego"]
const RUTAS_PROHIBIDAS = ["/catastro", "/inventario", "/admin", "/historial", "/?modulo=catastro", "/#catastro", "/#historial"]

async function entrarCoordinacion(request: APIRequestContext): Promise<void> {
  const res = await request.post(`${API}/sesion`, { data: { usuario: "coordinacion", clave: CLAVE } })
  const cuerpo = await res.text()
  expect(res.ok(), `${res.status()} ${cuerpo}`).toBeTruthy()
}

async function crearEnNorte(request: APIRequestContext, titulo: string): Promise<string> {
  const id = randomUUID()
  const res = await request.post(`${API}/operacion/actividades`, {
    data: {
      id,
      tipo: "limpieza",
      titulo,
      detalle: "Preparada por el recorrido e2e del capataz. Nombre ficticio.",
      lon: -77.07955,
      lat: -12.06955,
      assigned_capataz_id: "cap-norte",
      ejecutor: "propia",
    },
  })
  const cuerpo = await res.text()
  expect(res.ok(), `${res.status()} ${cuerpo}`).toBeTruthy()
  return id
}

async function archivar(request: APIRequestContext, id: string): Promise<void> {
  const res = await request.post(`${API}/operacion/actividades/${id}/archivar`, {
    data: { motivo: "error" },
  })
  if (!res.ok()) {
    throw new Error(`no se pudo archivar ${id}: ${res.status()} ${await res.text()}`)
  }
}

async function entrarNorte(page: Page): Promise<void> {
  await page.goto("/")
  await page.getByLabel("Usuario").fill("norte")
  await page.getByLabel("Clave").fill(CLAVE)
  await page.getByRole("button", { name: "Entrar" }).click()
  await expect(page.getByRole("heading", { name: /Cuadrilla Norte/ })).toBeVisible()
}

async function sinModulosDeOficina(page: Page): Promise<void> {
  for (const nombre of ["Catastro", "Áreas y sectores", "Inventario", "Administración", "Historial"]) {
    await expect(page.getByRole("button", { name: nombre, exact: true })).toHaveCount(0)
  }
  await expect(page.getByRole("button", { name: "Más", exact: true })).toHaveCount(0)
  await expect(page.getByRole("heading", { name: "Áreas y sectores", exact: true })).toHaveCount(0)
  await expect(page.getByRole("heading", { name: "Administración", exact: true })).toHaveCount(0)
  await expect(page.getByRole("heading", { name: "Historial", exact: true })).toHaveCount(0)
  await expect(page.getByRole("heading", { name: "Nueva cuenta" })).toHaveCount(0)
  await expect(page.getByRole("group", { name: "Inventario" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar área" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar zona de supervisión" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar", exact: true })).toHaveCount(0)
}

test("el capataz a 390 recorre Hoy, cambia un estado y no entra a oficina", async ({ page, context, request }) => {
  test.setTimeout(120_000)
  await page.setViewportSize({ width: 390, height: 844 })
  const marca = randomUUID().slice(0, 8)
  const tituloEnLinea = `Control e2e norte ${marca} a`
  const tituloCola = `Control e2e norte ${marca} b`
  const ids: string[] = []
  const fallosArchivo: string[] = []
  let recorridoOk = false

  await entrarCoordinacion(request)
  try {
    ids.push(await crearEnNorte(request, tituloEnLinea))
    ids.push(await crearEnNorte(request, tituloCola))

    await entrarNorte(page)
    const recorrido = page.getByRole("dialog", { name: "Hoy" })
    if (await recorrido.isVisible()) {
      await recorrido.getByRole("button", { name: "Omitir" }).click()
      await expect(recorrido).toBeHidden()
    }
    await expect(page.getByRole("button", { name: "Hoy", exact: true })).toBeVisible()
    await expect(page.getByRole("heading", { name: "Mis actividades de hoy" })).toBeVisible()

    const filaEnLinea = page.getByRole("listitem").filter({ hasText: tituloEnLinea })
    const filaCola = page.getByRole("listitem").filter({ hasText: tituloCola })
    await expect(filaEnLinea.getByRole("button", { name: "Iniciar", exact: true })).toBeVisible()
    await expect(filaCola.getByRole("button", { name: "Iniciar", exact: true })).toBeVisible()
    for (const ajena of AJENAS) {
      await expect(page.getByRole("button", { name: ajena })).toHaveCount(0)
    }
    await sinModulosDeOficina(page)
    await page.screenshot({ path: "e2e-artifacts/o4-e2e/capataz-hoy-390.png" })

    for (const ruta of RUTAS_PROHIBIDAS) {
      await page.goto(ruta)
      await expect(page.getByRole("heading", { name: /Cuadrilla Norte/ })).toBeVisible()
      await expect(page.getByRole("button", { name: "Hoy", exact: true })).toBeVisible()
      await sinModulosDeOficina(page)
    }

    await filaEnLinea.getByRole("button", { name: "Iniciar", exact: true }).click()
    await expect(filaEnLinea.getByRole("button", { name: tituloEnLinea })).toContainText("En proceso")
    await expect(filaEnLinea.getByRole("button", { name: "Terminar y Foto" })).toBeVisible()
    await page.screenshot({ path: "e2e-artifacts/o4-e2e/capataz-en-proceso-390.png" })

    await context.setOffline(true)
    await filaCola.getByRole("button", { name: "Iniciar", exact: true }).click()
    const chip = page.getByRole("status").filter({ hasText: /Sin conexión · [1-9]\d* por enviar/ })
    await chip.scrollIntoViewIfNeeded()
    await expect(chip).toBeVisible()
    await page.screenshot({ path: "e2e-artifacts/o4-e2e/capataz-sin-conexion-390.png" })

    const envio = page.waitForRequest(
      (req) => req.method() === "PATCH" && req.url().includes(`/${ids[1]}/estado`) && (req.postData() ?? "").includes("en_proceso"),
    )
    await context.setOffline(false)
    await envio
    const enLinea = page.getByRole("status").filter({ hasText: /^En línea$/ })
    await enLinea.scrollIntoViewIfNeeded()
    await expect(enLinea).toBeVisible()
    await expect(filaCola.getByRole("button", { name: tituloCola })).toContainText("En proceso")
    await expect(page.getByText(/por enviar/)).toHaveCount(0)
    await page.screenshot({ path: "e2e-artifacts/o4-e2e/capataz-cola-enviada-390.png" })
    recorridoOk = true
  } finally {
    for (const id of ids) {
      try {
        await archivar(request, id)
      } catch (error) {
        fallosArchivo.push(error instanceof Error ? error.message : String(error))
      }
    }
  }
  if (recorridoOk && fallosArchivo.length > 0) {
    throw new Error(fallosArchivo.join("\n"))
  }
})
