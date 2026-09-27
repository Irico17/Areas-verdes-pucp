# Auditoría de UI: scroll, layout, diseño y movimiento

**Superficie:** `apps/web` (React 19, Vite, MapLibre 6.11). Rutas relativas a `apps/web/`.
**Fecha:** 27/09/2026. **Rama sugerida:** `fix/ui-scroll-diseno`.
**Estado:** auditoría y plan, implementados. *(Nota de una fase posterior: los lotes A, B, C y D están implementados en la rama `fix/ui-scroll-diseno`; ver "Progreso" más abajo.)*

### Progreso (actualizado durante la implementación)

- **Lote A · completo.** S01–S19 implementados y verificados con `npm run lint`, `npm test`, `npm run build` en `apps/web`, y con capturas a 1280×800 y 390×844 (prefijo `loteA` en `/workspace/verdepucp-shots`).
- Fuera de la auditoría: se corrigió que `FichaLabor` (`Labores.tsx`) conservaba valores tipeados al cambiar de labor (riesgo de guardar en la labor equivocada). Ver O1 y el commit `fix: FichaLabor ya no arrastra datos tipeados de una labor a otra`.
- **Lote B · completo.** V01–V14 implementados en seis bloques (B1–B6) y verificados con `npm run lint`, `npm test`, `npm run build` en `apps/web`, y con capturas a 1280×800 y 390×844 (prefijo `loteB` en `/workspace/verdepucp-shots`).
- **Lote C · completo.** M01–M04 implementados y verificados con `npm run lint`, `npm test`, `npm run build` en `apps/web`, con capturas a 1280×800 y 390×844 (prefijo `loteC`) y con un script de Playwright aparte (`/workspace/shots-tool/verify-loteD.mjs`) que confirma, con `prefers-reduced-motion: reduce` emulado, que la hoja móvil solo transiciona `opacity, display` (sin traslado) y que los botones conservan `background, color, transform, border-color` en su lista de transición.
- **Lote D · verificado dentro de lo posible en este entorno.** D.1 (lint/test/build) en verde. D.3: sin desbordes horizontales ni contenedores recortados en ningún módulo (script equivalente ya integrado en `shots.mjs`, ver reporte `loteC` en la consola); sin objetivos táctiles bajo 44 px a 390 px en Mapa, Labores, Catastro e Inventario. D.2 (matriz completa de anchos), D.4 (regresión de gestos del mapa con puntero real) y D.6 (lector de pantalla) quedan como huecos: este entorno no tiene un iPhone o Android físico ni un lector de pantalla para probarlos; no bloquean el cierre del plan, tal como indica la guía del Lote D.
- **D.5 no se podía dar por cumplido.** Una revisión senior de `git diff main...HEAD` encontró que «crear una labor» y «revertir un lote» ya fallaban en `main` (no eran regresión de este lote): `operacion/store.go` rompía la creación de labores con `lugar_id` por un `CASE WHEN … ::bigint` que Postgres evaluaba antes del filtro, y `auditoria/store.go` (`leerActual`) no conocía las entidades de `bajaPorActivo` (lugares, tachos, etc.), así que revertir un lote de esas entidades deshacía toda la transacción. Ambos bugs, más uno de arrastre en `insertarCambio` (casteaba `antes`/`despues` nulos a `''` en vez de `NULL`), quedaron corregidos con tests de integración en verde (`go test ./...` en `apps/api`); D.5 queda pendiente de repetir a mano en un navegador real, pero ya no está bloqueada por estos bugs del backend.

## Cómo se hizo

- Lectura completa de `src/App.tsx`, `src/styles.css`, `src/ui/BottomSheet.tsx`, `src/ui/bottomSheet.ts`, `src/map/CampusMap.tsx`, `src/map/MapBoundary.tsx`, todos los `src/panel/*.tsx`, `src/FechaCampo.tsx`, `src/main.tsx`, `index.html`, los tests SSR y el CSS y la cámara de MapLibre (`node_modules/maplibre-gl/src/css/maplibre-gl.css`, `src/ui/map.ts`, `src/ui/camera.ts`).
- Criterio: `DESIGN.md`, `PRODUCT.md`, `.impeccable/surfaces/apps-web.md`, `docs/UI-AUDIT.md` y la guía Impeccable (`audit`, `adapt`, `layout`, `animate`, `operate`, `polish`). Modo **Operate**: la herramienta desaparece en la tarea.
- Evidencia: análisis estático del código y del modelo de layout (flex, grid, overflow, `touch-action`, viewport). En esta sesión no hubo permiso para ejecutar Chrome ni Node, así que **no hay capturas**. Los anchos de texto marcados con «≈» son estimaciones con métricas típicas de una grotesca a ese tamaño. El Lote D los confirma en navegador y en dispositivo.
- Severidad: **P0** bloquea una tarea. **P1** dificultad seria o falla AA. **P2** molestia con rodeo. **P3** acabado.

---

## 1. Resumen ejecutivo

La falla de scroll más grave es estructural, no de estilo. `App.tsx` envuelve el contenido de cada módulo en un `<div key={moduloActivo}>` que desarma la regla `.panel-view:has(.split)`. En **Catastro** e **Inventario**, los dos módulos de edición con panel partido, ni la lista ni la ficha tienen un alto acotado, y `.panel-view` recorta con `overflow: hidden`. No se puede bajar a la ficha de un área ni al botón **Guardar** del inventario, ni en escritorio ni en el teléfono (**S01, P0**).

En el teléfono se suman fallas independientes que explican el «se pone buggy»:

- iOS amplía la página al enfocar cualquier campo (15 px, menos de 16 px) y deja la interfaz fija corrida (S05).
- El teclado tapa el campo y la barra inferior queda encima (S06).
- En el anclaje máximo la hoja sube bajo el notch y tapa la barra superior, y el arrastre tiene una zona muerta (S03).
- El acceso no desplaza en pantallas bajas porque el bloqueo de scroll es global (S04).
- Con 8 o 9 módulos, la barra inferior monta los rótulos (S07).
- Al elegir una labor, su ficha queda fuera de vista (S08).

En diseño, el sistema de DESIGN.md está bien planteado pero tiene huecos visibles. Hay pestañas cuyo estado activo no se pinta (V01), anillos de foco recortados en los controles segmentados (V02), botones con el estilo del navegador (V03), sombra y radios fuera del sistema (V05), vacíos falsos durante la carga (V09) y `input type="date"`, que DESIGN.md prohíbe (V11). El movimiento es brusco justo en el elemento que lo necesita, la hoja móvil: aparece con un parpadeo, desaparece de golpe y salta al anclaje (M01). El modo de movimiento reducido apaga también los cambios de color que DESIGN.md pide conservar (M02).

### Puntuación técnica (Impeccable, audit)

| # | Dimensión | Nota | Hallazgo clave |
|---|---|---|---|
| 1 | Accesibilidad | 2 | Foco recortado en segmentados, estado activo invisible, objetivos de 32 px, Escape solo en el asa |
| 2 | Rendimiento | 3 | El arrastre re-estiliza todo el documento en cada movimiento. Nada grave |
| 3 | Responsive | 1 | Catastro e Inventario sin scroll. Teclado, zoom de iOS, notch, barra inferior |
| 4 | Tema | 3 | Tokens de color usados. Radios y z-index como literales |
| 5 | Integridad | 2 | Panel partido roto por un envoltorio. Botones sin estilo. Deriva de elevación y fechas |
| **Total** | | **11/20** | **Aceptable: trabajo significativo** |

**Hallazgos:** 37 en total. 1 P0, 9 P1, 21 P2 y 6 P3.

### Lo que funciona y se conserva

- Ya están `100dvh`, `env(safe-area-inset-*)` y `viewport-fit=cover`.
- Foco visible global de 2 px en yucca con 2 px de separación (DESIGN.md).
- Tokens de color en `:root`, un solo acento y cifras tabulares.
- Objetivos de 44–48 px en casi todos los controles, y zoom del mapa a 44 px en el teléfono.
- La lógica de anclajes de la hoja vive en funciones puras con test (`ui/bottomSheet.ts`).
- MapLibre 6 observa su contenedor con `ResizeObserver` (throttle de 50 ms). Ningún cambio de layout de este plan obliga a llamar `map.resize()`.
- Los controles del mapa ya suben por encima de la hoja con `--sheet-h`.

### Orden

El Lote A (scroll y layout) es condición para todo lo demás. B (visual) y C (movimiento) se apoyan en la estructura que deja A. D verifica a 390 px y a 1280 px.

---

## 2. Hallazgos

### 2.1 Tabla resumen

| ID | Sev | Dónde | Síntoma | Causa raíz | Corrección | Lote |
|---|---|---|---|---|---|---|
| S01 | P0 | `App.tsx:566,707`; `styles.css:145-149,303-317` | Catastro e Inventario no desplazan; ficha y **Guardar** inalcanzables (390 y 1280) | `<div key>` entre `.panel-view` y `.split`; ítem flex con `min-height: auto`; grilla con alto indefinido; recorte | Quitar el envoltorio; `key` en `.panel-view`; `:has(> .split)` y `flex: 1 1 0` | A |
| S02 | P1 | `styles.css:303-317`; `CatastroEditor.tsx:301-430`; `InventarioCapas.tsx:219-246` | En la hoja, la partición (mín. 392 px) no cabe; dos scrolls anidados | Patrón de escritorio sin variante | Un solo scroll en ≤820 px o alto ≤560 px; maestro-detalle en Catastro | A |
| S03 | P1 | `styles.css:470-490`; `BottomSheet.tsx:41-77,94` | Al 92 % la hoja sube bajo el notch y tapa la barra; zona muerta al arrastrar; sin tope | Anclajes contra `innerHeight`, tope CSS distinto, alto calculado y no medido, safe-area doble | Alto útil `--sheet-util`; medir y acotar el arrastre | A |
| S04 | P1 | `styles.css:417,548-553` | El acceso no desplaza en móvil apaisado o bajo | Bloqueo global `html, body, #root { overflow: hidden }` | Bloquear solo con `.shell` | A |
| S05 | P1 | `styles.css:20,31,219-226` | iPhone amplía al enfocar un campo; UI fija corrida | Campos a 15 px | 16 px en ≤820 px | A |
| S06 | P1 | `BottomSheet.tsx`; `styles.css:440-490` | El teclado tapa el campo; barra inferior sobre el teclado | Fijos contra el layout viewport; el teclado solo achica el visual viewport | Hook `useTeclado` y modo teclado de la hoja | A |
| S07 | P1 | `App.tsx:557-564`; `styles.css:440-469` | Con 8–9 módulos los rótulos se montan (≈53 px en 38–44 px) | Palabra indivisible en botón `flex: 1 1 0` | Cinco casillas como máximo y «Más» | A |
| S08 | P1 | `Labores.tsx:222-300`; `App.tsx:405-416,725-731` | Elegir una labor (lista o mapa) no muestra su ficha | Ficha debajo de la lista, sin desplazamiento | `mostrarEnPanel` al cambiar la selección | A |
| S09 | P2 | `BottomSheet.tsx:126` | Al cambiar de módulo se aterriza a media página | `.panel-view` persiste entre módulos | `key={vista}` en `.panel-view` | A |
| S10 | P2 | `styles.css:138-144,255,529`; `Importaciones.tsx:81,126-135` | Textos largos ensanchan el panel (1280) o se cortan sin paneo (390); vista previa ilegible | Scroll en ambos ejes, `pan-y`, `1fr` sin `minmax(0)`, vista previa en párrafos | `overflow: hidden auto`, `overflow-wrap: anywhere`, tabla en `.tabla-scroll` | A |
| S11 | P2 | `styles.css:96-106,411-414` | El riel desborda en alturas menores a ≈570 px y en apaisado; el notch tapa el riel | `.guard` sin overflow; sin variante de alto corto | Scroll propio del riel; regla de alto corto; safe-area izquierda y derecha | A |
| S12 | P2 | `styles.css:138-144,310-316` | «Tirar para recargar» en Android; rebote en iOS; salto al aparecer la barra | Sin `overscroll-behavior` ni `scrollbar-gutter` | `contain` en contenedores, `none` en raíz, gutter estable | A |
| S13 | P2 | `styles.css:424-439,567-570` | A 390 px la marca se monta sobre «Plano» (≈374 px en 350) | Regla compacta solo en ≤380 px | Compactar en ≤420 px | A |
| S14 | P2 | `CampusMap.tsx:453-460` | En el teléfono la labor elegida queda detrás de la hoja | `flyTo` centra en el contenedor completo | `offset` según `--sheet-h` (nunca `padding`) | A |
| S15 | P2 | `BottomSheet.tsx:53-92,105-121` | Escape solo con foco en el asa; el foco no vuelve; hoja trabada si se pierde la captura; tocar el asa no hace nada | Manejadores en el asa; sin `lostpointercapture`; sin umbral | Escape en el `aside`; umbral de 4 px; el toque cicla anclajes | A |
| S16 | P2 | `styles.css:128-137,418-422,454` | La UI queda corrida tras un foco o un `scrollIntoView` | `overflow: hidden` sigue siendo contenedor de scroll | `overflow: clip` con respaldo | A |
| S17 | P2 | `CatastroEditor.tsx:394-426` | «Nueva área» y «Exportar CSV» al final de 521 filas | Orden del JSX | Subirlos bajo la búsqueda | A |
| S18 | P3 | `styles.css:529` | No se puede ampliar con pellizco en la hoja | `touch-action: pan-y` | `pan-y pinch-zoom` | A |
| S19 | P3 | `App.tsx:114,535`; `BottomSheet.tsx:4`; `styles.css:54,335` | Hueco entre 820 y 821 px; `dvh` sin respaldo; controles desfasados tras rotar; «Saltar al panel» a un panel oculto | Corte duplicado en JS; sin respaldo; sin escucha de `resize` | `ui/media.ts`, respaldo `vh`, recálculo, abrir el panel | A |
| V01 | P1 | `CatastroEditor.tsx:304-311`; `CalendarioReservas.tsx:82-88`; `styles.css:182` | Áreas/Zonas y Mes/Semana/Día no muestran la opción activa | El CSS solo pinta `aria-pressed` | `role="group"` y `aria-pressed`; el CSS acepta ambos | B |
| V02 | P1 | `styles.css:164-170,180-181,318-319` | Anillo de foco invisible en segmentados; pestaña «reservas» cortada a 821–1199 px | `.roles { overflow: hidden }`; botones sin `min-width: 0` | Sin recorte; radios en los extremos; foco interior | B |
| V03 | P2 | `CalendarioReservas.tsx:91-97`; `InventarioCapas.tsx:239`; `CatastroEditor.tsx:325` | Botones con el estilo del navegador | Sin clase ni regla base | Base `:where(.panel-view button)`; filas como `.labor` | B |
| V04 | P2 | `App.tsx:567-571`; `BottomSheet.tsx` | La línea de estado (conteos, sin conexión, área tocada) se va con el scroll | No hay cabeza fija (DESIGN.md) | Prop `cabeza` y `.panel-head` | B |
| V05 | P2 | `styles.css:317,480,498,515,635-638,658-665` | Sombra dentro del panel; hoja a 16 px; chip y «Elegir archivo» en esquina viva; asa en marino | Literales fuera del sistema | Tokens `--r-*`; sin sombra; radios 6/10/4 | B |
| V06 | P2 | `Labores.tsx:344`; `styles.css` | La ficha de labor lleva el borde «groove» del navegador | `fieldset` sin normalizar | Reset de `fieldset` y clase `.grupo` | B |
| V07 | P2 | `styles.css:109-126,171-191,264-277` | Sin hover en botones quietos ni filas; filas y navegación se encogen al tocar; destello gris | Vocabulario de interacción incompleto | Hover con `hover: hover`; presión solo en botones | B |
| V08 | P2 | `App.tsx:572-699`; `styles.css` | Labores, Riego, Poda y Vivero pegados; Capas y Reservas igual | `.block` sin estilos | 32 px y una línea entre secciones | B |
| V09 | P2 | `Labores.tsx:203`; `Modulos.tsx:287,596,738`; `Poda.tsx:31`; `Vivero.tsx:48`; `EvidenciasCampo.tsx:221`; `InventarioCapas.tsx:237` | «No hay…» mientras carga | Sin estado de carga | `Esqueleto` y carga derivada | B |
| V10 | P2 | `styles.css:208-216,332`; `CatastroEditor.tsx`; `FechaCampo.tsx:14` | Errores como texto rojo suelto; campos inválidos sin marca | Sin estilo de aviso ni de `aria-invalid` | Aviso de error con tono; borde en alerta | B |
| V11 | P2 | `Labores.tsx:352,356`; `Poda.tsx:102,106`; `Vivero.tsx:98` | Fechas en mm/dd/aaaa según el navegador | `type="date"` (DESIGN.md lo prohíbe) | `FechaCampo` con sincronía externa | B |
| V12 | P2 | `styles.css:229-231,253-263` | Filtros de estado y capas con objetivos de 32–38 px en el teléfono | `min-height: 32px` | 44 px con `pointer: coarse` | B |
| V13 | P3 | `InventarioCapas.tsx:221-254` | Pestañas en minúscula («tachos»), claves técnicas en pantalla, título invertido | Copia provisional | Rótulos, pista legible, `h2` arriba | B |
| V14 | P3 | `styles.css` (varias) | Marca a 700 sin fuente cargada; listas sin línea de cierre; cerrar de los popups pequeño; z-index sueltos | Detalles | Ver detalle | B |
| M01 | P2 | `styles.css:61,483,572-575`; `BottomSheet.tsx:101` | La hoja parpadea al abrir, se va de golpe y salta al anclaje | `display: none` sin salida; sin transición de alto | `@starting-style` y `allow-discrete`; alto en 260 ms | C |
| M02 | P2 | `styles.css:617-622`; `BottomSheet.tsx:10-12` | El movimiento reducido apaga también el color | Regla global `transition: none` | Quitar solo el desplazamiento; la hoja entra con fundido | C |
| M03 | P3 | `styles.css:143,264-277,361-369` | Cambio de módulo sin transición; selección y avisos en seco; esqueleto con brillo fuerte | `rise` solo al montar; brillo con curva de estado | `rise` por módulo; fundidos cortos; pulso de opacidad | C |
| M04 | P3 | `styles.css:537-540`; `BottomSheet.tsx:41-51,64` | Los controles del mapa saltan; el arrastre re-estiliza todo el documento | Variable en `:root` por movimiento | Variable en `.stage` y transición sincronizada | C |

