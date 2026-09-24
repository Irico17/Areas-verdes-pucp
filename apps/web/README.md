# apps/web

Visor PWA del campus: React, TypeScript, Vite y MapLibre. El mapa no tiene backend propio; habla con `apps/api`.

## Arranque

Con la API en el puerto 8091:

```bash
npm install
npm run dev
```

Desde la raíz del repo: `make web`. Abre http://127.0.0.1:4317.

Vite reenvía `/api` y `/health` a `http://127.0.0.1:8091`.

## Qué muestra

- Base OpenStreetMap (teselas raster). Sin Street View.
- Capas de catastro: áreas, zonas, jardines de reserva y xerofítica.
- Sesión local (cookie). Cuentas: norte, sur, riego, coordinacion, jefatura, admin. Clave de desarrollo: `pando-local`.
- Módulos: Mapa, Labores, Catastro, Solicitudes, Reportes, Catálogos y Admin. El capataz solo ve mapa, labores y catastro, y solo sus labores.
- Labores abiertas: color = estado, letra = tipo. Pin para crear, propia o tercerizada, bitácora y evidencia en disco.
- Si no hay red, el alta y el cambio de estado se encolan; la última lista queda en este navegador.
- Vista Plano / Relieve. En relieve, las áreas se extruyen y se pueden encender las huellas OSM.
- Inventario opcional y una agenda marcada como ficticia.
- Sugerencia de tipo a partir del título: regla local, hay que confirmarla.

El manifiesto PWA y el service worker se generan con `vite-plugin-pwa`. En desarrollo el worker no intercepta la API.
