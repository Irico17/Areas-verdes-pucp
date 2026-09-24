import {
  ABIERTOS,
  ESTADOS,
  TIPOS,
  etiquetaEstado,
  etiquetaEvento,
  etiquetaTipo,
  type Capataz,
  type Evento,
} from "../operacion"
import type { Rol } from "../types"

export type LaborItem = {
  id: string
  titulo: string
  tipo: string
  estado: string
  equipo: string
  detalle: string
  capatazId: string
  queued?: boolean
}

type Props = {
  rol: Rol
  equipos: Capataz[]
  equipoId: string
  onEquipo: (id: string) => void
  items: LaborItem[]
  estados: Record<string, boolean>
  onToggleEstado: (id: string) => void
  tipo: string
  onTipo: (id: string) => void
  pinMode: boolean
  onPinMode: (on: boolean) => void
  draft: { lon: number; lat: number } | null
  formTipo: string
  formTitulo: string
  formDetalle: string
  formEquipo: string
  onForm: (patch: { tipo?: string; titulo?: string; detalle?: string; equipo?: string }) => void
  onCreate: () => void
  creating: boolean
  selected: LaborItem | null
  onSelect: (id: string) => void
  timeline: Evento[]
  timelineError: string
  estadoNuevo: string
  onEstadoNuevo: (id: string) => void
  onEstado: () => void
  reasignarA: string
  onReasignarA: (id: string) => void
  onReasignar: () => void
  onArchivar: () => void
  confirmarArchivo: boolean
  notice: string
  queueCount: number
  onFlush: () => void
}

