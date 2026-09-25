import { useEffect, useState } from "react"
import {
  aISO,
  diasDelMes,
  diasDeSemana,
  mover,
  queryReservas,
  rangoDe,
  reservasDelDia,
  tituloVista,
  type ReservaCalendario,
  type VistaCalendario,
} from "./calendario"

const VISTAS: { id: VistaCalendario; label: string }[] = [
  { id: "mes", label: "Mes" },
  { id: "semana", label: "Semana" },
  { id: "dia", label: "Día" },
]

const DIAS = ["Lun", "Mar", "Mié", "Jue", "Vie", "Sáb", "Dom"]

function leerReservas(body: unknown): ReservaCalendario[] {
  const raw = body && typeof body === "object" ? (body as { reservas?: unknown }).reservas : null
  if (!Array.isArray(raw)) return []
  return raw.flatMap((item) => {
    if (!item || typeof item !== "object") return []
    const row = item as Record<string, unknown>
    const fecha = String(row.fecha ?? "")
    if (!/^\d{4}-\d{2}-\d{2}/.test(fecha)) return []
    return [{
      id: (row.id as number | string) ?? fecha,
      fecha: fecha.slice(0, 10),
      hora_inicio: String(row.hora_inicio ?? ""),
      hora_fin: String(row.hora_fin ?? ""),
      estado: String(row.estado ?? ""),
      evento: String(row.evento ?? "Reserva"),
      unidad: String(row.unidad ?? ""),
    }]
  })
}

export function CalendarioReservas({ cliente = fetch }: { cliente?: typeof fetch }) {
  const [vista, setVista] = useState<VistaCalendario>("mes")
  const [cursor, setCursor] = useState(() => new Date())
  const [filas, setFilas] = useState<ReservaCalendario[]>([])
  const [aviso, setAviso] = useState("Agenda ficticia. La hoja institucional responde 401 y no se abre.")
  const [error, setError] = useState("")

  useEffect(() => {
    const { desde, hasta } = rangoDe(vista, cursor)
    let vivo = true
    cliente(queryReservas(desde, hasta), { credentials: "include", headers: { Accept: "application/json" } })
      .then(async (res) => {
        if (!res.ok) throw new Error("No se pudieron leer las reservas")
        return res.json()
      })
      .then((body: unknown) => {
        if (!vivo) return
        const avisoAPI = body && typeof body === "object" ? String((body as { aviso?: string }).aviso ?? "") : ""
        if (avisoAPI) setAviso(avisoAPI)
        setFilas(leerReservas(body))
        setError("")
      })
      .catch((err: unknown) => {
        if (!vivo) return
        setFilas([])
        setError(err instanceof Error ? err.message : "No se pudieron leer las reservas")
      })
    return () => {
      vivo = false
    }
  }, [cliente, cursor, vista])

  const titulo = tituloVista(vista, cursor)
  const hoy = aISO(new Date())

  return (
    <section className="calendario" aria-label="Calendario de reservas">
      <div className="cal-head">
        <h2>Reservas</h2>
        <div className="roles" role="tablist" aria-label="Vista del calendario">
          {VISTAS.map((item) => (
            <button key={item.id} type="button" role="tab" aria-selected={vista === item.id} onClick={() => setVista(item.id)}>
              {item.label}
            </button>
          ))}
        </div>
      </div>
      <div className="cal-nav">
        <button type="button" onClick={() => setCursor((fecha) => mover(vista, fecha, -1))} aria-label="Periodo anterior">
          Anterior
        </button>
        <p className="cal-titulo">{titulo}</p>
        <button type="button" onClick={() => setCursor(new Date())}>Hoy</button>
        <button type="button" onClick={() => setCursor((fecha) => mover(vista, fecha, 1))} aria-label="Periodo siguiente">
          Siguiente
        </button>
      </div>
      <p className="hint">{aviso}</p>
      {error && <p className="status error">{error}</p>}
      {vista === "mes" && (
        <div className="cal-mes">
          {DIAS.map((dia) => (
            <span key={dia} className="cal-dow">{dia}</span>
          ))}
          {diasDelMes(cursor).map((dia) => {
            const iso = aISO(dia)
            const eventos = reservasDelDia(filas, iso)
            const otroMes = dia.getMonth() !== cursor.getMonth()
            return (
              <button
                key={iso + (otroMes ? "-f" : "")}
                type="button"
                className={iso === hoy ? "cal-dia hoy" : otroMes ? "cal-dia fuera" : "cal-dia"}
                onClick={() => {
                  setCursor(dia)
                  setVista("dia")
                }}
              >
                <span>{dia.getDate()}</span>
                {eventos.slice(0, 2).map((evento) => (
                  <small key={String(evento.id)}>{evento.hora_inicio.slice(0, 5)} {evento.evento}</small>
                ))}
                {eventos.length > 2 && <small>+{eventos.length - 2}</small>}
              </button>
            )
          })}
        </div>
      )}
      {vista === "semana" && (
        <div className="cal-semana">
          {diasDeSemana(cursor).map((dia) => {
            const iso = aISO(dia)
            const eventos = reservasDelDia(filas, iso)
            return (
              <section key={iso} className={iso === hoy ? "cal-col hoy" : "cal-col"}>
                <h3>{DIAS[(dia.getDay() + 6) % 7]} {dia.getDate()}</h3>
                {eventos.length === 0 && <p className="empty">Sin reservas</p>}
                {eventos.map((evento) => (
                  <article key={String(evento.id)} className={`cal-evento ${evento.estado}`}>
                    <strong>{evento.evento}</strong>
                    <small>{evento.hora_inicio.slice(0, 5)}–{evento.hora_fin.slice(0, 5)} · {evento.estado}</small>
                    {evento.unidad && <small>{evento.unidad}</small>}
                  </article>
                ))}
              </section>
            )
          })}
        </div>
      )}
      {vista === "dia" && (
        <ul className="labor-list">
          {reservasDelDia(filas, aISO(cursor)).length === 0 && <li className="empty">Este día no tiene reservas.</li>}
          {reservasDelDia(filas, aISO(cursor)).map((evento) => (
            <li key={String(evento.id)} className="agenda">
              <strong>{evento.evento}</strong>
              <small>{evento.hora_inicio.slice(0, 5)}–{evento.hora_fin.slice(0, 5)} · {evento.estado}{evento.unidad ? ` · ${evento.unidad}` : ""}</small>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
