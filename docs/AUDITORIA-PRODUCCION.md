# Auditoría de salida a producción — VerdePUCP

Fecha de verificación: 2026-09-25. Solo lectura: no se modificó código, migraciones ni datos.

Fuentes primarias del cliente (leídas completas):

- `docs/fuente/Product_Backlog_Areas_Verdes.xlsm` — hoja `Product Backlog`, filas 6 a 54, HUID 01 a 49.
- `docs/fuente/Propuesta_Arquitectura_DP2_v0.2.docx` — Grupo 12, versión 0.2, 22 de septiembre de 2026.

Fuentes de contraste: código en `apps/api`, `apps/web`, `infra/terraform`, `docker-compose.yml`, `.github/workflows/ci.yml`; recuperación en `data/raw/`; web original https://mapa-web-6.vercel.app (GET públicos; no se llamó al Apps Script de escritura de flora). El PR #1 (`cursor/limpiar-avisos-ci-351f`, borrador, sin merge al 2026-09-25) limpia avisos de CI y no cambia este baseline de `main`.

Desalineación conocida del Excel, ya anotada en `docs/PLAN-PRODUCTO-Y-UI.md`: desde HUID 38 la celda *Funcionalidades* no coincide con *Como / Quiero*. La matriz usa el HUID, el texto de la historia y el RF impreso, y dice cuándo la celda de funcionalidades pide otra cosa.

## 1. Veredicto

El piloto cumple el núcleo del mapa (OSM, 521 áreas, labores con pin, sesión local, catálogos semilla, un reporte básico). No está listo para producción ni para operar con los datos reales del campus. Faltan el inventario de ejemplares (1081 filas vivas), las cuatro zonas de supervisión, el monitoreo 2026, integridad referencial, edición de geometrías, importación con vista previa, evidencias de campo offline, HTTPS, secretos de producción, backups de PostGIS y un despliegue que no recree la EC2.

## 2. Estado por área

Severidad: Alta (bloquea producción o pierde datos), Media (degrada la operación), Baja (deuda acotada). Esfuerzo: S (un módulo, pocos archivos), M (varios paquetes o una migración con UI), L (esquema + ETL + API + pantallas).

