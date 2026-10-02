import { useId, useState } from "react"
import { CUENTA } from "./nomenclatura"

export function AyudaTip(props: { texto: string; etiqueta: string }) {
  const id = useId()
  const [abierto, setAbierto] = useState(false)
  return (
    <span className="ayuda-tip">
      <button
        type="button"
        className="ayuda-boton"
        aria-describedby={abierto ? id : undefined}
        aria-expanded={abierto}
        aria-label={props.etiqueta}
        onClick={() => setAbierto((actual) => !actual)}
        onKeyDown={(event) => {
          if (event.key === "Escape") setAbierto(false)
        }}
      >
        ?
      </button>
      {abierto && (
        <span className="ayuda-texto" id={id} role="note">
          {props.texto}
          <button type="button" className="link" onClick={() => setAbierto(false)}>
            {CUENTA.cerrarAyuda}
          </button>
        </span>
      )}
    </span>
  )
}
