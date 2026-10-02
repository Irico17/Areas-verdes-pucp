import { useEffect, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFecha, hoyISO } from "../fecha"
import { Esqueleto } from "../ui/Esqueleto"
import { ApiError } from "../operacion"
import { crearRiego, fetchRiego } from "../producto"
import { COLA, RIEGO, etiquetaZonaSupervision } from "../ui/nomenclatura"
import { encolarRegistro, type QueuedRegistro } from "../offline/queue"
import { COLA_VACIADA, pendientesDe, rutaRiego, vaciarRegistros, type DetalleCola } from "../offline/registros"

type RiegoLocal = { id: string; sector: string; turno: string; fecha: string; nota: string }

function riegoLocal(item: QueuedRegistro): RiegoLocal | null {
  if (!item.body || typeof item.body !== "object") return null
  const row = item.body as Partial<RiegoLocal>
  if (!row.id || !row.sector) return null
  return { id: row.id, sector: row.sector, turno: row.turno ?? "", fecha: row.fecha ?? "", nota: row.nota ?? "" }
}

export function RiegoPanel(props: { capatazId: string }) {
  const [aviso, setAviso] = useState("")
  const [rows, setRows] = useState<Awaited<ReturnType<typeof fetchRiego>>["registros"]>([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState("")
  const [sector, setSector] = useState("")
  const [zona, setZona] = useState("Z1")
  const [turno, setTurno] = useState("manana")
  const [fecha, setFecha] = useState(hoyISO)
  const [nota, setNota] = useState("")
  const [locales, setLocales] = useState<RiegoLocal[]>([])
  const [avisoCola, setAvisoCola] = useState("")

  async function load() {
    try {
      const body = await fetchRiego()
      setAviso(body.aviso)
      setRows(body.registros)
      setError("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo leer el riego")
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
        setAviso(body.aviso)
        setRows(body.registros)
        setError("")
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "No se pudo leer el riego")
      })
      .finally(() => {
        if (!cancelled) setCargando(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <section className="block">
      <h2>{RIEGO.titulo}</h2>
      <p className="lede">{aviso || RIEGO.lede}</p>
      <p className="hint">
        {rows.length} {rows.length === 1 ? "turno registrado" : "turnos registrados"}. Cobertura: definición pendiente.
      </p>
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
      {!cargando && rows.length === 0 && !error && <p className="empty">Todavía no hay turnos registrados.</p>}
      <ul className="labor-list">
        {locales.map((row) => (
          <li key={row.id} className="agenda">
            <strong>{row.sector}</strong> <span className="marca-cola">{COLA.marca}</span>
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
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          if (event.currentTarget.querySelector('[aria-invalid="true"]')) {
            setError("Corrija la fecha antes de registrar.")
            return
          }
          const id = crypto.randomUUID()
          const cuerpo = {
            id,
            sector,
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
              if (err instanceof ApiError && err.status === 0) {
                void encolarRegistro({
                  id,
                  tipo: "riego",
                  path: rutaRiego(),
                  method: "POST",
                  body: cuerpo,
                  createdAt: new Date().toISOString(),
                }).then(() => {
                  setLocales((actual) => [...actual.filter((fila) => fila.id !== id), { id, sector, turno, fecha, nota }])
                  setNota("")
                  setError("")
                  setAvisoCola(RIEGO.enCola)
                })
                return
              }
              setError(err instanceof Error ? err.message : "No se pudo registrar")
            })
        }}
      >
        <label className="field">
          {RIEGO.zona}
          <select value={zona} onChange={(event) => setZona(event.target.value)}>
            {["Z1", "Z2", "Z3", "Z4"].map((codigo) => (
              <option key={codigo} value={codigo}>
                {etiquetaZonaSupervision(codigo)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          {RIEGO.sector}
          <input value={sector} onChange={(event) => setSector(event.target.value)} placeholder={RIEGO.sectorPlaceholder} required />
        </label>
        <label className="field">
          {RIEGO.turno}
          <select value={turno} onChange={(event) => setTurno(event.target.value)}>
            <option value="manana">{RIEGO.manana}</option>
            <option value="tarde">{RIEGO.tarde}</option>
          </select>
        </label>
        <label className="field">
          Fecha
          <FechaCampo value={fecha} onChange={setFecha} required />
        </label>
        <label className="field">
          Nota
          <input value={nota} onChange={(event) => setNota(event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Registrar turno
        </button>
      </form>
    </section>
  )
}
