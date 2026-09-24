import { useCallback, useEffect, useMemo, useState } from "react"
import { fetchCollection } from "./api"
import { CampusMap } from "./map/CampusMap"
import { enqueue, listQueue, removeQueued, type QueuedLabor } from "./offline/queue"
import {
  ApiError,
  archivar,
  asignar,
  cambiarEstado,
  crearActividad,
  fetchActividades,
  fetchCapataces,
  fetchTimeline,
  lonLat,
  type Capataz,
  type CreateBody,
  type Evento,
} from "./operacion"
import { Labores, type LaborItem } from "./panel/Labores"
import { readEquipo, readRol, writeEquipo, writeRol } from "./session"
import { LAYERS, ROLES, type FeatureCollection, type GeoFeature, type LayerId, type Rol } from "./types"

type LoadState = { kind: "loading" } | { kind: "error"; message: string } | { kind: "ready" }

function prop(feature: GeoFeature, key: string): string {
  const value = feature.properties?.[key]
  return value == null ? "" : String(value)
}

function toItem(feature: GeoFeature, queued = false): LaborItem | null {
  const id = prop(feature, "id") || String(feature.id ?? "")
  if (!id) return null
  return {
    id,
    titulo: prop(feature, "titulo") || "Labor",
    tipo: prop(feature, "tipo"),
    estado: prop(feature, "estado") || "pendiente",
    equipo: prop(feature, "equipo"),
    detalle: prop(feature, "detalle"),
    capatazId: prop(feature, "assigned_capataz_id"),
    queued,
  }
}

function queuedFeature(item: QueuedLabor): GeoFeature {
  return {
    type: "Feature",
    id: item.id,
    geometry: { type: "Point", coordinates: [item.body.lon, item.body.lat] },
    properties: {
      id: item.id,
      tipo: item.body.tipo,
      estado: "pendiente",
      titulo: item.body.titulo,
      detalle: item.body.detalle,
      assigned_capataz_id: item.body.assigned_capataz_id,
      equipo: "",
    },
  }
}

