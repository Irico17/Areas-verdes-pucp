import { useEffect, useMemo, useState } from "react"
import { FechaCampo } from "../FechaCampo"
import { Esqueleto } from "../ui/Esqueleto"
import { AREAS, VIVERO_VACIO, validarVivero, type ViveroItem } from "./vivero"
import { apiUrl } from "../api"
import { encolarRegistro } from "../offline/queue"
import { COLA_VACIADA, pendientesDe, rutaVivero, vaciarRegistros, type DetalleCola } from "../offline/registros"
import { COLA, VIVERO } from "../ui/nomenclatura"

type Props = {
  iniciales?: ViveroItem[]
  subprocesos?: string[]
  etapas?: string[]
  delta?: string
}

export function ViveroPanel(props: Props) {
  const [mes, setMes] = useState("")
  const [items, setItems] = useState<ViveroItem[]>(props.iniciales ?? [])
  const [cargando, setCargando] = useState(!props.iniciales)
  const [form, setForm] = useState<ViveroItem>(VIVERO_VACIO)
  const [aviso, setAviso] = useState("")
  const [enCola, setEnCola] = useState<string[]>([])
  const catalogo = { subproceso: props.subprocesos ?? [], etapa: props.etapas ?? [] }
  useEffect(() => {
    if (props.iniciales) return
    let cancelado = false
    void fetch(apiUrl("/vivero"), { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((body: { registros?: (ViveroItem & { lugar_libre?: string })[] } | null) => {
        if (cancelado || !body?.registros) return
        setItems(
          body.registros.map((item) => ({
            ...item,
            lugar: item.lugar || item.lugar_libre || "",
          })),
        )
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
      void pendientesDe("vivero")
        .then((filas) => {
          if (!vivo) return
          setEnCola(filas.map((fila) => fila.id))
          if (conflictos) setAviso(COLA.conflicto)
          else if (filas.length === 0) setAviso((actual) => (actual === VIVERO.enCola ? "" : actual))
        })
        .catch(() => {})
    }
    leer()
    const alVolver = (ev: Event) => {
      const detail = (ev as CustomEvent<DetalleCola>).detail
      leer(detail?.conflictos?.some((item) => item.tipo === "vivero") ?? false)
    }
    window.addEventListener(COLA_VACIADA, alVolver)
    return () => {
      vivo = false
      window.removeEventListener(COLA_VACIADA, alVolver)
    }
  }, [])
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
        {!cargando && visibles.length === 0 && <li className="empty">No hay registros de vivero en este mes.</li>}
        {visibles.map((item) => (
          <li key={item.id} className="agenda">
            <strong>{item.area || "Sin área"}</strong>
            <small>
              {item.fecha || "sin fecha"} · {item.lugar || "sin lugar"} · {item.subproceso || "sin subproceso"}
              {enCola.includes(item.id) ? ` · ${COLA.marca}` : ""}
            </small>
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
          const fallos = validarVivero(form, catalogo)
          if (fallos.length) {
            setAviso(fallos[0].motivo)
            return
          }
          const guardado = { ...form, id: form.id || crypto.randomUUID() }
          const cuerpo = {
            id: guardado.id,
            fecha: guardado.fecha,
            area: guardado.area,
            subproceso: guardado.subproceso,
            etapa: guardado.etapa,
            descripcion: guardado.descripcion,
            observaciones: guardado.observaciones,
            responsables: guardado.responsables,
            lugar_libre: guardado.lugar,
          }
          void fetch(rutaVivero(), {
            method: "POST",
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
                setAviso(VIVERO.error)
                return
              }
              setItems((current) => [...current.filter((item) => item.id !== guardado.id), guardado])
              setEnCola((actual) => actual.filter((id) => id !== guardado.id))
              setAviso(VIVERO.guardado)
              setForm(VIVERO_VACIO)
            })
            .catch(() => {
              void encolarRegistro({
                id: guardado.id,
                tipo: "vivero",
                path: rutaVivero(),
                method: "POST",
                body: cuerpo,
                createdAt: new Date().toISOString(),
              }).then(() => {
                setItems((current) => [...current.filter((item) => item.id !== guardado.id), guardado])
                setEnCola((actual) => [...actual.filter((id) => id !== guardado.id), guardado.id])
                setAviso(VIVERO.enCola)
              })
            })
        }}
      >
        <label className="field">
          Fecha
          <FechaCampo value={form.fecha} onChange={(iso) => set("fecha", iso)} />
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
