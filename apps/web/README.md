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
- Interruptor de rol (Jefatura, Coordinación, Capataz) guardado en el navegador. No es SSO.

El manifiesto PWA y el service worker se generan con `vite-plugin-pwa`. En desarrollo el worker no intercepta la API.
