export type Punto = { lat: number; lon: number }

export type ExifBasico = {
  fecha: string | null
  lat: number | null
  lon: number | null
  orientacion: number | null
}

export function leerExif(bytes: Uint8Array): ExifBasico {
  const vacio: ExifBasico = { fecha: null, lat: null, lon: null, orientacion: null }
  if (bytes.length < 4 || bytes[0] !== 0xff || bytes[1] !== 0xd8) return vacio
  let i = 2
  while (i + 4 < bytes.length) {
    if (bytes[i] !== 0xff) break
    const marker = bytes[i + 1]
    if (marker === 0xda || marker === 0xd9) break
    const size = (bytes[i + 2] << 8) | bytes[i + 3]
    if (size < 2 || i + 2 + size > bytes.length) break
    if (marker === 0xe1) {
      const segment = bytes.subarray(i + 4, i + 2 + size)
      const gps = leerApp1(segment)
      if (gps.fecha || gps.lat != null || gps.orientacion != null) return gps
    }
    i += 2 + size
  }
  return vacio
}

function leerApp1(seg: Uint8Array): ExifBasico {
  const vacio: ExifBasico = { fecha: null, lat: null, lon: null, orientacion: null }
  if (seg.length < 14) return vacio
  const head = String.fromCharCode(...seg.subarray(0, 6))
  if (head !== "Exif\0\0") return vacio
  const tiff = seg.subarray(6)
  const le = tiff[0] === 0x49 && tiff[1] === 0x49
  const be = tiff[0] === 0x4d && tiff[1] === 0x4d
  if (!le && !be) return vacio
  const u16 = (o: number) => (le ? tiff[o] | (tiff[o + 1] << 8) : (tiff[o] << 8) | tiff[o + 1])
  const u32 = (o: number) =>
    le ? tiff[o] | (tiff[o + 1] << 8) | (tiff[o + 2] << 16) | (tiff[o + 3] << 24) : (tiff[o] << 24) | (tiff[o + 1] << 16) | (tiff[o + 2] << 8) | tiff[o + 3]
  if (u16(2) !== 42) return vacio
  const ifd = u32(4)
  const fecha = buscarTexto(tiff, ifd, 0x0132, u16, u32)
  const orientacion = buscarCorto(tiff, ifd, 0x0112, u16, u32)
  const gpsOff = buscarLargo(tiff, ifd, 0x8825, u16, u32)
  if (gpsOff == null) return { fecha, lat: null, lon: null, orientacion }
  const lat = racionalGps(tiff, gpsOff, 0x0002, u16, u32)
  const lon = racionalGps(tiff, gpsOff, 0x0004, u16, u32)
  const latRef = buscarTexto(tiff, gpsOff, 0x0001, u16, u32)
  const lonRef = buscarTexto(tiff, gpsOff, 0x0003, u16, u32)
  return {
    fecha,
    lat: lat == null ? null : latRef === "S" ? -lat : lat,
    lon: lon == null ? null : lonRef === "W" ? -lon : lon,
    orientacion,
  }
}

function buscarCorto(
  tiff: Uint8Array,
  ifd: number,
  tag: number,
  u16: (o: number) => number,
  u32: (o: number) => number,
): number | null {
  if (ifd + 2 > tiff.length) return null
  const n = u16(ifd)
  for (let i = 0; i < n; i++) {
    const o = ifd + 2 + i * 12
    if (o + 12 > tiff.length) return null
    if (u16(o) !== tag || u16(o + 2) !== 3 || u32(o + 4) !== 1) continue
    return u16(o + 8)
  }
  return null
}

function buscarLargo(
  tiff: Uint8Array,
  ifd: number,
  tag: number,
  u16: (o: number) => number,
  u32: (o: number) => number,
): number | null {
  if (ifd + 2 > tiff.length) return null
  const n = u16(ifd)
  for (let i = 0; i < n; i++) {
    const o = ifd + 2 + i * 12
    if (o + 12 > tiff.length) return null
    if (u16(o) === tag && u16(o + 2) === 4) return u32(o + 8)
  }
  return null
}

function buscarTexto(
  tiff: Uint8Array,
  ifd: number,
  tag: number,
  u16: (o: number) => number,
  u32: (o: number) => number,
): string | null {
  if (ifd + 2 > tiff.length) return null
  const n = u16(ifd)
  for (let i = 0; i < n; i++) {
    const o = ifd + 2 + i * 12
    if (o + 12 > tiff.length) return null
    if (u16(o) !== tag || u16(o + 2) !== 2) continue
    const count = u32(o + 4)
    const start = count <= 4 ? o + 8 : u32(o + 8)
    if (start + count > tiff.length) return null
    return ascii(tiff.subarray(start, start + count))
  }
  return null
}

function racionalGps(
  tiff: Uint8Array,
  ifd: number,
  tag: number,
  u16: (o: number) => number,
  u32: (o: number) => number,
): number | null {
  if (ifd + 2 > tiff.length) return null
  const n = u16(ifd)
  for (let i = 0; i < n; i++) {
    const o = ifd + 2 + i * 12
    if (o + 12 > tiff.length) return null
    if (u16(o) !== tag || u16(o + 2) !== 5) continue
    const start = u32(o + 8)
    if (start + 24 > tiff.length) return null
    const partes = [0, 1, 2].map((k) => {
      const num = u32(start + k * 8)
      const den = u32(start + k * 8 + 4)
      return den === 0 ? 0 : num / den
    })
    return partes[0] + partes[1] / 60 + partes[2] / 3600
  }
  return null
}

function ascii(bytes: Uint8Array): string {
  let out = ""
  for (const b of bytes) {
    if (b === 0) break
    out += String.fromCharCode(b)
  }
  return out
}
