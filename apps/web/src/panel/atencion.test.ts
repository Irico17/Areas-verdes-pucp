import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { Labores, type LaborItem } from "./Labores.tsx"
import { PodaPanel } from "./Poda.tsx"
import { codigoExterno, validarPoda, type PodaItem } from "./poda.ts"
import { ViveroPanel } from "./Vivero.tsx"
import { validarVivero } from "./vivero.ts"

const labor: LaborItem = {
  id: "1",
  titulo: "Poda de ficus",
  tipo: "poda",
  estado: "sin_estado",
  equipo: "",
  detalle: "",
  capatazId: "",
  ejecutor: "propia",
}

const poda: PodaItem = {
  id: "p1",
  codigo: "PO-1",
  codigo_externo: "OSG-0478",
  tipo: "Hallazgo",
  tipo_actividad: "Poda de formación",
  fecha_reporte: "2026-07-06",
  fecha_ejecucion: "2026-07-08",
  personal: "Nora Beltrán",
  ubicacion: "Mac Gregor",
  unidad: "und",
  cantidad_pedida: "1",
  cantidad_ejecutada: "1",
  prioridad: "baja",
  comentario: "",
  nombre_comun: "Laurel",
  nombre_cientifico: "Laurus nobilis",
}

test("la ficha de labores conserva el pin y muestra los campos de escritorio", () => {
  const html = renderToStaticMarkup(
    createElement(Labores, {
      rol: "coordinacion",
      equipos: [],
      equipoId: "",
      onEquipo: () => {},
      items: [labor],
      estados: {},
      onToggleEstado: () => {},
      tipo: "",
      onTipo: () => {},
      pinMode: true,
      onPinMode: () => {},
      draft: { lon: -77.08, lat: -12.07 },
      formTipo: "poda",
      formTitulo: "",
      formDetalle: "",
      formEquipo: "",
      onForm: () => {},
      onCreate: () => {},
      creating: false,
      selected: labor,
      onSelect: () => {},
      timeline: [],
      timelineError: "",
      estadoNuevo: "sin_estado",
      onEstadoNuevo: () => {},
      onEstado: () => {},
      reasignarA: "",
      onReasignarA: () => {},
      onReasignar: () => {},
      onArchivar: () => {},
      confirmarArchivo: false,
      notice: "",
      queueCount: 0,
      onFlush: () => {},
      tipos: [{ id: "poda", label: "Poda" }],
      formEjecutor: "tercerizada",
      motivos: [],
      motivo: "",
      onMotivo: () => {},
      onSugerir: () => {},
      sugerencia: "",
      evidencias: [],
      onSubir: () => {},
    }),
  )
  assert.match(html, /Marcar labor|Cancelar marca/)
  assert.match(html, /Servicio tercerizado/)
  for (const etiqueta of ["Clase", "Fecha de solicitud", "Fecha de atención", "Lugar", "Comentario", "Sin estado"]) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
})

test("el código OSG no se inventa y la poda valida cantidades", () => {
  assert.equal(codigoExterno("").codigo, "")
  assert.equal(codigoExterno("aun no tiene codigo").codigo, "")
  assert.equal(codigoExterno("OSG-0478").codigo, "OSG-0478")
  assert.ok(codigoExterno("nuevo").fallo)
  assert.equal(validarPoda({ ...poda, cantidad_pedida: "-1" }).some((item) => item.campo === "cantidad_pedida"), true)
  const html = renderToStaticMarkup(createElement(PodaPanel, { iniciales: [poda] }))
  assert.match(html, /PO-1/)
  assert.match(html, /Código externo/)
})

test("el vivero filtra por mes y no acepta un área libre", () => {
  const html = renderToStaticMarkup(
    createElement(ViveroPanel, {
      delta: "copia local 683; publicado 689",
      iniciales: [
        {
          id: "v1",
          fecha: "2026-03-02",
          area: "Flora",
          subproceso: "Siembra",
          etapa: "Inicio",
          descripcion: "",
          observaciones: "",
          responsables: "Lucía Mendoza",
          lugar: "Vivero",
        },
      ],
      subprocesos: ["Siembra"],
      etapas: ["Inicio"],
    }),
  )
  assert.match(html, /683/)
  assert.match(html, /Mes/)
  assert.equal(validarVivero({ id: "x", fecha: "", area: "Inventada", subproceso: "", etapa: "", descripcion: "", observaciones: "", responsables: "", lugar: "" }, { subproceso: [], etapa: [] }).length, 1)
})
