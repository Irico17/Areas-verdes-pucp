# Mapa de Control Ambiental Campus (recuperación)

Aplicación web de control ambiental del campus (PUCP): mapa interactivo con capas de áreas verdes, flora, fauna, tachos, bebederos, vivero, monitoreo y más.

- **Sitio en vivo:** https://mapa-web-6.vercel.app/
- **Stack:** HTML/CSS/JS + Leaflet + MarkerCluster + leaflet.heat + Turf + PapaParse
- **Documentación de esta recuperación:**
  - [ANALISIS.md](./ANALISIS.md) — arquitectura, módulos, deuda y notas de modernización
  - [INVENTARIO-DATOS.md](./INVENTARIO-DATOS.md) — todas las fuentes de datos y su estado

## Estructura

```
mapa-web-6-recovery/
  index.html, style.css, script.js
  *.geojson / jefe_de_grupo.json     # capas estáticas
  data/
    sheets/       # CSV desde Google Sheets
    gviz/         # JSON flora (gviz)
    appscript/    # respuestas Apps Script
    drive_fotos/  # 70 fotos públicas de Drive
  ANALISIS.md
  INVENTARIO-DATOS.md
```

## Cómo abrir en local

Servir la carpeta con cualquier static server (los `fetch` relativos y CORS de Sheets requieren HTTP, no `file://`):

```bash
cd mapa-web-6-recovery
python3 -m http.server 8080
# abrir http://localhost:8080
```

Nota: la capa de **reservas de jardines** depende de un Sheet que hoy responde **401**; el resto de fuentes públicas se recuperó en `data/`.

## Paquete

`/workspace/mapa-web-6-completo.tar.gz` contiene esta carpeta completa.
