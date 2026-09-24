# Inventario de datos — Mapa de Control Ambiental Campus

Estado al momento de la recuperación (2026-09-24, America/Lima).  
Origen vivo: https://mapa-web-6.vercel.app/

Leyenda **status:** `ok` | `failed` | `ok_partial` | `local_only`

---

## A. Estáticos locales (también en Vercel)

| Nombre | Origen | Path local | Tamaño / registros | Status |
|--------|--------|------------|-------------------|--------|
| index.html | https://mapa-web-6.vercel.app/ | `index.html` | ~35 KB | ok |
| style.css | …/style.css | `style.css` | ~12 KB | ok |
| script.js | …/script.js | `script.js` | ~261 KB / 4921 líneas | ok |
| areas_verdes.geojson | …/areas_verdes.geojson | `areas_verdes.geojson` | ~984 KB / 521 features | ok |
| jefe_de_grupo.json | …/jefe_de_grupo.json | `jefe_de_grupo.json` | ~1002 KB / 534 features | ok |
| jardines_reserva.geojson | …/jardines_reserva.geojson | `jardines_reserva.geojson` | ~48 KB / 21 features | ok |
| fauna.geojson | …/fauna.geojson | `fauna.geojson` | ~3 KB / 19 features | ok |
| xerofitica.geojson | …/xerofitica.geojson | `xerofitica.geojson` | ~27 KB / 10 features | ok |
| puertas_entradas.geojson | …/puertas_entradas.geojson | `puertas_entradas.geojson` | ~1 KB / 7 features | ok |
| playas_de_estacionamiento.geojson | …/playas_de_estacionamiento.geojson | `playas_de_estacionamiento.geojson` | ~9 KB / 15 features | ok |
| area_vereda_peligro.geojson | …/area_vereda_peligro.geojson | `area_vereda_peligro.geojson` | ~81 KB / 1 MultiPolygon | ok |
| bebedero Tipo fuente.geojson | … | `bebedero Tipo fuente.geojson` | ~43 KB / 40 pts | ok |
| bebedero Tipo llenador de botella.geojson | … | `bebedero Tipo llenador de botella.geojson` | ~11 KB / 10 pts | ok |
| bebedero Nuevo.geojson | … | `bebedero Nuevo.geojson` | ~9 KB / 8 pts | ok |
| bebedero por deterioro.geojson | … | `bebedero por deterioro.geojson` | ~8 KB / 7 pts | ok |
| bebedero por baja del equipo.geojson | … | `bebedero por baja del equipo.geojson` | ~2 KB / 2 pts | ok |
| supervisores (inline) | embebido en `script.js` como `geojsonSupervisores` | (dentro de script.js) | ~39 KB / 4 features | ok |

No se encontraron en Vercel: `favicon.ico`, `icons/`, `logo.png`, `supervisoress.geojson` como archivo suelto (está inline).

---

## B. Google Sheets — CSV publicados

| Nombre | Variable en script.js | Origen URL | Path local | Tamaño / filas | Status |
|--------|----------------------|------------|------------|----------------|--------|
| puntos_pucp.csv | `urlPuntosPUCPCSV` | `…2PACX-1vTb8rgvl…/pub?gid=1330096436&output=csv` | `data/sheets/puntos_pucp.csv` | ~52 KB / 154 líneas | ok |
| tachos.csv | `urlPuntosEcologicosCSV` | `…2PACX-1vQK3aVBw…/pub?gid=657037007&output=csv` | `data/sheets/tachos.csv` | ~36 KB / 186 líneas | ok |
| flora.csv | `urlFloraCSV` | `…2PACX-1vTp1v1vt…/pub?gid=730733478&output=csv` | `data/sheets/flora.csv` | ~17 KB / 88 líneas | ok |
| cafetos.csv | `urlCafetosCSV` | `…2PACX-1vTp1v1vt…/pub?gid=746966399&output=csv` | `data/sheets/cafetos.csv` | ~10 KB / 88 líneas | ok |
| monitoreo_base.csv | `BASE_URL` (sin gid) | `…2PACX-1vTakay8F…/pub?output=csv` | `data/sheets/monitoreo_base.csv` | ~52 KB / 292 líneas | ok |
| lugares.csv | `URL_LUGARES_CSV` (`BASE_URL&gid=216669177`) | mismo libro + gid | `data/sheets/lugares.csv` | ~3 KB / 77 líneas | ok |
| actividades.csv | `URL_ACTIVIDADES_CSV` (gid=344271829) | mismo libro | `data/sheets/actividades.csv` | ~5 KB / 46 líneas | ok |
| poda.csv | `URL_PODA_CSV` (gid=1845362857) | mismo libro | `data/sheets/poda.csv` | ~7 KB / 27 líneas | ok |
| monitoreo_2026.csv | `URL_2026_CSV` (gid=530107837) | mismo libro | `data/sheets/monitoreo_2026.csv` | ~52 KB / 292 líneas | ok (idéntico a `monitoreo_base.csv`) |
| vivero.csv | `URL_VIVERO_CSV` | `…2PACX-1vRrEaaf7…/pub?gid=1814328638&output=csv` | `data/sheets/vivero.csv` | ~144 KB / 1070 líneas | ok |
| reservas_jardines.csv | `inicializarDescargaReservas` + `SHEET_ID_RESERVAS=1R3Xz8A5xIVm-s0duMQhav2YAJrQdlSAYgKYeoEvoT8s` | `…2PACX-1vSqBX3fI…/pub?gid=1907703639` y gviz/export del ID | `data/sheets/reservas_jardines.FAILED.txt` | — | **failed HTTP 401** (no público) |

