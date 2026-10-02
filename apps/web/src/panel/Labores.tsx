import { useEffect, useRef, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFechaHora } from "../fecha"
import { mostrarEnPanel } from "../ui/desplazar"
import { Esqueleto } from "../ui/Esqueleto"
import {
  estadosPermitidos,
  fetchTaxonomiaActividad,
  tipoGrueso,
  etiquetaEstado,
  etiquetaEvento,
  etiquetaTipo,
  type AltaCampos,
  type Capataz,
  type Evento,
  type FiltroActividades,
  type TaxonomiaActividad,
} from "../operacion"
import { etiquetaRol, type CatalogoItem } from "../producto"
import type { Rol } from "../types"
import { IconoClase } from "../map/iconoClase"
import { ACTIVIDAD, VISTA_ACTIVIDAD, marcaEjecutor } from "../ui/nomenclatura"
import { FiltrosActividad } from "./FiltrosActividad"
import { EvidenciasCampo } from "./EvidenciasCampo"
import { SelectorLugar } from "./SelectorLugar"
import { listarZonas } from "./catastro"
import { listarLugaresCatalogo, type LugarCatalogo } from "./zonificacion"
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
  clase?: string
  queued?: boolean
}

type Props = {
  rol: Rol
  cargando?: boolean
  equipos: Capataz[]
  equipoId: string
  onEquipo: (id: string) => void
  items: LaborItem[]
  filtro?: FiltroActividades
  onFiltro?: (filtro: FiltroActividades) => void
  presentacion?: "completa" | "detalle"
  pinMode: boolean
  onPinMode: (on: boolean) => void
  draft: { lon: number; lat: number } | null
  formTipo: string
  formTitulo: string
  formDetalle: string
  formEquipo: string
  onForm: (patch: { tipo?: string; titulo?: string; detalle?: string; equipo?: string; ejecutor?: string }) => void
  onCreate: (alta: AltaCampos) => void
  taxonomia?: TaxonomiaActividad
  lugares?: LugarCatalogo[]
  zonas?: { codigo: string; nombre: string }[]
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
  const avisoError = /no se|sin conexión|error/i.test(props.notice)
  const detalleRef = useRef<HTMLDivElement>(null)
  const elegida = props.selected?.id
  useEffect(() => {
    if (elegida) mostrarEnPanel(detalleRef.current, "inicio")
  }, [elegida])
  const detalle = props.selected ? (
        <div className="detail" id="detalle-actividad" ref={detalleRef}>
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
  ) : null
  if (props.presentacion === "detalle") {
    if (!detalle) return null
    return (
      <section className="block">
        <h2>{VISTA_ACTIVIDAD.detalle}</h2>
        {detalle}
      </section>
    )
  }
  return (
    <section className="block">
      <h2>{ACTIVIDAD.titulo}</h2>
      <p className="lede">{ACTIVIDAD.lede}</p>
      {props.rol === "capataz" && <p className="hint">{ACTIVIDAD.soloCuadrilla}</p>}
      <FiltrosActividad rol={props.rol} valor={props.filtro} onChange={props.onFiltro} tipos={props.tipos} />
      {puedeAsignar && (
        <button type="button" className={props.pinMode ? "primary sheet-action on" : "primary sheet-action"} onClick={() => props.onPinMode(!props.pinMode)}>
          {props.pinMode ? ACTIVIDAD.cancelarMarca : ACTIVIDAD.marcar}
        </button>
      )}
      {props.pinMode && !props.draft && <p className="hint">{ACTIVIDAD.ubicar}</p>}
      {props.draft && puedeAsignar && (
        <AltaActividad
          draft={props.draft}
          formTipo={props.formTipo}
          formTitulo={props.formTitulo}
          formDetalle={props.formDetalle}
          formEquipo={props.formEquipo}
          formEjecutor={props.formEjecutor}
          equipos={props.equipos}
          onForm={props.onForm}
          onCreate={props.onCreate}
          creating={props.creating}
          onSugerir={props.onSugerir}
          sugerencia={props.sugerencia}
          pista={props.pista}
          onConfirmarPista={props.onConfirmarPista}
          taxonomia={props.taxonomia}
          lugares={props.lugares}
          zonas={props.zonas}
        />
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
        {!props.cargando && props.items.length === 0 && <li className="empty">{ACTIVIDAD.vacio}</li>}
        {props.items.map((item) => (
          <li key={item.id}>
            <button type="button" className={props.selected?.id === item.id ? "labor on" : "labor"} onClick={() => props.onSelect(item.id)}>
              <span className="marca" data-estado={item.estado}>
                <IconoClase clase={item.clase} tipo={item.tipo} />
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
      {detalle}
    </section>
  )
}

type AltaProps = {
  draft: { lon: number; lat: number }
  formTipo: string
  formTitulo: string
  formDetalle: string
  formEquipo: string
  formEjecutor: string
  equipos: Capataz[]
  onForm: Props["onForm"]
  onCreate: (alta: AltaCampos) => void
  creating: boolean
  onSugerir: () => void
  sugerencia: string
  pista?: { codigo: string; etiqueta: string } | null
  onConfirmarPista?: () => void
  taxonomia?: TaxonomiaActividad
  lugares?: LugarCatalogo[]
  zonas?: { codigo: string; nombre: string }[]
}

const TAX_VACIA: TaxonomiaActividad = { clases: [], riesgos: [], origenes: [], personal: [] }

function claseDePista(codigo: string): string {
  if (codigo === "inspeccion") return "inspeccion_monitoreo"
  if (codigo === "poda" || codigo === "riego") return codigo
  return ""
}

function AltaActividad(props: AltaProps) {
  const [remota, setRemota] = useState<TaxonomiaActividad | null>(null)
  const [lugaresRemotos, setLugaresRemotos] = useState<LugarCatalogo[] | null>(null)
  const [zonasRemotas, setZonasRemotas] = useState<{ codigo: string; nombre: string }[] | null>(null)
  const [clase, setClase] = useState("")
  const [subtipo, setSubtipo] = useState("")
  const [origen, setOrigen] = useState("")
  const [codigoExterno, setCodigoExterno] = useState("")
  const [unidad, setUnidad] = useState("")
  const [riesgo, setRiesgo] = useState("")
  const [fecha, setFecha] = useState("")
  const [cantidad, setCantidad] = useState("")
  const [personal, setPersonal] = useState<string[]>([])
  const [lugarId, setLugarId] = useState("")
  const [zona, setZona] = useState("")
  const [aviso, setAviso] = useState("")

  useEffect(() => {
    if (props.taxonomia) return
    let vivo = true
    void fetchTaxonomiaActividad()
      .then((dato) => {
        if (vivo) setRemota(dato)
      })
      .catch(() => {})
    return () => {
      vivo = false
    }
  }, [props.taxonomia])

  useEffect(() => {
    if (props.lugares) return
    let vivo = true
    void listarLugaresCatalogo()
      .then((dato) => {
        if (vivo) setLugaresRemotos(dato)
      })
      .catch(() => {})
    return () => {
      vivo = false
    }
  }, [props.lugares])

  useEffect(() => {
    if (props.zonas) return
    let vivo = true
    void listarZonas()
      .then((dato) => {
        if (vivo) setZonasRemotas(dato.map((item) => ({ codigo: item.codigo, nombre: item.nombre })))
      })
      .catch(() => {})
    return () => {
      vivo = false
    }
  }, [props.zonas])

  const tax = props.taxonomia ?? remota ?? TAX_VACIA
  const lugares = props.lugares ?? lugaresRemotos ?? []
  const zonas = props.zonas ?? zonasRemotas ?? []
  const tipos = tax.clases.find((item) => item.codigo === clase)?.tipos ?? []

  function elegirClase(codigo: string) {
    setClase(codigo)
    setSubtipo("")
    setAviso("")
  }

  function alternarPersonal(nombre: string) {
    setPersonal((actual) => (actual.includes(nombre) ? actual.filter((item) => item !== nombre) : [...actual, nombre]))
  }

  function confirmarPista() {
    const siguiente = claseDePista(props.pista?.codigo ?? "")
    if (siguiente) elegirClase(siguiente)
    props.onConfirmarPista?.()
  }

  function enviar(event: { preventDefault: () => void }) {
    event.preventDefault()
    if (!clase) {
      setAviso(ACTIVIDAD.faltaClase)
      return
    }
    if (tipos.length > 0 && !subtipo) {
      setAviso(ACTIVIDAD.faltaTipo)
      return
    }
    setAviso("")
    props.onCreate({
      tipo: tipoGrueso(clase, props.formTipo),
      clase,
      subtipo,
      origen,
      codigo_externo: codigoExterno.trim(),
      unidad_solicitante: unidad.trim(),
      nivel_riesgo: riesgo,
      fecha_programada: fecha,
      cantidad,
      personal,
      lugar_id: lugarId,
      zona_supervision_id: zona,
    })
  }

  return (
    <form className="form alta-actividad" onSubmit={enviar}>
      <header className="alta-punto">
        <span>{ACTIVIDAD.punto}</span>
        <strong>
          {props.draft.lat.toFixed(5)}, {props.draft.lon.toFixed(5)}
        </strong>
      </header>
      <fieldset>
        <legend>{ACTIVIDAD.clasificacion}</legend>
        <div className="alta-par">
          <label className="field">
            {ACTIVIDAD.clase}
            <select name="clase" value={clase} onChange={(event) => elegirClase(event.target.value)}>
              <option value="">{ACTIVIDAD.elegirClase}</option>
              {tax.clases.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {ACTIVIDAD.tipo}
            <select name="subtipo" value={subtipo} disabled={!clase || tipos.length === 0} onChange={(event) => setSubtipo(event.target.value)}>
              <option value="">{ACTIVIDAD.elegirTipo}</option>
              {tipos.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
        </div>
        {clase && tipos.length === 0 && <p className="hint">{ACTIVIDAD.sinTipos}</p>}
        <p className="hint">{ACTIVIDAD.regla}</p>
        {props.sugerencia && <p className="hint">{props.sugerencia}</p>}
        <div className="row-actions">
          <button type="button" onClick={props.onSugerir}>
            {ACTIVIDAD.sugerir}
          </button>
          {props.pista?.codigo && (
            <button type="button" onClick={confirmarPista}>
              Confirmar {props.pista.etiqueta || props.pista.codigo}
            </button>
          )}
        </div>
      </fieldset>
      <fieldset>
        <legend>{ACTIVIDAD.pedido}</legend>
        <div className="alta-par">
          <label className="field">
            {ACTIVIDAD.origen}
            <select name="origen" value={origen} onChange={(event) => setOrigen(event.target.value)}>
              <option value="">{ACTIVIDAD.elegirOrigen}</option>
              {tax.origenes.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {ACTIVIDAD.riesgo}
            <select name="nivel_riesgo" value={riesgo} onChange={(event) => setRiesgo(event.target.value)}>
              <option value="">{ACTIVIDAD.elegirRiesgo}</option>
              {tax.riesgos.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {ACTIVIDAD.codigoExterno}
            <input name="codigo_externo" value={codigoExterno} maxLength={80} onChange={(event) => setCodigoExterno(event.target.value)} />
            <small className="alta-nota">{ACTIVIDAD.codigoExternoHint}</small>
          </label>
          <label className="field">
            {ACTIVIDAD.unidad}
            <input name="unidad_solicitante" value={unidad} maxLength={160} onChange={(event) => setUnidad(event.target.value)} />
          </label>
          <label className="field">
            {ACTIVIDAD.fechaProgramada}
            <FechaCampo value={fecha} onChange={setFecha} />
          </label>
          <label className="field">
            {ACTIVIDAD.cantidad}
            <input name="cantidad" inputMode="decimal" value={cantidad} onChange={(event) => setCantidad(event.target.value)} />
          </label>
        </div>
      </fieldset>
      <fieldset className="alta-personal">
        <legend>{ACTIVIDAD.personal}</legend>
        <p className="hint">{ACTIVIDAD.personalLede}</p>
        <ul>
          {tax.personal.map((item) => (
            <li key={item.id}>
              <label>
                <input
                  type="checkbox"
                  name="personal"
                  value={item.nombre_ficticio}
                  checked={personal.includes(item.nombre_ficticio)}
                  onChange={() => alternarPersonal(item.nombre_ficticio)}
                />
                {item.nombre_ficticio}
              </label>
            </li>
          ))}
        </ul>
      </fieldset>
      <div className="alta-par">
        <SelectorLugar id="alta-lugar" lugares={lugares} lugarId={lugarId} onChange={setLugarId} />
        <label className="field">
          {ACTIVIDAD.zonaSupervision}
          <select name="zona_supervision_id" value={zona} onChange={(event) => setZona(event.target.value)}>
            <option value="">{ACTIVIDAD.sinZona}</option>
            {zonas.map((item) => (
              <option key={item.codigo} value={item.codigo}>
                {item.nombre}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {ACTIVIDAD.tituloCampo}
          <input value={props.formTitulo} maxLength={160} onChange={(event) => props.onForm({ titulo: event.target.value })} required />
        </label>
        <label className="field">
          {ACTIVIDAD.quienEjecuta}
          <select value={props.formEjecutor} onChange={(event) => props.onForm({ ejecutor: event.target.value })}>
            <option value="propia">{ACTIVIDAD.personalPropio}</option>
            <option value="tercerizada">{ACTIVIDAD.servicioTercerizado}</option>
          </select>
        </label>
      </div>
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
      {aviso && <p className="status error">{aviso}</p>}
      <button type="submit" className="primary" disabled={props.creating}>
        {props.creating ? ACTIVIDAD.guardando : ACTIVIDAD.crear}
      </button>
    </form>
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
