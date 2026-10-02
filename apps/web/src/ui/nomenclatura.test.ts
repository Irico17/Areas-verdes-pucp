import assert from "node:assert/strict"
import { readdir, readFile } from "node:fs/promises"
import path from "node:path"
import { test } from "node:test"
import { fileURLToPath } from "node:url"
import { RIEGO, SECTOR_CAPATAZ, USO_FUENTE, etiquetaZonaSupervision } from "./nomenclatura.ts"

const SRC = fileURLToPath(new URL("..", import.meta.url))

const TEXTOS_USO = [
  "Uso Institucional",
  "Áreas de uso administrativo",
  "Áreas de uso recreativo/descanso",
  "Áreas de manejo sostenible y reducción de consumo de agua",
  "Áreas deportivas y recreación activa",
  "Áreas de conservación",
]

test("la cobertura de riego se dice provisional y no como fórmula oficial", () => {
  assert.equal(
    RIEGO.cobertura(50),
    "Cobertura provisional: 50 %. Pendiente de validar con la jefatura de sección.",
  )
  assert.equal(RIEGO.cobertura(0).startsWith("Cobertura provisional: 0 %."), true)
  assert.equal(RIEGO.lede.includes("definición pendiente"), false)
})

test("los usos y los sectores de capataz usan el texto oficial", () => {
  for (const texto of TEXTOS_USO) {
    assert.ok(Object.values(USO_FUENTE).includes(texto), texto)
  }
  assert.equal(USO_FUENTE["sin-uso"], "Sin uso")
  assert.match(SECTOR_CAPATAZ["cua-valeria"], /Sector de capataz — Valeria Quispe \(ficticio\)/)
  assert.match(SECTOR_CAPATAZ["cua-mateo"], /Mateo Salazar \(ficticio\)/)
  assert.match(SECTOR_CAPATAZ["cua-renato"], /Renato Cárdenas \(ficticio\)/)
  assert.equal(SECTOR_CAPATAZ["campo-deportivo"], "Campo deportivo")
  assert.equal(SECTOR_CAPATAZ["bosque-humedo"], "Bosque húmedo")
  assert.equal(etiquetaZonaSupervision("Z1"), "Zona 1")
  assert.equal(etiquetaZonaSupervision("Z4"), "Zona 4")
  assert.equal(etiquetaZonaSupervision("norte"), "norte")
})

type Regla = { id: string; re: RegExp; salvo?: RegExp }

const REGLAS: Regla[] = [
  { id: "labores", re: /\blabou?res?\b/i },
  { id: "equipo", re: /\bequipos?\b/i },
  { id: "zona-suelta", re: /\bzonas?\b/i, salvo: /zonas? de supervisión|zona [1-4]/gi },
  { id: "huellas", re: /huellas del recinto/i },
  { id: "osm", re: /edificios osm/i },
  { id: "puntos-pucp", re: /puntos pucp/i },
  { id: "sector-operativo", re: /sector(?:es)? operativ/i },
  { id: "sso", re: /\bSSO\b/ },
  { id: "universidad", re: /universidad/i },
  { id: "jefe-seccion", re: /jefe de sección/i },
  { id: "rol-sin-espacios", re: /Ingeniería\/Coordinación/ },
  { id: "uso-corto", re: /recreativo \/ descanso|(?<!de )manejo sostenible(?! y)/i },
  { id: "incidencia", re: /\bincidencias?\b/i },
  { id: "playas", re: /\bplayas\b(?! de estacionamiento)/i },
  { id: "puertas", re: /\bpuertas\b(?! y entradas)/i },
  { id: "supervisor", re: /\bsupervisor\b/i },
  { id: "operario", re: /\boperario\b/i },
  { id: "admin-suelto", re: /\bAdmin\b/ },
]

const CLASE = /^(?:[a-z][a-z0-9_-]*)(?: [a-z][a-z0-9_-]*)*$/

function visibles(fuente: string): string[] {
  const sin = fuente.replace(/\/\*[\s\S]*?\*\//g, "").replace(/(^|[^:\\])\/\/.*$/gm, "$1")
  const textos: string[] = []
  const literales = /`(?:\\[\s\S]|[^`])*`|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/g
  for (const coincidencia of sin.matchAll(literales)) {
    const crudo = coincidencia[0].slice(1, -1)
    for (const parte of crudo.split(/\$\{[^}]*\}/)) {
      if (seLee(parte)) textos.push(parte.trim())
    }
  }
  const jsx = />([^<>{}\n]+)</g
  for (const coincidencia of sin.matchAll(jsx)) {
    const texto = coincidencia[1].trim()
    if (seLee(texto)) textos.push(texto)
  }
  return textos
}

function seLee(crudo: string): boolean {
  const texto = crudo.trim()
  if (!texto || !/^[A-Za-zÁÉÍÓÚáéíóúÑñ¿¡]/.test(texto)) return false
  if (texto.includes("\n") || texto.includes("/") || texto.includes("${")) return false
  if (/[=;()[\]{}]/.test(texto)) return false
  if (CLASE.test(texto)) return false
  if (/^[a-z0-9_.:#-]+$/.test(texto)) return false
  if (/^[a-z][a-zA-Z0-9]*$/.test(texto)) return false
  return true
}

function viola(texto: string, regla: Regla): boolean {
  const limpio = regla.salvo ? texto.replace(regla.salvo, " ") : texto
  regla.re.lastIndex = 0
  return regla.re.test(limpio)
}

async function archivos(dir: string): Promise<string[]> {
  const entradas = await readdir(dir, { withFileTypes: true })
  const salida: string[] = []
  for (const entrada of entradas) {
    const ruta = path.join(dir, entrada.name)
    if (entrada.isDirectory()) salida.push(...(await archivos(ruta)))
    else if (/\.tsx?$/.test(entrada.name) && !entrada.name.endsWith(".test.ts")) salida.push(ruta)
  }
  return salida
}

test("el frontend no vuelve a mostrar un término prohibido por el glosario", async () => {
  const lista = await archivos(SRC)
  const fallos: string[] = []
  for (const archivo of lista) {
    const fuente = await readFile(archivo, "utf8")
    const textos = visibles(fuente)
    for (const texto of textos) {
      for (const regla of REGLAS) {
        if (viola(texto, regla)) {
          fallos.push(`${path.relative(SRC, archivo)} [${regla.id}]: ${texto.slice(0, 180)}`)
        }
      }
    }
  }
  assert.deepEqual(fallos, [])
})
