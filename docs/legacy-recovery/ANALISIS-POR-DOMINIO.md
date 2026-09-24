# Análisis por dominio — Mapa de Control Ambiental Campus (PUCP)

**Fuente viva:** https://mapa-web-6.vercel.app/  
**Recovery:** `/workspace/mapa-web-6-recovery/`  
**Fecha:** 2026-09-24 (America/Lima)  
**Stack:** Leaflet 1.9.4 + MarkerCluster + leaflet.heat + Turf 6 + PapaParse · monolito `script.js` (~4920 líneas)

Este documento detalla **cada dominio operativo** del mapa: datos reales (campos, volúmenes, URLs), cómo se cargan/renderizan, UI, dependencias cruzadas, estado de recuperación y riesgos concretos para una futura migración a React + mapa 3D.  
**No incluye reescritura de código.**

---

## Tabla rápida de estado

| Dominio | Estado recuperación | Fragilidad |
|---------|---------------------|------------|
| Áreas verdes (uso + riego) | **ok** | Baja (GeoJSON local) |
| Jefes de grupo / sectores | **ok** | Baja |
| Tachos / residuos | **ok** | Media (Sheet + zonas inline) |
| Flora (CSV + gviz + Drive + edición) | **parcial** | Alta (doble pipeline, APIs) |
| Cafetos | **ok** | Media (CSV legacy frágil) |
| Vivero | **ok** | Media (Sheet remoto) |
| Fauna | **ok** | Baja |
| Bebederos / agua | **ok** | Media (Drive fotos + cluster) |
| Monitoreo / poda / heatmap 2026 | **ok** | Media (4 Sheets + heat) |
| Reservas de jardines | **parcial / fallido sheet** | Alta (401) |
| Adicionales (puertas, estacionamientos, xerofítica) | **ok** (datos pobres) | Baja–media |
| Veredas de alto tránsito | **ok** | Baja |
| Barredoras 1 y 2 | **ok** (simulación local) | Baja datos / media lógica |
| Buscador / puntos PUCP | **ok** | Media (Sheet scrape) |
| Mapas base | **ok** | Alta ToS (Google tiles) |

---

## Áreas verdes (uso + riego)

### 1. Propósito
Inventariar y visualizar los **polígonos de áreas verdes del campus** clasificados por **uso de suelo** y por **sistema de riego actual / proyectado**, para planificación de riego tecnificado y manejo de jardines.

### 2. Fuentes de datos
| Fuente | Path / URL | Notas |
|--------|------------|-------|
| GeoJSON local | `areas_verdes.geojson` | Estático en Vercel/recovery |

**Schema real (properties):**

| Campo | Ejemplo | Rol |
|-------|---------|-----|
| `Nombre` | `"Bosque Húmedo"` | Etiqueta popup |
| `Uso` | `"Áreas de uso administrativo"` | Estilo por uso |
| `Riego act` | `"Sin riego tecnificado"` | Estilo por riego |
| `Proy riego` | `"Goteo"` / `"Por validar con unidad"` | Proyección |
| `Área`, `Perimetro`, `código` | `11147.353`, `712.399`, `"G 13"` | Métricas |

**Enums observados:**
- **Uso:** administrativo (351), manejo sostenible (96), recreativo (70), deportivas (2), institucional (1), conservación (1).
- **Riego act:** Sin riego tecnificado (516), aspersión (4), goteo (1).
- **Proy riego:** Por validar (350), Goteo (96), Falta aspersión (71), Cuenta con aspersión (4).

### 3. Volumen
- **521** features `MultiPolygon`
- **~984 KB**

### 4. Cómo se carga y renderiza
- `fetch('areas_verdes.geojson')` (~L3870).
- Dos paseos `L.geoJSON` sobre el mismo dataset:
  - Vista **uso** → `obtenerCategoriaUso` / `obtenerColorPorUso` → `subcapasAreasVerdesUso[cat]`.
  - Vista **riego** → `obtenerCategoriaRiego` / `obtenerColorPorRiego` (lee `properties["Riego act"]`) → `subcapasAreasVerdesRiego[cat]`.
- Popups: `vincularPopupAreaVerde` (Nombre, código, Área, Perimetro, Uso, Riego act, Proy riego).

### 5. UI / filtros
- Maestro `chk-nuevas-areas` + subchecks `.chk-sub-area` (`data-uso`: institucional, administrativo, recreativo, sostenible, otros).
- Maestro `chk-riego-actual` + `.chk-sub-riego` (`data-riego`: gotero, aspersion, cisterna, manual, sin_riego, otros).
- Sincronización: `actualizarEstadoMaestro`.
- **Desajuste UI↔datos:** la UI menciona “Cisterna / Camión” y “Riego Manual”, pero en el GeoJSON casi todo es `Sin riego tecnificado` (516/521); la categorización por string matching puede caer en “otros”.

### 6. Dependencias cruzadas
- Comparte geometría/atributos con **jefes de grupo** (`jefe_de_grupo.json` es casi el mismo shape + campo `jefes`).
- **Xerofítica** y **reservas** reutilizan campos similares (`Área`, `Riego act`).
- Panel de capas “Adicionales campus”.

