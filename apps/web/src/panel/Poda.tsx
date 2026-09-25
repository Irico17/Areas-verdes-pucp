import { useEffect, useMemo, useState } from "react"
import { PODA_VACIA, codigoExterno, validarPoda, type PodaItem } from "./poda"

type Props = {
  iniciales?: PodaItem[]
  onGuardar?: (item: PodaItem) => void
}

export function PodaPanel(props: Props) {
  const [items, setItems] = useState<PodaItem[]>(props.iniciales ?? [])
  const [form, setForm] = useState<PodaItem>({ ...PODA_VACIA, id: "nueva" })
  const [aviso, setAviso] = useState("")
  const fallos = useMemo(() => validarPoda(form), [form])
  useEffect(() => {
    if (props.iniciales) return
    void fetch("/api/v1/podas", { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((body: { podas?: PodaItem[] } | null) => {
        if (body?.podas) setItems(body.podas)
      })
      .catch(() => {})
  }, [props.iniciales])
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
          const nueva = !form.id || form.id === "nueva"
          const guardada = { ...form, codigo_externo: externo, id: nueva ? crypto.randomUUID() : form.id }
          const ruta = nueva ? "/api/v1/podas" : `/api/v1/podas/${guardada.id}`
          void fetch(ruta, {
            method: nueva ? "POST" : "PATCH",
            credentials: "include",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              ...guardada,
              cantidad_pedida: Number(guardada.cantidad_pedida),
              cantidad_ejecutada: Number(guardada.cantidad_ejecutada),
            }),
          })
            .then((res) => {
              if (!res.ok) {
                setAviso("No se pudo guardar la poda.")
                return
              }
              setItems((current) => {
                const sin = current.filter((item) => item.codigo !== guardada.codigo)
                return [...sin, guardada]
              })
              props.onGuardar?.(guardada)
              setAviso("Poda guardada. No se creó un código OSG.")
            })
            .catch(() => setAviso("Sin conexión con la API."))
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
