import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { apiUrl, fetchCollection } from "./api"
import { INVENTARIO } from "./inventario"
import { CampusMap } from "./map/CampusMap"
import { seleccionarActividad } from "./map/seleccionActividad"
import { ConmutadorVista, ListaActividades, type MotivoPlano, type Vista } from "./map/VistaActividades"
import type { ModoDibujo } from "./map/draw"
import { MapBoundary } from "./map/MapBoundary"
import { useEngancharCola } from "./offline/enganchar"
import { enqueue, enqueueEstado, listEstados, listQueue, listRegistros, loadLabores, removeEstado, removeQueued, saveLabores, type QueuedLabor } from "./offline/queue"
import {
  ApiError,
  archivar,
  asignar,
  cambiarEstado,
  crearActividad,
  type AltaCampos,
  FILTRO_ACTIVIDADES,
  fetchActividades,
  type FiltroActividades,
  fetchCapataces,
  fetchTimeline,
  lonLat,
  puedeEncolarEstado,
  TIPOS,
  type Capataz,
  type CreateBody,
  type Evento,
} from "./operacion"
import { BitacoraPanel } from "./panel/Bitacora"
import { Labores, type LaborItem } from "./panel/Labores"
import { CatastroEditor } from "./panel/CatastroEditor"
import { InventarioCapas } from "./panel/InventarioCapas"
import { conCategorias, conteoPorCategoria, sectoresDesdeCatalogo, SIN_SECTOR, USOS, type Categoria, type ColorPor } from "./map/categorias"
import { ACTIVIDAD, CUENTA, MAPA, MODULO, etiquetaCuadrilla, resumenCatastro } from "./ui/nomenclatura"
import { BottomSheet } from "./ui/BottomSheet"
import { ControlMapa } from "./map/ControlMapa"
import { Hoy } from "./panel/Hoy"
import { RegistrosCampo } from "./panel/RegistrosCampo"
import { Resumen } from "./panel/Resumen"
import { Recorrido } from "./ui/Recorrido"
import { ImportacionesPanel } from "./panel/Importaciones"
import { AdminPanel } from "./panel/Admin"
import { AuditoriaPanel } from "./panel/Auditoria"
import { CatalogosPanel } from "./panel/Catalogos"
import { EjemplaresPanel } from "./panel/Ejemplares"
import { Login } from "./panel/Login"
import { ReportesPanel } from "./panel/Reportes"
import { SolicitudesPanel } from "./panel/Solicitudes"
import { etiquetaRol, fetchCatalogo, fetchSesion, salir, sugerirTipo, type CatalogoItem, type Usuario } from "./producto"
import { readColorPor, readEquipo, writeColorPor, writeEquipo } from "./session"
import { listarSectores } from "./panel/zonificacion"
import { LAYERS, type FeatureCollection, type GeoFeature, type LayerId, type Rol } from "./types"
import { MQ_MOVIL, useMedia } from "./ui/media"
import { barraMovil } from "./ui/navegacion"
import { tienePermiso, permisosDeRol } from "./ui/permisos"
import { entradaDe, etiquetaCorta, etiquetaModulo, grupoDe, modulosPorPermisos, type GrupoNav, type Modulo } from "./ui/registroModulos"

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
    clase: prop(feature, "clase"),
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
      clase: item.body.clase ?? "",
      equipo: "",
    },
  }
}

