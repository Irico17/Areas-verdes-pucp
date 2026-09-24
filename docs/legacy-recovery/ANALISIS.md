# Análisis técnico — Mapa de Control Ambiental Campus

**Fuente viva:** https://mapa-web-6.vercel.app/  
**Recuperación local:** `/workspace/mapa-web-6-recovery/`  
**Stack:** HTML/CSS/JS vanilla + Leaflet 1.9.4 + MarkerCluster 1.5.3 + leaflet.heat 0.2.0 + Turf 6 + PapaParse 5.3.2  
**Tamaño JS:** ~261 KB / ~4 921 líneas en un único `script.js`

---

## 1. Arquitectura general

```
index.html  →  estructura DOM (mapa #map, paneles laterales, ~54 checkboxes de capas)
style.css   →  layout de paneles, sidebars, controles de capas y barredoras
script.js   →  TODA la lógica: mapa, capas, fetch/PapaParse, UI, animaciones
     │
     ├─ GeoJSON/JSON locales (fetch relativo)
     ├─ Google Sheets publicados (CSV vía PapaParse / gviz JSON)
     └─ Google Apps Script /exec (fotos Drive + edición flora)
```

- **Sin bundler, sin módulos ES, sin framework.** Un solo archivo monolítico con secciones comentadas (`// =====`).
- **Leaflet** monta el mapa sobre `#map` centrado en campus PUCP (`[-12.068, -77.08]`, zoom 15, `maxZoom: 22`).
- **Tres bases cartográficas:** Google Satellite (`mt{s}.google.com`), OSM, Esri World Imagery — conmutadas por radio buttons.
- La UI es un **panel de capas derecho** (checkboxes maestros + submenús) + **sidebars/popups** para detalle (agua, tachos, flora, barredoras, búsqueda PUCP).

---

## 2. Módulos / features principales en `script.js`

| Área | Qué hace | Datos |
|------|----------|-------|
| **Init mapa** | `L.map`, tile layers, `maxZoom: 22` | — |
| **Puntos PUCP / búsqueda** | Banco en memoria desde CSV; buscador con sugerencias; render dinámico | Sheets `puntos_pucp.csv` |
| **Jefes de grupo / áreas** | Capas por supervisor (responsable, responsable, responsable, Campos Depo, Bosque Húmedo) | `jefe_de_grupo.json` |
| **Zonas supervisores** | GeoJSON **embebido** en JS (`geojsonSupervisores`, 4 features) para zonificar tachos | inline |
| **Fauna** | Capas por tipo (aves, canes, insectos, gallinazos, ardillas, roedores, gatos) | `fauna.geojson` |
| **Tachos ecológicos** | Capas por fracción (no aprovechables, papel, plástico, vidrio, pilas, peligrosos, RAEE, metales, aniquem, intermedios); filtro por zona; sidebar resumen | Sheets `tachos.csv` |
| **Flora (CSV legacy)** | Palmeras/cafetos desde CSV; popups ricos; comentario menciona Proj4js pero **no se carga la lib** (usa lat/lng ya en CSV) | `flora.csv`, `cafetos.csv` |
| **Flora (gviz “nueva”)** | `cargarNuevosDatosFlora` (definida **dos veces**); panel lateral, filtros, métricas, edición | gviz Sheet `1QQx…RoI` |
| **Monitoreo / actividades** | Lugares, actividades, poda, registros 2026; heat map `L.heatLayer`; filtros clase/tipo/mes/responsable | Sheets monitoreo (`lugares`, `actividades`, `poda`, `monitoreo_2026`) |
| **Vivero** | Agrupa por lugar, filtro por mes | `vivero.csv` |
| **Reservas de jardines** | Polígonos + estado semanal desde Sheet; UI mes/semana/día | `jardines_reserva.geojson` + Sheet reservas (**401 privado**) |
| **Agua / bebederos** | MarkerCluster central; 5 GeoJSON por tipo; sidebar métricas | GeoJSON bebederos locales |
| **Áreas verdes / riego** | Estilos por uso de suelo y riego; popups | `areas_verdes.geojson` |
| **Adicionales campus** | Puertas, playas estacionamiento, xerofítica | GeoJSON locales |
| **Veredas peligro** | Capa MultiPolygon | `area_vereda_peligro.geojson` |
| **Barredoras 1 y 2** | Animación/simulación de recorrido con paneles de control | lógica en JS (rutas en código) |
| **Fotos Drive** | Apps Script devuelve mapa código→fileId; thumbnails `lh3.googleusercontent.com` | `/exec` fotos |
| **Edición flora** | POST/GET a segundo Apps Script con `nro, lat, lng` | `/exec` edición |

---

## 3. Modelo de datos

### Local (repo / Vercel estáticos)
- `areas_verdes.geojson` (521 features), `jefe_de_grupo.json` (534), `jardines_reserva.geojson` (21), bebederos (5 archivos), `fauna`, `xerofitica`, `puertas_entradas`, `playas_de_estacionamiento`, `area_vereda_peligro`
- GeoJSON supervisores **inline** en `script.js` (~39 KB)

