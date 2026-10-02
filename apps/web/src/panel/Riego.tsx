import { useEffect, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFecha, hoyISO } from "../fecha"
import { Esqueleto } from "../ui/Esqueleto"
import { crearRiego, fetchRiego } from "../producto"
import { RIEGO, etiquetaZonaSupervision } from "../ui/nomenclatura"
import { listarSectores, type SectorCapataz } from "./zonificacion"

export function RiegoPanel(props: { capatazId: string }) {
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
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && rows.length === 0 && !error && <p className="empty">{RIEGO.vacio}</p>}
      <ul className="labor-list">
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
            setError(RIEGO.errorFecha)
            return
          }
          const idSector = Number(sectorId)
          if (!Number.isInteger(idSector) || idSector <= 0) {
            setError(RIEGO.elegirSector)
            return
          }
          void crearRiego({
            id: crypto.randomUUID(),
            sector_id: idSector,
            turno,
            capataz_id: props.capatazId,
            fecha,
            nota,
            zona_supervision_id: zona,
          })
            .then(() => {
              setNota("")
              return load()
            })
            .catch((err: unknown) => setError(err instanceof Error ? err.message : RIEGO.errorRegistrar))
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
      </form>
    </section>
  )
}
