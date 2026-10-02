import { useEffect, useState } from "react"
import { ApiError } from "../operacion"
import { Esqueleto } from "../ui/Esqueleto"
import { crearOrden, crearSolicitud, fetchOrdenes, fetchSolicitudes, type Orden } from "../producto"
import { SOLICITUD } from "../ui/nomenclatura"

export function SolicitudesPanel(props: { actividadId: string }) {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof fetchSolicitudes>>>([])
  const [ordenes, setOrdenes] = useState<Orden[]>([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState("")
  const [titulo, setTitulo] = useState("")
  const [fuente, setFuente] = useState("osg")
  const [codigo, setCodigo] = useState("")
  const [prioridad, setPrioridad] = useState("media")
  const [lugar, setLugar] = useState("")
  const [empresa, setEmpresa] = useState("")
  const [referencia, setReferencia] = useState("")

  async function load() {
    try {
      const [sol, ord] = await Promise.all([fetchSolicitudes(), fetchOrdenes()])
      setRows(sol)
      setOrdenes(ord)
      setError("")
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "No se pudieron leer las solicitudes")
    }
  }

  useEffect(() => {
    let cancelled = false
    Promise.all([fetchSolicitudes(), fetchOrdenes()])
      .then(([sol, ord]) => {
        if (cancelled) return
        setRows(sol)
        setOrdenes(ord)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : "No se pudieron leer las solicitudes")
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
      <h2>Solicitudes</h2>
      <p className="lede">Captura manual. El código externo se conserva si viene de Centuria u OSG; el sistema no lo inventa.</p>
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && rows.length === 0 && !error && <p className="empty">No hay solicitudes registradas.</p>}
      <ul className="labor-list">
        {rows.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.titulo}</strong>
            <small>
              {row.fuente} · {row.estado} · {row.prioridad}
              {row.codigo_externo ? ` · ${row.codigo_externo}` : " · sin código externo"}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void crearSolicitud({
            id: crypto.randomUUID(),
            titulo,
            fuente,
            codigo_externo: codigo,
            prioridad,
            lugar,
            detalle: "",
          })
            .then(() => {
              setTitulo("")
              setCodigo("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo crear"))
        }}
      >
        <label className="field">
          Título
          <input value={titulo} onChange={(event) => setTitulo(event.target.value)} required />
        </label>
        <label className="field">
          {SOLICITUD.origen}
          <select value={fuente} onChange={(event) => setFuente(event.target.value)}>
            <option value="centuria">Centuria</option>
            <option value="osg">Matriz OSG</option>
            <option value="correo">Correo</option>
            <option value="interna">Interna</option>
          </select>
        </label>
        <label className="field">
          Código externo
          <input value={codigo} onChange={(event) => setCodigo(event.target.value)} placeholder="Opcional" />
        </label>
        <label className="field">
          Prioridad
          <select value={prioridad} onChange={(event) => setPrioridad(event.target.value)}>
            <option value="baja">Baja</option>
            <option value="media">Media</option>
            <option value="alta">Alta</option>
          </select>
        </label>
        <label className="field">
          Lugar
          <input value={lugar} onChange={(event) => setLugar(event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Registrar solicitud
        </button>
      </form>
      <h3>Órdenes de servicio</h3>
      <p className="lede">{SOLICITUD.vincula}</p>
      <p className="hint">
        {ordenes.length} {ordenes.length === 1 ? "orden registrada" : "órdenes registradas"}. {SOLICITUD.metricas}
      </p>
      {ordenes.length === 0 && <p className="empty">{SOLICITUD.vacioOrden}</p>}
      <ul className="labor-list">
        {ordenes.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.empresa}</strong>
            <small>
              {row.referencia} · {row.estado}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          if (!props.actividadId) {
            setError(SOLICITUD.elegir)
            return
          }
          void crearOrden({
            id: crypto.randomUUID(),
            actividad_id: props.actividadId,
            empresa,
            referencia,
            frecuencia: "",
          })
            .then(() => {
              setEmpresa("")
              setReferencia("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo crear la orden"))
        }}
      >
        <p className="hint">{props.actividadId ? SOLICITUD.seleccionada(props.actividadId.slice(0, 8)) : SOLICITUD.ninguna}</p>
        <label className="field">
          Empresa
          <input value={empresa} onChange={(event) => setEmpresa(event.target.value)} required />
        </label>
        <label className="field">
          Referencia de contratación
          <input value={referencia} onChange={(event) => setReferencia(event.target.value)} required />
        </label>
        <button type="submit">Registrar orden</button>
      </form>
    </section>
  )
}
