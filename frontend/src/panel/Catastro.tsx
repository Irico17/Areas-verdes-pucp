import { useEffect, useState } from "react"
import { crearAreaSinGeom, fetchFichas, guardarFicha, type Ficha } from "../producto"

function tituloFicha(row: Ficha): string {
  const nombre = row.nombre.trim()
  return nombre || row.feature_id
}

export function CatastroPanel() {
  const [q, setQ] = useState("")
  const [rows, setRows] = useState<Ficha[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [sel, setSel] = useState<Ficha | null>(null)
  const [nombreNuevo, setNombreNuevo] = useState("")
  const [usoNuevo, setUsoNuevo] = useState("")
  const [aviso, setAviso] = useState("")

  async function load(query = q) {
    setLoading(true)
    try {
      setRows(await fetchFichas(query))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchFichas("")
      .then((rows) => {
        if (cancelled) return
        setRows(rows)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="split">
      <div className="split-list">
        <h2>Catastro</h2>
        <p className="lede">Las áreas con nombre van primero. Si el catastro no trae nombre, se muestra el código.</p>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            void load(q)
          }}
        >
          <label className="field">
            Buscar
            <input value={q} onChange={(event) => setQ(event.target.value)} placeholder="Nombre, código o uso" />
          </label>
          <button type="submit">Buscar</button>
        </form>
        {error && <p className="status error">{error}</p>}
        {loading && (
          <div className="skel-wrap" aria-hidden="true">
            <div className="skel" />
            <div className="skel" />
            <div className="skel" />
          </div>
        )}
        {!loading && rows.length === 0 && !error && <p className="empty">Ningún área coincide con esa búsqueda.</p>}
        <ul className="labor-list">
          {rows.map((row) => (
            <li key={row.feature_id}>
              <button
                type="button"
                className={sel?.feature_id === row.feature_id ? "labor on" : "labor"}
                onClick={() => setSel(row)}
              >
                <span>
                  <strong>{tituloFicha(row)}</strong>
                  <small>
                    {row.nombre.trim() ? row.feature_id : "Sin nombre en el catastro"}
                    {row.uso ? ` · ${row.uso}` : ""}
                    {row.con_geometria ? "" : " · sin geometría"}
                  </small>
                </span>
              </button>
            </li>
          ))}
        </ul>
      </div>
      <div className="split-detail">
        {!sel && <p className="empty">Elija un área. La ficha queda en este panel, sin bajar por la lista.</p>}
        {sel && (
          <form
            className="form"
            onSubmit={(event) => {
              event.preventDefault()
              void guardarFicha(sel)
                .then(() => {
                  setAviso("Ficha guardada.")
                  setRows((current) => current.map((row) => (row.feature_id === sel.feature_id ? sel : row)))
                })
                .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo guardar"))
            }}
          >
            <h3>{tituloFicha(sel)}</h3>
            <label className="field">
              Nombre
              <input value={sel.nombre} onChange={(event) => setSel({ ...sel, nombre: event.target.value })} />
            </label>
            <label className="field">
              Uso
              <input value={sel.uso} onChange={(event) => setSel({ ...sel, uso: event.target.value })} />
            </label>
            <label className="field">
              Riego actual
              <input value={sel.riego_act} onChange={(event) => setSel({ ...sel, riego_act: event.target.value })} />
            </label>
            <label className="field">
              Referencia
              <input value={sel.referencia} onChange={(event) => setSel({ ...sel, referencia: event.target.value })} />
            </label>
            <button type="submit" className="primary">
              Guardar ficha
            </button>
          </form>
        )}
        <h3>Área sin GPS</h3>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            void crearAreaSinGeom(nombreNuevo, usoNuevo)
              .then(() => {
                setNombreNuevo("")
                setUsoNuevo("")
                setAviso("Área creada sin geometría.")
                return load(q)
              })
              .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo crear"))
          }}
        >
          <label className="field">
            Nombre
            <input value={nombreNuevo} onChange={(event) => setNombreNuevo(event.target.value)} required />
          </label>
          <label className="field">
            Uso
            <input value={usoNuevo} onChange={(event) => setUsoNuevo(event.target.value)} />
          </label>
          <button type="submit">Registrar sin geometría</button>
        </form>
        {aviso && <p className={aviso.toLowerCase().includes("no se") ? "status error" : "banner"}>{aviso}</p>}
      </div>
    </section>
  )
}
