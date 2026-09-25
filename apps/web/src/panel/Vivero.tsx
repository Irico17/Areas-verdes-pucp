import { useMemo, useState } from "react"

export type ViveroItem = {
  id: string
  fecha: string
  area: string
  subproceso: string
  etapa: string
  descripcion: string
  observaciones: string
  responsables: string
  lugar: string
}

export const AREAS = ["Fauna", "Flora", "Ambiental", "Otros"] as const

const VACIO: ViveroItem = {
  id: "",
  fecha: "",
  area: "Flora",
  subproceso: "",
  etapa: "",
  descripcion: "",
  observaciones: "",
  responsables: "",
  lugar: "",
}

export function validarVivero(item: ViveroItem, catalogo: { subproceso: string[]; etapa: string[] }): { campo: string; motivo: string }[] {
  const fallos: { campo: string; motivo: string }[] = []
  if (item.area && !AREAS.includes(item.area as (typeof AREAS)[number])) {
    fallos.push({ campo: "area", motivo: "El área tiene que estar en el catálogo." })
  }
  if (item.subproceso && catalogo.subproceso.length > 0 && !catalogo.subproceso.includes(item.subproceso)) {
    fallos.push({ campo: "subproceso", motivo: "El subproceso se da de alta al importar; después no es texto libre." })
  }
  if (item.etapa && catalogo.etapa.length > 0 && !catalogo.etapa.includes(item.etapa)) {
    fallos.push({ campo: "etapa", motivo: "La etapa se da de alta al importar; después no es texto libre." })
  }
  return fallos
}

type Props = {
  iniciales?: ViveroItem[]
  subprocesos?: string[]
  etapas?: string[]
  delta?: string
}

export function ViveroPanel(props: Props) {
  const [mes, setMes] = useState("")
  const [items, setItems] = useState<ViveroItem[]>(props.iniciales ?? [])
  const [form, setForm] = useState<ViveroItem>(VACIO)
  const [aviso, setAviso] = useState("")
  const catalogo = { subproceso: props.subprocesos ?? [], etapa: props.etapas ?? [] }
  const visibles = useMemo(
    () => items.filter((item) => !mes || item.fecha.startsWith(mes)),
    [items, mes],
  )
  function set<K extends keyof ViveroItem>(key: K, value: ViveroItem[K]) {
    setForm((current) => ({ ...current, [key]: value }))
  }
  return (
    <section className="block">
      <h2>Vivero</h2>
      <p className="lede">{props.delta || "Registros por mes. El marcador del mapa es el lugar, no cada fila."}</p>
      <label className="field">
        Mes
        <input type="month" value={mes} onChange={(event) => setMes(event.target.value)} />
      </label>
      <ul className="labor-list">
        {visibles.length === 0 && <li className="empty">No hay registros de vivero en este mes.</li>}
        {visibles.map((item) => (
          <li key={item.id} className="agenda">
            <strong>{item.area || "Sin área"}</strong>
            <small>
              {item.fecha || "sin fecha"} · {item.lugar || "sin lugar"} · {item.subproceso || "sin subproceso"}
            </small>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          const fallos = validarVivero(form, catalogo)
          if (fallos.length) {
            setAviso(fallos[0].motivo)
            return
          }
          const guardado = { ...form, id: form.id || crypto.randomUUID() }
          setItems((current) => [...current, guardado])
          setAviso("Registro de vivero guardado.")
          setForm(VACIO)
        }}
      >
        <label className="field">
          Fecha
          <input type="date" value={form.fecha} onChange={(event) => set("fecha", event.target.value)} />
        </label>
        <label className="field">
          Área
          <select value={form.area} onChange={(event) => set("area", event.target.value)}>
            {AREAS.map((area) => (
              <option key={area}>{area}</option>
            ))}
          </select>
        </label>
        <label className="field">
          Subproceso
          {catalogo.subproceso.length > 0 ? (
            <select value={form.subproceso} onChange={(event) => set("subproceso", event.target.value)}>
              <option value="">Elegir</option>
              {catalogo.subproceso.map((item) => (
                <option key={item}>{item}</option>
              ))}
            </select>
          ) : (
            <input value={form.subproceso} onChange={(event) => set("subproceso", event.target.value)} />
          )}
        </label>
        <label className="field">
          Etapa
          {catalogo.etapa.length > 0 ? (
            <select value={form.etapa} onChange={(event) => set("etapa", event.target.value)}>
              <option value="">Elegir</option>
              {catalogo.etapa.map((item) => (
                <option key={item}>{item}</option>
              ))}
            </select>
          ) : (
            <input value={form.etapa} onChange={(event) => set("etapa", event.target.value)} />
          )}
        </label>
        <label className="field">
          Descripción
          <textarea value={form.descripcion} rows={2} onChange={(event) => set("descripcion", event.target.value)} />
        </label>
        <label className="field">
          Responsables
          <input value={form.responsables} onChange={(event) => set("responsables", event.target.value)} placeholder="Nombres ficticios, separados por coma" />
        </label>
        <label className="field">
          Observaciones
          <textarea value={form.observaciones} rows={2} onChange={(event) => set("observaciones", event.target.value)} />
        </label>
        <label className="field">
          Lugar
          <input value={form.lugar} onChange={(event) => set("lugar", event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Guardar registro
        </button>
        {aviso && <p className="hint">{aviso}</p>}
      </form>
    </section>
  )
}
