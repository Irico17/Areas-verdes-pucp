export const TOPE_CLIENTE = Math.floor(1.5 * 1024 * 1024)
export const TOPE_API = 8 * 1024 * 1024
export const LADO_LARGO = 1600
export const CALIDAD_INICIAL = 0.7

export function medida(ancho: number, alto: number, lado = LADO_LARGO): { ancho: number; alto: number } {
  const largo = Math.max(ancho, alto)
  if (largo <= lado || largo === 0) return { ancho, alto }
  const ratio = lado / largo
  return {
    ancho: Math.max(1, Math.round(ancho * ratio)),
    alto: Math.max(1, Math.round(alto * ratio)),
  }
}

export function rechazarPorTamano(bytes: number): boolean {
  return bytes > TOPE_API
}

export async function comprimirFoto(file: File): Promise<{ blob: Blob; mime: string; nombre: string }> {
  const pdf = file.type === "application/pdf" || file.name.toLowerCase().endsWith(".pdf")
  if (pdf) {
    if (file.size > TOPE_API) throw new Error("el archivo supera 8 MB")
    return { blob: file, mime: "application/pdf", nombre: file.name || "documento.pdf" }
  }
  if (!file.type.startsWith("image/") && !/\.(jpe?g|png|webp)$/i.test(file.name)) {
    throw new Error("se admite jpg, png, webp o pdf")
  }
  const bitmap = await createImageBitmap(file)
  try {
    let lado = LADO_LARGO
    let calidad = CALIDAD_INICIAL
    let ultimo: Blob | null = null
    for (let intento = 0; intento < 8; intento++) {
      const box = medida(bitmap.width, bitmap.height, lado)
      const canvas = document.createElement("canvas")
      canvas.width = box.ancho
      canvas.height = box.alto
      const ctx = canvas.getContext("2d")
      if (!ctx) throw new Error("no se pudo preparar la foto")
      ctx.drawImage(bitmap, 0, 0, box.ancho, box.alto)
      const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, "image/jpeg", calidad))
      if (!blob) throw new Error("no se pudo comprimir la foto")
      ultimo = blob
      if (blob.size <= TOPE_CLIENTE) {
        const base = file.name.replace(/\.[^.]+$/, "") || "evidencia"
        return { blob, mime: "image/jpeg", nombre: `${base}.jpg` }
      }
      if (calidad > 0.45) calidad = Math.round((calidad - 0.1) * 10) / 10
      else lado = Math.round(lado * 0.8)
    }
    if (ultimo && ultimo.size <= TOPE_API && ultimo.size <= TOPE_CLIENTE) {
      return { blob: ultimo, mime: "image/jpeg", nombre: "evidencia.jpg" }
    }
    throw new Error(ultimo && ultimo.size > TOPE_API ? "el archivo supera 8 MB" : "la foto sigue por encima de 1.5 MB")
  } finally {
    bitmap.close()
  }
}
