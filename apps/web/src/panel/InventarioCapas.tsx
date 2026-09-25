import { useEffect, useState } from "react"
import {
  CAPAS_EDITABLES,
  CONTEOS_TACHO,
  cuerpoPunto,
  csvFichas,
  enviarInventario,
  esCapa,
  rutaListar,
  solicitudBaja,
  solicitudGuardar,
  validarBebedero,
  validarConteos,
  validarPunto,
  validarReserva,
  type CapaId,
  type Conteos,
  type EntidadInventario,
  type ErrorCampo,
} from "./inventarioCapas"

type Fila = { id: number; etiqueta: string }

const ceros = Object.fromEntries(CONTEOS_TACHO.map((campo) => [campo, 0])) as Conteos

const GEO_PUNTO = '{"type":"Point","coordinates":[-77.08,-12.07]}'

function numeroOpcional(texto: string): number | null {
  const limpio = texto.trim()
  if (!limpio) return null
  const n = Number(limpio)
  return Number.isFinite(n) ? n : null
}

function listaDe(body: unknown, entidad: EntidadInventario): Fila[] {
  const clave = esCapa(entidad) ? "filas" : entidad
  const raw = body && typeof body === "object" ? (body as Record<string, unknown>)[clave] : null
  if (!Array.isArray(raw)) return []
  return raw.flatMap((item) => {
    if (!item || typeof item !== "object") return []
    const row = item as Record<string, unknown>
    const id = Number(row.id)
    if (!Number.isInteger(id) || id < 1) return []
    const etiqueta = String(row.codigo || row.titulo || row.evento || row.feature_id || row.nombre || id)
    return [{ id, etiqueta }]
  })
}

