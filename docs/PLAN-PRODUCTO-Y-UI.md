# Plan de producto y de interfaz — Campus Verde

Fecha: 2026-09-24. Fuentes: `docs/ARQUITECTURA.md`, `docs/PLAN-INTEGRACION-MAPA.md`, backlog `Product_Backlog_Areas_Verdes` (CSV/XLSX) y las capturas del visor en `127.0.0.1:4317`.

Este documento se escribe **antes** de la ola de código. Las casillas de abajo se marcan al implementar.

El CSV desalinea, a partir de HUID 38, la columna *Funcionalidades* respecto de *Como / Quiero*. La matriz usa el HUID, el código RF impreso y el texto de la historia. Cuando la celda de funcionalidades dice otra cosa, se anota.

## A. Análisis de la UI actual

Capturas: plano de escritorio, relieve, popup de bebedero `PT_bb1`. Código: `apps/web/src/styles.css`, `App.tsx`, un solo riel.

**Lo que ya es de operación.** El mapa ocupa el lienzo. Hay 521 áreas y 534 zonas servidas en vivo, marcadores con letra de tipo y color de estado, filtros de labor, alta por pin, relieve con extrusión y un popup de inventario que dice la verdad («Sin fotografía recuperada»). No hay hero ni grilla de tarjetas. Eso se conserva.

**Jerarquía.** Compiten tres cosas en la misma barra: la marca en versalitas, Plano/Relieve y el rol (Jefatura / Coordinación / Capataz). El rol no es una sesión: es un interruptor del navegador. El panel izquierdo apila labores, ficha, catastro e inventario en un solo scroll. En la captura de relieve, la ficha de «Control de borde» empuja el resto hacia abajo. No hay lugar para solicitudes, reportes ni catálogos: no existen como módulos.

**Densidad.** La lista de labores es densa y legible, bien para supervisión. El resto del panel es un inventario de checkboxes con la misma fila repetida. Sirve para capas; no sirve como aplicación de varios módulos.

**Tipo.** IBM Plex Sans y Mono. El cuerpo está bien. El problema es el uso: «CAMPUS VERDE», «LABORES», «CATASTRO», «INVENTARIO» en mayúsculas con tracking. Es el tic de plantilla, no una necesidad de campo. Etiquetas a 11 px en mono se leen en la captura de escritorio y se pierden en móvil.

**Color.** Tokens actuales: papel `#f3f0e8`, ocre `#a8843d`, verde `#1e4d3a`, tinta `#1c211e`. Es el par crema + terracota que este plan retira como marca. El ocre puede quedar solo como color de estado «pendiente», no como segundo color de marca. El panel cálido sobre el mapa OSM se ve de lámina, no de mesa de guardia.

**Chrome.** Barra de 56 px translúcida y riel flotante con sombra. Correcto que no tape el campus; incorrecto que todo viva en ese riel. No hay navegación de módulos. Plano/Relieve está bien resuelto como par.

**Accesibilidad.** Contraste de tinta sobre crema es alto. El texto apagado `#5c6560` sobre crema pasa en cuerpo y falla en hints de 11 px. Los checks de estado son pequeños para dedo. El rol no se anuncia como sesión. No hay estados vacíos propios: si no hay labores, una línea gris. El fallo de WebGL2 ya tiene un mensaje; el resto de errores de API cabe en una sola línea de estado.

**Genérico frente a operación.** El mapa es específico (Pando, nombres de zonas, equipos Norte/Sur/Riego). El chrome no: versalitas, papel cálido, ocre, un panel para todo. Falta la estructura de un sistema de guardia (módulo activo, usuario real, cola offline visible).

## B. Sistema visual (brief)

Seis tokens. Nada de crema ni ocre de marca.

| Nombre | Hex | Uso |
|--------|-----|-----|
| Tinta | `#141816` | Riel, texto, mapa de controles |
| Campo | `#e7ece6` | Panel del módulo |
| Niebla | `#f4f7f4` | Popups y campos |
| Dosel | `#145c3e` | Acción primaria y módulo activo |
| Alerta | `#8d2f2c` | Error y estado bloqueada |
| Línea | `#c5cfc8` | Divisores |

