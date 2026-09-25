import { useCallback, useEffect, useState } from "react"
import { comprimirFoto, rechazarPorTamano } from "../offline/comprimir"
import { leerExif } from "../offline/exif"
import {
  encolarEvidencia,
  enviarColaEvidencias,
  listarEvidencias,
  type QueuedEvidencia,
  type ResultadoEnvio,
} from "../offline/queue"
import { fetchEvidencias, type Evidencia } from "../producto"

type Aviso = { tono: "ok" | "pendiente" | "error"; texto: string }

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

async function publicar(item: QueuedEvidencia): Promise<ResultadoEnvio> {
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
  data.set("archivo", new Blob([item.bytes], { type: item.mime }), item.nombre)
  let res: Response
  try {
    res = await fetch("/api/v1/evidencias", { method: "POST", body: data })
  } catch {
    return "despues"
  }
  if (res.status === 409) return "conflicto"
  if (res.ok) return "ok"
  if (res.status >= 500 || res.status === 0) return "despues"
  return "conflicto"
}

export function EvidenciasCampo({ actividadId }: { actividadId: string }) {
  const [enviadas, setEnviadas] = useState<Evidencia[]>([])
  const [pendientes, setPendientes] = useState<QueuedEvidencia[]>([])
  const [aviso, setAviso] = useState<Aviso | null>(null)
  const [ocupado, setOcupado] = useState(false)

  const refrescar = useCallback(async () => {
    const locales = await listarEvidencias().catch(() => [])
    setPendientes(locales.filter((item) => item.actividadId === actividadId))
    if (!actividadId) {
      setEnviadas([])
      return
    }
    try {
      setEnviadas(await fetchEvidencias(actividadId))
    } catch {
      setEnviadas([])
    }
  }, [actividadId])

  const drenar = useCallback(async () => {
    if (!navigator.onLine) return
    const r = await enviarColaEvidencias(publicar)
    if (r.enviadas.length > 0) setAviso({ tono: "ok", texto: "Foto enviada." })
    if (r.conflictos.length > 0) {
      setAviso({ tono: "error", texto: "Esa foto ya estaba registrada con otro contenido." })
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

  async function tomar(file: File) {
    if (!actividadId) return
    setOcupado(true)
    setAviso(null)
    try {
      if (rechazarPorTamano(file.size) && !file.type.startsWith("image/")) {
        setAviso({ tono: "error", texto: "El archivo supera 8 MB." })
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
        exif: { fecha: exif.fecha, lat, lon },
        createdAt: new Date().toISOString(),
      }
      const sinPunto = lat == null || lon == null
      if (!navigator.onLine) {
        await encolarEvidencia(item)
        setAviso({
          tono: "pendiente",
          texto: sinPunto ? "Guardado en este equipo, sin ubicación." : "Guardado en este equipo.",
        })
        await refrescar()
        return
      }
      const resultado = await publicar(item)
      if (resultado === "ok") {
        setAviso({ tono: "ok", texto: sinPunto ? "Foto enviada, sin ubicación." : "Foto enviada." })
      } else if (resultado === "conflicto") {
        setAviso({ tono: "error", texto: "Esa foto ya estaba registrada con otro contenido." })
      } else {
        await encolarEvidencia(item)
        setAviso({
          tono: "pendiente",
          texto: sinPunto ? "Guardado en este equipo, sin ubicación." : "Guardado en este equipo.",
        })
      }
      await refrescar()
    } catch (error) {
      setAviso({ tono: "error", texto: error instanceof Error ? error.message : "No se pudo preparar la foto." })
    } finally {
      setOcupado(false)
    }
  }

  return (
    <section className="evidencia-campo" aria-label="Evidencias de la labor">
      <h3>Evidencia</h3>
      <p className="evidencia-ayuda">La foto se reduce en el teléfono antes de enviarse.</p>
      <div className="evidencia-acciones">
        <label className="primary evidencia-captura">
          {ocupado ? "Preparando…" : "Tomar foto"}
          <input
            type="file"
            accept="image/*"
            capture="environment"
            disabled={ocupado || !actividadId}
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
            disabled={ocupado || !actividadId}
            onChange={(event) => {
              const file = event.target.files?.[0]
              event.target.value = ""
              if (file) void tomar(file)
            }}
          />
        </label>
      </div>
      {aviso && (
        <p className={`status${aviso.tono === "error" ? " error" : ""}`} role="status">
          {aviso.texto}
        </p>
      )}
      <ul className="labor-list evidencia-lista">
        {pendientes.map((item) => (
          <li key={item.id}>
            <strong>{item.nombre}</strong>
            <span className="chip pendiente">Pendiente</span>
            <small>{item.lat == null ? "Sin ubicación" : "Con ubicación"}</small>
          </li>
        ))}
        {enviadas.map((item) => (
          <li key={item.id}>
            <a href={`/api/v1/evidencias/${item.id}/archivo`}>{item.nombre}</a>
            <span className="chip enviada">Enviada</span>
            {item.nota ? <small>{item.nota}</small> : null}
          </li>
        ))}
      </ul>
      {pendientes.length === 0 && enviadas.length === 0 && (
        <p className="empty">{actividadId ? "Esta labor no tiene fotos." : "La foto se adjunta cuando la labor ya está en el servidor."}</p>
      )}
    </section>
  )
}
