import { useState } from "react"
import { formatFecha, parseFechaPE } from "./fecha"

export function FechaCampo(props: { value: string; onChange: (iso: string) => void; required?: boolean }) {
  const [text, setText] = useState(() => (props.value ? formatFecha(props.value) : ""))
  const [bad, setBad] = useState(false)

  return (
    <input
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
        if (iso) props.onChange(iso)
      }}
    />
  )
}
