import { ZONIFICACION } from "../ui/nomenclatura"
import type { LugarCatalogo } from "./zonificacion"

type Props = {
  id: string
  lugares: LugarCatalogo[]
  lugarId: string
  lugarLibre?: string
  onChange: (lugarId: string) => void
}

export function SelectorLugar({ id, lugares, lugarId, lugarLibre, onChange }: Props) {
  const historico = lugarLibre?.trim() ?? ""
  return (
    <div className="lugar-selector">
      <label className="field" htmlFor={id}>
        {ZONIFICACION.lugar}
        <select id={id} name="lugar_id" value={lugarId} onChange={(event) => onChange(event.target.value)}>
          <option value="">{ZONIFICACION.sinLugar}</option>
          {lugares.map((lugar) => (
            <option key={lugar.id} value={String(lugar.id)}>
              {lugar.nombre}
            </option>
          ))}
        </select>
      </label>
      {historico && !lugarId && (
        <p className="hint lugar-historico">
          {ZONIFICACION.lugarCargado}: {historico}
        </p>
      )}
    </div>
  )
}
