import { useState } from "react"
import { claveRecorrido, pasosDe } from "./ayuda"
import { CUENTA } from "./nomenclatura"

export function Recorrido(props: { rol: string; usuario: string; forzar?: boolean; onCerrar?: () => void }) {
  const clave = claveRecorrido(props.usuario)
  const [cerrado, setCerrado] = useState(() => !props.forzar && localStorage.getItem(clave) === "1")
  const [paso, setPaso] = useState(0)
  const [forzado, setForzado] = useState(props.forzar === true)
  if (props.forzar && !forzado) {
    setForzado(true)
    setCerrado(false)
    setPaso(0)
  } else if (!props.forzar && forzado) {
    setForzado(false)
  }
  const pasos = pasosDe(props.rol)
  if (cerrado) return null
  const actual = pasos[paso]
  if (!actual) return null

  function guardar() {
    localStorage.setItem(clave, "1")
    setCerrado(true)
    props.onCerrar?.()
  }

  return (
    <div className="recorrido" role="dialog" aria-labelledby="recorrido-titulo">
      <p className="recorrido-paso" id="recorrido-titulo">
        {actual.titulo}
      </p>
      <p>{actual.texto}</p>
      <div className="row-actions">
        <button type="button" onClick={guardar}>
          {CUENTA.omitir}
        </button>
        {paso < pasos.length - 1 ? (
          <button type="button" className="primary" onClick={() => setPaso((n) => n + 1)}>
            {CUENTA.siguiente}
          </button>
        ) : (
          <button type="button" className="primary" onClick={guardar}>
            {CUENTA.listo}
          </button>
        )}
      </div>
    </div>
  )
}