| Área | Qué hay | Qué falta | Severidad | Esfuerzo |
|------|---------|-----------|-----------|----------|
| Catastro de áreas | 521 MultiPolygon en `areas_verdes` (`apps/api/migrations/002_catastro.sql`). Ficha editable solo en `nombre`, `uso`, `riego_act`, `referencia` (`handlers/producto.go` `Fichas.Patch`). Alta sin geometría. | Editar el resto de atributos y la geometría en el mapa. El ETL hace `TRUNCATE` (`internal/etl/load.go` vía `cmd/etl`): una recarga borra ediciones humanas. Solo 187 códigos y 20 nombres en el GeoJSON local. | Alta | L |
| «Zonas» actuales | 534 polígonos de `data/raw/jefe_de_grupo.json` en la tabla `zonas`, sin el campo `jefes` (ADR-012, `internal/etl/baseline.go`). | No son las zonas de supervisión. En vivo (2026-09-25) `supervisoress.geojson` responde HTTP 200 y trae 4 MultiPolygon (`ZONA`, `Area`, `id`). La copia local del JSON ya no conserva la distribución original del campo `jefes` (un token repetido 531 veces, más «campo depo» ×2 y «Bosque húme» ×1). En vivo hay 3 nombres de persona (259, 168 y 104) más esas dos etiquetas de sector. Hay que anonimizar, no publicar los nombres. | Alta | M |
| Ejemplares / flora | Capa `flora` del inventario: 74 puntos de `flora.csv` (palmeras), solo lugar y nombres (`internal/etl/inventario.go` `readSheetPoints`). | Hoja gviz viva: 1081 filas, todas con coordenadas, 1040 con foto, 635 con código, 92 nombres científicos, 40 ubicaciones. Tipos: Árbol 665, Palmera 269, Arbusto 85, Herbácea 39, Trepadora 22, Suculenta 1. La copia `data/raw/gviz/flora_gviz_default.json` está vieja (1050 filas). | Alta | L |
| Inventario auxiliar | Bebederos 67, fauna 19, puertas 7, playas 15, vereda 1, cafetos 53, tachos 184. Conteos en `data/v1/inventario/`. | Tachos sin conteos por tipo (en el CSV local: no aprovechables 294, papel 142, plástico 260, vidrio 242, pilas 56, peligrosos 0, RAEE 12, metales 2, Aniquem 38, intermedios plástico 26, intermedios metal 4). Bebederos sin estado ni sede como columnas. Palmeras sin altura, fuste, DAP, radio ni zunchado. Cero JPEG en `data/raw/drive_fotos/` (solo `_index.json`). | Alta | L |
| Labores | 7 filas semilla en `003_operacion.sql` y `006_limpieza.sql`. Estados, pin, asignación, bitácora, ejecutor propia/tercerizada. | No están las 283 filas de monitoreo 2026 (102 con coordenadas, 219 con foto, 52 «Cerrado», 110 sin responsable; tres responsables con 69, 56 y 48). Catálogo real: 45 tipos en 7 clases (`data/raw/sheets/actividades.csv`), no los 5 tipos semilla. `geom` es `Point NOT NULL`: no se puede registrar por zona sin GPS, en contra de la propuesta DP2. | Alta | L |
| Poda, vivero, lugares, puntos, reservas | Jardines de reserva 21 y xerofítica 10 en `capas_auxiliares`. Mock de agenda: 71 reservas y 21 polígonos en `data/mocks/reservas_agenda.mock.json`. | Poda 25, vivero vivo 689 filas no vacías (la copia local tiene 683), lugares 76, puntos PUCP 153: no hay tablas ni ETL. La hoja de reservas sigue en HTTP 401. | Alta | L |
| Accesos | Cuentas semilla, bcrypt, cookie `cv_sesion` HttpOnly, 12 h, token solo como SHA-256 (`internal/accesos/accesos.go`). Matriz en código. | Misma clave `pando-local` (`CAMPUS_DEV_PASSWORD`). Cookie sin `Secure`. CORS `*` (`internal/server/server.go`). Sin CSRF, sin rate limit. Rutas de labores y geo no llaman a `exige`. `actor_rol` del JSON basta para mutar si no hay cookie (`handlers/operacion.go`). `permisos.rol` no tiene FK. | Alta | M |
| Catálogos | Tabla `catalogos`, alta y baja lógica. Clases: tipo, estado, prioridad, lugar, especie, motivo, turno, fuente (`internal/catalogos/store.go`). | Sin plagas, frecuencias, roles ni sedes (RF-24). Los CHECK de SQL no leen el catálogo. Lugares semilla son 4 textos, no los 76. Sin FK desde labores. | Media | M |
| Evidencias | `POST /api/v1/evidencias`, tope 8 MB, jpg/png/webp/pdf, disco o S3 (`internal/blobs/`). | Sin compresión, sin GPS de la foto, sin cola offline de archivos, sin vínculo al evento de la bitácora. El cubo S3 está apagado (`create_evidence_bucket` default false). Nginx permite 10 MB (`apps/web/nginx.conf`) y la API corta en 8 MB. | Alta | L |
| Offline | IndexedDB: alta de labor, cambio de estado y caché de lista (`apps/web/src/offline/queue.ts`). | No cubre riego, fotos, catálogos ni conflictos de versión. No hay Background Sync. DP2 pide UUID por evidencia y UNIQUE por operación. | Alta | L |
| Reportes e indicadores | `GET /api/v1/reportes/labores` en CSV con BOM y SpreadsheetML (`internal/atencion/export.go`). | Sin PDF, sin niveles intermedio/avanzado, sin historial por ejemplar. Indicadores oficiales explícitamente no definidos. | Media | M |
| IA | `POST /api/v1/ia/sugerir-tipo`: reglas locales, `requiere_humano` (`internal/atencion/ia.go`). | El caso de uso no está validado con el cliente (RF-25). No sustituye un modelo; no debe ampliarse a un proveedor externo sin acuerdo. | Baja | S |
| Seguridad de plataforma | IMDSv2 obligatorio, bucket de evidencias con acceso público bloqueado si se crea, Postgres no publicado en el compose de la EC2. | HTTP en el puerto 80. SSH `0.0.0.0/0` por defecto (`infra/terraform/variables.tf`). Clave de Postgres `campus-lab` y `pando-local` dentro de `user_data.sh.tftpl`. Sin cabeceras de seguridad en nginx. | Alta | M |
| Datos y backups | Volumen EBS de 20 GB para Postgres (`infra/terraform/main.tf`). | Sin `pg_dump` ni prueba de restauración. `user_data_replace_on_change = true` (línea 89): un `apply` que cambie el user data reemplaza la EC2. El disco de datos puede sobrevivir, pero el arranque no es un despliegue in-place. | Alta | M |
| Observabilidad | `gin.Logger()` y `/health`. | Sin métricas, alertas, correlación de petición ni tablero. RNF de logs no está cubierto. | Media | M |
| Pruebas | `go test` de validación, ETL, accesos, blobs y handlers sin base. Un test de cobertura de capas en `apps/web/src/map/coverage.test.ts`. CI en `.github/workflows/ci.yml`. | Sin tests de integración contra PostGIS, sin e2e, sin prueba móvil. | Alta | L |
| Migraciones | `001`–`006` aplicadas por `internal/migrate`. | No hay expansión compatible ni plan de numeración para varios agentes. El ETL no es una migración de datos editable. | Media | M |
| Accesibilidad | Algunas etiquetas `aria` en `App.tsx` y `FechaCampo.tsx`. UI en español. | Sin auditoría WCAG. Controles de mapa y checks densos. DP2 pide botones de campo y mensajes claros (RNF-06 en la historia de interfaz simple, que en el Excel está corrida de columna). | Media | M |
| Despliegue | Learner Lab documentado en `docs/DEPLOY-AWS.md`. CI de deploy solo `workflow_dispatch` y sale en verde si faltan secretos. | Credenciales del lab ~4 h. Sin dominio ni TLS. Salir del lab exige cuenta, red y costo propios. | Alta | L |

