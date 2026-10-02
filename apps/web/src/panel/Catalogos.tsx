import { useEffect, useState } from "react"
import { aplicarEstadosCatalogo } from "../operacion"
import { Esqueleto } from "../ui/Esqueleto"
import {
  crearCatalogo,
  desactivarCatalogo,
  fetchCatalogoCompleto,
  renombrarCatalogo,
  type CatalogoItem,
} from "../producto"

const ETIQUETA_CLASE: Record<string, string> = {
  tipo_actividad: "Tipos de actividad",
  estado: "Estados",
  prioridad: "Prioridades",
  lugar: "Lugares",
  especie: "Especies",
  motivo_archivo: "Motivos de archivo",
  turno: "Turnos",
  fuente: "Fuentes",
  clase_actividad: "Clases de actividad",
  plaga: "Plagas",
  producto_fitosanitario: "Productos fitosanitarios",
  frecuencia: "Frecuencias",
  sede: "Sedes",
  cuartel: "Cuarteles (histórico)",
  sector_capataz: "Sectores de capataz",
}

const CLASES_INICIALES = Object.keys(ETIQUETA_CLASE)

function etiquetaClase(id: string) {
  return ETIQUETA_CLASE[id] ?? id.replaceAll("_", " ")
}

export function ListaCatalogo(props: {
  items: CatalogoItem[]
  editable: boolean
  onRenombrar: (id: number, nombre: string) => Promise<void>
  onDesactivar: (id: number) => Promise<void>
}) {
  const [editando, setEditando] = useState<number | null>(null)
  const [nombre, setNombre] = useState("")

  return (
    <ul className="labor-list">
      {props.items.map((item) => (
        <li key={item.id} className="agenda">
          <strong>
            {item.nombre}
            {item.activo ? "" : " · inactivo"}
            {item.provisional ? " · provisional" : ""}
          </strong>
          <small>{item.codigo}</small>
          {props.editable && editando === item.id && (
            <form
              className="form"
              onSubmit={(event) => {
                event.preventDefault()
                void props.onRenombrar(item.id, nombre)
                  .then(() => setEditando(null))
                  .catch(() => {})
              }}
            >
              <label className="field">
                Nombre
                <input value={nombre} onChange={(event) => setNombre(event.target.value)} required />
              </label>
              <button type="submit" className="primary">
                Guardar nombre
              </button>
            </form>
          )}
          {props.editable && (editando !== item.id || item.activo) && (
            <div className="agenda-acciones">
              {editando !== item.id && (
                <button
                  type="button"
                  className="link"
                  onClick={() => {
                    setEditando(item.id)
                    setNombre(item.nombre)
                  }}
                >
                  Corregir nombre
                </button>
              )}
              {item.activo && (
                <button type="button" className="link" onClick={() => void props.onDesactivar(item.id)}>
                  Desactivar
                </button>
              )}
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}

export function CatalogosPanel(props: { editable: boolean }) {
  const [clase, setClase] = useState<string>("tipo_actividad")
  const [clases, setClases] = useState<string[]>(CLASES_INICIALES)
  const [items, setItems] = useState<CatalogoItem[]>([])
  const [cargadoDe, setCargadoDe] = useState<string | null>(null)
  const cargando = cargadoDe !== clase
  const [error, setError] = useState("")
  const [codigo, setCodigo] = useState("")
  const [nombre, setNombre] = useState("")

  async function load(next = clase) {
    try {
      const body = await fetchCatalogoCompleto(next, false)
      setItems(body.items)
      if (body.clases.length > 0) setClases(body.clases)
      if (next === "estado") aplicarEstadosCatalogo(body.items.map(comoEstado))
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el catálogo")
    }
  }

  useEffect(() => {
    let cancelled = false
    fetchCatalogoCompleto(clase, false)
      .then((body) => {
        if (cancelled) return
        setItems(body.items)
        if (body.clases.length > 0) setClases(body.clases)
        if (clase === "estado") aplicarEstadosCatalogo(body.items.map(comoEstado))
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
      <p className="lede">
        Valores que usa la operación. Desactivar no borra. El nombre se puede corregir y queda en el historial. «Bloqueada» sigue
        guardada, inactiva y provisional, hasta que el cliente decida.
      </p>
      <label className="field">
        Clase
        <select value={clase} onChange={(event) => setClase(event.target.value)}>
          {clases.map((id) => (
            <option key={id} value={id}>
              {etiquetaClase(id)}
            </option>
          ))}
        </select>
      </label>
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && items.length === 0 && !error && <p className="empty">Esta clase no tiene ítems.</p>}
      <ListaCatalogo
        items={items}
        editable={props.editable}
        onDesactivar={async (id) => {
          await desactivarCatalogo(id)
          await load()
        }}
        onRenombrar={async (id, next) => {
          try {
            await renombrarCatalogo(id, next)
            setError("")
            await load()
          } catch (err) {
            setError(err instanceof Error ? err.message : "No se pudo guardar")
            throw err
          }
        }}
      />
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

function comoEstado(item: CatalogoItem) {
  return {
    codigo: item.codigo,
    nombre: item.nombre,
    activo: item.activo,
    orden: item.orden,
    provisional: item.provisional,
  }
}
