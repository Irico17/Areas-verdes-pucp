# Plan de migración por fases — Mapa de Control Ambiental Campus

**Objetivo:** pasar del monolito Leaflet (HTML/CSS/`script.js`) a una app **React** con mapa moderno y opción **3D**, sin perder operación del campus.

**Principios**
1. **Datos primero**, UI después.
2. **Paridad por dominio** (feature flags): el sitio viejo puede convivir hasta que cada capa esté lista.
3. **Una fuente de verdad tipada** por dominio (nada de CSV + gviz duplicados).
4. **No reescribir el monolito de golpe** (~5k líneas, ~132 globals).
5. **3D solo cuando el 2D tipado y versionado esté estable.**

**Stack objetivo (recomendado)**
| Capa | Tecnología |
|------|------------|
| App | React + TypeScript + Vite |
| Estado UI capas | Zustand (o similar) |
| Datos remotos | TanStack Query |
| Validación | Zod |
| Mapa 2D/2.5D | MapLibre GL + `react-map-gl` |
| Capas densas / heat / extrusiones | deck.gl |
| 3D opcional | MapLibre `fill-extrusion` / terrain; R3F solo si hay glTF del campus |
| Estilos | CSS modules o Tailwind |
| Hosting | Vercel (cuenta nueva) + repo Git nuevo |

---

## Fase 0 — Congelar la recuperación (1–2 días)

**Meta:** que nada de lo recuperado se pierda si Sheets/Drive se cierran mañana.

| Entrega | Detalle |
|---------|---------|
| Repo Git nuevo | Subir `/workspace/mapa-web-6-recovery` (sin secretos; URLs públicas ya están en el JS) |
| Snapshot versionado | Tag `recovery-2026-09-24` |
| Inventario firmado | `INVENTARIO-DATOS.md` + checksums de GeoJSON/CSV/fotos |
| Decisión de ownership | Quién mantiene Sheets vs quién edita en la app nueva |

**Criterio de salida:** el tarball + docs viven en Git; se puede servir el legado con `npx serve` offline (salvo Sheet reservas 401).

**Riesgo inmediato a mitigar:** copiar flora gviz completa y fotos; documentar que **reservas agenda** está rota.

---

## Fase 1 — Capa de datos (ETL + schema) (3–5 días)

**Meta:** transformar archivos sueltos en datasets limpios, tipados y versionados. **Sin React aún** (scripts Node/Python).

### 1.1 Unificar geometrías duplicadas
- Fusionar `areas_verdes.geojson` + `jefe_de_grupo.json` → un solo `areas_verdes.v1.geojson` con propiedades:
  - `id`, `nombre`, `codigo`, `uso`, `riegoActual`, `riegoProyectado`, `areaM2`, `perimetroM`, `jefeId`
- Extraer `geojsonSupervisores` (inline en `script.js`) → `zonas_supervisores.v1.geojson` (Zona1–4).
- Documentar el **zoning triple** (jefes ≠ zonas tachos ≠ nombres de lugar).

### 1.2 Normalizar tablas
| Dominio | Entrada | Salida | Acciones |
|---------|---------|--------|----------|
| Tachos | `tachos.csv` | `tachos.v1.json` | Enums de fracción; marcar `Peligrosos` vacío; join espacial con zonas |
| Flora | CSV + gviz | **solo** `flora.v1.json` desde gviz (~1050) | Deprecar CSV legacy; mapear fotos Drive |
| Cafetos | `cafetos.csv` | `cafetos.v1.json` o merge en flora con `tipo=cafeto` | Decidir producto |
| Vivero | `vivero.csv` | `vivero.v1.json` | Tipar mes/lugar |
| Monitoreo | 4 CSV | `lugares`, `actividades`, `poda`, `registros` tipados | Fechas ISO; claves estables |
| Puntos PUCP | CSV | `puntos_campus.v1.json` | Limpiar campos tipo scrape Maps |
| Bebederos | 5 GeoJSON | `bebederos.v1.geojson` + `estado` | Unificar tipos |
| Fauna / veredas / xerofítica / puertas / playas | GeoJSON | `*.v1` | Enriquecer attrs de puertas/playas (hoy casi solo `id`) |
| Reservas | GeoJSON 21 + Sheet 401 | `reservas_poligonos.v1` + **agenda TBD** | Re-publicar Sheet o importar Excel manual |