## 3. Brechas frente a la propuesta DP2 v0.2

La propuesta pide una PWA, un monolito Go (Gin + GORM), PostgreSQL, objetos privados, mismo dominio con HTTPS, y este orden: base operativa, atención completa, capacidades priorizadas. El código sigue ese stack (`apps/api`, `apps/web`, PostGIS 16 en `docker-compose.yml`). Las brechas son de completitud y de operación, no de tecnología elegida.

| Tema DP2 | Propuesta | Implementado | Brecha |
|----------|-----------|--------------|--------|
| Contenedores | PWA + API + Postgres + objetos, mismo dominio, HTTPS | Compose local y una EC2 con nginx en el puerto 80 | Sin TLS, sin dominio, API también publicada en 8091 en el compose local |
| Módulos | Accesos, catastro, operación, tercerizados, reportes | Paquetes `accesos`, `catalogos`, `catastro`, `operacion`, `atencion`, `inventario` | Falta ejemplar, insumos, personal distinto de usuario, recodificación |
| Catastro progresivo | Zona o lugar sin GPS; ejemplar con código visible e historial | `areas_verdes.geom` admite NULL. `actividades.geom` es obligatoria. No hay tabla de ejemplar ni de códigos anteriores | RF-04 y RF-05 abiertos |
| Integridad | Transacciones, no cerrar sin ejecución, desactivar catálogos, conservar código externo | Cierre de labor tercerizada exige orden (ola 2 del plan de producto). Catálogos con baja lógica. Código externo único parcial | Área, zona y catálogos no son FK. `permisos.rol` es texto libre |
| Offline | Encargos, catálogos, avances, riego y fotos. UUID, rechazo si el mismo UUID cambia de contenido, versión al corregir | Cola de altas y de estados. Idempotencia de alta de labor en `internal/operacion` | Sin fotos, sin riego, sin catálogos offline, sin conflicto de versión |
| Evidencias | Bucket privado, validar tipo y tamaño, clave en Postgres | Disco local y cliente S3 opcional (`internal/blobs/s3.go`) | Cubo apagado; sin compresión ni geolocalización de la toma |
| Seguridad | Cookie Secure + HttpOnly + SameSite, CSRF, permisos en cada operación, secretos fuera del repo | HttpOnly y SameSite Lax. Permisos solo en rutas de `handlers/producto.go` | Labores, geo e inventario públicos de facto. CORS `*`. Claves semilla en Terraform |
| Sync | UNIQUE(usuario, operación), reintento, pendientes visibles | UUID de labor. Indicador de cola en la PWA | No hay clave única usuario+operación |
| Reportes | Básico, intermedio y avanzado; PDF o Excel | Un reporte, CSV y SpreadsheetML | Sin PDF y sin los otros dos niveles |
| IA | Caso por validar, sin PII hacia fuera, supervisión humana | Reglas locales | Correcto como límite; el caso sigue sin validar |
| Recuperación | Backup diario y prueba de restauración antes del piloto | No existe | RPO/RTO sin acordar |
| Mapas | No se presupone GeoServer. OSM. Ortofoto si hay acuerdo | MapLibre + OSM + extrusión. Sin GeoServer | Ortofoto diferida. Las «zonas» del mapa no son las de supervisión |
| Integraciones | Centuria, OSG y correo manuales. SSO pendiente. Proveedor fuera del MVP | Solicitud manual con fuente `centuria\|osg\|correo\|interna` | Alineado. No hay conector |
| Rendimiento | Panel cada 60 s, paginación, índices | Índices GIST y `limit` en geo. Sin paginación de labores ni de fichas | Sin meta de tiempo medida |
| Personal vs usuario | Operario sin cuenta | No hay tabla `personal`. El capataz es un equipo ficticio (`capataces`) | No se puede asignar una cuadrilla real anonimizada |

