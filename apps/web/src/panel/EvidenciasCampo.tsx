import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from "react"
import { formatFechaHora } from "../fecha"
import { comprimirFoto, rechazarPorTamano } from "../offline/comprimir"
import { leerExif } from "../offline/exif"
import {
  clasificarEstado,
  encolarEvidencia,
  enviarColaEvidencias,
  listarEvidencias,
  mensajeReintento,
  type QueuedEvidencia,
  type ResultadoEnvio,
} from "../offline/queue"
import {
  accionTrasFallo,
  avanzar,
  crearCacheUrls,
  enviarConProgreso,
  esImagen,
  ocupado,
  textoEstado,
  unirEvidencias,
  urlArchivo,
} from "../offline/subida"
import { fetchEvidencias, type Evidencia } from "../producto"
import { Esqueleto } from "../ui/Esqueleto"

async function sha256Hex(bytes: ArrayBuffer): Promise<string> {
  const dig = await crypto.subtle.digest("SHA-256", bytes)
  return [...new Uint8Array(dig)].map((b) => b.toString(16).padStart(2, "0")).join("")
}

function pedirPunto(): Promise<{ lat: number; lon: number } | null> {
  if (!navigator.geolocation) return Promise.resolve(null)
  return new Promise((resolve) => {
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ lat: pos.coords.latitude, lon: pos.coords.longitude }),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 5000, maximumAge: 60_000 },
    )
  })
}

async function publicar(
  item: QueuedEvidencia,
  alProgreso: (cargado: number, total: number | null) => void = () => {},
): Promise<ResultadoEnvio> {
  const data = new FormData()
  data.set("id", item.id)
  data.set("actividad_id", item.actividadId)
  data.set("nota", item.nota)
  data.set("sha256", item.sha256)
  data.set("exif", JSON.stringify(item.exif))
  if (item.lat != null && item.lon != null) {
    data.set("lat", String(item.lat))
    data.set("lon", String(item.lon))
  }
  if (item.ordenId) data.set("orden_id", item.ordenId)
  data.set("archivo", new Blob([item.bytes], { type: item.mime }), item.nombre)
  const status = await enviarConProgreso("/api/v1/evidencias", data, alProgreso)
  return clasificarEstado(status)
}

