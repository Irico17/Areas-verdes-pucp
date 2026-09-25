import { useEffect, useId, useState, type FormEvent, type KeyboardEvent } from "react"
import { listarVertices, medirGeom, moverVertice, type ModoDibujo, type Position, type VerticeRef } from "../map/draw"
import {
  areaNueva,
  areasDeFixture,
  avisoMedidas,
  bajaArea,
  bajaZona,
  CODIGOS_ZONA,
  csvAreas,
  csvZonas,
  guardarArea,
  guardarZona,
  listarAreas,
  listarZonas,
  medidasArea,
  payloadArea,
  payloadZona,
  PROYECTOS_RIEGO,
  RIEGOS_ACTUALES,
  USOS_AREA,
  validarArea,
  validarZona,
  zonaNueva,
  zonasDeFixture,
  type AreaVerde,
  type ErrorCampo,
  type ZonaSupervision,
} from "./catastro"

type Entidad = "area" | "zona"

type Props = {
  onModoDibujo?: (modo: ModoDibujo | null) => void
  cargarAreas?: () => Promise<AreaVerde[]>
  cargarZonas?: () => Promise<ZonaSupervision[]>
  persistirArea?: (area: AreaVerde, alta: boolean) => Promise<void>
  persistirZona?: (zona: ZonaSupervision, alta: boolean) => Promise<void>
  archivarArea?: (featureId: string) => Promise<void>
  archivarZona?: (codigo: string) => Promise<void>
  /** Datos ya resueltos. Si vienen, el primer render no espera a la API. */
  areasIniciales?: AreaVerde[]
  zonasIniciales?: ZonaSupervision[]
  entidadInicial?: Entidad
}

const ETIQUETA_AREA = {
  feature_id: "Identificador",
  codigo: "Código",
  nombre: "Nombre",
  uso: "Uso",
  proy_riego: "Proyecto de riego",
  riego_act: "Riego actual",
  referencia: "Referencia",
  perimetro_m: "Perímetro (m)",
  area_m2: "Área (m²)",
  geom: "Geometría",
  zona_supervision_id: "Zona de supervisión",
}

const ETIQUETA_ZONA = {
  codigo: "Código",
  nombre: "Nombre",
  area_m2: "Área (m²)",
  geom: "Geometría",
}

function tituloArea(row: AreaVerde): string {
  return row.nombre?.trim() || row.feature_id
}

function motivosDe(errores: ErrorCampo[], campo: string): string[] {
  return errores.filter((item) => item.campo === campo).map((item) => item.motivo)
}

function resumenErrores(errores: ErrorCampo[]): string {
  return `No se guardó. ${errores.map((item) => item.motivo).join(" ")}`
}

function ErroresCampo(props: { mensajes: string[] }) {
  if (props.mensajes.length === 0) return null
  return (
    <ul className="errores-campo">
      {props.mensajes.map((mensaje) => (
        <li key={mensaje} className="status error">
          {mensaje}
        </li>
      ))}
    </ul>
  )
}

function descargar(nombre: string, texto: string) {
  const blob = new Blob([texto], { type: "text/csv;charset=utf-8" })
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = nombre
  link.click()
  URL.revokeObjectURL(url)
}