## 4. Matriz de trazabilidad del backlog

Cada fila del Excel está aquí. Prioridad = columna *Prioridad del Product Owner*. Estado = código en `main` al 2026-09-25.

| HUID | Épica | Pri. | RF | Estado | Evidencia | Qué falta para cerrarla |
|------|-------|------|----|--------|-----------|-------------------------|
| 01 | A Acceso | Must | RF-01 | parcial | `POST /api/v1/sesion`, `internal/accesos/accesos.go`, cookie en `handlers/producto.go` | SSO diferido (el Excel lo deja pendiente). Cuentas con clave propia por usuario, cookie `Secure`, CSRF, rate limit, y que ninguna mutación acepte `actor_rol` sin sesión |
| 02 | A Acceso | Must | RF-02 | parcial | CHECK `usuarios_rol_chk` en `005_producto.sql`; semillas en `accesos.go` | Roles no son configurables: el enum está en SQL y en código. Operario sin cuenta: sí se respeta (no hay login de operario). Falta el rol como catálogo (RF-24) sin abrir un editor infinito de políticas |
| 03 | A Acceso | Must | RF-03 | parcial | Tabla `permisos`; `Matriz` y `Permite` en `accesos.go`; `exige` en `producto.go` | La matriz se borra y se reescribe al arrancar. Sin FK a un catálogo de roles. Labores (`handlers/operacion.go`) y geo no consultan `Permite`. La validación fina de la matriz sigue pendiente con el cliente; la semilla actual sí puede fijarse |
| 04 | B Catastro | Must | RF-04 | parcial | `areas_verdes`, `GET /api/v1/geo/areas`, ficha `PATCH /api/v1/catastro/areas/:id` | No hay ejemplar (código, especie, salud, coordenada). Los 1081 de la hoja gviz no se importan. Salud no viene en la hoja: ver datos a pedir. Georreferencia progresiva solo en áreas, no en labores |
| 05 | B Catastro | Should | RF-05 | falta | No hay columna de código anterior ni endpoint | Tabla de historial de códigos del ejemplar y recodificación que conserve el anterior |
| 06 | B Catastro | Must | RF-06 | parcial | `zonas` (534), catálogo `lugar` con 4 semillas, `005_producto.sql` | Zonas de supervisión Z1–Z4 sin cargar. Lugares 76 sin cargar. Texto libre en `actividades.zona_feature_id`, `solicitudes.lugar` y `riego_registros.sector`. Hace falta FK a catálogo |
| 07 | B Catastro | Should | RF-07 | falta | El plan de producto lo deja pendiente (`PLAN-PRODUCTO-Y-UI.md` §E ola 3) | Importador con vista previa para Excel/CSV/GeoJSON y exportación del inventario. La estructura del Excel de ejemplares ya se puede tomar de la hoja gviz (columnas en `MAPA-DATOS-Y-EDICION.md`) |
| 08 | C Actividades | Must | RF-08 | parcial | `actividades` + catálogo `tipo_actividad` de 5 códigos | Taxonomía real de 45 tipos / 7 clases. Sin personal, insumos, campos por tipo ni fechas de solicitud y atención. `geom` obligatoria impide alta solo por lugar |
| 09 | C Actividades | Must | RF-09 | hecho | `ejecutor` CHECK `propia\|tercerizada` en `005_producto.sql`; formulario en `panel/Labores.tsx` | Nada de este criterio mínimo. El «quién» nominal es el equipo (`capataces`), no una persona; eso se cierra con el modelo de cuadrilla anonimizada, no reabriendo este CHECK |
| 10 | C Actividades | Should | RF-10 | falta | No hay tabla de insumos | Consumo por actividad, espacio y periodo |
| 11 | C Actividades | Must | RF-11 | parcial | `solicitudes` y `GET/POST /api/v1/solicitudes` | Alta manual sí. No hay edición, ni las 283 labores 2026, ni poda 25 como incidencias. Integración automática Centuria/OSG: diferida y aprobada como captura manual |
| 12 | C Actividades | Must | RF-12 | parcial | `apps/web/src/offline/queue.ts`, uso en `App.tsx` | Cola de altas y estados. Faltan fotos, riego, reintento con error por registro, y rechazo si el UUID se reenvía con otro cuerpo |
| 13 | C Actividades | Must | RF-26 | parcial | `riego_registros`, `GET/POST /api/v1/riego`, `RiegoPanel` | Sector es texto. Sin ciclo, sin superficie, sin evitar doble conteo. El indicador oficial queda diferido; el registro por sector, turno y equipo debe poder editarse y cargarse |
| 14 | D Tercerizados | Must | RF-13 | parcial | `ordenes_servicio`, `GET/POST /api/v1/ordenes` | Falta conformidad usable, periodo, reporte del proveedor y regla visible de que la orden no cierra la atención por sí sola (el cierre de labor tercerizada ya exige orden; falta el flujo de estados de la solicitud) |
| 15 | D Tercerizados | Should | RF-14 | parcial | `evidencias.orden_id` en `005_producto.sql` | El `POST` de evidencias solo recibe `actividad_id` (`SubirEvidencia`). No adjunta a la orden |
| 16 | D Tercerizados | Should | RF-15 | diferido | El Excel dice «métricas por definir con el cliente» | No inventar KPI. Dejar el hueco rotulado hasta que jefatura defina fórmulas |
| 17 | D Tercerizados | Could | RF-27 | diferido | Fuera del MVP en el Excel y en la propuesta (CC-005) | Portal del proveedor no entra en este plan |
| 18 | E Seguimiento | Must | RF-16 | parcial | CHECK de estados en `003` y `005`; catálogo `estado` | Dos máquinas distintas (labor vs solicitud) y los estados de labor siguen hardcodeados. Falta impedir el cierre sin ejecución completa en todos los tipos, no solo la orden tercerizada |
| 19 | E Seguimiento | Should | RF-17 | falta | No hay avance por fecha ni checklist por área | Avances de varios días ligados al ejemplar o al área, como pide la propuesta §3 |
| 20 | E Seguimiento | Must | RF-18 | parcial | `GET .../timeline`, reporte filtrable por estado y fecha (`export.go`) | Filtros por ejemplar, zona, responsable y origen. Hoy la zona de la labor es texto |
| 21 | E Seguimiento | Must | RF-28 | hecho | `codigo_externo` único parcial; comentario en `005_producto.sql` | El sistema no genera el código. Falta poblar los que ya vienen en poda (`OSG-…`) al importar |
| 22 | F Evidencias | Must | RF-19 | parcial | `evidencias`, `POST /api/v1/evidencias`, `GET .../archivo` | Foto ligada al evento de trazabilidad, coordenadas opcionales, documentos de la orden, compresión y cola offline. Ver flujo en `MAPA-DATOS-Y-EDICION.md` |
| 23 | G Reportes | Must | RF-20 | parcial | `GET /api/v1/reportes/labores` | Solo nivel básico. Intermedio y avanzado esperan columnas que el Excel marca pendientes de validación: no inventarlas; sí dejar el gancho y exportar lo que ya está definido |
| 24 | G Reportes | Should | RF-21 | falta | No hay reportes de proceso | Depende de personal, horas y superficie, que hoy no existen. Mostrar el hueco, no un índice falso |
| 25 | G Reportes | Should | RF-22 | parcial | `CSV` y `ExcelXML` en `internal/atencion/export.go` | PDF no existe. Excel es SpreadsheetML, no xlsx. El mínimo del básico está; falta PDF cuando el contenido esté validado |
| 26 | G Reportes | Should | RF-23 | diferido | Aviso en la respuesta de riego (`handlers/producto.go`) | Indicadores oficiales pendientes de validación. Conteos operativos pueden mostrarse con esa frase; no hay meta |
| 27 | H Configuración | Must | RF-24 | parcial | `catalogos`, `GET/POST /api/v1/catalogos`, `POST .../desactivar` | Faltan clases del RF (plagas, frecuencias, roles, sedes) y los 45 tipos reales. No hay edición de nombre, solo alta y baja. Sin FK |
| 28 | I IA | Should | RF-25 | parcial | `internal/atencion/ia.go`, `POST /api/v1/ia/sugerir-tipo`, test `ia_test.go` | El caso no está cerrado con el cliente. Lo local cumple el límite de no sacar datos. No ampliar a API externa en este plan |
| 29 | J Mapa | Must | RF-29 | hecho | `apps/web/src/map/CampusMap.tsx`, `GET /api/v1/operacion/actividades` | Marcadores por estado y tipo, filtros, popup, sin Street View. Se mantiene |
| 30 | J Mapa | Must | RF-30 | parcial | `actividad_eventos`, `GET .../timeline` | El actor es `actor_rol`, no el usuario de sesión. El tipo `evidencia` está en el CHECK y no se llena al subir un archivo |
| 31 | J Mapa | Must | RF-31 | parcial | `POST /api/v1/operacion/actividades`, alta desde el mapa en `App.tsx` | Tipo de catálogo y pin sí. Capataz es selector de equipos semilla. Falta creación consecutiva verificada de punta a punta y permiso real (hoy vale el rol del cuerpo) |
| 32 | J Mapa | Must | RF-32 | parcial | Filtro de capataz en `handlers/operacion.go` `List` si hay cookie | Sin cookie, `?rol=capataz` es solo un filtro voluntario. Evidencias no entran en la cola offline. La lista cacheada sí (`saveLabores`) |
| 33 | J Mapa | Must | RF-33 | parcial | Capas en `apps/web/src/types.ts` `LAYERS`, `GET /api/v1/geo/zonas` | Se muestran 534 polígonos de jefe de grupo, no las 4 zonas de supervisión. El Excel pide validar nombres y límites con el cliente: los polígonos Z1–Z4 ya están publicados y se pueden cargar; la asignación a personas se anonimiza |
| 34 | J Mapa | Must | RF-34 | parcial | `POST .../archivar`, motivos semilla en `005_producto.sql` | Baja lógica sí. La lista definitiva de motivos sigue «pendiente de validación»: la semilla de cuatro motivos sirve hasta el acuerdo. No hay borrado físico |
| 35 | J Mapa | Should | RF-35 | hecho | `PATCH .../asignacion`, evento `reasignada` en `003_operacion.sql` | El cambio queda en la bitácora. Falta que el actor sea el usuario de sesión (se arrastra con HUID 30) |
| 36 | RNF | Must | RNF-01 | parcial | Misma cola que HUID 12; texto de pendientes en la PWA | Indicador sí, de forma básica. Falta la promesa verificable de DP2 §4 (corte, reapertura, pérdida de respuesta, fotos) |
| 37 | RNF | Must | RNF-02 | parcial | `vite-plugin-pwa` en `apps/web/vite.config.ts`, riel en `App.tsx` | PWA y riel responsive existen. No hay evidencia de prueba en Android ni tablet |
| 38 | RNF | Must | RNF-03 | parcial | Historia: tiempos de respuesta. Celda: «full cloud en AWS» | Hay `limit` en geo y un EC2 del Learner Lab (`docs/DEPLOY-AWS.md`). No hay paginación de listados grandes, ni meta de tiempo, ni hosting institucional. El lab no es el despliegue de producción |
| 39 | RNF | Must | RNF-04 | parcial | Historia: interfaz clara. Celda: autenticación, secretos, auditoría | La UI de guardia está (riel, español, sesión). No hay auditoría de operaciones críticas (tabla de cambios). Auth incompleta: ver HUID 01 |
| 40 | RNF | Must | RNF-05 | parcial | Historia: proteger comunicaciones. Celda: catálogos sin hardcode | HTTP, CORS `*`, sin HSTS ni cabeceras. Catálogos a medias: los CHECK de tipo y estado siguen en SQL |
| 41 | RNF | Must | RNF-06 | parcial | Historia: secretos fuera del código. Celda: interfaz de campo simple | `.env` no se versiona (`.env.example`, `.gitignore`). `user_data.sh.tftpl` incrusta `pando-local` y la clave de Postgres. Interfaz de campo: formularios existen; faltan controles grandes y mensajes de error por campo en el mapa |
| 42 | RNF | Should | RNF-07 | falta | Historia: logs. Celda: tiempos, paginación y caché | Solo log de Gin. Sin agregación ni alertas. Sin paginación de labores ni caché de lecturas |
| 43 | RNF | Should | RNF-08 | parcial | Historia: pruebas. Celda: capas separadas | Paquetes Go por módulo (monolito, como pide DP2: RNF-08 «sin añadir despliegues»). Pruebas: ver HUID 45 |
| 44 | RNF | Should | RNF-09 | parcial | Historia: español. Celda: Excel y REST | UI en español. Excel solo en el reporte básico. REST sí (`openapi.yaml`, incompleto respecto a sesión y reportes) |
| 45 | RNF | Must | RNF-10 | parcial | Historia: arquitectura por responsabilidades. Celda: pruebas web y móvil | Arquitectura de un solo API: sí. Pruebas: `go test` sin Postgres y un test Node. Sin e2e ni evidencia móvil |
| 46 | RNF | Should | RNF-11 | parcial | `CampusMap.tsx`, edificios OSM `GET /api/v1/geo/edificios` | OSM y capas sí. Ortofoto diferida (sin acuerdo con OSG). Tecnología de visor ya decidida: MapLibre, no queda pendiente |
| 47 | RNF | Must | RNF-12 | parcial | Historia: desplegar. Celda: mensajes en español | Español cubierto. Despliegue: Learner Lab, no infraestructura acordada con TI. `user_data_replace_on_change` impide un apply seguro |
| 48 | RNF | Must | RNF-13 | parcial | Historia: API definida. Celda: IA responsable | OpenAPI en `apps/api/openapi.yaml` no lista todas las rutas de `server.go`. La IA local no envía datos fuera: cumple el espíritu de la celda, no cierra el contrato |
| 49 | RNF | Must | RNF-14 | parcial | Historia: límites de IA. Celda: conservar trazabilidad | Trazabilidad de eventos sí, sin política de retención. IA con confirmación humana sí. Retención, RPO y volumen: pendientes con jefatura, no se inventan |

