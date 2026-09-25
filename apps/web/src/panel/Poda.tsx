import { useMemo, useState } from "react"

export type PodaItem = {
  id: string
  codigo: string
  codigo_externo: string
  tipo: string
  tipo_actividad: string
  fecha_reporte: string
  fecha_ejecucion: string
  personal: string
  ubicacion: string
  unidad: string
  cantidad_pedida: string
  cantidad_ejecutada: string
  prioridad: string
  comentario: string
  nombre_comun: string
  nombre_cientifico: string
}

export type Fallo = { campo: string; motivo: string }

const VACIA: PodaItem = {
  id: "",
  codigo: "",
  codigo_externo: "",
  tipo: "",
  tipo_actividad: "",
  fecha_reporte: "",
  fecha_ejecucion: "",
  personal: "",
  ubicacion: "",
  unidad: "",
  cantidad_pedida: "0",
  cantidad_ejecutada: "0",
  prioridad: "media",
  comentario: "",
  nombre_comun: "",
  nombre_cientifico: "",
}

export function codigoExterno(valor: string): { codigo: string; fallo?: Fallo } {
  const v = valor.trim()
  if (!v || /^(aun no tiene codigo|no aplica|sin codigo)$/i.test(v)) return { codigo: "" }
  if (/^OSG-/i.test(v)) return { codigo: v.toUpperCase() }
  return { codigo: "", fallo: { campo: "codigo_externo", motivo: "El código OSG no se genera. Déjelo vacío o copie el que ya trae la fuente." } }
}

export function validarPoda(item: PodaItem): Fallo[] {
  const fallos: Fallo[] = []
  if (!/^PO-\d+$/.test(item.codigo.trim())) fallos.push({ campo: "codigo", motivo: "El código es PO-n." })
  const externo = codigoExterno(item.codigo_externo)
  if (externo.fallo) fallos.push(externo.fallo)
  if (item.fecha_reporte && item.fecha_ejecucion && item.fecha_ejecucion < item.fecha_reporte) {
    fallos.push({ campo: "fecha_ejecucion", motivo: "La ejecución no puede ser anterior al reporte." })
  }
  for (const campo of ["cantidad_pedida", "cantidad_ejecutada"] as const) {
    const n = Number(item[campo])
    if (!Number.isFinite(n) || n < 0) fallos.push({ campo, motivo: "La cantidad es cero o más." })
  }
  if (!["baja", "media", "alta"].includes(item.prioridad)) fallos.push({ campo: "prioridad", motivo: "Prioridad no catalogada." })
  return fallos
}

type Props = {
  iniciales?: PodaItem[]
  onGuardar?: (item: PodaItem) => void
}

export function PodaPanel(props: Props) {
  const [items, setItems] = useState<PodaItem[]>(props.iniciales ?? [])
  const [form, setForm] = useState<PodaItem>({ ...VACIA, id: "nueva" })
  const [aviso, setAviso] = useState("")
  const fallos = useMemo(() => validarPoda(form), [form])
  function set<K extends keyof PodaItem>(key: K, value: PodaItem[K]) {
    setForm((current) => ({ ...current, [key]: value }))
  }
  return (
    <section className="block">
      <h2>Poda</h2>
      <p className="lede">Incidencias de poda. El código externo solo se conserva si ya viene como OSG.</p>
      <ul className="labor-list">
        {items.length === 0 && <li className="empty">No hay podas en esta vista.</li>}
        {items.map((item) => (
          <li key={item.codigo}>
            <button type="button" className="labor" onClick={() => setForm(item)}>
              <span>
                <strong>{item.codigo}</strong>
                <small>
                  {item.ubicacion || "Sin ubicación"} · {item.prioridad}
                  {item.codigo_externo ? ` · ${item.codigo_externo}` : ""}
                </small>
              </span>
            </button>
          </li>
        ))}
      </ul>
      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          const errores = validarPoda(form)
          if (errores.length) {
            setAviso(errores[0].motivo)
            return
          }
          const externo = codigoExterno(form.codigo_externo).codigo
          const guardada = { ...form, codigo_externo: externo, id: form.id || crypto.randomUUID() }
          setItems((current) => {
            const sin = current.filter((item) => item.codigo !== guardada.codigo)
            return [...sin, guardada]
          })
          props.onGuardar?.(guardada)
          setAviso("Poda guardada. No se creó un código OSG.")
        }}
      >
        <label className="field">
          Código
          <input value={form.codigo} onChange={(event) => set("codigo", event.target.value)} required />
        </label>
        <label className="field">
          Código externo
          <input value={form.codigo_externo} onChange={(event) => set("codigo_externo", event.target.value)} placeholder="OSG-… si ya existe" />
        </label>
        <label className="field">
          Tipo
          <input value={form.tipo} onChange={(event) => set("tipo", event.target.value)} />
        </label>
        <label className="field">
          Tipo de actividad
          <input value={form.tipo_actividad} onChange={(event) => set("tipo_actividad", event.target.value)} />
        </label>
        <label className="field">
          Fecha de reporte
          <input type="date" value={form.fecha_reporte} onChange={(event) => set("fecha_reporte", event.target.value)} />
        </label>
        <label className="field">
          Fecha de ejecución
          <input type="date" value={form.fecha_ejecucion} onChange={(event) => set("fecha_ejecucion", event.target.value)} />
        </label>
        <label className="field">
          Personal
          <input value={form.personal} onChange={(event) => set("personal", event.target.value)} />
        </label>
        <label className="field">
          Ubicación
          <input value={form.ubicacion} onChange={(event) => set("ubicacion", event.target.value)} />
        </label>
        <label className="field">
          Unidad
          <input value={form.unidad} onChange={(event) => set("unidad", event.target.value)} />
        </label>
        <label className="field">
          Cantidad pedida
          <input value={form.cantidad_pedida} onChange={(event) => set("cantidad_pedida", event.target.value)} />
        </label>
        <label className="field">
          Cantidad ejecutada
          <input value={form.cantidad_ejecutada} onChange={(event) => set("cantidad_ejecutada", event.target.value)} />
        </label>
        <label className="field">
          Prioridad
          <select value={form.prioridad} onChange={(event) => set("prioridad", event.target.value)}>
            <option value="baja">Baja</option>
            <option value="media">Media</option>
            <option value="alta">Alta</option>
          </select>
        </label>
        <label className="field">
          Nombre común
          <input value={form.nombre_comun} onChange={(event) => set("nombre_comun", event.target.value)} />
        </label>
        <label className="field">
          Nombre científico
          <input value={form.nombre_cientifico} onChange={(event) => set("nombre_cientifico", event.target.value)} />
        </label>
        <label className="field">
          Comentario
          <textarea value={form.comentario} rows={2} onChange={(event) => set("comentario", event.target.value)} />
        </label>
        <button type="submit" className="primary">
          Guardar poda
        </button>
        {aviso && <p className="hint">{aviso}</p>}
        {fallos.length > 0 && form.codigo !== "" && <p className="status error">{fallos[0].motivo}</p>}
      </form>
    </section>
  )
}
