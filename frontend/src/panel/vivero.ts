export type ViveroItem = {
  id: string
  fecha: string
  area: string
  subproceso: string
  etapa: string
  descripcion: string
  observaciones: string
  responsables: string
  lugar: string
}

export const AREAS = ["Fauna", "Flora", "Ambiental", "Otros"] as const

export const VIVERO_VACIO: ViveroItem = {
  id: "",
  fecha: "",
  area: "Flora",
  subproceso: "",
  etapa: "",
  descripcion: "",
  observaciones: "",
  responsables: "",
  lugar: "",
}

export function validarVivero(item: ViveroItem, catalogo: { subproceso: string[]; etapa: string[] }): { campo: string; motivo: string }[] {
  const fallos: { campo: string; motivo: string }[] = []
  if (item.area && !AREAS.includes(item.area as (typeof AREAS)[number])) {
    fallos.push({ campo: "area", motivo: "El área tiene que estar en el catálogo." })
  }
  if (item.subproceso && catalogo.subproceso.length > 0 && !catalogo.subproceso.includes(item.subproceso)) {
    fallos.push({ campo: "subproceso", motivo: "El subproceso se da de alta al importar; después no es texto libre." })
  }
  if (item.etapa && catalogo.etapa.length > 0 && !catalogo.etapa.includes(item.etapa)) {
    fallos.push({ campo: "etapa", motivo: "La etapa se da de alta al importar; después no es texto libre." })
  }
  return fallos
}