### 7. Estado de recuperación
**ok** — archivo completo y legible offline.

### 8. Riesgos y recomendaciones React + mapa 3D
- Unificar `areas_verdes` y `jefe_de_grupo` en **una FeatureCollection tipada** con `uso`, `riegoActual`, `riegoProyectado`, `jefeId` (hoy hay duplicación ~1 MB×2).
- En MapLibre/deck.gl: `fill` + `fill-extrusion` con altura proporcional a `Área` o dummy height para “relieve” de jardines; evitar dos copias de geometría en memoria.
- Normalizar enums de riego en ETL (mapear UI `cisterna`/`manual` a valores reales o documentar que son placeholders).
- No depender de nombres de propiedad con tilde (`Área`, `código`) sin alias en el schema Zod.

---

## Jefes de grupo / sectores

### 1. Propósito
Asignar **responsabilidad operativa de jardinería** por polígono a jefes de grupo (responsable, responsable, responsable) y sectores especiales (campo deportivo, bosque húmedo), con métricas agregadas de área/perímetro.

### 2. Fuentes de datos
| Fuente | Path | Notas |
|--------|------|-------|
| JSON GeoJSON-like | `jefe_de_grupo.json` | 534 features |
| Zonas supervisores | **inline** `geojsonSupervisores` en `script.js` (~L550) | 4 MultiPolygons Zona1–4; usado por tachos, no por jefes |

**Campos clave:** mismos que áreas verdes + **`jefes`**.

| `jefes` (valor crudo) | Conteo |
|----------------------|--------|
| `responsable` | 259 |
| `responsable` | 168 |
| `responsable` | 104 |
| `campo depo` | 2 |
| `Bosque húme` | 1 (truncado) |

### 3. Volumen
- **534** features (505 Polygon + 29 MultiPolygon)
- **~1002 KB**

### 4. Cómo se carga y renderiza
- `fetch('jefe_de_grupo.json')` (~L3666).
- `capaBaseGeoJSON` con `style` vía `obtenerEstiloPoligono(jefe)`.
- `onEachFeature` clasifica por substring (`oscar`/`óscar`, `andres`, `alfonso`, `campo`/`depo`, `bosque`/`húme`) hacia:
  - `capasresponsable`, `capasresponsable`, `capasresponsable`, `capasCamposDepo`, `capasBosqueHumedo`
- Acumuladores `totalesPorJefe` + popup con área/perímetro **global del jefe** (`mostrarResumenMétricas`).

### 5. UI / filtros
- Maestro `chk-todas-verdes`.
- `chk-oscar`, `chk-andres`, `chk-alfonso`, `chk-camposdepo`, `chk-bosque`.
- Vinculación: `vincularCapaPro(...)`.

### 6. Dependencias cruzadas
- Solapa conceptualmente con **áreas verdes** (mismo dominio espacial).
- **Monitoreo 2026** usa responsables `responsable`/`responsable`/`responsable` (texto libre, no FK).
- **Tachos** usan otro recorte: `geojsonSupervisores` (Zona1–4), distinto de jefes.

### 7. Estado de recuperación
**ok**.

### 8. Riesgos y recomendaciones React + mapa 3D
- Truncamiento `"Bosque húme"` → normalizar catálogo de jefes (`jefeId` estable).
- Separar **capa de responsabilidad** de **capa de uso/riego** en el store (Zustand); un solo GeoJSON fuente.
- Colores por jefe en `obtenerEstiloPoligono` hoy hardcodeados → tokens de diseño.
- En 3D, extruir por jefe con hue distinto; tooltips con métricas precomputadas (no recalcular en click como strings HTML).

---

## Tachos / residuos

### 1. Propósito
Inventario de **estaciones de tachos ecológicos** por fracción (papel, plástico, etc.), con filtro espacial por zona de supervisor y panel de resumen operativo.

### 2. Fuentes de datos
| Fuente | Variable / path | URL |
|--------|-----------------|-----|
| CSV Sheet | `urlPuntosEcologicosCSV` → `data/sheets/tachos.csv` | `…2PACX-1vQK3aVBw…/pub?gid=657037007&output=csv` |
| Zonas | `geojsonSupervisores` inline | properties `ZONA`: Zona1…Zona4 |

**Columnas reales (sample):** `ID`, `Latitude`, `Longitude`, `Note`, `No Aprovechables`, `Papel y Cartón`, `Plástico`, `Vidrio`, `Pilas`, `Peligrosos`, `RAEE`, `Metales`, `Aniquem`, `Intermedios Plastico`, `Intermedios Metal`, `Foto`, `Espacios`, `Lugar`, `Accion`, `Tacho Actual`, `Tacho Nuevo`, `Recomendaciones`, `GOOGLE MAPS`.

**Fracciones con cantidad > 0 (sobre 185 filas):**

| Fracción | Estaciones con ≥1 |
|----------|-------------------|
| No Aprovechables | 144 |
| Plástico | 128 |
| Vidrio | 120 |
| Papel y Cartón | 70 |
| Pilas | 29 |
| Aniquem | 20 |
| RAEE | 7 |
| Intermedios Plastico | 6 |
| Intermedios Metal | 3 |
| Metales | 2 |
| Peligrosos | **0** |

