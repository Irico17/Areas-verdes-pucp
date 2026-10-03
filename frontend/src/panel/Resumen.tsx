import { useEffect, useState } from "react"
import { apiUrl } from "../api"
import { hoyISO } from "../fecha"
import { FILTRO_ACTIVIDADES, etiquetaEstado, type FiltroActividades } from "../operacion"
import { HUECOS, etiquetaRol, fetchOrdenes, fetchReporte, fetchRiego, fetchSolicitudes, reporteHref, type Orden, type Riego, type Solicitud } from "../producto"
import { RESUMEN } from "../ui/nomenclatura"
import { Esqueleto } from "../ui/Esqueleto"
import type { LaborItem } from "./Labores"

type Cambio = { id: number; entidad: string; accion: string; nombre?: string; usuario?: string }

type Props = {
  rol: string
  items: LaborItem[]
  porEnviar: number
  puedeValidar: boolean
  puedeSolicitudes: boolean
  puedeUsuarios: boolean
  puedeReportes: boolean
  onVerActividades: (filtro: FiltroActividades) => void
  onNueva: () => void
  onSolicitud: () => void
  onReportes: () => void
  onImportar: () => void
}

function Bloque(props: { titulo: string; filas: { id: string; texto: string; meta?: string }[]; onVer: () => void }) {
  return (
    <section className="resumen-bloque">
      <header>
        <h3>{props.titulo}</h3>
        <span className="contador-n">{props.filas.length}</span>
      </header>
      {props.filas.length === 0 && <p className="empty">{RESUMEN.vacio}</p>}
      <ul>
        {props.filas.slice(0, 3).map((fila) => (
          <li key={fila.id}>
            <strong>{fila.texto}</strong>
            {fila.meta && <small>{fila.meta}</small>}
          </li>
        ))}
      </ul>
      {props.filas.length > 0 && (
        <button type="button" className="link" onClick={props.onVer}>
          {RESUMEN.verTodas}
        </button>
      )}
    </section>
  )
}