export function Labores(props: Props) {
  const puedeAsignar = props.rol !== "capataz"
  return (
    <section className="block">
      <h2>Labores</h2>
      <p className="lede">Puntos abiertos. El color es el estado y la letra, el tipo.</p>
      {props.rol === "capataz" && (
        <label className="field">
          Equipo en vista
          <select value={props.equipoId} onChange={(event) => props.onEquipo(event.target.value)}>
            {props.equipos.map((equipo) => (
              <option key={equipo.id} value={equipo.id}>
                {equipo.equipo} · {equipo.turno}
              </option>
            ))}
          </select>
        </label>
      )}
      <div className="checks">
        {ESTADOS.filter((estado) => (ABIERTOS as readonly string[]).includes(estado.id)).map((estado) => (
          <label key={estado.id}>
            <input
              type="checkbox"
              checked={props.estados[estado.id] !== false}
              onChange={() => props.onToggleEstado(estado.id)}
            />
            <i style={{ background: estado.color }} />
            {estado.label}
          </label>
        ))}
      </div>
      <label className="field">
        Tipo
        <select value={props.tipo} onChange={(event) => props.onTipo(event.target.value)}>
          <option value="">Todos</option>
          {TIPOS.map((tipo) => (
            <option key={tipo.id} value={tipo.id}>
              {tipo.marca} · {tipo.label}
            </option>
          ))}
        </select>
      </label>
      {puedeAsignar && (
        <button type="button" className={props.pinMode ? "primary on" : "primary"} onClick={() => props.onPinMode(!props.pinMode)}>
          {props.pinMode ? "Cancelar marca" : "Marcar labor"}
        </button>
      )}
      {props.pinMode && !props.draft && <p className="hint">Haga clic en el mapa para ubicar la labor.</p>}
      {props.draft && puedeAsignar && (
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            props.onCreate()
          }}
        >
          <p className="hint">
            {props.draft.lat.toFixed(5)}, {props.draft.lon.toFixed(5)}
          </p>
          <label className="field">
            Tipo
            <select value={props.formTipo} onChange={(event) => props.onForm({ tipo: event.target.value })}>
              {TIPOS.map((tipo) => (
                <option key={tipo.id} value={tipo.id}>
                  {tipo.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            Título
            <input value={props.formTitulo} maxLength={160} onChange={(event) => props.onForm({ titulo: event.target.value })} required />
          </label>
          <label className="field">
            Detalle
            <textarea value={props.formDetalle} maxLength={2000} rows={3} onChange={(event) => props.onForm({ detalle: event.target.value })} />
          </label>
          <label className="field">
            Equipo
            <select value={props.formEquipo} onChange={(event) => props.onForm({ equipo: event.target.value })}>
              <option value="">Sin asignar</option>
              {props.equipos.map((equipo) => (
                <option key={equipo.id} value={equipo.id}>
                  {equipo.equipo}
                </option>
              ))}
            </select>
          </label>
          <button type="submit" className="primary" disabled={props.creating}>
            {props.creating ? "Guardando…" : "Crear labor"}
          </button>
        </form>
      )}
      {props.queueCount > 0 && (
        <p className="hint">
          {props.queueCount} en cola local.{" "}
          <button type="button" className="link" onClick={props.onFlush}>
            Reintentar envío
          </button>
        </p>
      )}
      {props.notice && <p className="hint">{props.notice}</p>}
      <ul className="labor-list">
        {props.items.length === 0 && <li className="empty">No hay labores con este filtro.</li>}
        {props.items.map((item) => (
          <li key={item.id}>
            <button type="button" className={props.selected?.id === item.id ? "labor on" : "labor"} onClick={() => props.onSelect(item.id)}>
              <span className="marca" data-estado={item.estado}>
                {TIPOS.find((tipo) => tipo.id === item.tipo)?.marca ?? "·"}
              </span>
              <span>
                <strong>{item.titulo}</strong>
                <small>
                  {etiquetaTipo(item.tipo)} · {etiquetaEstado(item.estado)} · {item.equipo || "Sin equipo"}
                  {item.queued ? " · pendiente de envío" : ""}
                </small>
              </span>
            </button>
          </li>
        ))}
      </ul>
      {props.selected && (
        <div className="detail">
          <h3>{props.selected.titulo}</h3>
          <p className="meta">
            {etiquetaTipo(props.selected.tipo)} · {etiquetaEstado(props.selected.estado)} · {props.selected.equipo || "Sin equipo"}
          </p>
          {props.selected.detalle && <p className="lede">{props.selected.detalle}</p>}
          {props.selected.queued ? (
            <p className="hint">Aún no está en el servidor. El id ya quedó reservado para el reintento.</p>
          ) : (
            <>
              <label className="field">
                Estado
                <select value={props.estadoNuevo} onChange={(event) => props.onEstadoNuevo(event.target.value)}>
                  {ESTADOS.map((estado) => (
                    <option key={estado.id} value={estado.id}>
                      {estado.label}
                    </option>
                  ))}
                </select>
              </label>
              <button type="button" className="primary" onClick={props.onEstado}>
                Guardar estado
              </button>
              {puedeAsignar && (
                <>
                  <label className="field">
                    Reasignar a
                    <select value={props.reasignarA} onChange={(event) => props.onReasignarA(event.target.value)}>
                      {props.equipos.map((equipo) => (
                        <option key={equipo.id} value={equipo.id}>
                          {equipo.equipo}
                        </option>
                      ))}
                    </select>
                  </label>
                  <div className="row-actions">
                    <button type="button" onClick={props.onReasignar}>
                      Reasignar
                    </button>
                    <button type="button" className="danger" onClick={props.onArchivar}>
                      {props.confirmarArchivo ? "Confirmar archivo" : "Archivar"}
                    </button>
                  </div>
                </>
              )}
              <h3>Bitácora</h3>
              {props.timelineError && <p className="status error">{props.timelineError}</p>}
              <ol className="timeline">
                {props.timeline.map((evento) => (
                  <li key={evento.id}>
                    <strong>{etiquetaEvento(evento.tipo)}</strong>
                    <span>
                      {evento.estado ? etiquetaEstado(evento.estado) : ""}
                      {evento.equipo ? ` · ${evento.equipo}` : ""}
                      {` · ${evento.actor_rol}`}
                    </span>
                    <time>{evento.created_at.replace("T", " ").slice(0, 16)}</time>
                  </li>
                ))}
              </ol>
            </>
          )}
        </div>
      )}
    </section>
  )
}
