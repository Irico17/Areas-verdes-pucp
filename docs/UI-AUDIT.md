# Auditoría de interfaz — Campus Verde

Fecha: 2026-09-24. Modo Operate. El sistema vigente se escaneó desde `apps/web/src/styles.css`, `App.tsx`, `panel/Modulos.tsx`, `panel/Labores.tsx` y `map/CampusMap.tsx`, más las capturas de `/opt/cursor/artifacts/screenshots-v2/`. Esta nota se escribe y se commitea antes de cualquier cambio de UI.

Preguntas de cierre de la crítica: omitidas. El encargo pide ejecución autónoma y prohíbe preguntar.

Crítica (Impeccable): Assessment A en un subagente de diseño; Assessment B con `impeccable detect --json apps/web/src` (salida `[]`, código 0) más revisión manual. El detector no vio los problemas de composición. La revisión manual sí.

Paso extra Taste (`redesign-existing-projects`): lista al final. Donde choca con Impeccable, manda Impeccable (Operate: familiaridad ganada, un acento, sin reescritura de framework).

Camino de construcción de esta sesión: código directo, no comps. No quedó guardado en `.impeccable/config.json` porque nadie respondió esa pregunta. Un flujo de comps con aprobación no cabe en un encargo que prohíbe preguntas y pide la app entera más infra.

## Sistema vigente (escaneo)

Tokens en `:root`: tinta `#141816`, campo `#e7ece6`, niebla `#f4f7f4`, dosel `#145c3e`, alerta `#8d2f2c`, línea `#c5cfc8`, apagado `#3e4943`, pendiente `#8a6a12`. Un hex suelto: `#d5ddd7` en `.guard .who`. Tipografía: Source Sans 3 en la UI, Newsreader solo en el nombre. Escala plana: h1 40 px, h2 16 px, cuerpo 15 px. Radio 0 en casi todo. Sin sombras, sin transiciones, sin `prefers-reduced-motion`. Grilla `76 px | 280–340 px | mapa`. Corte móvil a 820 px. Controles a 36 px de alto. El mapa usa ocre `#a8843d` en zonas, un color que el producto ya había retirado de la marca.

## Pantalla por pantalla

### Acceso

El formulario vive en `.login` (`max-width: 420px; padding: 48px 24px`) sin centrado. En 1440×900 el resto del lienzo es campo vacío. Se leen el nombre, la nota de SSO, usuario, clave, Entrar y la lista de cuentas locales. Jerarquía pobre: el título no ancla la página porque no tiene un plano detrás. No hay estado de error distinguido del texto de ayuda más que el color. No hay esqueleto: o está la sesión o no.

### Mapa, plano

Riel negro de 76 px con etiquetas de 12 px (Mapa, Labores, Catastro, Solicitudes, Reportes, Catálogos, Coordinación, Salir). El panel izquierdo es la lista de capas, larga, del catastro al inventario y la agenda ficticia. El mapa OSM ocupa el resto, con pines de labor. Plano / Relieve son dos botones flotantes. El cambio a relieve anima pitch con `easeTo` 650 ms, no con `flyTo`. La leyenda de labores (color = estado, letra = tipo) no está en esta vista. La agenda ficticia se mezcla con capas operativas.

### Mapa, relieve

La misma cáscara. La extrusión de áreas y las huellas OSM se leen, pero el riel y el panel no cambian de jerarquía. El control de modo no se distingue como estado persistente más allá de un botón pulsado.

### Labores (coordinación)

Una sola columna con scroll: filtros, alta con pin, lista, detalle (estado, reasignar, archivar, evidencias, bitácora) y, debajo, el bloque de riego. Elegir una labor empuja el alta fuera de vista. Hay una labor tercerizada visible (“Poda contratada”). `equipoId` llega al panel y no se dibuja: coordinación no cambia de equipo desde la UI. El aviso, el error y la ayuda comparten `.hint` / `.status`.

### Formulario de labor

Coordenadas, tipo, título, “Sugerir tipo” con la frase de regla local, ejecutor, detalle, equipo. Funciona. Visualmente es una pila de campos idénticos, sin agrupación (lugar / qué / quién). El pin en el mapa y el formulario no comparten un mismo gesto visual.

