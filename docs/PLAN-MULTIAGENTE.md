# Plan de ejecución multiagente — VerdePUCP

Fecha: 2026-09-25. Solo planificación. El orden es el de la propuesta DP2 v0.2 (`docs/fuente/Propuesta_Arquitectura_DP2_v0.2.docx`, §6): base operativa, después atención completa, después capacidades priorizadas. Dentro de eso se mete la carga de datos y la edición que pidió el producto, porque sin esos datos la base operativa no refleja el campus.

El PR #1 (limpieza de CI, borrador, sin merge) no es dependencia. Si se mergea, no cambia el reparto de archivos de abajo.

Contrato previo a cualquier ola de código: este documento y `docs/MAPA-DATOS-Y-EDICION.md`. Quien implemente no redefine columnas.

## Cómo no pisarse

1. Migraciones. El último archivo es `apps/api/migrations/006_limpieza.sql`. Cada frente usa solo su rango (`MAPA-DATOS-Y-EDICION.md` §1). Un frente no renumera el rango de otro.
2. Rutas. `apps/api/internal/server/server.go` es el punto de conflicto. La ola 0 deja una función `registrar(r, deps)` por paquete. Cada frente edita su archivo `handlers/<frente>.go` y una sola línea de registro. No se reescribe el bloque entero.
3. OpenAPI. `apps/api/openapi.yaml` se parte por tag. Cada frente añade paths bajo su tag y no reordena el archivo completo.
4. Web. Un frente no edita `App.tsx` salvo para montar un módulo ya exportado. El módulo vive en `apps/web/src/panel/<frente>.tsx`.
5. ETL. `internal/etl/run.go` hoy hace `TRUNCATE`. A partir de la ola 1 el cargador nuevo es `internal/etl/lote.go` (upsert por `origen_ref`). El `TRUNCATE` de catastro e inventario no se vuelve a llamar sobre tablas ya editadas.
6. Datos personales. El ETL aplica la tabla de ficticios de `MAPA-DATOS-Y-EDICION.md` §2. Ningún PR incluye nombres reales.
7. Rama. `cursor/<frente>-77cf` no aplica: el sufijo `-77cf` es de este agente de documentación. Los agentes de implementación usan el prefijo que les indique su entorno y no comparten rama.

## Ola 0 — Contrato (un solo agente, bloquea al resto)

Objetivo: dejar el esqueleto de registro de rutas, la tabla `cambios` y los rangos de migración, sin cargar datos todavía.

Alcance: `server.go` (extraer registro), migración `007_auditoria.sql`, `008_fk_minimas.sql` (FK de `permisos.rol` hacia un catálogo de roles semilla, sin cambiar el comportamiento visible), prueba de que `006` sigue aplicando.

HUIDs que prepara, no cierra: 01, 03, 20, 30, 39.

Dependencias: ninguna.

Criterio: `go test ./...` sigue en verde; existe `cambios`; no hay `TRUNCATE` nuevo.

Riesgo: bajo, si es el único que toca `server.go` en esta ola.

Prompt:

```text
Eres el agente de la ola 0 de VerdePUCP. Lee docs/PLAN-MULTIAGENTE.md y docs/MAPA-DATOS-Y-EDICION.md. No cargues datos del campus. Extrae el registro de rutas de apps/api/internal/server/server.go a funciones por paquete sin cambiar URLs. Añade apps/api/migrations/007_auditoria.sql (tabla cambios) y 008_fk_minimas.sql según el mapa de datos. No uses números 009 en adelante. No pongas nombres de personas. Prueba: go test ./... en apps/api. Criterio: las rutas actuales responden igual.
```

## Ola 1 — Base operativa (en paralelo después de la ola 0)

Pueden correr a la vez: frentes 1A, 1B, 1C y 1D. El 1E espera a 1A (necesita las tablas). No esperan entre sí 1A–1D.

### 1A — Esquema de catastro y ejemplares

Objetivo: tablas nuevas vacías y FK, sin volcar todavía los 1081.

Alcance: migraciones `010`–`019` solamente. Paquete `internal/catastro`. No toca `apps/web` ni `run.go`.

HUIDs: 04, 05, 06, 33 (modelo, no la pantalla).

Dependencias: ola 0.

Archivos: `apps/api/migrations/010_*.sql` … `019_*.sql`, `internal/catastro/`.