### 2.2 Mapa de contenedores de scroll (objetivo)

Es el modelo que deja el Lote A. Cualquier contenedor con scroll fuera de esta tabla es un defecto.

| Región | Ancho ≥821 px y alto >560 px | Ancho ≤820 px, o alto ≤560 px |
|---|---|---|
| Documento (`html`, `body`) | Bloqueado mientras existe `.shell`. En el acceso desplaza | Igual |
| `.shell`, `.panel` | Sin scroll (`overflow: clip`) | Sin scroll (`overflow: clip`) |
| `.stage` | Sin scroll; no se toca | Igual |
| `.guard` (riel o barra) | Scroll vertical propio si no cabe | En ≤820 px, barra fija de cinco casillas, sin scroll (`clip`). En escritorio de alto corto sigue siendo el riel, con su scroll |
| `.panel-head` | Fija | Fija, en una línea |
| `.panel-view` | Scroll vertical. En módulos partidos, sin scroll | Único scroll vertical de la hoja |
| `.split-list`, `.split-detail` | Scroll vertical propio cada uno | Sin scroll: fluyen dentro de `.panel-view` |
| `.tabla-scroll` | Scroll horizontal | Scroll horizontal (`pan-x`) |
| `.map-host` | MapLibre (`touch-action: none` en el lienzo) | Igual |

### 2.3 Scroll y layout

#### S01 · P0 · Catastro e Inventario: el panel partido no desplaza

**Dónde.** `src/App.tsx:566` y `:707` (el `<div key={moduloActivo}>`), `src/ui/BottomSheet.tsx:126`, `src/styles.css:145-149` y `:303-317`. Lo sufren `CatastroEditor.tsx:301` e `InventarioCapas.tsx:219`.

**Síntoma.**
- 1280×800: en Catastro, la lista de 521 áreas no desplaza, o desplaza un tramo y se detiene. La ficha, que DESIGN.md quiere «visible debajo, con su propio scroll», nunca aparece. En Inventario → Tachos el formulario tiene 16 campos (≈1 300 px) y se corta: **Guardar** y **Eliminar** no se alcanzan.
- 390×844: lo mismo dentro de la hoja.

**Causa raíz.**
1. `.panel-view:has(.split)` (L145-149) vuelve a `.panel-view` flex en columna con `overflow: hidden`, suponiendo que `.split` es su hijo directo.
2. App mete todo el módulo en `<div key={moduloActivo}>` (L566). Ese div es el ítem flex. Como su `overflow` es `visible`, su `min-height: auto` equivale a su contenido y no se encoge.
3. `.split { flex: 1; min-height: 0 }` no tiene efecto: su padre, el div, no es flex.
4. `grid-template-rows: minmax(160px, 42%) minmax(220px, 1fr)` se resuelve contra un alto indefinido y las filas toman el alto del contenido (≈27 000 px para 521 filas). `.split-list` y `.split-detail` nunca quedan más bajos que su contenido, así que su `overflow: auto` no actúa, o actúa sobre una caja de miles de píxeles cuya parte baja ya está recortada.
5. `.panel-view` recorta todo con `overflow: hidden`.

**Corrección.**

`src/App.tsx`: quitar el envoltorio y pasar el módulo a la hoja.

```tsx
<BottomSheet open={railOpen} onClose={() => setRailOpen(false)} vista={moduloActivo}>
  <p className={load.kind === "error" || activityError ? "status error" : "status"}>…</p>
  {moduloActivo === "mapa" && (…)}
  {/* … resto de módulos, sin <div key> alrededor … */}
  {moduloActivo === "admin" && <AdminPanel />}
</BottomSheet>
```

El paso A4 agrega `onAltura`, y el A6 cambia `onClose` por `cerrarPanel`. En el Lote B la línea de estado pasa a la prop `cabeza` (V04).

`src/ui/BottomSheet.tsx`: la prop `vista` es la `key` del cuerpo. Esto también resuelve S09 y M03.

```tsx
<div className="panel-view" key={vista}>{children}</div>
```

`src/styles.css`: reemplazar L145-149 y L303-317.

```css
.panel-view:has(> .split) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.panel-view:has(> .split) > * { flex: none; }
.panel-view:has(> .split) > .split {
  flex: 1 1 0;
  min-height: 0;
}
.split {
  display: grid;
  grid-template-rows: minmax(160px, 42%) minmax(0, 1fr);
  gap: 12px;
}
.split-list,
.split-detail {
  min-height: 0;
  overflow: hidden auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  background: var(--hoja);
  border: 1px solid var(--linea);
  border-radius: var(--r-hoja);
  padding: 12px;
}
/* Sin box-shadow en .split-detail: se retira aquí; el motivo está en V05. */
```

`flex: 1 1 0` da a `.split` un alto definido, y con él se resuelven el 42 % y el `1fr`. La segunda fila pasa de `minmax(220px, 1fr)` a `minmax(0, 1fr)`. Con poco alto manda S02, no un mínimo que desborda.

**Acepta si** (1280×800): en Catastro la lista llega a la última área con la rueda, la ficha desplaza hasta **Dar de baja**, y `document.querySelector('.panel-view').scrollTop` sigue en 0. En Inventario → Tachos, **Guardar** se alcanza desplazando la ficha.

#### S02 · P1 · La partición no cabe en la hoja móvil ni en alturas cortas

**Dónde.** `styles.css:303-317` (sin variante), `CatastroEditor.tsx:301-430`, `InventarioCapas.tsx:219-246`.

**Síntoma.** A 390×844, aun corregido S01, la grilla pide 160 + 12 + 220 = 392 px más la línea de estado. El cuerpo útil de la hoja es de ≈156 px al 28 % y ≈384 px al 55 %, así que la ficha queda recortada. Además, dos zonas con scroll dentro de una hoja que se arrastra atrapan el dedo: el gesto mueve la lista, no la hoja ni la ficha. En 844×390 apaisado (usa el layout de escritorio) la fila de la ficha queda en ≈70 px.

**Causa raíz.** El panel partido es un patrón de escritorio sin variante para pantallas bajas.

**Corrección.** Un solo scroll (`.panel-view`) cuando el ancho es ≤820 px o el alto es ≤560 px, y navegación maestro-detalle en Catastro.

CSS, después del bloque de S01:

```css
@media (max-width: 820px), (max-height: 560px) {
  .panel-view:has(> .split) { display: block; overflow: hidden auto; }
  .split { display: block; }
  .split-list,
  .split-detail {
    overflow: visible;
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
  }
  .split-detail { margin-top: 16px; }
  .split[data-vista="ficha"] > .split-list,
  .split[data-vista="lista"] > .split-detail { display: none; }
}
.volver-lista { display: none; }
@media (max-width: 820px), (max-height: 560px) {
  .volver-lista { display: inline-flex; }
}
```

`CatastroEditor.tsx`:

```tsx
const [verLista, setVerLista] = useState(false)
const seleccion = entidad === "area" ? area : zona
const vista = seleccion && !verLista ? "ficha" : "lista"
const listaRef = useRef<HTMLDivElement>(null)
const fichaRef = useRef<HTMLDivElement>(null)

// Otra ficha: empieza arriba (escritorio: .split-detail; móvil: .panel-view).
useEffect(() => {
  if (vista !== "ficha") return
  fichaRef.current?.scrollTo({ top: 0 })
  mostrarEnPanel(fichaRef.current, "inicio")
}, [vista, idActivo])

// De vuelta en la lista: la fila elegida a la vista, o el inicio si no hay.
useEffect(() => {
  if (vista !== "lista") return
  const fila = listaRef.current?.querySelector(".labor.on")
  mostrarEnPanel(fila ?? listaRef.current, fila ? "centro" : "inicio")
}, [vista])

// Al editar la geometría, los controles de vértice quedan arriba del cuerpo visible.
useEffect(() => {
  if (editandoGeom) mostrarEnPanel(fichaRef.current?.querySelector(".vertice-box"), "inicio")
}, [editandoGeom])
```

- `<section className="split catastro-editor" data-vista={vista}>`, `ref={listaRef}` en `.split-list`, `ref={fichaRef}` en `.split-detail`.
- En cada `onClick` que elige o crea (filas de áreas y de zonas, **Nueva área**, **Nueva zona**) agregar `setVerLista(false)`.
- Primera línea de `.split-detail`: `<button type="button" className="link volver-lista" onClick={() => setVerLista(true)}>Volver a la lista</button>`.
- El aviso (L678) se muestra dentro de `.split-detail` solo si hay selección (`seleccion && aviso`). Si no la hay, por ejemplo después de **Dar de baja**, se muestra al inicio de `.split-list`, bajo la búsqueda. Así en el teléfono la baja no queda sin respuesta.
- `verLista` es solo presentación. No borra `area` ni `zona`, y en escritorio se siguen viendo ambos lados.

`InventarioCapas.tsx` (la lista es corta; basta saltar al formulario):

```tsx
const formRef = useRef<HTMLFormElement>(null)
const listaRef = useRef<HTMLDivElement>(null)
const [salto, setSalto] = useState(0)
useEffect(() => {
  if (salto > 0) mostrarEnPanel(formRef.current, "inicio")
}, [salto])
useEffect(() => {
  formRef.current?.closest(".split-detail")?.scrollTo({ top: 0 })   // escritorio: otra entidad empieza arriba
}, [entidad])
// En cargarFila(): setSalto((n) => n + 1)
```

- `ref={listaRef}` en `.split-list` y `ref={formRef}` en el `<form>`.
- Primera línea del formulario: `<button type="button" className="link volver-lista" onClick={() => mostrarEnPanel(listaRef.current?.querySelector(".labor.on") ?? listaRef.current, "centro")}>Volver a la lista</button>`.
- En escritorio `mostrarEnPanel` encuentra `.split-detail` como contenedor y deja el formulario arriba. En el teléfono desplaza `.panel-view`.

`mostrarEnPanel` vive en `src/ui/desplazar.ts` (paso A1).

**Acepta si** (390×844): tocar un área muestra su ficha arriba de la hoja; **Volver a la lista** regresa con la fila elegida a la vista; la ficha llega a **Guardar área** al 55 % y al 92 %; al activar **Editar geometría** los botones Oeste/Este/Norte/Sur quedan visibles. En Inventario, tocar un registro lleva al formulario y **Guardar** se alcanza. A 1280×540 el comportamiento es el mismo, de un solo scroll.

#### S03 · P1 · Geometría de la hoja móvil

**Dónde.** `styles.css:470-490`, `BottomSheet.tsx:41-77` y `:94`, `bottomSheet.ts:1-2`.

**Síntoma** (iPhone 14: 390×844, safe-area de 47 px arriba y 34 px abajo):
- Al 92 %: alto pedido 0,92 × 844 = 776 px, tope CSS `100dvh − 64 − 34` = 746 px, y la hoja arranca en y = 0. El asa y **Cerrar** quedan bajo la isla o la barra de estado, y la barra superior (Plano/Relieve, Salir) queda tapada. DESIGN.md pide identidad y Salir visibles.
- Al arrastrar desde el 92 %, los primeros ≈30 px no mueven nada: `alBajar` parte de 776 px calculados, no de los 746 px reales.
- Hacia arriba el alto crece sin tope, `--sheet-h` sube y los controles del mapa salen de la pantalla.
- `.panel.sheet { padding-bottom: env(safe-area-inset-bottom) }` suma otra vez los 34 px que la barra inferior ya reserva: queda una franja vacía al pie de la hoja.
- `.panel { height: min(74dvh, 680px) }` (L478) nunca se usa: la pisan `.panel.sheet` y el estilo en línea.

**Causa raíz.** Los anclajes se calculan contra `window.innerHeight` (JS) y `dvh` (CSS), mientras el tope real es otro. No hay «zona de la barra superior». El arrastre no mide ni acota. El comentario de `bottomSheet.ts` ya dice «alturas respecto al espacio útil», pero el código no lo aplica.

