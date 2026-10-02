import { IconoClase } from "./iconoClase"
import { etiquetaEstado, etiquetaTipo } from "../operacion"
import { ACTIVIDAD, VISTA_ACTIVIDAD, marcaEjecutor } from "../ui/nomenclatura"

export type Vista = "mapa" | "lista"
export type MotivoPlano = "red" | "teselas" | null

type Item = {
  id: string
  titulo: string
  tipo: string
  clase?: string
  estado: string
  equipo: string
  ejecutor?: string
}

export function ConmutadorVista({
  vista,
  motivo,
  onVista,
}: {
  vista: Vista
  motivo: MotivoPlano
  onVista: (vista: Vista) => void
}) {
  const forzada = motivo != null
  const texto = motivo === "red" ? VISTA_ACTIVIDAD.sinRed : motivo === "teselas" ? VISTA_ACTIVIDAD.sinTeselas : ""
  return (
    <div className="vista-actividades">
      <div className="roles" role="group" aria-label={VISTA_ACTIVIDAD.grupo}>
        <button type="button" aria-pressed={vista === "mapa"} disabled={forzada} onClick={() => onVista("mapa")}>
          {VISTA_ACTIVIDAD.mapa}
        </button>
        <button type="button" aria-pressed={vista === "lista"} onClick={() => onVista("lista")}>
          {VISTA_ACTIVIDAD.lista}
        </button>
      </div>
      {texto && (
        <p className="vista-motivo" role="status">
          {texto}
        </p>
      )}
    </div>
  )
}

export function ListaActividades({
  items,
  selectedId,
  onSelect,
}: {
  items: Item[]
  selectedId: string | null
  onSelect: (id: string) => void
}) {
  return (
    <div className="lista-sobre-mapa">
      <h2>{ACTIVIDAD.titulo}</h2>
      <ul className="labor-list">
        {items.length === 0 && <li className="empty">{VISTA_ACTIVIDAD.vacio}</li>}
        {items.map((item) => (
          <li key={item.id}>
            <button type="button" className={selectedId === item.id ? "labor on" : "labor"} onClick={() => onSelect(item.id)}>
              <span className="marca" data-estado={item.estado}>
                <IconoClase clase={item.clase} tipo={item.tipo} />
              </span>
              <span>
                <strong>{item.titulo}</strong>
                <small>
                  {etiquetaTipo(item.tipo)} · {etiquetaEstado(item.estado)} · {item.equipo || ACTIVIDAD.sinCuadrilla}
                  {item.ejecutor === "tercerizada" ? ` · ${marcaEjecutor(item.ejecutor)}` : ""}
                </small>
              </span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
