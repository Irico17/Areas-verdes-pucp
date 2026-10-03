import { useEffect, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFecha, hoyISO } from "../fecha"
import { Esqueleto } from "../ui/Esqueleto"
import { ApiError } from "../operacion"
import { crearRiego, fetchRiego } from "../producto"
import { COLA, FALLO, RIEGO, etiquetaZonaSupervision } from "../ui/nomenclatura"
import { listarSectores, type SectorCapataz } from "./zonificacion"
import { encolarRegistro, type QueuedRegistro } from "../offline/queue"
import { COLA_VACIADA, pendientesDe, rutaRiego, vaciarRegistros, type DetalleCola } from "../offline/registros"

type RiegoLocal = { id: string; sectorId: number; turno: string; fecha: string; nota: string }

function riegoLocal(item: QueuedRegistro): RiegoLocal | null {
  if (!item.body || typeof item.body !== "object") return null
  const row = item.body as { id?: string; sector_id?: number; turno?: string; fecha?: string; nota?: string }
  if (!row.id || !row.sector_id) return null
  return { id: row.id, sectorId: row.sector_id, turno: row.turno ?? "", fecha: row.fecha ?? "", nota: row.nota ?? "" }
}

export function RiegoPanel(props: { capatazId: string; mostrarFormulario?: boolean }) {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof fetchRiego>>["registros"]>([])
  const [sectores, setSectores] = useState<SectorCapataz[]>([])
  const [cobertura, setCobertura] = useState<number | null>(null)
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState("")
  const [sectorId, setSectorId] = useState("")
  const [zona, setZona] = useState("Z1")
  const [turno, setTurno] = useState("manana")
  const [fecha, setFecha] = useState(hoyISO)
  const [nota, setNota] = useState("")
  const [locales, setLocales] = useState<RiegoLocal[]>([])
  const [avisoCola, setAvisoCola] = useState("")

  async function load() {
    try {
      const body = await fetchRiego()
      setCobertura(body.cobertura)
      setRows(body.registros)
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : RIEGO.errorLeer)
    }
  }

  useEffect(() => {
    let vivo = true
    void pendientesDe("riego")
      .then((filas) => {
        if (!vivo) return
        setLocales(filas.map(riegoLocal).filter((fila): fila is RiegoLocal => fila != null))
      })
      .catch(() => {})
    const alVolver = (ev: Event) => {
      const detail = (ev as CustomEvent<DetalleCola>).detail
      void pendientesDe("riego")
        .then((filas) => {
          if (!vivo) return
          setLocales(filas.map(riegoLocal).filter((fila): fila is RiegoLocal => fila != null))
          if (detail?.conflictos?.some((item) => item.tipo === "riego")) setAvisoCola(COLA.conflicto)
          else if (filas.length === 0) setAvisoCola("")
        })
        .catch(() => {})
      void load()
    }
    window.addEventListener(COLA_VACIADA, alVolver)
    return () => {
      vivo = false
      window.removeEventListener(COLA_VACIADA, alVolver)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    fetchRiego()
      .then((body) => {
        if (cancelled) return
        setCobertura(body.cobertura)
        setRows(body.registros)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : RIEGO.errorLeer)
      })
      .finally(() => {
        if (!cancelled) setCargando(false)
      })
    listarSectores(true)
      .then((filas) => {
        if (!cancelled) setSectores(filas)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : RIEGO.errorSectores)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block riego" id="riego" aria-labelledby="riego-titulo">
      <h2 id="riego-titulo">{RIEGO.titulo}</h2>
      <p className="lede">{RIEGO.lede}</p>
      {cobertura !== null && (
        <p className="riego-cobertura" role="status">
          {RIEGO.cobertura(cobertura)}
        </p>
      )}
      <p className="hint">{RIEGO.turnos(rows.length)}</p>
      {avisoCola && (
        <p className="hint" role="status">
          {avisoCola} {locales.length > 0 && <span className="marca-cola">{COLA.marca}</span>}
        </p>
      )}
      {locales.length > 0 && (
        <p className="hint" role="status">
          {COLA.local(locales.length)}{" "}
          <button type="button" className="link" onClick={() => void vaciarRegistros()}>
            {COLA.reintentar}
          </button>
        </p>
      )}
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && rows.length === 0 && !error && <p className="empty">{RIEGO.vacio}</p>}
      <ul className="labor-list">
        {locales.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{sectores.find((sector) => sector.id === row.sectorId)?.nombre ?? `Sector ${row.sectorId}`}</strong> <span className="marca-cola">{COLA.marca}</span>
            <small>
              {formatFecha(row.fecha)} · {row.turno === "manana" ? RIEGO.manana.toLowerCase() : row.turno}
            </small>
          </li>
        ))}
        {rows.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.sector}</strong>
            <small>
              {formatFecha(row.fecha)} · {row.turno === "manana" ? RIEGO.manana.toLowerCase() : row.turno} · {row.equipo || RIEGO.sinCuadrilla}
            </small>
          </li>
        ))}
      </ul>
      {props.mostrarFormulario !== false && <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          if (event.currentTarget.querySelector('[aria-invalid="true"]')) {
            setError(RIEGO.errorFecha)
            return
          }
          const idSector = Number(sectorId)
          if (!Number.isInteger(idSector) || idSector <= 0) {
            setError(RIEGO.elegirSector)
            return
          }
          const id = crypto.randomUUID()
          const cuerpo = {
            id,
            sector_id: idSector,
            turno,
            capataz_id: props.capatazId,
            fecha,
            nota,
            zona_supervision_id: zona,
          }
          void crearRiego(cuerpo)
            .then(() => {
              setNota("")
              setAvisoCola("")
              return load()
            })
            .catch((err: unknown) => {
              if (err instanceof ApiError && err.status === 409) {
                setAvisoCola(COLA.conflicto)
                return
              }
              if (err instanceof ApiError && err.status === 403) {
                setError(FALLO.riego)
                return
              }
              if (err instanceof ApiError && err.status >= 400 && err.status < 500) {
                setError(err.message)
                return
              }
              if (err instanceof ApiError && err.status === 0) {
                void encolarRegistro({
                  id,
                  tipo: "riego",
                  path: rutaRiego(),
                  method: "POST",
                  body: cuerpo,
                  createdAt: new Date().toISOString(),
                }).then(() => {
                  setLocales((actual) => [...actual.filter((fila) => fila.id !== id), { id, sectorId: idSector, turno, fecha, nota }])
                  setNota("")
                  setError("")
                  setAvisoCola(RIEGO.enCola)
                })
                return
              }
              setError(err instanceof Error ? err.message : RIEGO.errorRegistrar)
            })
        }}
      >
        <label className="field" htmlFor="riego-zona">
          {RIEGO.zona}
          <select id="riego-zona" name="zona_supervision_id" value={zona} onChange={(event) => setZona(event.target.value)}>
            {["Z1", "Z2", "Z3", "Z4"].map((codigo) => (
              <option key={codigo} value={codigo}>
                {etiquetaZonaSupervision(codigo)}
              </option>
            ))}
          </select>
        </label>
        <label className="field" htmlFor="riego-sector">
          {RIEGO.sector}
          <select
            id="riego-sector"
            name="sector_id"
            value={sectorId}
            onChange={(event) => setSectorId(event.target.value)}
            required
            disabled={sectores.length === 0}
          >
            <option value="">{RIEGO.elegirSector}</option>
            {sectores.map((sector) => (
              <option key={sector.id} value={sector.id}>
                {sector.nombre}
              </option>
            ))}
          </select>
        </label>
        {sectores.length === 0 && !cargando && <p className="empty">{RIEGO.sinSectores}</p>}
        <label className="field" htmlFor="riego-turno">
          {RIEGO.turno}
          <select id="riego-turno" name="turno" value={turno} onChange={(event) => setTurno(event.target.value)}>
            <option value="manana">{RIEGO.manana}</option>
            <option value="tarde">{RIEGO.tarde}</option>
          </select>
        </label>
        <label className="field" htmlFor="riego-fecha">
          {RIEGO.fecha}
          <FechaCampo id="riego-fecha" value={fecha} onChange={setFecha} required />
        </label>
        <label className="field" htmlFor="riego-nota">
          {RIEGO.nota}
          <input id="riego-nota" name="nota" value={nota} onChange={(event) => setNota(event.target.value)} />
        </label>
        <button type="submit" className="primary" disabled={sectores.length === 0}>
          {RIEGO.registrar}
        </button>
      </form>}
    </section>
  )
}