### Google Sheets (CSV publicados `2PACX-…/pub?output=csv`)
- Puntos PUCP, tachos, flora, cafetos, vivero, monitoreo (4 pestañas vía `gid`)
- **Reservas:** publicado pero responde **401** (ya no es público)

### Google Sheets gviz
- Flora “nueva”: `SHEET_ID=1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI` (+ `gid=1270218104`) — ~1050 filas

### Google Apps Script
- Fotos Drive: público, JSON `{codigo: fileId}` (70 entradas) — **fotos descargadas**
- Edición flora: público pero exige parámetros (`Faltan parámetros: nro, lat o lng`)

### Drive file IDs
- 70 IDs públicos vía `lh3.googleusercontent.com/d/{id}=s800` (recuperados en `data/drive_fotos/`)

---

## 4. Estado global y acoplamiento

- **~132** declaraciones top-level (`var`/`let`/`const`) en un solo scope global.
- Capas Leaflet (`capa*`, `clusterCentralAgua`, `capaCalorActividades`) y bancos (`bancoDatosPUCP`, `diccionarioLugares`, `datosPodaGuardados`, etc.) viven como globals.
- **~68–91 IDs DOM** acoplados por `getElementById` / atributos HTML (`chk-*`, `panel-*`, `btn-*`, `select*`).
- Checkboxes “maestro” (`chk-toda-flora`, `chk-todo-agua`, …) orquestan subcapas con listeners imperativos.
- `window.map` se asigna defensivamente en un `DOMContentLoaded` tardío (hay lógica que asume `map` ya creado al parsear el script).
- **Deuda visible:** `cargarNuevosDatosFlora` duplicada; `capaPuntosPUCP` / bebederos redeclarados; comentario Proj4js sin CDN; Sheet reservas referenciado por `SHEET_ID_RESERVAS` y por URL `2PACX` distinta.

---

## 5. Dependencias externas (CDN)

| Lib | Versión | Uso |
|-----|---------|-----|
| Leaflet | 1.9.4 (unpkg) | Mapa, capas, popups |
| leaflet.markercluster | 1.5.3 | Bebederos |
| leaflet.heat | 0.2.0 | Heatmap actividades |
| @turf/turf | 6 (jsDelivr) | Geo espacial puntual |
| PapaParse | 5.3.2 (cdnjs) | CSV remoto |

**No incluidas pero mencionadas:** Proj4js (UTM).  
**Tiles:** Google Maps VT, OSM, Esri — dependencia de terceros y ToS de Google tiles.

---

## 6. Fortalezas

- Cubre un dominio operativo amplio (residuos, flora, fauna, agua, riego, vivero, reservas, monitoreo) en una sola UI.
- Datos operativos en Sheets permiten edición sin redeploy.
- MarkerCluster + heat + GeoJSON mixto es adecuado al volumen actual.
- Recuperación casi completa: todos los estáticos Vercel + la mayoría de Sheets/APIs públicas.

## 7. Deuda técnica

- Monolito de ~5k líneas sin módulos; difícil testear y razonar.
- Globals + IDs DOM → alto acoplamiento UI/datos.
- Doble pipeline de flora (CSV + gviz) y funciones duplicadas.
- Auth/publicación frágil: reservas 401; Apps Script y Sheets pueden cerrarse sin aviso.
- Google raster tiles y Drive thumbnails fuera de control del proyecto.
- Sin `package.json`, sin tests, sin tipado, sin CI.
- Mezcla de encoding/columnas frágiles en CSV sin header (`flora.csv` / `cafetos.csv`).
- Animaciones de barredoras y HTML de popups embebidos como strings gigantes.

---

## 8. Modernización React + mapa 3D (solo recomendaciones)

**No reescribir aún.** Dirección sugerida:

1. **Mapa 2D/2.5D:** MapLibre GL JS + `react-map-gl` (o Mapbox GL si hay presupuesto). Sustituye Leaflet; mejor rendimiento con muchos puntos.
2. **Capas analíticas / extrusiones:** `deck.gl` (GeoJsonLayer, HeatmapLayer, IconLayer) sobre MapLibre.
3. **3D campus (opcional):** `three.js` / `react-three-fiber` o edificios con MapLibre `fill-extrusion` si hay alturas; glTF del campus solo si existe modelo.
4. **Datos:**
   - Congelar snapshots versionados (GeoJSON/Parquet) en el repo o object storage.
   - API propia (o Edge Functions) que lea Sheets/Drive con service account — no exponer `/exec` públicos ni CSV publicados.
   - Capas tipadas (Zod) + React Query; una fuente de verdad por dominio (flora, tachos, …).
5. **UI:** React + panel de capas como estado (Zustand/Redux); eliminar IDs globales.
6. **Migración incremental:** un feature flag por dominio (p.ej. solo bebederos en React primero) manteniendo el monolito hasta paridad.

---

## 9. Archivos de esta recuperación

- `INVENTARIO-DATOS.md` — cada fuente, URL, path, estado
- `data/sheets/` — CSV recuperados
- `data/gviz/` — JSON flora
- `data/appscript/` — respuestas `/exec`
- `data/drive_fotos/` — 70 JPEG + `_index.json`
