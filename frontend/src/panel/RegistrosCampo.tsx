import { useState } from "react"
import { PodaPanel } from "./Poda"
import { RiegoPanel } from "./Riego"
import { ViveroPanel } from "./Vivero"
import { REGISTRO_CAMPO } from "../ui/nomenclatura"

type Sub = "riego" | "poda" | "vivero"

export function RegistrosCampo(props: { capatazId: string }) {
  const [sub, setSub] = useState<Sub>("riego")
  const [nuevo, setNuevo] = useState(false)

  function elegir(siguiente: Sub) {
    setSub(siguiente)
    setNuevo(false)
  }

  return (
    <section className="registros-campo" aria-labelledby="registros-titulo">
      <h2 id="registros-titulo">{REGISTRO_CAMPO.titulo}</h2>
      <p className="lede">{REGISTRO_CAMPO.lede}</p>
      <div className="roles" role="tablist" aria-label={REGISTRO_CAMPO.titulo}>
        {(
          [
            ["riego", REGISTRO_CAMPO.riego],
            ["poda", REGISTRO_CAMPO.poda],
            ["vivero", REGISTRO_CAMPO.vivero],
          ] as const
        ).map(([id, etiqueta]) => (
          <button key={id} type="button" role="tab" aria-selected={sub === id} onClick={() => elegir(id)}>
            {etiqueta}
          </button>
        ))}
      </div>
      <p className="row-actions">
        <button type="button" className="primary" aria-expanded={nuevo} onClick={() => setNuevo((abierto) => !abierto)}>
          {nuevo ? REGISTRO_CAMPO.cerrar : REGISTRO_CAMPO.nuevo}
        </button>
      </p>
      {sub === "riego" && <RiegoPanel capatazId={props.capatazId} mostrarFormulario={nuevo} />}
      {sub === "poda" && <PodaPanel mostrarFormulario={nuevo} />}
      {sub === "vivero" && <ViveroPanel mostrarFormulario={nuevo} delta="" />}
    </section>
  )
}
