import { useEffect, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { formatFecha, hoyISO } from "../fecha"
import { Esqueleto } from "../ui/Esqueleto"
import { crearRiego, fetchRiego } from "../producto"
import { RIEGO, etiquetaZonaSupervision } from "../ui/nomenclatura"

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
      {error && <p className="status error">{error}</p>}
      {cargando && <Esqueleto />}
      {!cargando && rows.length === 0 && !error && <p className="empty">Todavía no hay turnos registrados.</p>}
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
            setError("Corrija la fecha antes de registrar.")
            return
          }
          void crearRiego({
            id: crypto.randomUUID(),
            sector,
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
            .catch((err: unknown) => setError(err instanceof Error ? err.message : "No se pudo registrar"))
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
