import { expect, test } from "./fixtures"

test("al volver la red el cambio de estado sale de la cola sin pulsar el botón", async ({ page, context }) => {
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: /Riego del eje central/ }).click()

  const detalle = page.locator(".detail")
  await expect(detalle.getByRole("heading", { name: "Riego del eje central" })).toBeVisible()
  const estado = detalle.getByLabel("Estado")
  const actual = await estado.inputValue()
  const siguiente = actual === "en_proceso" ? "pendiente" : "en_proceso"
  await estado.selectOption(siguiente)

  await context.setOffline(true)
  await detalle.getByRole("button", { name: "Cambiar estado" }).click()
  const enCola = page.locator(".marca-cola", { hasText: "por enviar" }).first()
  await expect(enCola).toBeVisible()
  await enCola.scrollIntoViewIfNeeded()
  await page.screenshot({ path: "e2e-artifacts/capturas/offline-estado-en-cola.png" })

  const envio = page.waitForRequest((req) => req.method() === "PATCH" && req.url().includes("/estado"))
  await context.setOffline(false)
  const pedido = await envio
  expect(pedido.postData() ?? "").toContain(siguiente)
  await expect(page.getByRole("button", { name: "Reintentar envío" })).toHaveCount(0)
  await expect(page.getByText("por enviar")).toHaveCount(0)
  await page.screenshot({ path: "e2e-artifacts/capturas/offline-estado-enviado.png", fullPage: true })
})