**Corrección.**

Tokens en `:root` (paso A1):

```css
--nav-h: 64px;
--topbar-zona: calc(env(safe-area-inset-top) + 64px);   /* 8 + 48 + 8 */
--sheet-util: calc(100dvh - var(--nav-h) - env(safe-area-inset-bottom) - var(--topbar-zona));
```

En teléfonos apaisados bajos la hoja puede cubrir la barra superior en el máximo:

```css
@media (max-width: 820px) and (max-height: 480px) {
  :root { --topbar-zona: env(safe-area-inset-top); }
}
```

Bloque móvil: reemplazar L470-490.

```css
.panel {
  position: fixed;
  z-index: var(--z-hoja);
  left: 0;
  right: 0;
  bottom: calc(var(--nav-h) + env(safe-area-inset-bottom));
  border-right: 0;
  border-radius: var(--r-hoja) var(--r-hoja) 0 0;
  box-shadow: var(--sombra);
  background: var(--papel);
}
.panel.sheet {
  height: calc(var(--sheet-util) * 0.55);   /* respaldo; el alto real va en línea */
  max-height: var(--sheet-util);
  touch-action: none;
}
```

Sin `padding-bottom` y sin `animation`: el movimiento va en M01.

`BottomSheet.tsx`: ver la versión objetivo en el paso A4. El alto en línea es `calc(var(--sheet-util) * ${anclaje})`. Al bajar el dedo se miden `getBoundingClientRect().height` y el tope (`parseFloat(getComputedStyle(hoja).maxHeight)`); al mover, el alto se acota a `[48, tope]`; al soltar, se aplica `anclajeCercano(alto / tope)`. `alturaPx(anclaje, tope)`, que hoy no se usa, da el alto que se publica para el mapa. No cambian `ANCLAJES` ni las firmas de `bottomSheet.ts`, que tienen test.

**Acepta si** (390×844 emulado, y en un iPhone real): al 92 % el borde alto de la hoja queda bajo la barra superior, que sigue visible y operable; el asa responde desde el primer píxel; arrastrar hacia arriba se detiene en el tope; no hay franja vacía al pie.

#### S04 · P1 · El acceso no desplaza en el teléfono

**Dónde.** `styles.css:417` y `:548-553`.

**Síntoma.** A 740×360 (Android apaisado, entra en ≤820 px), o en cualquier pantalla baja con el teclado abierto, la franja de marca más el formulario superan el alto y no hay scroll. **Entrar** puede quedar fuera.

**Causa raíz.** `html, body, #root { overflow: hidden }` vale para toda pantalla de 820 px o menos, también el acceso. `.gate { overflow: auto }` no ayuda: `.gate` crece con `min-height` y quien recorta es `#root`.

**Corrección.** Quitar L417 y el `overflow: auto` de `.gate` (L552). Bloquear solo cuando existe la cáscara de la app, fuera de toda media query:

```css
html:has(.shell),
html:has(.shell) body {
  overflow: hidden;
  overscroll-behavior: none;
}
```

**Acepta si** (740×360, y 390×844 con teclado): el acceso desplaza hasta **Entrar**; dentro de la app el documento no desplaza (`document.scrollingElement.scrollTop === 0`).

#### S05 · P1 · iOS amplía la página al enfocar un campo

**Dónde.** `styles.css:20` (15 px), `:31` (`font: inherit`), `:219-226`.

**Síntoma.** En iPhone (Safari o PWA), al tocar cualquier campo la página se amplía. La barra, la hoja y el mapa quedan corridos y hay que pellizcar para volver. Es la explicación más probable del «se pone buggy» en iPhone.

**Causa raíz.** Safari amplía todo campo con fuente menor a 16 px.

**Corrección** (bloque móvil):

```css
input, select, textarea { font-size: 16px; }
```

No usar `maximum-scale=1` ni `user-scalable=no` en el viewport: bloquean el zoom a quien lo necesita.

**Acepta si** (iPhone real, 390): enfocar Título, Detalle, Buscar y una fecha no cambia la escala (`visualViewport.scale === 1`).

#### S06 · P1 · El teclado tapa el campo y deja la barra inferior encima

**Dónde.** `BottomSheet.tsx` (no escucha `visualViewport`), `styles.css:440-490`.

**Síntoma** (teléfono real). En Labores → Detalle, o en Poda → Comentario (al final del formulario), el teclado cubre el campo y los botones. La hoja sigue anclada al pie del layout viewport, detrás del teclado. En iOS, además, el visual viewport se desplaza y arrastra toda la interfaz fija.

**Causa raíz.** Los elementos fijos se ubican contra el layout viewport. En iOS, y en Chrome Android desde la versión 108 por omisión, el teclado solo achica el visual viewport.

**Corrección.** Hook `useTeclado` (`src/ui/teclado.ts`, paso A1) y modo teclado de la hoja (paso A5). Mientras hay un campo enfocado y el teclado tapa más de 120 px, la hoja ocupa exactamente el visual viewport (`top = offsetTop`, `height = visualViewport.height`), se ocultan la barra inferior y la superior, y el campo enfocado se lleva a la vista con `mostrarEnPanel(…, "cerca")`.

```css
@media (max-width: 820px) {
  html[data-teclado] .guard,
  html[data-teclado] .topbar { visibility: hidden; }
  .panel.sheet.con-teclado {
    max-height: none;
    border-radius: 0;
    transition: none;
  }
  .panel.sheet.con-teclado .sheet-action { position: static; box-shadow: none; }
}
```

No usar `interactive-widget=resizes-content`: haría que el teclado cambie el alto de `.shell` y redimensione el lienzo del mapa en cada apertura.

**Acepta si** (iPhone y Android reales, 390): con el teclado abierto se ven el campo enfocado y su etiqueta, la barra inferior no aparece sobre el teclado y **Cerrar** sigue arriba; al cerrar el teclado la hoja vuelve a su anclaje. Si el hook falla en algún equipo, el mínimo aceptable es ocultar la barra inferior.

#### S07 · P1 · Barra inferior: los rótulos se montan con seis o más módulos

**Dónde.** `App.tsx:557-564`, `styles.css:440-469`.

**Síntoma** (390 px). Coordinación tiene 8 módulos: (390 − 8) / 8 − 4 ≈ 43,8 px de texto por botón. Admin tiene 9: ≈38,4 px. «Solicitudes» (≈53 px a 11 px y peso 500), «Inventario» (≈48 px) y «Catálogos» (≈48 px) no caben y, como son una sola palabra, desbordan centrados sobre los vecinos. A 360 px es peor. Jefatura (7 módulos) queda al límite.

**Causa raíz.** El botón tiene `flex: 1 1 0` y `min-width: 0`, pero el texto es un ítem flex anónimo con `min-width: auto`, igual al ancho de la palabra. `justify-content: center` lo desborda hacia ambos lados.

**Corrección.** Cinco casillas como máximo en el teléfono: los cuatro primeros módulos permitidos y «Más», que despliega el resto. Si el módulo activo está en «Más», la quinta casilla muestra su nombre y se pinta activa. El código está en el paso A6. CSS:

```css
@media (max-width: 820px) {
  .guard-menu {
    position: fixed;
    right: calc(8px + env(safe-area-inset-right));
    bottom: calc(var(--nav-h) + env(safe-area-inset-bottom) + 8px);
    z-index: var(--z-menu);
    display: grid;
    min-width: 208px;
    padding: 4px;
    background: var(--hoja);
    border: 1px solid var(--linea);
    border-radius: var(--r-hoja);
    box-shadow: var(--sombra);
  }
  .guard-menu button {
    min-height: 44px;
    padding: 10px 12px;
    border: 0;
    border-radius: var(--r-control);
    background: transparent;
    text-align: left;
    font-weight: 500;
  }
  .guard-menu button[aria-current="page"],
  .guard button.activo {
    background: color-mix(in srgb, var(--yucca) 14%, var(--hoja));
    color: var(--yucca-deep);
    font-weight: 600;
  }
}
```

En escritorio el riel sigue mostrando todos los módulos. `modulosDe(rol)` no cambia.

**Acepta si** (390 y 360, roles coordinación y admin): ninguna casilla desborda (`[...document.querySelectorAll('.guard button')].every(b => b.scrollWidth <= b.clientWidth)`); «Más» abre la lista; Escape y tocar fuera la cierran y el foco vuelve a «Más»; capataz (4 módulos) no ve «Más».

#### S08 · P1 · Elegir una labor no lleva a su ficha

**Dónde.** `Labores.tsx:222-300`, `App.tsx:405-416` (`choose`) y `:725-731` (clic en el mapa).

**Síntoma.** A 390, tocar un pin abre la hoja al 55 % en Labores, arriba de todo (filtros, botón, lista). La ficha de la labor está debajo de la lista y no se ve. A 1280, con más de 10 o 15 labores, pasa lo mismo al elegir desde el mapa o desde el final de la lista.

**Causa raíz.** La ficha se pinta debajo de la lista y nadie desplaza el panel.

**Corrección** (`Labores.tsx`):

```tsx
const detalleRef = useRef<HTMLDivElement>(null)
const elegida = props.selected?.id
useEffect(() => {
  if (elegida) mostrarEnPanel(detalleRef.current, "inicio")
}, [elegida])
// …
<div className="detail" ref={detalleRef}>
```

No agregar `key` al `.detail` (ver O1).

**Acepta si** (390 y 1280): elegir una labor en la lista o en el mapa deja el título de su ficha arriba del panel, con desplazamiento suave, o instantáneo con movimiento reducido.

#### S09 · P2 · Al cambiar de módulo se hereda el scroll

**Dónde.** `BottomSheet.tsx:126`.

**Síntoma.** Si Labores está desplazado 2 000 px y se pasa a Reportes, el panel queda al final de Reportes.

**Causa raíz.** `.panel-view` es el mismo elemento para todos los módulos.

**Corrección.** `key={vista}` en `.panel-view` (S01). Al montarse de nuevo empieza en 0 y repite `rise` (M03).

**Acepta si:** cada cambio de módulo empieza arriba, y cerrar y abrir la hoja sin cambiar de módulo conserva la posición.

#### S10 · P2 · Desborde horizontal y vista previa del importador

**Dónde.** `styles.css:138-144` (`overflow: auto` en ambos ejes), `:255` (`.layer` con `1fr`), `:529` (`touch-action: pan-y`); `Importaciones.tsx:81` y `:126-135`.

**Síntoma.** A 1280, un valor sin espacios (UUID, GeoJSON, URL) en la vista previa del importador, en un nombre de archivo de evidencia o en una fila ensancha `.panel-view`: aparece scroll horizontal del panel entero y el trackpad lo mueve de lado. A 390 el mismo contenido se corta y no se puede ver, porque `pan-y` impide el paneo horizontal. La vista previa muestra las filas como párrafos «col: valor · col: valor», limitados a 6 columnas.

**Causa raíz.** El panel desplaza en ambos ejes, las palabras largas no se cortan, hay ítems flex y grid con mínimo automático, y los datos tabulares no tienen un contenedor horizontal propio.

**Corrección.**

```css
.panel-view {
  overflow: hidden auto;            /* x recortado, y con scroll */
  overflow-wrap: anywhere;
}
.layer { grid-template-columns: 14px minmax(0, 1fr) auto; }
.labor > span { min-width: 0; }

.tabla-scroll {
  overflow: auto hidden;            /* x con scroll, y recortado */
  overscroll-behavior-x: contain;
  touch-action: pan-x pan-y;
  margin: 8px 0 16px;
  border: 1px solid var(--linea);
  border-radius: var(--r-control);
  background: var(--hoja);
}
.tabla { border-collapse: separate; border-spacing: 0; min-width: 100%; font-size: 13px; }
.tabla th,
.tabla td {
  padding: 8px 10px;
  text-align: left;
  white-space: nowrap;
  border-bottom: 1px solid var(--linea);
}
.tabla th { background: var(--papel); font-size: 12px; font-weight: 600; }
.tabla tbody tr:last-child td { border-bottom: 0; }
.tabla :is(th, td):first-child {
  position: sticky;
  left: 0;
  background: var(--hoja);
  box-shadow: 1px 0 0 var(--linea);
}
.tabla th:first-child { background: var(--papel); }
.tabla .celda { display: block; max-width: 28ch; overflow: hidden; text-overflow: ellipsis; }
```

`Importaciones.tsx`. Solo cambia la presentación: `previsualizar`, `confirmarImportacion`, `revertirLote` y `window.confirm` quedan igual.

```tsx
const columnas = vista?.filas[0] ? Object.keys(vista.filas[0]) : []
// …
{vista.filas.length > 0 && (
  <>
    <h3>Primeras {vista.filas.length} filas</h3>
    <div className="tabla-scroll" role="region" aria-label="Vista previa del archivo" tabIndex={0}>
      <table className="tabla">
        <thead>
          <tr>{columnas.map((col) => <th key={col} scope="col">{col}</th>)}</tr>
        </thead>
        <tbody>
          {vista.filas.map((fila, i) => (
            <tr key={i}>
              {columnas.map((col) => {
                const valor = String(fila[col] ?? "")
                return <td key={col}><span className="celda" title={valor}>{valor}</span></td>
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  </>
)}
{vista.errores.length > 0 && (
  <ul className="labor-list">
    {vista.errores.map((item, i) => (
      <li key={`${item.fila}-${item.campo}-${i}`} className="agenda">
        <strong>Fila {item.fila} · {item.campo}</strong>
        <small>{item.motivo}</small>
      </li>
    ))}
  </ul>
)}
```

Mostrar todas las columnas (antes `slice(0, 6)`) es solo lectura: la tabla desplaza de lado.

**Acepta si:** a 390 y a 1280 el detector de desbordes (D.3) devuelve `[]` en todos los módulos; la vista previa desplaza de lado con el dedo y con Shift + rueda; el panel no se mueve de lado.

#### S11 · P2 · Riel de escritorio en alturas cortas y teléfonos apaisados

**Dónde.** `styles.css:96-106`, `:411-414`.

**Síntoma.** El riel del rol admin necesita ≈514 px (9 botones de 44 px, la marca y los huecos). Con una ventana de menos de ≈570 px de alto (1280×550, pantalla partida) o en 844×390 apaisado, que usa el layout de escritorio, los botones se salen de la grilla y aparece scroll en el documento: se mueven la barra y el mapa. En apaisado, la muesca tapa el riel y los controles del mapa de la derecha.

**Causa raíz.** `.guard` no tiene `overflow`, no hay variante de alto corto, y la safe-area izquierda y derecha solo se aplica en el teléfono vertical.

**Corrección.**

```css
.guard {
  overflow: hidden auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  padding-left: max(8px, env(safe-area-inset-left));
}
.topbar {
  padding-left: max(16px, env(safe-area-inset-left));
  padding-right: max(12px, env(safe-area-inset-right));
}
@media (min-width: 821px) {
  .maplibregl-ctrl-bottom-right {
    margin-right: env(safe-area-inset-right);
    margin-bottom: env(safe-area-inset-bottom);
  }
}
@media (min-width: 821px) and (max-height: 560px) {
  .shell { grid-template-rows: 48px minmax(0, 1fr); }
  .rail-mark, .brand-sub { display: none; }
}
```

