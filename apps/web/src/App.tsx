import { useCallback, useEffect, useMemo, useState } from "react"
import { fetchCollection } from "./api"
import { INVENTARIO } from "./inventario"
import { CampusMap } from "./map/CampusMap"
import { MapBoundary } from "./map/MapBoundary"
import { enqueue, enqueueEstado, listEstados, listQueue, loadLabores, removeEstado, removeQueued, saveLabores, type QueuedLabor } from "./offline/queue"
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
  TIPOS,
  type Capataz,
  type CreateBody,
  type Evento,
} from "./operacion"
import { Labores, type LaborItem } from "./panel/Labores"
import { AdminPanel, CatalogosPanel, CatastroPanel, Login, ReportesPanel, RiegoPanel, SolicitudesPanel } from "./panel/Modulos"
import { fetchCatalogo, fetchSesion, salir, sugerirTipo, type CatalogoItem, type Usuario } from "./producto"
import { readEquipo, writeEquipo } from "./session"
import { LAYERS, type FeatureCollection, type GeoFeature, type LayerId, type Rol } from "./types"

type LoadState = { kind: "loading" } | { kind: "error"; message: string } | { kind: "ready" }

type Modulo = "mapa" | "labores" | "catastro" | "solicitudes" | "reportes" | "catalogos" | "admin"

const MODULOS: { id: Modulo; label: string }[] = [
  { id: "mapa", label: "Mapa" },
  { id: "labores", label: "Labores" },
  { id: "catastro", label: "Catastro" },
  { id: "solicitudes", label: "Solicitudes" },
  { id: "reportes", label: "Reportes" },
  { id: "catalogos", label: "Catálogos" },
  { id: "admin", label: "Admin" },
]

