import { expect, test } from "./fixtures"
import { deflateSync } from "node:zlib"

function pngSolido(ancho: number, alto: number): Buffer {
  const firma = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10])
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(ancho, 0)
  ihdr.writeUInt32BE(alto, 4)
  ihdr[8] = 8
  ihdr[9] = 2
  const crudo: number[] = []
  for (let y = 0; y < alto; y += 1) {
    crudo.push(0)
    for (let x = 0; x < ancho; x += 1) {
      crudo.push(0x1f, 0x6b, 0x43)
    }
  }
  const trozo = (tipo: string, datos: Buffer) => {
    const largo = Buffer.alloc(4)
    largo.writeUInt32BE(datos.length, 0)
    const nombre = Buffer.from(tipo)
    const crc = Buffer.alloc(4)
    crc.writeUInt32BE(crc32(Buffer.concat([nombre, datos])) >>> 0, 0)
    return Buffer.concat([largo, nombre, datos, crc])
  }
  return Buffer.concat([firma, trozo("IHDR", ihdr), trozo("IDAT", deflateSync(Buffer.from(crudo))), trozo("IEND", Buffer.alloc(0))])
}

function crc32(buf: Buffer): number {
  let c = ~0
  for (const byte of buf) {
    c ^= byte
    for (let i = 0; i < 8; i += 1) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
  }
  return ~c
}

test("la bitácora muestra la reasignación y la foto bajo ese hito", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto("/")
  await page.getByLabel("Usuario").fill("coordinacion")
  await page.getByLabel("Clave").fill("pando-local")
  await page.getByRole("button", { name: "Entrar" }).click()

  await page.getByRole("button", { name: "Actividades", exact: true }).click()
  await page.getByRole("button", { name: /Riego del eje central/ }).click()
  await page.getByRole("button", { name: "Más acciones" }).click()
  await page.getByLabel("Reasignar a").selectOption({ label: "Cuadrilla Sur" })
  await page.getByRole("button", { name: "Reasignar", exact: true }).click()
  await page.locator("#detalle-actividad summary", { hasText: "Bitácora" }).click()

  const hito = page.locator(".bitacora-hito", { hasText: "Reasignada" })
  await expect(hito).toBeVisible()
  await expect(hito.getByText("Cuadrilla Norte")).toBeVisible()
  await expect(hito.getByText("Cuadrilla Sur")).toBeVisible()
  await hito.scrollIntoViewIfNeeded()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/bitacora-reasignada.png" })

  await hito.locator("input[type=file]").setInputFiles({
    name: "aspersor.png",
    mimeType: "image/png",
    buffer: pngSolido(112, 84),
  })
  const foto = hito.locator("img[alt='aspersor.png']").first()
  await expect(foto).toBeVisible()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/bitacora-evidencia.png" })

  await page.locator(".guard").getByRole("button", { name: "Bitácora", exact: true }).click()
  await expect(page.getByText("Quién hizo qué, a qué hora")).toBeVisible()
  const enPanel = page.locator(".bitacora-panel .bitacora-hito", { hasText: "Reasignada" })
  const fotoPanel = enPanel.locator("img[alt='aspersor.png']").first()
  await expect(fotoPanel).toBeVisible()
  await enPanel.scrollIntoViewIfNeeded()
  await fotoPanel.evaluate((img: HTMLImageElement) => img.complete && img.naturalWidth > 0)
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/bitacora-panel.png" })

  await page.setViewportSize({ width: 390, height: 844 })
  await enPanel.scrollIntoViewIfNeeded()
  await page.screenshot({ path: "/opt/cursor/artifacts/screenshots/bitacora-movil.png" })
})
