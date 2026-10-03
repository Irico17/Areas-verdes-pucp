import { existsSync, readFileSync } from "node:fs"
import { createRequire } from "node:module"
import { dirname, join } from "node:path"
import { fileURLToPath, pathToFileURL } from "node:url"

const ts = createRequire(import.meta.url)("typescript")
const extensiones = [".tsx", ".ts", ".js", ".mjs"]

export async function resolve(specifier, context, nextResolve) {
  if (specifier.startsWith(".") && !extensiones.some((ext) => specifier.endsWith(ext))) {
    const base = join(dirname(fileURLToPath(context.parentURL)), specifier)
    for (const ext of [".tsx", ".ts"]) {
      if (existsSync(base + ext)) {
        return { url: pathToFileURL(base + ext).href, shortCircuit: true }
      }
    }
  }
  return nextResolve(specifier, context)
}

export async function load(url, context, nextLoad) {
  if (!url.endsWith(".ts") && !url.endsWith(".tsx")) return nextLoad(url, context)
  const filename = fileURLToPath(url)
  const source = ts.transpileModule(readFileSync(filename, "utf8"), {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
      jsx: ts.JsxEmit.ReactJSX,
    },
    fileName: filename,
  }).outputText
  return { format: "module", source, shortCircuit: true }
}