### 3. Volumen
- **185** filas de datos / **~36 KB** CSV
- Marcadores: N estaciones × M fracciones (offset longitudinal `0.000035`)

### 4. Cómo se carga y renderiza
- `cargarPuntosEcologicos()` + PapaParse `header: true`.
- Zona: `obtenerZonaDeCoordenada(lon, lat, geojsonSupervisores)` (ray casting propio, no Turf).
- `configuracionTachos` mapea nombre → color → `capa*`.
- Iconos: `crearIconoTacho` / `generarHTMLIconoTacho`.
- Click → `abrirSidebarTacho(datosEstacion)`.
- Totales: `actualizarResumenTachos`, filtro `filtrarTachosPorZona`.

### 5. UI / filtros
- Maestro `chk-todos-tachos` + 11 checkboxes (`chk-no-aprov` … `chk-intermedios-metal`).
- `#panel-resumen-tachos`, `#sidebar-info`.
- Selector de zona (variable global `zonaSeleccionadaFiltro`).

### 6. Dependencias cruzadas
- **Supervisores inline** (obligatorio para zonificación).
- Fotos Drive vía `transformarEnlaceDrive` (misma utilidad que flora/bebederos).
- Independiente de jefes de jardinería (otro zoning).

### 7. Estado de recuperación
**ok** (CSV + lógica + zonas inline).

### 8. Riesgos y recomendaciones React + mapa 3D
- Modelo: estación 1:N fracciones (hoy se “explota” en N markers offset) — en deck.gl usar `IconLayer` con clustering o un solo punto + popup multi-fracción.
- Columna `Peligrosos` vacía: o retirar del UI o completar datos.
- Inconsistencia nombre CSV `Intermedios Plastico` vs config `"Intermedios Plástico"` (`nombreFila` sin tilde) — frágil.
- Extraer `geojsonSupervisores` a archivo; no dejar 39 KB embebidos en el bundle React.
- Auth: Sheet público → API con service account.

---

## Flora (CSV legacy + gviz nueva + Apps Script + fotos Drive)

### 1. Propósito
Inventario arbóreo/palmeras del campus con ficha (especie, ubicación, foto), filtros por tipo de vegetación, panel lateral y **edición de coordenadas** hacia Sheets.

### 2. Fuentes de datos

| Pipeline | Origen | Path recovery | Schema / keys |
|----------|--------|---------------|---------------|
| **Legacy CSV** | `urlFloraCSV` gid=`730733478` | `data/sheets/flora.csv` | Sin header limpio; columnas por índice: N°, lugar, Lat, Lng, UTM N/E, nombre común/científico, flags Árbol/Palmera/Arbusto, medidas |
| **Gviz “nueva”** | Sheet `1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI` gid=`1270218104` | `data/gviz/flora_gviz_*.json` | `N°`, `Ubicación`, `Referencia`, `Latitud `, `Longitud`, `Nombre común`, `Nombre científico`, `Tipo de vegetación (Formas biológicas)`, `campus pucp Cantidad`, `Foto`, `Código`, `OBSERVACIÓN FEN 2026` |
| **Fotos Drive API** | `URL_MI_API_DRIVE` `/exec` | `data/appscript/drive_fotos_api.json` + `data/drive_fotos/*.jpg` | mapa `codigo → fileId` (70) — **usado sobre todo por bebederos**, no por flora gviz |
| **Edición coords** | `WEB_APP_URL` `/exec` | `data/appscript/flora_edit_api.txt` | GET `?nro=&lat=&lng=` |

**Tipos vegetación (gviz, 1050 filas):** Árbol 652, Palmera 253, Arbusto 83, herbácea 39, Trepadora 22, suculenta 1.  
**Top especies:** Palmera Real (170), Ponciana (135), Palo borracho (71), Cafeto (55), …

### 3. Volumen
- Legacy: **~76** filas útiles / 17 KB  
- Gviz: **1050** puntos / ~354 KB JSON  
- Drive fotos: **70 JPEG** (~8.8 MB) — índice bebederos  
- Edición API: endpoint vivo, sin snapshot de escritura

### 4. Cómo se carga y renderiza
- Legacy: `cargarPuntosFlora()` → PapaParse `header: false` → `capaPalmeras` (comentario Proj4js; **Proj4 no está en CDN**; usa lat/lng del CSV).
- Nueva: `cargarNuevosDatosFlora()` (**definida dos veces**, ~L2875 y ~L3263) → parse gviz `text.substring(47).slice(0,-2)` → `renderizarPuntosFlora` → `capaNuevaFlora`.
- Símbolos: `obtenerSimboloVegetacion(tipo)`.
- Panel: `mostrarPanelLateralFlora`, `crearPanelControlFlora`, `ejecutarFiltroFlora`, `actualizarMetricasPanelFlora`.
- Edición: `guardarCoordenadasEnSheets(nro, lat, lng)` → `WEB_APP_URL`.

