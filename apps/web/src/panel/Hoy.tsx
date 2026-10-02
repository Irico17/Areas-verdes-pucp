import { hoyISO } from "../fecha"
import { etiquetaEstado } from "../operacion"
import type { LaborItem } from "./Labores"
import { AyudaTip } from "../ui/AyudaTip"
import { AYUDA, HOY } from "../ui/nomenclatura"

const ORDEN = ["pendiente", "en_proceso", "bloqueada", "ejecutado"]

type Props = {
  cuadrilla: string
  enLinea: boolean
  porEnviar: number
  items: LaborItem[]
  filtroEstado: string
  onFiltro: (estado: string) => void
  onAccion: (item: LaborItem) => void
  onAbrir: (id: string) => void
  onRiego: () => void
  onPoda: () => void
  onVivero: () => void
  onBitacora?: () => void
  onEjemplares?: () => void
  vista: "lista" | "mapa"
  onVista: (vista: "lista" | "mapa") => void
  mapaBloqueado: boolean
}

function fechaVisible(): string {
  const iso = hoyISO()
  const [anio, mes, dia] = iso.split("-").map(Number)
  return new Intl.DateTimeFormat("es-PE", { day: "numeric", month: "long", timeZone: "America/Lima" }).format(
    new Date(Date.UTC(anio, mes - 1, dia, 12)),
  )
}

function accionDe(estado: string): string | null {
  if (estado === "pendiente") return HOY.iniciar
  if (estado === "en_proceso") return HOY.terminar
  if (estado === "bloqueada") return HOY.verMotivo
  return null
}

export function Hoy(props: Props) {
  const conteo = (estado: string) => props.items.filter((item) => item.estado === estado).length
  const visibles = [...props.items]
    .filter((item) => !props.filtroEstado || item.estado === props.filtroEstado)
    .sort((a, b) => ORDEN.indexOf(a.estado) - ORDEN.indexOf(b.estado) || a.titulo.localeCompare(b.titulo, "es"))

  return (
    <section className="hoy" aria-labelledby="hoy-titulo">
      <header className="hoy-cabecera">
        <h2 id="hoy-titulo">
          {props.cuadrilla}
          <span className="hoy-fecha"> · {fechaVisible()}</span>
        </h2>
        <p className={props.enLinea ? "chip" : "chip chip-alerta"} role="status">
          {props.enLinea ? HOY.enLinea : HOY.sinConexion(props.porEnviar)}
        </p>
      </header>
      <div className="hoy-contadores" role="group" aria-label="Estados de la cuadrilla">
        {(
          [
            ["pendiente", HOY.porIniciar],
            ["en_proceso", HOY.enProceso],
            ["bloqueada", HOY.bloqueadas],
          ] as const
        ).map(([estado, etiqueta]) => (
          <button
            key={estado}
            type="button"
            className={props.filtroEstado === estado ? "contador on" : "contador"}
            aria-pressed={props.filtroEstado === estado}
            onClick={() => props.onFiltro(props.filtroEstado === estado ? "" : estado)}
          >
            <span className="contador-n">{conteo(estado)}</span>
            <span>{etiqueta}</span>
          </button>
        ))}
        <AyudaTip texto={AYUDA.bloqueada} etiqueta="Qué significa Bloqueada" />
      </div>
      <div className="roles" role="group" aria-label="Vista de hoy">
        <button type="button" aria-pressed={props.vista === "lista" || props.mapaBloqueado} onClick={() => props.onVista("lista")}>
          {HOY.lista}
        </button>
        <button type="button" aria-pressed={props.vista === "mapa" && !props.mapaBloqueado} disabled={props.mapaBloqueado} onClick={() => props.onVista("mapa")}>
          {HOY.mapa}
        </button>
      </div>
      {props.mapaBloqueado && (
        <p className="status" role="status">
          {HOY.forzada}
        </p>
      )}
      <h3>{HOY.actividades}</h3>
      {visibles.length === 0 && <p className="empty">{HOY.vacio}</p>}
      <ul className="labor-list">
        {visibles.map((item) => {
          const accion = accionDe(item.estado)
          return (
            <li key={item.id} className="hoy-fila">
              <button type="button" className="labor" onClick={() => props.onAbrir(item.id)}>
                <strong>{item.titulo}</strong>
                <small>{etiquetaEstado(item.estado)}</small>
              </button>
              {accion && (
                <button type="button" className="primary" onClick={() => props.onAccion(item)}>
                  {accion}
                </button>
              )}
            </li>
          )
        })}
      </ul>
      {(props.onBitacora || props.onEjemplares) && (
        <p className="row-actions">
          {props.onBitacora && (
            <button type="button" className="link" onClick={props.onBitacora}>
              {HOY.bitacora}
            </button>
          )}
          {props.onEjemplares && (
            <button type="button" className="link" onClick={props.onEjemplares}>
              {HOY.ejemplares}
            </button>
          )}
        </p>
      )}
      <div className="accion-fija">
        <button type="button" className="primary" onClick={props.onRiego}>
          {HOY.riego}
        </button>
        <details className="menu-secundario">
          <summary>{HOY.mas}</summary>
          <button type="button" onClick={props.onPoda}>
            {HOY.poda}
          </button>
          <button type="button" onClick={props.onVivero}>
            {HOY.vivero}
          </button>
        </details>
      </div>
    </section>
  )
}
