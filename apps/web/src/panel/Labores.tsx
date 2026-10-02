import { useEffect, useRef, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFechaHora } from "../fecha"
import { mostrarEnPanel } from "../ui/desplazar"
import { Esqueleto } from "../ui/Esqueleto"
import {
  ABIERTOS,
  ESTADOS,
  estadosPermitidos,
  TIPOS,
  etiquetaEstado,
  etiquetaEvento,
  etiquetaTipo,
  type Capataz,
  type Evento,
} from "../operacion"
import { etiquetaRol, type CatalogoItem } from "../producto"
import type { Rol } from "../types"
import { ACTIVIDAD, marcaEjecutor } from "../ui/nomenclatura"
import { EvidenciasCampo } from "./EvidenciasCampo"
import { apiUrl } from "../api"

export type LaborItem = {
  id: string
  titulo: string
  tipo: string
  estado: string
  equipo: string
  detalle: string
  capatazId: string
  ejecutor?: string
  queued?: boolean
}

type Props = {
  rol: Rol
  cargando?: boolean
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
  onForm: (patch: { tipo?: string; titulo?: string; detalle?: string; equipo?: string; ejecutor?: string }) => void
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
  tipos: { id: string; label: string }[]
  formEjecutor: string
  motivos: CatalogoItem[]
  motivo: string
  onMotivo: (id: string) => void
  onSugerir: () => void
  sugerencia: string
  pista?: { codigo: string; etiqueta: string } | null
  onConfirmarPista?: () => void
}

