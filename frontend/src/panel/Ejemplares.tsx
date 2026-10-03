import { useEffect, useState, type FormEvent } from "react"
import { apiUrl } from "../api"
import { fetchCatalogo, type CatalogoItem } from "../producto"
import type { Rol } from "../types"
import { Esqueleto } from "../ui/Esqueleto"
import { CLASE_CATALOGO, EJEMPLAR } from "../ui/nomenclatura"
import {
  cuerpoFicha,
  etiquetaSalud,
  lineaCodigo,
  puedeEditarEjemplar,
  tituloLista,
  type CodigoHistorico,
  type Ejemplar,
  type Opcion,
} from "./ejemplares"

type Props = { rol: Rol }

const PAGINA = 40

export function EjemplaresPanel({ rol }: Props) {
  const edita = puedeEditarEjemplar(rol)
  const [q, setQ] = useState("")
  const [consulta, setConsulta] = useState("")
  const [items, setItems] = useState<Ejemplar[]>([])
  const [total, setTotal] = useState(0)
  const [cargando, setCargando] = useState(true)
  const [aviso, setAviso] = useState("")
  const [sel, setSel] = useState<number | null>(null)
  const [codigos, setCodigos] = useState<CodigoHistorico[]>([])
  const [especies, setEspecies] = useState<Opcion[]>([])
  const [lugares, setLugares] = useState<Opcion[]>([])
  const [sectores, setSectores] = useState<CatalogoItem[]>([])
  const [salud, setSalud] = useState("")
  const [especieId, setEspecieId] = useState("")
  const [lugarId, setLugarId] = useState("")
  const [sectorId, setSectorId] = useState("")
  const [lat, setLat] = useState("")
  const [lon, setLon] = useState("")
  const [codigoNuevo, setCodigoNuevo] = useState("")
  const [guardando, setGuardando] = useState(false)

  useEffect(() => {
    let vivo = true
    const id = window.setTimeout(() => {
      const texto = q.trim()
      setConsulta(texto)
      setCargando(true)
      void leerPagina(texto, 0)
        .then((pagina) => {
          if (!vivo) return
          setItems(pagina.ejemplares)
          setTotal(pagina.total)
          setAviso("")
        })
        .catch(() => {
          if (vivo) setAviso(EJEMPLAR.errorLista)
        })
        .finally(() => {
          if (vivo) setCargando(false)
        })
    }, 200)
    return () => {
      vivo = false
      window.clearTimeout(id)
    }
  }, [q])

  useEffect(() => {
    let vivo = true
    void Promise.all([
      leerOpciones("/catastro/especies", "especies"),
      leerOpciones("/catastro/lugares", "lugares"),
      fetchCatalogo("sector_capataz", true),
      fetchCatalogo("cuartel", true),
    ])
      .then(([esp, lug, sec, cua]) => {
        if (!vivo) return
        setEspecies(esp)
        setLugares(lug)
        setSectores([...sec, ...cua])
      })
      .catch(() => {})
    return () => {
      vivo = false
    }
  }, [])

  const elegido = items.find((item) => item.id === sel) ?? null

  function abrir(item: Ejemplar) {
    setSel(item.id)
    setSalud(item.salud ?? "")
    setEspecieId(item.especie_id ? String(item.especie_id) : "")
    setLugarId(item.ubicacion_lugar_id ? String(item.ubicacion_lugar_id) : "")
    setSectorId(item.sector_cuartel_id ? String(item.sector_cuartel_id) : "")
    setLat(item.lat == null ? "" : String(item.lat))
    setLon(item.lon == null ? "" : String(item.lon))
    setCodigoNuevo("")
    setAviso("")
    setCodigos([])
    void fetch(apiUrl(`/catastro/ejemplares/${item.id}/codigos`), { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((body: { codigos?: CodigoHistorico[] } | null) => {
        if (body?.codigos) setCodigos(body.codigos)
      })
      .catch(() => {})
  }

  async function guardar(event: FormEvent) {
    event.preventDefault()
    if (!elegido || !edita) return
    const listo = cuerpoFicha({ salud, especieId, lugarId, sectorId, lat, lon })
    if (!listo.ok) {
      setAviso(listo.aviso)
      return
    }
    setGuardando(true)
    try {
      const res = await fetch(apiUrl(`/catastro/ejemplares/${elegido.id}`), {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(listo.cuerpo),
      })
      if (!res.ok) {
        setAviso(EJEMPLAR.error)
        return
      }
      const fresco = (await res.json()) as Ejemplar
      setItems((lista) => lista.map((item) => (item.id === fresco.id ? { ...item, ...fresco } : item)))
      setSalud(fresco.salud ?? "")
      setAviso(EJEMPLAR.guardada)
    } catch {
      setAviso(EJEMPLAR.error)
    } finally {
      setGuardando(false)
    }
  }

  async function recodificar(event: FormEvent) {
    event.preventDefault()
    if (!elegido || !edita) return
    const nuevo = codigoNuevo.trim()
    if (!nuevo) {
      setAviso(EJEMPLAR.errorCodigo)
      return
    }
    setGuardando(true)
    try {
      const res = await fetch(apiUrl(`/catastro/ejemplares/${elegido.id}/codigos`), {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ codigo_nuevo: nuevo }),
      })
      if (!res.ok) {
        setAviso(EJEMPLAR.error)
        return
      }
      const fila = (await res.json()) as CodigoHistorico
      setCodigos((lista) => [...lista, fila])
      setItems((lista) => lista.map((item) => (item.id === elegido.id ? { ...item, codigo: fila.codigo_nuevo || nuevo } : item)))
      setCodigoNuevo("")
      setAviso(EJEMPLAR.recodificado)
    } catch {
      setAviso(EJEMPLAR.error)
    } finally {
      setGuardando(false)
    }
  }

  async function mas() {
    const pagina = await leerPagina(consulta, items.length)
    setItems((lista) => [...lista, ...pagina.ejemplares])
    setTotal(pagina.total)
  }

  const sectoresCapataz = sectores.filter((item) => item.clase === "sector_capataz")
  const cuarteles = sectores.filter((item) => item.clase === "cuartel")

  return (
    <section className="block ejemplares">
      <h2>{EJEMPLAR.titulo}</h2>
      <p className="lede">{EJEMPLAR.lede}</p>
      {!edita && <p className="hint">{EJEMPLAR.soloConsulta}</p>}
      <label className="field">
        {EJEMPLAR.buscar}
        <input value={q} onChange={(event) => setQ(event.target.value)} autoComplete="off" />
      </label>
      {cargando && <Esqueleto />}
      {!cargando && items.length === 0 && <p className="empty">{EJEMPLAR.vacio}</p>}
      <ul className="labor-list">
        {items.map((item) => (
          <li key={item.id}>
            <button
              type="button"
              className={item.id === sel ? "labor on" : "labor"}
              aria-pressed={item.id === sel}
              onClick={() => abrir(item)}
            >
              <span className="ejemplar-codigo">{tituloLista(item)}</span>
              <span>
                <strong>{item.nombre_comun.trim() || EJEMPLAR.sinNombre}</strong>
                <small>
                  <span className="ejemplar-salud" data-salud={item.salud || "sin-dato"}>
                    {etiquetaSalud(item.salud)}
                  </span>
                  {" · "}
                  {item.sector_cuartel_nombre || EJEMPLAR.sinSector}
                </small>
              </span>
            </button>
          </li>
        ))}
      </ul>
      {items.length < total && (
        <button type="button" onClick={() => void mas()}>
          {EJEMPLAR.mas}
        </button>
      )}

      {!elegido && !cargando && items.length > 0 && <p className="hint">{EJEMPLAR.elegir}</p>}

      {elegido && (
        <div className="detail">
          <h3>{tituloLista(elegido)}</h3>
          <p className="meta">
            {elegido.especie_cientifico || elegido.nombre_comun || EJEMPLAR.sinEspecie}
            {elegido.lugar_nombre ? ` · ${elegido.lugar_nombre}` : ""}
          </p>
          <form className="form" onSubmit={(event) => void guardar(event)}>
            <label className="field">
              {EJEMPLAR.salud}
              <select value={salud} disabled={!edita || guardando} onChange={(event) => setSalud(event.target.value)}>
                <option value="">{EJEMPLAR.sinSalud}</option>
                <option value="bueno">{EJEMPLAR.saludBueno}</option>
                <option value="regular">{EJEMPLAR.saludRegular}</option>
                <option value="deteriorado">{EJEMPLAR.saludDeteriorado}</option>
              </select>
            </label>
            <label className="field">
              {EJEMPLAR.especie}
              <select value={especieId} disabled={!edita || guardando} onChange={(event) => setEspecieId(event.target.value)}>
                <option value="">{EJEMPLAR.sinEspecie}</option>
                {especies.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.nombre}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              {EJEMPLAR.lugar}
              <select value={lugarId} disabled={!edita || guardando} onChange={(event) => setLugarId(event.target.value)}>
                <option value="">{EJEMPLAR.sinLugar}</option>
                {lugares.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.nombre}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              {EJEMPLAR.sector}
              <select value={sectorId} disabled={!edita || guardando} onChange={(event) => setSectorId(event.target.value)}>
                <option value="">{EJEMPLAR.sinSector}</option>
                <optgroup label={CLASE_CATALOGO.sector_capataz}>
                  {sectoresCapataz.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.nombre}
                    </option>
                  ))}
                </optgroup>
                <optgroup label={CLASE_CATALOGO.cuartel}>
                  {cuarteles.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.nombre}
                    </option>
                  ))}
                </optgroup>
              </select>
            </label>
            <div className="ejemplar-par">
              <label className="field">
                {EJEMPLAR.lat}
                <input
                  inputMode="decimal"
                  value={lat}
                  disabled={!edita || guardando}
                  onChange={(event) => setLat(event.target.value)}
                />
              </label>
              <label className="field">
                {EJEMPLAR.lon}
                <input
                  inputMode="decimal"
                  value={lon}
                  disabled={!edita || guardando}
                  onChange={(event) => setLon(event.target.value)}
                />
              </label>
            </div>
            {edita && (
              <button className="primary" type="submit" disabled={guardando}>
                {guardando ? EJEMPLAR.guardando : EJEMPLAR.guardar}
              </button>
            )}
          </form>

          {edita && (
            <form className="form" onSubmit={(event) => void recodificar(event)}>
              <label className="field">
                {EJEMPLAR.codigoNuevo}
                <input value={codigoNuevo} disabled={guardando} onChange={(event) => setCodigoNuevo(event.target.value)} />
              </label>
              <button type="submit" disabled={guardando}>
                {EJEMPLAR.recodificar}
              </button>
            </form>
          )}

          <h3>{EJEMPLAR.historial}</h3>
          {codigos.length === 0 ? (
            <p className="empty">{EJEMPLAR.sinHistorial}</p>
          ) : (
            <ol className="ejemplar-codigos">
              {codigos.map((fila) => (
                <li key={fila.id}>{lineaCodigo(fila.codigo_anterior, fila.codigo_nuevo)}</li>
              ))}
            </ol>
          )}
          {aviso && (
            <p className="hint" role="status">
              {aviso}
            </p>
          )}
        </div>
      )}
      {!elegido && aviso && (
        <p className="hint" role="status">
          {aviso}
        </p>
      )}
    </section>
  )
}

async function leerPagina(q: string, offset: number): Promise<{ ejemplares: Ejemplar[]; total: number }> {
  const params = new URLSearchParams({ limit: String(PAGINA), offset: String(offset) })
  if (q) params.set("q", q)
  const res = await fetch(apiUrl(`/catastro/ejemplares?${params}`), { credentials: "include" })
  if (!res.ok) throw new Error(String(res.status))
  const body = (await res.json()) as { ejemplares?: Ejemplar[]; total?: number }
  return { ejemplares: body.ejemplares ?? [], total: body.total ?? 0 }
}

async function leerOpciones(ruta: string, clave: "especies" | "lugares"): Promise<Opcion[]> {
  const res = await fetch(apiUrl(ruta), { credentials: "include" })
  if (!res.ok) return []
  const body = (await res.json()) as {
    especies?: { id: number; nombre_cientifico?: string; nombre_comun?: string }[]
    lugares?: { id: number; nombre: string }[]
  }
  if (clave === "especies") {
    return (body.especies ?? []).map((item) => ({
      id: item.id,
      nombre: item.nombre_cientifico || item.nombre_comun || String(item.id),
    }))
  }
  return (body.lugares ?? []).map((item) => ({ id: item.id, nombre: item.nombre }))
}