Las reglas móviles de `.guard` y `.topbar`, que van después en el archivo, siguen mandando en ≤820 px.

**Acepta si** (1280×540 con admin, y 844×390): el riel desplaza por dentro, el documento no, y el último módulo se alcanza.

#### S12 · P2 · Sin contención de overscroll ni reserva para la barra de scroll

**Dónde.** `styles.css:138-144`, `:310-316`.

**Síntoma.** En Chrome Android, en pestaña, al llegar al tope del panel y seguir tirando puede dispararse «tirar para recargar», y se pierde lo escrito. En iOS rebota toda la página. En Windows y Linux (1280), al pasar de esqueleto a lista aparece la barra de scroll y el contenido se corre ≈15 px.

**Corrección.**

```css
.panel-view {
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
}
```

Lo mismo para `.split-list` y `.split-detail` (ya incluido en S01), y `overscroll-behavior: none` en la raíz (S04). Donde el navegador no soporta `scrollbar-gutter` o `scrollbar-width`, no pasa nada.

**Acepta si:** en Android, tirar hacia abajo en el tope de la hoja no recarga; en escritorio el contenido no se corre al aparecer la barra.

#### S13 · P2 · Barra superior apretada entre 381 y 420 px

**Dónde.** `styles.css:424-439`, `:567-570`.

**Síntoma** (390). Isotipo de 36 px + 10 px + «VerdePUCP» a 18 px (≈88 px) + Plano/Relieve (≈144 px) + Salir (≈60 px) + tres huecos de 12 px suman ≈374 px en 350 px útiles: la marca se monta sobre «Plano». La regla compacta existe solo en ≤380 px. Es una estimación que se confirma en D.

**Corrección.** Reemplazar L567-570:

```css
.brand { overflow: hidden; }
@media (max-width: 420px) {
  .topbar { gap: 8px; padding-left: 10px; padding-right: 8px; }
  .word { font-size: 16px; }
  .roles button,
  .session button { padding-inline: 10px; }
}
@media (max-width: 359px) {
  .brand .word { display: none; }   /* queda el isotipo como identidad */
}
```

**Acepta si** (390 y 360): ningún elemento de la barra se superpone (detector de desbordes sobre `.topbar`).

#### S14 · P2 · En el teléfono, la labor elegida queda detrás de la hoja

**Dónde.** `CampusMap.tsx:453-460`.

**Síntoma** (390). `flyTo` centra la labor en y ≈ 422 px. Con la hoja al 55 %, todo lo que está por debajo de y ≈ 290 px queda tapado.

**Corrección.** Solo en este efecto; el resto del mapa no se toca.

```ts
function desfaseHoja(contenedor: HTMLElement): [number, number] {
  const hoja = parseFloat(getComputedStyle(contenedor).getPropertyValue("--sheet-h")) || 0
  if (hoja <= 0) return [0, 0]
  const alto = contenedor.clientHeight
  const arriba = document.querySelector(".topbar")?.getBoundingClientRect().bottom ?? 0
  const nav = document.querySelector(".guard")?.getBoundingClientRect().height ?? 0
  const bordeHoja = alto - nav - hoja
  return [0, Math.round((arriba + bordeHoja) / 2 - alto / 2)]
}

useEffect(() => {
  const map = mapRef.current
  if (!map || !ready || !focus) return
  const quiet = window.matchMedia("(prefers-reduced-motion: reduce)").matches
  const zoom = Math.max(map.getZoom(), 16.4)
  const offset = desfaseHoja(map.getContainer())
  if (quiet) map.easeTo({ center: [focus.lon, focus.lat], zoom, offset, duration: 0 })
  else map.flyTo({ center: [focus.lon, focus.lat], zoom, offset, duration: 900, essential: true })
}, [focus, ready])
```

`offset` es parte de `AnimationOptions` de MapLibre y no persiste. `padding` sí persiste en la transformación y movería `fitBounds`, el relieve y el dibujo: no usarlo. `jumpTo` no acepta `offset`, por eso el camino reducido usa `easeTo` con `duration: 0`. En escritorio `--sheet-h` vale 0 y el vuelo es idéntico al actual.

**Acepta si** (390): al elegir una labor, su pin queda entre la barra superior y el borde de la hoja. A 1280 el vuelo no cambia.

#### S15 · P2 · Teclado y gestos de la hoja

**Dónde.** `BottomSheet.tsx:53-92`, `:105-121`.

**Síntoma.**
- Escape solo cierra con el foco en el asa (`onKeyDown={tecla}` vive en el asa). Dentro de un campo no hace nada.
- Al cerrar, el foco queda en un elemento oculto (`display: none`) y el lector de pantalla se pierde.
- Si el navegador retira la captura del puntero sin `pointerup` ni `pointercancel` (llamada entrante, gesto del sistema), `arrastre` queda con valor y la hoja se traba a media altura, sin transición.
- Tocar el asa no hace nada, y un temblor de 1 px ya cuenta como arrastre.

**Corrección.** Versión objetivo en el paso A4.
- `onKeyDown` en el `aside`, solo para Escape y solo en el teléfono. Las flechas siguen solo en el asa: `ControlVertices` usa las flechas para mover vértices y no se deben capturar más arriba.
- App cierra y devuelve el foco al botón del módulo activo (`cerrarPanel`, paso A6).
- `onLostPointerCapture={alSoltar}`. Es idempotente: la segunda llamada ve `arrastreRef.current === null`.
- Un umbral de 4 px antes de tratar el gesto como arrastre. Si no hubo arrastre, el clic cicla 28 % → 55 % → 92 % → 55 %.

**Acepta si:** con teclado físico, Escape desde cualquier campo cierra la hoja y el foco vuelve al módulo; las flechas siguen moviendo el vértice en la edición de geometría; tocar el asa cambia de anclaje; no queda estado trabado tras cancelar un gesto.

#### S16 · P2 · `overflow: hidden` en contenedores que no deben desplazar

**Dónde.** `styles.css:128-137` (`.panel`), `:418-422` (`.shell` en móvil), `:454` (`.guard` en móvil).

**Síntoma.** Después de un `focus()` o un `scrollIntoView` dentro de la hoja, el navegador puede desplazar un ancestro con `overflow: hidden`, que sigue siendo contenedor de scroll por programa. La interfaz queda corrida unos píxeles sin forma de volver.

**Corrección.** `overflow: clip`, con `hidden` como respaldo:

```css
.shell { overflow: hidden; overflow: clip; }
.panel { overflow: hidden; overflow: clip; }
@media (max-width: 820px) {
  .guard { overflow: hidden; overflow: clip; }
}
```

En el bloque móvil, la regla `.shell, .shell.panel-off` (L418-422) declara su propio `overflow: hidden`: cambiarlo por el mismo par `hidden` + `clip`, o quitarlo para que herede la regla base.

`clip` recorta igual, esquinas redondeadas incluidas, pero no crea un contenedor de scroll. Los desplazamientos por programa pasan por `mostrarEnPanel`, que nunca toca el documento.

**Acepta si:** después de recorrer todos los formularios con Tab a 390 y a 1280, `[document.scrollingElement, ...document.querySelectorAll('#root, .shell, .panel')].every(e => e.scrollTop === 0 && e.scrollLeft === 0)`.

#### S17 · P2 · Acciones de Catastro al final de 521 filas

**Dónde.** `CatastroEditor.tsx:394-426`.

**Corrección.** Mover el `<p className="row-actions">` (Nueva área o Nueva zona, Exportar CSV) justo después del formulario de búsqueda (L322), antes de los avisos y de la lista. La lógica no cambia.

**Acepta si** (1280 y 390): **Nueva área** y **Exportar CSV** se ven sin desplazar la lista.

#### S18 · P3 · El panel móvil bloquea el zoom con pellizco

**Dónde.** `styles.css:529`.

**Corrección.** `.panel-view { touch-action: pan-y pinch-zoom; }`. El mapa conserva su `touch-action: none`. El hook de S06 ignora el teclado mientras hay zoom (`visualViewport.scale ≠ 1`).

**Acepta si** (teléfono real): pellizcar sobre el texto de la hoja amplía la página, y el arrastre vertical sigue desplazando.

#### S19 · P3 · Robustez responsive

- **Un solo corte en CSS y JS.** `App.tsx:114` usa `(min-width: 821px)` y `BottomSheet.tsx:4` usa `(max-width: 820px)`. Con anchos fraccionarios (820,5 px al hacer zoom) ninguna coincide y el panel nace oculto. Usar `MQ_MOVIL` de `ui/media.ts` y `!matches` para escritorio.
- **Respaldo de `dvh`:** `height: 100vh; height: 100dvh;` en `.shell` (L54), y `min-height: 100vh; min-height: 100dvh;` en `.gate` (L335).
- **`--sheet-h` tras rotar.** Hoy se calcula solo al cambiar el anclaje. La versión objetivo (A4) lo recalcula en `resize`.
- **Tableta vertical (600–820 px).** La hoja ocupa todo el ancho. Opcional:
  ```css
  @media (min-width: 600px) and (max-width: 820px) {
    .panel { left: 12px; right: auto; width: min(560px, calc(100% - 24px)); }
  }
  ```
- **«Saltar al panel»** (`App.tsx:535`) apunta a un panel oculto en el teléfono: `onClick={() => setRailOpen(true)}`.

### 2.4 Diseño visual

#### V01 · P1 · Pestañas sin estado activo visible

**Dónde.** `CatastroEditor.tsx:304-311` (Áreas verdes / Zonas), `CalendarioReservas.tsx:82-88` (Mes / Semana / Día), `styles.css:182`.

**Síntoma.** A 390 y a 1280 ninguna opción se ve elegida; solo cambia el contenido.

**Causa raíz.** Usan `role="tab"` con `aria-selected`, pero el CSS solo pinta `.roles button[aria-pressed="true"]`. Además un `tablist` exige `tabpanel` y navegación con flechas, que no existen.

**Corrección.** Son conmutadores de vista: `role="group"` en el contenedor y `aria-pressed` en cada botón, igual que Plano/Relieve e Inventario. Por robustez el CSS acepta ambos. Sacar `.roles button[aria-pressed="true"]` de la lista combinada de L182 y L187 y agregar:

```css
.roles button:is([aria-pressed="true"], [aria-selected="true"]) {
  background: var(--yucca);
  color: var(--yucca-ink);
}
.roles button:is([aria-pressed="true"], [aria-selected="true"]):hover {
  background: var(--yucca-deep);
}
```

**Acepta si:** la opción activa se ve en yucca en los cuatro segmentados.

#### V02 · P1 · Foco recortado y pestañas cortadas en los segmentados

**Dónde.** `styles.css:164-170` (`.roles { overflow: hidden }`), `:180-181`, `:318-319`; `InventarioCapas.tsx:221-227`.

**Síntoma.** Con Tab, el anillo de 2 px con 2 px de separación queda fuera de `.roles` y se recorta: en Plano/Relieve y en las pestañas casi no se ve (WCAG 2.4.7). Entre 821 y 1199 px el panel mide 280–340 px y las cuatro pestañas de Inventario piden ≈332 px: «reservas» queda cortada y no se puede tocar entera.

**Causa raíz.** Se usó `overflow: hidden` para redondear las esquinas, y los botones no tienen `min-width: 0`.

**Corrección.** Reemplazar L164-170 y L180-181:

```css
.roles {
  display: flex;
  background: var(--papel);
  border: 1px solid var(--linea);
  border-radius: var(--r-control);
}
.roles button {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 36px;
  border: 0;
  border-radius: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.roles button + button { border-left: 1px solid var(--linea); }
.roles button:first-child { border-radius: calc(var(--r-control) - 1px) 0 0 calc(var(--r-control) - 1px); }
.roles button:last-child { border-radius: 0 calc(var(--r-control) - 1px) calc(var(--r-control) - 1px) 0; }
.roles button:focus-visible { outline-offset: -4px; }
```

Y L319, que hoy fuerza anchos iguales con `flex: 1`:

```css
.catastro-editor .roles button { flex: 1 1 auto; padding-inline: 8px; font-size: 14px; }
```

Con 14 px y 8 px de margen interno, las cuatro pestañas piden ≈270 px y caben en los ≈282 px del panel a 1024 px. La elipsis queda solo como red de seguridad.

**Acepta si:** el anillo se ve completo en cada segmento al navegar con Tab, y a 1024 px las cuatro pestañas de Inventario se ven enteras, sin elipsis.

#### V03 · P2 · Botones con el estilo del navegador

**Dónde.** `CalendarioReservas.tsx:91-97` (Anterior, Hoy, Siguiente), `InventarioCapas.tsx:239` (registros), `CatastroEditor.tsx:325` (Abrir GeoJSON de prueba).

**Síntoma.** Botón gris con el borde del sistema, con altura y radio distintos al resto.

**Causa raíz.** El botón quieto se define con una lista de selectores de contexto (`.form button`, `.row-actions button`…). Lo que cae fuera queda sin estilo.

**Corrección.** Una base de especificidad cero para todo botón del panel. Cualquier clase existente la pisa:

```css
:where(.panel-view button) {
  min-height: 44px;
  padding: 8px 14px;
  border: 1px solid var(--linea);
  border-radius: var(--r-control);
  background: var(--hoja);
  color: var(--tinta);
  transition:
    background-color var(--dur-feedback) var(--ease),
    border-color var(--dur-feedback) var(--ease),
    transform var(--dur-feedback) var(--ease);
}
```

Los registros de Inventario pasan a filas de lista:

```tsx
<button type="button" className={id === fila.id ? "labor on" : "labor"} onClick={() => cargarFila(fila.raw)}>
  <span><strong>{fila.etiqueta}</strong></span>
</button>
```

No usar un selector global `button { … }`: alcanzaría los controles de MapLibre y el cerrar de los popups.

**Acepta si:** no queda ningún botón con aspecto del sistema en Mapa (calendario), Catastro (con error de API) e Inventario.

#### V04 · P2 · El panel no tiene cabeza fija

**Dónde.** `App.tsx:567-571`, `BottomSheet.tsx`.

**Síntoma.** La línea de estado («521 áreas · 534 zonas · 8 labores», «Sin conexión…», el área tocada en el mapa) es el primer párrafo del scroll. Al bajar se pierde, y en el teléfono tocar un polígono no da respuesta visible si la hoja está desplazada. DESIGN.md: «el panel es una hoja con cabeza fija».

**Corrección.** `BottomSheet` recibe `cabeza?: ReactNode` y la pinta entre `.sheet-chrome` y `.panel-view`. App le pasa el `<p className="status">` actual, con `role="status"`, y lo quita del cuerpo.

```css
.panel-head {
  flex: none;
  padding: 10px 16px;
  background: var(--papel);
  border-bottom: 1px solid var(--linea);
}
.panel-head .status { margin: 0; }
@media (max-width: 820px) {
  .panel-head { padding: 6px 16px; }
  .panel-head .status:not(.error) { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
}
```

**Acepta si:** a 390 y a 1280 la línea de estado sigue visible al desplazar cualquier módulo, y tocar un polígono con la hoja al 28 % muestra su nombre en la cabeza.

