import { useEffect, useMemo, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { Esqueleto } from "../ui/Esqueleto"
import { PODA_VACIA, codigoExterno, validarPoda, type PodaItem } from "./poda"
import { encolarRegistro } from "../offline/queue"
import { COLA_VACIADA, pendientesDe, rutaPoda, vaciarRegistros, type DetalleCola } from "../offline/registros"
import { COLA, PODA } from "../ui/nomenclatura"

type Props = {
  iniciales?: PodaItem[]
  onGuardar?: (item: PodaItem) => void
}

export function PodaPanel(props: Props) {
  const [items, setItems] = useState<PodaItem[]>(props.iniciales ?? [])
  const [cargando, setCargando] = useState(!props.iniciales)
  const [form, setForm] = useState<PodaItem>({ ...PODA_VACIA, id: "nueva" })
  const [aviso, setAviso] = useState("")
  const [enCola, setEnCola] = useState<string[]>([])
  const fallos = useMemo(() => validarPoda(form), [form])
  useEffect(() => {
    if (props.iniciales) return
    let cancelado = false
    void fetch(apiUrl("/podas"), { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((body: { podas?: PodaItem[] } | null) => {
        if (!cancelado && body?.podas) setItems(body.podas)
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelado) setCargando(false)
      })
    return () => {
      cancelado = true
    }
  }, [props.iniciales])
  useEffect(() => {
    let vivo = true
    const leer = (conflictos = false) => {
      void pendientesDe("poda")
        .then((filas) => {
          if (!vivo) return
          setEnCola(filas.map((fila) => fila.id))
          if (conflictos) setAviso(COLA.conflicto)
          else if (filas.length === 0) setAviso((actual) => (actual === PODA.enCola ? "" : actual))
        })
        .catch(() => {})
    }
    leer()
    const alVolver = (ev: Event) => {
      const detail = (ev as CustomEvent<DetalleCola>).detail
      leer(detail?.conflictos?.some((item) => item.tipo === "poda") ?? false)
    }
    window.addEventListener(COLA_VACIADA, alVolver)
    return () => {
      vivo = false
      window.removeEventListener(COLA_VACIADA, alVolver)
    }
  }, [])
  function set<K extends keyof PodaItem>(key: K, value: PodaItem[K]) {
    setForm((current) => ({ ...current, [key]: value }))
  }
  return (
    <section className="block">
      <h2>Poda</h2>
      <p className="lede">{PODA.lede}</p>
      {enCola.length > 0 && (
        <p className="hint" role="status">
          {COLA.local(enCola.length)} <span className="marca-cola">{COLA.marca}</span>{" "}
          <button type="button" className="link" onClick={() => void vaciarRegistros()}>
            {COLA.reintentar}
          </button>
        </p>
      )}
      {cargando && <Esqueleto />}
      <ul className="labor-list">
        {!cargando && items.length === 0 && <li className="empty">No hay podas en esta vista.</li>}
        {items.map((item) => (
          <li key={item.codigo}>
            <button type="button" className="labor" onClick={() => setForm(item)}>
              <span>
                <strong>{item.codigo}</strong>
                <small>
                  {item.ubicacion || "Sin ubicación"} · {item.prioridad}
                  {item.codigo_externo ? ` · ${item.codigo_externo}` : ""}
                  {enCola.includes(item.id) ? ` · ${COLA.marca}` : ""}
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
          if (event.currentTarget.querySelector('[aria-invalid="true"]')) {
            setAviso("Corrija la fecha antes de guardar.")
            return
          }
          const errores = validarPoda(form)
          if (errores.length) {
            setAviso(errores[0].motivo)
            return
          }
          const externo = codigoExterno(form.codigo_externo).codigo
          const nueva = !form.id || form.id === "nueva"
          const guardada = { ...form, codigo_externo: externo, id: nueva ? crypto.randomUUID() : form.id }
          const cuerpo = {
            ...guardada,
            cantidad_pedida: Number(guardada.cantidad_pedida),
            cantidad_ejecutada: Number(guardada.cantidad_ejecutada),
          }
          const ruta = rutaPoda(guardada.id, nueva)
          void fetch(ruta, {
            method: nueva ? "POST" : "PATCH",
            credentials: "include",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(cuerpo),
          })
            .then((res) => {
              if (res.status === 409) {
                setAviso(COLA.conflicto)
                return
              }
              if (!res.ok) {
                setAviso(PODA.error)
                return
              }
              setItems((current) => {
                const sin = current.filter((item) => item.codigo !== guardada.codigo)
                return [...sin, guardada]
              })
              setEnCola((actual) => actual.filter((id) => id !== guardada.id))
              props.onGuardar?.(guardada)
              setAviso(PODA.guardada)
            })
            .catch(() => {
              void encolarRegistro({
                id: guardada.id,
                tipo: "poda",
                path: ruta,
                method: nueva ? "POST" : "PATCH",
                body: cuerpo,
                createdAt: new Date().toISOString(),
              }).then(() => {
                setItems((current) => {
                  const sin = current.filter((item) => item.codigo !== guardada.codigo && item.id !== guardada.id)
                  return [...sin, guardada]
                })
                setEnCola((actual) => [...actual.filter((id) => id !== guardada.id), guardada.id])
                setAviso(PODA.enCola)
              })
            })
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
          <FechaCampo value={form.fecha_reporte} onChange={(iso) => set("fecha_reporte", iso)} />
        </label>
        <label className="field">
          Fecha de ejecución
          <FechaCampo value={form.fecha_ejecucion} onChange={(iso) => set("fecha_ejecucion", iso)} />
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