Conteo de la matriz: 4 hecho (09, 21, 29, 35), 2 diferido de producto (16, 17) más indicadores oficiales dentro de 26, 1 falta neta marcada falta en varias filas de Should/Must (05, 07, 10, 19, 24, 42), y el resto parcial. Ningún HUID Must está cerrado del todo salvo 09, 21, 29 y 35, y esos cuatro arrastran huecos de actor o de datos.

## 5. Checklist de salida a producción

No se marca nada como cumplido si el código no lo sostiene.

- [ ] Claves semilla fuera del artefacto: nada de `pando-local` ni `campus-lab` en imágenes ni user data.
- [ ] HTTPS en el mismo dominio para la PWA y la API; HTTP redirige.
- [ ] Cookie `Secure`, `HttpOnly`, `SameSite`, y CSRF en las mutaciones con cookie.
- [ ] CORS acotado al origen de la PWA, no `*`.
- [ ] Rate limit en `POST /api/v1/sesion` y en subida de archivos.
- [ ] Cabeceras: `Content-Security-Policy`, `X-Content-Type-Options`, `Referrer-Policy`, HSTS.
- [ ] Todas las mutaciones exigen sesión y `Permite`. Geo de escritura también.
- [ ] Postgres sin puerto público. SSH cerrado o restringido a una IP.
- [ ] Secretos en un almacén, no en `user_data` ni en el estado de Terraform en claro.
- [ ] Backup diario de PostGIS (`pg_dump` custom, con extensión) y prueba de restauración anotada.
- [ ] Versionado del volumen de evidencias (S3) y bloqueo de acceso público.
- [ ] `terraform apply` no reemplaza la EC2 (`user_data_replace_on_change` en false; cloud-init solo en el primer arranque; despliegue = pull de imágenes).
- [ ] Healthcheck con alerta si `/health` falla.
- [ ] Logs con identificador de petición, sin claves ni cuerpos de archivos.
- [ ] Migraciones solo hacia adelante, numeración reservada, sin `TRUNCATE` de tablas ya editadas.
- [ ] Importación con vista previa, informe de errores y lote reversible.
- [ ] Datos faltantes cargados y anonimizados (lista en `MAPA-DATOS-Y-EDICION.md`).
- [ ] Pruebas de integración PostGIS y un e2e del flujo entrar → editar área → importar → revertir.
- [ ] Documentación de operación: encender, backup, restaurar, rotar clave, leer logs.
- [ ] Acuerdo escrito de RPO/RTO, retención y de lo que la universidad debe entregar (sección 6).

