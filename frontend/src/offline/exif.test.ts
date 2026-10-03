import assert from "node:assert/strict"
import test from "node:test"
import { leerExif } from "./exif.ts"

test("lee latitud, longitud y fecha de un JPEG con EXIF", () => {
  const punto = leerExif(jpegConGps())
  assert.equal(punto.fecha, "2026:09:25 10:00:00")
  assert.ok(punto.lat != null && Math.abs(punto.lat + 12.07) < 0.0001)
  assert.ok(punto.lon != null && Math.abs(punto.lon + 77.08) < 0.0001)
})

test("un archivo sin EXIF no trae punto", () => {
  const punto = leerExif(new Uint8Array([0xff, 0xd8, 0xff, 0xd9]))
  assert.equal(punto.lat, null)
  assert.equal(punto.lon, null)
  assert.equal(punto.orientacion, null)
})

test("lee la orientación del JPEG", () => {
  assert.equal(leerExif(jpegConOrientacion(6)).orientacion, 6)
})

function jpegConOrientacion(valor: number): Uint8Array {
  const tiff = new Uint8Array(26)
  const v = new DataView(tiff.buffer)
  v.setUint8(0, 0x49)
  v.setUint8(1, 0x49)
  v.setUint16(2, 42, true)
  v.setUint32(4, 8, true)
  v.setUint16(8, 1, true)
  v.setUint16(10, 0x0112, true)
  v.setUint16(12, 3, true)
  v.setUint32(14, 1, true)
  v.setUint16(18, valor, true)
  const body = new Uint8Array(6 + tiff.length)
  body.set(ascii("Exif\0\0"), 0)
  body.set(tiff, 6)
  const out = new Uint8Array(6 + body.length + 2)
  out[0] = 0xff
  out[1] = 0xd8
  out[2] = 0xff
  out[3] = 0xe1
  out[4] = (body.length + 2) >> 8
  out[5] = (body.length + 2) & 0xff
  out.set(body, 6)
  out[out.length - 2] = 0xff
  out[out.length - 1] = 0xd9
  return out
}

function jpegConGps(): Uint8Array {
  const fecha = ascii("2026:09:25 10:00:00")
  const tiff = new Uint8Array(200)
  const v = new DataView(tiff.buffer)
  v.setUint8(0, 0x49)
  v.setUint8(1, 0x49)
  v.setUint16(2, 42, true)
  v.setUint32(4, 8, true)
  v.setUint16(8, 2, true)
  const fechaOff = 8 + 2 + 2 * 12 + 4
  escribirEntrada(v, 10, 0x0132, 2, fecha.length, fechaOff)
  const gpsOff = fechaOff + fecha.length
  escribirEntrada(v, 22, 0x8825, 4, 1, gpsOff)
  v.setUint32(34, 0, true)
  tiff.set(fecha, fechaOff)
  v.setUint16(gpsOff, 4, true)
  const rat = gpsOff + 2 + 4 * 12 + 4
  escribirEntrada(v, gpsOff + 2, 0x0001, 2, 2, 0)
  v.setUint8(gpsOff + 10, 83)
  v.setUint8(gpsOff + 11, 0)
  escribirEntrada(v, gpsOff + 14, 0x0002, 5, 3, rat)
  escribirEntrada(v, gpsOff + 26, 0x0003, 2, 2, 0)
  v.setUint8(gpsOff + 34, 87)
  v.setUint8(gpsOff + 35, 0)
  escribirEntrada(v, gpsOff + 38, 0x0004, 5, 3, rat + 24)
  v.setUint32(gpsOff + 2 + 4 * 12, 0, true)
  racional(v, rat, [12, 1, 4, 1, 12, 1])
  racional(v, rat + 24, [77, 1, 4, 1, 48, 1])
  const body = new Uint8Array(6 + tiff.length)
  body.set(ascii("Exif\0\0"), 0)
  body.set(tiff, 6)
  const out = new Uint8Array(4 + body.length + 2)
  out[0] = 0xff
  out[1] = 0xd8
  out[2] = 0xff
  out[3] = 0xe1
  out[4] = (body.length + 2) >> 8
  out[5] = (body.length + 2) & 0xff
  out.set(body, 6)
  out[out.length - 2] = 0xff
  out[out.length - 1] = 0xd9
  return out
}

function escribirEntrada(v: DataView, offset: number, tag: number, tipo: number, count: number, valor: number) {
  v.setUint16(offset, tag, true)
  v.setUint16(offset + 2, tipo, true)
  v.setUint32(offset + 4, count, true)
  v.setUint32(offset + 8, valor, true)
}

function racional(v: DataView, offset: number, nums: number[]) {
  for (let i = 0; i < nums.length; i++) v.setUint32(offset + i * 4, nums[i], true)
}

function ascii(text: string): Uint8Array {
  return Uint8Array.from(text, (ch) => ch.charCodeAt(0))
}