### Catastro

Búsqueda y lista. De 521 áreas, 501 no tienen nombre. La consulta ordena por `feature_id` y corta en 40, así que la primera página es casi toda “Sin nombre” (los jardines con nombre, código tipo `C 34`, quedan fuera del corte). La ficha (nombre, uso, riego, referencia, Guardar ficha, área sin GPS) se pinta debajo de la lista: hay que hacer scroll. Muchas filas iguales cansan antes de llegar al formulario.

### Solicitudes y órdenes

La solicitud `OSG-2026-0142` está. El bloque de órdenes dice “No hay órdenes.” porque `ordenes_servicio` está vacía, aunque existe “Poda contratada” como labor tercerizada. Crear una orden exige una labor elegida en otro módulo y solo muestra un UUID cortado. El vacío no enseña el siguiente paso con una acción clara.

### Riego

Dos registros de demostración y el formulario sector / turno / fecha / nota. La fecha es `input type="date"`: en este navegador se ve `09/24/2026`. El aviso de cobertura no definida está. El capataz `norte` ve la fila de Equipo Riego: `ListarRiego` no filtra por equipo. El alta sí fuerza el `capataz_id` de la sesión. El hueco es de lectura y está en el servidor.

### Reportes

Conteos (bloqueada 1, cerrada 1, en proceso 1, pendiente 5), Descargar CSV y Descargar Excel, y la frase de que no son el indicador oficial. El filtro de fechas también es `type="date"`. En la tabla aparece “Prueba de pin”, una labor de prueba que no está en la semilla SQL: quedó en la base. El estado del filtro es texto libre.

### Catálogos

Clases, filas con Desactivar, formulario de alta. Correcto y plano. Desactivar es un enlace con `padding: 0`, por debajo de 44 px. Coordinación ve el módulo y el servidor rechaza la escritura: el botón promete algo que la API niega.

### Admin

Seis cuentas, ids de equipo, matriz de permisos, frase de SSO pendiente. Es una lista densa sin agrupación por rol. Sirve. No tiene jerarquía de “quién puede entrar hoy”.

### Vista capataz

El riel se reduce a Mapa, Labores y Catastro. La lista de labores sí es solo Equipo Norte (tres). El bloque de riego, en el mismo panel, sigue listando Bosque húmedo de Equipo Riego. En el teléfono, `.guard .who` se esconde: no se ve el nombre ni Salir.

### Móvil

Mapa a 390×844 con la fila de módulos y Relieve. La labor muestra el detalle y “Esta labor no tiene archivos”. No hay indicador de cola si la cola está vacía, y cuando hay cola el texto compite con el detalle. Los objetivos del riel son bajos.

## Qué mejorar

### Jerarquía

- El acceso tiene que ocupar el viewport: plano de marca y formulario centrado. El lienzo vacío no es una decisión, es un formulario sin layout.
- Dentro de la app, una barra de identidad (nombre, conteos, modo, persona, salir) y un panel con cabeza fija. La lista y la ficha no comparten el mismo scroll.
- Riego sale de debajo de labores y vive en su propia sección del panel de labores, con cabeza, no como cola infinita.
- La agenda ficticia se separa con un rótulo que no parezca una capa más del catastro.

### Tipografía

- Familjen Grotesk para UI. Newsreader solo en “Campus Verde”.
- Escala fija: 56 / 22 / 16 / 15 / 12. Cifras tabulares en conteos, coordenadas y fechas.
- Títulos en oración. Sin ceja en mayúsculas. `text-wrap: balance` en el nombre del acceso.

### Color

- Ver la ficha de tokens de abajo. Se retira el riel negro y el ocre de zonas en el mapa.
- Un acento (yucca). Alerta y pendiente solo en estado.
- Texto secundario teñido de la tinta, nunca gris puro. Selección, caret y foco salen de la misma paleta.

### Espaciado

- 4 / 8 / 12 / 16 / 24 / 40. La etiqueta queda a 4 px de su campo; el bloque siguiente a 24 px.
- Más aire arriba de un título de panel que debajo.

