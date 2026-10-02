import { useState } from "react"
import { formatFecha, parseFechaPE } from "./fecha"

export function FechaCampo(props: { id?: string; value: string; onChange: (iso: string) => void; required?: boolean }) {
  const [text, setText] = useState(() => (props.value ? formatFecha(props.value) : ""))
  const [bad, setBad] = useState(false)
  const [externo, setExterno] = useState(props.value)
  if (props.value !== externo) {
    setExterno(props.value)
    // Solo si el cambio vino de fuera: no reformatear mientras se escribe.
    if (props.value !== (parseFechaPE(text) ?? "")) {
      setText(props.value ? formatFecha(props.value) : "")
      setBad(false)
    }
  }

  return (
    <input
      id={props.id}
      inputMode="numeric"
      autoComplete="off"
      placeholder="dd/mm/aaaa"
      value={text}
      aria-invalid={bad}
      required={props.required}
      onChange={(event) => {
        const next = event.target.value
        setText(next)
        if (next.trim() === "") {
          setBad(false)
          props.onChange("")
          return
        }
        const iso = parseFechaPE(next)
        setBad(iso == null)
        // Texto inválido: el padre no debe conservar la última fecha válida
        // (si no, un formulario podría enviarla mientras el campo se ve inválido).
        props.onChange(iso ?? "")
      }}
    />
  )
}
