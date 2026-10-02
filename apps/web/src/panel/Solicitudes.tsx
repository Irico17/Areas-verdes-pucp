import { useEffect, useState } from "react"
import { ApiError } from "../operacion"
import { Esqueleto } from "../ui/Esqueleto"
import { SOLICITUD } from "../ui/nomenclatura"
import {
  crearOrden,
  crearSolicitud,
  fetchCatalogo,
  fetchOrdenes,
  fetchSolicitudes,
  type CatalogoItem,
  type Orden,
  type Solicitud,
} from "../producto"
import { SelectorLugar } from "./SelectorLugar"
import { listarLugaresCatalogo, type LugarCatalogo } from "./zonificacion"

export function SolicitudesPanel(props: { actividadId: string }) {
  const [rows, setRows] = useState<Solicitud[]>([])
  const [ordenes, setOrdenes] = useState<Orden[]>([])
  const [fuentes, setFuentes] = useState<CatalogoItem[]>([])
  const [estados, setEstados] = useState<CatalogoItem[]>([])
  const [empresas, setEmpresas] = useState<CatalogoItem[]>([])
  const [frecuencias, setFrecuencias] = useState<CatalogoItem[]>([])
  const [lugares, setLugares] = useState<LugarCatalogo[]>([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState("")
  const [titulo, setTitulo] = useState("")
  const [fuente, setFuente] = useState("osg")
  const [codigo, setCodigo] = useState("")
  const [prioridad, setPrioridad] = useState("media")
  const [estado, setEstado] = useState("por_iniciar")
  const [pedida, setPedida] = useState("")
  const [ejecutada, setEjecutada] = useState("")
  const [lugarId, setLugarId] = useState("")
  const [latitud, setLatitud] = useState("")
  const [longitud, setLongitud] = useState("")
  const [empresaId, setEmpresaId] = useState("")
  const [frecuenciaId, setFrecuenciaId] = useState("")
  const [referencia, setReferencia] = useState("")
  const [conformidad, setConformidad] = useState("")

  async function load() {
    const [sol, ord, fuenteItems, estadoItems, empresaItems, frecuenciaItems, lugarItems] = await Promise.all([
      fetchSolicitudes(),
      fetchOrdenes(),
      fetchCatalogo("fuente", true),
      fetchCatalogo("estado_solicitud", true),
      fetchCatalogo("empresa", true),
      fetchCatalogo("frecuencia", true),
      listarLugaresCatalogo().catch(() => [] as LugarCatalogo[]),
    ])
    setRows(sol)
    setOrdenes(ord)
    setFuentes(fuenteItems)
    setEstados(estadoItems)
    setEmpresas(empresaItems)
    setFrecuencias(frecuenciaItems)
    setLugares(lugarItems)
    setError("")
  }

  useEffect(() => {
    let cancelled = false
    Promise.all([
      fetchSolicitudes(),
      fetchOrdenes(),
      fetchCatalogo("fuente", true),
      fetchCatalogo("estado_solicitud", true),
      fetchCatalogo("empresa", true),
      fetchCatalogo("frecuencia", true),
      listarLugaresCatalogo().catch(() => [] as LugarCatalogo[]),
    ])
      .then(([sol, ord, fuenteItems, estadoItems, empresaItems, frecuenciaItems, lugarItems]) => {
        if (cancelled) return
        setRows(sol)
        setOrdenes(ord)
        setFuentes(fuenteItems)
        setEstados(estadoItems)
        setEmpresas(empresaItems)
        setFrecuencias(frecuenciaItems)
        setLugares(lugarItems)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : SOLICITUD.errorLectura)
      })
      .finally(() => {
        if (!cancelled) setCargando(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>{SOLICITUD.titulo}</h2>
      <p className="lede">{SOLICITUD.lede}</p>
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && rows.length === 0 && !error && <p className="empty">{SOLICITUD.vacio}</p>}
      <ul className="labor-list">
        {rows.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.titulo}</strong>
            <small>
              {etiqueta(fuentes, row.fuente)} · {etiqueta(estados, row.estado)} · {etiquetaPrioridad(row.prioridad)}
              {row.codigo_externo ? ` · ${row.codigo_externo}` : ""}
            </small>
            <small>
              {SOLICITUD.pedida}: {cantidadVisible(row.cantidad_solicitada ?? row.cantidad)} · {SOLICITUD.ejecutada}:{" "}
              {cantidadVisible(row.cantidad_ejecutada)}
            </small>
            {row.lugar_nombre && <small>{row.lugar_nombre}</small>}
            {row.lugar && (
              <small>
                {SOLICITUD.lugarCargado}: {row.lugar}
              </small>
            )}
            {row.lat != null && row.lon != null && (
              <small className="solicitud-punto">
                {row.lat.toFixed(5)}, {row.lon.toFixed(5)}
              </small>
            )}
          </li>
        ))}
      </ul>
      <form
        className="form solicitud-alta"
        onSubmit={(event) => {
          event.preventDefault()
          const pedidaN = entero(pedida)
          const ejecutadaN = entero(ejecutada)
          const lat = decimal(latitud)
          const lon = decimal(longitud)
          if ([pedidaN, ejecutadaN, lat, lon].some((n) => Number.isNaN(n))) {
            setError(SOLICITUD.numeroInvalido)
            return
          }
          if ((lat === undefined) !== (lon === undefined)) {
            setError(SOLICITUD.puntoJunto)
            return
          }
          void crearSolicitud({
            id: crypto.randomUUID(),
            titulo,
            fuente,
            codigo_externo: codigo,
            prioridad,
            estado,
            detalle: "",
            lugar_id: lugarId ? Number(lugarId) : undefined,
            lat,
            lon,
            cantidad_solicitada: pedidaN,
            cantidad_ejecutada: ejecutadaN,
          })
            .then(() => {
              setTitulo("")
              setCodigo("")
              setPedida("")
              setEjecutada("")
              setLugarId("")
              setLatitud("")
              setLongitud("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : SOLICITUD.errorCrear))
        }}
      >
        <label className="field">
          {SOLICITUD.tituloCampo}
          <input name="titulo" value={titulo} onChange={(event) => setTitulo(event.target.value)} required />
        </label>
        <div className="alta-par">
          <label className="field">
            {SOLICITUD.origen}
            <select name="fuente" value={fuente} onChange={(event) => setFuente(event.target.value)}>
              {fuentes.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {SOLICITUD.estado}
            <select name="estado" value={estado} onChange={(event) => setEstado(event.target.value)}>
              {estados.map((item) => (
                <option key={item.codigo} value={item.codigo}>
                  {item.nombre}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {SOLICITUD.prioridad}
            <select name="prioridad" value={prioridad} onChange={(event) => setPrioridad(event.target.value)}>
              <option value="baja">{SOLICITUD.prioridadBaja}</option>
              <option value="media">{SOLICITUD.prioridadMedia}</option>
              <option value="alta">{SOLICITUD.prioridadAlta}</option>
            </select>
          </label>
          <label className="field">
            {SOLICITUD.codigo}
            <input name="codigo_externo" value={codigo} onChange={(event) => setCodigo(event.target.value)} placeholder={SOLICITUD.codigoPlaceholder} />
          </label>
          <label className="field">
            {SOLICITUD.pedida}
            <input name="cantidad_solicitada" inputMode="numeric" value={pedida} onChange={(event) => setPedida(event.target.value)} />
          </label>
          <label className="field">
            {SOLICITUD.ejecutada}
            <input name="cantidad_ejecutada" inputMode="numeric" value={ejecutada} onChange={(event) => setEjecutada(event.target.value)} />
          </label>
        </div>
        <fieldset>
          <legend>{SOLICITUD.ubicacion}</legend>
          <p className="hint">{SOLICITUD.ubicacionHint}</p>
          <SelectorLugar id="solicitud-lugar" lugares={lugares} lugarId={lugarId} onChange={setLugarId} />
          <div className="alta-par">
            <label className="field">
              {SOLICITUD.latitud}
              <input name="lat" inputMode="decimal" value={latitud} onChange={(event) => setLatitud(event.target.value)} />
            </label>
            <label className="field">
              {SOLICITUD.longitud}
              <input name="lon" inputMode="decimal" value={longitud} onChange={(event) => setLongitud(event.target.value)} />
            </label>
          </div>
        </fieldset>
        <button type="submit" className="primary">
          {SOLICITUD.registrar}
        </button>
      </form>
      <h3>{SOLICITUD.ordenes}</h3>
      <p className="lede">{SOLICITUD.vincula}</p>
      <p className="hint">
        {ordenes.length} {ordenes.length === 1 ? SOLICITUD.unaOrden : SOLICITUD.variasOrdenes}. {SOLICITUD.metricas}
      </p>
      <p className="hint">{SOLICITUD.evidencias}</p>
      {ordenes.length === 0 && <p className="empty">{SOLICITUD.vacioOrden}</p>}
      <ul className="labor-list">
        {ordenes.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.empresa_catalogo || row.empresa}</strong>
            {row.empresa && row.empresa_catalogo && row.empresa !== row.empresa_catalogo && (
              <small>
                {SOLICITUD.empresaCargada}: {row.empresa}
              </small>
            )}
            <small>
              {row.referencia} · {row.estado}
              {row.frecuencia_catalogo || row.frecuencia ? ` · ${row.frecuencia_catalogo || row.frecuencia}` : ""}
            </small>
            {row.frecuencia && row.frecuencia_catalogo && row.frecuencia !== row.frecuencia_catalogo && (
              <small>
                {SOLICITUD.frecuenciaCargada}: {row.frecuencia}
              </small>
            )}
            {row.conformidad && (
              <small>
                {SOLICITUD.conformidad}: {row.conformidad}
              </small>
            )}
            <p className="hint">{SOLICITUD.evidencias}</p>
            {(row.evidencias ?? []).length === 0 ? (
              <p className="empty">{SOLICITUD.sinEvidencias}</p>
            ) : (
              <ul className="solicitud-evidencias">
                {(row.evidencias ?? []).map((ev) => (
                  <li key={ev.id}>
                    {ev.nombre}
                    <small>{ev.mime}</small>
                  </li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ul>
      <form
        className="form orden-alta"
        onSubmit={(event) => {
          event.preventDefault()
          if (!props.actividadId) {
            setError(SOLICITUD.elegir)
            return
          }
          void crearOrden({
            id: crypto.randomUUID(),
            actividad_id: props.actividadId,
            empresa_id: Number(empresaId),
            referencia,
            frecuencia_id: Number(frecuenciaId),
            conformidad,
          })
            .then(() => {
              setEmpresaId("")
              setFrecuenciaId("")
              setReferencia("")
              setConformidad("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : SOLICITUD.errorOrden))
        }}
      >
        <p className="hint">{props.actividadId ? SOLICITUD.seleccionada(props.actividadId.slice(0, 8)) : SOLICITUD.ninguna}</p>
        <label className="field">
          {SOLICITUD.empresa}
          <select name="empresa_id" value={empresaId} onChange={(event) => setEmpresaId(event.target.value)} required>
            <option value="">{SOLICITUD.elegirEmpresa}</option>
            {empresas.map((item) => (
              <option key={item.id} value={String(item.id)}>
                {item.nombre}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {SOLICITUD.frecuencia}
          <select name="frecuencia_id" value={frecuenciaId} onChange={(event) => setFrecuenciaId(event.target.value)} required>
            <option value="">{SOLICITUD.elegirFrecuencia}</option>
            {frecuencias.map((item) => (
              <option key={item.id} value={String(item.id)}>
                {item.nombre}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {SOLICITUD.referencia}
          <input name="referencia" value={referencia} onChange={(event) => setReferencia(event.target.value)} required />
        </label>
        <label className="field">
          {SOLICITUD.conformidad}
          <textarea name="conformidad" value={conformidad} onChange={(event) => setConformidad(event.target.value)} />
          <small className="alta-nota">{SOLICITUD.conformidadHint}</small>
        </label>
        <button type="submit">{SOLICITUD.registrarOrden}</button>
      </form>
    </section>
  )
}

function etiqueta(items: CatalogoItem[], codigo: string) {
  return items.find((item) => item.codigo === codigo)?.nombre ?? codigo
}

function etiquetaPrioridad(codigo: string) {
  if (codigo === "baja") return SOLICITUD.prioridadBaja
  if (codigo === "alta") return SOLICITUD.prioridadAlta
  if (codigo === "media") return SOLICITUD.prioridadMedia
  return codigo
}

function cantidadVisible(n: number | undefined) {
  return n == null ? SOLICITUD.sinCantidad : String(n)
}

function entero(valor: string): number | undefined {
  const limpio = valor.trim()
  if (!limpio) return undefined
  if (!/^\d+$/.test(limpio)) return Number.NaN
  return Number(limpio)
}

function decimal(valor: string): number | undefined {
  const limpio = valor.trim()
  if (!limpio) return undefined
  if (!/^-?\d+(\.\d+)?$/.test(limpio)) return Number.NaN
  return Number(limpio)
}
