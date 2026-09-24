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
- Labores abiertas: color = estado, letra = tipo. Pin para crear, panel para asignar y ver la bitácora.
- Capataz solo consulta el equipo elegido. Si no hay red, el alta se encola en IndexedDB.
- Vista Plano / Relieve. En relieve, las áreas se extruyen y se pueden encender las huellas OSM.
- Inventario opcional (bebederos, fauna, tachos, flora, cafetos, playas, puertas, vereda) y una agenda marcada como ficticia.

El manifiesto PWA y el service worker se generan con `vite-plugin-pwa`. En desarrollo el worker no intercepta la API.