#### V05 · P2 · Elevación y radios fuera del sistema

**Dónde.** `styles.css:317` (sombra en `.split-detail`), `:480` y `:498` (hoja a 16 px), `:515` (asa en marino), `:635-638` (`.evidencia-archivo` sin radio), `:658-665` (`.chip` sin radio).

**Síntoma.** La ficha de Catastro flota con sombra dentro del panel; en el teléfono, es una sombra dentro de una hoja que ya tiene sombra. La hoja móvil usa 16 px cuando DESIGN.md fija 10 px para hojas flotantes. «Elegir archivo» y las etiquetas de estado son rectángulos de esquina viva junto a botones de 6 px. El asa en azul marino compite con el acento.

**Causa raíz.** Literales sueltos. La regla del escritorio plano (sombra solo sobre el mapa) no se aplicó a `.split-detail`.

**Corrección.**

```css
/* :root (A1): --r-control: 6px; --r-hoja: 10px; --r-marca: 4px; */
/* .split-detail: quitar box-shadow. S01 ya le da una línea de 1 px. */
@media (max-width: 820px) {
  .panel,
  .sheet-chrome { border-radius: var(--r-hoja) var(--r-hoja) 0 0; }
}
.sheet-grab { background: color-mix(in srgb, var(--tinta) 35%, transparent); }
.sheet-handle:active .sheet-grab,
.panel.arrastrando .sheet-grab { background: color-mix(in srgb, var(--tinta) 60%, transparent); }
.evidencia-archivo { border-radius: var(--r-control); padding: 8px 14px; }
.chip { border-radius: var(--r-marca); }
.chip.pendiente {
  background: color-mix(in srgb, var(--pendiente) 8%, var(--hoja));
  border-color: color-mix(in srgb, var(--pendiente) 40%, var(--linea));
}
.chip.enviada {
  background: color-mix(in srgb, var(--verde) 8%, var(--hoja));
  border-color: color-mix(in srgb, var(--verde) 40%, var(--linea));
}
```

Reemplazar los `6px`, `10px` y `4px` literales por los tokens donde aparecen (`.guard button`, botones, `.field`, `.banner`, `.counts li`, `.cal-*`, `.topbar` móvil, popups). Elevación permitida: barra superior móvil, hoja móvil, popups, menú «Más» y controles del mapa. Nada más.

**Acepta si:** a 390 y a 1280 no hay sombras dentro del panel, y los radios son 6, 10 y 4 según DESIGN.md.

#### V06 · P2 · `fieldset` sin normalizar

**Dónde.** `Labores.tsx:344-370` (Ficha de la labor), `CatastroEditor.tsx:544` y `:636` (`.geom-box`), `styles.css` (sin reset).

**Síntoma.** La ficha de labor se dibuja con el borde «groove» de 2 px del navegador, con su margen y su leyenda: distinta a todo lo demás. Por omisión `fieldset` tiene `min-inline-size: min-content`, que impide encogerlo y puede desbordar en anchos chicos.

**Corrección.**

```css
fieldset { min-inline-size: 0; margin: 0; padding: 0; border: 0; }
.grupo,
.geom-box,
.vertice-box {
  margin: 0 0 16px;
  padding: 8px 12px 4px;
  border: 1px solid var(--linea);
  border-radius: var(--r-control);
  background: transparent;
}
.grupo > legend,
.geom-box > legend { padding: 0 4px; font-size: 13px; font-weight: 600; }
```

En `FichaLabor`: `<fieldset className="form grupo">`. El grupo no lleva fondo propio: se marca con la línea, no con otra tarjeta.

**Acepta si:** la ficha de labor y la geometría de Catastro se ven como el mismo grupo, de 1 px y 6 px.

#### V07 · P2 · Vocabulario de interacción incompleto

**Dónde.** `styles.css:109-126`, `:171-191`, `:264-277`.

**Síntoma.** Los botones quietos, los de peligro y las filas de lista no tienen hover. `button:active { transform: scale(0.98) }` encoge también las filas, las casillas del calendario y los botones de navegación: la fila entera late. En Android e iOS aparece el rectángulo gris de toque, sin radio.

**Corrección.** Reemplazar L190:

```css
button, a, label, summary { -webkit-tap-highlight-color: transparent; }
button:not(.labor, .cal-dia, .sheet-handle, .link, .guard *, .roles *):active,
.row-actions a:active { transform: scale(0.98); }
.labor:active,
.guard button:active { background: color-mix(in srgb, var(--yucca) 12%, var(--hoja)); }
@media (hover: hover) {
  :where(.panel-view button:not(.primary, .danger, .labor, .link, .cal-dia, [aria-pressed="true"]):not(:disabled)):hover,
  .session button:hover {
    background: color-mix(in srgb, var(--yucca) 6%, var(--hoja));
    border-color: var(--linea-fuerte);
  }
  .danger:not(:disabled):hover { background: color-mix(in srgb, var(--alerta) 6%, var(--hoja)); }
  .labor:not(.on):hover { background: color-mix(in srgb, var(--yucca) 5%, var(--hoja)); }
  .cal-dia:hover { border-color: var(--linea-fuerte); }
}
```

`--linea-fuerte` se define en `:root` (A1).

**Acepta si:** a 1280 todo botón y toda fila responden al hover; a 390, tocar una fila la tiñe sin encogerla y no aparece el rectángulo gris.

#### V08 · P2 · Secciones sin ritmo

**Dónde.** `App.tsx:572-609` (Capas, Inventario, Reservas), `:610-699` (Labores, Riego, Poda y Vivero); `styles.css` (sin reglas para `.block`).

**Síntoma.** Cuatro secciones con `h2` de 22 px, una debajo de otra, sin aire ni corte: el título de Riego queda pegado a la bitácora de la labor.

**Corrección.**

```css
.block + .block,
.block > .calendario {
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid var(--linea);
}
.panel-view h2 { margin: 0 0 4px; }
.panel-view h3 { margin: 24px 0 8px; }
.detail { margin-top: 24px; padding-top: 16px; }
```

**Acepta si:** a 390 y a 1280, con la vista entrecerrada, se distinguen Labores, Riego, Poda y Vivero como grupos.

#### V09 · P2 · Vacío falso mientras carga

**Dónde.** `Labores.tsx:203`, `Modulos.tsx:287` (Solicitudes), `:596` (Catálogos) y `:738` (Riego), `Poda.tsx:31`, `Vivero.tsx:48`, `EvidenciasCampo.tsx:221-223`, `InventarioCapas.tsx:237` (sin nada).

**Síntoma.** Con red lenta se lee «No hay labores con este filtro.», «No hay solicitudes registradas.» o «Esta labor no tiene fotos.» antes de que llegue el dato. DESIGN.md pide tres filas de esqueleto.

**Corrección.** Componente `src/ui/Esqueleto.tsx`:

```tsx
export function Esqueleto({ filas = 3 }: { filas?: number }) {
  return (
    <div className="skel-wrap" aria-hidden="true">
      {Array.from({ length: filas }, (_, i) => <div className="skel" key={i} />)}
    </div>
  )
}
```

Estado de carga sin `setState` síncrono dentro de efectos:
- **Solicitudes, Riego, Poda y Vivero:** `const [cargando, setCargando] = useState(true)` (`useState(!props.iniciales)` en Poda y Vivero) y `setCargando(false)` en el `then` y en el `catch` del efecto inicial.
- **Catálogos:** `const [cargadoDe, setCargadoDe] = useState<string | null>(null)`, `const cargando = cargadoDe !== clase`, y `setCargadoDe(clase)` en `then` y `catch`.
- **Registros de Inventario:** igual, con `entidad`.
- **Evidencias:** igual, con `actividadId`, dentro de `refrescar`.
- **Labores:** App agrega `const [laboresListas, setLaboresListas] = useState(false)`, lo pone en `true` en el `then` y en las dos ramas del `catch` del efecto de L269-291, y pasa `cargando={!laboresListas}` a `Labores` como prop opcional. `Labores` muestra `<Esqueleto />` en lugar del `<ul>`.

Pintar `{cargando ? <Esqueleto /> : filas.length === 0 && !error && <p className="empty">…</p>}`, con el esqueleto fuera de `<ul>`. En CSS: `.panel-view .skel-wrap { padding: 4px 0; }`.

**Acepta si** (DevTools con red «Slow 3G», a 390 y a 1280): ningún «No hay…» aparece antes de que responda la API.

#### V10 · P2 · Errores sin jerarquía y campos inválidos sin marca

**Dónde.** `styles.css:208-216`, `:332`; `aria-invalid` en `CatastroEditor.tsx` y en `FechaCampo.tsx:14`.

**Síntoma.** Un error al guardar es una línea roja de 13 px entre campos. Una fecha mal escrita no cambia de aspecto.

**Corrección.**

```css
.status.error {
  color: var(--alerta);
  background: color-mix(in srgb, var(--alerta) 7%, var(--hoja));
  border: 1px solid color-mix(in srgb, var(--alerta) 35%, var(--linea));
  border-radius: var(--r-control);
  padding: 8px 10px;
}
.errores-campo .status.error { background: none; border: 0; padding: 0; }
.banner { border: 1px solid color-mix(in srgb, var(--yucca) 30%, var(--linea)); }
.field :is(input, select, textarea)[aria-invalid="true"] {
  border-color: var(--alerta);
  background: color-mix(in srgb, var(--alerta) 4%, var(--hoja));
}
```

Agregar `role="status"` al aviso de Labores (`Labores.tsx:201`).

**Acepta si:** un error de la API se ve como aviso con tono, y escribir «31/02/2026» en una fecha marca el campo en alerta.

#### V11 · P2 · `input type="date"` donde DESIGN.md exige `dd/mm/aaaa`

**Dónde.** `Labores.tsx:352` y `:356` (Ficha de la labor), `Poda.tsx:102` y `:106`, `Vivero.tsx:98`.

**Síntoma.** Según el navegador se ve `09/24/2026` (mm/dd/aaaa), que PRODUCT.md cita como rechazo explícito.

**Corrección.** Usar `FechaCampo`: el valor que viaja a la API sigue siendo `AAAA-MM-DD`. Antes hay que hacer que `FechaCampo` acepte cambios externos. Hoy guarda el texto solo al montarse, así que en Poda elegir otra poda no actualizaría la fecha y en Vivero el reinicio tras guardar dejaría el texto viejo.

```tsx
export function FechaCampo(props: { value: string; onChange: (iso: string) => void; required?: boolean }) {
  const [text, setText] = useState(() => (props.value ? formatFecha(props.value) : ""))
  const [bad, setBad] = useState(false)
  const [externo, setExterno] = useState(props.value)
  if (props.value !== externo) {
    setExterno(props.value)
    // Solo si el cambio vino de fuera: no reformatear mientras se escribe.
    if (props.value !== (parseFechaPE(text) ?? "")) {
      setText(props.value ? formatFecha(props.value) : "")
      setBad(false)
    }
  }
  // … el <input> actual, sin cambios
}
```

Reemplazos: `<FechaCampo value={solicitud} onChange={setSolicitud} />`, lo mismo para `atencion`; `(iso) => set("fecha_reporte", iso)` y `(iso) => set("fecha_ejecucion", iso)` en Poda; `(iso) => set("fecha", iso)` en Vivero. El `type="month"` de Vivero se queda.

**Acepta si:** las cinco fechas muestran `dd/mm/aaaa`; en la pestaña Network el cuerpo sigue llevando `AAAA-MM-DD`; elegir otra poda actualiza sus fechas; `atencion.test.ts` sigue en verde.

#### V12 · P2 · Objetivos táctiles menores a 44 px

**Dónde.** `styles.css:229-231` (`.checks label`, 32 px), `:253-263` (`.layer label`, ≈38 px).

**Corrección.**

```css
@media (pointer: coarse) {
  .checks label { min-height: 44px; }
  .layer label { display: block; min-height: 44px; }
}
```

**Acepta si:** el detector de objetivos (D.3) no lista filtros de estado ni capas.

#### V13 · P3 · Inventario: rótulos, copia técnica y jerarquía

**Dónde.** `InventarioCapas.tsx:221-227`, `:233`, `:254`.

**Síntoma.** Las pestañas dicen «tachos», «bebederos»…, en minúscula, porque son ids. Bajo cada capa se listan claves técnicas (`area_m2, perimetro_m`). La lista no tiene título y el formulario sí («Edición de inventario»), al revés que en Catastro.

**Corrección.**

```tsx
const PESTANAS = { tachos: "Tachos", bebederos: "Bebederos", puntos: "Puntos", reservas: "Reservas" } as const
const CAMPO = {
  nombre: "nombre", codigo: "código", nota: "nota", clase: "clase", riego: "riego",
  area_m2: "área", perimetro_m: "perímetro", pertenecen: "pertenecen", uso: "uso",
} as const
```

- `<div className="roles" role="group" aria-label="Inventario">` y como texto `{PESTANAS[item]}`.
- `<small>Campos: {item.campos.map((campo) => CAMPO[campo]).join(", ")}</small>`.
- `<h2>Inventario</h2>` al inicio de `.split-list`. El título del formulario baja a `<h3>{id == null ? "Nuevo registro" : "Editar registro"}</h3>`.

**Acepta si:** las pestañas llevan mayúscula inicial y no se cortan, y no queda ninguna clave técnica en pantalla.

#### V14 · P3 · Acabado

- `.word { font-weight: 600; }`. `main.tsx` no carga el peso 700 y, con `font-synthesis: none`, se ve 600 de todos modos: declarar lo que se ve.
- Listas con línea de cierre: `.labor-list > li:last-child > .labor, .labor-list > li.agenda:last-child, .timeline li:last-child { border-bottom: 1px solid var(--linea); }`. `.evidencia-lista li` pasa a `border-top`, como el resto.
- Popups del mapa (solo estilo):
  ```css
  .cv-popup .maplibregl-popup-content { overflow-wrap: anywhere; padding-right: 32px; }
  .cv-popup .maplibregl-popup-close-button {
    width: 32px;
    height: 32px;
    font-size: 20px;
    border-radius: 0 var(--r-hoja) 0 var(--r-control);
  }
  @media (pointer: coarse) {
    .cv-popup .maplibregl-popup-close-button { width: 44px; height: 44px; }
  }
  ```
- Usar la escala de z-index de `:root` (A1) en `.skip`, `.topbar`, `.guard`, `.panel` y `.guard-menu`.
- Retirar `.menu-btn` (`styles.css:163`, `:438`; `App.tsx:544-546`): nunca se muestra.

### 2.5 Movimiento

**Tesis.** El único momento con autoría es la hoja que sube desde la barra de módulos: continuidad entre «toqué Labores» y «Labores está aquí». Todo lo demás es respuesta de estado de 120–180 ms. La única propiedad de layout que se anima es el alto de la hoja; el resto es `transform` y `opacity`.

#### M01 · P2 · La hoja parpadea al abrir, se va de golpe y salta al anclaje

**Dónde.** `styles.css:483` (`animation: sheet 180ms`), `:572-575` (`@keyframes sheet`: 12 px y opacidad desde 0,6), `:61` (`display: none` al cerrar), `BottomSheet.tsx:101`.