### 1.3 Contratos Zod
Un schema por dominio en `packages/schemas` (o `src/schemas`). Ningún componente React lee CSV crudo.

### 1.4 Fotos
- Mantener `data/drive_fotos/` como mirror.
- Manifest `fotos.v1.json`: `codigo → path local | url pública propia` (subir a storage controlado; no depender de `lh3`).

**Criterio de salida:** carpeta `data/v1/` con todos los dominios recuperables; script `npm run etl` regenera desde raw; CI valida Zod.

---

## Fase 2 — Esqueleto React + mapa 2D (3–4 días)

**Meta:** app vacía con el mismo campus, sin todas las capas.

| Entrega | Detalle |
|---------|---------|
| Scaffold | Vite + React + TS |
| Mapa | MapLibre centrado `[-12.068, -77.08]`, zoom 15 |
| Bases | OSM + Esri (o MapTiler); **evitar** tiles Google raster sin licencia |
| Shell UI | Panel de capas (estado Zustand), sidebar detalle, layout responsivo |
| Legacy side-by-side | Ruta `/legacy` iframe o link al estático servido, para comparar |

**Criterio de salida:** mapa base usable; panel de capas enciende/apaga capas vacías; deploy preview en Vercel cuenta nueva.

---

## Fase 3 — Migración por dominio (orden sugerido) (2–4 semanas)

Cada dominio: **cargar v1 → capa MapLibre/deck → filtros → popup/sidebar → tests de paridad visual** → flag `features.<dominio>=true`.

### Oleada A — Quick wins (días 1–4)
1. **Fauna** — pocos puntos, poca lógica  
2. **Veredas** — un MultiPolygon  
3. **Xerofítica**  
4. **Puertas / estacionamientos** — tras enriquecer attrs (si no, popups mínimos honestos)  
5. **Bebederos** — cluster deck.gl / supercluster; panel resumen  

### Oleada B — Polígonos core (días 5–10)
6. **Áreas verdes** — vistas uso / riego (una geometría, dos style expressions)  
7. **Jefes de grupo** — mismos polígonos, filtro por `jefeId` + métricas  
8. **Reservas polígonos** — sin agenda hasta resolver Sheet  

### Oleada C — Operación densa (días 11–18)
9. **Tachos** — multi-fracción + filtro zona (point-in-polygon o precomputar `zonaId` en ETL) + sidebar  
10. **Puntos PUCP / buscador**  
11. **Vivero** — filtro mes  
12. **Cafetos** (o absorbidos en flora)  

### Oleada D — Complejos (días 19–28)
13. **Monitoreo + poda + heatmap** — filtros cruzados; `HeatmapLayer`  
14. **Flora** — fuente única gviz; panel; fotos; edición **solo lectura** primero  
15. **Barredoras** — exportar rutas a GeoJSON; animación React; etiquetar como simulación  

### Oleada E — Cierre de huecos
16. **Edición flora** — sustituir Apps Script público por API autenticada  
17. **Agenda reservas** — cuando haya fuente (Sheet re-publicado o CSV manual)  
18. **Fotos restantes** de IDs en gviz no bajados en el lote de 70  

**Criterio de salida por dominio:** checklist de paridad (misma geometría ±1 m, mismos conteos ±0, filtros equivalentes).

---

## Fase 4 — Backend / acceso a datos vivos (en paralelo a oleadas C–E)

**Meta:** dejar de depender de CSV `2PACX` públicos y `/exec` abiertos.

| Opción | Cuándo |
|--------|--------|
| **A. Solo estáticos v1** | Si los datos cambian poco; editores actualizan Git o un admin sube archivos |
| **B. API Edge + service account** | Si Sheets sigue siendo el CMS operativo |
| **C. DB (Postgres/PostGIS)** | Si quieren consultas espaciales, historial y multi-usuario |

