export const TRAZO_CLASE: Record<string, string> = {
  habilitacion: "M12 21V10 M12 14c-4 0-5-5-3-7 0 4 3 5 3 5s3-1 3-5c2 2 1 7-3 7 M8 21h8",
  rehabilitacion: "M19 12a7 7 0 1 1-2.5-5.8 M16 3v4h-4",
  mantenimiento: "M5 8h14 M5 12h14 M5 16h14 M8 8v10 M16 8v10",
  poda: "M8 8l8 8 M16 8l-8 8 M7 7a2 2 0 1 0 .01 0 M17 7a2 2 0 1 0 .01 0",
  propagacion: "M12 21V12 M12 14C8 12 6 8 8 5c2 3 4 6 4 9 M12 14c4-2 6-6 4-9-2 3-4 6-4 9",
  riego: "M12 3s-6 8-6 12a6 6 0 0 0 12 0c0-4-6-12-6-12z",
  residuos: "M6 16c0-4 3-5 6-3 3-2 6-1 6 3 0 3-3 4-6 2-3 2-6 1-6-2z M8 11c0-3 2-4 4-2 2-2 4-1 4 2",
  fitosanitario: "M12 3l7 3v6c0 4-3 7-7 9-4-2-7-5-7-9V6z",
  inspeccion_monitoreo: "M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6z M12 9a3 3 0 1 0 .01 0",
  generica: "M12 4a8 8 0 1 0 .01 0",
}

const TIPO_A_CLASE: Record<string, string> = {
  riego: "riego",
  poda: "poda",
  limpieza: "mantenimiento",
  incidencia: "fitosanitario",
  inspeccion: "inspeccion_monitoreo",
}

/** Icono de la clase. Si la clase no viene, se usa el tipo grueso. Nunca una letra. */
export function iconoDe(clase: string, tipo: string): string {
  if (clase && TRAZO_CLASE[clase] && clase !== "generica") return clase
  if (TIPO_A_CLASE[tipo]) return TIPO_A_CLASE[tipo]
  return "generica"
}

export function pintarIcono(path: string, size = 32): ImageData {
  const canvas = document.createElement("canvas")
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext("2d")
  if (!ctx) throw new Error("sin lienzo")
  ctx.scale(size / 24, size / 24)
  ctx.strokeStyle = "#f6f3ec"
  ctx.lineWidth = 1.7
  ctx.lineJoin = "round"
  ctx.lineCap = "round"
  ctx.stroke(new Path2D(path))
  return ctx.getImageData(0, 0, size, size)
}