### 5. UI / filtros
- Maestro `chk-toda-flora` (“VER ESPECIES”) dispara pipeline **nueva**.
- Legacy aún tiene `chk-palmeras` / `chk-cafetos` (maestro `chk-toda-flora` antiguo comentado ~L2839).
- Panel lateral con filtros de tipo/ubicación y métricas.

### 6. Dependencias cruzadas
- **Cafetos** = sibling legacy del mismo libro Sheets.
- **Vivero** / **monitoreo** comparten dominio Flora pero otras fuentes.
- `transformarEnlaceDrive` redefinida **5 veces** en el monolito.
- Fotos en gviz son links Drive por fila; API Drive indexada es de códigos `bb*` (bebederos).

### 7. Estado de recuperación
**parcial**
- Datos lectura: ok (CSV + gviz + fotos bebederos).
- Escritura flora: **ok_partial** (API exige params; no se replayaron writes).
- Deuda: función duplicada, dos verdades (76 vs 1050).

### 8. Riesgos y recomendaciones React + mapa 3D
- **Elegir una sola fuente de verdad** (gviz 1050) y archivar CSV legacy o migrarlo con ETL.
- Tipar columnas con espacios (`Latitud `) y tildes; índices mágicos `idxFoto = 9` son bomba.
- Clusterizar 1050 puntos (MapLibre + Supercluster / deck.gl); popups con foto vía proxy propio, no `lh3`/`drive` directo.
- Edición: React form + API autenticada (no Apps Script público GET).
- Eliminar duplicados de `cargarNuevosDatosFlora` antes de portar.

---

## Cafetos

### 1. Propósito
Capa específica de **Coffea arabica** / cafetos del campus (subset del inventario botánico).

### 2. Fuentes de datos
| Fuente | Variable | Path |
|--------|----------|------|
| CSV | `urlCafetosCSV` gid=`746966399` (mismo libro que flora legacy) | `data/sheets/cafetos.csv` |

Schema análogo a flora legacy (filas multi-header). Sample: `N°=1`, lugar=`Biblioteca Central`, lat/lng, `Cafeto`, `Coffea arabica`, flag Arbusto=`1`.

### 3. Volumen
- **~76** filas útiles / **~10 KB**  
- Nota: gviz flora ya incluye **55** registros “Cafeto” — posible solape parcial, no identidad 1:1.

### 4. Cómo se carga y renderiza
- `cargarPuntosCafetos()` → PapaParse → markers en `capaCafetos` (declarada **dos veces**).
- Popups HTML con especie/lugar (mismo patrón que palmeras legacy).

### 5. UI / filtros
- `chk-cafetos` vía `vincularCapaPro`.
- No entra en el panel de la flora “nueva”.

### 6. Dependencias cruzadas
- Hermano de **flora legacy**; solapa semántica con gviz (`Cafeto`/`Coffea arabica`).

### 7. Estado de recuperación
**ok** (CSV recuperado).

### 8. Riesgos y recomendaciones React + mapa 3D
- Fusionar cafetos como filtro `tipoVegetacion === '…' || nombre === 'Cafeto'` sobre el inventario único.
- Evitar segunda capa Leaflet paralela en la migración.

---

## Vivero

### 1. Propósito
Registrar **actividades del vivero / zoocriadero / ARAC** (flora y fauna bajo cuidado), georreferenciadas por `Lugar` y filtrables por mes.

### 2. Fuentes de datos
| Fuente | Variable | Path |
|--------|----------|------|
| CSV | `URL_VIVERO_CSV` | `data/sheets/vivero.csv` |

**Columnas:** `Fecha`, `Área`, `Subproceso`, `Etapa`, `Descripción`, `Responsables del reporte`, `Observaciones`, `Lugar`.

**Enums (filas no vacías ~683):**
- `Área`: Fauna 335, Flora 299, Otros/otros ~12, Ambiental 3.
- `Responsables`: Magno 335, Hector 164, Efrain 76, …
- `Lugar`: Bosque húmedo 102, Vivero 29, EEGGCC 16, …

### 3. Volumen
- Archivo **~144 KB** / ~719 líneas crudas; **~683** con contenido.
- Fechas recientes tipo `23/9/2026`.

### 4. Cómo se carga y renderiza
- `cargarVivero()` + PapaParse.
- `poblarDesplegableMeses`, `renderizarMarcadoresVivero` → `capaVivero`.
- Coordenadas: resolución por nombre de `Lugar` (diccionario / matching; acoplado a lugares del monitoreo).

### 5. UI / filtros
- `chk-vivero`.
- Select de mes (rellenado dinámicamente).

### 6. Dependencias cruzadas
- Cruza **Flora** y **Fauna** en el campo `Área`.
- Lugares solapan con `lugares.csv` del monitoreo y con polígonos de jardines.

### 7. Estado de recuperación
**ok**.

### 8. Riesgos y recomendaciones React + mapa 3D
- Muchas filas sin coordenadas propias: geocoding por `Lugar` es frágil (typos “Neeskens”/“Neskeens”).
- Modelo evento (fecha, área, subproceso) + join a `lugares`; heatmap temporal en deck.gl `HexagonLayer` / `HeatmapLayer`.
- Separar dominio “vivero operativo” de “inventario estático flora”.