Criterio: las tablas del diagrama existen; `zonas` actual se puede leer igual (vista o renombre con vista `zonas` de compatibilidad hasta el frente 1E); `actividades.geom` admite NULL si hay lugar o zona.

Riesgo: el renombre de `zonas`. Mitigación: crear `poligonos_cuadrilla`, copiar, dejar una vista `zonas` hasta que 1E corte los clientes. No borrar la tabla en esta ola.

Prompt:

```text
Frente 1A de VerdePUCP. Lee docs/MAPA-DATOS-Y-EDICION.md secciones 1, 4.1 a 4.6 y 4.13. Crea solo migraciones 010-019 y el paquete Go de catastro. No importes GeoJSON. No edites apps/web ni server.go salvo la línea de registro que dejó la ola 0. Mantén una vista de compatibilidad llamada zonas. actividades.geom puede ser NULL si hay lugar o zona de supervisión. HUIDs 04, 05, 06 y 33 (modelo). go test ./...
```

### 1B — Seguridad de la sesión y de las mutaciones

Objetivo: que no se pueda mutar con `actor_rol` suelto ni con CORS `*`.

Alcance: `internal/accesos`, `handlers/operacion.go`, `handlers/producto.go`, `server.go` solo en el middleware, `apps/web/nginx.conf` cabeceras, tests. No cambia el modelo de ejemplares.

HUIDs: 01, 02, 03, 32, 39, 40, 41 (secretos de código; el user data de Terraform es el frente 3C).

Dependencias: ola 0.

Criterio: sin cookie, `POST /api/v1/operacion/actividades` responde 401; la cookie de pruebas sigue HttpOnly; CORS solo refleja el origen configurado; hay rate limit básico en el login; `go test` cubre el 401.

Riesgo: la PWA manda `actor_rol` (`apps/web/src/operacion.ts`). Mitigación: el servidor ignora el cuerpo y usa la sesión; el cliente se actualiza en el mismo frente, archivos `operacion.ts` y `producto.ts` solamente.

Prompt:

```text
Frente 1B de VerdePUCP. Cierra el hueco de actor_rol sin sesión descrito en docs/AUDITORIA-PRODUCCION.md (HUID 01, 03, 32, 39, 40). Toda mutación exige cookie y Permite. CORS configurable, no asterisco. Rate limit en POST /api/v1/sesion. No implementes SSO. No edites migraciones fuera de 007-009 si hace falta un índice, y no toques 010+. Ajusta apps/web/src/operacion.ts para no depender de actor_rol. Prueba el 401 sin cookie.
```

### 1C — Formularios de escritorio del catastro (contra el contrato)

Objetivo: pantallas que editan atributos y geometría cuando el API del 1A exista. Hasta entonces, tipos y cliente contra el contrato escrito, con pruebas de componente que no exigen la API.

Alcance: `apps/web/src/panel/CatastroEditor.tsx`, `apps/web/src/map/` solo un control de dibujo nuevo en archivo propio `draw.ts`. No reescribe `CampusMap.tsx` entero: añade un modo.

HUIDs: 04, 06, 33, 37, 41 (controles de campo).

Dependencias: contrato de la ola 0 y de `MAPA-DATOS`. Puede desarrollarse en paralelo al 1A usando fixtures.

Criterio: se puede cambiar un vértice en un GeoJSON de prueba y el payload coincide con el contrato; el formulario lista los campos de la sección 4.1 y 4.2.

Riesgo: conflicto en `CampusMap.tsx`. Mitigación: archivo nuevo `draw.ts` importado con una prop `modoDibujo`.

Prompt:

```text
Frente 1C de VerdePUCP. En apps/web, añade edición de atributos y de geometría de áreas y de zonas de supervisión según docs/MAPA-DATOS-Y-EDICION.md 4.1 y 4.2. No reescribas CampusMap.tsx: un módulo draw.ts. No inventes campos. HUIDs 04, 06, 33. La API puede no estar: usa el contrato del documento y fixtures. Texto en español. No toques apps/api/migrations.
```

### 1D — Evidencias de móvil

Objetivo: el flujo de la sección 4.16.

Alcance: `apps/web/src/offline/queue.ts` (nuevo store, no cambiar el formato de los stores viejos), `internal/blobs/`, `handlers/producto.go` `SubirEvidencia`, migración `040` solo si el 1A no la tomó: la reserva 040–049 es de este frente y del 2C. Usar `040_evidencia_meta.sql` para lat/lon/exif y hash.

