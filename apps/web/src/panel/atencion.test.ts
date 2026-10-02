import assert from "node:assert/strict"
import { test } from "node:test"
import { createElement } from "react"
import { renderToStaticMarkup } from "react-dom/server"
import { Labores, type LaborItem } from "./Labores.tsx"
import { PodaPanel } from "./Poda.tsx"
import { codigoExterno, validarPoda, type PodaItem } from "./poda.ts"
import { etiquetaRol } from "../producto.ts"
import { estadosPermitidos, puedeEncolarEstado } from "../operacion.ts"
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
      sugerencia: "Regla local sobre el título. No es un modelo externo: confirme antes de guardar.",
      pista: { codigo: "poda", etiqueta: "Poda" },
      onConfirmarPista: () => {},
    }),
  )
  assert.match(html, /Actividades/)
  assert.match(html, /Cuadrilla/)
  assert.match(html, /Marcar actividad|Cancelar marca/)
  assert.match(html, /Servicio tercerizado/)
  assert.match(html, /en proceso/)
  assert.match(html, /Confirmar Poda/)
  for (const etiqueta of ["Clase", "Fecha de solicitud", "Fecha de atención", "Lugar", "Comentario", "Sin estado"]) {
    assert.ok(html.includes(etiqueta), etiqueta)
  }
})

test("el alta de actividad pide clase, tipo, riesgo y lugar de catálogo", () => {
  const html = renderToStaticMarkup(
    createElement(Labores, {
      rol: "coordinacion",
      equipos: [],
      equipoId: "",
      onEquipo: () => {},
      items: [],
      pinMode: true,
      onPinMode: () => {},
      draft: { lon: -77.08, lat: -12.07 },
      formTipo: "riego",
      formTitulo: "",
      formDetalle: "",
      formEquipo: "",
      onForm: () => {},
      onCreate: () => {},
      creating: false,
      selected: null,
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
      tipos: [{ id: "riego", label: "Riego" }],
      formEjecutor: "propia",
      motivos: [],
      motivo: "",
      onMotivo: () => {},
      onSugerir: () => {},
      sugerencia: "",
      pista: null,
      taxonomia: {
        clases: [
          {
            codigo: "riego",
            nombre: "Riego",
            tipos: [{ codigo: "riego_manual", nombre: "Riego manual" }],
          },
        ],
        riesgos: [
          { codigo: "bajo", nombre: "Bajo" },
          { codigo: "alto", nombre: "Alto" },
        ],
        origenes: [{ codigo: "interna", nombre: "Interna" }],
        personal: [{ id: "pf-elsa", nombre_ficticio: "Elsa Mamani" }],
      },
      lugares: [{ id: 1, nombre: "Jardín de Letras" }],
      zonas: [{ codigo: "Z1", nombre: "Zona 1" }],
    }),
  )
  const alta = html.slice(html.indexOf('class="form alta-actividad"'), html.indexOf("</form>") + 7)
  for (const etiqueta of [
    "Clase de actividad",
    "Tipo de actividad",
    "Origen",
    "Código externo",
    "Unidad solicitante",
    "Nivel de riesgo",
    "Fecha programada",
    "Cantidad",
    "Personal de la actividad",
    "Zona de supervisión",
    "Bajo",
    "Alto",
    "Elsa Mamani",
    "Riego",
  ]) {
    assert.ok(alta.includes(etiqueta), etiqueta)
  }
  assert.equal(alta.includes("Medio"), false)
  assert.match(alta, /name="lugar_id"/)
  assert.equal(alta.includes('name="lugar"'), false)
  assert.equal(alta.includes("lugar_libre"), false)
  assert.match(alta, /name="clase"/)
  assert.match(alta, /name="subtipo"/)
  assert.match(alta, /name="nivel_riesgo"/)
  assert.match(alta, /name="personal"/)
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

test("etiquetaRol y roles v2 muestran los nombres visibles correctos", () => {
  assert.equal(etiquetaRol("capataz"), "Capataz")
  assert.equal(etiquetaRol("coordinacion"), "Ingeniería / Coordinación")
  assert.equal(etiquetaRol("jefatura"), "Jefatura de sección")
  assert.equal(etiquetaRol("admin"), "Administrador del sistema")
  assert.equal(etiquetaRol("otro", "Rol Personalizado"), "Rol Personalizado")
  assert.equal(etiquetaRol("desconocido"), "desconocido")
})

test("capataz no puede cerrar ni cancelar labores ni encolar esos estados", () => {
  const permitidosCapataz = estadosPermitidos("capataz")
  assert.equal(permitidosCapataz.some((e) => e.id === "cerrada"), false)
  assert.equal(permitidosCapataz.some((e) => e.id === "cancelada"), false)
  assert.equal(permitidosCapataz.some((e) => e.id === "pendiente"), true)
  assert.equal(permitidosCapataz.some((e) => e.id === "en_proceso"), true)

  const permitidosCoord = estadosPermitidos("coordinacion")
  assert.equal(permitidosCoord.some((e) => e.id === "cerrada"), true)
  assert.equal(permitidosCoord.some((e) => e.id === "cancelada"), true)

  assert.equal(puedeEncolarEstado("capataz", "cerrada"), false)
  assert.equal(puedeEncolarEstado("capataz", "cancelada"), false)
  assert.equal(puedeEncolarEstado("capataz", "en_proceso"), true)
  assert.equal(puedeEncolarEstado("coordinacion", "cerrada"), true)
  assert.equal(puedeEncolarEstado("jefatura", "cancelada"), true)

  const htmlCapataz = renderToStaticMarkup(
    createElement(Labores, {
      rol: "capataz",
      equipos: [],
      equipoId: "cap-1",
      onEquipo: () => {},
      items: [labor],
      pinMode: false,
      onPinMode: () => {},
      draft: null,
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
      estadoNuevo: "pendiente",
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
      formEjecutor: "propia",
      motivos: [],
      motivo: "",
      onMotivo: () => {},
      onSugerir: () => {},
      sugerencia: "",
      pista: null,
      onConfirmarPista: () => {},
    }),
  )
  assert.equal(htmlCapataz.includes('value="cerrada"'), false)
  assert.equal(htmlCapataz.includes('value="cancelada"'), false)
  assert.equal(htmlCapataz.includes('value="pendiente"'), true)
  assert.match(htmlCapataz, /Por iniciar/)
  assert.match(htmlCapataz, /Ejecutado/)
  assert.equal(htmlCapataz.includes(">ejecutado<"), false)
  assert.equal(htmlCapataz.includes(">pendiente<"), false)
})

