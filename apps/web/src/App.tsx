import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { apiUrl, fetchCollection } from "./api"
import { INVENTARIO } from "./inventario"
import { CampusMap } from "./map/CampusMap"
import { seleccionarActividad } from "./map/seleccionActividad"
import type { ModoDibujo } from "./map/draw"
import { MapBoundary } from "./map/MapBoundary"
import { useEngancharCola } from "./offline/enganchar"
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
  puedeEncolarEstado,
  TIPOS,
  type Capataz,
  type CreateBody,
  type Evento,
} from "./operacion"
import { Labores, type LaborItem } from "./panel/Labores"
import { PodaPanel } from "./panel/Poda"
import { ViveroPanel } from "./panel/Vivero"
import { CatastroEditor } from "./panel/CatastroEditor"
import { InventarioCapas } from "./panel/InventarioCapas"
import { CalendarioReservas } from "./panel/CalendarioReservas"
import { conCategorias, conteoPorCategoria, sectoresDesdeCatalogo, SIN_SECTOR, USOS, type Categoria, type ColorPor } from "./map/categorias"
import { ACTIVIDAD, CAPA, MAPA, ZONIFICACION, conteoCatastro, resumenCatastro } from "./ui/nomenclatura"
import { BottomSheet } from "./ui/BottomSheet"
import { ImportacionesPanel } from "./panel/Importaciones"
import { AdminPanel } from "./panel/Admin"
import { CatalogosPanel } from "./panel/Catalogos"
import { Login } from "./panel/Login"
import { ReportesPanel } from "./panel/Reportes"
import { RiegoPanel } from "./panel/Riego"
import { SolicitudesPanel } from "./panel/Solicitudes"
import { etiquetaRol, fetchCatalogo, fetchSesion, salir, sugerirTipo, type CatalogoItem, type Usuario } from "./producto"
import { readColorPor, readEquipo, writeColorPor, writeEquipo } from "./session"
import { listarSectores } from "./panel/zonificacion"
import { LAYERS, type FeatureCollection, type GeoFeature, type LayerId, type Rol } from "./types"
import { MQ_MOVIL, useMedia } from "./ui/media"
import { repartirModulos } from "./ui/navegacion"
import { MODULOS, modulosDe, type Modulo } from "./ui/registroModulos"

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
    titulo: prop(feature, "titulo") || ACTIVIDAD.sinTitulo,
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
  const [colorPor, setColorPor] = useState<ColorPor>(() => readColorPor())
  const [sectoresCat, setSectoresCat] = useState<Categoria<string>[]>(() => [SIN_SECTOR])
  const [catastroOn, setCatastroOn] = useState(true)
  const [ocultas, setOcultas] = useState<Record<ColorPor, string[]>>({ uso: [], sector: [] })
  const [data, setData] = useState<Partial<Record<LayerId, FeatureCollection>>>({})
  const [load, setLoad] = useState<LoadState>({ kind: "loading" })
  const [activities, setActivities] = useState<FeatureCollection>({ type: "FeatureCollection", features: [] })
  const [activityError, setActivityError] = useState("")
  const [queue, setQueue] = useState<QueuedLabor[]>([])
  const [laboresListas, setLaboresListas] = useState(false)
  const [estados, setEstados] = useState<Record<string, boolean>>({ pendiente: true, en_proceso: true, bloqueada: true })
  const [tipoFiltro, setTipoFiltro] = useState("")
  const [railOpen, setRailOpen] = useState(() => !window.matchMedia(MQ_MOVIL).matches)
  const [picked, setPicked] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [pinMode, setPinMode] = useState(false)
  const [modoDibujo, setModoDibujo] = useState<ModoDibujo | null>(null)
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
  const [pista, setPista] = useState<{ codigo: string; etiqueta: string } | null>(null)
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
  const reloadActivities = useCallback(async () => {
    if (!sesion) return
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
        const message = error instanceof Error ? error.message : ACTIVIDAD.noLeer
        setActivityError(message)
      }
    }
  }, [sesion, rol, equipoId])

  const reloadQueue = useCallback(async () => {
    try {
      setQueue(await listQueue())
    } catch {
      setQueue([])
    }
  }, [])

  const flush = useCallback(async () => {
    if (!sesion) return
    const pending = await listQueue().catch(() => [] as QueuedLabor[])
    for (const item of pending) {
      try {
        await crearActividad(item.body)
        await removeQueued(item.id)
      } catch (error) {
        if (error instanceof ApiError && error.status === 409) {
          await removeQueued(item.id)
          setNotice(ACTIVIDAD.duplicada)
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
  }, [sesion, reloadActivities, reloadQueue, rol, equipoId])

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
    if (!sesion) return
    let cancelled = false
    listarSectores(true)
      .then((filas) => {
        if (!cancelled) setSectoresCat(sectoresDesdeCatalogo(filas))
      })
      .catch(() => {
        if (!cancelled) setSectoresCat([SIN_SECTOR])
      })
    return () => {
      cancelled = true
    }
  }, [sesion])

  useEffect(() => {
    if (!sesion) return
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
          return [layer.id, await fetchCollection(apiUrl(`/geo/inventario/${layer.id}`))] as const
        } catch {
          return [layer.id, { type: "FeatureCollection" as const, features: [] }] as const
        }
      }),
    ).then((rows) => {
      if (!cancelled) setInventory(Object.fromEntries(rows))
    })
    fetchCollection(apiUrl("/geo/edificios"))
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
  }, [sesion])

  useEffect(() => {
    if (!sesion) return
    let cancelled = false
    fetchActividades(rol, equipoId)
      .then((fc) => {
        if (cancelled) return
        setActivities(fc)
        void saveLabores(fc)
        setActivityError("")
        setLaboresListas(true)
      })
      .catch(async (error: unknown) => {
        if (cancelled) return
        const cached = await loadLabores<FeatureCollection>().catch(() => null)
        if (cached && Array.isArray(cached.features)) {
          setActivities(cached)
          setActivityError("Sin conexión: se muestra la última lista guardada en este navegador.")
          setLaboresListas(true)
          return
        }
        setActivityError(error instanceof Error ? error.message : ACTIVIDAD.noLeer)
        setLaboresListas(true)
      })
    return () => {
      cancelled = true
    }
  }, [sesion, rol, equipoId])

  useEngancharCola(sesion, flush)

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

  useEffect(() => {
    writeColorPor(colorPor)
  }, [colorPor])

  const dataMapa = useMemo<Partial<Record<LayerId, FeatureCollection>>>(
    () => ({ ...data, areas: conCategorias(data.areas, "uso"), zonas: conCategorias(data.zonas, "sector", sectoresCat) }),
    [data, sectoresCat],
  )
  const visibleMapa = useMemo<Record<LayerId, boolean>>(
    () => ({ ...visible, areas: catastroOn && colorPor === "uso", zonas: catastroOn && colorPor === "sector" }),
    [visible, catastroOn, colorPor],
  )
  const conteoUso = useMemo(() => conteoPorCategoria(dataMapa.areas, "cat_uso", USOS), [dataMapa.areas])
  const conteoSector = useMemo(() => conteoPorCategoria(dataMapa.zonas, "cat_sector", sectoresCat), [dataMapa.zonas, sectoresCat])
  const catsColor = colorPor === "uso" ? USOS : sectoresCat
  const conteoColor: Record<string, number> = colorPor === "uso" ? conteoUso : conteoSector

  const selected = items.find((item) => item.id === selectedId) ?? null
  const summary = useMemo(() => {
    const areas = data.areas?.features.length
    const zonas = data.zonas?.features.length
    if (areas == null || zonas == null) return MAPA.leyendo
    return resumenCatastro(areas, zonas, items.length)
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
      setNotice(ACTIVIDAD.creada)
      await reloadActivities()
      setSelectedId(body.id)
    } catch (error) {
      if (error instanceof ApiError && error.status === 0) {
        await enqueue({ id: body.id, body, createdAt: new Date().toISOString() })
        await reloadQueue()
        setNotice(ACTIVIDAD.enCola)
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
    if (!puedeEncolarEstado(rol, estadoNuevo)) {
      setNotice(ACTIVIDAD.noCierra)
      return
    }
    try {
      await cambiarEstado(selected.id, estadoNuevo, rol, equipoId)
      setNotice("Estado actualizado.")
      await reloadActivities()
    } catch (error) {
      if (error instanceof ApiError && error.status === 0) {
        try {
          await enqueueEstado(
            {
              id: crypto.randomUUID(),
              actividadId: selected.id,
              estado: estadoNuevo,
              createdAt: new Date().toISOString(),
            },
            rol,
          )
          setNotice("Sin conexión: el cambio de estado quedó en la cola.")
        } catch (err) {
          setNotice(err instanceof Error ? err.message : "No se pudo encolar el estado")
        }
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
      setNotice(ACTIVIDAD.archivada)
      await reloadActivities()
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "No se pudo archivar")
    }
  }

  const stageRef = useRef<HTMLDivElement>(null)
  const publicarAltura = useCallback((px: number) => {
    if (px > 0) stageRef.current?.style.setProperty("--sheet-h", `${Math.round(px)}px`)
    else stageRef.current?.style.removeProperty("--sheet-h")
  }, [])
  const movil = useMedia(MQ_MOVIL)
  const [masAbierto, setMasAbierto] = useState(false)
  useEffect(() => {
    if (!masAbierto) return
    document.querySelector<HTMLButtonElement>(".guard-menu button")?.focus({ preventScroll: true })
    const cerrar = (event: PointerEvent) => {
      if (!(event.target instanceof Element) || !event.target.closest(".guard-menu, .guard-mas")) setMasAbierto(false)
    }
    document.addEventListener("pointerdown", cerrar)
    return () => document.removeEventListener("pointerdown", cerrar)
  }, [masAbierto])

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
  const visiblesModulos = MODULOS.filter((item) => permitidos.includes(item.id))
  const { barra, mas } = movil ? repartirModulos(visiblesModulos) : { barra: visiblesModulos, mas: [] as typeof visiblesModulos }
  const enMas = mas.find((item) => item.id === moduloActivo)
  const menuMas = masAbierto && movil && mas.length > 0

  function irA(id: Modulo) {
    setModulo(id)
    setRailOpen(true)
    setMasAbierto(false)
  }
  function cerrarPanel() {
    setRailOpen(false)
    requestAnimationFrame(() => {
      document
        .querySelector<HTMLButtonElement>('.guard button[aria-current="page"], .guard .guard-mas.activo')
        ?.focus({ preventScroll: true })
    })
  }

  return (
    <div className={railOpen ? "shell" : "shell panel-off"}>
      <a className="skip" href="#panel" onClick={() => setRailOpen(true)}>Saltar al panel</a>
      <header className="topbar">
        <div className="brand">
          <img className="brand-mark" src="/isotipo.svg" alt="" />
          <div>
            <strong className="word"><span className="wv">Verde</span><span className="wp">PUCP</span></strong>
            <span className="brand-sub">Gestión de Áreas Verdes</span>
          </div>
        </div>
        <div className="top-spacer" />
        <div className="roles" role="group" aria-label="Vista del mapa">
          <button type="button" aria-pressed={!relieve} onClick={() => setRelieve(false)}>Plano</button>
          <button type="button" aria-pressed={relieve} onClick={() => setRelieve(true)}>Relieve</button>
        </div>
        <div className="session">
          <span>{sesion.nombre}</span>
          <span className="chip">{etiquetaRol(sesion.rol, sesion.rol_nombre)}</span>
          <button type="button" onClick={() => void salir().then(() => setSesion(null))}>Salir</button>
        </div>
      </header>
      <nav className="guard" aria-label="Módulos">
        <div className="rail-mark"><img src="/isotipo.svg" alt="VerdePUCP" /></div>
        {barra.map((item) => (
          <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
            {item.label}
          </button>
        ))}
        {mas.length > 0 && (
          <button
            type="button"
            className={enMas ? "guard-mas activo" : "guard-mas"}
            aria-expanded={menuMas}
            aria-controls="mas-modulos"
            onClick={() => setMasAbierto((abierto) => !abierto)}
          >
            {enMas?.label ?? "Más"}
          </button>
        )}
      </nav>
      {menuMas && (
        <div
          className="guard-menu"
          id="mas-modulos"
          role="group"
          aria-label="Más módulos"
          onKeyDown={(event) => {
            if (event.key !== "Escape") return
            event.stopPropagation()
            setMasAbierto(false)
            document.querySelector<HTMLButtonElement>(".guard-mas")?.focus({ preventScroll: true })
          }}
          onBlur={(event) => {
            const siguiente = event.relatedTarget
            if (!(siguiente instanceof Element) || !siguiente.closest(".guard-menu, .guard-mas")) setMasAbierto(false)
          }}
        >
          {mas.map((item) => (
            <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
              {item.label}
            </button>
          ))}
        </div>
      )}
      <BottomSheet
        open={railOpen}
        onClose={cerrarPanel}
        vista={moduloActivo}
        onAltura={publicarAltura}
        cabeza={
          <p className={load.kind === "error" || activityError ? "status error" : "status"} role="status">
            {load.kind === "error" ? load.message : summary}
            {activityError ? ` · ${activityError}` : ""}
            {picked ? ` · ${picked}` : ""}
          </p>
        }
      >
        {moduloActivo === "mapa" && (
          <section className="block">
            <h2>{MAPA.capas}</h2>
            <div className="layer">
              <span className="swatch" style={{ background: "#c8c0b2" }} />
              <label>
                <input type="checkbox" checked={showEdificios} onChange={() => setShowEdificios((on) => !on)} /> {CAPA.edificios.label}
                <small>{CAPA.edificios.hint}</small>
              </label>
              <span className="count">{edificios.features.length || "—"}</span>
            </div>
            <fieldset className="grupo catastro-color">
              <legend>{MAPA.catastro}</legend>
              <label className="layer-toggle">
                <input type="checkbox" checked={catastroOn} onChange={() => setCatastroOn((on) => !on)} /> {MAPA.mostrarCatastro}
                <small>{conteoCatastro(data.areas?.features.length ?? "—", data.zonas?.features.length ?? "—")}</small>
              </label>
              <div className="roles segmentado" role="group" aria-label={MAPA.colorear}>
                <button type="button" aria-pressed={colorPor === "uso"} onClick={() => setColorPor("uso")}>
                  {MAPA.uso}
                </button>
                <button type="button" aria-pressed={colorPor === "sector"} onClick={() => setColorPor("sector")}>
                  {MAPA.sector}
                </button>
              </div>
              <ul className="leyenda" aria-label={colorPor === "uso" ? MAPA.leyendaUso : MAPA.leyendaSector}>
                {catsColor.map((cat) => {
                  const n = conteoColor[cat.id]
                  return (
                    <li key={cat.id}>
                      <label>
                        <input
                          type="checkbox"
                          checked={!ocultas[colorPor].includes(cat.id)}
                          disabled={n === 0}
                          onChange={() =>
                            setOcultas((current) => {
                              const activas = current[colorPor]
                              const siguiente = activas.includes(cat.id) ? activas.filter((id) => id !== cat.id) : [...activas, cat.id]
                              return { ...current, [colorPor]: siguiente }
                            })
                          }
                        />
                        <span className="swatch" style={{ background: `color-mix(in srgb, ${cat.fill} 55%, var(--hoja))`, borderColor: cat.line }} />
                        <span>{cat.label}</span>
                        <span className="count">{n}</span>
                      </label>
                    </li>
                  )
                })}
              </ul>
              {colorPor === "sector" && <p className="hint">{MAPA.cuadrillasFicticias}</p>}
            </fieldset>
            {LAYERS.filter((layer) => layer.id === "jardines_reserva" || layer.id === "xerofitica" || layer.id === "vias" || layer.id === "cuarteles").map((layer) => (
              <div className="layer" key={layer.id}>
                <span className="swatch" style={{ background: layer.fill }} />
                <label>
                  <input type="checkbox" checked={visible[layer.id]} onChange={() => setVisible((current) => ({ ...current, [layer.id]: !current[layer.id] }))} />{" "}
                  {layer.label}
                  <small>{layer.id === "cuarteles" && (data.cuarteles?.features.length ?? 0) === 0 ? ZONIFICACION.sinCuarteles : layer.hint}</small>
                </label>
                <span className="count">{data[layer.id]?.features.length ?? "—"}</span>
              </div>
            ))}
            <h2>{MAPA.inventario}</h2>
            <p className="lede">{MAPA.inventarioLede}</p>
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
            <CalendarioReservas />
          </section>
        )}
        {moduloActivo === "labores" && (
          <>
            <Labores
              rol={rol}
              cargando={!laboresListas}
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
                if (patch.titulo != null) {
                  setFormTitulo(patch.titulo)
                  setPista(null)
                  setSugerencia("")
                }
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
                    setPista(s.codigo ? { codigo: s.codigo, etiqueta: s.etiqueta } : null)
                  })
                  .catch((error: unknown) => {
                    setPista(null)
                    setSugerencia(error instanceof Error ? error.message : "Sin sugerencia")
                  })
              }}
              onConfirmarPista={() => {
                if (pista?.codigo) setFormTipo(pista.codigo)
              }}
              pista={pista}
              sugerencia={sugerencia}
            />
            <RiegoPanel capatazId={rol === "capataz" ? equipoId : formEquipo} />
            <PodaPanel />
            <ViveroPanel delta="La copia local trae 683 filas; la hoja viva del 2026-09-25 trae 689." />
          </>
        )}
        {moduloActivo === "catastro" && <CatastroEditor onModoDibujo={setModoDibujo} />}
        {moduloActivo === "inventario" && <InventarioCapas />}
        {moduloActivo === "solicitudes" && <SolicitudesPanel actividadId={selected?.queued ? "" : selected?.id ?? ""} />}
        {moduloActivo === "reportes" && <ReportesPanel />}
        {moduloActivo === "catalogos" && <CatalogosPanel editable={rol === "admin"} />}
        {moduloActivo === "importaciones" && <ImportacionesPanel />}
        {moduloActivo === "admin" && <AdminPanel />}
      </BottomSheet>
      <div className="stage" ref={stageRef}>
        <MapBoundary>
          <CampusMap
            data={dataMapa}
            visible={visibleMapa}
            activities={mapActivities}
            pinMode={pinMode && rol !== "capataz" && moduloActivo === "labores"}
            modoDibujo={moduloActivo === "catastro" ? modoDibujo : null}
            draft={draft}
            focus={focus}
            relieve={relieve}
            edificios={edificios}
            showEdificios={showEdificios}
            inventory={inventory}
            inventoryOn={inventoryOn}
            sectores={sectoresCat}
            ocultas={ocultas[colorPor]}
            onSelectCatastro={(hit) => setPicked(hit ? `${hit.layer}: ${String(hit.props.nombre || hit.props.feature_id || "polígono")}` : null)}
            onSelectActividad={(id) =>
              seleccionarActividad(id, {
                elegir: choose,
                irALabores: () => setModulo("labores"),
                abrirPanel: () => setRailOpen(true),
                limpiar: () => setSelectedId(null),
              })
            }
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