**Síntoma** (390). Al abrir, la hoja ya está casi en su sitio y late 12 px con opacidad. Al cerrar desaparece en un cuadro. Al soltar el asa salta al anclaje, sin transición.

**Corrección.** Entrada desde la barra, salida más corta y ajuste de alto con la misma curva, sin JS de animación. Quitar L483 y L572-575, y agregar al bloque móvil:

```css
.panel.sheet {
  transition:
    height var(--dur-hoja) var(--ease-llegada),
    transform var(--dur-hoja) var(--ease-llegada),
    display var(--dur-hoja) allow-discrete;
}
@starting-style {
  .panel.sheet { transform: translateY(calc(100% + var(--nav-h) + env(safe-area-inset-bottom))); }
}
.shell.panel-off .panel.sheet {
  transform: translateY(calc(100% + var(--nav-h) + env(safe-area-inset-bottom)));
  transition-duration: var(--dur-salida);
  transition-timing-function: var(--ease-salida);
}
.panel.sheet.arrastrando,
.panel.sheet.con-teclado { transition: none; }
.panel.sheet.arrastrando { will-change: height; }
.panel.sheet.arrastrando .sheet-handle { cursor: grabbing; }
```

`BottomSheet` conserva el alto del último anclaje también cerrada, para que la salida no cambie de tamaño, y marca el arrastre con la clase `arrastrando` en lugar de `style.transition`. Donde no hay soporte de `@starting-style` o de `allow-discrete` (Safari anterior a 17.5), la hoja aparece y desaparece sin animación: aceptable.

**Acepta si** (390): abrir dura ≈260 ms deslizando desde la barra; cerrar, ≈180 ms; soltar el asa anima el ajuste; durante el arrastre no hay retardo.

#### M02 · P2 · Movimiento reducido demasiado brusco

**Dónde.** `styles.css:617-622`, `BottomSheet.tsx:10-12` y `:101`.

**Síntoma.** Con `prefers-reduced-motion`, la regla apaga toda transición de los botones, color incluido, contra DESIGN.md («anula el desplazamiento y deja el cambio de color»). La hoja no tiene alternativa.

**Corrección.** Reemplazar L617-622 y dejar el bloque al final del archivo:

```css
@media (prefers-reduced-motion: reduce) {
  .panel-view,
  .guard-menu { animation-name: aparecer; }
  .skel { animation: none; }
  button:active,
  .row-actions a:active { transform: none; }
  .maplibregl-ctrl-bottom-left,
  .maplibregl-ctrl-bottom-right { transition: none; }
}
@media (prefers-reduced-motion: reduce) and (max-width: 820px) {
  .panel.sheet { transition: opacity 120ms linear, display 120ms allow-discrete; }
  @starting-style { .panel.sheet { opacity: 0; transform: none; } }
  .shell.panel-off .panel.sheet { opacity: 0; transform: none; }
}
@keyframes aparecer { from { opacity: 0; } }
```

Quitar `reducido()` de `BottomSheet`: lo resuelve el CSS. El mapa ya usa `jumpTo` o `duration: 0`, y `mostrarEnPanel` usa `behavior: "auto"`.

**Acepta si** (emulación en DevTools): ningún elemento se traslada; la hoja aparece y se va con un fundido de 120 ms; los botones siguen cambiando de color.

#### M03 · P3 · Transiciones de estado ausentes

**Dónde.** `styles.css:143` (`rise` solo al montar el panel), `:264-277`, `:361-369`.

**Síntoma.** Cambiar de módulo no anima. Elegir una fila, ver un aviso o abrir la ficha es instantáneo. El esqueleto brilla con un degradado fuerte y una curva de estado en bucle.

**Corrección.**
- Con `key={vista}` (S09), el `rise` de `.panel-view` se repite en cada módulo: 180 ms y 6 px.
- `.labor { transition: background-color var(--dur-feedback) var(--ease); }`. La transición de `.roles button` agrega `color`.
- `.panel-view :is(.banner, .status.error), .detail { animation: aparecer var(--dur-estado) var(--ease); }`. Solo al montarse; sin `key` en `.detail`.
- Esqueleto con pulso de opacidad, sin degradado. Retirar `@keyframes shimmer` (L401-404):
  ```css
  .skel {
    height: 52px;
    margin: 8px 0;
    border-radius: var(--r-control);
    background: color-mix(in srgb, var(--linea) 55%, var(--hoja));
    animation: pulso 1.4s ease-in-out infinite alternate;
  }
  @keyframes pulso { to { opacity: 0.55; } }
  ```
- Menú «Más»: `animation: menu-entra var(--dur-estado) var(--ease-llegada)` con `@keyframes menu-entra { from { opacity: 0; transform: translateY(6px); } }`. Se cierra sin animación: la salida es más rápida que la entrada.

#### M04 · P3 · Controles del mapa y costo del arrastre

**Dónde.** `styles.css:537-540`, `BottomSheet.tsx:41-51` y `:64`.

**Síntoma.** Los controles saltan a su sitio mientras la hoja anima. Cada movimiento del dedo cambia `--sheet-h` en `:root` y obliga a recalcular los estilos de todo el documento; con Catastro abierto son más de 2 500 nodos.

**Corrección.** App publica el alto en `.stage` con `onAltura`, no en `document.documentElement`, y los controles lo heredan:

```css
@media (max-width: 820px) {
  .maplibregl-ctrl-bottom-left,
  .maplibregl-ctrl-bottom-right {
    bottom: calc(var(--sheet-h, 0px) + var(--nav-h) + env(safe-area-inset-bottom) + 12px) !important;
    transition: bottom var(--dur-hoja) var(--ease-llegada);
  }
  .shell:has(.panel.arrastrando) :is(.maplibregl-ctrl-bottom-left, .maplibregl-ctrl-bottom-right) {
    transition: none;
  }
  .shell.panel-off :is(.maplibregl-ctrl-bottom-left, .maplibregl-ctrl-bottom-right) {
    transition-duration: var(--dur-salida);
  }
}
```

Esto reemplaza L537-540. La regla de L541-543 (`top-right`) no aplica a ningún control y se puede quitar.

**Acepta si:** los controles acompañan a la hoja al abrir, cerrar y ajustar; en el panel Performance de Chrome, el arrastre no muestra «Recalculate Style» de todo el documento en cada cuadro.

### 2.6 Observaciones fuera de alcance

- **O1.** `FichaLabor` (`Labores.tsx:305-372`) conserva su estado local al cambiar de labor: lo escrito para la labor A puede guardarse en la B. Es comportamiento de datos. No agregar `key` al `.detail` sin una decisión de producto.
- **O2.** `map/draw.ts` edita vértices con eventos de mouse. En el tacto, el camino son los botones Oeste/Este/Norte/Sur del panel, así que `ControlVertices` debe quedar alcanzable (S01, S02) y la hoja no debe capturar las flechas.
- **O3.** `CatastroPanel` (`Modulos.tsx:85-240`) no se usa. No invertir en él; retirarlo en una limpieza aparte.
- **O4.** La prosa de DESIGN.md cita hex antiguos (`#1a5c44`…) que no coinciden con el frontmatter ni con `styles.css`. Implementar siempre con las variables CSS.
- **O5.** El `mousemove` del mapa proyecta todos los puntos de inventario encendidos en cada movimiento (`CampusMap.tsx:308-316`). No es de este plan.

---

## 3. Plan de implementación

### Principios

1. Un scroll por región, según 2.2. No crear contenedores con scroll fuera de esa tabla.
2. Todo hijo flex o grid que deba encogerse lleva `min-height: 0`, `min-width: 0` o `minmax(0, …)`.
3. `overflow: hidden` no arregla desbordes. En contenedores sin scroll se usa `clip`, y el contenido ancho va en su propio `.tabla-scroll`.
4. Los desplazamientos por programa pasan solo por `mostrarEnPanel`: nada de `scrollIntoView` ni `window.scrollTo`. Al enfocar por código, `focus({ preventScroll: true })`.
5. Un solo corte en CSS y JS: `MQ_MOVIL`, 820 px.
6. Movimiento: `transform` y `opacity`. La única propiedad de layout animada es el alto de la hoja. De 120 a 260 ms, con salidas más cortas que las entradas y sin rebote.
7. Con `prefers-reduced-motion` no hay traslación, pero el color y la opacidad se quedan.
8. Manda DESIGN.md: un acento, sombra solo sobre el mapa, radios de 6, 10 y 4 px. Sin gradientes decorativos, sin vidrio ni `backdrop-filter`, sin emojis, sin píldoras, sin tarjetas anidadas, sin fuentes ni dependencias nuevas.
9. La copia no cambia, salvo lo listado en V13 y los botones nuevos «Más» y «Volver a la lista».
10. Antes de editar UI, leer `/home/box/agent-data/workflows/impeccable/reference/craft-floor.md`.
11. Un commit por paso y un PR por lote. Después de cada paso: `npm run lint && npm test && npm run build` en `apps/web`.

### Lote A · Scroll y layout

**Archivos:** `src/styles.css`, `src/App.tsx`, `src/ui/BottomSheet.tsx`, nuevos `src/ui/media.ts`, `src/ui/desplazar.ts`, `src/ui/teclado.ts`, `src/ui/navegacion.ts` con sus tests, `src/panel/CatastroEditor.tsx`, `src/panel/InventarioCapas.tsx`, `src/panel/Labores.tsx`, `src/panel/Importaciones.tsx`, `src/map/CampusMap.tsx` (solo el efecto `focus`) y `package.json` (lista de tests).

**A0 · Preparación.** Rama `fix/ui-scroll-diseno`. Línea base: `cd apps/web && npm run lint && npm test && npm run build`. Capturas de cada módulo a 390×844 y a 1280×800 (coordinación y capataz) para comparar.

**A1 · Tokens y utilidades.** Sin cambio visible.

En `:root` de `styles.css`, junto a los tokens actuales:

```css
--r-control: 6px;
--r-hoja: 10px;
--r-marca: 4px;
--linea-fuerte: color-mix(in srgb, var(--tinta) 24%, var(--linea));
--nav-h: 64px;
--topbar-zona: calc(env(safe-area-inset-top) + 64px);
--sheet-util: calc(100dvh - var(--nav-h) - env(safe-area-inset-bottom) - var(--topbar-zona));
--dur-feedback: 120ms;
--dur-estado: 180ms;
--dur-hoja: 260ms;
--dur-salida: 180ms;
--ease-llegada: cubic-bezier(0.16, 1, 0.3, 1);
--ease-salida: cubic-bezier(0.4, 0, 1, 1);
--z-mapa-ui: 2;
--z-topbar: 3;
--z-hoja: 5;
--z-nav: 6;
--z-menu: 7;
--z-skip: 8;
```

`--ease` sigue siendo la curva de estado.

`src/ui/media.ts`:

```ts
import { useCallback, useSyncExternalStore } from "react"

/** Único corte entre teléfono y escritorio. styles.css usa el mismo valor. */
export const MQ_MOVIL = "(max-width: 820px)"

export function useMedia(query: string): boolean {
  const suscribir = useCallback(
    (avisar: () => void) => {
      const lista = window.matchMedia(query)
      lista.addEventListener("change", avisar)
      return () => lista.removeEventListener("change", avisar)
    },
    [query],
  )
  return useSyncExternalStore(suscribir, () => window.matchMedia(query).matches, () => false)
}
```

`src/ui/desplazar.ts`:

```ts
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
```

`src/ui/teclado.ts`:

```ts
import { useEffect, useState } from "react"

export type Teclado = { abierto: boolean; alto: number; top: number }
const CERRADO: Teclado = { abierto: false, alto: 0, top: 0 }

/** Parte baja del layout viewport que tapa el teclado, en px. */
export function altoTapado(innerHeight: number, vv: { height: number; offsetTop: number }): number {
  return Math.max(0, Math.round(innerHeight - vv.height - vv.offsetTop))
}

export function tecladoAbierto(tapado: number, escala: number, editable: boolean): boolean {
  return editable && tapado > 120 && Math.abs(escala - 1) < 0.01
}

function esEditable(el: Element | null): boolean {
  if (el instanceof HTMLTextAreaElement || el instanceof HTMLSelectElement) return true
  if (el instanceof HTMLInputElement) {
    return !["checkbox", "radio", "button", "submit", "reset", "file", "range", "color"].includes(el.type)
  }
  return el instanceof HTMLElement && el.isContentEditable
}

/** Estado del teclado del teléfono a partir del visual viewport. */
export function useTeclado(activo: boolean): Teclado {
  const [estado, setEstado] = useState<Teclado>(CERRADO)
  useEffect(() => {
    const vv = window.visualViewport
    if (!activo || !vv) return
    let cuadro = 0
    const medir = () => {
      cancelAnimationFrame(cuadro)
      cuadro = requestAnimationFrame(() => {
        const abierto = tecladoAbierto(altoTapado(window.innerHeight, vv), vv.scale, esEditable(document.activeElement))
        const siguiente = abierto ? { abierto, alto: Math.round(vv.height), top: Math.round(vv.offsetTop) } : CERRADO
        setEstado((prev) =>
          prev.abierto === siguiente.abierto && prev.alto === siguiente.alto && prev.top === siguiente.top ? prev : siguiente,
        )
      })
    }
    vv.addEventListener("resize", medir)
    vv.addEventListener("scroll", medir)
    document.addEventListener("focusin", medir)
    document.addEventListener("focusout", medir)
    medir()
    return () => {
      cancelAnimationFrame(cuadro)
      vv.removeEventListener("resize", medir)
      vv.removeEventListener("scroll", medir)
      document.removeEventListener("focusin", medir)
      document.removeEventListener("focusout", medir)
    }
  }, [activo])
  return activo ? estado : CERRADO
}
```

`src/ui/navegacion.ts`:

```ts
/** En el teléfono caben cinco casillas: cuatro módulos y «Más» con el resto. */
export function repartirModulos<T>(modulos: readonly T[], casillas = 5): { barra: T[]; mas: T[] } {
  if (modulos.length <= casillas) return { barra: [...modulos], mas: [] }
  return { barra: modulos.slice(0, casillas - 1), mas: modulos.slice(casillas - 1) }
}
```

Tests nuevos, con imports `.ts` como los existentes, agregados a la lista explícita del script `test` de `package.json`:

```ts
// src/ui/navegacion.test.ts
import assert from "node:assert/strict"
import test from "node:test"
import { repartirModulos } from "./navegacion.ts"

test("la barra del teléfono deja cuatro módulos y manda el resto a Más", () => {
  assert.deepEqual(repartirModulos(["a", "b", "c", "d", "e"]), { barra: ["a", "b", "c", "d", "e"], mas: [] })
  const nueve = repartirModulos(["1", "2", "3", "4", "5", "6", "7", "8", "9"])
  assert.deepEqual(nueve.barra, ["1", "2", "3", "4"])
  assert.deepEqual(nueve.mas, ["5", "6", "7", "8", "9"])
})
```