Estado, solo en el dato: pendiente `#8a6a12`, en proceso `#145c3e`, bloqueada `#8d2f2c`, cerrada `#5c6560`. No pintan el fondo de la app.

**Tipo.** [Source Sans 3](https://fontsource.org/fonts/source-sans-3) para interfaz (13–15 px, peso 400/600). [Newsreader](https://fontsource.org/fonts/newsreader) solo en el nombre «Campus Verde», una línea, sentencia. No IBM Plex como sistema, no Inter, no versalitas de sección.

**Layout.** Mapa a sangre. A la izquierda, el **riel de guardia** (firma): 56 px, fondo tinta, módulos en columna — Mapa, Labores, Catastro, Solicitudes, Reportes, Catálogos, Admin. El panel del módulo activo (320 px, fondo campo) se abre al lado y se cierra en móvil con un botón. Plano/Relieve vive sobre el mapa, no junto al usuario. El usuario es una sesión abajo del riel (nombre de equipo, salir).

**Firma única.** El riel oscuro. Un solo acento dosel. Si una pantalla nueva pide otra tarjeta, otra ceja o otro gradiente, no entra.

**Autochequeo anti-slop.**

- Crema `#f3f0e8` y ocre `#a8843d` dejan de ser marca.
- No hay grilla de tarjetas ni hero.
- No hay Inter ni púrpura.
- Los módulos van en sentencia, no en mayúsculas espaciadas.
- El mapa sigue siendo la vista de entrada tras entrar.

## C. Matriz del backlog

Prioridad del CSV. Ola: (1) base operativa, (2) atención, (3) priorizadas. «Hecho» es lo que el repo demuestra hoy (fases A–F del plan de mapa), no el cierre del HUID.

| HUID | RF | Must | Estado | Ola | Nota |
|------|----|------|--------|-----|------|
| 01 | RF-01 | MUST | parcial | 1 | Interruptor local de rol. Falta usuario, clave y sesión. SSO real queda fuera. |
| 02 | RF-02 | MUST | parcial | 1 | Tres roles en el cliente. Falta Admin y roles persistidos. Operario sin cuenta: se mantiene. |
| 03 | RF-03 | MUST | falta | 1 | Sin matriz de permisos. Semilla por rol, no editor infinito. |
| 04 | RF-04 | MUST | parcial | 1 | 521 áreas y overlays de lectura. Falta ficha, edición de metadatos y ejemplar. Geom ya admite NULL. |
| 05 | RF-05 | SHOULD | falta | 3 | Recodificar conservando el código anterior. |
| 06 | RF-06 | MUST | parcial | 1 | Zonas cargadas (534) sin catálogo editable. Lugares entran como catálogo semilla. |
| 07 | RF-07 | SHOULD | falta | 3 | Import/export de inventario. Definición del Excel pendiente; esta ola no inventa el archivo. |
| 08 | RF-08 | MUST | parcial | 1 | Alta por pin con tipo fijo. Faltan clase, responsable real, insumos y campos por tipo. |
| 09 | RF-09 | MUST | falta | 1 | Propia vs tercerizada no existe en el modelo. |
| 10 | RF-10 | SHOULD | falta | 3 | Insumos por actividad. No bloquea la base. |
| 11 | RF-11 | MUST | falta | 2 | Solicitudes manuales (Centuria / OSG / correo) sin API externa. |
| 12 | RF-12 | MUST | parcial | 1 | Cola IndexedDB solo del alta. Falta lista offline del capataz y reintento de estado/evidencia. |
| 13 | RF-26 | MUST | falta | 1 | Riego por sector, turno y capataz. Indicador oficial de cobertura: pendiente, no se inventa. |
| 14 | RF-13 | MUST | falta | 2 | Orden de servicio tercerizado mínima. |
| 15 | RF-14 | SHOULD | falta | 2 | Evidencia de proveedor reutiliza el stub de evidencias si la orden existe. |
| 16 | RF-15 | SHOULD | falta | — | Métricas de proveedor: el propio backlog dice que no están definidas. No se fabrican. |
| 17 | RF-27 | COULD | falta | — | Portal del proveedor. Fuera del MVP. |
| 18 | RF-16 | MUST | parcial | 2 | Estados fijos en código. Catálogo en ola 1; flujo de cierre con tercerización en ola 2. |
| 19 | RF-17 | SHOULD | parcial | 2 | El cambio de estado existe. Falta avance por área. |
| 20 | RF-18 | MUST | parcial | 2 | Bitácora por labor. Falta historial filtrable (zona, origen, fecha). |
| 21 | RF-28 | MUST | falta | 2 | Código externo conservado, no generado. |
| 22 | RF-19 | MUST | falta | 1 | Stub: metadatos + archivo en disco local. Sin bucket S3 real. |
| 23 | RF-20 | MUST | falta | 2 | Un reporte básico. Intermedio/avanzado no cierran columnas (pendiente de validación). |
| 24 | RF-21 | SHOULD | falta | 3 | Reportes de proceso. Dependen de columnas no validadas. |
| 25 | RF-22 | SHOULD | parcial | 2 | CSV y Excel mínimos del reporte básico. PDF no en esta pasada. |
| 26 | RF-23 | SHOULD | falta | 3 | Indicadores: etiqueta explícita de definición pendiente. Sin número oficial inventado. |
| 27 | RF-24 | MUST | falta | 1 | CRUD de catálogos con baja lógica. |
| 28 | RF-25 | SHOULD | falta | 3 | Una función de IA del curso, etiquetada, sin datos personales hacia fuera. |
| 29 | RF-29 | MUST | hecho | — | Mapa, filtros, popup, sin Street View. Se mantiene al cambiar el chrome. |
| 30 | RF-30 | MUST | parcial | 1 | Cadena de eventos sin usuario de sesión ni evidencia por evento. |
| 31 | RF-31 | MUST | parcial | 1 | Pin + capataz. Falta catálogo de tipos y alta consecutiva (el formulario se limpia y permanece). |
| 32 | RF-32 | MUST | parcial | 1 | Capataz ve solo lo asignado. Offline de lista y evidencias, incompleto. |
| 33 | RF-33 | MUST | hecho | — | Zonas y áreas como capas. |
| 34 | RF-34 | MUST | parcial | 1 | Archivo lógico sin motivo de catálogo. |
| 35 | RF-35 | SHOULD | hecho | — | Reasignación con evento en la bitácora. |
| 36 | RNF-01 | MUST | parcial | 1 | Misma deuda que HUID 12. Indicador de pendientes visible. |
| 37 | RNF-02 | MUST | parcial | 1 | El riel colapsa en ancho estrecho. Revisar tras el rediseño. |
| 38 | RNF-03 | MUST | parcial | — | La historia pide tiempo de respuesta; la celda de funcionalidades dice AWS. Despliegue cloud diferido. Paginación donde el listado pase de unos cientos. |
| 39 | RNF-04 | MUST | parcial | 1 | La historia pide interfaz clara (este rediseño). La celda habla de auth: se cubre con la sesión de ola 1, no con SSO. |
| 40 | RNF-05 | MUST | parcial | 1 | Historia: proteger comunicaciones. Celda: catálogos sin hardcode (ola 1). Secreto solo en `.env`. |
| 41 | RNF-06 | MUST | parcial | 1 | Historia: secretos fuera del repo (ya `.env`). Celda: interfaz de campo simple (este rediseño). |
| 42 | RNF-07 | SHOULD | falta | — | Historia: logs. Celda: tiempos. Se mantiene el log de Gin. Sin plataforma de observabilidad. |
| 43 | RNF-08 | SHOULD | parcial | 1 | Hay `go test` de API. Celda pide capas separadas: se respeta el monolito modular. |
| 44 | RNF-09 | SHOULD | hecho | — | UI en español. Celda pide Excel: el stub de ola 2. |
| 45 | RNF-10 | MUST | parcial | 1 | Arquitectura de un solo API. Pruebas nuevas de accesos y catálogos. |
| 46 | RNF-11 | SHOULD | hecho | — | OSM + zonas. Celda dice «interfaz en español», ya cubierta. Ortofoto no. |
| 47 | RNF-12 | MUST | falta | — | Despliegue institucional. Celda habla de IA responsable: ver HUID 49. |
| 48 | RNF-13 | MUST | parcial | 1 | REST/JSON existente. OpenAPI se extiende con lo que se agregue. |
| 49 | RNF-14 | MUST | parcial | 3 | Trazabilidad de labores ya persiste. IA con supervisión humana y sin enviar PII fuera. |

### Diferido, con motivo

| Tema | Motivo |
|------|--------|
| SSO PUCP | HUID 01: pendiente de validación con el cliente. Sesión propia de desarrollo. |
| API automática Centuria | HUID 11 y 21: captura manual aprobada. |
| Portal del proveedor (RF-27) | Could, fuera del MVP. |
| Definiciones oficiales de indicadores y de cobertura de riego | HUID 13, 16, 23, 26. Se muestra el hueco, no un KPI falso. |
| Ortofoto, GeoServer, bucket S3 de producción, AWS | Sin acuerdo de TI. Disco local para evidencias. |
| PDF de reportes, insumos, recodificación, import Excel de inventario | Should de ola 3 o explícitamente no cerrados. |
| React Three Fiber / maqueta fotoreal | Sección D. No en esta pasada. |

## D. 3D

La extrusión de MapLibre (`fill-extrusion` sobre áreas, huellas OSM en `data/osm/edificios_pando.geojson`) sigue siendo el 3D del producto. Plano no se rompe.

React Three Fiber queda como maqueta opcional posterior, cuando exista un glTF del campus. Esta pasada no lo agrega y no fabrica un campus fotorrealista sin ese archivo.

## E. Olas de implementación

Misma API Go y la misma PWA. Profundidad honesta: cada módulo lista, crea o filtra contra Postgres. Lo que no está se dice en la pantalla.

### Ola 1 — Base operativa

- [ ] Riel de guardia, tokens nuevos, tipo Source Sans 3 + Newsreader, módulos en sentencia.
- [ ] Sesión: usuarios semilla (capataz de cada equipo, coordinación, jefatura, admin), clave solo de desarrollo, cookie HttpOnly. El interruptor de rol sale.
- [ ] Permisos semilla por rol (consultar, registrar, validar, reportes, catálogos, solicitudes).
- [ ] Catálogos RF-24: tipos, estados, prioridades, lugares, especies, motivos de archivo. Alta y baja lógica.
- [ ] Catastro: lista, ficha, edición de metadatos, registro sin geometría.
- [ ] Labor: ejecutada por personal propio o tercerizado; motivo al archivar; alta consecutiva; tipos desde catálogo.
- [ ] Riego mínimo: sector, turno, equipo, fecha, nota. Sin porcentaje oficial.
- [ ] Offline del capataz: lista de asignadas en IndexedDB y cola de altas y cambios de estado, con UUID.
- [ ] Evidencia: metadatos + archivo en disco local, ligada a la labor.
- [ ] Mapa CORE intacto (capas, marcadores, bitácora, capataz).

### Ola 2 — Atención

- [ ] Solicitudes e incidencias: fuente, código externo, prioridad, estado, vínculo opcional a labor.
- [ ] Orden tercerizada mínima (empresa, referencia, estado) ligada a una labor tercerizada.
- [ ] Historial filtrable y un reporte básico con exportación CSV y Excel.
- [ ] Cierre: una labor tercerizada no pasa a cerrada sin orden registrada.

### Ola 3 — Priorizadas, delgadas

- [ ] IA etiquetada: sugerir el tipo de labor a partir del título, en proceso, sin red externa ni PII. La persona confirma.
- [ ] Indicadores: conteos operativos con la frase de definición pendiente. Sin meta oficial.
- [ ] Pulido del mapa (filtros ya existentes, evidencia visible en la bitácora).
- [ ] Import de inventario: no se inventa el Excel. Queda nombrado como pendiente de estructura.

### Verificación

`make bootstrap`, `make api`, `make web`. Pruebas Go de sesión y catálogos. Recorrido en el navegador: entrar, mapa, labor, catastro, solicitud, reporte, catálogo. Capturas antes/después en artefactos. OpenAPI y README en español al día.

## F. Cómo se lee esto junto al plan de mapa

`docs/PLAN-INTEGRACION-MAPA.md` cierra las fases A–F del visor (API, catastro, PWA, labores, relieve, inventario). Este documento no las reabre: las usa como base y marca lo que el backlog todavía no tiene (acceso real, catálogos, solicitudes, reportes, evidencias, riego).
