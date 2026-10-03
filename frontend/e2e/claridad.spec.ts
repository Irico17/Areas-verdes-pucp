import { randomUUID } from "node:crypto"

import { expect, test, type APIRequestContext } from "./fixtures"

async function entrar(page: import("@playwright/test").Page, usuario: string) {
  await page.goto("/")
  await page.getByLabel("Usuario").fill(usuario)
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()
}

test("capataz a 390 ve Hoy y no ve Catastro ni Inventario", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  await expect(page.getByRole("heading", { name: /Cuadrilla Norte/ })).toBeVisible()
  await expect(page.getByRole("button", { name: "Hoy", exact: true })).toBeVisible()
  await expect(page.getByRole("button", { name: "Catastro", exact: true })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Áreas y sectores", exact: true })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Inventario", exact: true })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Más", exact: true })).toHaveCount(0)
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/capataz-hoy-390.png" })
})

test("capataz inicia una actividad y sin red ve el chip por enviar", async ({ page, context }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  const iniciar = page.getByRole("button", { name: "Iniciar" }).first()
  await expect(iniciar).toBeVisible()
  await iniciar.click()
  await context.setOffline(true)
  await expect(page.getByText(/Sin conexión · \d+ por enviar/)).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/capataz-sin-conexion-390.png" })
  await context.setOffline(false)
})

test("jefatura no ve Guardar de catastro ni formularios de poda, vivero o riego", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "jefatura")
  await page.getByRole("button", { name: "Áreas y sectores", exact: true }).click()
  const fila = page.locator(".catastro-editor .labor-list .labor").first()
  if (await fila.count()) await fila.click()
  await expect(page.getByRole("button", { name: "Guardar área" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Nueva área" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar zona de supervisión" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar referente" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Registros de campo" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Guardar poda" })).toHaveCount(0)
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/jefatura-sin-guardar-1280.png" })
})

test("la pestaña dice Administración y el historial distingue los roles", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "admin")
  const admin = page.getByRole("button", { name: "Administración", exact: true })
  await expect(admin).toBeVisible()
  await expect(page.getByRole("button", { name: "Admin", exact: true })).toHaveCount(0)
  await admin.click()
  await expect(page.getByRole("heading", { name: "Nueva cuenta" })).toBeVisible()
  await page.getByRole("button", { name: "Historial", exact: true }).click()
  await expect(page.getByRole("heading", { name: "Historial" })).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/admin-administracion-1280.png" })

  await page.getByRole("button", { name: "Salir" }).click()
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  await expect(page.getByRole("button", { name: "Historial", exact: true })).toHaveCount(0)
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/capataz-sin-historial-390.png" })
})

test("registros de campo abre el formulario solo con Nuevo registro", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("button", { name: "Registros de campo", exact: true }).click()
  await expect(page.getByRole("tab", { name: "Riego" })).toHaveAttribute("aria-selected", "true")
  await expect(page.getByRole("tab", { name: "Poda" })).toBeVisible()
  await expect(page.getByRole("tab", { name: "Vivero" })).toBeVisible()
  await expect(page.locator("#riego select[name='sector_id']")).toHaveCount(0)
  await page.getByRole("button", { name: "Nuevo registro" }).click()
  await expect(page.locator("#riego select[name='sector_id']")).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/registros-campo-1280.png" })
})

test("el mapa muestra capas y leyenda sin abrir el panel", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("navigation", { name: "Módulos" }).getByRole("button", { name: "Mapa", exact: true }).click()
  const control = page.getByRole("complementary", { name: "Capas" })
  await expect(control.getByRole("button", { name: "Plano" })).toBeVisible()
  await expect(control.getByRole("button", { name: "Relieve" })).toBeVisible()
  await expect(control.getByRole("button", { name: "Sector de capataz", exact: true })).toBeVisible()
  await expect(page.locator(".shell")).toHaveClass(/panel-off/)
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/mapa-capas-1280.png" })
  await page.setViewportSize({ width: 390, height: 844 })
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/mapa-capas-390.png" })
})

test("el nombre del capataz se lee y el botón de ayuda toma el foco", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  await expect(page.getByText("Elsa Quispe")).toBeVisible()
  await expect(page.locator(".chip", { hasText: "Capataz" })).toBeVisible()
  const ayuda = page.getByRole("button", { name: "Qué significa Bloqueada" })
  await ayuda.focus()
  await expect(ayuda).toBeFocused()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/capataz-identidad-390.png" })
})

async function prepararActividad(request: APIRequestContext, titulo: string) {
  const sesion = await request.post("/areas-verdes/v1/sesion", {
    data: { usuario: "coordinacion", clave: "pando-local" },
  })
  const cuerpoSesion = await sesion.text()
  expect(sesion.ok(), `${sesion.status()} ${cuerpoSesion}`).toBeTruthy()
  const alta = await request.post("/areas-verdes/v1/operacion/actividades", {
    data: {
      id: randomUUID(),
      tipo: "limpieza",
      titulo,
      detalle: "Preparada para repetir el diálogo de archivo.",
      lon: -77.0789,
      lat: -12.0682,
      assigned_capataz_id: "cap-norte",
      ejecutor: "propia",
    },
  })
  const cuerpoAlta = await alta.text()
  expect(alta.ok(), `${alta.status()} ${cuerpoAlta}`).toBeTruthy()
}

