import { useEffect, useState } from "react"
import { apiUrl } from "../api"
import { formatFechaHora } from "../fecha"
import { Esqueleto } from "../ui/Esqueleto"

type EventoAuditoria = {
  id: number
  entidad: string
  entidad_id: string
  accion: string
  usuario: string
  nombre: string
  created_at: string
  antes?: string
  despues?: string
}

const RETENCION =
  "No hay un plazo de retención. El cliente no lo fijó. Este historial no se borra por antigüedad."

function actor(evento: EventoAuditoria): string {
  const nombre = evento.nombre?.trim()
  if (nombre) return nombre
  const usuario = evento.usuario?.trim()
  if (usuario) return usuario
  return "Semilla"
}

export function AuditoriaPanel() {
  const [eventos, setEventos] = useState<EventoAuditoria[] | null>(null)
  const [error, setError] = useState("")

  useEffect(() => {
    let vivo = true
    fetch(apiUrl("/auditoria/cambios"), { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) {
          const body = (await res.json().catch(() => null)) as { error?: string } | null
          throw new Error(body?.error || "No se pudo leer el historial")
        }
        return res.json() as Promise<{ eventos?: EventoAuditoria[] }>
      })
      .then((body) => {
        if (vivo) setEventos(body.eventos ?? [])
      })
      .catch((err: unknown) => {
        if (vivo) setError(err instanceof Error ? err.message : "No se pudo leer el historial")
      })
    return () => {
      vivo = false
    }
  }, [])

  return (
    <section className="historial-panel" aria-labelledby="historial-titulo">
      <h2 id="historial-titulo">Historial</h2>
      <p className="lede">Cambios de la semilla y de la sesión, con el antes y el después.</p>
      <p className="historial-retencion">{RETENCION}</p>
      {error && <p className="status error">{error}</p>}
      {eventos === null && !error && <Esqueleto filas={4} />}
      {eventos?.length === 0 && <p className="hint">Todavía no hay cambios.</p>}
      {eventos && eventos.length > 0 && (
        <ul className="historial-lista">
          {eventos.map((evento) => (
            <li key={evento.id}>
              <strong>{evento.entidad}</strong>
              <span>{evento.accion}</span>
              <span>{actor(evento)}</span>
              <time dateTime={evento.created_at}>{formatFechaHora(evento.created_at)}</time>
              {(evento.antes || evento.despues) && (
                <p>
                  {evento.antes || "—"}
                  {" → "}
                  {evento.despues || "—"}
                </p>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
