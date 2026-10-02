import { useEffect, useId, useState, type FormEvent } from "react"
import { formatFechaHora } from "../fecha"
import { etiquetaEstado, etiquetaEvento, type Evento } from "../operacion"
import { urlArchivo } from "../offline/subida"
import { etiquetaRol } from "../producto"
import type { Rol } from "../types"
import { BITACORA } from "../ui/nomenclatura"
import { adjuntarEvidencia, anotarHito, leerCadena, listarActividadesBitacora, TIPOS_HITO, type ActividadBitacora } from "./bitacora"

type Props = {
  actividadId: string
  eventos: Evento[]
  error?: string
  embebida?: boolean
  autonomo?: boolean
  sinTitulo?: boolean
}

function nombre(valor: string | undefined, vacio: string): string {
  const limpio = valor?.trim()
  return limpio ? limpio : vacio
}

function mensaje(error: unknown): string {
  return error instanceof Error && error.message ? error.message : BITACORA.error
}

export function Bitacora(props: Props) {
  const [propias, setPropias] = useState<Evento[] | null>(null)
  const [firma, setFirma] = useState(props.eventos)
  const [fallo, setFallo] = useState<string | null>(null)
  const [ocupado, setOcupado] = useState("")
  const [tipo, setTipo] = useState<(typeof TIPOS_HITO)[number]["id"]>("inicio")
  const [texto, setTexto] = useState("")
  const tituloId = useId()
  if (!props.autonomo && props.eventos !== firma) {
    setFirma(props.eventos)
    setPropias(null)
    setFallo(null)
  }
  const filas = propias ?? props.eventos
  const aviso = fallo ?? (props.autonomo ? "" : (props.error ?? ""))

  useEffect(() => {
    if (!props.autonomo || !props.actividadId) return
    let vivo = true
    leerCadena(props.actividadId)
      .then((rows) => {
        if (!vivo) return
        setPropias(rows)
        setFallo(null)
      })
      .catch((error: unknown) => {
        if (vivo) setFallo(mensaje(error))
      })
    return () => {
      vivo = false
    }
  }, [props.autonomo, props.actividadId])

  async function recargar() {
    if (!props.actividadId) return
    const rows = await leerCadena(props.actividadId)
    setPropias(rows)
    setFallo(null)
  }

  async function onAnotar(event: FormEvent) {
    event.preventDefault()
    if (!props.actividadId || ocupado) return
    setOcupado(BITACORA.anotando)
    setFallo("")
    try {
      await anotarHito(props.actividadId, tipo, texto)
      setTexto("")
      await recargar()
    } catch (error: unknown) {
      setFallo(mensaje(error))
    } finally {
      setOcupado("")
    }
  }

  async function onArchivo(eventoId: number, archivo: File | undefined) {
    if (!archivo || !props.actividadId || ocupado) return
    setOcupado(BITACORA.subiendo)
    setFallo("")
    try {
      await adjuntarEvidencia(props.actividadId, eventoId, archivo)
      await recargar()
    } catch (error: unknown) {
      setFallo(mensaje(error))
    } finally {
      setOcupado("")
    }
  }

  const Titulo = props.embebida ? "h3" : "h2"
  const tituloEnElPanel = props.embebida && props.autonomo

  return (
    <section className="bitacora" aria-labelledby={tituloEnElPanel || props.sinTitulo ? undefined : tituloId} aria-label={tituloEnElPanel || props.sinTitulo ? BITACORA.titulo : undefined}>
      {!tituloEnElPanel && !props.sinTitulo && <Titulo id={tituloId}>{BITACORA.titulo}</Titulo>}
      {!props.embebida && <p className="lede">{BITACORA.lede}</p>}
      {aviso && <p className="status error">{aviso}</p>}
      {props.autonomo && propias === null && !aviso && <p className="hint">{BITACORA.cargando}</p>}
      {filas.length === 0 && !(props.autonomo && propias === null) && <p className="bitacora-vacio">{BITACORA.vacio}</p>}
      <ol className="bitacora-lista">
        {filas.map((evento) => (
          <li key={evento.id} className="bitacora-hito" data-tipo={evento.tipo}>
            <time dateTime={evento.created_at}>{formatFechaHora(evento.created_at)}</time>
            <div>
              <strong>{etiquetaEvento(evento.tipo)}</strong>
              {evento.estado ? <span className="bitacora-estado">{etiquetaEstado(evento.estado)}</span> : null}
              {evento.tipo === "reasignada" && (
                <dl className="bitacora-cuadrillas">
                  <div>
                    <dt>{BITACORA.anterior}</dt>
                    <dd>{nombre(evento.cuadrilla_anterior, BITACORA.sinCuadrilla)}</dd>
                  </div>
                  <div>
                    <dt>{BITACORA.nueva}</dt>
                    <dd>{nombre(evento.equipo, BITACORA.sinCuadrilla)}</dd>
                  </div>
                </dl>
              )}
              {evento.nota ? <p>{evento.nota}</p> : null}
              <p className="bitacora-quien">
                {BITACORA.quien} <span>{evento.usuario_nombre || etiquetaRol(evento.actor_rol)}</span>
              </p>
              {(evento.evidencias ?? []).length > 0 && (
                <ul className="bitacora-fotos">
                  {(evento.evidencias ?? []).map((archivo) => (
                    <li key={archivo.id}>
                      {archivo.mime.startsWith("image/") ? (
                        <figure>
                          <img src={urlArchivo(archivo.id)} alt={archivo.nombre} />
                          <figcaption>{archivo.nombre}</figcaption>
                        </figure>
                      ) : (
                        <a href={urlArchivo(archivo.id)}>{archivo.nombre}</a>
                      )}
                    </li>
                  ))}
                </ul>
              )}
              {props.actividadId && (
                <label className="bitacora-archivo">
                  {ocupado === BITACORA.subiendo ? BITACORA.subiendo : BITACORA.adjuntar}
                  <input
                    type="file"
                    accept="image/jpeg,image/png,image/webp,application/pdf"
                    disabled={ocupado !== ""}
                    onChange={(event) => {
                      const archivo = event.target.files?.[0]
                      event.target.value = ""
                      void onArchivo(evento.id, archivo)
                    }}
                  />
                </label>
              )}
            </div>
          </li>
        ))}
      </ol>
      {props.actividadId && (
        <form className="bitacora-anotar" onSubmit={(event) => void onAnotar(event)}>
          <label className="field">
            {BITACORA.tipo}
            <select value={tipo} onChange={(event) => setTipo(event.target.value as (typeof TIPOS_HITO)[number]["id"])}>
              {TIPOS_HITO.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            {BITACORA.texto}
            <textarea value={texto} maxLength={500} required onChange={(event) => setTexto(event.target.value)} />
          </label>
          <button type="submit" className="primary" disabled={ocupado !== "" || texto.trim() === ""}>
            {ocupado === BITACORA.anotando ? BITACORA.anotando : BITACORA.anotar}
          </button>
        </form>
      )}
    </section>
  )
}

export function BitacoraPanel(props: { rol: Rol; capatazId: string; actividadId: string }) {
  const [items, setItems] = useState<ActividadBitacora[]>([])
  const [elegida, setElegida] = useState(props.actividadId)
  const [vista, setVista] = useState(props.actividadId)
  const [fallo, setFallo] = useState("")
  if (props.actividadId !== vista) {
    setVista(props.actividadId)
    if (props.actividadId) setElegida(props.actividadId)
  }

  useEffect(() => {
    let vivo = true
    listarActividadesBitacora(props.rol, props.capatazId)
      .then((lista) => {
        if (!vivo) return
        setItems(lista)
        setElegida((actual) => actual || lista[0]?.id || "")
        setFallo("")
      })
      .catch((error: unknown) => {
        if (vivo) setFallo(mensaje(error))
      })
    return () => {
      vivo = false
    }
  }, [props.rol, props.capatazId])

  return (
    <section className="bitacora-panel">
      <header className="bitacora-cabeza">
        <h2>{BITACORA.titulo}</h2>
        <p className="lede">{BITACORA.lede}</p>
      </header>
      {fallo && <p className="status error">{fallo}</p>}
      <div className="bitacora-reparto">
        <ul className="bitacora-actividades" aria-label={BITACORA.actividades}>
          {items.length === 0 && <li className="empty">{BITACORA.sinActividad}</li>}
          {items.map((item) => (
            <li key={item.id}>
              <button type="button" aria-pressed={item.id === elegida} onClick={() => setElegida(item.id)}>
                {item.titulo}
              </button>
            </li>
          ))}
        </ul>
        {elegida ? (
          <Bitacora actividadId={elegida} eventos={[]} autonomo embebida />
        ) : (
          <p className="bitacora-vacio">{BITACORA.sinActividad}</p>
        )}
      </div>
    </section>
  )
}