HUIDs: 12, 15, 22, 30, 36.

Dependencias: ola 0. No depende de 1A.

Criterio: una foto se comprime bajo 1.5 MB en el cliente; sin red queda en IndexedDB; al volver la red se envía una vez; el evento `evidencia` aparece en la bitácora; otro cuerpo con el mismo UUID responde 409.

Riesgo: `producto.go` también lo toca 1B. Mitigación: 1D solo edita las funciones `SubirEvidencia` y `Archivo`. 1B no las reescribe; si ambas necesitan middleware, el middleware queda en 1B y 1D solo llama `exige`.

Prompt:

```text
Frente 1D de VerdePUCP. Implementa docs/MAPA-DATOS-Y-EDICION.md sección 4.16. Compresión en el cliente, cola IndexedDB nueva sin romper cola-labores, POST de evidencia con hash, evento de bitácora tipo evidencia, 409 si el UUID se repite con otro contenido. Migración solo 040_evidencia_meta.sql. HUIDs 12, 15, 22, 30, 36. No implementes S3 nuevo: usa internal/blobs. No llames Apps Script.
```

### 1E — Carga de la base (después de 1A, en paralelo con el final de 1C y 1D)

Objetivo: volcar lo que ya es base operativa: áreas (si hace falta reconciliar), zonas Z1–Z4, polígonos de cuadrilla anonimizados, lugares 76, especies y 1081 ejemplares, palmeras con medidas, cafetos, catálogo de 45 tipos.

Alcance: `internal/etl/lote.go` y lectores nuevos. Lee la web en vivo una vez, o `data/raw` si el vivo falla, y escribe por upsert. No `TRUNCATE`.

HUIDs: 04, 06, 08 (taxonomía), 27, 33.

Dependencias: 1A mergeado.

Criterio: conteos verificables contra el documento (4 zonas, 76 lugares, 45 tipos, 1081 ejemplares o el delta que el reporte de errores explique, 74 palmeras válidas y 2 rechazadas). Cero nombres reales en la base (consulta de que no aparecen los fragmentos que `baseline.go` ya trata como PII).

Riesgo: `run.go`. Mitigación: no modificar el camino viejo; el comando nuevo es `etl-lote`.

Prompt:

```text
Frente 1E de VerdePUCP. Con las tablas del frente 1A ya mergeadas, carga por upsert (sin TRUNCATE) zonas de supervisión, polígonos de cuadrilla, lugares, especies, ejemplares de la gviz, medidas de palmera, cafetos y el catálogo de 45 tipos. Fuentes y rechazos: docs/MAPA-DATOS-Y-EDICION.md. Anonimiza responsables con la regla de la sección 2; no escribas nombres reales en código ni en commits. HUIDs 04, 06, 08, 27, 33. Entrega un reporte de conteos y de filas rechazadas.
```

## Ola 2 — Atención completa (después de 1A y 1E)

En paralelo: 2A, 2B y 2C. 2D (reportes) después de 2A, porque necesita las labores cargadas para probar filtros.

### 2A — Labores, poda, vivero, solicitudes

HUIDs: 08, 09, 11, 13, 14, 18, 19, 20, 21, 34.

Migraciones `020`–`029`. Módulos `panel/Labores.tsx` (ampliar, no sustituir el pin), `panel/Poda.tsx`, `panel/Vivero.tsx`.

Criterio: 283 labores importadas con el reporte de estados vacíos; 25 podas; vivero con el conteo vivo o el delta explicado; una labor tercerizada no cierra sin orden; el código OSG no se autogenera; riego guarda zona por FK.

Prompt:

```text
Frente 2A de VerdePUCP. Migraciones solo 020-029. Importa monitoreo 2026 (283), poda (25) y vivero según docs/MAPA-DATOS-Y-EDICION.md 4.8 a 4.10. Formularios de edición en escritorio. Respeta ejecutor propia/tercerizada y el código externo. Estado vacío se importa como sin_estado. Anonimiza responsables. HUIDs 08, 09, 11, 13, 14, 18, 19, 20, 21, 34. No TRUNCATE. No SSO ni Centuria automática.
```

### 2B — Tachos, bebederos, puntos, reservas ficticias, capas ya cargadas editables