### Componentes

- Botón primario, botón quieto, botón de peligro. Misma altura (44 px).
- Campo con etiqueta visible. Fecha propia `dd/mm/aaaa`.
- Fila de lista, no tarjeta. Ficha de catastro siempre visible.
- Vacío con la acción. Error con el problema y cómo seguir. Carga con esqueleto de tres filas.
- Selector de estado en reportes, no texto libre.
- Desactivar y Salir con área táctil real. Coordinación no ve Desactivar si no puede.

### Vacíos, carga, movimiento

- Esqueleto al entrar y al cambiar de módulo, no un spinner al centro.
- Transición de panel 180 ms (opacidad y 6 px). Presión de botón 140 ms.
- `map.flyTo` al elegir una labor o un área.
- `prefers-reduced-motion`: sin traslación ni shimmer; el color de estado permanece.

### Responsive

- ≤820 px: banda de módulos horizontal con 44 px de alto, Salir en la barra, panel como hoja. Mapa del capataz y detalle de labor se verifican a 390×844.

### Accesibilidad

- `lang="es-PE"`. Foco visible. Contraste de cuerpo ≥ 4.5:1 sobre hoja y sobre yucca. No hay trampa de teclado en la hoja móvil (Escape cierra).

### Mapa

- Áreas: relleno `#1a5c44` al 28 %, línea `#0e3b2c`.
- Zonas: línea `#3e5148`, relleno `#5c6a62` al 12 %. Sin ocre.
- Jardines de reserva `#2f4a44`. Xerofítica `#6e5344`.
- Popup con la hoja, radio 10 px, sombra flotante.
- Plano / Relieve como control segmentado con estado pulsado claro.

## Plan, en este orden

1. Tokens, fuentes, `lang`, foco, selección, scrollbar. Login centrado con franja de yucca.
2. Cáscara: barra, navegación clara, panel con cabeza. Movimiento corto y reducible.
3. Catastro: orden con nombre primero, título = nombre o código, ficha visible sin scroll de la lista.
4. Fechas `es-PE` en riego, reportes y cualquier otro texto de fecha.
5. Labores: formulario agrupado, sugerencia visible, riego en bloque propio. Selector de equipo para quien no es capataz.
6. Solicitudes con órdenes de semilla y vacío que explique la tercerización.
7. Reportes sin “Prueba de pin”. Catálogos y admin con la misma hoja.
8. Capataz: el servidor no devuelve riego de otro equipo. Prueba incluida.
9. Móvil: identidad visible, objetivos de 44 px, mapa y detalle.
10. Endurecer vacíos, errores y carga. Pulir contraste y estados hover / foco / disabled.

Fuera de esta pasada visual, pero del mismo encargo: Docker, CI y Terraform. No cambian la interfaz.

## Ficha de tokens

| Token | Valor | Uso |
| --- | --- | --- |
| tinta | `#1a2420` | texto |
| papel | `#e4ebe4` | fondo |
| hoja | `#f7faf6` | panel, barra, campo |
| linea | `#c5d2c8` | borde 1 px |
| yucca | `#1a5c44` | acción, selección, en proceso |
| yucca-deep | `#0e3b2c` | hover, franja de acceso |
| yucca-ink | `#f3faf6` | texto sobre yucca |
| alerta | `#8c3832` | error, bloqueada |
| pendiente | `#8a6410` | pendiente |
| muted | `#3e5148` | texto secundario |
| cerrado | `#5c6a62` | cerrada, cancelada |

Tipo: Newsreader 56/500 solo en el acceso; Familjen Grotesk 22/600, 16/600, 15/400, 12/600. Tabulares en datos.

Espacio: 4, 8, 12, 16, 24, 40 px.

Radio: control 6 px, hoja flotante 10 px, marca 4 px.

Elevación: solo flotante, `0 10px 28px -16px rgba(26, 36, 32, 0.45)`.

Movimiento: 140 ms presión, 180 ms panel, 900 ms vuelo del mapa. Curva `cubic-bezier(0.2, 0.8, 0.2, 1)`.