export default function App() {
  const [sesion, setSesion] = useState<Usuario | null>(null)
  const [sesionLista, setSesionLista] = useState(false)
  const [modulo, setModulo] = useState<Modulo>("resumen")
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
  const [filtro, setFiltro] = useState<FiltroActividades>(FILTRO_ACTIVIDADES)
  const [vista, setVista] = useState<Vista>("mapa")
  const [planoForzado, setPlanoForzado] = useState<MotivoPlano>(() =>
    navigator.onLine === false ? "red" : null,
  )
  const [railOpen, setRailOpen] = useState(false)
  const [enLinea, setEnLinea] = useState(() => navigator.onLine)
  const [estadosPendientes, setEstadosPendientes] = useState(0)
  const [registrosPendientes, setRegistrosPendientes] = useState(0)
  const [filtroHoy, setFiltroHoy] = useState("")
  const [ayudaForzada, setAyudaForzada] = useState(false)
  const [cuentaAbierta, setCuentaAbierta] = useState(false)
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
      const fc = await fetchActividades(rol, equipoId, filtro)
      setActivities(fc)
      void saveLabores(fc)
      setActivityError("")
    } catch (error) {
      const cached = await loadLabores<FeatureCollection>().catch(() => null)
      if (cached && Array.isArray(cached.features)) {
        setActivities(cached)
        setActivityError("Sin conexión. Se muestra la última lista guardada en este teléfono.")
      } else {
        const message = error instanceof Error ? error.message : ACTIVIDAD.noLeer
        setActivityError(message)
      }
    }
  }, [sesion, rol, equipoId, filtro])

  const reloadQueue = useCallback(async () => {
    try {
      setQueue(await listQueue())
    } catch {
      setQueue([])
    }
    try {
      setEstadosPendientes((await listEstados()).length)
    } catch {
      setEstadosPendientes(0)
    }
    try {
      setRegistrosPendientes((await listRegistros()).length)
    } catch {
      setRegistrosPendientes(0)
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
        if (user) {
          setModulo(entradaDe(user.rol))
          setRailOpen(true)
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
    fetchActividades(rol, equipoId, filtro)
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
          setActivityError("Sin conexión. Se muestra la última lista guardada en este teléfono.")
          setLaboresListas(true)
          return
        }
        setActivityError(error instanceof Error ? error.message : ACTIVIDAD.noLeer)
        setLaboresListas(true)
      })
    return () => {
      cancelled = true
    }
  }, [sesion, rol, equipoId, filtro])

  useEffect(() => {
    const caer = () => {
      setPlanoForzado("red")
      setEnLinea(false)
    }
    const volver = () => {
      setPlanoForzado((actual) => (actual === "red" ? null : actual))
      setEnLinea(true)
    }
    window.addEventListener("offline", caer)
    window.addEventListener("online", volver)
    return () => {
      window.removeEventListener("offline", caer)
      window.removeEventListener("online", volver)
    }
  }, [])

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
      if (filtro.estado && item.estado !== filtro.estado) return false
      if (filtro.tipo && item.tipo !== filtro.tipo && item.clase !== filtro.tipo) return false
      if (filtro.ejecutor && item.ejecutor !== filtro.ejecutor) return false
      if (rol !== "capataz" && filtro.cuadrillaId && item.capatazId !== filtro.cuadrillaId) return false
      return true
    })
  }, [activities, queue, filtro, rol, equipoId])

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

  async function onCreate(alta?: AltaCampos) {
    if (!draft || !formTitulo.trim()) {
      setNotice("Indique un título y un punto en el mapa.")
      return
    }
    const body: CreateBody = {
      id: crypto.randomUUID(),
      tipo: alta?.tipo || formTipo,
      titulo: formTitulo.trim(),
      detalle: formDetalle.trim(),
      lon: draft.lon,
      lat: draft.lat,
      assigned_capataz_id: formEquipo,
      actor_rol: rol,
      ejecutor: formEjecutor,
    }
    if (alta?.clase) body.clase = alta.clase
    if (alta?.subtipo) body.subtipo = alta.subtipo
    if (alta?.origen) body.origen = alta.origen
    if (alta?.codigo_externo) body.codigo_externo = alta.codigo_externo
    if (alta?.unidad_solicitante) body.unidad_solicitante = alta.unidad_solicitante
    if (alta?.nivel_riesgo) body.nivel_riesgo = alta.nivel_riesgo
    if (alta?.fecha_programada) body.fecha_programada = alta.fecha_programada
    if (alta?.cantidad) {
      const n = Number(alta.cantidad)
      if (Number.isFinite(n)) body.cantidad = n
    }
    if (alta?.personal.length) body.personal = alta.personal
    if (alta?.lugar_id) body.lugar_id = alta.lugar_id
    if (alta?.zona_supervision_id) body.zona_supervision_id = alta.zona_supervision_id
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
          setNotice(ACTIVIDAD.enCola)
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
  if (!sesion) {
    return (
      <Login
        onIn={(user) => {
          setSesion(user)
          if (user.rol === "capataz" && user.capataz_id) {
            setEquipoId(user.capataz_id)
            writeEquipo(user.capataz_id)
          }
          setModulo(entradaDe(user.rol))
          setRailOpen(true)
        }}
      />
    )
  }

  const permisos = permisosDeRol(rol)
  const permitidos = modulosPorPermisos(permisos, rol)
  const moduloActivo = permitidos.includes(modulo) ? modulo : entradaDe(rol)
  const tipos = tiposCat.length > 0 ? tiposCat.map((item) => ({ id: item.codigo, label: item.nombre })) : TIPOS.map((item) => ({ id: item.id, label: item.label }))
  const entradas = permitidos.map((id) => ({ id, label: etiquetaModulo(id, rol), corto: etiquetaCorta(id, rol) }))
  const { barra, mas } = movil ? barraMovil(rol, entradas) : { barra: entradas, mas: [] as typeof entradas }
  const enMas = mas.find((item) => item.id === moduloActivo)
  const menuMas = masAbierto && movil && mas.length > 0
  const puede = (accion: "consultar" | "registrar" | "validar" | "solicitudes" | "reportes" | "catalogos" | "usuarios" | "evidencias") => tienePermiso(permisos, accion)
  const porEnviar = (rol === "capataz" ? 0 : queue.length) + estadosPendientes + registrosPendientes
  const cuadrillaNombre = etiquetaCuadrilla(equipoId, equipos.find((item) => item.id === equipoId)?.equipo || "Cuadrilla")
  const gruposNav: GrupoNav[] = ["operacion", "datos", "configuracion"]

  function irA(id: Modulo) {
    setModulo(id)
    setRailOpen(id !== "mapa")
    setMasAbierto(false)
    setCuentaAbierta(false)
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
        <div className="session">
          <span className="session-nombre">{sesion.nombre}</span>
          <span className="chip">{etiquetaRol(sesion.rol, sesion.rol_nombre)}</span>
          {!movil && (
            <button type="button" onClick={() => void salir().then(() => setSesion(null))}>
              {CUENTA.salir}
            </button>
          )}
          <div className="cuenta">
            <button type="button" aria-expanded={cuentaAbierta} aria-controls="menu-cuenta" onClick={() => setCuentaAbierta((abierto) => !abierto)}>
              {CUENTA.menu}
            </button>
            {cuentaAbierta && (
              <div className="cuenta-menu" id="menu-cuenta" role="group" aria-label={CUENTA.menu}>
                {movil && (
                  <button type="button" onClick={() => void salir().then(() => setSesion(null))}>
                    {CUENTA.salir}
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => {
                    setAyudaForzada(true)
                    setCuentaAbierta(false)
                  }}
                >
                  {CUENTA.ayuda}
                </button>
              </div>
            )}
          </div>
        </div>
      </header>
      <nav className="guard" aria-label="Módulos">
        {movil
          ? barra.map((item) => (
              <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
                {item.corto}
              </button>
            ))
          : gruposNav.map((grupo) => {
              const items = entradas.filter((item) => grupoDe(item.id) === grupo)
              if (items.length === 0) return null
              const titulo = grupo === "operacion" ? MODULO.operacion : grupo === "datos" ? MODULO.datos : MODULO.configuracion
              return (
                <div key={grupo} className="guard-grupo">
                  <p className="guard-sep">{titulo}</p>
                  {items.map((item) => (
                    <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
                      {item.label}
                    </button>
                  ))}
                </div>
              )
            })}
        {mas.length > 0 && (
          <button
            type="button"
            className={enMas ? "guard-mas activo" : "guard-mas"}
            aria-expanded={menuMas}
            aria-controls="mas-modulos"
            onClick={() => setMasAbierto((abierto) => !abierto)}
          >
            {enMas ? enMas.corto : MODULO.mas}
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
          {gruposNav.map((grupo) => {
            const items = mas.filter((item) => grupoDe(item.id) === grupo)
            if (items.length === 0) return null
            const titulo = grupo === "operacion" ? MODULO.operacion : grupo === "datos" ? MODULO.datos : MODULO.configuracion
            return (
              <div key={grupo} className="guard-grupo">
                <p className="guard-sep">{titulo}</p>
                {items.map((item) => (
                  <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
                    {item.label}
                  </button>
                ))}
              </div>
            )
          })}
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
        {selected && moduloActivo !== "labores" && (
          <Labores
            presentacion="detalle"
            rol={rol}
            equipos={equipos}
            equipoId={equipoId}
            onEquipo={() => {}}
            items={items}
            pinMode={false}
            onPinMode={() => {}}
            draft={null}
            formTipo={formTipo}
            formTitulo=""
            formDetalle=""
            formEquipo={formEquipo}
            onForm={() => {}}
            onCreate={() => {}}
            creating={false}
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
            notice=""
            queueCount={0}
            onFlush={() => {}}
            tipos={tipos}
            formEjecutor={formEjecutor}
            motivos={motivos}
            motivo={motivo}
            onMotivo={setMotivo}
            onSugerir={() => {}}
            sugerencia=""
          />
        )}
        {moduloActivo === "hoy" && (
          <Hoy
            cuadrilla={cuadrillaNombre}
            enLinea={enLinea}
            porEnviar={porEnviar}
            items={items}
            filtroEstado={filtroHoy}
            onFiltro={setFiltroHoy}
            onAccion={(item) => {
              if (item.estado === "bloqueada") {
                choose(item.id)
                return
              }
              const siguiente = item.estado === "pendiente" ? "en_proceso" : "ejecutado"
              setSelectedId(item.id)
              setEstadoNuevo(siguiente)
              void cambiarEstado(item.id, siguiente, rol, equipoId)
                .then(() => reloadActivities())
                .catch(async (error: unknown) => {
                  if (error instanceof ApiError && error.status === 0) {
                    await enqueueEstado({ id: crypto.randomUUID(), actividadId: item.id, estado: siguiente, createdAt: new Date().toISOString() }, rol)
                    setNotice(ACTIVIDAD.enCola)
                    await reloadQueue()
                    return
                  }
                  setNotice(error instanceof Error ? error.message : ACTIVIDAD.noLeer)
                })
              if (item.estado === "en_proceso") {
                choose(item.id)
                setRailOpen(true)
              }
            }}
            onAbrir={choose}
            onRiego={() => irA("registros")}
            onPoda={() => irA("registros")}
            onVivero={() => irA("registros")}
            onBitacora={permitidos.includes("bitacora") ? () => irA("bitacora") : undefined}
            onEjemplares={permitidos.includes("ejemplares") ? () => irA("ejemplares") : undefined}
            vista={vista}
            onVista={setVista}
            mapaBloqueado={planoForzado !== null}
          />
        )}
        {moduloActivo === "resumen" && (
          <Resumen
            rol={rol}
            items={items}
            porEnviar={porEnviar}
            puedeValidar={puede("validar")}
            puedeSolicitudes={puede("solicitudes")}
            puedeUsuarios={puede("usuarios")}
            puedeReportes={puede("reportes")}
            onVerActividades={(siguiente) => {
              setFiltro(siguiente)
              irA("labores")
            }}
            onNueva={() => {
              setPinMode(true)
              irA("labores")
            }}
            onSolicitud={() => irA("solicitudes")}
            onReportes={() => irA("reportes")}
            onImportar={() => irA("importaciones")}
          />
        )}
        {moduloActivo === "mapa" && (
          <section className="block">
            <h2>{MAPA.capas}</h2>
            <p className="lede">Las capas, la leyenda y Plano o Relieve están sobre el mapa.</p>
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
              filtro={filtro}
              onFiltro={setFiltro}
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
              onCreate={(alta) => void onCreate(alta)}
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
          </>
        )}
        {moduloActivo === "registros" && <RegistrosCampo capatazId={rol === "capataz" ? equipoId : formEquipo} />}
        {moduloActivo === "catastro" && <CatastroEditor editable={puede("registrar")} onModoDibujo={setModoDibujo} />}
        {moduloActivo === "inventario" && <InventarioCapas />}
        {moduloActivo === "solicitudes" && <SolicitudesPanel actividadId={selected?.queued ? "" : selected?.id ?? ""} />}
        {moduloActivo === "reportes" && <ReportesPanel />}
        {moduloActivo === "catalogos" && <CatalogosPanel editable={rol === "admin"} />}
        {moduloActivo === "ejemplares" && <EjemplaresPanel rol={rol} />}
        {moduloActivo === "bitacora" && (
          <BitacoraPanel rol={rol} capatazId={rol === "capataz" ? equipoId : ""} actividadId={selected?.queued ? "" : (selected?.id ?? "")} />
        )}
        {moduloActivo === "importaciones" && <ImportacionesPanel />}
        {moduloActivo === "admin" && <AdminPanel />}
        {moduloActivo === "historial" && <AuditoriaPanel />}
      </BottomSheet>
      <Recorrido rol={rol} usuario={sesion.usuario} forzar={ayudaForzada} onCerrar={() => setAyudaForzada(false)} />
      <div className="stage" ref={stageRef}>
        <ConmutadorVista
          vista={planoForzado ? "lista" : vista}
          motivo={planoForzado}
          onVista={setVista}
        />
        {(planoForzado || vista === "lista") && (
          <ListaActividades
            items={items}
            selectedId={selectedId}
            onSelect={(id) => {
              choose(id)
              setRailOpen(true)
            }}
          />
        )}
        <ControlMapa
          relieve={relieve}
          onRelieve={setRelieve}
          showEdificios={showEdificios}
          onEdificios={() => setShowEdificios((on) => !on)}
          edificios={edificios.features.length}
          catastroOn={catastroOn}
          onCatastro={() => setCatastroOn((on) => !on)}
          areas={data.areas?.features.length ?? "—"}
          zonas={data.zonas?.features.length ?? "—"}
          colorPor={colorPor}
          onColorPor={setColorPor}
          categorias={catsColor}
          conteo={conteoColor}
          ocultas={ocultas[colorPor]}
          onOculta={(id) =>
            setOcultas((current) => {
              const activas = current[colorPor]
              const siguiente = activas.includes(id) ? activas.filter((item) => item !== id) : [...activas, id]
              return { ...current, [colorPor]: siguiente }
            })
          }
          visible={visible}
          onCapa={(id) => setVisible((current) => ({ ...current, [id]: !current[id] }))}
          data={data}
          inventoryOn={inventoryOn}
          onInventario={(id) => setInventoryOn((current) => ({ ...current, [id]: !current[id] }))}
          inventory={inventory}
          resumen={summary}
        />
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
            onSinPlano={() => setPlanoForzado((actual) => (actual === "red" ? actual : "teselas"))}
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