export function Resumen(props: Props) {
  const [solicitudes, setSolicitudes] = useState<Solicitud[] | null>(null)
  const [ordenes, setOrdenes] = useState<Orden[] | null>(null)
  const [riego, setRiego] = useState<Riego[] | null>(null)
  const [porEstado, setPorEstado] = useState<{ estado: string; n: number }[] | null>(null)
  const [cuentas, setCuentas] = useState<{ rol: string; n: number }[] | null>(null)
  const [cambios, setCambios] = useState<Cambio[] | null>(null)
  const [error, setError] = useState("")
  const hoy = hoyISO()

  useEffect(() => {
    let vivo = true
    const tareas: Promise<void>[] = []
    if (props.puedeSolicitudes) {
      tareas.push(
        fetchSolicitudes()
          .then((filas) => {
            if (vivo) setSolicitudes(filas)
          })
          .catch(() => {
            if (vivo) setSolicitudes([])
          }),
        fetchOrdenes()
          .then((filas) => {
            if (vivo) setOrdenes(filas)
          })
          .catch(() => {
            if (vivo) setOrdenes([])
          }),
      )
    }
    tareas.push(
      fetchRiego()
        .then((body) => {
          if (vivo) setRiego(body.registros)
        })
        .catch(() => {
          if (vivo) setRiego([])
        }),
    )
    if (props.puedeReportes && props.rol === "jefatura") {
      tareas.push(
        fetchReporte({})
          .then((body) => {
            if (vivo) setPorEstado(body.por_estado ?? [])
          })
          .catch((err: unknown) => {
            if (vivo) setError(err instanceof Error ? err.message : RESUMEN.operacion)
          }),
      )
    }
    if (props.puedeUsuarios && props.rol === "admin") {
      tareas.push(
        fetch(apiUrl("/accesos/usuarios"), { credentials: "include" })
          .then((res) => (res.ok ? res.json() : { usuarios: [] }))
          .then((body: { usuarios?: { rol: string }[] }) => {
            if (!vivo) return
            const mapa = new Map<string, number>()
            for (const cuenta of body.usuarios ?? []) mapa.set(cuenta.rol, (mapa.get(cuenta.rol) ?? 0) + 1)
            setCuentas([...mapa.entries()].map(([rol, n]) => ({ rol, n })))
          })
          .catch(() => {
            if (vivo) setCuentas([])
          }),
        fetch(apiUrl("/auditoria/cambios"), { credentials: "include" })
          .then((res) => (res.ok ? res.json() : { eventos: [] }))
          .then((body: { eventos?: Cambio[] }) => {
            if (vivo) setCambios(body.eventos ?? [])
          })
          .catch(() => {
            if (vivo) setCambios([])
          }),
      )
    }
    void Promise.all(tareas)
    return () => {
      vivo = false
    }
  }, [props.puedeSolicitudes, props.puedeReportes, props.puedeUsuarios, props.rol])

  const bloqueadas = props.items.filter((item) => item.estado === "bloqueada")
  const sinCuadrilla = props.items.filter((item) => !item.capatazId)
  const ordenesVisibles = props.puedeSolicitudes ? ordenes : []
  const solicitudesVisibles = props.puedeSolicitudes ? solicitudes : []
  const conOrden = new Set((ordenesVisibles ?? []).map((orden) => orden.actividad_id))
  const tercerizadas = props.items.filter((item) => item.ejecutor === "tercerizada" && !conOrden.has(item.id))
  const sueltas = (solicitudesVisibles ?? []).filter((item) => !item.actividad_id)
  const riegoHoy = (riego ?? []).filter((item) => item.fecha.slice(0, 10) === hoy)
  const cargando = (props.puedeSolicitudes && solicitudes === null) || riego === null

  return (
    <section className="resumen" aria-labelledby="resumen-titulo">
      <h2 id="resumen-titulo">{RESUMEN.titulo}</h2>
      <p className="lede">{RESUMEN.lede}</p>
      {error && <p className="status error">{error}</p>}
      {props.rol === "jefatura" && (
        <section className="resumen-bloque">
          <h3>{RESUMEN.operacion}</h3>
          {porEstado === null && <Esqueleto filas={3} />}
          <ul className="resumen-estados">
            {(porEstado ?? []).map((fila) => (
              <li key={fila.estado}>
                <span>{etiquetaEstado(fila.estado)}</span>
                <strong>{fila.n}</strong>
              </li>
            ))}
          </ul>
          <p className="hint">{RESUMEN.provisional}</p>
          <p className="row-actions">
            <button type="button" onClick={props.onReportes}>
              {RESUMEN.verReporte}
            </button>
            <a className="quiet" href={reporteHref("xls", {})}>
              {RESUMEN.excel}
            </a>
          </p>
          <details>
            <summary>{RESUMEN.indicadores}</summary>
            <ul>
              {HUECOS.map((hueco) => (
                <li key={hueco.clave}>
                  {hueco.nombre}: {hueco.estado}. {hueco.nota}
                </li>
              ))}
            </ul>
          </details>
        </section>
      )}
      {props.rol === "admin" && (
        <section className="resumen-bloque">
          <h3>{RESUMEN.cuentas}</h3>
          <ul>
            {(cuentas ?? []).map((fila) => (
              <li key={fila.rol}>
                <span>{etiquetaRol(fila.rol)}</span>
                <strong>{fila.n}</strong>
              </li>
            ))}
          </ul>
          <h3>{RESUMEN.cambios}</h3>
          <ul>
            {(cambios ?? []).slice(0, 3).map((fila) => (
              <li key={fila.id}>
                <strong>{fila.entidad}</strong>
                <small>
                  {fila.accion}
                  {fila.nombre ? ` · ${fila.nombre}` : ""}
                </small>
              </li>
            ))}
          </ul>
          <button type="button" className="primary" onClick={props.onImportar}>
            {RESUMEN.importar}
          </button>
        </section>
      )}
      {cargando && props.rol !== "admin" && <Esqueleto filas={4} />}
      {!cargando && props.rol !== "admin" && (
        <>
          <Bloque
            titulo={RESUMEN.bloqueadas}
            filas={bloqueadas.map((item) => ({ id: item.id, texto: item.titulo, meta: etiquetaEstado(item.estado) }))}
            onVer={() => props.onVerActividades({ ...FILTRO_ACTIVIDADES, estado: "bloqueada" })}
          />
          <Bloque
            titulo={RESUMEN.sinCuadrilla}
            filas={sinCuadrilla.map((item) => ({ id: item.id, texto: item.titulo }))}
            onVer={() => props.onVerActividades({ ...FILTRO_ACTIVIDADES, cuadrillaId: "" })}
          />
          <Bloque
            titulo={RESUMEN.tercerizadas}
            filas={tercerizadas.map((item) => ({ id: item.id, texto: item.titulo }))}
            onVer={() => props.onVerActividades({ ...FILTRO_ACTIVIDADES, ejecutor: "tercerizada" })}
          />
          <Bloque
            titulo={RESUMEN.solicitudes}
            filas={sueltas.map((item) => ({ id: item.id, texto: item.titulo }))}
            onVer={props.onSolicitud}
          />
          <Bloque
            titulo={RESUMEN.sincronizar}
            filas={props.porEnviar > 0 ? [{ id: "cola", texto: `${props.porEnviar} por enviar` }] : []}
            onVer={() => props.onVerActividades(FILTRO_ACTIVIDADES)}
          />
          <Bloque
            titulo={RESUMEN.riego}
            filas={riegoHoy.map((item) => ({ id: item.id, texto: item.sector, meta: item.turno }))}
            onVer={() => props.onVerActividades(FILTRO_ACTIVIDADES)}
          />
        </>
      )}
      <div className="accion-fija">
        {props.puedeValidar && (
          <button type="button" className="primary" onClick={props.onNueva}>
            {RESUMEN.nueva}
          </button>
        )}
        {props.puedeSolicitudes && props.rol !== "admin" && (
          <button type="button" onClick={props.onSolicitud}>
            {RESUMEN.solicitud}
          </button>
        )}
      </div>
    </section>
  )
}
