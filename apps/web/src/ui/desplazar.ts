const REDUCIDO = "(prefers-reduced-motion: reduce)"

function contenedorConScroll(el: Element): HTMLElement | null {
  for (let p = el.parentElement; p && !p.classList.contains("panel"); p = p.parentElement) {
    const oy = getComputedStyle(p).overflowY
    if ((oy === "auto" || oy === "scroll") && p.scrollHeight > p.clientHeight + 1) return p
  }
  return null
}

/** Desplaza solo el contenedor del panel, nunca el documento: la UI fija no se corre. */
export function mostrarEnPanel(el: Element | null | undefined, donde: "inicio" | "centro" | "cerca" = "inicio") {
  if (!el) return
  const caja = contenedorConScroll(el)
  if (!caja) return
  const r = el.getBoundingClientRect()
  const c = caja.getBoundingClientRect()
  const margen = 8
  let destino = caja.scrollTop + (r.top - c.top) - margen
  if (donde === "centro") destino = caja.scrollTop + (r.top - c.top) - (c.height - r.height) / 2
  if (donde === "cerca") {
    if (r.top >= c.top + margen && r.bottom <= c.bottom - margen) return
    if (r.bottom > c.bottom - margen) destino = caja.scrollTop + (r.bottom - c.bottom) + margen
  }
  caja.scrollTo({ top: Math.max(0, destino), behavior: window.matchMedia(REDUCIDO).matches ? "auto" : "smooth" })
}
