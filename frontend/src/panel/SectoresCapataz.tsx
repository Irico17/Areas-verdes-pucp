import { useEffect, useId, useState, type FormEvent } from "react"
import { apiUrl, fetchCollection } from "../api"
import { ZONIFICACION } from "../ui/nomenclatura"
import { SelectorLugar } from "./SelectorLugar"
import {
  crearSector,
  desactivarSector,
  guardarReferente,
  importarSectores,
  importarVias,
  listarEdificiosReferente,
  listarLugaresCatalogo,
  listarSectores,
  type EdificioRef,
  type LugarCatalogo,
  type SectorCapataz,
} from "./zonificacion"

export function SectoresCapataz(props: { editable?: boolean }) {
  const editable = props.editable !== false
  const baseId = useId()
  const [sectores, setSectores] = useState<SectorCapataz[]>([])
  const [lugares, setLugares] = useState<LugarCatalogo[]>([])
  const [edificios, setEdificios] = useState<EdificioRef[]>([])
  const [codigo, setCodigo] = useState("")
  const [nombre, setNombre] = useState("")
  const [color, setColor] = useState("#3f73b0")
  const [lugarId, setLugarId] = useState("")
  const [edificioId, setEdificioId] = useState("")
  const [aviso, setAviso] = useState("")
  const [error, setError] = useState("")
  const [sinCuarteles, setSinCuarteles] = useState(true)

  function pedir() {
    return Promise.all([
      listarSectores(false),
      listarLugaresCatalogo(),
      listarEdificiosReferente(),
      fetchCollection(apiUrl("/zonificacion/cuarteles")).catch(() => null),
    ] as const)
  }

  function aplicar(filas: SectorCapataz[], sitios: LugarCatalogo[], obras: EdificioRef[], cuarteles: { features: unknown[] } | null) {
    setSectores(filas)
    setLugares(sitios)
    setEdificios(obras)
    setSinCuarteles(!cuarteles || cuarteles.features.length === 0)
  }

  async function cargar() {
    const [filas, sitios, obras, cuarteles] = await pedir()
    aplicar(filas, sitios, obras, cuarteles)
  }

  useEffect(() => {
    let vivo = true
    pedir()
      .then(([filas, sitios, obras, cuarteles]) => {
        if (vivo) aplicar(filas, sitios, obras, cuarteles)
      })
      .catch((err: unknown) => {
        if (vivo) setError(err instanceof Error ? err.message : ZONIFICACION.vacio)
      })
    return () => {
      vivo = false
    }
  }, [])

  async function onAlta(event: FormEvent) {
    event.preventDefault()
    setError("")
    setAviso("")
    try {
      await crearSector({ codigo: codigo.trim().toLowerCase(), nombre: nombre.trim(), color })
      setCodigo("")
      setNombre("")
      setAviso(ZONIFICACION.guardado)
      await cargar()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : ZONIFICACION.duplicado)
    }
  }

  async function onArchivoSectores(file: File) {
    setError("")
    const texto = await file.text()
    const cuerpo = JSON.parse(texto) as { sectores?: { codigo: string; nombre: string; color: string }[] }
    const filas = Array.isArray(cuerpo) ? cuerpo : cuerpo.sectores
    if (!Array.isArray(filas) || filas.length === 0) throw new Error(ZONIFICACION.vacio)
    await importarSectores(filas)
    setAviso(ZONIFICACION.guardado)
    await cargar()
  }

  async function onArchivoVias(file: File) {
    setError("")
    const texto = await file.text()
    await importarVias(JSON.parse(texto) as unknown)
    setAviso(ZONIFICACION.viasLede)
  }

  return (
    <section className="sectores-capataz" aria-labelledby={`${baseId}-titulo`}>
      <header className="sectores-cabecera">
        <h2 id={`${baseId}-titulo`}>{ZONIFICACION.titulo}</h2>
        <p className="lede">{ZONIFICACION.lede}</p>
      </header>
      <ul className="sector-lista">
        {sectores.map((sector) => (
          <li key={sector.codigo} className={sector.activo ? "sector-fila" : "sector-fila apagado"}>
            <span className="sector-muestra" style={{ background: sector.color }} aria-hidden="true" />
            <span>
              <strong>{sector.nombre}</strong>
              <small>
                {sector.codigo}
                {sector.activo ? "" : ` · ${ZONIFICACION.inactivo}`}
              </small>
            </span>
            {editable && sector.activo && (
              <button
                type="button"
                className="link"
                onClick={() => {
                  void desactivarSector(sector.codigo)
                    .then(() => cargar())
                    .catch((err: unknown) => setError(err instanceof Error ? err.message : ZONIFICACION.desactivar))
                }}
              >
                {ZONIFICACION.desactivar}
              </button>
            )}
          </li>
        ))}
      </ul>
      {sectores.filter((sector) => sector.activo).length === 0 && !error && <p className="empty">{ZONIFICACION.vacio}</p>}
      {editable && <form className="form sector-alta" onSubmit={(event) => void onAlta(event)}>
        <label className="field" htmlFor={`${baseId}-codigo`}>
          {ZONIFICACION.codigo}
          <input id={`${baseId}-codigo`} name="codigo" value={codigo} required onChange={(event) => setCodigo(event.target.value)} />
        </label>
        <label className="field" htmlFor={`${baseId}-nombre`}>
          {ZONIFICACION.nombre}
          <input id={`${baseId}-nombre`} name="nombre" value={nombre} required onChange={(event) => setNombre(event.target.value)} />
        </label>
        <label className="field sector-color" htmlFor={`${baseId}-color`}>
          {ZONIFICACION.color}
          <input id={`${baseId}-color`} name="color" type="color" value={color} onChange={(event) => setColor(event.target.value)} />
        </label>
        <button type="submit" className="primary">
          {ZONIFICACION.alta}
        </button>
      </form>}
      {editable && <div className="sector-archivos">
        <label className="field">
          {ZONIFICACION.importarSectores}
          <input
            type="file"
            accept="application/json,.json"
            onChange={(event) => {
              const file = event.target.files?.[0]
              if (!file) return
              void onArchivoSectores(file).catch((err: unknown) => setError(err instanceof Error ? err.message : ZONIFICACION.importarSectores))
            }}
          />
        </label>
        <label className="field">
          {ZONIFICACION.importarVias}
          <input
            type="file"
            accept=".geojson,application/geo+json,application/json"
            onChange={(event) => {
              const file = event.target.files?.[0]
              if (!file) return
              void onArchivoVias(file).catch((err: unknown) => setError(err instanceof Error ? err.message : ZONIFICACION.importarVias))
            }}
          />
          <small>{ZONIFICACION.viasLede}</small>
        </label>
      </div>}
      <aside className="cuarteles-vacio">
        <h3>{ZONIFICACION.cuarteles}</h3>
        {sinCuarteles && <p>{ZONIFICACION.sinCuarteles}</p>}
      </aside>
      {editable && <form
        className="form referente"
        onSubmit={(event) => {
          event.preventDefault()
          if (!lugarId || !edificioId) return
          void guardarReferente(Number(lugarId), edificioId)
            .then(() => setAviso(ZONIFICACION.guardado))
            .catch((err: unknown) => setError(err instanceof Error ? err.message : ZONIFICACION.edificio))
        }}
      >
        <SelectorLugar id={`${baseId}-lugar`} lugares={lugares} lugarId={lugarId} onChange={setLugarId} />
        <label className="field" htmlFor={`${baseId}-edificio`}>
          {ZONIFICACION.edificio}
          <select id={`${baseId}-edificio`} name="edificio_id" value={edificioId} onChange={(event) => setEdificioId(event.target.value)}>
            <option value="">{ZONIFICACION.sinEdificio}</option>
            {edificios.map((edificio) => (
              <option key={edificio.id} value={edificio.id}>
                {edificio.nombre ? `${edificio.nombre} · ${edificio.id}` : edificio.id}
              </option>
            ))}
          </select>
        </label>
        <button type="submit" disabled={!lugarId || !edificioId}>
          {ZONIFICACION.guardarReferente}
        </button>
      </form>}
      {error && (
        <p className="status error" role="alert">
          {error}
        </p>
      )}
      {aviso && <p className="banner">{aviso}</p>}
    </section>
  )
}