export function CatastroEditor({
  onModoDibujo,
  cargarAreas = listarAreas,
  cargarZonas = listarZonas,
  persistirArea = guardarArea,
  persistirZona = guardarZona,
  archivarArea = bajaArea,
  archivarZona = bajaZona,
  areasIniciales,
  zonasIniciales,
  entidadInicial = "area",
}: Props) {
  const baseId = useId()
  const conDatos = areasIniciales != null || zonasIniciales != null
  const [entidad, setEntidad] = useState<Entidad>(entidadInicial)
  const [areas, setAreas] = useState<AreaVerde[]>(areasIniciales ?? [])
  const [zonas, setZonas] = useState<ZonaSupervision[]>(zonasIniciales ?? [])
  const [q, setQ] = useState("")
  const [loading, setLoading] = useState(!conDatos)
  const [error, setError] = useState("")
  const [aviso, setAviso] = useState("")
  const [fixture, setFixture] = useState(false)
  const [area, setArea] = useState<AreaVerde | null>(entidadInicial === "area" ? (areasIniciales?.[0] ?? null) : null)
  const [zona, setZona] = useState<ZonaSupervision | null>(entidadInicial === "zona" ? (zonasIniciales?.[0] ?? null) : null)
  const [altaArea, setAltaArea] = useState(false)
  const [altaZona, setAltaZona] = useState(false)
  const [errores, setErrores] = useState<ErrorCampo[]>([])
  const [editandoGeom, setEditandoGeom] = useState(false)
  const [vertice, setVertice] = useState(0)

  useEffect(() => {
    return () => onModoDibujo?.(null)
  }, [onModoDibujo])

  async function load(forzarFixture = false) {
    setLoading(true)
    setError("")
    try {
      if (forzarFixture) throw new Error("fixture")
      const [a, z] = await Promise.all([cargarAreas(), cargarZonas()])
      setAreas(a)
      setZonas(z)
      setFixture(false)
    } catch (err) {
      if (forzarFixture || (err instanceof Error && err.message === "fixture")) {
        setAreas(areasDeFixture())
        setZonas(zonasDeFixture())
        setFixture(true)
        setError("")
      } else {
        setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (conDatos) return
    let cancelled = false
    Promise.all([cargarAreas(), cargarZonas()])
      .then(([a, z]) => {
        if (cancelled) return
        setAreas(a)
        setZonas(z)
        setFixture(false)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el catastro")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [cargarAreas, cargarZonas, conDatos])

  const geomActiva = entidad === "area" ? area?.geom ?? null : zona?.geom ?? null
  const idActivo = entidad === "area" ? area?.feature_id ?? "" : zona?.codigo ?? ""

  useEffect(() => {
    if (!editandoGeom || !onModoDibujo || !idActivo) {
      onModoDibujo?.(null)
      return
    }
    onModoDibujo({
      entidad,
      id: idActivo,
      geom: geomActiva,
      onGeom: (geom) => {
        if (entidad === "area") setArea((current) => (current ? { ...current, geom } : current))
        else setZona((current) => (current ? { ...current, geom } : current))
      },
    })
  }, [editandoGeom, entidad, idActivo, geomActiva, onModoDibujo])

  const filtro = q.trim().toLowerCase()
  const areasVisibles = areas.filter((row) => {
    if (!filtro) return true
    return [row.feature_id, row.codigo, row.nombre, row.uso].some((value) => value?.toLowerCase().includes(filtro))
  })
  const zonasVisibles = zonas.filter((row) => {
    if (!filtro) return true
    return row.codigo.toLowerCase().includes(filtro) || row.nombre.toLowerCase().includes(filtro)
  })

  const medidas = area?.geom ? medidasArea(area) : null
  const avisoPerimetro = area && medidas ? avisoMedidas(area.perimetro_m, medidas.perimetro_m) : null
  const avisoArea = area && medidas ? avisoMedidas(area.area_m2, medidas.area_m2) : null
  const medidasZona = zona?.geom ? medirGeom(zona.geom) : null
  const avisoAreaZona = zona && medidasZona ? avisoMedidas(zona.area_m2, medidasZona.area_m2) : null
  const vertices = geomActiva ? listarVertices(geomActiva) : []
  const refActiva: VerticeRef | null = vertices[vertice] ?? null

  function nudgir(dx: number, dy: number) {
    if (!geomActiva || !refActiva) return
    const actual = geomActiva.coordinates[refActiva.polygon][refActiva.ring][refActiva.index]
    const destino: Position = [actual[0] + dx, actual[1] + dy]
    const next = moverVertice(geomActiva, refActiva, destino)
    if (entidad === "area" && area) setArea({ ...area, geom: next })
    if (entidad === "zona" && zona) setZona({ ...zona, geom: next })
  }

  function onTeclaVertice(event: KeyboardEvent) {
    const paso = 0.00005
    if (event.key === "ArrowLeft") nudgir(-paso, 0)
    else if (event.key === "ArrowRight") nudgir(paso, 0)
    else if (event.key === "ArrowUp") nudgir(0, paso)
    else if (event.key === "ArrowDown") nudgir(0, -paso)
    else return
    event.preventDefault()
  }

  async function onGuardarArea(event: FormEvent) {
    event.preventDefault()
    if (!area) return
    const payload = payloadArea(area)
    const vistos = altaArea ? areas : areas.filter((row) => row.feature_id !== area.feature_id)
    const fallos = validarArea(payload, vistos)
    setErrores(fallos)
    if (fallos.length) {
      setAviso(resumenErrores(fallos))
      return
    }
    if (fixture) {
      setAreas((rows) => (altaArea ? [payload, ...rows] : rows.map((row) => (row.feature_id === payload.feature_id ? payload : row))))
      setAltaArea(false)
      setArea(payload)
      setAviso("Payload listo. El GeoJSON de prueba no escribe en la base.")
      setEditandoGeom(false)
      return
    }
    try {
      await persistirArea(payload, altaArea)
      setAviso("Área guardada.")
      setAltaArea(false)
      setEditandoGeom(false)
      await load()
      setArea(payload)
    } catch (err) {
      setAviso(err instanceof Error ? err.message : "No se pudo guardar")
    }
  }

  async function onGuardarZona(event: FormEvent) {
    event.preventDefault()
    if (!zona) return
    const payload = payloadZona(zona)
    const vistos = altaZona ? zonas : zonas.filter((row) => row.codigo !== zona.codigo)
    const fallos = validarZona(payload, vistos)
    setErrores(fallos)
    if (fallos.length) {
      setAviso(resumenErrores(fallos))
      return
    }
    if (fixture) {
      setZonas((rows) => (altaZona ? [payload, ...rows] : rows.map((row) => (row.codigo === payload.codigo ? payload : row))))
      setAltaZona(false)
      setZona(payload)
      setAviso("Payload listo. El GeoJSON de prueba no escribe en la base.")
      setEditandoGeom(false)
      return
    }
    try {
      await persistirZona(payload, altaZona)
      setAviso("Zona guardada.")
      setAltaZona(false)
      setEditandoGeom(false)
      await load()
      setZona(payload)
    } catch (err) {
      setAviso(err instanceof Error ? err.message : "No se pudo guardar")
    }
  }

  return (
    <section className="split catastro-editor">
      <div className="split-list">
        <h2>Catastro</h2>
        <div className="roles" role="tablist" aria-label="Entidad del catastro">
          <button type="button" role="tab" aria-selected={entidad === "area"} onClick={() => { setEntidad("area"); setErrores([]); setEditandoGeom(false) }}>
            Áreas verdes
          </button>
          <button type="button" role="tab" aria-selected={entidad === "zona"} onClick={() => { setEntidad("zona"); setErrores([]); setEditandoGeom(false) }}>
            Zonas
          </button>
        </div>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
          }}
        >
          <label className="field" htmlFor={`${baseId}-q`}>
            Buscar
            <input id={`${baseId}-q`} value={q} onChange={(event) => setQ(event.target.value)} placeholder="Código, nombre o uso" />
          </label>
        </form>
        {error && <p className="status error">{error}</p>}
        {error && (
          <button type="button" onClick={() => void load(true)}>
            Abrir GeoJSON de prueba
          </button>
        )}
        {fixture && <p className="hint">GeoJSON de prueba. Guardar arma el payload y no escribe en la base.</p>}
        {loading && (
          <div className="skel-wrap" aria-hidden="true">
            <div className="skel" />
            <div className="skel" />
            <div className="skel" />
          </div>
        )}
        {!loading && entidad === "area" && areasVisibles.length === 0 && !error && (
          <p className="empty">Ningún área coincide. Puede registrar una sin geometría.</p>
        )}
        {!loading && entidad === "zona" && zonasVisibles.length === 0 && !error && (
          <p className="empty">No hay zonas de supervisión cargadas.</p>
        )}
        <ul className="labor-list">
          {entidad === "area" &&
            areasVisibles.map((row) => (
              <li key={row.feature_id}>
                <button
                  type="button"
                  className={area?.feature_id === row.feature_id ? "labor on" : "labor"}
                  onClick={() => {
                    setArea(row)
                    setAltaArea(false)
                    setErrores([])
                    setEditandoGeom(false)
                    setVertice(0)
                  }}
                >
                  <span>
                    <strong>{tituloArea(row)}</strong>
                    <small>
                      {row.feature_id}
                      {row.uso ? ` · ${row.uso}` : ""}
                      {row.geom ? "" : " · sin geometría"}
                    </small>
                  </span>
                </button>
              </li>
            ))}
          {entidad === "zona" &&
            zonasVisibles.map((row) => (
              <li key={row.codigo}>
                <button
                  type="button"
                  className={zona?.codigo === row.codigo ? "labor on" : "labor"}
                  onClick={() => {
                    setZona(row)
                    setAltaZona(false)
                    setErrores([])
                    setEditandoGeom(false)
                    setVertice(0)
                  }}
                >
                  <span>
                    <strong>{row.nombre || row.codigo}</strong>
                    <small>
                      {row.codigo}
                      {row.area_m2 != null ? ` · ${row.area_m2} m²` : ""}
                    </small>
                  </span>
                </button>
              </li>
            ))}
        </ul>
        <p className="row-actions">
          {entidad === "area" ? (
            <button
              type="button"
              onClick={() => {
                setArea(areaNueva(areas.map((row) => row.feature_id)))
                setAltaArea(true)
                setErrores([])
                setEditandoGeom(false)
              }}
            >
              Nueva área
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                setZona(zonaNueva(zonas.map((row) => row.codigo)))
                setAltaZona(true)
                setErrores([])
                setEditandoGeom(false)
              }}
            >
              Nueva zona
            </button>
          )}
          <button
            type="button"
            onClick={() => descargar(entidad === "area" ? "areas-verdes.csv" : "zonas-supervision.csv", entidad === "area" ? csvAreas(areas) : csvZonas(zonas))}
          >
            Exportar CSV
          </button>
        </p>
      </div>
      <div className="split-detail">
        {entidad === "area" && !area && <p className="empty">Elija un área. La ficha queda en este panel.</p>}
        {entidad === "zona" && !zona && <p className="empty">Elija una zona de supervisión. El mapa muestra su polígono al editar.</p>}
        {entidad === "area" && area && (
          <form className="form" onSubmit={(event) => void onGuardarArea(event)}>
            <h3>{altaArea ? "Nueva área" : tituloArea(area)}</h3>
            <label className="field" htmlFor={`${baseId}-feature_id`}>
              {ETIQUETA_AREA.feature_id}
              <input
                id={`${baseId}-feature_id`}
                name="feature_id"
                value={area.feature_id}
                readOnly={!altaArea}
                required
                onChange={(event) => setArea({ ...area, feature_id: event.target.value })}
                aria-invalid={motivosDe(errores, "feature_id").length > 0}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "feature_id")} />
            <label className="field" htmlFor={`${baseId}-codigo`}>
              {ETIQUETA_AREA.codigo}
              <input id={`${baseId}-codigo`} name="codigo" value={area.codigo ?? ""} onChange={(event) => setArea({ ...area, codigo: event.target.value || null })} aria-invalid={motivosDe(errores, "codigo").length > 0} />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "codigo")} />
            <label className="field" htmlFor={`${baseId}-nombre`}>
              {ETIQUETA_AREA.nombre}
              <input id={`${baseId}-nombre`} name="nombre" value={area.nombre ?? ""} onChange={(event) => setArea({ ...area, nombre: event.target.value || null })} />
            </label>
            <label className="field" htmlFor={`${baseId}-uso`}>
              {ETIQUETA_AREA.uso}
              <select id={`${baseId}-uso`} name="uso" value={area.uso ?? ""} onChange={(event) => setArea({ ...area, uso: event.target.value || null })}>
                <option value="">Sin uso</option>
                {USOS_AREA.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
              </select>
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "uso")} />
            <label className="field" htmlFor={`${baseId}-proy`}>
              {ETIQUETA_AREA.proy_riego}
              <select id={`${baseId}-proy`} name="proy_riego" value={area.proy_riego ?? ""} onChange={(event) => setArea({ ...area, proy_riego: event.target.value || null })}>
                <option value="">Sin proyecto</option>
                {PROYECTOS_RIEGO.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
              </select>
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "proy_riego")} />
            <label className="field" htmlFor={`${baseId}-riego`}>
              {ETIQUETA_AREA.riego_act}
              <select id={`${baseId}-riego`} name="riego_act" value={area.riego_act ?? ""} onChange={(event) => setArea({ ...area, riego_act: event.target.value || null })}>
                <option value="">Sin riego</option>
                {RIEGOS_ACTUALES.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
              </select>
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "riego_act")} />
            <label className="field" htmlFor={`${baseId}-ref`}>
              {ETIQUETA_AREA.referencia}
              <textarea
                id={`${baseId}-ref`}
                name="referencia"
                maxLength={500}
                value={area.referencia ?? ""}
                onChange={(event) => setArea({ ...area, referencia: event.target.value || null })}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "referencia")} />
            <label className="field" htmlFor={`${baseId}-peri`}>
              {ETIQUETA_AREA.perimetro_m}
              <input
                id={`${baseId}-peri`}
                name="perimetro_m"
                inputMode="decimal"
                value={area.perimetro_m ?? ""}
                onChange={(event) => setArea({ ...area, perimetro_m: event.target.value === "" ? null : Number(event.target.value) })}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "perimetro_m")} />
            {avisoPerimetro && <p className="hint">{avisoPerimetro}</p>}
            <label className="field" htmlFor={`${baseId}-am2`}>
              {ETIQUETA_AREA.area_m2}
              <input
                id={`${baseId}-am2`}
                name="area_m2"
                inputMode="decimal"
                value={area.area_m2 ?? ""}
                onChange={(event) => setArea({ ...area, area_m2: event.target.value === "" ? null : Number(event.target.value) })}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "area_m2")} />
            {avisoArea && <p className="hint">{avisoArea}</p>}
            <label className="field" htmlFor={`${baseId}-zona`}>
              {ETIQUETA_AREA.zona_supervision_id}
              <select
                id={`${baseId}-zona`}
                name="zona_supervision_id"
                value={area.zona_supervision_id ?? ""}
                onChange={(event) => setArea({ ...area, zona_supervision_id: event.target.value || null })}
              >
                <option value="">Sin zona</option>
                {CODIGOS_ZONA.map((item) => (
                  <option key={item} value={item}>
                    {zonas.find((row) => row.codigo === item)?.nombre || item}
                  </option>
                ))}
              </select>
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "zona_supervision_id")} />
            <fieldset className="geom-box">
              <legend>{ETIQUETA_AREA.geom}</legend>
              <p className="hint">{area.geom ? "MultiPolygon listo para guardar." : "Sin geometría. Puede dibujarla en el mapa."}</p>
              <ErroresCampo mensajes={motivosDe(errores, "geom")} />
              <button
                type="button"
                className={editandoGeom ? "primary" : undefined}
                aria-pressed={editandoGeom}
                onClick={() => setEditandoGeom((on) => !on)}
              >
                {editandoGeom ? "Cerrar edición de vértices" : "Editar geometría"}
              </button>
            </fieldset>
            {editandoGeom && <ControlVertices vertices={vertices} vertice={vertice} onVertice={setVertice} onTecla={onTeclaVertice} onNudge={nudgir} />}
            {errores.length > 0 && (
              <p className="status error" role="alert">
                {resumenErrores(errores)}
              </p>
            )}
            <button type="submit" className="primary">
              Guardar área
            </button>
            {!altaArea && (
              <button
                type="button"
                className="danger"
                onClick={() => {
                  if (fixture) {
                    setAreas((rows) => rows.filter((row) => row.feature_id !== area.feature_id))
                    setArea(null)
                    setAviso("Baja lógica en el GeoJSON de prueba.")
                    return
                  }
                  void archivarArea(area.feature_id)
                    .then(() => {
                      setArea(null)
                      setAviso("Área dada de baja.")
                      return load()
                    })
                    .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo dar de baja"))
                }}
              >
                Dar de baja
              </button>
            )}
          </form>
        )}
        {entidad === "zona" && zona && (
          <form className="form" onSubmit={(event) => void onGuardarZona(event)}>
            <h3>{altaZona ? "Nueva zona" : zona.nombre || zona.codigo}</h3>
            <label className="field" htmlFor={`${baseId}-zcode`}>
              {ETIQUETA_ZONA.codigo}
              <select
                id={`${baseId}-zcode`}
                name="codigo"
                required
                value={zona.codigo}
                disabled={!altaZona}
                onChange={(event) => setZona({ ...zona, codigo: event.target.value })}
              >
                <option value="">Elegir</option>
                {CODIGOS_ZONA.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
              </select>
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "codigo")} />
            <label className="field" htmlFor={`${baseId}-znombre`}>
              {ETIQUETA_ZONA.nombre}
              <input
                id={`${baseId}-znombre`}
                name="nombre"
                required
                value={zona.nombre}
                onChange={(event) => setZona({ ...zona, nombre: event.target.value })}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "nombre")} />
            <label className="field" htmlFor={`${baseId}-zam2`}>
              {ETIQUETA_ZONA.area_m2}
              <input
                id={`${baseId}-zam2`}
                name="area_m2"
                inputMode="decimal"
                value={zona.area_m2 ?? ""}
                onChange={(event) => setZona({ ...zona, area_m2: event.target.value === "" ? null : Number(event.target.value) })}
              />
            </label>
            <ErroresCampo mensajes={motivosDe(errores, "area_m2")} />
            {avisoAreaZona && <p className="hint">{avisoAreaZona}</p>}
            <fieldset className="geom-box">
              <legend>{ETIQUETA_ZONA.geom}</legend>
              <p className="hint">{zona.geom ? "MultiPolygon de la zona." : "La zona necesita polígono para guardarse."}</p>
              <ErroresCampo mensajes={motivosDe(errores, "geom")} />
              <button type="button" className={editandoGeom ? "primary" : undefined} aria-pressed={editandoGeom} onClick={() => setEditandoGeom((on) => !on)}>
                {editandoGeom ? "Cerrar edición de vértices" : "Editar geometría"}
              </button>
            </fieldset>
            {editandoGeom && <ControlVertices vertices={vertices} vertice={vertice} onVertice={setVertice} onTecla={onTeclaVertice} onNudge={nudgir} />}
            {errores.length > 0 && (
              <p className="status error" role="alert">
                {resumenErrores(errores)}
              </p>
            )}
            <button type="submit" className="primary">
              Guardar zona
            </button>
            {!altaZona && (
              <button
                type="button"
                className="danger"
                onClick={() => {
                  if (fixture) {
                    setZonas((rows) => rows.filter((row) => row.codigo !== zona.codigo))
                    setZona(null)
                    setAviso("Baja lógica en el GeoJSON de prueba.")
                    return
                  }
                  void archivarZona(zona.codigo)
                    .then(() => {
                      setZona(null)
                      setAviso("Zona dada de baja.")
                      return load()
                    })
                    .catch((err: unknown) => setAviso(err instanceof Error ? err.message : "No se pudo dar de baja"))
                }}
              >
                Dar de baja
              </button>
            )}
          </form>
        )}
        {aviso && <p className={aviso.toLowerCase().includes("no se") || aviso.toLowerCase().includes("respondió") ? "status error" : "banner"}>{aviso}</p>}
      </div>
    </section>
  )
}