test("archivar pide el motivo en un diálogo y la actividad sale de la lista", async ({ page, request }) => {
  const titulo = `Limpieza e2e ${randomUUID().slice(0, 8)}`
  await prepararActividad(request, titulo)
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  const fila = page.getByRole("button", { name: new RegExp(titulo) })
  await fila.click()
  await page.getByRole("button", { name: "Más acciones" }).click()
  await page.getByRole("button", { name: "Archivar", exact: true }).click()
  const dialogo = page.locator("dialog.dialogo-archivo")
  await expect(dialogo).toBeVisible()
  await expect(dialogo).toContainText(`¿Archivar "${titulo}"?`)
  await expect(dialogo).toContainText("Deja de verse en el mapa; no se borra.")
  await dialogo.getByRole("button", { name: "Cancelar" }).click()
  await expect(dialogo).toBeHidden()
  await expect(fila).toBeVisible()

  await page.getByRole("button", { name: "Más acciones" }).click()
  await page.getByRole("button", { name: "Archivar", exact: true }).click()
  await expect(dialogo.getByRole("button", { name: "Archivar" })).toBeDisabled()
  await dialogo.getByLabel("Motivo de archivo").selectOption({ index: 1 })
  await dialogo.getByRole("button", { name: "Archivar" }).click()
  await expect(fila).toHaveCount(0)
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/detalle-archivo-1280.png" })
  await page.setViewportSize({ width: 390, height: 844 })
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/detalle-archivo-390.png" })
})

test("el alta es un banner y luego una hoja, y permite otra seguida", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: "Nueva actividad" }).click()
  await expect(page.getByText("Toque el mapa donde está el trabajo.")).toBeVisible()
  await expect(page.locator("form.alta-actividad")).toHaveCount(0)

  const mapa = page.locator(".maplibregl-canvas")
  await mapa.waitFor({ state: "visible" })
  await page.waitForTimeout(800)
  await mapa.click({ position: { x: 72, y: 280 }, force: true })
  const alta = page.locator("form.alta-actividad")
  await expect(alta).toBeVisible()
  await expect(alta.getByRole("group", { name: "Qué" })).toBeVisible()
  await expect(alta.getByRole("group", { name: "Quién" })).toBeVisible()
  await expect(alta.getByRole("group", { name: "Detalle" })).toBeVisible()
  await alta.getByLabel("Clase de actividad").selectOption("riego")
  await alta.getByLabel("Tipo de actividad").selectOption({ label: "Riego manual" })
  await alta.getByLabel("Título").fill("Riego de prueba claridad")
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/alta-pasos-1280.png" })
  await alta.getByRole("button", { name: "Crear actividad" }).click()
  await expect(page.getByText("Actividad creada. Marque el siguiente punto.")).toBeVisible()
  await expect(page.locator("form.alta-actividad")).toHaveCount(0)
  await expect(page.getByText("Toque el mapa donde está el trabajo.")).toBeVisible()

  await mapa.click({ position: { x: 120, y: 300 }, force: true })
  await expect(page.locator("form.alta-actividad")).toBeVisible()
  await page.setViewportSize({ width: 390, height: 844 })
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/alta-pasos-390.png" })
})

test("los contadores de estado filtran y Quitar filtros los suelta", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  const porIniciar = page.getByRole("button", { name: /Por iniciar \d+/ })
  await expect(porIniciar).toBeVisible()
  await expect(page.getByRole("button", { name: /En proceso \d+/ })).toBeVisible()
  await porIniciar.click()
  await expect(porIniciar).toHaveAttribute("aria-pressed", "true")
  await expect(page.getByRole("button", { name: "Quitar filtros" })).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/filtros-contadores-1280.png" })
  await page.getByRole("button", { name: "Quitar filtros" }).click()
  await expect(porIniciar).toHaveAttribute("aria-pressed", "false")
  await expect(page.getByRole("button", { name: "Quitar filtros" })).toHaveCount(0)
  await page.setViewportSize({ width: 390, height: 844 })
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/filtros-contadores-390.png" })
})

test("ficha, evidencia y bitácora nacen plegadas y el rol oculta lo que no puede", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "coordinacion")
  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: /Riego del eje central/ }).click()
  const detalle = page.locator("#detalle-actividad")
  await expect(detalle.getByRole("button", { name: "Guardar ficha" })).toBeHidden()
  await detalle.locator("summary", { hasText: "Ficha de la actividad" }).click()
  await expect(detalle.getByRole("button", { name: "Guardar ficha" })).toBeVisible()
  await expect(detalle.getByRole("button", { name: "Cambiar estado" })).toBeVisible()
  await expect(detalle.getByText("Foto", { exact: true })).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/detalle-pliegues-1280.png" })

  await page.getByRole("button", { name: "Salir" }).click()
  await page.setViewportSize({ width: 390, height: 844 })
  await entrar(page, "norte")
  await page.getByRole("button", { name: "Mis actividades", exact: true }).click()
  const omitir = page.getByRole("button", { name: "Omitir" })
  if (await omitir.isVisible()) await omitir.click()
  await page.locator(".labor-list .labor").first().click()
  await expect(page.getByRole("button", { name: "Más acciones" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Cambiar estado" })).toBeVisible()
  await page.screenshot({ path: "e2e-artifacts/o4-claridad/detalle-pliegues-390.png" })
})

test("jefatura cambia el estado y no ve guardar ficha ni avance", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
  await entrar(page, "jefatura")
  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: /Riego del eje central/ }).click()
  const fichaJefe = page.locator("#detalle-actividad")
  await fichaJefe.locator("summary", { hasText: "Ficha de la actividad" }).click()
  await expect(fichaJefe.getByRole("button", { name: "Guardar ficha" })).toHaveCount(0)
  await expect(fichaJefe.getByRole("button", { name: "Guardar avance" })).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Más acciones" })).toBeVisible()
})