function modulosDe(rol: Rol): Modulo[] {
  if (rol === "capataz") return ["mapa", "labores", "catastro"]
  if (rol === "jefatura") return ["mapa", "labores", "catastro", "solicitudes", "reportes"]
  if (rol === "admin") return MODULOS.map((item) => item.id)
  return ["mapa", "labores", "catastro", "solicitudes", "reportes", "catalogos"]
}

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
    ejecutor: prop(feature, "ejecutor"),
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
  const [sesion, setSesion] = useState<Usuario | null>(null)
  const [sesionLista, setSesionLista] = useState(false)
  const [modulo, setModulo] = useState<Modulo>("mapa")
  const rol = (sesion?.rol ?? "coordinacion") as Rol
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
  const [railOpen, setRailOpen] = useState(() => window.matchMedia("(min-width: 821px)").matches)
  const [picked, setPicked] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [pinMode, setPinMode] = useState(false)
  const [draft, setDraft] = useState<{ lon: number; lat: number } | null>(null)
  const [focus, setFocus] = useState<{ lon: number; lat: number; token: number } | null>(null)
  const [formTipo, setFormTipo] = useState("riego")
  const [formTitulo, setFormTitulo] = useState("")
  const [formDetalle, setFormDetalle] = useState("")
  const [formEquipo, setFormEquipo] = useState("cap-norte")
  const [formEjecutor, setFormEjecutor] = useState("propia")
  const [motivo, setMotivo] = useState("")
  const [tiposCat, setTiposCat] = useState<CatalogoItem[]>([])
  const [motivos, setMotivos] = useState<CatalogoItem[]>([])
  const [sugerencia, setSugerencia] = useState("")
  const [creating, setCreating] = useState(false)
  const [notice, setNotice] = useState("")
  const [timeline, setTimeline] = useState<Evento[]>([])
  const [timelineError, setTimelineError] = useState("")
  const [estadoNuevo, setEstadoNuevo] = useState("pendiente")
  const [reasignarA, setReasignarA] = useState("cap-norte")
  const [confirmarArchivo, setConfirmarArchivo] = useState(false)
  const [relieve, setRelieve] = useState(false)
  const [showEdificios, setShowEdificios] = useState(false)
  const [edificios, setEdificios] = useState<FeatureCollection>({ type: "FeatureCollection", features: [] })
  const [inventory, setInventory] = useState<Partial<Record<string, FeatureCollection>>>({})
  const [inventoryOn, setInventoryOn] = useState<Record<string, boolean>>({})
  const [agenda, setAgenda] = useState<{ aviso: string; total: number; reservas: { id: string; jardin: string; fecha: string; hora: string; evento: string; estado: string }[] } | null>(null)

  const reloadActivities = useCallback(async () => {
    try {
      const fc = await fetchActividades(rol, equipoId)
      setActivities(fc)
      void saveLabores(fc)
      setActivityError("")
    } catch (error) {
      const cached = await loadLabores<FeatureCollection>().catch(() => null)
      if (cached && Array.isArray(cached.features)) {
        setActivities(cached)
        setActivityError("Sin conexión: se muestra la última lista guardada en este navegador.")
      } else {
        const message = error instanceof Error ? error.message : "No se pudieron leer las labores"
        setActivityError(message)
      }
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
    const estadosPendientes = await listEstados().catch(() => [])
    for (const item of estadosPendientes) {
      try {
        await cambiarEstado(item.actividadId, item.estado, rol, equipoId)
        await removeEstado(item.id)
      } catch (error) {
        if (!(error instanceof ApiError) || error.status !== 0) {
          await removeEstado(item.id)
        } else {
          break
        }
      }
    }
    await reloadQueue()
    await reloadActivities()
  }, [reloadActivities, reloadQueue, rol, equipoId])

  useEffect(() => {
    let cancelled = false
    fetchSesion()
      .then((user) => {
        if (cancelled) return
        setSesion(user)
        if (user?.rol === "capataz" && user.capataz_id) {
          setEquipoId(user.capataz_id)
          writeEquipo(user.capataz_id)
        }
      })
      .catch(() => {
        if (!cancelled) setSesion(null)
      })
      .finally(() => {
        if (!cancelled) setSesionLista(true)
      })
    return () => {
      cancelled = true
    }
  }, [])

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
    Promise.all(
      INVENTARIO.map(async (layer) => {
        try {
          return [layer.id, await fetchCollection(`/api/v1/geo/inventario/${layer.id}`)] as const
        } catch {
          return [layer.id, { type: "FeatureCollection" as const, features: [] }] as const
        }
      }),
    ).then((rows) => {
      if (!cancelled) setInventory(Object.fromEntries(rows))
    })
    fetch("/api/v1/geo/reservas-mock")
      .then((res) => (res.ok ? res.json() : null))
      .then((body) => {
        if (!cancelled && body && Array.isArray(body.reservas)) setAgenda(body)
      })
      .catch(() => {
        if (!cancelled) setAgenda(null)
      })
    fetchCollection("/api/v1/geo/edificios")
      .then((fc) => {
        if (!cancelled) setEdificios(fc)
      })
      .catch(() => {
        if (!cancelled) setEdificios({ type: "FeatureCollection", features: [] })
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
    let cancelled = false
    fetchActividades(rol, equipoId)
      .then((fc) => {
        if (cancelled) return
        setActivities(fc)
        void saveLabores(fc)
        setActivityError("")
      })
      .catch(async (error: unknown) => {
        if (cancelled) return
        const cached = await loadLabores<FeatureCollection>().catch(() => null)
        if (cached && Array.isArray(cached.features)) {
          setActivities(cached)
          setActivityError("Sin conexión: se muestra la última lista guardada en este navegador.")
          return
        }
        setActivityError(error instanceof Error ? error.message : "No se pudieron leer las labores")
      })
    return () => {
      cancelled = true
    }
  }, [rol, equipoId])

  useEffect(() => {
    let cancelled = false
    void Promise.resolve().then(() => {
      if (!cancelled) return flush()
    })
    return () => {
      cancelled = true
    }
  }, [flush])

  const seleccionEnCola = selectedId != null && queue.some((item) => item.id === selectedId)
  const timelineVisible = selectedId && !seleccionEnCola ? timeline : []
  const timelineErrorVisible = selectedId && !seleccionEnCola ? timelineError : ""

  useEffect(() => {
    if (!selectedId || queue.some((item) => item.id === selectedId)) return
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

  useEffect(() => {
    if (!sesion) return
    let cancelled = false
    fetchCatalogo("tipo_actividad", true)
      .then((rows) => {
        if (!cancelled) setTiposCat(rows)
      })
      .catch(() => {
        if (!cancelled) setTiposCat([])
      })
    fetchCatalogo("motivo_archivo", true)
      .then((rows) => {
        if (!cancelled) setMotivos(rows)
      })
      .catch(() => {
        if (!cancelled) setMotivos([])
      })
    return () => {
      cancelled = true
    }
  }, [sesion])

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
      ejecutor: formEjecutor,
    }
    setCreating(true)
    setNotice("")
    try {
      await crearActividad(body)
      setDraft(null)
      setFormTitulo("")
      setFormDetalle("")
      setSugerencia("")
      setNotice("Labor creada. Marque el siguiente punto.")
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
      if (error instanceof ApiError && error.status === 0) {
        await enqueueEstado({
          id: crypto.randomUUID(),
          actividadId: selected.id,
          estado: estadoNuevo,
          createdAt: new Date().toISOString(),
        })
        setNotice("Sin conexión: el cambio de estado quedó en la cola.")
      } else {
        setNotice(error instanceof Error ? error.message : "No se pudo cambiar el estado")
      }
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
    if (!motivo) {
      setNotice("Elija un motivo de archivo.")
      return
    }
    try {
      await archivar(selected.id, rol, motivo)
      setSelectedId(null)
      setNotice("Labor archivada. Ya no aparece en el mapa abierto.")
      await reloadActivities()
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "No se pudo archivar")
    }
  }

  if (!sesionLista) {
    return (
      <main className="gate" aria-busy="true">
        <section className="gate-brand">
          <img className="gate-logo" src="/logo-verdepucp.png" alt="VerdePUCP" />
        </section>
        <section className="gate-form">
          <div className="skel-wrap">
            <div className="skel" />
            <div className="skel" />
          </div>
        </section>
      </main>
    )
  }
  if (!sesion) return <Login onIn={setSesion} />

  const permitidos = modulosDe(rol)
  const moduloActivo = permitidos.includes(modulo) ? modulo : "mapa"
  const tipos = tiposCat.length > 0 ? tiposCat.map((item) => ({ id: item.codigo, label: item.nombre })) : TIPOS.map((item) => ({ id: item.id, label: item.label }))

  return (
    <div className={railOpen ? "shell" : "shell panel-off"}>
      <a className="skip" href="#panel">Saltar al panel</a>
      <header className="topbar">
        <div className="brand">
          <img className="brand-mark" src="/isotipo.svg" alt="" />
          <div>
            <strong className="word"><span className="wv">Verde</span><span className="wp">PUCP</span></strong>
            <span className="brand-sub">Gestión de Áreas Verdes</span>
          </div>
        </div>
        <button type="button" className="menu-btn" onClick={() => setRailOpen((open) => !open)}>
          {railOpen ? "Ocultar" : "Panel"}
        </button>
        <div className="top-spacer" />
        <div className="roles" role="group" aria-label="Vista del mapa">
          <button type="button" aria-pressed={!relieve} onClick={() => setRelieve(false)}>Plano</button>
          <button type="button" aria-pressed={relieve} onClick={() => setRelieve(true)}>Relieve</button>
        </div>
        <div className="session">
          <span>{sesion.nombre}</span>
          <button type="button" onClick={() => void salir().then(() => setSesion(null))}>Salir</button>
        </div>
      </header>
      <nav className="guard" aria-label="Módulos">
        <div className="rail-mark"><img src="/isotipo.svg" alt="VerdePUCP" /></div>
        {MODULOS.filter((item) => permitidos.includes(item.id)).map((item) => (
          <button key={item.id} type="button" aria-pressed={moduloActivo === item.id} onClick={() => { setModulo(item.id); setRailOpen(true) }}>
            {item.label}
          </button>
        ))}
      </nav>
      <aside className="panel" id="panel">
        <button type="button" className="sheet-close" onClick={() => setRailOpen(false)}>
          Cerrar hoja
        </button>
        <div key={moduloActivo} className="panel-view">
        <p className={load.kind === "error" || activityError ? "status error" : "status"}>
          {load.kind === "error" ? load.message : summary}
          {activityError ? ` · ${activityError}` : ""}
          {picked ? ` · ${picked}` : ""}
        </p>
        {moduloActivo === "mapa" && (
          <section className="block">
            <h2>Capas</h2>
            <div className="layer">
              <span className="swatch" style={{ background: "#c8c0b2" }} />
              <label>
                <input type="checkbox" checked={showEdificios} onChange={() => setShowEdificios((on) => !on)} /> Edificios OSM
                <small>Huellas del recinto, solo en relieve</small>
              </label>
              <span className="count">{edificios.features.length || "—"}</span>
            </div>
            {LAYERS.map((layer) => (
              <div className="layer" key={layer.id}>
                <span className="swatch" style={{ background: layer.fill }} />
                <label>
                  <input type="checkbox" checked={visible[layer.id]} onChange={() => setVisible((current) => ({ ...current, [layer.id]: !current[layer.id] }))} />{" "}
                  {layer.label}
                  <small>{layer.hint}</small>
                </label>
                <span className="count">{data[layer.id]?.features.length ?? "—"}</span>
              </div>
            ))}
            <h2>Inventario</h2>
            <p className="lede">Capas opcionales. Apagadas hasta que se necesiten.</p>
            {INVENTARIO.map((layer) => (
              <div className="layer" key={layer.id}>
                <span className="swatch" style={{ background: layer.color }} />
                <label>
                  <input type="checkbox" checked={inventoryOn[layer.id] === true} onChange={() => setInventoryOn((current) => ({ ...current, [layer.id]: !current[layer.id] }))} />{" "}
                  {layer.label}
                  <small>{layer.hint}</small>
                </label>
                <span className="count">{inventory[layer.id]?.features.length ?? "—"}</span>
              </div>
            ))}
            <h2>Agenda ficticia</h2>
            <p className="lede">{agenda?.aviso ?? "Leyendo la agenda de demostración…"}</p>
            {(agenda?.reservas ?? []).length === 0 && <p className="empty">No hay reservas de demostración.</p>}
            <ul className="labor-list">
              {(agenda?.reservas ?? []).slice(0, 6).map((item) => (
                <li key={item.id} className="agenda">
                  <strong>{item.jardin}</strong>
                  <small>{item.fecha} · {item.hora} · {item.evento}</small>
                </li>
              ))}
            </ul>
          </section>
        )}
        {moduloActivo === "labores" && (
          <>
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
                if (patch.ejecutor) setFormEjecutor(patch.ejecutor)
              }}
              onCreate={() => void onCreate()}
              creating={creating}
              selected={selected}
              onSelect={choose}
              timeline={timelineVisible}
              timelineError={timelineErrorVisible}
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
              tipos={tipos}
              formEjecutor={formEjecutor}
              motivos={motivos}
              motivo={motivo}
              onMotivo={setMotivo}
              onSugerir={() => {
                void sugerirTipo(formTitulo)
                  .then((s) => {
                    setSugerencia(s.explicacion)
                    if (s.codigo) setFormTipo(s.codigo)
                  })
                  .catch((error: unknown) => setSugerencia(error instanceof Error ? error.message : "Sin sugerencia"))
              }}
              sugerencia={sugerencia}
            />
            <RiegoPanel capatazId={rol === "capataz" ? equipoId : formEquipo} />
          </>
        )}
        {moduloActivo === "catastro" && <CatastroPanel />}
        {moduloActivo === "solicitudes" && <SolicitudesPanel actividadId={selected?.queued ? "" : selected?.id ?? ""} />}
        {moduloActivo === "reportes" && <ReportesPanel />}
        {moduloActivo === "catalogos" && <CatalogosPanel editable={rol === "admin"} />}
        {moduloActivo === "admin" && <AdminPanel />}
        </div>
      </aside>
      <div className="stage">
        <MapBoundary>
          <CampusMap
            data={data}
            visible={visible}
            activities={mapActivities}
            pinMode={pinMode && rol !== "capataz" && moduloActivo === "labores"}
            draft={draft}
            focus={focus}
            relieve={relieve}
            edificios={edificios}
            showEdificios={showEdificios}
            inventory={inventory}
            inventoryOn={inventoryOn}
            onSelectCatastro={(hit) => setPicked(hit ? `${hit.layer}: ${String(hit.props.nombre || hit.props.feature_id || "polígono")}` : null)}
            onSelectActividad={(id) => {
              if (id) {
                choose(id)
                setModulo("labores")
                setRailOpen(true)
              } else setSelectedId(null)
            }}
            onPin={(lon, lat) => {
              setDraft({ lon, lat })
              setNotice("")
            }}
          />
        </MapBoundary>
      </div>
    </div>
  )
}