---

## Fauna

### 1. Propósito
Visualizar puntos del **Plan de Manejo de Fauna** (aves, canes, insectos, etc.) con recomendaciones de conservación.

### 2. Fuentes de datos
| Fuente | Path |
|--------|------|
| GeoJSON | `fauna.geojson` |

**Properties:** `id`, `Animal` (valores: Aves 6, Roedores 4, Insectos 3, Canes 2, Ardillas 2, Gallinazos 1, `Colonia Fe` 1 — truncado “Colonias Felinas”).

### 3. Volumen
- **19** Points / **~3 KB** — capa pequeña, simbólica.

### 4. Cómo se carga y renderiza
- `fetch('fauna.geojson')` (~L3783).
- `capaFaunaGeoJSON` + `obtenerConfiguracionAnimal` → subcapas `capaAves`, `capaCanes`, `capaInsectos`, `capaGallinazos`, `capaArdillas`, `capaRoedores`, `capaColoniasGatos`.
- DivIcons / class `fauna-marker-custom`; popups con texto de recomendación.

### 5. UI / filtros
- Maestro `chk-toda-fauna`.
- `chk-ardillas`, `chk-aves`, `chk-canes`, `chk-insectos`, `chk-gallinazos`, `chk-roedores`, `chk-gatos`.

### 6. Dependencias cruzadas
- **Vivero** reporta Área=Fauna (operativo diario) ≠ estos 19 puntos de plan.
- Independiente del resto espacialmente.

### 7. Estado de recuperación
**ok**.

### 8. Riesgos y recomendaciones React + mapa 3D
- Enriquecer attributes (`especie`, `estado`, `fecha_avistamiento`); hoy solo `Animal`.
- Corregir truncado `Colonia Fe`.
- 3D de bajo valor; priorizar iconografía 2D clara. Posible capa GLTF solo si hay modelos educativos.

---

## Bebederos / agua

### 1. Propósito
Inventario de **bebederos de agua** del campus por tipología/estado (fuente, llenador, nuevo, deterioro, baja), con fotos Drive y panel de totales.

### 2. Fuentes de datos

| Archivo | Features | Estado implícito en carga |
|---------|----------|---------------------------|
| `bebedero Tipo fuente.geojson` | 40 | Operativo fuente |
| `bebedero Tipo llenador de botella.geojson` | 10 | Llenador |
| `bebedero Nuevo.geojson` | 8 | Proyecto nuevo |
| `bebedero por deterioro.geojson` | 7 | Renovación |
| `bebedero por baja del equipo.geojson` | 2 | Baja |
| Apps Script fotos | 70 ids | `data/appscript/drive_fotos_api.json` |
| JPEGs | 70 | `data/drive_fotos/` |

**Campos útiles en properties (export KML-like):**  
`Name` (ej. `PT_bb1`), `COL5BBF168` (código), `COL5BBF1_1`/`_2` (lat/lng), `COL5BBF1_3` (ubicación texto), `COL5BBF1_4` (estado: operativo/nuevo/remodelación), `COL5BBF1_5` (ámbito: CAMPUS/AULARIO/…).  
El render principal usa **`Name`** + geometría Point; el resto queda en el archivo.

### 3. Volumen
- **67** puntos totales (40+10+8+7+2)
- ~75 KB GeoJSON agregados + ~8.8 MB fotos

### 4. Cómo se carga y renderiza
- `inicializarFotosDrive` → fetch `URL_MI_API_DRIVE` → `cacheImagenesDrive`.
- `cargarCapaBebedero(archivo, capa, color, símbolo, …)` × 5.
- MarkerCluster: `clusterCentralAgua`.
- Click → `abrirSidebarAgua`; totales → `actualizarPanelResumenAgua` + `#panel-resumen-agua`.
- Nota: deterioro+baja se cargan ambos en `capaBebedRenovacion` / checkbox `chk-bebed-renovacion`; existen también `capaBebedDeterioro`/`capaBebedBaja` y checks `chk-bebed-deterioro`/`chk-bebed-baja` en otro bloque → **inconsistencia de IDs** entre HTML (`chk-bebed-renovacion`) y `vincularCapaPro` tardío.

### 5. UI / filtros
- Maestro `chk-todo-agua`.
- `chk-bebed-fuente`, `chk-bebed-llenador`, `chk-bebed-nuevo`, `chk-bebed-renovacion`.
- Panel flotante de resumen de conteos.

### 6. Dependencias cruzadas
- Fotos Drive compartidas conceptualmente con flora (misma infra Apps Script distinta URL).
- Independiente de riego de áreas verdes (otro “agua”).

### 7. Estado de recuperación
**ok** (GeoJSON + API + 70/70 fotos).

### 8. Riesgos y recomendaciones React + mapa 3D
- Limpiar schema: renombrar `COL5BBF*` → `codigo`, `estado`, `ubicacion`, `ambito`.
- Unificar checkboxes renovación (un solo contrato UI).
- Hostear thumbnails en CDN propio; `lh3.googleusercontent.com` es frágil.
- Cluster nativo MapLibre; en 3D, pins con billboard + panel detalle React.

