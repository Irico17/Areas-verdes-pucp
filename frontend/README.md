# frontend

Visor de VerdePUCP. React 19, TypeScript, Vite 8 y MapLibre. El mapa no tiene backend propio: habla con `backend/app` por el proxy de Vite (`/api`, `/areas-verdes` y `/health`).

## Arranque

Con la API en el puerto 8091:

```bash
npm ci
npm run dev
```

Desde la raíz del repositorio: `make web`. Abre http://127.0.0.1:4317 (`strictPort`).

`VITE_DEV_API` cambia el destino del proxy (por defecto `http://127.0.0.1:8091`). No es una variable del navegador: la lee el proceso de Vite.

## Scripts

| Script | Qué hace |
| --- | --- |
| `npm run dev` | Servidor de desarrollo. |
| `npm run build` | `tsc -b` y `vite build`. |
| `npm run preview` | Sirve el build en el 4317. |
| `npm run lint` | `oxlint --max-warnings 0`. |
| `npm test` | Pruebas unitarias con el runner de Node (`node --test`). |
| `npm run test:e2e` | Playwright. `npm run e2e` es el mismo comando. |

## Estructura

```
src/main.tsx          entrada
src/App.tsx           sesión y marco
src/api.ts            cliente HTTP
src/session.ts        cookie de sesión
src/map/             mapa, capas, dibujo y selección
src/panel/           pantallas (Hoy, labores, catastro, solicitudes, reportes…)
src/offline/         cola IndexedDB, compresión y subida
src/ui/              navegación, permisos, nomenclatura, ayuda
e2e/                 Playwright
```

En pantallas de hasta 640 px el panel de capas arranca cerrado. En escritorio se puede ocultar; la elección queda en este navegador, por persona y por ancho.

El service worker lo genera `vite-plugin-pwa`. En desarrollo no intercepta la API. Las teselas de OpenStreetMap se cachean en el build de producción.

## Pruebas

Unitarias, sin navegador:

```bash
npm test
```

Cubren mapa, paneles, permisos, nomenclatura, cola offline y teclado. La lista de archivos está en el script `test` de `package.json`.

Extremo a extremo: hace falta la API real y una base migrada con la semilla ficticia (`go run ./cmd/migrate -semilla-ficticia` en `backend/app`). La clave de demostración local es `pando-local`.

```bash
npm run test:e2e
```

Playwright usa `E2E_BASE_URL` (por defecto `http://127.0.0.1:4317`) y levanta Vite si no está ya arriba. En CI (`CI=true`) no reutiliza un servidor previo. Las capturas quedan en `e2e-artifacts/` y el informe en `playwright-report/`. Esos directorios no se versionan.

El job `frontend` de `.github/workflows/ci.yml` corre `npm ci`, `oxlint --max-warnings 0`, las pruebas unitarias y `tsc -b && vite build`. El job `e2e` es el recorrido de extremo a extremo: PostGIS, `migrate`, semilla, API en `127.0.0.1:8091` y `npm run test:e2e`.

## Nomenclatura

Las etiquetas visibles están en `src/ui/nomenclatura.ts` (`MODULO`, `CAPA`, `EJEMPLAR` y el resto). Los códigos de rol, de estado, de capa y de ruta no se renombran ahí.

El vocabulario acordado, con los sinónimos que no deben volver a la interfaz, está en [docs/GLOSARIO-NOMENCLATURA.md](../docs/GLOSARIO-NOMENCLATURA.md).

## Roles y permisos

`GET /sesion` no devuelve la lista de permisos. Las pestañas salen de `src/ui/permisos.ts`, copia de la matriz semilla del backend. Los códigos de acción no se renombran.

| Rol | Acciones |
| --- | --- |
| `capataz` | consultar, registrar |
| `coordinacion` | consultar, registrar, validar, solicitudes, reportes |
| `jefatura` | consultar, validar, reportes, solicitudes, evidencias, usuarios |
| `admin` | consultar, registrar, validar, reportes, catálogos, solicitudes, usuarios |

Cuentas de demostración: `norte`, `sur` y `riego` son capataces (cuadrilla norte, sur y riego). `coordinacion`, `jefatura` y `admin` usan el rol del mismo nombre.

El orden de pestañas está en `src/ui/registroModulos.ts`. El capataz ve Hoy, mapa, actividades, registros, bitácora y ejemplares. Coordinación añade resumen, solicitudes, catastro, inventario, reportes, importación, catálogos e historial. Jefatura no tiene el alta de campo: no ve Guardar en catastro ni los formularios de poda, vivero o riego. Administración añade el módulo de cuentas.

Si no hay red, el alta y el cambio de estado se encolan en este navegador y se reintentan con el mismo UUID.