export function Labores(props: Props) {
  const puedeAsignar = props.rol !== "capataz"
  const [equipoVista, setEquipoVista] = useState("")
  const visibles = props.items.filter((item) => !equipoVista || item.capatazId === equipoVista)
  const avisoError = /no se|sin conexión|error/i.test(props.notice)
  const detalleRef = useRef<HTMLDivElement>(null)
  const elegida = props.selected?.id
  useEffect(() => {
    if (elegida) mostrarEnPanel(detalleRef.current, "inicio")
  }, [elegida])
  return (
    <section className="block">
      <h2>{ACTIVIDAD.titulo}</h2>
      <p className="lede">{ACTIVIDAD.lede}</p>
      {props.rol === "capataz" && <p className="hint">{ACTIVIDAD.soloCuadrilla}</p>}
      {puedeAsignar && (
        <label className="field">
          {ACTIVIDAD.cuadrilla}
          <select value={equipoVista} onChange={(event) => setEquipoVista(event.target.value)}>
            <option value="">{ACTIVIDAD.todas}</option>
            {props.equipos.map((equipo) => (
              <option key={equipo.id} value={equipo.id}>
                {equipo.equipo}
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
        {ACTIVIDAD.tipo}
        <select value={props.tipo} onChange={(event) => props.onTipo(event.target.value)}>
          <option value="">{ACTIVIDAD.todosTipos}</option>
          {props.tipos.map((tipo) => (
            <option key={tipo.id} value={tipo.id}>
              {tipo.label}
            </option>
          ))}
        </select>
      </label>
      {puedeAsignar && (
        <button type="button" className={props.pinMode ? "primary sheet-action on" : "primary sheet-action"} onClick={() => props.onPinMode(!props.pinMode)}>
          {props.pinMode ? ACTIVIDAD.cancelarMarca : ACTIVIDAD.marcar}
        </button>
      )}
      {props.pinMode && !props.draft && <p className="hint">{ACTIVIDAD.ubicar}</p>}
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
            {ACTIVIDAD.tipo}
            <select value={props.formTipo} onChange={(event) => props.onForm({ tipo: event.target.value })}>
              {props.tipos.map((tipo) => (
                <option key={tipo.id} value={tipo.id}>
                  {tipo.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {ACTIVIDAD.tituloCampo}
            <input value={props.formTitulo} maxLength={160} onChange={(event) => props.onForm({ titulo: event.target.value })} required />
          </label>
          <p className="hint">{ACTIVIDAD.regla}</p>
          {props.sugerencia && <p className="hint">{props.sugerencia}</p>}
          <button type="button" onClick={props.onSugerir}>
            {ACTIVIDAD.sugerir}
          </button>
          {props.pista?.codigo && (
            <button type="button" onClick={props.onConfirmarPista}>
              Confirmar {props.pista.etiqueta || props.pista.codigo}
            </button>
          )}
          <label className="field">
            {ACTIVIDAD.quienEjecuta}
            <select value={props.formEjecutor} onChange={(event) => props.onForm({ ejecutor: event.target.value })}>
              <option value="propia">{ACTIVIDAD.personalPropio}</option>
              <option value="tercerizada">{ACTIVIDAD.servicioTercerizado}</option>
            </select>
          </label>
          <FichaLabor />
          <label className="field">
            {ACTIVIDAD.detalle}
            <textarea value={props.formDetalle} maxLength={2000} rows={3} onChange={(event) => props.onForm({ detalle: event.target.value })} />
          </label>
          <label className="field">
            {ACTIVIDAD.cuadrilla}
            <select value={props.formEquipo} onChange={(event) => props.onForm({ equipo: event.target.value })}>
              <option value="">{ACTIVIDAD.sinAsignar}</option>
              {props.equipos.map((equipo) => (
                <option key={equipo.id} value={equipo.id}>
                  {equipo.equipo}
                </option>
              ))}
            </select>
          </label>
          <button type="submit" className="primary" disabled={props.creating}>
            {props.creating ? ACTIVIDAD.guardando : ACTIVIDAD.crear}
          </button>
        </form>
      )}
      {props.queueCount > 0 && (
        <p className="hint">
          {ACTIVIDAD.colaLocal(props.queueCount)}{" "}
          <button type="button" className="link" onClick={props.onFlush}>
            {ACTIVIDAD.reintentar}
          </button>
        </p>
      )}
      {props.notice && (
        <p className={avisoError ? "status error" : "banner"} role="status">
          {props.notice}
        </p>
      )}
      {props.cargando && <Esqueleto />}
      <ul className="labor-list">
        {!props.cargando && visibles.length === 0 && <li className="empty">{ACTIVIDAD.vacio}</li>}
        {visibles.map((item) => (
          <li key={item.id}>
            <button type="button" className={props.selected?.id === item.id ? "labor on" : "labor"} onClick={() => props.onSelect(item.id)}>
              <span className="marca" data-estado={item.estado}>
                {TIPOS.find((tipo) => tipo.id === item.tipo)?.marca ?? "·"}
              </span>
              <span>
                <strong>{item.titulo}</strong>
                <small>
                  {etiquetaTipo(item.tipo)} · {etiquetaEstado(item.estado)} · {item.equipo || ACTIVIDAD.sinCuadrilla}
                  {item.ejecutor === "tercerizada" ? ` · ${marcaEjecutor(item.ejecutor)}` : ""}
                  {item.queued ? " · pendiente de envío" : ""}
                </small>
              </span>
            </button>
          </li>
        ))}
      </ul>
      {props.selected && (
        <div className="detail" ref={detalleRef}>
          <h3>{props.selected.titulo}</h3>
          <p className="meta">
            {etiquetaTipo(props.selected.tipo)} · {etiquetaEstado(props.selected.estado)} · {props.selected.equipo || ACTIVIDAD.sinCuadrilla}
            {` · ${marcaEjecutor(props.selected.ejecutor) || ACTIVIDAD.propioMarca}`}
          </p>
          {props.selected.detalle && <p className="lede">{props.selected.detalle}</p>}
          <FichaLabor key={props.selected.id} actividadId={props.selected.queued ? "" : props.selected.id} />
          {props.selected.queued ? (
            <p className="hint">Aún no está en el servidor. El id ya quedó reservado para el reintento.</p>
          ) : (
            <>
              <label className="field">
                {ACTIVIDAD.estado}
                <select value={props.estadoNuevo} onChange={(event) => props.onEstadoNuevo(event.target.value)}>
                  {estadosPermitidos(props.rol).map((estado) => (
                    <option key={estado.id} value={estado.id}>
                      {estado.label}
                    </option>
                  ))}
                </select>
              </label>
              <button type="button" className="primary" onClick={props.onEstado}>
                {ACTIVIDAD.guardarEstado}
              </button>
              {puedeAsignar && (
                <>
                  <label className="field">
                    {ACTIVIDAD.reasignarA}
                    <select value={props.reasignarA} onChange={(event) => props.onReasignarA(event.target.value)}>
                      {props.equipos.map((equipo) => (
                        <option key={equipo.id} value={equipo.id}>
                          {equipo.equipo}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="field">
                    {ACTIVIDAD.motivo}
                    <select value={props.motivo} onChange={(event) => props.onMotivo(event.target.value)}>
                      <option value="">{ACTIVIDAD.elegir}</option>
                      {props.motivos.map((item) => (
                        <option key={item.codigo} value={item.codigo}>
                          {item.nombre}
                        </option>
                      ))}
                    </select>
                  </label>
                  <div className="row-actions">
                    <button type="button" onClick={props.onReasignar}>
                      {ACTIVIDAD.reasignar}
                    </button>
                    <button type="button" className="danger" onClick={props.onArchivar}>
                      {props.confirmarArchivo ? ACTIVIDAD.confirmarArchivo : ACTIVIDAD.archivar}
                    </button>
                  </div>
                </>
              )}
              <EvidenciasCampo key={props.selected.id} actividadId={props.selected.queued ? "" : props.selected.id} />
              <h3>{ACTIVIDAD.bitacora}</h3>
              {props.timelineError && <p className="status error">{props.timelineError}</p>}
              <ol className="timeline">
                {props.timeline.map((evento) => (
                  <li key={evento.id}>
                    <strong>{etiquetaEvento(evento.tipo)}</strong>
                    <span>
                      {evento.estado ? etiquetaEstado(evento.estado) : ""}
                      {evento.equipo ? ` · ${evento.equipo}` : ""}
                      {` · ${etiquetaRol(evento.actor_rol)}`}
                    </span>
                    <time dateTime={evento.created_at}>{formatFechaHora(evento.created_at)}</time>
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

function FichaLabor(props: { actividadId?: string }) {
  const [clase, setClase] = useState("Mantenimiento de jardines")
  const [solicitud, setSolicitud] = useState("")
  const [atencion, setAtencion] = useState("")
  const [lugar, setLugar] = useState("")
  const [comentario, setComentario] = useState("")
  const [aviso, setAviso] = useState("")
  const fichaRef = useRef<HTMLFieldSetElement>(null)
  async function guardar() {
    if (fichaRef.current?.querySelector('[aria-invalid="true"]')) {
      setAviso("Corrija la fecha antes de guardar.")
      return
    }
    if (solicitud && atencion && atencion < solicitud) {
      setAviso("La atención no puede ser anterior a la solicitud.")
      return
    }
    if (!props.actividadId) {
      setAviso(ACTIVIDAD.creePrimero)
      return
    }
    try {
      const res = await fetch(apiUrl(`/operacion/actividades/${props.actividadId}/ficha`), {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          clase,
          fecha_solicitud: solicitud,
          fecha_atencion: atencion,
          lugar,
          comentario,
        }),
      })
      if (!res.ok) {
        setAviso("No se pudo guardar la ficha.")
        return
      }
      setAviso(lugar || "Ficha guardada. El pin del mapa no cambia.")
    } catch {
      setAviso("Sin conexión con la API.")
    }
  }
  return (
    <fieldset className="form grupo" ref={fichaRef}>
      <legend>{ACTIVIDAD.ficha}</legend>
      <label className="field">
        {ACTIVIDAD.clase}
        <input value={clase} onChange={(event) => setClase(event.target.value)} />
      </label>
      <label className="field">
        {ACTIVIDAD.fechaSolicitud}
        <FechaCampo value={solicitud} onChange={setSolicitud} />
      </label>
      <label className="field">
        {ACTIVIDAD.fechaAtencion}
        <FechaCampo value={atencion} onChange={setAtencion} />
      </label>
      <label className="field">
        {ACTIVIDAD.lugar}
        <input value={lugar} onChange={(event) => setLugar(event.target.value)} placeholder={ACTIVIDAD.lugarPlaceholder} />
      </label>
      <label className="field">
        {ACTIVIDAD.comentario}
        <textarea value={comentario} rows={2} maxLength={2000} onChange={(event) => setComentario(event.target.value)} />
      </label>
      <button type="button" onClick={() => void guardar()}>
        {ACTIVIDAD.guardarFicha}
      </button>
      {aviso && <p className="hint">{aviso}</p>}
    </fieldset>
  )
}