---

## Monitoreo / actividades / poda / heatmap 2026

### 1. Propósito
Seguimiento de **actividades de jardinería 2026** (mantenimiento, poda, riego, habilitación…), catálogo de tipos, lugares georreferenciados, registros de poda e **mapa de calor** de intensidad.

### 2. Fuentes de datos
Libro base `BASE_URL` = `…2PACX-1vTakay8F…/pub?output=csv`:

| Variable | gid | Path | Filas |
|----------|-----|------|-------|
| `URL_LUGARES_CSV` | 216669177 | `lugares.csv` | ~76 | lugar, latitud, longitud (coma decimal) |
| `URL_ACTIVIDADES_CSV` | 344271829 | `actividades.csv` | ~45 | Clase, Tipo de actividad, Descripción (+ columnas duplicadas) |
| `URL_PODA_CSV` | 1845362857 | `poda.csv` | ~25 | ID `PO-1`, Tipo, fechas, Personal, Ubicación, especie… |
| `URL_2026_CSV` | 530107837 | `monitoreo_2026.csv` | **283** | Clase, Estado, fechas, Mes, lugar, lat/lng, Actividad, comentario, Foto, Responsable… |
| (base sin gid) | — | `monitoreo_base.csv` | **idéntico** a 2026 |

**Clases top 2026:** Mantenimiento de jardines 59, Propagación 38, Poda 22, Habilitación 18, …  
**Responsables:** responsable 69, responsable 56, responsable 48.

### 3. Volumen
- ~283 eventos + catálogos pequeños; CSV ~52 KB el principal.

### 4. Cómo se carga y renderiza
- `inicializarDescargaExcel()` / `parsearCSV` / `cargarDatosEnMapa` / `diagnosticarYRenderizar`.
- Marcadores: capa asociada a `chk-actividades`.
- Calor: `capaCalorActividades = L.heatLayer([], {…})` + `setLatLngs(puntosCalor)`.
- Helpers: `obtenerCoordenadas`, `poblarFiltrosClaseYTipo`, `poblarFiltroResponsables`, `actualizarSelectTipos`, `vincularEscuchadoresCheckboxes`.
- Control Leaflet custom inyecta checkboxes de actividades/calor (~L253).

### 5. UI / filtros
- `chk-actividades`, `chk-calor-actividades`.
- Selects de clase, tipo, mes, responsable (IDs vía `obtenerElementosSelect`).

### 6. Dependencias cruzadas
- Responsables ↔ **jefes de grupo** (mismo personal, sin FK).
- Lugares ↔ **vivero** / jardines.
- Fotos: a veces Wikimedia placeholder en Sheet.

### 7. Estado de recuperación
**ok** (4 CSV + base duplicada).

### 8. Riesgos y recomendaciones React + mapa 3D
- Unificar `monitoreo_base` ≡ `monitoreo_2026` (no versionar dos archivos idénticos).
- Decimal con coma → normalizar en ETL a `.`.
- Heatmap: deck.gl `HeatmapLayer` o MapLibre densify; filtrar por mes en React Query.
- Poda como subdominio tipado (`PO-*`) enlazable a flora por nombre científico.

---

## Reservas de jardines

### 1. Propósito
Mostrar **polígonos reservables** y su **ocupación semanal** (mes/semana/día) según agenda en Google Sheet.

### 2. Fuentes de datos
| Fuente | Estado | Path / ID |
|--------|--------|-----------|
| GeoJSON local | **ok** | `jardines_reserva.geojson` — 21 MultiPolygons |
| Sheet reservas | **failed 401** | `SHEET_ID_RESERVAS=1R3Xz8A5xIVm-s0duMQhav2YAJrQdlSAYgKYeoEvoT8s` gid=`1907703639`; pub `2PACX-1vSqBX3fI…` → `data/sheets/reservas_jardines.FAILED.txt` |

**GeoJSON properties:** `Nombre`, `Uso`, `Proy riego`, `Riego act`, `código`, `Pertenecen` (`DAF` 15 / `Unidades` 6), etc.

### 3. Volumen
- 21 polígonos / ~48 KB.
- Filas de agenda: **0 recuperadas** (401).

### 4. Cómo se carga y renderiza
- `cargarJardinesReserva()` → `fetch('jardines_reserva.geojson')` + `inicializarDescargaReservas()` (CSV fetch).
- Match nombre Sheet↔shape: `coincidenJardines`, `obtenerNombreDesdeShape`.
- Pintado semanal: `procesarReservasSemanales` → `capaReservasJardines`.
- `capaReservasJardines` redeclarada varias veces (riesgo de shadowing).

### 5. UI / filtros
- `chk-reservas-jardines`.
- `#reserva-mes`, `#reserva-semana`, `#reserva-dia`.

### 6. Dependencias cruzadas
- Geometría subset de áreas verdes / jardines DAF.
- Sin Sheet, la capa solo muestra polígonos “vacíos” de estado.

### 7. Estado de recuperación
**parcial / fallido en agenda** — shapes ok; temporalidad **rota**.