export default function App() {
  const [rol, setRol] = useState<Rol>(() => readRol())
  const [equipoId, setEquipoId] = useState(() => readEquipo())
  const [equipos, setEquipos] = useState<Capataz[]>([])
  const [visible, setVisible] = useState<Record<LayerId, boolean>>(() =>
    Object.fromEntries(LAYERS.map((layer) => [layer.id, layer.defaultOn])) as Record<LayerId, boolean>,
  )
  const [data, setData] = useState<Partial<Record<LayerId, FeatureCollection>>>({})
  const [load, setLoad] = useState<LoadState>({ kind: "loading" })
  const [activities, setActivities] = useState<FeatureCollection>({ type: "FeatureCollection", features: [] })
  const [activityError, setActivityError] = useState("")
  const [queue, setQueue] = useState<QueuedLabor[]>([])
  const [estados, setEstados] = useState<Record<string, boolean>>({ pendiente: true, en_proceso: true, bloqueada: true })
  const [tipoFiltro, setTipoFiltro] = useState("")
  const [railOpen, setRailOpen] = useState(true)
  const [picked, setPicked] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [pinMode, setPinMode] = useState(false)
  const [draft, setDraft] = useState<{ lon: number; lat: number } | null>(null)
  const [focus, setFocus] = useState<{ lon: number; lat: number; token: number } | null>(null)
  const [formTipo, setFormTipo] = useState("riego")
  const [formTitulo, setFormTitulo] = useState("")
  const [formDetalle, setFormDetalle] = useState("")
  const [formEquipo, setFormEquipo] = useState("cap-norte")
  const [creating, setCreating] = useState(false)
  const [notice, setNotice] = useState("")
  const [timeline, setTimeline] = useState<Evento[]>([])
  const [timelineError, setTimelineError] = useState("")
  const [estadoNuevo, setEstadoNuevo] = useState("pendiente")
  const [reasignarA, setReasignarA] = useState("cap-norte")
  const [confirmarArchivo, setConfirmarArchivo] = useState(false)

  const reloadActivities = useCallback(async () => {
    try {
      const fc = await fetchActividades(rol, equipoId)
      setActivities(fc)
      setActivityError("")
    } catch (error) {
      const message = error instanceof Error ? error.message : "No se pudieron leer las labores"
      setActivityError(message)
    }
  }, [rol, equipoId])

  const reloadQueue = useCallback(async () => {
    try {
      setQueue(await listQueue())
    } catch {
      setQueue([])
    }
  }, [])

  const flush = useCallback(async () => {
    const pending = await listQueue().catch(() => [] as QueuedLabor[])
    for (const item of pending) {
      try {
        await crearActividad(item.body)
        await removeQueued(item.id)
      } catch (error) {
        if (error instanceof ApiError && error.status === 409) {
          await removeQueued(item.id)
          setNotice("Una labor en cola ya existía con otro contenido y se descartó.")
        } else if (!(error instanceof ApiError) || error.status !== 0) {
          await removeQueued(item.id)
        } else {
          break
        }
      }
    }
    await reloadQueue()
    await reloadActivities()
  }, [reloadActivities, reloadQueue])

  useEffect(() => {
    let cancelled = false
    Promise.all(LAYERS.map(async (layer) => [layer.id, await fetchCollection(layer.path)] as const))
      .then((rows) => {
        if (cancelled) return
        setData(Object.fromEntries(rows) as Partial<Record<LayerId, FeatureCollection>>)
        setLoad({ kind: "ready" })
      })
      .catch((error: unknown) => {
        if (cancelled) return
        const message = error instanceof Error ? error.message : "No se pudo leer el catastro"
        setLoad({ kind: "error", message })
      })
    fetchCapataces()
      .then((rows) => {
        if (!cancelled) setEquipos(rows)
      })
      .catch(() => {
        if (!cancelled) setEquipos([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    void reloadActivities()
  }, [reloadActivities])

  useEffect(() => {
    void flush()
  }, [flush])

  useEffect(() => {
    if (!selectedId || queue.some((item) => item.id === selectedId)) {
      setTimeline([])
      return
    }
    let cancelled = false
    fetchTimeline(selectedId)
      .then((rows) => {
        if (!cancelled) {
          setTimeline(rows)
          setTimelineError("")
        }
      })
      .catch((error: unknown) => {
        if (!cancelled) setTimelineError(error instanceof Error ? error.message : "Sin bitácora")
      })
    return () => {
      cancelled = true
    }
  }, [selectedId, queue, activities])

  const items = useMemo(() => {
    const fromApi = activities.features
      .map((feature) => toItem(feature))
      .filter((item): item is LaborItem => item != null)
    const local = queue
      .filter((item) => rol !== "capataz" || item.body.assigned_capataz_id === equipoId)
      .filter((item) => !fromApi.some((existing) => existing.id === item.id))
      .map((item) => toItem(queuedFeature(item), true))
      .filter((item): item is LaborItem => item != null)
    return [...local, ...fromApi].filter((item) => {
      if (estados[item.estado] === false) return false
      if (tipoFiltro && item.tipo !== tipoFiltro) return false
      return true
    })
  }, [activities, queue, estados, tipoFiltro, rol, equipoId])

  const mapActivities = useMemo<FeatureCollection>(() => {
    const ids = new Set(items.map((item) => item.id))
    const features = [
      ...queue.map(queuedFeature),
      ...activities.features,
    ].filter((feature) => ids.has(prop(feature, "id") || String(feature.id ?? "")))
    const seen = new Set<string>()
    return {
      type: "FeatureCollection",
      features: features.filter((feature) => {
        const id = prop(feature, "id") || String(feature.id ?? "")
        if (seen.has(id)) return false
        seen.add(id)
        return true
      }),
    }
  }, [items, queue, activities])

  const selected = items.find((item) => item.id === selectedId) ?? null
  const role = ROLES.find((item) => item.id === rol) ?? ROLES[1]
  const summary = useMemo(() => {
    const areas = data.areas?.features.length
    const zonas = data.zonas?.features.length
    if (areas == null || zonas == null) return "Leyendo catastro…"
    return `${areas} áreas · ${zonas} zonas · ${items.length} labores`
  }, [data, items.length])

  function choose(id: string) {
    setSelectedId(id)
    setConfirmarArchivo(false)
    const feature = activities.features.find((item) => prop(item, "id") === id) ?? queue.map(queuedFeature).find((item) => item.id === id)
    const point = feature ? lonLat(feature) : null
    if (point) setFocus({ ...point, token: Date.now() })
    const item = items.find((row) => row.id === id)
    if (item) {
      setEstadoNuevo(item.estado)
      setReasignarA(item.capatazId || equipos[0]?.id || "cap-norte")
    }
  }

  async function onCreate() {
    if (!draft || !formTitulo.trim()) {
      setNotice("Indique un título y un punto en el mapa.")
      return
    }
    const body: CreateBody = {
      id: crypto.randomUUID(),
      tipo: formTipo,
      titulo: formTitulo.trim(),
      detalle: formDetalle.trim(),
      lon: draft.lon,
      lat: draft.lat,
      assigned_capataz_id: formEquipo,
      actor_rol: rol,
    }
    setCreating(true)
    setNotice("")
    try {
      await crearActividad(body)
      setDraft(null)
      setPinMode(false)
      setFormTitulo("")
      setFormDetalle("")
      setNotice("Labor creada.")
      await reloadActivities()
      setSelectedId(body.id)
    } catch (error) {
      if (error instanceof ApiError && error.status === 0) {
        await enqueue({ id: body.id, body, createdAt: new Date().toISOString() })
        await reloadQueue()
        setNotice("Sin conexión: la labor quedó en la cola de este navegador.")
        setDraft(null)
        setPinMode(false)
      } else {
        setNotice(error instanceof Error ? error.message : "No se pudo crear")
      }
    } finally {
      setCreating(false)
    }
  }

  async function onEstado() {
    if (!selected || selected.queued) return
    try {
      await cambiarEstado(selected.id, estadoNuevo, rol, equipoId)
      setNotice("Estado actualizado.")
      await reloadActivities()
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "No se pudo cambiar el estado")
    }
  }

  async function onReasignar() {
    if (!selected || selected.queued) return
    try {
      await asignar(selected.id, reasignarA, rol)
      setNotice("Asignación registrada.")
      await reloadActivities()
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "No se pudo reasignar")
    }
  }

  async function onArchivar() {
    if (!selected || selected.queued) return
    if (!confirmarArchivo) {
      setConfirmarArchivo(true)
      return
    }
    try {
      await archivar(selected.id, rol)
      setSelectedId(null)
      setNotice("Labor archivada. Ya no aparece en el mapa abierto.")
      await reloadActivities()
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "No se pudo archivar")
    }
  }

  return (
    <div className="shell">
      <CampusMap
        data={data}
        visible={visible}
        activities={mapActivities}
        pinMode={pinMode && rol !== "capataz"}
        draft={draft}
        focus={focus}
        onSelectCatastro={(hit) => setPicked(hit ? `${hit.layer}: ${String(hit.props.nombre || hit.props.feature_id || "polígono")}` : null)}
        onSelectActividad={(id) => {
          if (id) choose(id)
          else setSelectedId(null)
        }}
        onPin={(lon, lat) => {
          setDraft({ lon, lat })
          setNotice("")
        }}
      />
      <header className="topbar">
        <div className="brand">
          <strong>Campus Verde</strong>
          <span>PUCP Pando · supervisión</span>
        </div>
        <button type="button" className="menu-btn" onClick={() => setRailOpen((open) => !open)}>
          Panel
        </button>
        <div className="top-spacer" />
        <div className="roles" role="group" aria-label="Rol de consulta">
          {ROLES.map((item) => (
            <button
              key={item.id}
              type="button"
              aria-pressed={item.id === rol}
              onClick={() => {
                setRol(item.id)
                writeRol(item.id)
                setPinMode(false)
                setDraft(null)
                setSelectedId(null)
              }}
            >
              {item.label}
            </button>
          ))}
        </div>
      </header>
      <aside className={railOpen ? "rail open" : "rail"}>
        <p className="role-note">{role.note}</p>
        <p className={load.kind === "error" || activityError ? "status error" : "status"}>
          {load.kind === "error" ? load.message : summary}
          {activityError ? ` · ${activityError}` : ""}
          {picked ? ` · ${picked}` : ""}
        </p>
        <Labores
          rol={rol}
          equipos={equipos}
          equipoId={equipoId}
          onEquipo={(id) => {
            setEquipoId(id)
            writeEquipo(id)
            setSelectedId(null)
          }}
          items={items}
          estados={estados}
          onToggleEstado={(id) => setEstados((current) => ({ ...current, [id]: current[id] === false }))}
          tipo={tipoFiltro}
          onTipo={setTipoFiltro}
          pinMode={pinMode}
          onPinMode={(on) => {
            setPinMode(on)
            if (!on) setDraft(null)
          }}
          draft={draft}
          formTipo={formTipo}
          formTitulo={formTitulo}
          formDetalle={formDetalle}
          formEquipo={formEquipo}
          onForm={(patch) => {
            if (patch.tipo) setFormTipo(patch.tipo)
            if (patch.titulo != null) setFormTitulo(patch.titulo)
            if (patch.detalle != null) setFormDetalle(patch.detalle)
            if (patch.equipo != null) setFormEquipo(patch.equipo)
          }}
          onCreate={() => void onCreate()}
          creating={creating}
          selected={selected}
          onSelect={choose}
          timeline={timeline}
          timelineError={timelineError}
          estadoNuevo={estadoNuevo}
          onEstadoNuevo={setEstadoNuevo}
          onEstado={() => void onEstado()}
          reasignarA={reasignarA}
          onReasignarA={setReasignarA}
          onReasignar={() => void onReasignar()}
          onArchivar={() => void onArchivar()}
          confirmarArchivo={confirmarArchivo}
          notice={notice}
          queueCount={queue.length}
          onFlush={() => void flush()}
        />
        <h2>Catastro</h2>
        {LAYERS.map((layer) => (
          <div className="layer" key={layer.id}>
            <span className="swatch" style={{ background: layer.fill }} />
            <label>
              <input
                type="checkbox"
                checked={visible[layer.id]}
                onChange={() => setVisible((current) => ({ ...current, [layer.id]: !current[layer.id] }))}
              />{" "}
              {layer.label}
              <small>{layer.hint}</small>
            </label>
            <span className="count">{data[layer.id]?.features.length ?? "—"}</span>
          </div>
        ))}
        <p className="foot">Sin Street View. El rol es local: no hay sesión institucional.</p>
      </aside>
    </div>
  )
}
