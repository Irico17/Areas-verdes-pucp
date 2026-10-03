import { test as base, expect } from "@playwright/test"

/** PNG de 1×1. Las teselas y los glifos no deben tumbar la prueba si esos hosts no responden. */
const PNG_MINIMO = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "base64",
)

const MAPA_EXTERNO = /tile\.openstreetmap\.org|demotiles\.maplibre\.org/i

export const test = base.extend({
  page: async ({ page }, aplicar) => {
    await page.route(MAPA_EXTERNO, async (route) => {
      const url = route.request().url()
      if (/demotiles\.maplibre\.org/i.test(url) && !/\.png(?:$|\?)/i.test(url)) {
        await route.fulfill({ status: 200, contentType: "application/octet-stream", body: Buffer.alloc(0) })
        return
      }
      await route.fulfill({ status: 200, contentType: "image/png", body: PNG_MINIMO })
    })
    await aplicar(page)
  },
})

export { expect }
