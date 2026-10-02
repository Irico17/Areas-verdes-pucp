import { useEffect, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { ESTADOS, etiquetaEstado, etiquetaTipo } from "../operacion"
import { fetchReporte, HUECOS, reporteHref } from "../producto"

export function ReportesPanel() {
  const [estado, setEstado] = useState("")
  const [desde, setDesde] = useState("")
  const [hasta, setHasta] = useState("")
  const [zona, setZona] = useState("")
  const [cuadrilla, setCuadrilla] = useState("")
  const [origen, setOrigen] = useState("")
  const [data, setData] = useState<Awaited<ReturnType<typeof fetchReporte>> | null>(null)
  const [error, setError] = useState("")
  const filtro = { estado, desde, hasta, zona, cuadrilla, origen }

  async function load() {
    try {
      setData(await fetchReporte(filtro))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo armar el reporte")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchReporte({})
      .then((report) => {
        if (cancelled) return
        setData(report)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo armar el reporte")
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>Reportes</h2>
      <p className="lede">Reporte básico de labores, filtrable por zona, cuadrilla, origen y fechas. Los conteos no son indicadores oficiales.</p>
      <h3>Sin fórmula acordada</h3>
      <ul className="huecos">
        {HUECOS.map((hueco) => (
          <li key={hueco.clave}>
            <div>
              <span>{hueco.nombre}</span>
              <p>{hueco.nota}</p>
            </div>
            <strong>{hueco.estado}</strong>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          if (event.currentTarget.querySelector('[aria-invalid="true"]')) {
            setError("Corrija la fecha antes de actualizar.")
            return
          }
          void load()
        }}
      >
        <label className="field">
          Estado
          <select value={estado} onChange={(event) => setEstado(event.target.value)}>
            <option value="">Todos</option>
            {ESTADOS.map((item) => (
              <option key={item.id} value={item.id}>
                {item.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Zona
          <input value={zona} onChange={(event) => setZona(event.target.value)} placeholder="Z1" />
        </label>
        <label className="field">
          Cuadrilla
          <input value={cuadrilla} onChange={(event) => setCuadrilla(event.target.value)} placeholder="Nombre ficticio" />
        </label>
        <label className="field">
          Origen
          <input value={origen} onChange={(event) => setOrigen(event.target.value)} placeholder="monitoreo" />
        </label>
        <label className="field">
          Desde
          <FechaCampo value={desde} onChange={setDesde} />
        </label>
        <label className="field">
          Hasta
          <FechaCampo value={hasta} onChange={setHasta} />
        </label>
        <button type="submit" className="primary">
          Actualizar
        </button>
      </form>
      {error && <p className="status error">{error}</p>}
      {data && (
        <>
          <h3>Conteos operativos</h3>
          <p className="hint">{data.aviso}</p>
          <ul className="counts">
            {data.por_estado.map((row) => (
              <li key={row.estado}>
                <strong>{row.n}</strong>
                <span>{etiquetaEstado(row.estado)}</span>
              </li>
            ))}
          </ul>
          <p className="row-actions">
            <a href={reporteHref("csv", filtro)}>Descargar CSV</a>
            <a href={reporteHref("xls", filtro)}>Descargar Excel</a>
          </p>
          {data.filas.length === 0 && <p className="empty">No hay labores en ese rango.</p>}
          <ul className="labor-list">
            {data.filas.slice(0, 40).map((row) => (
              <li key={row.id} className="agenda">
                <strong>{row.titulo}</strong>
                <small>
                  {row.clase || etiquetaTipo(row.tipo)} · {etiquetaEstado(row.estado)}
                  {row.lugar ? ` · ${row.lugar}` : ""}
                  {row.cuadrilla ? ` · ${row.cuadrilla}` : ""}
                  {row.fecha_solicitud ? ` · ${row.fecha_solicitud}` : ""}
                </small>
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  )
}
