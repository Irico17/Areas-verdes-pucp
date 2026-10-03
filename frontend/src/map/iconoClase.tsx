import { TRAZO_CLASE, iconoDe } from "./trazosClase"

export function IconoClase({ clase = "", tipo = "" }: { clase?: string; tipo?: string }) {
  const id = iconoDe(clase, tipo)
  return (
    <svg className="icono-clase" viewBox="0 0 24 24" aria-hidden="true" data-icono={id}>
      <path d={TRAZO_CLASE[id]} fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  )
}
