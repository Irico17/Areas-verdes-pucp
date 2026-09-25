import assert from "node:assert/strict"
import test from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { CalendarioReservas } from "./CalendarioReservas.tsx"
import { mover, queryReservas, rangoDe, reservasDelDia, type ReservaCalendario } from "./calendario.ts"

const filas: ReservaCalendario[] = [
  { id: 1, fecha: "2026-09-22", hora_inicio: "11:00", hora_fin: "12:00", estado: "reservado", evento: "Tarde", unidad: "EEGGLL" },
  { id: 2, fecha: "2026-09-22", hora_inicio: "09:00", hora_fin: "10:00", estado: "realizado", evento: "Mañana", unidad: "" },
  { id: 3, fecha: "2026-09-23", hora_inicio: "08:00", hora_fin: "09:00", estado: "cancelado", evento: "Otro", unidad: "" },
]

test("el mes, la semana y el día acotan el filtro de la API", () => {
  const cursor = new Date(2026, 8, 22)
  assert.deepEqual(rangoDe("dia", cursor), { desde: "2026-09-22", hasta: "2026-09-22" })
  const semana = rangoDe("semana", cursor)
  assert.equal(semana.desde <= "2026-09-22" && semana.hasta >= "2026-09-22", true)
  assert.equal(rangoDe("mes", cursor).desde, "2026-09-01")
  assert.equal(rangoDe("mes", cursor).hasta, "2026-09-30")
  assert.equal(queryReservas("2026-09-01", "2026-09-30"), "/api/v1/inventario/reservas?desde=2026-09-01&hasta=2026-09-30")
  assert.equal(mover("mes", cursor, 1).getMonth(), 9)
})

test("las reservas del día salen por hora", () => {
  const del = reservasDelDia(filas, "2026-09-22")
  assert.deepEqual(del.map((fila) => fila.evento), ["Mañana", "Tarde"])
  assert.equal(reservasDelDia(filas, "2026-09-01").length, 0)
})

test("el calendario pinta mes, semana y día", () => {
  const html = renderToStaticMarkup(createElement(CalendarioReservas, {
    cliente: (() => Promise.reject(new Error("sin red"))) as typeof fetch,
  }))
  for (const texto of ["Mes", "Semana", "Día", "Reservas", "Anterior", "Siguiente"]) {
    assert.ok(html.includes(texto), texto)
  }
})
