import { useEffect, useMemo, useState } from "react"
import { fetchCollection } from "./api"
import { CampusMap } from "./map/CampusMap"
import { readRol, writeRol } from "./session"
import { LAYERS, ROLES, type FeatureCollection, type LayerId, type Rol } from "./types"

type LoadState =
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "ready" }

export default function App() {
  const [rol, setRol] = useState<Rol>(() => readRol())
  const [visible, setVisible] = useState<Record<LayerId, boolean>>(() =>
    Object.fromEntries(LAYERS.map((layer) => [layer.id, layer.defaultOn])) as Record<LayerId, boolean>,
  )
  const [data, setData] = useState<Partial<Record<LayerId, FeatureCollection>>>({})
  const [load, setLoad] = useState<LoadState>({ kind: "loading" })
  const [railOpen, setRailOpen] = useState(false)
  const [picked, setPicked] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    Promise.all(LAYERS.map(async (layer) => [layer.id, await fetchCollection(layer.path)] as const))
      .then((rows) => {
        if (cancelled) return
        setData(Object.fromEntries(rows) as Partial<Record<LayerId, FeatureCollection>>)
        setLoad({ kind: "ready" })
      })
      .catch((error: unknown) => {
        if (cancelled) return
        const message = error instanceof Error ? error.message : "No se pudo leer el catastro"
        setLoad({ kind: "error", message })
      })
    return () => {
      cancelled = true
    }
  }, [])

  const role = ROLES.find((item) => item.id === rol) ?? ROLES[1]
  const summary = useMemo(() => {
    const areas = data.areas?.features.length
    const zonas = data.zonas?.features.length
    if (areas == null || zonas == null) return "Leyendo catastro…"
    return `${areas} áreas · ${zonas} zonas`
  }, [data])

  return (
    <div className="shell">
      <CampusMap data={data} visible={visible} onSelect={(hit) => setPicked(hit ? `${hit.layer}: ${String(hit.props.nombre || hit.props.feature_id || "polígono")}` : null)} />
      <header className="topbar">
        <div className="brand">
          <strong>Campus Verde</strong>
          <span>PUCP Pando · supervisión</span>
        </div>
        <button type="button" className="menu-btn" onClick={() => setRailOpen((open) => !open)}>
          Capas
        </button>
        <div className="top-spacer" />
        <div className="roles" role="group" aria-label="Rol de consulta">
          {ROLES.map((item) => (
            <button
              key={item.id}
              type="button"
              aria-pressed={item.id === rol}
              onClick={() => {
                setRol(item.id)
                writeRol(item.id)
              }}
            >
              {item.label}
            </button>
          ))}
        </div>
      </header>
      <aside className={railOpen ? "rail open" : "rail"}>
        <h2>Catastro</h2>
        <p className="lede">Polígonos servidos por la API en EPSG:4326. La base es OpenStreetMap.</p>
        <p className="role-note">{role.note}</p>
        <p className={load.kind === "error" ? "status error" : "status"}>
          {load.kind === "error" ? load.message : summary}
          {picked ? ` · ${picked}` : ""}
        </p>
        {LAYERS.map((layer) => (
          <div className="layer" key={layer.id}>
            <span className="swatch" style={{ background: layer.fill }} />
            <label>
              <input
                type="checkbox"
                checked={visible[layer.id]}
                onChange={() => setVisible((current) => ({ ...current, [layer.id]: !current[layer.id] }))}
              />{" "}
              {layer.label}
              <small>{layer.hint}</small>
            </label>
            <span className="count">{data[layer.id]?.features.length ?? "—"}</span>
          </div>
        ))}
        <p className="foot">
          Sin Street View. El rol es un interruptor local: no hay sesión institucional.
          {load.kind === "loading" ? " Esperando /api/v1/geo." : ""}
        </p>
      </aside>
    </div>
  )
}
