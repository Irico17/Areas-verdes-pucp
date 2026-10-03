import assert from "node:assert/strict"
import test from "node:test"
import { CALIDAD_INICIAL, LADO_LARGO, TOPE_CLIENTE, comprimirFoto, medida, rechazarPorTamano } from "./comprimir.ts"

test("el lado largo baja a 1600 px", () => {
  assert.deepEqual(medida(3200, 1600), { ancho: 1600, alto: 800 })
  assert.deepEqual(medida(800, 400), { ancho: 800, alto: 400 })
  assert.equal(LADO_LARGO, 1600)
  assert.equal(CALIDAD_INICIAL, 0.7)
  assert.equal(TOPE_CLIENTE, 1_572_864)
})

test("por encima de 8 MB se rechaza antes de salir", () => {
  assert.equal(rechazarPorTamano(8 * 1024 * 1024 + 1), true)
  assert.equal(rechazarPorTamano(TOPE_CLIENTE), false)
})

test("comprimirFoto baja una imagen grande por debajo de 1.5 MB", async () => {
  const originalDoc = globalThis.document
  const originalBitmap = globalThis.createImageBitmap
  globalThis.createImageBitmap = async () => ({
    width: 4000,
    height: 3000,
    close() {},
  })
  globalThis.document = {
    createElement() {
      const canvas = { width: 0, height: 0 }
      return {
        set width(v: number) {
          canvas.width = v
        },
        get width() {
          return canvas.width
        },
        set height(v: number) {
          canvas.height = v
        },
        get height() {
          return canvas.height
        },
        getContext() {
          return { drawImage() {} }
        },
        toBlob(cb: (blob: Blob | null) => void, _tipo?: string, calidad?: number) {
          const pixeles = canvas.width * canvas.height
          const peso = Math.floor(pixeles * (calidad ?? 0.7) * 2)
          cb(new Blob([new Uint8Array(peso)], { type: "image/jpeg" }))
        },
      }
    },
  } as unknown as Document
  try {
    const file = new File([new Uint8Array(32)], "campo.png", { type: "image/png" })
    const out = await comprimirFoto(file)
    assert.equal(out.mime, "image/jpeg")
    assert.ok(out.blob.size > 0)
    assert.ok(out.blob.size <= TOPE_CLIENTE, `quedó en ${out.blob.size}`)
  } finally {
    globalThis.document = originalDoc
    globalThis.createImageBitmap = originalBitmap
  }
})