```ts
// src/ui/teclado.test.ts
import assert from "node:assert/strict"
import test from "node:test"
import { altoTapado, tecladoAbierto } from "./teclado.ts"

test("el teclado cuenta si tapa más de 120 px, sin zoom y con un campo enfocado", () => {
  assert.equal(altoTapado(844, { height: 500, offsetTop: 0 }), 344)
  assert.equal(altoTapado(844, { height: 844, offsetTop: 0 }), 0)
  assert.equal(tecladoAbierto(344, 1, true), true)
  assert.equal(tecladoAbierto(344, 1, false), false)
  assert.equal(tecladoAbierto(80, 1, true), false)
  assert.equal(tecladoAbierto(344, 1.6, true), false)
})
```

**A2 · Estructura del panel** (S01, S02, S09, S17). App sin el envoltorio y con `vista`; BottomSheet con `key={vista}`; CSS de S01 y S02; CatastroEditor (maestro-detalle, refs, «Volver a la lista», acciones arriba); InventarioCapas (refs, `salto`, «Volver a la lista»).

**A3 · Higiene de scroll** (S04, S05, S10 en su parte CSS, S12, S16, S18, S19 `dvh`). Solo `styles.css`. En el bloque móvil, `.panel-view` queda `touch-action: pan-y pinch-zoom; padding-bottom: 20px;`.

**A4 · Hoja móvil** (S03, S15, S19 recálculo). CSS de S03. `BottomSheet.tsx`, versión objetivo:

```tsx
import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type PointerEvent as ReactPointerEvent,
  type ReactNode,
} from "react"
import { ANCLAJES, alturaPx, anclajeCercano, siguienteAnclaje, type Anclaje } from "./bottomSheet"
import { mostrarEnPanel } from "./desplazar"
import { MQ_MOVIL, useMedia } from "./media"
import { useTeclado } from "./teclado"

const MEDIA: Anclaje = 0.55
const UMBRAL = 4

type Props = {
  open: boolean
  onClose: () => void
  /** Módulo activo: al cambiar, el cuerpo vuelve arriba. */
  vista: string
  /** Línea fija sobre el cuerpo (Lote B). */
  cabeza?: ReactNode
  /** Alto que tapa la hoja en el teléfono, para los controles y el encuadre del mapa. */
  onAltura?: (px: number) => void
  children: ReactNode
}

export function BottomSheet({ open, onClose, vista, cabeza, onAltura, children }: Props) {
  const movil = useMedia(MQ_MOVIL)
  const hoja = useRef<HTMLElement>(null)
  const [anclaje, setAnclaje] = useState<Anclaje>(MEDIA)
  const [arrastre, setArrastre] = useState<number | null>(null)
  const [vistoAbierto, setVistoAbierto] = useState(open)
  const inicio = useRef({ y: 0, alto: 0, tope: 0 })
  const arrastreRef = useRef<number | null>(null)
  const movio = useRef(false)
  const teclado = useTeclado(movil && open)
  if (open !== vistoAbierto) {
    setVistoAbierto(open)
    if (open) setAnclaje(MEDIA)
  }

  function tope(): number {
    const px = hoja.current ? parseFloat(getComputedStyle(hoja.current).maxHeight) : NaN
    return Number.isFinite(px) && px > 0 ? px : window.innerHeight * 0.8
  }

  useEffect(() => {
    if (!onAltura) return
    if (!movil || !open || teclado.abierto) {
      onAltura(0)
      return
    }
    const publicar = () => onAltura(arrastre ?? alturaPx(anclaje, tope()))
    publicar()
    window.addEventListener("resize", publicar)
    return () => window.removeEventListener("resize", publicar)
  }, [movil, open, anclaje, arrastre, teclado.abierto, onAltura])

  useEffect(() => {
    const raiz = document.documentElement
    raiz.toggleAttribute("data-teclado", teclado.abierto)
    if (teclado.abierto) requestAnimationFrame(() => mostrarEnPanel(document.activeElement, "cerca"))
    return () => raiz.removeAttribute("data-teclado")
  }, [teclado.abierto, teclado.alto])

  function alBajar(event: ReactPointerEvent<HTMLElement>) {
    if (!movil) return
    inicio.current = { y: event.clientY, alto: hoja.current?.getBoundingClientRect().height ?? 0, tope: tope() }
    movio.current = false
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  function alMover(event: ReactPointerEvent<HTMLElement>) {
    if (!movil || !event.currentTarget.hasPointerCapture(event.pointerId)) return
    const delta = inicio.current.y - event.clientY
    if (!movio.current && Math.abs(delta) < UMBRAL) return
    movio.current = true
    const alto = Math.min(inicio.current.tope, Math.max(48, inicio.current.alto + delta))
    arrastreRef.current = alto
    setArrastre(alto)
  }

  function alSoltar(event: ReactPointerEvent<HTMLElement>) {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    const alto = arrastreRef.current
    if (!movil || alto == null) return
    arrastreRef.current = null
    setArrastre(null)
    const destino = anclajeCercano(alto / inicio.current.tope)
    if (destino == null) onClose()
    else setAnclaje(destino)
  }

  function alTocarAsa() {
    if (movio.current) {
      movio.current = false
      return
    }
    const ultimo = ANCLAJES[ANCLAJES.length - 1]
    setAnclaje((actual) => (actual === ultimo ? MEDIA : (siguienteAnclaje(actual, 1) ?? MEDIA)))
  }

  function flechas(event: ReactKeyboardEvent) {
    if (!movil || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) return
    event.preventDefault()
    const siguiente = siguienteAnclaje(anclaje, event.key === "ArrowUp" ? 1 : -1)
    if (siguiente == null) onClose()
    else setAnclaje(siguiente)
  }

  function escape(event: ReactKeyboardEvent) {
    if (!movil || event.key !== "Escape" || event.defaultPrevented) return
    event.stopPropagation()
    onClose()
  }

  const estilo = !movil
    ? undefined
    : teclado.abierto
      ? { top: `${teclado.top}px`, bottom: "auto", height: `${teclado.alto}px` }
      : { height: arrastre != null ? `${arrastre}px` : `calc(var(--sheet-util) * ${anclaje})` }
  const clases = ["panel", movil && "sheet", arrastre != null && "arrastrando", teclado.abierto && "con-teclado"]
    .filter(Boolean)
    .join(" ")
  const indice = ANCLAJES.indexOf(anclaje)

  return (
    <aside ref={hoja} className={clases} id="panel" style={estilo} aria-label="Panel" onKeyDown={escape}>
      <div className="sheet-chrome">
        <button
          type="button"
          className="sheet-handle"
          role="slider"
          aria-label="Arrastrar el panel. Flechas para subir o bajar."
          aria-valuemin={0}
          aria-valuemax={ANCLAJES.length - 1}
          aria-valuenow={Math.max(indice, 0)}
          aria-valuetext={indice === 0 ? "Vista mínima" : indice === 1 ? "Media altura" : "Casi pantalla completa"}
          onKeyDown={flechas}
          onClick={alTocarAsa}
          onPointerDown={alBajar}
          onPointerMove={alMover}
          onPointerUp={alSoltar}
          onPointerCancel={alSoltar}
          onLostPointerCapture={alSoltar}
        >
          <span className="sheet-grab" aria-hidden="true" />
        </button>
        <button type="button" className="sheet-close" onClick={onClose}>
          Cerrar
        </button>
      </div>
      {cabeza && <div className="panel-head">{cabeza}</div>}
      <div className="panel-view" key={vista}>
        {children}
      </div>
    </aside>
  )
}
```

La hoja ya no escribe `--sheet-h` en `document.documentElement`. En el mismo paso, App lo publica en `.stage`; si no, los controles del mapa quedarían tapados por la hoja. Hay que agregar `useRef` al import de React en `App.tsx`, y los hooks van **antes de los `return` tempranos (L512)**:

```tsx
const stageRef = useRef<HTMLDivElement>(null)
const publicarAltura = useCallback((px: number) => {
  stageRef.current?.style.setProperty("--sheet-h", `${Math.round(px)}px`)
}, [])
// …
<BottomSheet open={railOpen} onClose={() => setRailOpen(false)} vista={moduloActivo} onAltura={publicarAltura}>
// …
<div className="stage" ref={stageRef}>
```

Las reglas actuales de los controles (`styles.css:537-540`) leen `--sheet-h` por herencia desde `.stage` sin cambios. M04 las afina después.

**A5 · Teclado** (S06). CSS de S06. El hook ya está en la versión objetivo de A4.

**A6 · Navegación y barras** (S07, S11, S13, S19). En `App.tsx`:

- Imports: `MQ_MOVIL` y `useMedia` de `./ui/media`, y `repartirModulos` de `./ui/navegacion`. `useRef` ya entró en A4.
- `railOpen` inicial: `useState(() => !window.matchMedia(MQ_MOVIL).matches)`.
- **Antes de los `return` tempranos (L512)**, por la regla de hooks, junto a `stageRef` y `publicarAltura` de A4:
  ```tsx
  const movil = useMedia(MQ_MOVIL)
  const [masAbierto, setMasAbierto] = useState(false)
  useEffect(() => {
    if (!masAbierto) return
    document.querySelector<HTMLButtonElement>(".guard-menu button")?.focus({ preventScroll: true })
    const cerrar = (event: PointerEvent) => {
      if (!(event.target instanceof Element) || !event.target.closest(".guard-menu, .guard-mas")) setMasAbierto(false)
    }
    document.addEventListener("pointerdown", cerrar)
    return () => document.removeEventListener("pointerdown", cerrar)
  }, [masAbierto])
  ```
- Después de `permitidos` (no son hooks):
  ```tsx
  const visibles = MODULOS.filter((item) => permitidos.includes(item.id))
  const { barra, mas } = movil ? repartirModulos(visibles) : { barra: visibles, mas: [] as typeof visibles }
  const enMas = mas.find((item) => item.id === moduloActivo)
  const menuMas = masAbierto && movil && mas.length > 0

  function irA(id: Modulo) {
    setModulo(id)
    setRailOpen(true)
    setMasAbierto(false)
  }
  function cerrarPanel() {
    setRailOpen(false)
    requestAnimationFrame(() => {
      document
        .querySelector<HTMLButtonElement>('.guard button[aria-pressed="true"], .guard .guard-mas.activo')
        ?.focus({ preventScroll: true })
    })
  }
  ```
- JSX de la navegación:
  ```tsx
  <nav className="guard" aria-label="Módulos">
    <div className="rail-mark"><img src="/isotipo.svg" alt="VerdePUCP" /></div>
    {barra.map((item) => (
      <button key={item.id} type="button" aria-pressed={moduloActivo === item.id} onClick={() => irA(item.id)}>
        {item.label}
      </button>
    ))}
    {mas.length > 0 && (
      <button
        type="button"
        className={enMas ? "guard-mas activo" : "guard-mas"}
        aria-expanded={menuMas}
        aria-controls="mas-modulos"
        onClick={() => setMasAbierto((abierto) => !abierto)}
      >
        {enMas?.label ?? "Más"}
      </button>
    )}
  </nav>
  {menuMas && (
    <div
      className="guard-menu"
      id="mas-modulos"
      role="group"
      aria-label="Más módulos"
      onKeyDown={(event) => {
        if (event.key !== "Escape") return
        event.stopPropagation()
        setMasAbierto(false)
        document.querySelector<HTMLButtonElement>(".guard-mas")?.focus({ preventScroll: true })
      }}
    >
      {mas.map((item) => (
        <button key={item.id} type="button" aria-current={moduloActivo === item.id ? "page" : undefined} onClick={() => irA(item.id)}>
          {item.label}
        </button>
      ))}
    </div>
  )}
  ```
- `<BottomSheet … onClose={cerrarPanel}>` y `onClick={() => setRailOpen(true)}` en «Saltar al panel».
- CSS de S07, S11 y S13. `.guard` en móvil lleva `z-index: var(--z-nav)`.

**A7 · Llevar a la vista** (S08, S14). `Labores.tsx` (ref y efecto) y el efecto `focus` de `CampusMap.tsx`.

**A8 · Importador** (S10). `Importaciones.tsx` y el CSS de `.tabla-scroll` y `.tabla`.

**Criterios de aceptación del Lote A**

A 1280×800 (coordinación):
- [ ] Catastro: la lista llega a la última área; la ficha llega a **Dar de baja**; `.panel-view` no desplaza; **Nueva área** y **Exportar CSV** se ven sin desplazar la lista.
- [ ] Inventario → Tachos: **Guardar** se alcanza; al tocar un registro la ficha empieza arriba.
- [ ] Cada cambio de módulo empieza arriba.
- [ ] Labores: elegir una labor en el mapa o en la lista deja su ficha arriba.
- [ ] Importar: la vista previa desplaza de lado y el panel no.
- [ ] A 1280×540 con admin: el riel desplaza por dentro y el documento no.
- [ ] D.3: el detector de desbordes da `[]` y todos los contenedores dicen «fin visible», en todos los módulos.
- [ ] D.4: el mapa se comporta igual que antes.

A 390×844 (coordinación y admin; los puntos con [dispositivo] se verifican en iPhone y Android reales):
- [ ] Hoja: al 92 % la barra superior sigue visible [dispositivo]; el asa responde desde el primer píxel; hay tope al subir; no hay franja vacía al pie; tocar el asa cicla anclajes.
- [ ] Catastro: maestro-detalle, con la ficha arriba y **Volver a la lista** que deja la fila a la vista; la ficha llega a **Guardar área** al 55 % y al 92 %; con **Editar geometría** activo se ven Oeste/Este/Norte/Sur.
- [ ] Inventario: tocar un registro lleva al formulario y **Guardar** se alcanza.
- [ ] Labores: tocar un pin abre la hoja con la ficha a la vista y el pin visible sobre la hoja.
- [ ] Barra inferior: cinco casillas como máximo, ninguna desborda, y «Más» abre, cierra y devuelve el foco. También a 360 px.
- [ ] Barra superior sin superposición, también a 360 px.
- [ ] Acceso a 740×360 y con teclado: desplaza hasta **Entrar**.
- [ ] Sin desbordes horizontales; la tabla de importación desplaza con el dedo.
- [ ] [dispositivo] iPhone: enfocar campos no amplía la página.
- [ ] [dispositivo] Teclado en iPhone y Android: campo visible, barra inferior oculta, y el anclaje vuelve al cerrar.
- [ ] [dispositivo] Android en pestaña: tirar en el tope de la hoja no recarga.
- [ ] Tras recorrer formularios con Tab: `document.scrollingElement`, `#root`, `.shell` y `.panel` siguen con scroll 0.
- [ ] Escape con teclado físico cierra la hoja desde un campo y devuelve el foco al módulo.

### Lote B · Diseño visual y bordes

**Archivos:** `src/styles.css`, `src/App.tsx` (cabeza, `laboresListas`, `.menu-btn`), `src/ui/BottomSheet.tsx` (cabeza), nuevo `src/ui/Esqueleto.tsx`, `src/panel/CatastroEditor.tsx`, `src/panel/CalendarioReservas.tsx`, `src/panel/InventarioCapas.tsx`, `src/panel/Labores.tsx`, `src/panel/Modulos.tsx`, `src/panel/Poda.tsx`, `src/panel/Vivero.tsx`, `src/panel/EvidenciasCampo.tsx`, `src/FechaCampo.tsx`.

