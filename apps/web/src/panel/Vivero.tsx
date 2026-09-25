import { useEffect, useMemo, useState } from "react"
import { AREAS, VIVERO_VACIO, validarVivero, type ViveroItem } from "./vivero"

type Props = {
  iniciales?: ViveroItem[]
  subprocesos?: string[]
  etapas?: string[]
  delta?: string
}

export function ViveroPanel(props: Props) {
  const [mes, setMes] = useState("")
  const [items, setItems] = useState<ViveroItem[]>(props.iniciales ?? [])
  const [form, setForm] = useState<ViveroItem>(VIVERO_VACIO)
  const [aviso, setAviso] = useState("")
  const catalogo = { subproceso: props.subprocesos ?? [], etapa: props.etapas ?? [] }
  useEffect(() => {
    if (props.iniciales) return
    void fetch("/api/v1/vivero", { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((body: { registros?: (ViveroItem & { lugar_libre?: string })[] } | null) => {
        if (!body?.registros) return
        setItems(
          body.registros.map((item) => ({
            ...item,
            lugar: item.lugar || item.lugar_libre || "",
          })),
        )
      })
      .catch(() => {})
  }, [props.iniciales])
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
          void fetch("/api/v1/vivero", {
            method: "POST",
            credentials: "include",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              id: guardado.id,
              fecha: guardado.fecha,
              area: guardado.area,
              subproceso: guardado.subproceso,
              etapa: guardado.etapa,
              descripcion: guardado.descripcion,
              observaciones: guardado.observaciones,
              responsables: guardado.responsables,
              lugar_libre: guardado.lugar,
            }),
          })
            .then((res) => {
              if (!res.ok) {
                setAviso("No se pudo guardar el registro de vivero.")
                return
              }
              setItems((current) => [...current, guardado])
              setAviso("Registro de vivero guardado.")
              setForm(VIVERO_VACIO)
            })
            .catch(() => setAviso("Sin conexión con la API."))
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
