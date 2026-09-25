import { useState, type FormEvent } from "react"
import { ENTIDADES, confirmarImportacion, previsualizar, revertirLote, type VistaPrevia } from "./importaciones"

export function ImportacionesPanel() {
  const [entidad, setEntidad] = useState(ENTIDADES[0].id)
  const [archivo, setArchivo] = useState<File | null>(null)
  const [vista, setVista] = useState<VistaPrevia | null>(null)
  const [aviso, setAviso] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  async function vistaPrevia(event: FormEvent) {
    event.preventDefault()
    if (!archivo) {
      setError("Elija un archivo CSV, XLSX o GeoJSON.")
      return
    }
    setPending(true)
    setError("")
    setAviso("")
    try {
      const previa = await previsualizar(entidad, archivo)
      setVista(previa)
      setAviso(previa.validas === 0 ? "Ninguna fila válida. No se escribió nada." : "Vista previa lista. Confirme para escribir.")
    } catch (err) {
      setVista(null)
      setError(err instanceof Error ? err.message : "No se pudo leer el archivo")
    } finally {
      setPending(false)
    }
  }

  async function confirmar() {
    if (!vista) return
    setPending(true)
    setError("")
    try {
      const hecho = await confirmarImportacion(vista.id)
      setAviso(`Se escribieron ${hecho.validas} filas en el lote ${hecho.lote_id}.`)
      setVista({ ...vista, escrito: true, lote_id: hecho.lote_id })
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo confirmar")
    } finally {
      setPending(false)
    }
  }

  async function revertir() {
    if (!vista?.escrito) return
    setPending(true)
    setError("")
    try {
      const rep = await revertirLote(vista.lote_id, false)
      const extra = rep.excluidas?.length ? ` ${rep.excluidas.length} filas quedaron fuera porque alguien las editó después.` : ""
      setAviso(`Lote ${rep.lote_id} revertido.${extra}`)
      setVista({ ...vista, escrito: false })
    } catch (err) {
      const texto = err instanceof Error ? err.message : "No se pudo revertir"
      if (texto.includes("editadas")) {
        const seguir = window.confirm("Hay filas editadas después del lote. ¿Revierte solo las que nadie tocó?")
        if (!seguir) {
          setError(texto)
          setPending(false)
          return
        }
        try {
          const rep = await revertirLote(vista.lote_id, true)
          setAviso(`Lote ${rep.lote_id} revertido. ${rep.excluidas?.length ?? 0} filas no se pisaron.`)
          setVista({ ...vista, escrito: false })
        } catch (err2) {
          setError(err2 instanceof Error ? err2.message : "No se pudo revertir")
        }
      } else {
        setError(texto)
      }
    } finally {
      setPending(false)
    }
  }

  const columnas = vista?.filas[0] ? Object.keys(vista.filas[0]).slice(0, 6) : []

  return (
    <section className="form">
      <h2>Importar</h2>
      <p className="lede">Suba CSV, XLSX o GeoJSON. Primero se revisa; solo la confirmación escribe. El lote se puede revertir.</p>
      <form onSubmit={(event) => void vistaPrevia(event)}>
        <label className="field">
          Entidad
          <select value={entidad} onChange={(event) => setEntidad(event.target.value)}>
            {ENTIDADES.map((item) => (
              <option key={item.id} value={item.id}>
                {item.etiqueta}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Archivo
          <input
            type="file"
            accept=".csv,.xlsx,.geojson,.json,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/geo+json"
            onChange={(event) => setArchivo(event.target.files?.[0] ?? null)}
          />
        </label>
        <button type="submit" className="primary" disabled={pending}>
          {pending ? "Revisando…" : "Vista previa"}
        </button>
      </form>
      {error && <p className="status error">{error}</p>}
      {aviso && <p className="status">{aviso}</p>}
      {vista && (
        <div>
          <p>
            {vista.validas} filas válidas · {vista.errores.length} con error · formato {vista.formato}
          </p>
          {vista.aviso_omitidas && (
            <p>
              {vista.aviso_omitidas}
              {vista.columnas_omitidas?.length ? `: ${vista.columnas_omitidas.join(", ")}` : ""}
            </p>
          )}
          {vista.avisos?.map((item) => (
            <p key={item}>{item}</p>
          ))}
          {vista.filas.length > 0 && (
            <div className="labor-list">
              <p>Primeras {vista.filas.length} filas</p>
              {vista.filas.map((fila, i) => (
                <p key={i}>
                  {columnas.map((col) => `${col}: ${String(fila[col] ?? "")}`).join(" · ")}
                </p>
              ))}
            </div>
          )}
          {vista.errores.length > 0 && (
            <ul className="labor-list">
              {vista.errores.map((item, i) => (
                <li key={`${item.fila}-${item.campo}-${i}`}>
                  Fila {item.fila}, {item.campo}: {item.motivo}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
      <div className="row-actions">
        <button type="button" className="primary" disabled={pending || !vista || vista.validas === 0 || vista.escrito} onClick={() => void confirmar()}>
          Confirmar escritura
        </button>
        <button type="button" disabled={pending || !vista?.escrito} onClick={() => void revertir()}>
          Revertir lote
        </button>
      </div>
    </section>
  )
}