## 6. Qué hay que pedir a la universidad

No se puede sacar de la web pública, o no se debe usar tal cual:

1. Hoja de reservas de jardines (el publicado responde 401). Mientras tanto, datos ficticios.
2. Definición oficial de cobertura de riego, indicadores y columnas de reportes intermedio/avanzado (el backlog las deja pendientes).
3. SSO PUCP, dominio, certificado y cuenta de nube fuera del Learner Lab.
4. Ortofoto y cartografía OSG, si se quieren capas distintas de las ya publicadas.
5. Estado de salud del ejemplar, si lo exigen: la hoja gviz no trae un campo de salud, solo «OBSERVACIÓN FEN 2026».
6. Confirmación de qué persona real corresponde a cada cuadrilla, por un canal interno. En el sistema se guardan nombres ficticios.
7. Fotos que dejen de ser públicas en Drive (bebederos: hay índice de 70 ids en `data/raw/appscript/drive_fotos_api.json` y no hay JPEG en el repo).
8. GPS real de barredoras. Lo embebido en `data/raw/legacy-app/script.js` es una simulación, no una ruta operativa.
9. RPO, RTO y retención.
10. Fichas de poda enlazadas a documentos privados, si el enlace de Drive no abre.

Detalle de fuentes y campos: `docs/MAPA-DATOS-Y-EDICION.md`. Orden de construcción: `docs/PLAN-MULTIAGENTE.md`.