### 8. Riesgos y recomendaciones React + mapa 3D
- **Prioridad #1 de datos:** republicar Sheet o exportar snapshot CSV al repo.
- Modelo: `GardenPlot` + `Reservation(slot)`; UI calendario React (no 3 selects frágiles).
- Colorear fill por estado (libre/reservado/conflicto) en MapLibre feature-state.
- No hardcodear dos URLs distintas (SHEET_ID vs 2PACX) sin test de salud.

---

## Adicionales campus (puertas, estacionamientos, xerofítica)

### 1. Propósito
Capas de contexto: **accesos**, **playas de estacionamiento** y **zonas xerofíticas** (bajo riego).

### 2. Fuentes de datos

| Archivo | n | Keys reales | Problema |
|---------|---|-------------|----------|
| `puertas_entradas.geojson` | 7 Points | solo `id` | Popup espera `Nombre`/`Tipo`/`Control` → “Sin nombre” |
| `playas_de_estacionamiento.geojson` | 15 MultiPolygon | solo `id` | Popup espera `Nombre`/`Capacidad`/`Zona` |
| `xerofitica.geojson` | 10 MultiPolygon | `Nombre` (null), `Área`, `Riego act`, `Class=Ornato`, `layer`, `path` | Export shapefile crudo |

### 3. Volumen
- ~1 KB + 9 KB + 27 KB.

### 4. Cómo se carga y renderiza
- `fetch` → `capaPuertasEntradas` (emoji 🚪), `capaPlayasEstacionamiento` (fill gris), `capaXerofitica` (naranja) + `vincularPopupXerofitica`.

### 5. UI / filtros
- Maestro `chk-todo-adicionales`.
- `chk-puertas`, `chk-estacionamientos`, `chk-xerofitica` (+ áreas uso/riego en el mismo bloque UI).

### 6. Dependencias cruzadas
- Xerofítica ⊂ lógica de áreas verdes / riego.
- Puertas útiles al **buscador PUCP** como POIs, hoy no enlazados.

### 7. Estado de recuperación
**ok** archivos; **datos atributos incompletos** (parcial funcional).

### 8. Riesgos y recomendaciones React + mapa 3D
- Completar attributes o regenerar GeoJSON desde fuente GIS.
- Extrusión baja de estacionamientos; puertas como POI 3D simples.
- Quitar props basura `path`/`layer` del shapefile.

---

## Veredas de alto tránsito

### 1. Propósito
Resaltar zonas de **vereda de alto riesgo / alto tránsito peatonal** para mantenimiento y seguridad.

### 2. Fuentes de datos
- `area_vereda_peligro.geojson` — **1** MultiPolygon, property `id: 2`, **~81 KB** (geometría densa).

### 3. Volumen
1 feature / 81 KB.

### 4. Cómo se carga y renderiza
- `capaVeredasPeligro`, carga async `fetch('area_vereda_peligro.geojson')` (~L4382+).
- Estilo alerta (capa dedicada).

### 5. UI / filtros
- `chk-veredas-peligro`.

### 6. Dependencias cruzadas
- Independiente; podría cruzar con **barredoras** (rutas de limpieza) y áreas verdes.

### 7. Estado de recuperación
**ok**.

### 8. Riesgos y recomendaciones React + mapa 3D
- Simplificar geometría (Douglas-Peucker) — 81 KB para 1 polígono es excesivo en web.
- Feature-state “peligro” con outline animado; en 3D, cinta sobre DEM/campus.

---

## Barredoras 1 y 2

### 1. Propósito
**Simulación animada** de recorrido de barredoras mecánicas (capacitación / visualización de cobertura), no telemetría GPS real.

### 2. Fuentes de datos
- **Hardcode** en `script.js`: `rutaBarredora1` (~25 waypoints), `rutaBarredora2` (~15 waypoints).
- Turf: `lineString`, `length`, `along`, `lineSliceAlong`, `buffer` (franja ~6 m).

### 3. Volumen
- Solo coordenadas en JS; sin CSV/GeoJSON externo.

### 4. Cómo se carga y renderiza
- `mostrarSimulacionBarredora` / `mostrarSimulacionBarredora2`.
- Polyline + marker custom + `setInterval` animación.
- Buffer visual del área ya “barrida”.
- Paneles `#panel-control-barredora` / `-2`, botones iniciar/reiniciar/cerrar.

### 5. UI / filtros
- `chk-barredora-1`, `chk-barredora-2`.

### 6. Dependencias cruzadas
- Podría alinearse con **veredas** y **tachos**; hoy no hay join.

### 7. Estado de recuperación
**ok** (lógica 100% local).

### 8. Riesgos y recomendaciones React + mapa 3D
- Extraer rutas a GeoJSON/GPX versionado.
- Sustituir `setInterval` por requestAnimationFrame + clock React.
- En MapLibre: `line` layer + animated point; buffer con Turf en worker.
- Si en el futuro hay GPS real: capa distinta “telemetría” vs “simulación”.

---

## Buscador / puntos PUCP

