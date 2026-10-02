import { useEffect, useRef, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { hoyISO } from "../fecha"
import { mostrarEnPanel } from "../ui/desplazar"
import { Esqueleto } from "../ui/Esqueleto"
import {
  ApiError,
  estadosPermitidos,
  fetchTaxonomiaActividad,
  tipoGrueso,
  etiquetaEstado,
  etiquetaTipo,
  type AltaCampos,
  type Capataz,
  type Evento,
  type FiltroActividades,
  type TaxonomiaActividad,
} from "../operacion"
import { send, type CatalogoItem } from "../producto"
import type { Rol } from "../types"
import { IconoClase } from "../map/iconoClase"
import { encolarRegistro, listEstados, ofreceColaDeAltas } from "../offline/queue"
import { COLA_VACIADA, publicarOEncolar, rutaAvance, rutaFicha, vaciarRegistros, type DetalleCola } from "../offline/registros"
import { ACTIVIDAD, COLA, EVIDENCIA, FALLO, VISTA_ACTIVIDAD, etiquetaCuadrilla, marcaEjecutor } from "../ui/nomenclatura"
import { Bitacora } from "./Bitacora"
import { FiltrosActividad } from "./FiltrosActividad"
import { EvidenciasCampo } from "./EvidenciasCampo"
import { SelectorLugar } from "./SelectorLugar"
import { listarZonas } from "./catastro"
import { listarLugaresCatalogo, type LugarCatalogo } from "./zonificacion"
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
  onArchivar: (motivo: string) => void
  conteos?: Record<string, number>
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
  const puedeValidar = props.rol !== "capataz"
  const puedeRegistrar = props.rol !== "jefatura"
  const [menuAbierto, setMenuAbierto] = useState(false)
  const [dialogoArchivo, setDialogoArchivo] = useState(false)
  const [motivoDialogo, setMotivoDialogo] = useState("")
  const dialogoRef = useRef<HTMLDialogElement>(null)
  const estadosPendientes = usePendientesEstados(props.notice)
  const totalCola = (ofreceColaDeAltas(props.rol) ? props.queueCount : 0) + estadosPendientes
  const avisoError = /no se|sin conexión|error/i.test(props.notice)
  const detalleRef = useRef<HTMLDivElement>(null)
  const elegida = props.selected?.id
  const [vistaId, setVistaId] = useState(elegida)
  if (elegida !== vistaId) {
    setVistaId(elegida)
    setMenuAbierto(false)
    setDialogoArchivo(false)
    setMotivoDialogo("")
  }
  useEffect(() => {
    if (elegida) mostrarEnPanel(detalleRef.current, "inicio")
  }, [elegida])
  useEffect(() => {
    const el = dialogoRef.current
    if (!el) return
    if (dialogoArchivo && !el.open) el.showModal()
    if (!dialogoArchivo && el.open) el.close()
  }, [dialogoArchivo])
  const detalle = props.selected ? (
        <div className="detail" id="detalle-actividad" ref={detalleRef}>
          <h3>{props.selected.titulo}</h3>
          <p className="meta">
            {etiquetaTipo(props.selected.tipo)} · {etiquetaEstado(props.selected.estado)} · {props.selected.equipo || ACTIVIDAD.sinCuadrilla}
            {` · ${marcaEjecutor(props.selected.ejecutor) || ACTIVIDAD.propioMarca}`}
          </p>
          {props.selected.detalle && <p className="lede">{props.selected.detalle}</p>}
          {props.selected.queued ? (
            <p className="hint">Aún no está en el servidor. El id ya quedó reservado para el reintento.</p>
          ) : (
            <>
              <div className="detalle-acciones">
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
                <label className="btn-foto" htmlFor="foto-actividad">
                  {ACTIVIDAD.foto}
                </label>
                {puedeValidar && (
                  <div className="menu-detalle">
                    <button
                      type="button"
                      aria-expanded={menuAbierto}
                      aria-haspopup="menu"
                      aria-label={ACTIVIDAD.masAcciones}
                      onClick={() => setMenuAbierto((abierto) => !abierto)}
                    >
                      ⋯
                    </button>
                    {menuAbierto && (
                      <div className="menu-panel" role="menu">
                        <label className="field">
                          {ACTIVIDAD.reasignarA}
                          <select value={props.reasignarA} onChange={(event) => props.onReasignarA(event.target.value)}>
                            {props.equipos.map((equipo) => (
                              <option key={equipo.id} value={equipo.id}>
                                {etiquetaCuadrilla(equipo.id, equipo.equipo)}
                              </option>
                            ))}
                          </select>
                        </label>
                        <button
                          type="button"
                          role="menuitem"
                          onClick={() => {
                            setMenuAbierto(false)
                            props.onReasignar()
                          }}
                        >
                          {ACTIVIDAD.reasignar}
                        </button>
                        <button
                          type="button"
                          role="menuitem"
                          className="danger"
                          onClick={() => {
                            setMenuAbierto(false)
                            setMotivoDialogo("")
                            setDialogoArchivo(true)
                          }}
                        >
                          {ACTIVIDAD.archivar}
                        </button>
                      </div>
                    )}
                  </div>
                )}
              </div>
              {puedeValidar && (
                <dialog
                  ref={dialogoRef}
                  className="dialogo-archivo"
                  aria-labelledby="archivo-pregunta"
                  onClose={() => setDialogoArchivo(false)}
                >
                  <h3 id="archivo-pregunta">{ACTIVIDAD.archivarPregunta(props.selected.titulo)}</h3>
                  <label className="field">
                    {ACTIVIDAD.motivo}
                    <select value={motivoDialogo} onChange={(event) => setMotivoDialogo(event.target.value)}>
                      <option value="">{ACTIVIDAD.elegir}</option>
                      {props.motivos.map((item) => (
                        <option key={item.codigo} value={item.codigo}>
                          {item.nombre}
                        </option>
                      ))}
                    </select>
                  </label>
                  {!motivoDialogo && (
                    <p className="hint" id="archivo-falta">
                      {ACTIVIDAD.faltaMotivo}
                    </p>
                  )}
                  <div className="row-actions">
                    <button type="button" onClick={() => setDialogoArchivo(false)}>
                      {ACTIVIDAD.cancelar}
                    </button>
                    <button
                      type="button"
                      className="danger"
                      disabled={!motivoDialogo}
                      onClick={() => {
                        const motivo = motivoDialogo
                        setDialogoArchivo(false)
                        props.onArchivar(motivo)
                      }}
                    >
                      {ACTIVIDAD.archivar}
                    </button>
                  </div>
                </dialog>
              )}
              <details className="pliegue">
                <summary>{ACTIVIDAD.ficha}</summary>
                <FichaLabor key={props.selected.id} actividadId={props.selected.id} puedeRegistrar={puedeRegistrar} />
              </details>
              {puedeRegistrar && (
                <details className="pliegue">
                  <summary>{ACTIVIDAD.avance}</summary>
                  <AvanceCampo key={`avance-${props.selected.id}`} actividadId={props.selected.id} />
                </details>
              )}
              <details className="pliegue">
                <summary>{EVIDENCIA.titulo}</summary>
                <EvidenciasCampo sinTitulo actividadId={props.selected.id} />
              </details>
              <details className="pliegue">
                <summary>{ACTIVIDAD.bitacora}</summary>
                <Bitacora embebida sinTitulo actividadId={props.selected.id} eventos={props.timeline} error={props.timelineError} />
              </details>
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
      <FiltrosActividad rol={props.rol} valor={props.filtro} onChange={props.onFiltro} tipos={props.tipos} conteos={props.conteos} />
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
      {totalCola > 0 && (
        <p className="hint" role="status">
          {ACTIVIDAD.colaLocal(totalCola)} <span className="marca-cola">{COLA.marca}</span>{" "}
          <button
            type="button"
            className="link"
            onClick={() => {
              props.onFlush()
              void vaciarRegistros()
            }}
          >
            {ACTIVIDAD.reintentar}
          </button>
        </p>
      )}
      {props.notice && !(totalCola === 0 && props.notice.includes("Sin señal")) && (
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
        <legend>{ACTIVIDAD.que}</legend>
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
          <label className="field">
            {ACTIVIDAD.tituloCampo}
            <input value={props.formTitulo} maxLength={160} onChange={(event) => props.onForm({ titulo: event.target.value })} required />
          </label>
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
        <legend>{ACTIVIDAD.quien}</legend>
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
          <label className="field">
            {ACTIVIDAD.cuadrilla}
            <select value={props.formEquipo} onChange={(event) => props.onForm({ equipo: event.target.value })}>
              <option value="">{ACTIVIDAD.sinAsignar}</option>
              {props.equipos.map((equipo) => (
                <option key={equipo.id} value={equipo.id}>
                  {etiquetaCuadrilla(equipo.id, equipo.equipo)}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {ACTIVIDAD.quienEjecuta}
            <select value={props.formEjecutor} onChange={(event) => props.onForm({ ejecutor: event.target.value })}>
              <option value="propia">{ACTIVIDAD.personalPropio}</option>
              <option value="tercerizada">{ACTIVIDAD.servicioTercerizado}</option>
            </select>
          </label>
        </div>
      </fieldset>
      <fieldset>
        <legend>{ACTIVIDAD.detalle}</legend>
        <label className="field">
          <span className="sr-only">{ACTIVIDAD.detalle}</span>
          <textarea value={props.formDetalle} maxLength={2000} rows={3} onChange={(event) => props.onForm({ detalle: event.target.value })} />
        </label>
      </fieldset>
      {aviso && <p className="status error">{aviso}</p>}
      <button type="submit" className="primary" disabled={props.creating}>
        {props.creating ? ACTIVIDAD.guardando : ACTIVIDAD.crear}
      </button>
    </form>
  )
}

function usePendientesEstados(aviso: string): number {
  const [n, setN] = useState(0)
  useEffect(() => {
    let vivo = true
    const leer = () => {
      void listEstados()
        .then((filas) => {
          if (vivo) setN(filas.length)
        })
        .catch(() => {
          if (vivo) setN(0)
        })
    }
    leer()
    window.addEventListener(COLA_VACIADA, leer)
    return () => {
      vivo = false
      window.removeEventListener(COLA_VACIADA, leer)
    }
  }, [aviso])
  return n
}

function FichaLabor(props: { actividadId?: string; puedeRegistrar?: boolean }) {
  const editable = props.puedeRegistrar !== false
  const [clase, setClase] = useState("Mantenimiento de jardines")
  const [solicitud, setSolicitud] = useState("")
  const [atencion, setAtencion] = useState("")
  const [lugar, setLugar] = useState("")
  const [comentario, setComentario] = useState("")
  const [aviso, setAviso] = useState("")
  const fichaRef = useRef<HTMLFieldSetElement>(null)
  useEffect(() => {
    const alVolver = (ev: Event) => {
      const detail = (ev as CustomEvent<DetalleCola>).detail
      if (detail?.conflictos?.some((item) => item.tipo === "ficha")) {
        setAviso(COLA.conflicto)
        return
      }
      setAviso((actual) => (actual === COLA.ficha ? "" : actual))
    }
    window.addEventListener(COLA_VACIADA, alVolver)
    return () => window.removeEventListener(COLA_VACIADA, alVolver)
  }, [])
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
    const cuerpo = {
      clase,
      fecha_solicitud: solicitud,
      fecha_atencion: atencion,
      lugar,
      comentario,
    }
    const ruta = rutaFicha(props.actividadId)
    try {
      await send(ruta, "PATCH", cuerpo)
      setAviso(lugar || COLA.fichaGuardada)
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) {
        setAviso(COLA.conflicto)
        return
      }
      if (error instanceof ApiError && error.status === 403) {
        setAviso(FALLO.ficha)
        return
      }
      if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
        setAviso(error.message)
        return
      }
      if (error instanceof ApiError && error.status === 0) {
        await encolarRegistro({
          id: crypto.randomUUID(),
          tipo: "ficha",
          path: ruta,
          method: "PATCH",
          body: cuerpo,
          createdAt: new Date().toISOString(),
        })
        setAviso(COLA.ficha)
        return
      }
      setAviso(COLA.fichaError)
    }
  }
  return (
    <fieldset className="form grupo" ref={fichaRef}>
      <legend>{ACTIVIDAD.ficha}</legend>
      <label className="field">
        {ACTIVIDAD.clase}
        <input value={clase} readOnly={!editable} onChange={(event) => setClase(event.target.value)} />
      </label>
      <label className="field">
        {ACTIVIDAD.fechaSolicitud}
        <FechaCampo value={solicitud} onChange={setSolicitud} soloLectura={!editable} />
      </label>
      <label className="field">
        {ACTIVIDAD.fechaAtencion}
        <FechaCampo value={atencion} onChange={setAtencion} soloLectura={!editable} />
      </label>
      <label className="field">
        {ACTIVIDAD.lugar}
        <input value={lugar} readOnly={!editable} onChange={(event) => setLugar(event.target.value)} placeholder={ACTIVIDAD.lugarPlaceholder} />
      </label>
      <label className="field">
        {ACTIVIDAD.comentario}
        <textarea value={comentario} readOnly={!editable} rows={2} maxLength={2000} onChange={(event) => setComentario(event.target.value)} />
      </label>
      {editable && (
        <button type="button" onClick={() => void guardar()}>
          {ACTIVIDAD.guardarFicha}
        </button>
      )}
      {aviso && <p className="hint">{aviso}</p>}
    </fieldset>
  )
}

function AvanceCampo(props: { actividadId: string }) {
  const [fecha, setFecha] = useState(hoyISO)
  const [nota, setNota] = useState("")
  const [area, setArea] = useState("")
  const [ejemplar, setEjemplar] = useState("")
  const [aviso, setAviso] = useState("")
  const ref = useRef<HTMLFieldSetElement>(null)
  useEffect(() => {
    const alVolver = (ev: Event) => {
      const detail = (ev as CustomEvent<DetalleCola>).detail
      if (detail?.conflictos?.some((item) => item.tipo === "avance")) {
        setAviso(COLA.conflicto)
        return
      }
      setAviso((actual) => (actual === COLA.avance ? "" : actual))
    }
    window.addEventListener(COLA_VACIADA, alVolver)
    return () => window.removeEventListener(COLA_VACIADA, alVolver)
  }, [])
  async function guardar() {
    if (ref.current?.querySelector('[aria-invalid="true"]')) {
      setAviso("Corrija la fecha antes de guardar.")
      return
    }
    if (!area.trim() && !ejemplar.trim()) {
      setAviso(ACTIVIDAD.faltaLugarAvance)
      return
    }
    const id = crypto.randomUUID()
    const resultado = await publicarOEncolar({
      id,
      tipo: "avance",
      path: rutaAvance(props.actividadId),
      method: "POST",
      body: {
        id,
        fecha,
        nota: nota.trim(),
        area_feature_id: area.trim(),
        ejemplar_ref: ejemplar.trim(),
      },
    })
    if (resultado === "ok") {
      setNota("")
      setAviso(ACTIVIDAD.avanceGuardado)
      return
    }
    if (resultado === "conflicto") {
      setAviso(COLA.conflicto)
      return
    }
    if (resultado === "encolado") {
      setAviso(COLA.avance)
      return
    }
    setAviso(ACTIVIDAD.avanceError)
  }
  return (
    <fieldset className="form grupo" ref={ref}>
      <legend>{ACTIVIDAD.avance}</legend>
      <p className="hint">{ACTIVIDAD.avanceLede}</p>
      <label className="field">
        {ACTIVIDAD.fechaAvance}
        <FechaCampo value={fecha} onChange={setFecha} required />
      </label>
      <label className="field">
        {ACTIVIDAD.comentario}
        <textarea value={nota} rows={2} maxLength={2000} onChange={(event) => setNota(event.target.value)} />
      </label>
      <label className="field">
        {ACTIVIDAD.area}
        <input value={area} onChange={(event) => setArea(event.target.value)} placeholder={ACTIVIDAD.areaPlaceholder} />
      </label>
      <label className="field">
        {ACTIVIDAD.ejemplar}
        <input value={ejemplar} onChange={(event) => setEjemplar(event.target.value)} placeholder={ACTIVIDAD.ejemplarPlaceholder} />
      </label>
      <button type="button" onClick={() => void guardar()}>
        {ACTIVIDAD.guardarAvance}
      </button>
      {aviso && (
        <p className="hint">
          {aviso}
          {aviso === COLA.avance && <span className="marca-cola">{COLA.marca}</span>}
        </p>
      )}
    </fieldset>
  )
}
