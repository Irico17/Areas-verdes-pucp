import { useEffect, useState } from "react"
import { FILTRO_ACTIVIDADES, TIPOS, estadosPermitidos, fetchCapataces, fetchTaxonomiaActividad, type FiltroActividades } from "../operacion"
import type { Rol } from "../types"
import { ACTIVIDAD, MAPA, SECTOR_CAPATAZ, etiquetaCuadrilla } from "../ui/nomenclatura"

type Opcion = { id: string; label: string }

type Props = {
  rol: Rol
  valor?: FiltroActividades
  onChange?: (valor: FiltroActividades) => void
  cuadrillas?: Opcion[]
  tipos?: Opcion[]
  riesgos?: Opcion[]
  origenes?: Opcion[]
}

const SECTORES = Object.entries(SECTOR_CAPATAZ)
  .filter(([id]) => id !== "sin-sector")
  .map(([id, label]) => ({ id, label }))

export function FiltrosActividad(props: Props) {
  const [local, setLocal] = useState<FiltroActividades>(FILTRO_ACTIVIDADES)
  const [cuadrillasRemotas, setCuadrillasRemotas] = useState<Opcion[]>([])
  const [tiposRemotos, setTiposRemotos] = useState<Opcion[]>([])
  const [riesgosRemotos, setRiesgosRemotos] = useState<Opcion[]>([])
  const [origenesRemotos, setOrigenesRemotos] = useState<Opcion[]>([])
  const valor = props.valor ?? local
  const avisar = props.onChange ?? setLocal
  const oficina = props.rol !== "capataz"

  useEffect(() => {
    if (props.cuadrillas || !oficina) return
    let vivo = true
    void fetchCapataces()
      .then((filas) => {
        if (!vivo) return
        setCuadrillasRemotas(filas.map((fila) => ({ id: fila.id, label: etiquetaCuadrilla(fila.id, fila.equipo) })))
      })
      .catch(() => {
        if (vivo) setCuadrillasRemotas([])
      })
    return () => {
      vivo = false
    }
  }, [oficina, props.cuadrillas])

  useEffect(() => {
    if (props.riesgos && props.origenes && props.tipos) return
    let vivo = true
    void fetchTaxonomiaActividad()
      .then((tax) => {
        if (!vivo) return
        setRiesgosRemotos(tax.riesgos.map((item) => ({ id: item.codigo, label: item.nombre })))
        setOrigenesRemotos(tax.origenes.map((item) => ({ id: item.codigo, label: item.nombre })))
        const tipos = tax.clases.flatMap((clase) => clase.tipos.map((tipo) => ({ id: tipo.codigo, label: tipo.nombre })))
        setTiposRemotos(tipos)
      })
      .catch(() => {})
    return () => {
      vivo = false
    }
  }, [props.riesgos, props.origenes, props.tipos])

  function patch(parcial: Partial<FiltroActividades>) {
    avisar({ ...valor, ...parcial })
  }

  const cuadrillas = props.cuadrillas ?? cuadrillasRemotas
  const tipos = props.tipos ?? (tiposRemotos.length > 0 ? tiposRemotos : TIPOS.map((tipo) => ({ id: tipo.id, label: tipo.label })))
  const riesgos = props.riesgos ?? (riesgosRemotos.length > 0 ? riesgosRemotos : [
    { id: "bajo", label: "Bajo" },
    { id: "alto", label: "Alto" },
  ])
  const origenes = props.origenes ?? origenesRemotos

  return (
    <fieldset className="filtros-actividad">
      <legend>{ACTIVIDAD.filtros}</legend>
      <label className="field">
        {ACTIVIDAD.estado}
        <select name="estado" value={valor.estado} onChange={(event) => patch({ estado: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierEstado}</option>
          {estadosPermitidos(props.rol).map((estado) => (
            <option key={estado.id} value={estado.id}>
              {estado.label}
            </option>
          ))}
        </select>
      </label>
      <label className="field">
        {ACTIVIDAD.tipo}
        <select name="tipo" value={valor.tipo} onChange={(event) => patch({ tipo: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierTipo}</option>
          {tipos.map((tipo) => (
            <option key={tipo.id} value={tipo.id}>
              {tipo.label}
            </option>
          ))}
        </select>
      </label>
      {oficina && (
        <label className="field">
          {ACTIVIDAD.cuadrilla}
          <select name="cuadrilla_id" value={valor.cuadrillaId} onChange={(event) => patch({ cuadrillaId: event.target.value })}>
            <option value="">{ACTIVIDAD.todas}</option>
            {cuadrillas.map((cuadrilla) => (
              <option key={cuadrilla.id} value={cuadrilla.id}>
                {cuadrilla.label}
              </option>
            ))}
          </select>
        </label>
      )}
      <label className="field">
        {MAPA.sector}
        <select name="sector" value={valor.sector} onChange={(event) => patch({ sector: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierSector}</option>
          {SECTORES.map((sector) => (
            <option key={sector.id} value={sector.id}>
              {sector.label}
            </option>
          ))}
        </select>
      </label>
      <label className="field">
        {ACTIVIDAD.quienEjecuta}
        <select name="ejecutor" value={valor.ejecutor} onChange={(event) => patch({ ejecutor: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierEjecutor}</option>
          <option value="propia">{ACTIVIDAD.personalPropio}</option>
          <option value="tercerizada">{ACTIVIDAD.servicioTercerizado}</option>
        </select>
      </label>
      <label className="field">
        {ACTIVIDAD.origen}
        <select name="origen" value={valor.origen} onChange={(event) => patch({ origen: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierOrigen}</option>
          {origenes.map((origen) => (
            <option key={origen.id} value={origen.id}>
              {origen.label}
            </option>
          ))}
        </select>
      </label>
      <label className="field">
        {ACTIVIDAD.riesgo}
        <select name="nivel_riesgo" value={valor.nivelRiesgo} onChange={(event) => patch({ nivelRiesgo: event.target.value })}>
          <option value="">{ACTIVIDAD.cualquierRiesgo}</option>
          {riesgos.map((riesgo) => (
            <option key={riesgo.id} value={riesgo.id}>
              {riesgo.label}
            </option>
          ))}
        </select>
      </label>
      <div className="fechas">
        <label className="field">
          {ACTIVIDAD.desde}
          <input name="desde" type="date" value={valor.desde} onChange={(event) => patch({ desde: event.target.value })} />
        </label>
        <label className="field">
          {ACTIVIDAD.hasta}
          <input name="hasta" type="date" value={valor.hasta} onChange={(event) => patch({ hasta: event.target.value })} />
        </label>
      </div>
      <label className="check-linea">
        <input name="historico" type="checkbox" checked={valor.historico} onChange={(event) => patch({ historico: event.target.checked })} />
        {ACTIVIDAD.historico}
      </label>
    </fieldset>
  )
}