Recomendación: **A para lanzar**, **B o C en cuanto haya dueño de datos**. Nunca reexponer Apps Script de escritura sin auth.

---

## Fase 5 — Capacidad 3D (después de paridad 2D de áreas verdes + jefes) (1–2 semanas)

**Solo cuando** áreas verdes y jefes estén estables en MapLibre.

| Nivel | Qué | Esfuerzo |
|-------|-----|----------|
| **3D ligero** | `fill-extrusion` con altura ∝ √área o altura fija por uso; terrain DEM si aporta | Bajo |
| **3D medio** | deck.gl ColumnLayer / Grid sobre densidad flora o tachos | Medio |
| **3D fuerte** | Modelo glTF del campus + R3F | Alto (hace falta el modelo; hoy no está en el recovery) |

**Alcance inicial recomendado:** 3D ligero sobre **áreas verdes** (el pedido original), con toggle 2D/3D. No bloquear el resto de dominios por un mesh que no existe.

---

## Fase 6 — Corte y apagado del legado (2–3 días)

1. Checklist UAT con operadores de campus (jardinería, residuos, vivero).  
2. Redirect `mapa-web-6.vercel.app` → proyecto nuevo **solo si** recuperan la cuenta Vercel; si no, dominio/proyecto nuevo y comunicar URL.  
3. Archivar monolito en `/legacy` 30–60 días.  
4. Rotar/cerrar Apps Script públicos de escritura si ya no se usan.

---

## Orden de valor vs riesgo

```
Alto valor / bajo riesgo     →  Fauna, veredas, bebederos, áreas verdes 2D
Alto valor / alto riesgo     →  Flora, monitoreo, tachos+zonas
Pedidos “wow” (3D)           →  Después de áreas verdes tipadas
Bloqueantes de negocio       →  Sheet reservas 401, ownership de Sheets/Drive
```

---

## Decisiones que necesitas tomar (antes de código de producto)

1. **¿Sheets sigue siendo el CMS** o migran a formularios/admin en la app?  
2. **¿Cuenta Vercel/GitHub nuevas** ya existen? (despliegue y CI)  
3. **¿Prioridad #1 es áreas verdes 3D** o paridad operativa total (tachos/flora)?  
4. **¿Tienen modelo 3D del campus** o solo extrusión de polígonos?  
5. **¿Cómo recuperan la agenda de reservas** (re-publicar Sheet, Excel, o posponer)?

---

## Cronograma orientativo (1 persona full-time)

| Fase | Duración |
|------|----------|
| 0 Congelar | 1–2 días |
| 1 ETL + schemas | 3–5 días |
| 2 Esqueleto React + mapa | 3–4 días |
| 3 Dominios A–B | ~1.5 semanas |
| 3 Dominios C–D | ~2 semanas |
| 4 Backend (mínimo viable) | solapado, +3–5 días |
| 5 3D ligero áreas verdes | 3–5 días |
| 6 Corte | 2–3 días |
| **Total orientativo** | **~6–8 semanas** a paridad útil + 3D ligero |

(Equipo de 2 puede comprimir ~30–40 % si uno hace ETL/API y otro UI/mapa.)

---

## Qué no hacer al inicio

- Reescribir `script.js` línea a línea en React.  
- Meter Three.js antes de tener GeoJSON tipado.  
- Seguir leyendo CSV publicados desde el cliente en producción.  
- Asumir que “Peligrosos”, “cisterna/manual” o la agenda de reservas existen en datos reales.  
- Usar tiles Google satélite sin cumplir ToS / facturación.

---

## Próximo paso sugerido (cuando apruebes el plan)

**Fase 0 + inicio Fase 1:** crear repo, `data/v1` con áreas verdes unificadas + zonas supervisores extraídas, schemas Zod de esos dos. Eso desbloquea el 3D de áreas verdes sin esperar flora/monitoreo.