HUIDs: 04 (inventario), 07 (el formato, la herramienta genérica es 3A), 33.

Migraciones `030`–`039`.

Criterio: los 11 conteos de tachos suman lo publicado en el mapa de datos; bebedero tiene estado y sede; reservas marcadas ficticias; puntos PUCP sin teléfono ni placeId.

Prompt:

```text
Frente 2B de VerdePUCP. Migraciones solo 030-039. Completa tachos con conteos, bebederos con estado y sede, puntos PUCP sin datos de contacto, y reservas ficticias desde data/mocks. Haz editables fauna, puertas, playas, vereda, xerofítica y jardines de reserva. docs/MAPA-DATOS-Y-EDICION.md 4.11 a 4.15. No abras la hoja de reservas (401). No guardes teléfonos.
```

### 2C — Historial filtrable y reversión de lote

HUIDs: 20, 30, 34, 39 (auditoría), 49 (conservar traza).

Usa `cambios` de la ola 0. No nueva numeración fuera de `041_lotes.sql` si 040 ya lo usó 1D; si 040 está tomado, usa `041`.

Criterio: revertir un lote de prueba restaura el `antes`; una fila editada después no se pisa; el timeline muestra usuario de sesión.

Prompt:

```text
Frente 2C de VerdePUCP. Completa la tabla cambios (ola 0) con lotes de importación y reversión segura descrita en docs/MAPA-DATOS-Y-EDICION.md 4.17. El actor del timeline es el usuario de sesión, no actor_rol. Migración 041_lotes.sql solamente. HUIDs 20, 30, 34, 39, 49. Prueba de integración: importar 2 filas, revertir, comprobar el estado previo.
```

### 2D — Reporte básico cerrado

HUIDs: 23, 25 (Excel del básico; PDF no).

Depende de 2A.

Criterio: el CSV y el SpreadsheetML incluyen clase, lugar, cuadrilla ficticia y fechas; no aparece un nombre real; no se añade un indicador oficial.

Prompt:

```text
Frente 2D de VerdePUCP. Amplía GET /api/v1/reportes/labores para filtrar por zona, cuadrilla, origen y fechas, con las columnas que ya existen. Mantén CSV y SpreadsheetML. No implementes PDF ni KPIs. HUIDs 23 y 25. Si una columna no está en el backlog como definida, no la inventes: déjala fuera y anótala en el OpenAPI.
```

## Ola 3 — Capacidades priorizadas y producción

En paralelo: 3A, 3B y 3C.

### 3A — Importación genérica con vista previa

HUIDs: 07, 44 (Excel).

Un endpoint `POST /api/v1/importaciones?entidad=` que reutiliza los lectores de 1E y 2A. Vista previa, errores por fila, confirmar, revertir vía 2C.

Criterio: subir un CSV inválido no escribe; confirmar escribe; revertir deshace.

Prompt:

```text
Frente 3A de VerdePUCP. Importador único CSV/XLSX/GeoJSON con vista previa, validación y reporte de errores, para cada entidad de docs/MAPA-DATOS-Y-EDICION.md. No dupliques reglas: llama a los lectores ya mergeados. HUID 07 y 44. Confirmación en un segundo paso. Integración con la reversión del frente 2C.
```

### 3B — Indicadores honestos e IA como está

HUIDs: 16 y 26 (diferidos: mostrar el hueco), 24 (hueco de proceso), 28, 48, 49.

No llamar a un modelo externo. La regla local de `internal/atencion/ia.go` se mantiene y se etiqueta.

Criterio: la pantalla dice «definición pendiente» donde no hay fórmula acordada; la IA sigue en proceso.

Prompt:

```text
Frente 3B de VerdePUCP. No inventes indicadores. En la UI, muestra conteos operativos con la frase de definición pendiente para cobertura, rendimiento y métricas de proveedor (HUID 16, 24, 26). Deja la IA local de internal/atencion/ia.go con confirmación humana y sin red (HUID 28, 48, 49). Sin PDF y sin proveedor externo.
```

### 3C — Producción operable

HUIDs: 38, 40, 41, 42, 45, 47. Checklist de `docs/AUDITORIA-PRODUCCION.md` §5.

