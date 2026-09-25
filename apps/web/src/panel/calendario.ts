export type VistaCalendario = "mes" | "semana" | "dia"

export type ReservaCalendario = {
  id: number | string
  fecha: string
  hora_inicio: string
  hora_fin: string
  estado: string
  evento: string
  unidad: string
}

const DIA = 86_400_000

export function parseISO(fecha: string): Date {
  const [y, m, d] = fecha.split("-").map(Number)
  return new Date(y, (m || 1) - 1, d || 1)
}

export function aISO(fecha: Date): string {
  const m = String(fecha.getMonth() + 1).padStart(2, "0")
  const d = String(fecha.getDate()).padStart(2, "0")
  return `${fecha.getFullYear()}-${m}-${d}`
}

export function inicioSemana(fecha: Date): Date {
  const copia = new Date(fecha.getFullYear(), fecha.getMonth(), fecha.getDate())
  const dia = (copia.getDay() + 6) % 7
  copia.setDate(copia.getDate() - dia)
  return copia
}

export function rangoDe(vista: VistaCalendario, cursor: Date): { desde: string; hasta: string } {
  if (vista === "dia") {
    const iso = aISO(cursor)
    return { desde: iso, hasta: iso }
  }
  if (vista === "semana") {
    const inicio = inicioSemana(cursor)
    const fin = new Date(inicio.getTime() + 6 * DIA)
    return { desde: aISO(inicio), hasta: aISO(fin) }
  }
  const inicio = new Date(cursor.getFullYear(), cursor.getMonth(), 1)
  const fin = new Date(cursor.getFullYear(), cursor.getMonth() + 1, 0)
  return { desde: aISO(inicio), hasta: aISO(fin) }
}

export function diasDelMes(cursor: Date): Date[] {
  const primero = new Date(cursor.getFullYear(), cursor.getMonth(), 1)
  const inicio = inicioSemana(primero)
  const dias: Date[] = []
  for (let i = 0; i < 42; i++) dias.push(new Date(inicio.getTime() + i * DIA))
  return dias
}

export function diasDeSemana(cursor: Date): Date[] {
  const inicio = inicioSemana(cursor)
  return Array.from({ length: 7 }, (_, i) => new Date(inicio.getTime() + i * DIA))
}

export function mover(vista: VistaCalendario, cursor: Date, delta: number): Date {
  if (vista === "mes") return new Date(cursor.getFullYear(), cursor.getMonth() + delta, 1)
  if (vista === "semana") return new Date(cursor.getTime() + delta * 7 * DIA)
  return new Date(cursor.getTime() + delta * DIA)
}

export function reservasDelDia(filas: ReservaCalendario[], iso: string): ReservaCalendario[] {
  return filas
    .filter((fila) => fila.fecha === iso)
    .slice()
    .sort((a, b) => a.hora_inicio.localeCompare(b.hora_inicio))
}

export function tituloVista(vista: VistaCalendario, cursor: Date): string {
  const fmt = new Intl.DateTimeFormat("es-PE", vista === "dia"
    ? { weekday: "long", day: "numeric", month: "long", year: "numeric" }
    : { month: "long", year: "numeric" })
  const texto = fmt.format(cursor)
  return texto.charAt(0).toUpperCase() + texto.slice(1)
}

export function queryReservas(desde: string, hasta: string): string {
  const params = new URLSearchParams({ desde, hasta })
  return `/api/v1/inventario/reservas?${params.toString()}`
}
