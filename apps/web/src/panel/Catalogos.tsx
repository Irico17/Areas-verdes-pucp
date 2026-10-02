import { useEffect, useState } from "react"
import { Esqueleto } from "../ui/Esqueleto"
import { crearCatalogo, desactivarCatalogo, fetchCatalogo, type CatalogoItem } from "../producto"

const CLASES = [
  ["tipo_actividad", "Tipos de labor"],
  ["estado", "Estados"],
  ["prioridad", "Prioridades"],
  ["lugar", "Lugares"],
  ["especie", "Especies"],
  ["motivo_archivo", "Motivos de archivo"],
  ["turno", "Turnos"],
  ["fuente", "Fuentes"],
] as const

export function CatalogosPanel(props: { editable: boolean }) {
  const [clase, setClase] = useState<string>("tipo_actividad")
  const [items, setItems] = useState<CatalogoItem[]>([])
  const [cargadoDe, setCargadoDe] = useState<string | null>(null)
  const cargando = cargadoDe !== clase
  const [error, setError] = useState("")
  const [codigo, setCodigo] = useState("")
  const [nombre, setNombre] = useState("")

  async function load(next = clase) {
    try {
      setItems(await fetchCatalogo(next, false))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el catálogo")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchCatalogo(clase, false)
      .then((next) => {
        if (cancelled) return
        setItems(next)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el catálogo")
      })
      .finally(() => {
        if (!cancelled) setCargadoDe(clase)
      })
    return () => {
      cancelled = true
    }
  }, [clase])

  return (
    <section className="block">
      <h2>Catálogos</h2>
      <p className="lede">Valores que usa la operación. Desactivar no borra el historial.</p>
      <label className="field">
        Clase
        <select value={clase} onChange={(event) => setClase(event.target.value)}>
          {CLASES.map(([id, label]) => (
            <option key={id} value={id}>
              {label}
            </option>
          ))}
        </select>
      </label>
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && items.length === 0 && !error && <p className="empty">Esta clase no tiene ítems.</p>}
      <ul className="labor-list">
        {items.map((item) => (
          <li key={item.id} className="agenda">
            <strong>
              {item.nombre}
              {item.activo ? "" : " · inactivo"}
            </strong>
            <small>{item.codigo}</small>
            {props.editable && item.activo && (
              <button type="button" className="link" onClick={() => void desactivarCatalogo(item.id).then(() => load())}>
                Desactivar
              </button>
            )}
          </li>
        ))}
      </ul>
      {props.editable && (
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          void crearCatalogo(clase, codigo, nombre)
            .then(() => {
              setCodigo("")
              setNombre("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo guardar"))
        }}
      >
        <label className="field">
          Código
          <input value={codigo} onChange={(event) => setCodigo(event.target.value)} placeholder="minusculas_sin_espacio" required />
        </label>
        <label className="field">
          Nombre
          <input value={nombre} onChange={(event) => setNombre(event.target.value)} required />
        </label>
        <button type="submit" className="primary">
          Agregar
        </button>
      </form>
      )}
      {!props.editable && <p className="hint">Puede consultar el catálogo. Solo administración lo modifica.</p>}
    </section>
  )
}