- **B1 · Segmentados y botones** (V01, V02, V03, V07): CSS; `role="group"` y `aria-pressed` en CatastroEditor y CalendarioReservas; filas de Inventario como `.labor`.
- **B2 · Contenedores, grupos y ritmo** (V05, V06, V08): CSS; `className="form grupo"` en `FichaLabor`; tokens de radio en todos los literales.
- **B3 · Cabeza fija** (V04): prop `cabeza` en BottomSheet (ya prevista en A4); App le pasa la línea de estado con `role="status"`.
- **B4 · Estados** (V09, V10): `Esqueleto`, carga derivada en los ocho lugares, estilos de error y de `aria-invalid`, y `role="status"` en el aviso de Labores.
- **B5 · Fechas** (V11): `FechaCampo` con sincronía externa; reemplazos en Labores, Poda y Vivero.
- **B6 · Táctil y acabado** (V12, V13, V14).

**Criterios de aceptación del Lote B** (a 390×844 y a 1280×800):
- [ ] Los cuatro segmentados muestran la opción activa y el anillo de foco completo.
- [ ] Ningún botón tiene el estilo del sistema.
- [ ] No hay sombras dentro del panel: solo en la barra móvil, la hoja, los popups, «Más» y los controles del mapa.
- [ ] Radios: controles 6 px, hoja móvil 10 px, chips 4 px.
- [ ] Grupos de formulario con línea de 1 px y radio de 6 px; ningún borde «groove».
- [ ] Secciones separadas por 32 px y una línea; la cabeza sigue visible al desplazar.
- [ ] Con «Slow 3G» nunca aparece «No hay…» antes del dato.
- [ ] Un error de la API se ve como aviso con tono; una fecha inválida marca el campo.
- [ ] Las fechas se muestran `dd/mm/aaaa` y la API recibe `AAAA-MM-DD` (pestaña Network).
- [ ] A 390, los filtros de estado y las capas miden 44 px o más.
- [ ] Inventario muestra pestañas con mayúscula inicial y ninguna clave técnica.

### Lote C · Animaciones y microinteracciones

**Archivos:** `src/styles.css`. `src/ui/BottomSheet.tsx` y `src/App.tsx` no cambian si se aplicó la versión objetivo de A4: ya no tiene `reducido()` ni `transition` en línea, marca el arrastre con clases, y App publica `--sheet-h` en `.stage`.

- **C1 · Movimiento reducido** (M02). Va primero, para que cada animación nueva nazca con su alternativa.
- **C2 · Hoja y controles del mapa** (M01, M04).
- **C3 · Estados** (M03): `rise` por módulo, filas, avisos, ficha, esqueleto y menú «Más».

**Criterios de aceptación del Lote C:**
- [x] 390: la hoja entra en ≈260 ms desde la barra y sale en ≈180 ms (`@starting-style` + `allow-discrete` sobre `.panel.sheet`); el ajuste al anclaje anima; sin transición durante el arrastre (`.panel.sheet.arrastrando { transition: none }`); los controles del mapa acompañan (`transition: bottom var(--dur-hoja)…`, congelada durante `.arrastrando`).
- [x] 390 y 1280: el cambio de módulo es un fundido con 6 px en 180 ms (`rise` con `key={vista}`, ya de Lote A); la selección de fila cambia de tono en 120 ms (`.labor { transition: background-color var(--dur-feedback)… }`); los avisos y la ficha aparecen con fundido (`@keyframes aparecer` en `.banner`, `.status.error`, `.detail`).
- [x] Movimiento reducido: verificado con Playwright (`reducedMotion: "reduce"`, ver `/workspace/shots-tool/verify-loteD.mjs`) — `.panel.sheet` solo lista `opacity, display` en su transición (nada se traslada); los botones conservan `background, color, transform, border-color` en su transición (el color sigue cambiando); `CampusMap.tsx` ya usaba `jumpTo`/`duration: 0` desde antes de este lote y no se tocó; `mostrarEnPanel` ya usa `behavior: "auto"` desde el Lote A.
- [x] Ninguna transición de UI nueva supera 300 ms (260/180/120 ms), salvo el vuelo del mapa (900 ms, sin cambios, ya existente).
- [ ] Con la CPU a ×4 en DevTools, el arrastre no produce tareas largas de más de 50 ms — sin verificar: requiere el panel Performance de un Chrome interactivo, no disponible en este entorno headless. Hueco, no bloquea (el arrastre solo cambia `height`/`transform` de un elemento y `bottom` de los controles del mapa, sin recalcular el layout del documento; ver M04).

### Lote D · Verificación

Pasadas acotadas, según la guía: una ronda completa, que mezcla escritorio y teléfono; se corrige todo lo que aparezca en un solo lote; se hace una ronda de confirmación y se cierra. Cada resultado registra con qué se obtuvo: emulación de Chromium, toque sintetizado, iPhone o Android real. Lo que quede sin probar se anota como hueco, no bloquea.

**D.1 · Automático.** `cd apps/web && npm run lint && npm test && npm run build`. Los tests SSR deben seguir en verde sin tocar sus aserciones.

**D.2 · Matriz**

| Vista | Motivo | Roles |
|---|---|---|
| 390×844 | Referencia del teléfono (iPhone 12–14) | coordinación, admin, capataz |
| 360×740 | Android chico | coordinación, admin |
| 740×360 y 844×390 | Apaisado: acceso, riel y hoja | coordinación |
| 768×1024 | Tableta vertical (layout de teléfono) | coordinación |
| 1024×768 | Tableta horizontal (821–1199 px) | coordinación |
| 1280×800 | Referencia de escritorio | coordinación, admin, capataz |
| 1280×540 | Alto corto | admin |
| 1440×900 | Escritorio amplio | coordinación |

**D.3 · Scripts de consola**

Desbordes horizontales. Se ejecuta en cada módulo y el resultado esperado es `[]`:

```js
(() => {
  const out = []
  for (const raiz of document.querySelectorAll(".panel, .topbar, .guard")) {
    const lim = raiz.getBoundingClientRect()
    for (const el of raiz.querySelectorAll("*")) {
      if (el.closest(".tabla-scroll")) continue
      const r = el.getBoundingClientRect()
      if (r.width > 0 && (r.right > lim.right + 1 || r.left < lim.left - 1)) {
        out.push(`${el.tagName.toLowerCase()}.${[...el.classList].join(".")}`)
      }
    }
  }
  return out
})()
```

Alcance de cada contenedor con scroll. Todos deben decir «fin visible»:

```js
for (const el of document.querySelectorAll(".panel-view, .split-list, .split-detail, .guard")) {
  const oy = getComputedStyle(el).overflowY
  if (oy !== "auto" && oy !== "scroll") continue
  el.scrollTop = el.scrollHeight
  const ultimo = el.lastElementChild?.getBoundingClientRect()
  const caja = el.getBoundingClientRect()
  console.log(el.className, !ultimo || ultimo.bottom <= caja.bottom + 1 ? "fin visible" : "FIN OCULTO")
}
```

Documento quieto. Tras usar formularios, el resultado esperado es `true`:

```js
[document.scrollingElement, ...document.querySelectorAll("#root, .shell, .panel")]
  .every((e) => e.scrollTop === 0 && e.scrollLeft === 0)
```

Objetivos táctiles a 390. La lista esperada es vacía, salvo enlaces de texto dentro de párrafos:

```js
[...document.querySelectorAll(".panel button, .panel label, .guard button, .topbar button")]
  .map((el) => [el, el.getBoundingClientRect()])
  .filter(([, r]) => r.width > 0 && r.height < 44)
  .map(([el, r]) => `${el.textContent.trim().slice(0, 24)} ${Math.round(r.width)}×${Math.round(r.height)}`)
```

**D.4 · Regresión del mapa** (a 390 y a 1280):
- [ ] Arrastrar, hacer zoom con rueda y pellizco, rotar e inclinar (clic derecho en escritorio, dos dedos en el teléfono).
- [ ] Plano/Relieve: la animación de inclinación y las extrusiones.
- [ ] Encender y apagar cada capa del catastro y del inventario.
- [ ] Popups de labor, área y bebedero; cerrarlos.
- [ ] «Marcar labor» y clic en el mapa: el pin de borrador aparece.
- [ ] Catastro → Editar geometría: arrastrar un vértice (escritorio) y moverlo con los botones y las flechas (teléfono y teclado).
- [ ] Redimensionar la ventana y rotar el teléfono: el lienzo llena `.stage`, sin franjas en blanco.
- [ ] Con la hoja abierta, los controles del mapa quedan sobre la hoja; cerrada, sobre la barra.

**D.5 · Regresión de datos:**
- [ ] Crear una labor, con y sin conexión (cola).
- [ ] Cambiar el estado, reasignar y archivar.
- [ ] Subir una evidencia.
- [ ] Guardar la ficha de la labor.
- [ ] Guardar y dar de baja un área y una zona en Catastro.
- [ ] Guardar y eliminar en Inventario.
- [ ] Poda, Vivero y Riego.
- [ ] Solicitud y orden.
- [ ] Descargar CSV y Excel en Reportes.
- [ ] Importar: vista previa → Confirmar escritura → Revertir lote, incluido el `confirm` de filas editadas.
- [ ] Comparar en Network los cuerpos de las peticiones con la línea base: deben ser idénticos, salvo el formato visible de las fechas.

**D.6 · Accesibilidad:**
- [ ] Recorrido con Tab en cada módulo: el orden sigue al visual y el foco se ve en todo.
- [ ] Escape cierra la hoja y el menú «Más».
- [ ] `role="status"` anuncia el área tocada y los avisos.
- [ ] Contraste AA de los avisos nuevos (alerta sobre el tinte de alerta y yucca sobre hoja).

---

## 4. Riesgos y qué no tocar

### 4.1 MapLibre

- **Tamaño del contenedor.** `.stage` (`position: relative; min-width: 0; min-height: 0`) y `.map-host` (`position: absolute; inset: 0`) se quedan como están. No agregar padding ni borde. MapLibre observa `.map-host` con `ResizeObserver`: no animar `grid-template-columns` ni el ancho del panel en escritorio, porque el lienzo WebGL se redimensionaría cada 50 ms durante la animación. El modo teclado no cambia el tamaño de `.stage`; por eso no se usa `interactive-widget=resizes-content`.
- **Orden del CSS.** `styles.css` debe cargar después de `maplibre-gl.css` (import en `CampusMap` → `App` → `main.tsx`, antes de `import "./styles.css"`). `.map-host { position: absolute }` le gana a `.maplibregl-map { position: relative }` solo por ese orden.
- **Gestos.** El lienzo usa `touch-action: none`, que pone MapLibre. No poner `touch-action`, `overflow` ni `pointer-events` en `.stage`, `.map-host` ni `.maplibregl-*`, salvo las posiciones de controles documentadas. La hoja, la barra y el menú «Más» son hermanos del mapa, no ancestros: su scroll no puede encadenarse al mapa. Mantenerlos fuera de `.stage`. Nunca `pointer-events: none` en sus contenedores, porque los toques caerían al mapa.
- **Controles.** Las reglas de `.maplibregl-ctrl-bottom-*` necesitan `!important` en el teléfono porque MapLibre fija `bottom` y `right`. Desde A4, `--sheet-h` vive en `.stage` y no en `:root`: solo lo leen esas reglas y `desfaseHoja`, por herencia. No leerlo desde fuera de `.stage`.
- **Cámara.** Solo se toca el efecto `focus`. Usar `offset`, que no persiste; nunca `padding`, que persiste y movería `fitBounds`, el relieve y el dibujo. El camino reducido queda sin animación.
- **Dibujo.** `draw.ts` desactiva `dragPan` mientras se arrastra un vértice. No cerrar la hoja con clics en el mapa mientras `modoDibujo` está activo. No capturar las flechas fuera del asa.
- **Popups.** Se pueden ajustar sus estilos (`.cv-popup`), no su creación.
- **Rendimiento.** Nada de `backdrop-filter` ni `blur` sobre el lienzo; además, el vidrio está vetado por DESIGN.md.

### 4.2 Datos y comportamiento

- No se toca `apps/api`, y ningún cuerpo de petición cambia. Las fechas siguen viajando como `AAAA-MM-DD`.
- Importador: solo cambia la vista previa. Previsualizar, confirmar, revertir y el `confirm` quedan igual.
- Catastro: `verLista` es solo presentación. `area`, `zona`, `altaArea`, `altaZona` y `editandoGeom` conservan su semántica.
- Labores: sin `key` en `.detail` (O1).
- «Más» cambia solo la presentación en el teléfono. `modulosDe(rol)` y los permisos quedan igual.
- Tests SSR: conservar los textos que se buscan: «Vista previa», «Confirmar escritura», «Revertir lote», «Lugares», «Puntos PUCP», «Mes», «Semana», «Día», «Reservas», «Anterior», «Siguiente», «Zonas», las etiquetas de campo de Catastro, Inventario y Labores, «Marcar labor» o «Cancelar marca», «Servicio tercerizado», «Confirmar Poda», «Fecha de solicitud», «Fecha de atención», «Sin estado», «PO-1», «Código externo», «683» y «Mes» (Vivero), y los de Reportes. **En `ReportesPanel` no puede aparecer «%» ni «PDF» en el marcado**; ojo con estilos en línea que usen porcentajes.
- Los tests nuevos se agregan a la lista explícita del script `test` de `package.json`.

### 4.3 Compatibilidad

| Recurso | Soporte | Si falta |
|---|---|---|
| `:has()` | Chrome 105, Safari 15.4, Firefox 121 (ya se usa) | Sin bloqueo de raíz ni partición: degradación visible, aceptable |
| `@starting-style`, `allow-discrete` | Chrome 117, Safari 17.5, Firefox 129 | La hoja aparece y se va sin animación |
| `overflow: clip` | Chrome 90, Safari 16, Firefox 81 | Respaldo `hidden` declarado antes |
| `dvh` | Chrome 108, Safari 15.4, Firefox 101 | Respaldo `vh` declarado antes |
| `visualViewport` | Chrome 61, Safari 13 | El hook no hace nada |
| `scrollbar-gutter`, `scrollbar-width` | Chromium y Firefox | No hace nada |

### 4.4 Qué no tocar

- `apps/api/**`.
- `src/api.ts`, `src/producto.ts`, `src/operacion.ts`, `src/session.ts`, `src/types.ts`, `src/inventario.ts`, `src/offline/**`.
- `src/map/draw.ts`, `src/map/coverage.ts` y, en `src/map/CampusMap.tsx`, todo salvo el efecto `focus`.
- Los módulos de lógica y sus tests: `panel/catastro.ts`, `panel/inventarioCapas.ts`, `panel/importaciones.ts`, `panel/calendario.ts`, `panel/poda.ts`, `panel/vivero.ts`. En `ui/bottomSheet.ts` no cambian ni los valores de `ANCLAJES` ni las firmas; se reutilizan.
- `vite.config.ts`, el manifiesto PWA, `index.html` (salvo que D encuentre algo) y los `nginx*.conf`.
- Colores y capas del mapa.
- Sin dependencias nuevas: nada de librerías de animación, kits de UI ni Tailwind.