---

## C. Google Sheets — gviz JSON

| Nombre | Origen | Path local | Tamaño / registros | Status |
|--------|--------|------------|-------------------|--------|
| flora_gviz (default) | `https://docs.google.com/spreadsheets/d/1QQxHKnefZ94Nk2raTwOe0N_zxBlb3AdQC_MP3ZzfRoI/gviz/tq?tqx=out:json` | `data/gviz/flora_gviz_default.json` | ~354 KB / 1050 filas | ok |
| flora_gviz gid 1270218104 | misma hoja `&gid=1270218104` | `data/gviz/flora_gviz_gid1270218104.json` | ~354 KB / 1050 filas | ok (idéntico al default) |

---

## D. Google Apps Script `/exec`

| Nombre | Variable | URL | Path local | Notas | Status |
|--------|----------|-----|------------|-------|--------|
| drive_fotos_api.json | `URL_MI_API_DRIVE` | `https://script.google.com/macros/s/AKfycbyul5ascYRZ1L5KpXECCnOzQC-SGS8iTHCBRkAgQfDvy4vAEJZiK0JI115PY4P2yyuQ/exec` | `data/appscript/drive_fotos_api.json` | JSON mapa código→Drive fileId, 70 claves | ok |
| flora_edit_api.txt | `WEB_APP_URL` | `https://script.google.com/macros/s/AKfycbxGSndXYFuJj3NxpW7TLz3USQ4EgNqWkK6NctE2oKYXRE3Q8XOZ5f6mQCgOc7G1hPj0/exec` | `data/appscript/flora_edit_api.txt` | Respuesta GET sin params: `Faltan parámetros: nro, lat o lng` (34 B) | ok_partial (endpoint vivo; requiere params de escritura) |

---

## E. Google Drive — fotos (vía API + lh3)

| Nombre | Origen | Path local | Tamaño | Status |
|--------|--------|------------|--------|--------|
| 70 fotos JPEG (s800) | IDs de `drive_fotos_api.json` → `https://lh3.googleusercontent.com/d/{id}=s800` | `data/drive_fotos/{codigo}_{id}.jpg` + `_index.json` | ~8.8 MB total / 70 ok, 0 fail | ok |

Índice: `data/drive_fotos/_index.json` (mapeo key → id → status → bytes).

---

## F. CDN / tiles (no descargados como datos de dominio)

- Leaflet 1.9.4, MarkerCluster 1.5.3, leaflet.heat 0.2.0, Turf 6, PapaParse 5.3.2
- Tiles: Google VT, OSM, Esri World Imagery

---

## Resumen de recuperación externa

| Resultado | Cantidad |
|-----------|----------|
| Sheets CSV ok | 10 archivos |
| Sheets CSV failed (401) | 1 (reservas) |
| gviz JSON ok | 2 (duplicados entre sí) |
| Apps Script ok / parcial | 1 ok + 1 parcial |
| Drive fotos ok | 70 / 70 |
| Estáticos Vercel ya locales | 16 archivos de app/datos |
