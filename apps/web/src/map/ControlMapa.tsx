import { INVENTARIO } from "../inventario"
import type { Categoria, ColorPor } from "./categorias"
import { AYUDA, CAPA, MAPA, ZONIFICACION, conteoCatastro } from "../ui/nomenclatura"
import { AyudaTip } from "../ui/AyudaTip"
import { LAYERS, type FeatureCollection, type LayerId } from "../types"

type Props = {
  relieve: boolean
  onRelieve: (valor: boolean) => void
  showEdificios: boolean
  onEdificios: () => void
  edificios: number
  catastroOn: boolean
  onCatastro: () => void
  areas: number | string
  zonas: number | string
  colorPor: ColorPor
  onColorPor: (valor: ColorPor) => void
  categorias: Categoria<string>[]
  conteo: Record<string, number>
  ocultas: string[]
  onOculta: (id: string) => void
  visible: Record<LayerId, boolean>
  onCapa: (id: LayerId) => void
  data: Partial<Record<LayerId, FeatureCollection>>
  inventoryOn: Record<string, boolean>
  onInventario: (id: string) => void
  inventory: Partial<Record<string, FeatureCollection>>
  resumen: string
}

export function ControlMapa(props: Props) {
  return (
    <aside className="control-mapa" aria-label={MAPA.capas}>
      <div className="roles" role="group" aria-label="Vista del mapa">
        <button type="button" aria-pressed={!props.relieve} onClick={() => props.onRelieve(false)}>
          Plano
        </button>
        <button type="button" aria-pressed={props.relieve} onClick={() => props.onRelieve(true)}>
          Relieve
        </button>
      </div>
      <details open>
        <summary>{MAPA.capas}</summary>
        <div className="layer">
          <label>
            <input type="checkbox" checked={props.showEdificios} onChange={props.onEdificios} /> {CAPA.edificios.label}
          </label>
          <span className="count">{props.edificios || "—"}</span>
        </div>
        <fieldset className="grupo catastro-color">
          <legend>{MAPA.catastro}</legend>
          <label className="layer-toggle">
            <input type="checkbox" checked={props.catastroOn} onChange={props.onCatastro} /> {MAPA.mostrarCatastro}
            <small>{conteoCatastro(props.areas, props.zonas)}</small>
          </label>
          <div className="roles segmentado" role="group" aria-label={MAPA.colorear}>
            <button type="button" aria-pressed={props.colorPor === "uso"} onClick={() => props.onColorPor("uso")}>
              {MAPA.uso}
            </button>
            <button type="button" aria-pressed={props.colorPor === "sector"} onClick={() => props.onColorPor("sector")}>
              {MAPA.sector}
            </button>
            <AyudaTip texto={AYUDA.sector} etiqueta="Qué es un sector de capataz" />
          </div>
          <ul className="leyenda" aria-label={props.colorPor === "uso" ? MAPA.leyendaUso : MAPA.leyendaSector}>
            {props.categorias.map((cat) => {
              const n = props.conteo[cat.id]
              return (
                <li key={cat.id}>
                  <label>
                    <input
                      type="checkbox"
                      checked={!props.ocultas.includes(cat.id)}
                      disabled={n === 0}
                      onChange={() => props.onOculta(cat.id)}
                    />
                    <span className="swatch" style={{ background: `color-mix(in srgb, ${cat.fill} 55%, var(--hoja))`, borderColor: cat.line }} />
                    <span>{cat.label}</span>
                    <span className="count">{n}</span>
                  </label>
                </li>
              )
            })}
          </ul>
          {props.colorPor === "sector" && <p className="hint">{MAPA.cuadrillasFicticias}</p>}
        </fieldset>
        {LAYERS.filter((layer) => layer.id === "jardines_reserva" || layer.id === "xerofitica" || layer.id === "vias" || layer.id === "cuarteles").map((layer) => (
          <div className="layer" key={layer.id}>
            <label>
              <input type="checkbox" checked={props.visible[layer.id]} onChange={() => props.onCapa(layer.id)} /> {layer.label}
              <small>{layer.id === "cuarteles" && (props.data.cuarteles?.features.length ?? 0) === 0 ? ZONIFICACION.sinCuarteles : layer.hint}</small>
            </label>
            <span className="count">{props.data[layer.id]?.features.length ?? "—"}</span>
          </div>
        ))}
        <h3>{MAPA.inventario}</h3>
        {INVENTARIO.map((layer) => (
          <div className="layer" key={layer.id}>
            <label>
              <input type="checkbox" checked={props.inventoryOn[layer.id] === true} onChange={() => props.onInventario(layer.id)} /> {layer.label}
            </label>
            <span className="count">{props.inventory[layer.id]?.features.length ?? "—"}</span>
          </div>
        ))}
      </details>
      <p className="control-resumen">{props.resumen}</p>
    </aside>
  )
}