Mapa: áreas `#1a5c44` / `#0e3b2c`; zonas `#5c6a62` / `#3e5148`; reserva `#2f4a44`; xerofítica `#6e5344`.

## Puntuación técnica (auditoría, antes del cambio)

| Dimensión | Nota | Por qué |
| --- | --- | --- |
| Accesibilidad | 2 | Etiquetas sí; foco flojo, objetivos de 36 px, sin reduced-motion, Salir oculto en móvil. |
| Rendimiento | 3 | Sin animaciones caras. El panel largo re-renderiza con el mapa, aceptable. |
| Tema | 2 | Hay tokens, más un hex suelto y colores de mapa fuera del sistema. Sin tema oscuro: no se pide. |
| Responsive | 2 | Hay corte a 820 px. Identidad oculta, objetivos bajos, fechas en locale del navegador. |
| Integridad | 2 | El detector no marcó nada. La composición sí: acceso sin layout, lista y ficha en el mismo scroll, riego de otro equipo en pantalla. |

## Taste, pasada extra

Se corrige: fuente con carácter (no Inter ni Source Sans genérico), pesos 500 y 600, tabulares, un solo acento, sombras teñidas, hover y presión, foco, esqueletos, vacíos, fechas, `100dvh` donde haga falta, objetivos táctiles, semántica (`nav`, `main`, `aside`).

No se aplica, porque Impeccable en Operate lo prohíbe o el producto no lo pide: fotos de stock, grano en toda la app, asimetría de afiche, parallax, tres columnas de tarjetas, iconos de “lanzamiento”, migrar a Tailwind, reescribir desde cero el mapa.

## Auditoría móvil (campo)

El teléfono no es un escritorio encogido. Lo usa el capataz al sol, con una mano, a veces sin red, y puede instalar la PWA. Anchos a cubrir: 360–430 px. Tableta (821–1199): la misma grilla de escritorio, más estrecha. Escritorio 1280–1920: riel lateral y panel.

### Lo que había

- A 820 px la navegación era una fila horizontal con scroll. En 360 px “Solicitudes” y “Catálogos” se salían.
- El panel era una columna lateral absoluta, no una hoja sobre el mapa. El mapa quedaba a la mitad o tapado de lado.
- Salir se escondía (`.who { display: none }`).
- Controles a 36 px. El zoom de MapLibre, más chico.
- Sin `viewport-fit=cover` ni `safe-area-inset`. En un teléfono con barra de gestos, la última fila queda bajo el sistema.
- “Marcar labor” y “Adjuntar” vivían a mitad del scroll, lejos del pulgar.
- La animación, si se agregaba igual que en escritorio, era el mismo desplazamiento. En un equipo lento tiene que ser corta y solo de opacidad/traslación pequeña.

### Lo que debe quedar

- Barra inferior fija, sin scroll horizontal. Cada módulo cabe en un flex igual, texto en dos líneas si hace falta, alto mínimo 48 px, más `env(safe-area-inset-bottom)`.
- El mapa a sangre. La barra superior flota, con Plano/Relieve y Salir al alcance, y `safe-area-inset-top`.
- La ficha es una hoja inferior (`bottom sheet`) de hasta 74 dvh, con asa y “Cerrar hoja”. Al entrar en el teléfono la hoja nace cerrada: lo primero es el mapa.
- “Marcar labor” y “Adjuntar foto o PDF” se pegan al borde inferior de la hoja (`position: sticky`), ancho completo, 48 px.
- Zoom del mapa a 44 px y subido para no quedar bajo la barra.
- Contraste de cuerpo sobre hoja y de texto claro sobre yucca, pensado para sol: sin grises lavados.
- Movimiento: 180 ms de la hoja, 140 ms del botón. `prefers-reduced-motion` anula la traslación.
- PWA: `display: standalone`, `theme-color` yucca, `viewport-fit=cover`. El service worker sigue solo en producción.
- Tableta: no usa la barra inferior. Mantiene riel y panel para quien coordina con un iPad en horizontal.
- Sin scroll horizontal en 360 px: `overflow: hidden` en la cáscara y en la barra.