function ControlVertices(props: {
  vertices: VerticeRef[]
  vertice: number
  onVertice: (index: number) => void
  onTecla: (event: KeyboardEvent) => void
  onNudge: (dx: number, dy: number) => void
}) {
  const paso = 0.00005
  return (
    <div className="vertice-box" onKeyDown={props.onTecla}>
      <p className="hint">Arrastre un vértice en el mapa, o muévalo con las flechas.</p>
      <label className="field">
        Vértice
        <select value={props.vertices.length ? props.vertice : ""} onChange={(event) => props.onVertice(Number(event.target.value))} disabled={props.vertices.length === 0}>
          {props.vertices.length === 0 && <option value="">Dibuje el polígono en el mapa</option>}
          {props.vertices.map((ref, index) => (
            <option key={`${ref.polygon}-${ref.ring}-${ref.index}`} value={index}>
              {index + 1} de {props.vertices.length}
            </option>
          ))}
        </select>
      </label>
      <p className="row-actions">
        <button type="button" onClick={() => props.onNudge(-paso, 0)} disabled={!props.vertices.length}>
          Oeste
        </button>
        <button type="button" onClick={() => props.onNudge(paso, 0)} disabled={!props.vertices.length}>
          Este
        </button>
        <button type="button" onClick={() => props.onNudge(0, paso)} disabled={!props.vertices.length}>
          Norte
        </button>
        <button type="button" onClick={() => props.onNudge(0, -paso)} disabled={!props.vertices.length}>
          Sur
        </button>
      </p>
    </div>
  )
}