export function InventarioCapas({
  entidadInicial = "tachos",
  cliente = fetch,
}: {
  entidadInicial?: EntidadInventario
  cliente?: typeof fetch
}) {
  const [entidad, setEntidad] = useState<EntidadInventario>(entidadInicial)
  const [filas, setFilas] = useState<Fila[]>([])
  const [id, setId] = useState<number | null>(null)
  const [errores, setErrores] = useState<ErrorCampo[]>([])
  const [aviso, setAviso] = useState("")
  const [codigo, setCodigo] = useState("PT_")
  const [lat, setLat] = useState("-12.07")
  const [lon, setLon] = useState("-77.08")
  const [conteos, setConteos] = useState<Conteos>(ceros)
  const [recomendacion, setRecomendacion] = useState("")
  const [notaTacho, setNotaTacho] = useState("")
  const [estado, setEstado] = useState("operativo")
  const [sede, setSede] = useState("CAMPUS")
  const [subtipo, setSubtipo] = useState("fuente")
  const [titulo, setTitulo] = useState("")
  const [url, setUrl] = useState("")
  const [reserva, setReserva] = useState({
    origen: "ficticio",
    fecha: "2026-09-22",
    hora_inicio: "09:00",
    hora_fin: "11:00",
    estado: "reservado",
    evento: "",
    unidad: "",
  })
  const [ficha, setFicha] = useState({
    feature_id: "",
    nombre: "",
    codigo: "",
    nota: "",
    clase: "",
    riego: "",
    area_m2: "",
    perimetro_m: "",
    pertenecen: "",
    uso: "",
    geojson: "",
  })

  useEffect(() => {
    let vivo = true
    enviarInventario(cliente, { method: "GET", path: rutaListar(entidad) })
      .then((body) => {
        if (vivo) setFilas(listaDe(body, entidad))
      })
      .catch(() => {
        if (vivo) setFilas([])
      })
    return () => {
      vivo = false
    }
  }, [cliente, entidad])

  function elegir(siguiente: EntidadInventario) {
    setEntidad(siguiente)
    setId(null)
    setErrores([])
    setAviso("")
  }

  function cuerpo(): unknown {
    const latN = Number(lat)
    const lonN = Number(lon)
    if (entidad === "tachos") {
      return {
        codigo,
        lat: Number.isFinite(latN) ? latN : null,
        lon: Number.isFinite(lonN) ? lonN : null,
        nota: notaTacho,
        recomendaciones: recomendacion,
        ...conteos,
      }
    }
    if (entidad === "bebederos") {
      return { codigo, subtipo, estado, sede, lat: Number.isFinite(latN) ? latN : null, lon: Number.isFinite(lonN) ? lonN : null }
    }
    if (entidad === "puntos") return cuerpoPunto({ titulo, lat: latN, lon: lonN, url })
    if (entidad === "reservas") return { ...reserva, origen: "ficticio" }
    return {
      feature_id: ficha.feature_id.trim(),
      nombre: ficha.nombre,
      codigo: ficha.codigo,
      nota: ficha.nota,
      clase: ficha.clase,
      riego: ficha.riego,
      area_m2: numeroOpcional(ficha.area_m2),
      perimetro_m: numeroOpcional(ficha.perimetro_m),
      pertenecen: ficha.pertenecen,
      uso: ficha.uso,
      geojson: ficha.geojson.trim(),
    }
  }

  function validar(): ErrorCampo[] {
    if (entidad === "tachos") {
      const lista = validarConteos(conteos)
      if (!codigo.startsWith("PT")) lista.push({ campo: "codigo", motivo: "El código empieza por PT." })
      return lista
    }
    if (entidad === "bebederos") return validarBebedero({ codigo, subtipo, estado, sede })
    if (entidad === "puntos") return validarPunto({ titulo, lat: Number(lat), lon: Number(lon), url })
    if (entidad === "reservas") return validarReserva(reserva)
    if (!ficha.feature_id.trim()) return [{ campo: "feature_id", motivo: "Falta el identificador." }]
    return []
  }

  async function guardar() {
    const lista = validar()
    setErrores(lista)
    if (lista.length > 0) {
      setAviso("")
      return
    }
    try {
      const guardado = await enviarInventario(cliente, solicitudGuardar(entidad, id, cuerpo()))
      const nuevo = guardado && typeof guardado === "object" ? Number((guardado as { id?: number }).id) : NaN
      if (Number.isInteger(nuevo) && nuevo > 0) setId(nuevo)
      setAviso(id == null ? "Guardado en la API." : "Actualizado en la API.")
      const body = await enviarInventario(cliente, { method: "GET", path: rutaListar(entidad) })
      setFilas(listaDe(body, entidad))
    } catch (error) {
      setAviso(error instanceof Error ? error.message : "No se pudo guardar.")
    }
  }

  async function eliminar() {
    if (id == null) return
    try {
      await enviarInventario(cliente, solicitudBaja(entidad, id))
      setId(null)
      setAviso("Dado de baja en la API.")
      const body = await enviarInventario(cliente, { method: "GET", path: rutaListar(entidad) })
      setFilas(listaDe(body, entidad))
    } catch (error) {
      setAviso(error instanceof Error ? error.message : "No se pudo dar de baja.")
    }
  }

  const capa = esCapa(entidad) ? CAPAS_EDITABLES.find((item) => item.id === entidad) : undefined
  const campos = capa ? (capa.campos as readonly string[]) : []

  return (
    <section className="split catastro-editor">
      <div className="split-list">
        <div className="roles" role="tablist" aria-label="Inventario">
          {(["tachos", "bebederos", "puntos", "reservas"] as const).map((item) => (
            <button key={item} type="button" aria-pressed={entidad === item} onClick={() => elegir(item)}>
              {item}
            </button>
          ))}
        </div>
        <ul className="labor-list">
          {CAPAS_EDITABLES.map((item) => (
            <li key={item.id}>
              <button type="button" className={entidad === item.id ? "labor on" : "labor"} onClick={() => elegir(item.id)}>
                <span>{item.label}</span>
                <small>{item.campos.join(", ")}</small>
              </button>
            </li>
          ))}
          {filas.map((fila) => (
            <li key={fila.id}>
              <button type="button" onClick={() => setId(fila.id)}>
                {fila.etiqueta}
              </button>
            </li>
          ))}
        </ul>
      </div>
      <div className="split-detail">
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            void guardar()
          }}
        >
          <h2>Edición de inventario</h2>
          {entidad === "tachos" && (
            <>
              <label className="field">
                Código
                <input value={codigo} onChange={(event) => setCodigo(event.target.value)} />
              </label>
              <label className="field">
                Latitud
                <input value={lat} onChange={(event) => setLat(event.target.value)} />
              </label>
              <label className="field">
                Longitud
                <input value={lon} onChange={(event) => setLon(event.target.value)} />
              </label>
              {CONTEOS_TACHO.map((campo) => (
                <label className="field" key={campo}>
                  {campo}
                  <input
                    inputMode="numeric"
                    value={conteos[campo]}
                    onChange={(event) => setConteos({ ...conteos, [campo]: Number(event.target.value) })}
                  />
                </label>
              ))}
              <label className="field">
                Nota
                <input value={notaTacho} onChange={(event) => setNotaTacho(event.target.value)} />
              </label>
              <label className="field">
                Recomendación
                <input value={recomendacion} onChange={(event) => setRecomendacion(event.target.value)} />
              </label>
            </>
          )}
          {entidad === "bebederos" && (
            <>
              <label className="field">
                Código
                <input value={codigo} onChange={(event) => setCodigo(event.target.value)} />
              </label>
              <label className="field">
                Subtipo
                <select value={subtipo} onChange={(event) => setSubtipo(event.target.value)}>
                  {["fuente", "llenador", "nuevo", "deterioro", "baja"].map((item) => (
                    <option key={item}>{item}</option>
                  ))}
                </select>
              </label>
              <label className="field">
                Estado
                <input value={estado} onChange={(event) => setEstado(event.target.value)} />
              </label>
              <label className="field">
                Sede
                <input value={sede} onChange={(event) => setSede(event.target.value)} />
              </label>
              <label className="field">
                Latitud
                <input value={lat} onChange={(event) => setLat(event.target.value)} />
              </label>
              <label className="field">
                Longitud
                <input value={lon} onChange={(event) => setLon(event.target.value)} />
              </label>
            </>
          )}
          {entidad === "puntos" && (
            <>
              <label className="field">
                Título
                <input value={titulo} onChange={(event) => setTitulo(event.target.value)} />
              </label>
              <label className="field">
                Latitud
                <input value={lat} onChange={(event) => setLat(event.target.value)} />
              </label>
              <label className="field">
                Longitud
                <input value={lon} onChange={(event) => setLon(event.target.value)} />
              </label>
              <label className="field">
                URL
                <input value={url} onChange={(event) => setUrl(event.target.value)} />
              </label>
            </>
          )}
          {entidad === "reservas" && (
            <>
              <p className="hint">Agenda ficticia. La hoja institucional responde 401 y no se abre.</p>
              <label className="field">
                Fecha
                <input value={reserva.fecha} onChange={(event) => setReserva({ ...reserva, fecha: event.target.value })} />
              </label>
              <label className="field">
                Hora inicio
                <input value={reserva.hora_inicio} onChange={(event) => setReserva({ ...reserva, hora_inicio: event.target.value })} />
              </label>
              <label className="field">
                Hora fin
                <input value={reserva.hora_fin} onChange={(event) => setReserva({ ...reserva, hora_fin: event.target.value })} />
              </label>
              <label className="field">
                Evento
                <input value={reserva.evento} onChange={(event) => setReserva({ ...reserva, evento: event.target.value })} />
              </label>
              <label className="field">
                Unidad
                <input value={reserva.unidad} onChange={(event) => setReserva({ ...reserva, unidad: event.target.value })} />
              </label>
            </>
          )}
          {capa && (
            <>
              <label className="field">
                Identificador
                <input value={ficha.feature_id} onChange={(event) => setFicha({ ...ficha, feature_id: event.target.value })} />
              </label>
              {campos.includes("nombre") && (
                <label className="field">
                  Nombre
                  <input value={ficha.nombre} onChange={(event) => setFicha({ ...ficha, nombre: event.target.value })} />
                </label>
              )}
              {campos.includes("codigo") && (
                <label className="field">
                  Código
                  <input value={ficha.codigo} onChange={(event) => setFicha({ ...ficha, codigo: event.target.value })} />
                </label>
              )}
              {campos.includes("nota") && (
                <label className="field">
                  Nota
                  <input value={ficha.nota} onChange={(event) => setFicha({ ...ficha, nota: event.target.value })} />
                </label>
              )}
              {campos.includes("clase") && (
                <label className="field">
                  Clase
                  <input value={ficha.clase} onChange={(event) => setFicha({ ...ficha, clase: event.target.value })} />
                </label>
              )}
              {campos.includes("riego") && (
                <label className="field">
                  Riego
                  <input value={ficha.riego} onChange={(event) => setFicha({ ...ficha, riego: event.target.value })} />
                </label>
              )}
              {campos.includes("area_m2") && (
                <label className="field">
                  Área
                  <input value={ficha.area_m2} onChange={(event) => setFicha({ ...ficha, area_m2: event.target.value })} />
                </label>
              )}
              {campos.includes("perimetro_m") && (
                <label className="field">
                  Perímetro
                  <input value={ficha.perimetro_m} onChange={(event) => setFicha({ ...ficha, perimetro_m: event.target.value })} />
                </label>
              )}
              {campos.includes("pertenecen") && (
                <label className="field">
                  Pertenecen
                  <input value={ficha.pertenecen} onChange={(event) => setFicha({ ...ficha, pertenecen: event.target.value })} />
                </label>
              )}
              {campos.includes("uso") && (
                <label className="field">
                  Uso
                  <input value={ficha.uso} onChange={(event) => setFicha({ ...ficha, uso: event.target.value })} />
                </label>
              )}
              <label className="field">
                Geometría
                <textarea value={ficha.geojson} placeholder={GEO_PUNTO} onChange={(event) => setFicha({ ...ficha, geojson: event.target.value })} />
              </label>
              <p className="row-actions">
                <button
                  type="button"
                  onClick={() => {
                    const blob = new Blob(
                      [csvFichas([{ feature_id: ficha.feature_id || "nuevo", nombre: ficha.nombre, codigo: ficha.codigo }])],
                      { type: "text/csv" },
                    )
                    const enlace = URL.createObjectURL(blob)
                    const link = document.createElement("a")
                    link.href = enlace
                    link.download = `${entidad}.csv`
                    link.click()
                    URL.revokeObjectURL(enlace)
                  }}
                >
                  Exportar CSV
                </button>
              </p>
            </>
          )}
          {errores.length > 0 && (
            <ul className="errores-campo">
              {errores.map((item) => (
                <li key={item.campo + item.motivo} className="status error">
                  {item.motivo}
                </li>
              ))}
            </ul>
          )}
          {aviso && <p className="hint">{aviso}</p>}
          <button type="submit">Guardar</button>
          {id != null && (
            <button type="button" onClick={() => void eliminar()}>
              Eliminar
            </button>
          )}
        </form>
      </div>
    </section>
  )
}

export type { CapaId }
