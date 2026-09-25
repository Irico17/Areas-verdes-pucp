import { useState } from "react"
import {
  CAPAS_EDITABLES,
  CONTEOS_TACHO,
  csvFichas,
  validarBebedero,
  validarConteos,
  validarPunto,
  validarReserva,
  type CapaId,
  type Conteos,
  type ErrorCampo,
} from "./inventarioCapas"

type Entidad = "tachos" | "bebederos" | "puntos" | "reservas" | CapaId

const ceros = Object.fromEntries(CONTEOS_TACHO.map((campo) => [campo, 0])) as Conteos

function motivos(errores: ErrorCampo[], campo: string): string[] {
  return errores.filter((item) => item.campo === campo).map((item) => item.motivo)
}

export function InventarioCapas() {
  const [entidad, setEntidad] = useState<Entidad>("tachos")
  const [errores, setErrores] = useState<ErrorCampo[]>([])
  const [aviso, setAviso] = useState("")
  const [codigo, setCodigo] = useState("PT_")
  const [conteos, setConteos] = useState<Conteos>(ceros)
  const [recomendacion, setRecomendacion] = useState("")
  const [estado, setEstado] = useState("")
  const [sede, setSede] = useState("")
  const [subtipo] = useState("fuente")
  const [titulo, setTitulo] = useState("")
  const [reserva, setReserva] = useState({ origen: "ficticio", fecha: "2026-09-22", hora_inicio: "09:00", hora_fin: "11:00", estado: "reservado", evento: "" })
  const [ficha, setFicha] = useState({ feature_id: "", nombre: "", codigo: "" })

  function guardar() {
    let lista: ErrorCampo[] = []
    if (entidad === "tachos") lista = validarConteos(conteos).concat(codigo.startsWith("PT") ? [] : [{ campo: "codigo", motivo: "El código empieza por PT." }])
    if (entidad === "bebederos") lista = validarBebedero({ codigo, subtipo, estado, sede })
    if (entidad === "puntos") lista = validarPunto({ titulo, lat: -12.07, lon: -77.08 })
    if (entidad === "reservas") lista = validarReserva(reserva)
    if (entidad !== "tachos" && entidad !== "bebederos" && entidad !== "puntos" && entidad !== "reservas" && !ficha.feature_id.trim()) {
      lista = [{ campo: "feature_id", motivo: "Falta el identificador." }]
    }
    setErrores(lista)
    setAviso(lista.length === 0 ? "Listo para guardar. La hoja de reservas no se consulta." : "")
  }

  return (
    <section className="split catastro-editor">
      <div className="split-list">
        <div className="roles" role="tablist" aria-label="Inventario">
          {(["tachos", "bebederos", "puntos", "reservas"] as const).map((id) => (
            <button key={id} type="button" aria-pressed={entidad === id} onClick={() => setEntidad(id)}>
              {id}
            </button>
          ))}
        </div>
        <ul className="labor-list">
          {CAPAS_EDITABLES.map((capa) => (
            <li key={capa.id}>
              <button type="button" className={entidad === capa.id ? "labor on" : "labor"} onClick={() => setEntidad(capa.id)}>
                <span>{capa.label}</span>
                <small>{capa.campos.join(", ")}</small>
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
            guardar()
          }}
        >
          <h2>Edición de inventario</h2>
          {entidad === "tachos" && (
            <>
              <label className="field">
                Código
                <input value={codigo} onChange={(event) => setCodigo(event.target.value)} />
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
                Estado
                <input value={estado} onChange={(event) => setEstado(event.target.value)} />
              </label>
              <label className="field">
                Sede
                <input value={sede} onChange={(event) => setSede(event.target.value)} />
              </label>
            </>
          )}
          {entidad === "puntos" && (
            <label className="field">
              Título
              <input value={titulo} onChange={(event) => setTitulo(event.target.value)} />
            </label>
          )}
          {entidad === "reservas" && (
            <>
              <p className="hint">Agenda ficticia. La hoja institucional responde 401 y no se abre.</p>
              <label className="field">
                Evento
                <input value={reserva.evento} onChange={(event) => setReserva({ ...reserva, evento: event.target.value })} />
              </label>
            </>
          )}
          {entidad !== "tachos" && entidad !== "bebederos" && entidad !== "puntos" && entidad !== "reservas" && (
            <>
              <label className="field">
                Identificador
                <input value={ficha.feature_id} onChange={(event) => setFicha({ ...ficha, feature_id: event.target.value })} />
              </label>
              <label className="field">
                Nombre o código
                <input value={ficha.nombre} onChange={(event) => setFicha({ ...ficha, nombre: event.target.value })} />
              </label>
              <p className="row-actions">
                <button
                  type="button"
                  onClick={() => {
                    const blob = new Blob([csvFichas([{ feature_id: ficha.feature_id || "nuevo", nombre: ficha.nombre, codigo: ficha.codigo }])], { type: "text/csv" })
                    const url = URL.createObjectURL(blob)
                    const link = document.createElement("a")
                    link.href = url
                    link.download = `${entidad}.csv`
                    link.click()
                    URL.revokeObjectURL(url)
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
          <p className="hint">{motivos(errores, "origen").join(" ")}</p>
          <button type="submit">Guardar</button>
        </form>
      </div>
    </section>
  )
}
