import { useState, type FormEvent } from "react"
import { ENTIDADES, GRUPOS_IMPORTACION, confirmarImportacion, previsualizar, revertirLote, type VistaPrevia } from "./importaciones"
import { IMPORTAR_UI } from "../ui/nomenclatura"

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
      setAviso(previa.validas === 0 ? "Ninguna fila válida. No se escribió nada." : "Archivo revisado. Confirme para escribir.")
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

  const columnas = vista ? Array.from(new Set(vista.filas.flatMap((fila) => Object.keys(fila)))) : []

  return (
    <section className="form">
      <h2>{IMPORTAR_UI.titulo}</h2>
      <p className="lede">{IMPORTAR_UI.lede}</p>
      <form onSubmit={(event) => void vistaPrevia(event)}>
        <label className="field">
          {IMPORTAR_UI.paso1}
          <select value={entidad} onChange={(event) => setEntidad(event.target.value)}>
            {GRUPOS_IMPORTACION.map((grupo) => (
              <optgroup key={grupo.id} label={grupo.etiqueta}>
                {grupo.ids.map((id) => {
                  const item = ENTIDADES.find((fila) => fila.id === id)
                  if (!item) return null
                  return (
                    <option key={item.id} value={item.id}>
                      {item.etiqueta}
                    </option>
                  )
                })}
              </optgroup>
            ))}
          </select>
        </label>
        <label className="field">
          {IMPORTAR_UI.paso2}
          <input
            type="file"
            accept=".csv,.xlsx,.geojson,.json,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/geo+json"
            onChange={(event) => setArchivo(event.target.files?.[0] ?? null)}
          />
        </label>
        <p className="hint" id="importar-motivo">
          {!archivo ? IMPORTAR_UI.faltaArchivo : IMPORTAR_UI.paso3}
        </p>
        <button type="submit" className="primary" disabled={pending || !archivo} aria-describedby="importar-motivo">
          {pending ? "Revisando…" : IMPORTAR_UI.revisar}
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
            <>
              <h3>Primeras {vista.filas.length} filas</h3>
              <div className="tabla-scroll" role="region" aria-label={IMPORTAR_UI.region} tabIndex={0}>
                <table className="tabla">
                  <thead>
                    <tr>{columnas.map((col) => <th key={col} scope="col">{col}</th>)}</tr>
                  </thead>
                  <tbody>
                    {vista.filas.map((fila, i) => (
                      <tr key={i}>
                        {columnas.map((col) => {
                          const valor = String(fila[col] ?? "")
                          return (
                            <td key={col}>
                              <span className="celda" title={valor}>{valor}</span>
                            </td>
                          )
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
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
        <p className="hint" id="importar-confirmar">
          {IMPORTAR_UI.paso4}. {!vista ? IMPORTAR_UI.faltaRevision : IMPORTAR_UI.confirmar}
        </p>
        <button type="button" className="primary" disabled={pending || !vista || vista.validas === 0 || vista.escrito} aria-describedby="importar-confirmar" onClick={() => void confirmar()}>
          {IMPORTAR_UI.confirmar}
        </button>
        <button type="button" disabled={pending || !vista?.escrito} onClick={() => void revertir()}>
          {IMPORTAR_UI.revertir}
        </button>
      </div>
    </section>
  )
}