Alcance: `infra/terraform/main.tf` (`user_data_replace_on_change = false`), secretos por variable sensible sin default de laboratorio, nginx TLS cuando haya certificado, script `scripts/backup-postgis.sh` y `scripts/restore-postgis.sh`, prueba documentada, healthcheck, logs sin cuerpos. Documento `docs/OPERACION.md` (este frente sí puede añadirlo: es el runbook, no está en la ola de auditoría).

Criterio: un apply que solo cambia la imagen no reemplaza la instancia (plan de Terraform lo demuestra); el backup restaura en una base vacía de prueba; la guía de costos está actualizada.

Costos aproximados fuera del Learner Lab (us-east-1, un mes, orden de magnitud, no cotización):

| Pieza | Aprox. |
|-------|--------|
| t3.small encendida | 15 USD |
| RDS Postgres `db.t3.micro` con 20 GB, o la EC2 actual con backup | 15–20 USD si se separa la base; 0 extra si sigue en el volumen |
| EBS gp3 20 GB + snapshots diarios de 20 GB | 4–8 USD |
| S3 evidencias, pocos GB, más versionado | 1–3 USD |
| Route 53 + certificado ACM | menos de 1 USD |
| IP elástica en uso | 0 |
| **Orden de magnitud** | **25–45 USD al mes** sin NAT ni ALB |

El Learner Lab sigue siendo el piloto (saldo de curso, credenciales de unas 4 horas, EC2 detenida al cerrar la sesión). Producción de verdad no cabe ahí: hace falta una cuenta con presupuesto propio. No encender ALB ni ECS (el propio `DEPLOY-AWS.md` estima 16–20 USD solo por el balanceador).

Prompt:

```text
Frente 3C de VerdePUCP. Prepara producción según docs/AUDITORIA-PRODUCCION.md sección 5 y docs/DEPLOY-AWS.md. Quita user_data_replace_on_change o ignóralo para que apply no reemplace la EC2. Saca pando-local y campus-lab del user data. Añade scripts de backup y restauración de PostGIS y docs/OPERACION.md. No despliegues contra AWS real. HUIDs 38, 40, 41, 42, 45, 47. Incluye en el runbook el costo aproximado 25-45 USD/mes fuera del Learner Lab.
```

## Ola 4 — Diferido (no se lanza agente)

Queda escrito para no reabrirlo por inercia. Motivo entre paréntesis.

| Tema | HUIDs | Motivo |
|------|-------|--------|
| SSO PUCP | 01 | Pendiente de validación con el cliente |
| API Centuria / OSG / correo | 11, 21 | Captura manual aprobada |
| Portal del proveedor | 17 | Could, fuera del MVP |
| Métricas de proveedor e indicadores oficiales | 16, 26 | Fórmulas no definidas |
| PDF y reportes intermedio/avanzado | 23, 24, 25 | Columnas pendientes de validación |
| Ortofoto | 46 | Sin acuerdo OSG |
| Offline completo (catálogos, riego, Background Sync, conflicto de versión en todos los módulos) | 12, 36 | La ola 1D cubre fotos y labores; el resto espera prueba de campo |
| Insumos y avance por área con horas | 10, 19 | Should, después de que exista personal |
| Recodificación masiva | 05 | La tabla queda en 1A; la herramienta, cuando haya códigos nuevos de la universidad |
| Barredoras con GPS real | — | La ruta del script es demo |
| React Three Fiber / glTF | — | `PLAN-PRODUCTO-Y-UI.md` §D |

## Resumen de paralelismo

| Ola | A la vez | Espera |
|-----|----------|--------|
| 0 | Un agente | — |
| 1 | 1A, 1B, 1C, 1D | 1E después de 1A |
| 2 | 2A, 2B, 2C | 2D después de 2A |
| 3 | 3A, 3B, 3C | 3A después de 2C |
| 4 | Nadie | Acuerdo externo |

Conflictos más probables y el corte:

| Archivos | Frentes | Corte |
|----------|---------|-------|
| `server.go` | todos | Ola 0 deja el registro; luego una línea por frente |
| `producto.go` | 1B y 1D | 1B middleware; 1D solo subida |
| `operacion.ts` | 1B | solo 1B |
| `CampusMap.tsx` | 1C, 2B | 1C en `draw.ts`; 2B en capas de `inventario.ts` |
| migraciones | todos | rangos 007–009, 010–019, 020–029, 030–039, 040–049 |
| `run.go` | 1E | archivo nuevo `lote.go`, el viejo no se llama sobre datos editados |