### 1. Propósito
**Búsqueda tipo Google Maps** de lugares del campus (facultades, servicios) con sugerencias y ficha foto/enlace.

### 2. Fuentes de datos
| Fuente | Variable | Path |
|--------|----------|------|
| CSV | `urlPuntosPUCPCSV` gid=`1330096436` | `data/sheets/puntos_pucp.csv` |

**Columnas:** `location/lat`, `location/lng`, `phone`, `placeId`, `searchPageUrl`, `searchString`, `title`, `url`, `website`, `image` (export estilo scraper Maps).

### 3. Volumen
- **~153** lugares / **~52 KB**.

### 4. Cómo se carga y renderiza
- PapaParse al inicio → `bancoDatosPUCP` (no pinta todo).
- `buscarYLimpiarMapa(texto)` filtra y crea markers en `capaPuntosPUCP`.
- Control de búsqueda custom (`onAdd`).
- Fotos: `transformarEnlaceDrive`.
- `capaPuntosPUCP` declarada **múltiples veces** (shadowing).

### 5. UI / filtros
- Input de búsqueda (control Leaflet).
- `chk-fotos-pucp` vincula la capa.

### 6. Dependencias cruzadas
- POIs genéricos; no unificados con puertas/bebederos/lugares de monitoreo (`lugares.csv` es otro set).

### 7. Estado de recuperación
**ok**.

### 8. Riesgos y recomendaciones React + mapa 3D
- Tres catálogos de lugares (PUCP scrape, monitoreo, vivero) → **gazetteer único**.
- Buscador React con Fuse.js / Elastic; geocoder propio.
- Cumplimiento ToS: el CSV parece scrape de Google Maps (`placeId`, `searchPageUrl`) — riesgo legal/ToS; regenerar desde datos institucionales.

---

## Mapas base (Google / OSM / Esri)

### 1. Propósito
Fondos cartográficos conmutables para lectura satelital o callejero del campus.

### 2. Fuentes de datos
| Variable | URL tiles | maxNativeZoom |
|----------|-----------|---------------|
| `googleSatelite` (default) | `mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}` | 20 (estirado a 22) |
| `mapaNormal` | `{s}.tile.openstreetmap.org/...` | 19 |
| `esriSatelite` | `server.arcgisonline.com/.../World_Imagery/...` | 18 |

Sin archivos locales de tiles.

### 3. Volumen
N/A (streaming).

### 4. Cómo se carga y renderiza
- `L.map('map', { maxZoom: 22 }).setView([-12.068, -77.08], 15)`.
- `cambiarMapaBase(idRadio, capaBase, nombre)`.
- Radios: `rb-satelite`, `rb-esri`, `rb-clasico`.

### 5. UI / filtros
- Grupo `name="mapa-base"` en `#panel-capas-derecho`.

### 6. Dependencias cruzadas
- Base de **todas** las capas vectoriales.
- Superzoom 22 asume estirado digital (artefactos).

### 7. Estado de recuperación
**ok** (código); dependencia runtime de terceros.

### 8. Riesgos y recomendaciones React + mapa 3D
- **Google raster tiles sin API key** viola ToS habituales → migrar a MapLibre + estilo propio / MapTiler / Esri con licencia / mosaico campus.
- Para 3D: MapLibre `globe` o terrain DEM + posible mesh del campus; satélite como raster source con attribution correcta.
- OSM: respetar usage policy (User-Agent, caching).

---

## Hallazgos transversales (migración)

1. **Monolito global:** ~132 bindings top-level, funciones duplicadas (`cargarNuevosDatosFlora`×2, `transformarEnlaceDrive`×5, capas redeclaradas).
2. **Doble inventario flora** (76 CSV vs 1050 gviz) + cafetos paralelos.
3. **Reservas rotas** por Sheet 401 — único fallo duro de datos de negocio.
4. **Zoning triple:** jefes (`jefes`), supervisores tachos (`ZONA` 1–4), lugares texto libre.
5. **Atributos huecos** en puertas/estacionamientos (popup engañoso).
6. **Fotos:** 70 OK offline; flora gviz apunta a otros fileIds Drive no descargados en lote.
7. **Barredoras** = demo Turf, no IoT.
8. **Puntos PUCP** con olor a scrape Maps.

### Complejidad relativa para React + mapa 3D

| Rank | Dominio | Por qué |
|------|---------|---------|
| 1 | **Flora** | Doble pipeline, edición, fotos, 1050 pts, UI panel |
| 2 | **Monitoreo + heatmap** | 4 sheets, filtros cruzados, heat, fechas locales |
| 3 | **Tachos + zonas** | Multi-fracción, sidebar, point-in-polygon, resumen |

Candidatos “quick wins”: fauna, veredas, xerofítica/puertas (tras enriquecer attrs), barredoras (rutas a GeoJSON).

---

## Referencias de archivos en este recovery

- Código: `index.html`, `style.css`, `script.js`
- Docs: `ANALISIS.md`, `INVENTARIO-DATOS.md`, **este archivo**
- Datos: GeoJSON/JSON en raíz; `data/sheets/`, `data/gviz/`, `data/appscript/`, `data/drive_fotos/`