export function EvidenciasCampo({ actividadId }: { actividadId: string }) {
  const [enviadas, setEnviadas] = useState<Evidencia[]>([])
  const [pendientes, setPendientes] = useState<QueuedEvidencia[]>([])
  const [cargadoDe, setCargadoDe] = useState<string | null>(null)
  const cargando = cargadoDe !== actividadId
  const [estado, despachar] = useReducer(avanzar, { fase: "inactivo" } as const)
  const [rotas, setRotas] = useState<Set<string>>(new Set())
  const cacheRef = useRef(crearCacheUrls())

  useEffect(() => () => cacheRef.current.liberar(), [])

  const marcarRota = useCallback(
    (id: string) => () => setRotas((prev) => new Set(prev).add(id)),
    [],
  )

  const refrescar = useCallback(async () => {
    const locales = await listarEvidencias().catch(() => [])
    setPendientes(locales.filter((item) => item.actividadId === actividadId))
    try {
      setEnviadas(actividadId ? await fetchEvidencias(actividadId) : [])
    } catch {
      setEnviadas([])
    } finally {
      setCargadoDe(actividadId)
    }
  }, [actividadId])

  const drenar = useCallback(async () => {
    if (!navigator.onLine) return
    const r = await enviarColaEvidencias(publicar)
    if (r.reintentos.length > 0) {
      const ultimo = r.reintentos[r.reintentos.length - 1]
      despachar({ tipo: "cola", texto: mensajeReintento(ultimo.status) })
    } else if (r.conflictos.length > 0) {
      despachar({ tipo: "fallo", texto: "Esa foto ya estaba registrada con otro contenido.", reintentable: false })
    } else if (r.enviadas.length > 0) {
      despachar({ tipo: "ok", sinPunto: false })
    }
    await refrescar()
  }, [refrescar])

  useEffect(() => {
    const cargar = () => {
      void refrescar().then(() => drenar())
    }
    const id = window.setTimeout(cargar, 0)
    window.addEventListener("online", cargar)
    return () => {
      window.clearTimeout(id)
      window.removeEventListener("online", cargar)
    }
  }, [refrescar, drenar])

  const enviar = useCallback(
    async (item: QueuedEvidencia) => {
      const sinPunto = item.lat == null || item.lon == null
      if (!navigator.onLine) {
        await encolarEvidencia(item)
        despachar({
          tipo: "cola",
          texto: sinPunto ? "Guardado en este equipo, sin ubicación." : "Guardado en este equipo.",
        })
        await refrescar()
        return
      }
      const resultado = await publicar(item, (cargado, total) => despachar({ tipo: "progreso", cargado, total }))
      if (resultado.tipo === "ok") {
        despachar({ tipo: "ok", sinPunto })
      } else if (resultado.tipo === "conflicto") {
        despachar({ tipo: "fallo", texto: "Esa foto ya estaba registrada con otro contenido.", reintentable: false })
      } else {
        const accion = accionTrasFallo(resultado, item)
        if (accion) {
          await encolarEvidencia(accion.encolar)
          despachar(accion.evento)
        }
      }
      await refrescar()
    },
    [refrescar],
  )

  async function tomar(file: File) {
    if (!actividadId) return
    despachar({ tipo: "preparar" })
    try {
      if (rechazarPorTamano(file.size) && !file.type.startsWith("image/")) {
        despachar({ tipo: "fallo", texto: "El archivo supera 8 MB.", reintentable: false })
        return
      }
      const crudo = new Uint8Array(await file.arrayBuffer())
      const exif = leerExif(crudo)
      const listo = await comprimirFoto(file)
      let lat = exif.lat
      let lon = exif.lon
      if (lat == null || lon == null) {
        const punto = await pedirPunto()
        lat = punto?.lat ?? null
        lon = punto?.lon ?? null
      }
      const bytes = await listo.blob.arrayBuffer()
      const item: QueuedEvidencia = {
        id: crypto.randomUUID(),
        actividadId,
        nota: "",
        nombre: listo.nombre,
        mime: listo.mime,
        bytes,
        sha256: await sha256Hex(bytes),
        lat,
        lon,
        exif: {
          fecha: exif.fecha,
          lat,
          lon,
          ...(exif.orientacion != null ? { orientacion: exif.orientacion } : {}),
        },
        createdAt: new Date().toISOString(),
      }
      await enviar(item)
    } catch (error) {
      despachar({
        tipo: "fallo",
        texto: error instanceof Error ? error.message : "No se pudo preparar la foto.",
        reintentable: false,
      })
    }
  }

  const items = useMemo(() => unirEvidencias(pendientes, enviadas), [pendientes, enviadas])
  const [urls, setUrls] = useState<Map<string, string>>(new Map())
  useEffect(() => {
    setUrls(
      cacheRef.current.sincronizar(
        pendientes.map((item) => ({ id: item.id, blob: new Blob([item.bytes], { type: item.mime }) })),
      ),
    )
  }, [pendientes])

  const etiquetaCaptura =
    estado.fase === "preparando" ? "Preparando…" : estado.fase === "subiendo" ? "Subiendo…" : "Tomar foto"

  return (
    <section className="evidencia-campo" aria-label="Evidencias de la labor">
      <h3>Evidencia</h3>
      <p className="evidencia-ayuda">La foto se reduce en el teléfono antes de enviarse.</p>
      <div className="evidencia-acciones">
        <label className="primary evidencia-captura">
          {etiquetaCaptura}
          <input
            type="file"
            accept="image/*"
            capture="environment"
            disabled={ocupado(estado) || !actividadId}
            onChange={(event) => {
              const file = event.target.files?.[0]
              event.target.value = ""
              if (file) void tomar(file)
            }}
          />
        </label>
        <label className="evidencia-archivo">
          Elegir archivo
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp,application/pdf"
            disabled={ocupado(estado) || !actividadId}
            onChange={(event) => {
              const file = event.target.files?.[0]
              event.target.value = ""
              if (file) void tomar(file)
            }}
          />
        </label>
      </div>
      <div className="evidencia-estado" aria-live="polite">
        <p className={estado.fase === "error" ? "status error" : "status"}>{textoEstado(estado)}</p>
        {estado.fase === "preparando" && <progress aria-label="Preparando foto" />}
        {estado.fase === "subiendo" &&
          (estado.pct == null ? (
            <progress aria-label="Preparando foto" />
          ) : (
            <progress max={100} value={estado.pct} aria-label="Progreso de la subida" />
          ))}
        {estado.fase === "error" && estado.reintentable && (
          <button type="button" className="link" onClick={() => void drenar()}>
            Reintentar
          </button>
        )}
        {estado.fase === "en-cola" && (
          <button type="button" className="link" onClick={() => void drenar()}>
            Reintentar envío
          </button>
        )}
      </div>
      <ul className="evidencia-grid" aria-label="Fotos de la labor">
        {items.map((it) => {
          const rota = rotas.has(it.id)
          const contenido = rota ? (
            <span className="evidencia-doc" aria-hidden="true">
              No disponible
            </span>
          ) : esImagen(it.mime) ? (
            <img
              src={it.estado === "enviada" ? urlArchivo(it.id) : urls.get(it.id)}
              alt={`Foto de evidencia del ${formatFechaHora(it.fecha)}`}
              width={112}
              height={112}
              loading="lazy"
              decoding="async"
              onError={marcarRota(it.id)}
            />
          ) : (
            <span className="evidencia-doc" aria-hidden="true">
              PDF
            </span>
          )
          return (
            <li key={it.id}>
              <figure className="evidencia-mini">
                {it.estado === "enviada" ? (
                  <a href={urlArchivo(it.id)} target="_blank" rel="noopener" aria-label={`Abrir ${it.nombre}`}>
                    {contenido}
                  </a>
                ) : (
                  contenido
                )}
                <figcaption>
                  <span className={`chip ${it.estado}`}>{it.estado === "enviada" ? "Enviada" : "Pendiente"}</span>
                  <small>{it.nombre}</small>
                  {it.estado === "pendiente" && <small>{it.conPunto ? "Con ubicación" : "Sin ubicación"}</small>}
                </figcaption>
              </figure>
            </li>
          )
        })}
      </ul>
      {cargando && <Esqueleto filas={2} />}
      {!cargando && items.length === 0 && (
        <p className="empty">{actividadId ? "Esta labor no tiene fotos." : "La foto se adjunta cuando la labor ya está en el servidor."}</p>
      )}
    </section>
  )
}
