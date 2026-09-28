# Auditoría del Product Backlog v2 frente al v1 y a la app actual

> **Alcance.** Compara `docs/fuente/Product_Backlog_Areas_Verdes.xlsm` (v1) con `docs/fuente/Product_Backlog_Areas_Verdes_v2.xlsm` (v2) y contrasta cada RF/RNF de v2 con el código de la rama `docs/fase1-arquitectura-backlog` (base `115a56d`). También cruza los roles con la arquitectura objetivo del equipo (`GRUPO-12-DP2/-areas-verdes-pucp`, rama `init/backend`, archivo `db/schema_nucleo_v0.2.sql`).
>
> **Método.** Se extrajo el texto de ambas hojas (v1: hoja `Product Backlog`, 54×12; v2: hoja `Copia de Product Backlog`, 1000×27, de la que solo 14 columnas tienen contenido). Ninguno de los dos libros tiene comentarios de celda. La cobertura se verificó leyendo rutas, handlers, stores, migraciones y componentes de la PWA. Cuando algo no se pudo comprobar en el código, se indica.
>
> **Leyenda de cobertura.** **Sí** = implementado de punta a punta (API + BD + UI cuando aplica). **Parcial** = existe una base, pero falta parte del texto de v2 o solo existe en una capa. **No** = no hay implementación.

---

## 1. Resumen ejecutivo

- **El SSO queda descartado.** RF-01 v2 aprueba "credenciales propias administradas por el jefe de sección (sin SSO)" (3.ª reunión). Esto confirma el modelo de cuentas locales de `apps/api/internal/accesos/accesos.go`, pero la app **no tiene alta, baja ni cambio de clave de usuarios**: solo trae semillas y una lista de solo lectura (`GET /api/v1/accesos/usuarios`, solo para `admin`).
- **Se reorganiza la épica J (mapa/CORE).** Cambian de número RF-30 y RF-31 (registro ↔ trazabilidad). La baja lógica (v1 RF 34) y la reasignación (v1 RF 35) se integran en **RF-31**. Los números RF-34 y RF-35 se reutilizan para **requisitos nuevos**: sectores de capataz con vigencia y sugerencia automática (Should), y vista de concentración de intervenciones (Could). Comparar por número induce a error.
- **RF-30 amplía mucho el alta desde el mapa.** Ahora exige un tipo en dos niveles (eje + subtipo), ejecutor, origen, código externo, unidad solicitante, nivel de riesgo, fecha programada y cantidad pedida. La app solo captura tipo (un nivel), título, detalle, capataz y ejecutor (`apps/api/internal/handlers/operacion.go`, `createBody`).
- **RF-31 amplía la cadena de eventos.** Añade inicio, actualizaciones intermedias, supervisión, derivación a tercerizado, observación/subsanación y conformidad. La BD solo admite `creada, asignada, reasignada, estado, cancelada, archivada, evidencia` (CHECK en `migrations/005_producto.sql`).
- **RF-32 es nuevo como requisito de visibilidad y filtros por rol.** Suma filtros por sector, ejecutor, origen, nivel de riesgo y fechas. La UI solo filtra por estado, tipo y equipo (`apps/web/src/panel/Labores.tsx`). La API ya admite `zona_supervision_id`, `cuadrilla_id` y `origen`, pero la UI no los usa.
- **La zonificación cambia de modelo (RF-06).** Pasa a "sectores de capataz, lugares/jardines y referentes (edificios, vías)", y los cuarteles quedan como referencia histórica. RF-33 (capas) **baja de Must a Should** y pasa a cargarse desde shapes/GeoJSON del cliente.
- **Se precisa la taxonomía de actividades (RF-08).** Aparecen "habilitación", "manejo fitosanitario" e "inspección y monitoreo". Se retira del texto la obligación explícita de "campos específicos por tipo", aunque la observación la mantiene como pendiente. En BD, `clase_actividad` no incluye fitosanitario ni inspección (`migrations/020_labores.sql`).
- **RF-24 retira del texto los "motivos de cancelación"** y los deja como pendientes de catalogar. La app **ya** tiene la clase `motivo_archivo` con los 4 motivos (`migrations/005_producto.sql`).
- **Los RNF de v2 corrigen la desalineación de v1.** En v1, 11 de las 14 filas RNF tenían un "Como/Quiero" que no correspondía a su RNF. Además, RNF-11 (mapa) **sube de Should a Must** y RNF-03 deja de decir "full cloud en AWS" para decir "nube según restricciones institucionales".
- **Brechas de permisos en el código actual:**
  - El capataz puede poner cualquier estado (incluido `cerrada`/`cancelada`) en sus labores.
  - Puede editar la ficha y registrar avances de labores **ajenas**, y crear especies, lugares y cuadrillas.
  - La jefatura no puede subir evidencias ni registrar riego.
  - Hay endpoints de lectura **sin autenticación**: `GET /api/v1/operacion/actividades`, `/timeline`, `/capataces` y `/api/v1/geo/*`.
- **Los estados no están parametrizados de verdad.** El catálogo `estado` existe, pero `SetEstado` acepta cualquier *slug* sin validarlo (`apps/api/internal/operacion/validate.go:102`). La web tiene los estados codificados en duro (`apps/web/src/operacion.ts:21`). Los nombres tampoco coinciden con v2: v2 dice *Por iniciar / En proceso / Ejecutado / Cerrado / Cancelado / Archivado*, y la app usa *pendiente / en_proceso / bloqueada / cerrada / cancelada*.
- **No existen:** insumos (RF-10), métricas de proveedores (RF-15), indicadores (RF-23), reportes intermedio/avanzado (RF-20/21), exportación a PDF (RF-22), sugerencia de sector/capataz por ubicación (RF-34), vista de concentración (RF-35) y portal del proveedor (RF-27).

---

## 2. Nota sobre la numeración y la alineación v1 ↔ v2

### 2.1 Convenciones

| Aspecto | v1 | v2 |
|---|---|---|
| Historia | `HUID nn`, una por fila (HUID 01–49) | `HU-nn`, **reagrupadas**: una fila puede citar varias HU (`HU-07, HU-09, HU-21, HU-31`) y una HU aparece en varias filas |
| Requisito | `RF nn`, `RNF nn` (con espacio) | `RF-nn`, `RNF-nn` (con guion) |
| Columna "Estado" | Mezcla estado y observaciones | Pasa a **"Observaciones"**; se añaden **"Dependencias con otros RF"** y una columna **"Estado" vacía** |
| Épica J | "Módulo CORE — Gestión, supervisión y trazabilidad por mapa" | "Mapa de supervisión y trazabilidad de actividades" |

**Consecuencia:** v2 no define los títulos de las HU. Las referencias `HU-01…HU-34` apuntan a un "Catálogo v0.4" que no está en `docs/fuente/`, así que **no se puede verificar a qué historia corresponde cada HU-nn**. La alineación de este documento se hace **por RF/RNF y por contenido**, no por el número de HU.

Algunas reagrupaciones de v2 que conviene revisar con el equipo:

- **HU-12** cubre RF-09, RF-13 y RF-14.
- **HU-14** cubre RF-15 (métricas) y RF-17 (avance).
- **HU-19** cubre RF-06, RF-24 y RNF-05.
- **HU-30** cubre RF-32, RNF-02 y RNF-06.
- **RF-03** no tiene HU propia (usa `HU-01, HU-02`, igual que RF-01/RF-02).
- Diez RNF tienen HU "—".

### 2.2 Errores de v1 (filas RNF desalineadas)

En v1, los textos "Como / Quiero / Para" de las filas HUID 38–49 están **desplazados** respecto de la columna RNF. Se toma como verdadero el texto de la columna RNF, porque la prioridad de la fila coincide con él:

| Fila v1 | "Quiero…" (historia) | RNF de la fila | ¿Coinciden? | La historia corresponde a… |
|---|---|---|---|---|
| HUID 36 | usar las vistas de campo sin conexión | RNF 01 offline | Sí | — |
| HUID 37 | usar la app desde Android, tablet y PC | RNF 02 PWA | Sí | — |
| HUID 38 | contar con tiempos de respuesta adecuados | RNF 03 nube AWS | **No** | RNF 07 |
| HUID 39 | contar con una interfaz clara y consistente | RNF 04 seguridad | **No** | RNF 06 |
| HUID 40 | proteger comunicaciones y credenciales | RNF 05 parametrizable | **No** | RNF 04 |
| HUID 41 | mantener los secretos fuera del código | RNF 06 interfaz de campo | **No** | RNF 04 |
| HUID 42 | contar con registros de errores y eventos | RNF 07 rendimiento | **No** | *ningún RNF* (observabilidad/logs) |
| HUID 43 | contar con pruebas automatizadas | RNF 08 capas | **No** | RNF 10 |
| HUID 44 | usar la aplicación en español | RNF 09 interoperabilidad | **No** | RNF 12 |
| HUID 45 | mantener una arquitectura organizada | RNF 10 pruebas | **No** | RNF 08 |
| HUID 46 | usar funciones de mapa | RNF 11 mapa | Sí | — |
| HUID 47 | desplegar en la infraestructura definida | RNF 12 español | **No** | RNF 03 |
| HUID 48 | exponer y consumir servicios mediante una interfaz | RNF 13 IA responsable | **No** | RNF 09 |
| HUID 49 | recibir resultados de IA con límites | RNF 14 trazabilidad | **No** | RNF 13 |

Efectos:

- En v1, **RNF 05 (parametrizable)** y **RNF 14 (trazabilidad histórica)** no tenían historia propia. v2 las corrige ("Administrador del sistema / ajustar reglas sin modificar código" y "Supervisor / conservar el historial completo").
- La historia **"registros de errores y eventos" (HUID 42)** no tiene RNF en v1 y **desaparece en v2**. La cobertura real se comenta en §6.
- v2 fusiona las dos historias de seguridad (HUID 40 + 41) en RNF-04.

### 2.3 Otras desalineaciones e inconsistencias

- **v1 HUID 17 / RF 27** no tenía "Como/Quiero/Para". v2 la completa con "Ingeniero de campo de empresa proveedora autorizada".
- **RF 26 (riego)** está numerado fuera de secuencia en las dos versiones (va dentro de la épica C, después de RF-12).
- **La ejecución del capataz queda implícita en v2.** v1 RF 32 decía "registra actualizaciones/evidencias avanzando estados". v2 RF-32 solo habla de **visibilidad**, y la ejecución solo se deduce de RF-12, RF-19 y RF-31 (vía HU-31/HU-32). Conviene explicitarla.
- **RNF-11 v2 se contradice.** El texto dice "proveedor por definir (Google Maps API frente a Leaflet/MapLibre)", la observación dice "Definido (base OSM)… actualizar con la decisión OSM", y la app ya usa MapLibre + OSM (`apps/web/src/map/CampusMap.tsx`).
- **RF-04 v2** tiene la observación truncada ("Definido · Datos maestros", sin "Pendiente de validación").
- **RNF-14 v2** dice en su observación "No figura en el Catálogo v0.4: incorporar como RNF-14 o retirar". Está pendiente de decisión.
- **RF-31 v2** dice en su observación que la reasignación tras iniciar está "definida en backlog v1.0, pero el Catálogo v0.4 aún la lista como pendiente".

---

## 3. Diff por requisito (alineado por contenido)

Tipos de cambio: **Sin cambio** (solo redacción), **Modificada** (cambia el alcance), **Renumerada**, **Nueva**, **Eliminada/Absorbida**.

| v1 | v2 | Tipo | Prioridad v1 → v2 | Resumen del cambio de texto |
|---|---|---|---|---|
| RF 01 | RF-01 | Modificada (observación) | Must → Must | Se resuelve "SSO vs credenciales propias": **credenciales propias administradas por el jefe de sección, sin SSO** (aprobado). |
| RF 02 | RF-02 | Sin cambio | Must → Must | Añade "(campo)" y "(oficina)"; "operarios sin cuenta **individual para registrar labores**". |
| RF 03 | RF-03 | Sin cambio | Must → Must | "…y, **cuando corresponda**, gestionar solicitudes/incidencias". Matriz pendiente. |
| RF 04 | RF-04 | Sin cambio (menor) | Must → Must | "ejemplares **(arbolado)**"; quita "(georreferenciación progresiva)". |
| RF 05 | RF-05 | Sin cambio | Should → Should | — |
| RF 06 | RF-06 | **Modificada** | Must → Must | De "cuarteles, sectores, lugares" a "**sectores de capataz, lugares/jardines y referentes (edificios, vías)**"; **cuarteles solo como referencia histórica** (validado). |
| RF 07 | RF-07 | Sin cambio | Should → Should | "exportar **información**". |
| RF 08 | RF-08 | **Modificada** | Must → Must | Nueva lista de ejes: mantenimiento de jardines, poda, riego, propagación **y plantación**, **habilitación**, manejo de residuos **vegetales**, manejo fitosanitario, **inspección y monitoreo**. Se quita del texto "Debe permitir registrar los campos específicos de cada tipo" (sigue en la observación). |
| RF 09 | RF-09 | Sin cambio | Must → Must | "diferenciando el **tipo de** ejecutor". |
| RF 10 | RF-10 | Sin cambio | Should → Should | "por **intervención**". |
| RF 11 | RF-11 | Modificada (menor) | Must → Must | Añade "cantidad **solicitada/ejecutada** cuando aplique", "observaciones", "estado **de atención**". |
| RF 12 | RF-12 | Sin cambio | Must → Must | — |
| RF 26 | RF-26 | Sin cambio | Must → Must | "sector/turno". |
| RF 13 | RF-13 | Sin cambio | Must → Must | "atención **realizada**". |
| RF 14 | RF-14 | Sin cambio | Should → Should | — |
| RF 15 | RF-15 | Sin cambio | Should → Should | — |
| RF 27 | RF-27 | Modificada (historia) | Could → Could | Se completa la historia: "Ingeniero de campo de empresa proveedora autorizada". Sigue fuera del MVP. |
| RF 16 | RF-16 | Sin cambio | Must → Must | Redacción: "el cierre ocurre después de ejecutar el servicio". |
| RF 17 | RF-17 | **Modificada** | Should → Should | Se quita del texto "avance/checklist por jardín o área"; pasa a la observación como pendiente. |
| RF 18 | RF-18 | Sin cambio | Must → Must | "historial de **intervenciones y atenciones**". |
| RF 28 | RF-28 | Sin cambio | Must → Must | Redacción: "sin asumir que el nuevo sistema genera dicho código". |
| RF 19 | RF-19 | **Modificada** | Must → Must | Se quita del texto "a un evento de su trazabilidad"; la observación lo remite a RF-31. |
| RF 20 | RF-20 | Sin cambio | Must → Must | Añade "el contenido de cada nivel debe validarse". |
| RF 21 | RF-21 | Sin cambio | Should → Should | "rendimiento por espacio **o actividad**", "trazabilidad **de atenciones**". |
| RF 22 | RF-22 | Sin cambio | Should → Should | "**cuando corresponda**". |
| RF 23 | RF-23 | Modificada (menor) | Should → Should | "cobertura de **mantenimiento**/riego", "**frecuencia**/cumplimiento". |
| RF 24 | RF-24 | **Modificada** | Must → Must | Se **quitan "motivos de cancelación"** de la lista (quedan pendientes en la observación). |
| RF 25 | RF-25 | Modificada (menor) | Should → Should | Lista candidatos: clasificación por imágenes, reportes asistidos, consulta en lenguaje natural. |
| RF 29 | RF-29 | **Modificada** | Must → Must | "Campus Pando"; el icono es el **eje** de actividad; resumen "con acceso al detalle sin abandonar el mapa". El **panel de filtros sale** a RF-32. |
| RF 31 | **RF-30** | **Renumerada + Modificada** | Must → Must | Añade tipo en **2 niveles (eje/subtipo)**, **ejecutor, origen, código externo, unidad solicitante, nivel de riesgo, fecha programada, cantidad pedida**; alta consecutiva "sin volver al mapa". |
| RF 30 | **RF-31** | **Renumerada + Modificada** | Must → Must | Más eventos: registro, inicio, actualizaciones intermedias, **supervisión**, reasignación, **derivación a tercerizado**, **observación/subsanación**, **conformidad**, cierre. Absorbe baja lógica "previa confirmación" (v1 RF 34) y reasignación (v1 RF 35, en la observación). |
| RF 32 | RF-32 | **Modificada** | Must → Must | Pasa a "**visibilidad y filtros por rol**": oficina ve todo y filtra por estado, tipo, capataz, **sector, ejecutor, origen, riesgo, fechas**; el capataz ve lo suyo en mapa **o** lista. Se pierde la mención explícita a "registra actualizaciones/evidencias avanzando estados". Rol: "Supervisor / Capataz". |
| RF 33 | RF-33 | **Modificada** | **Must → Should** | De "zonas/sectores con nombre y límites" a "**capas de referencia**: sectores de capataz, jardines del inventario y de préstamo, tipos de uso, inventario forestal, cuarteles históricos", **desde shapes/GeoJSON** del cliente. |
| RF 34 | (RF-31) | **Absorbida** | Must → Must (dentro de RF-31) | Baja lógica con motivo predefinido. La lista de motivos queda pendiente de validación. |
| RF 35 | (RF-31, observación) | **Absorbida** | **Should → Must** (implícito, al integrarse en RF-31) | Reasignación registrando anterior y nuevo. Aparece solo en la observación. |
| — | **RF-34** | **Nueva** | — → Should | Sectores de capataz como **polígonos con vigencia**, reasignables (p. ej. por obras), y **sugerencia automática de sector, capataz y referente** según el punto marcado. |
| — | **RF-35** | **Nueva** | — → Could | Vista de **concentración de intervenciones** por sector/jardín en un periodo (incluye cerradas). Fuera del MVP. |
| RNF 01 | RNF-01 | Sin cambio | Must → Must | "indicador **visible** de pendientes". |
| RNF 02 | RNF-02 | Sin cambio | Must → Must | — |
| RNF 03 | RNF-03 | **Modificada** | Must → Must | De "full cloud en AWS" a "**desplegable en nube** según arquitectura y restricciones institucionales" (la observación mantiene AWS). |
| RNF 04 | RNF-04 | Sin cambio | Must → Must | Historia corregida (fusiona HUID 40 y 41). |
| RNF 05 | RNF-05 | Modificada (menor) | Must → Must | Detalla qué no debe estar codificado en duro: clases/tipos, estados, prioridades, frecuencias, lugares, roles. Historia nueva. |
| RNF 06 | RNF-06 | Sin cambio | Must → Must | Historia corregida (Capataz). |
| RNF 07 | RNF-07 | Sin cambio | Should → Should | "en las operaciones principales". |
| RNF 08 | RNF-08 | Sin cambio | Should → Should | Historia corregida. |
| RNF 09 | RNF-09 | Sin cambio | Should → Should | "formatos existentes (especialmente Excel)". |
| RNF 10 | RNF-10 | Modificada (menor) | Must → Must | Se quita "(web y móvil)" del texto; queda en la observación. |
| RNF 11 | RNF-11 | **Modificada** | **Should → Must** | Vista aérea, capas vectoriales, "proveedor por definir" (contradice OSM) y degradación a lista sin conexión. |
| RNF 12 | RNF-12 | Sin cambio | Must → Must | — |
| RNF 13 | RNF-13 | Sin cambio | Must → Must | — |
| RNF 14 | RNF-14 | Sin cambio (en riesgo) | Must → Must | Observación: "no figura en el Catálogo v0.4: incorporar o retirar". |
| (HUID 42) | — | **Eliminada** | Should → — | Historia "registros de errores y eventos" (observabilidad). No tenía RNF en v1 y en v2 no aparece. |

### 3.1 Historias/requisitos **nuevos** en v2
1. **RF-34** (Should): sectores de capataz como polígonos con vigencia, reasignables, y sugerencia automática de sector, capataz y referente.
2. **RF-35** (Could, fuera del MVP): concentración de intervenciones por sector/jardín.
3. **RF-32** como requisito de *visibilidad + filtros por rol*. El contenido existía repartido entre v1 RF 29 y RF 32; lo nuevo son los filtros sector/ejecutor/origen/riesgo/fechas.
4. Historias nuevas para RNF-05 (Administrador del sistema) y RNF-14 (Supervisor), y para RF-27 (Ingeniero de campo del proveedor).

### 3.2 Historias/requisitos **eliminados o absorbidos**
1. **v1 RF 34** (cancelar/archivar) → absorbido en RF-31.
2. **v1 RF 35** (reasignar capataz) → absorbido en la observación de RF-31.
3. **v1 HUID 42** (registros de errores y eventos) → desaparece.
4. Elementos retirados del texto y dejados como pendientes: "motivos de cancelación" (RF-24), "avance/checklist por jardín" (RF-17), "campos específicos por tipo" (RF-08), "evidencia por evento" (RF-19 → RF-31), "georreferenciación progresiva" (RF-04), "AWS" (RNF-03, sigue en la observación) y "web y móvil" (RNF-10, sigue en la observación).

### 3.3 Cambios de prioridad
| Requisito | v1 | v2 | Comentario |
|---|---|---|---|
| Capas de zonas/sectores (v1 RF 33 → v2 RF-33) | Must | **Should** | Depende de shapes que el cliente aún no entrega. |
| Mapa/georreferenciación (RNF 11 → RNF-11) | Should | **Must** | Coherente con que el mapa sea la vista principal. |
| Reasignación (v1 RF 35 → dentro de RF-31) | Should | **Must** (implícito) | Hay que confirmar si la reasignación hereda el Must de RF-31. |
| RF-34 nuevo | — | Should | — |
| RF-35 nuevo | — | Could | — |

---

## 4. Roles

### 4.1 Inventario de roles por fuente

| Fuente | Roles |
|---|---|
| **(a) Backlog v1**, columna "Como…" | Usuario del sistema; Administrador del sistema; Supervisor; Capataz; Usuario autorizado; Administrador técnico; Equipo de desarrollo; (RF 27 sin rol). Texto de RF 02: Capataz, Ingeniería/Coordinación, Jefatura; operarios sin cuenta. |
| **(b) Backlog v2**, columna "Como…" y textos | Los mismos de v1, más **"Ingeniero de campo de empresa proveedora autorizada"** (RF-27) y **"Supervisor / Capataz"** (RF-32). RF-02: Capataz (campo), Ingeniería/Coordinación y Jefatura (oficina), operarios sin cuenta individual. RF-01 (observación): cuentas **administradas por el jefe de sección**. RF-32: "la jefatura y el personal de oficina ven todas las actividades". |
| **(c) Código actual** | `capataz`, `coordinacion`, `jefatura`, `admin`. Aparecen en: CHECK `usuarios_rol_chk` (`apps/api/migrations/005_producto.sql`); tabla `roles` (`migrations/008_fk_minimas.sql`, nombres "Capataz", "Coordinación", "Jefatura", "Administración"); `accesos.Matriz` y `semillas` (`apps/api/internal/accesos/accesos.go:31-46`); constantes `Rol*` (`apps/api/internal/operacion/model.go`); tipo `Rol` (`apps/web/src/types.ts:1`). **Operario** no es rol: existe como `personal_labor.rol_campo = 'operario de cuadrilla'` con nombre ficticio (`migrations/020_labores.sql`). |
| Arquitectura objetivo (`init/backend`: `db/schema_nucleo_v0.2.sql`) | Tabla `rol`: Administrador, Jefe de Sección, Ingeniero/Coordinador, Capataz. RBAC `permiso`/`rol_permiso`; `personal` separado de `usuario` (operarios sin cuenta). |

> Aviso: el esquema objetivo (`schema_nucleo_v0.2.sql`, comentario del admin inicial) y su README incluyen datos personales reales (nombre y correo). Choca con la regla "solo nombres ficticios". Conviene sustituirlos antes de mezclar ese esquema con este repo.

### 4.2 Tabla "rol antiguo → rol nuevo"

| Rol v1 | Rol v2 | Rol en el código | Arquitectura objetivo | Observación |
|---|---|---|---|---|
| Capataz | Capataz (campo) | `capataz` (con `capataz_id` → `capataces`) | Capataz | Coincide. El capataz está ligado a un **equipo** (`cap-norte`, `cap-sur`, `cap-riego`), no a un sector. |
| Supervisor *(persona de las historias)* | Supervisor *(persona)* = "jefatura y personal de oficina" (RF-32) | `jefatura` + `coordinacion` | Jefe de Sección + Ingeniero/Coordinador | "Supervisor" **no es un rol de sistema**: es la persona de oficina. |
| Ingeniería/Coordinación (texto RF 02) | Ingeniería/Coordinación (oficina) | `coordinacion` (etiqueta "Coordinación") | Ingeniero/Coordinador | Coincide en el código; hay que cambiar la etiqueta visible a "Ingeniería / Coordinación". |
| Jefatura (texto RF 02) | Jefatura (oficina) / "jefe de sección" (RF-01) | `jefatura` | Jefe de Sección | v2 le asigna **administrar credenciales**; en el código no puede. |
| Administrador del sistema | Administrador del sistema | `admin` (etiqueta "Administración") | Administrador | En el código es el único que ve cuentas y edita catálogos. |
| Administrador técnico | Administrador técnico | — (no hay cuenta) | (admin técnico "puede no ser personal") | Es un rol de operación de infraestructura (secretos, despliegue), no de la app. |
| Usuario autorizado | Usuario autorizado | Cualquier rol con `registrar` | — | Genérico (RF-19). |
| Usuario del sistema | Usuario del sistema | Cualquier sesión válida | — | Genérico. |
| (texto RF 27) | Ingeniero de campo de proveedor | — | — | No existe. Es Could y está fuera del MVP. |
| Operario (sin cuenta) | Operario (sin cuenta individual) | `personal_labor` (sin cuenta) | `personal` + `intervencion_personal` | Coincide. No hay endpoint/UI para cargar `personal_labor` (ver RF-08). |
| Equipo de desarrollo | Equipo de desarrollo | — | — | No es rol del sistema. |

### 4.3 Matriz de acciones: código actual vs. v2

**Código actual.** Fuente: `accesos.Matriz` (`accesos.go:41`) aplicada por `exige()` (`handlers/producto.go:39`) y reglas de `operacion/validate.go`. Leyenda: ✔ = permitido; ✔¹ = solo en sus labores asignadas; ✖ = denegado.

| Acción (endpoint) | capataz | coordinacion | jefatura | admin |
|---|---|---|---|---|
| Registrar labor desde el mapa (`POST /operacion/actividades`, exige `validar` + `puedeAsignar`) | ✖ | ✔ | ✔ | ✔ |
| Asignar/reasignar (`PATCH …/asignacion`, `validar`) | ✖ | ✔ | ✔ | ✔ |
| Avanzar estado, incluido **cerrar** y **cancelar** (`PATCH …/estado`, `registrar` o `validar`) | ✔¹ | ✔ | ✔ | ✔ |
| Registrar avance (`POST …/avances`, `registrar`) | ✔ **(cualquier labor)** | ✔ | ✖ | ✔ |
| Editar ficha de labor (`PATCH …/ficha`, `registrar`) | ✔ **(cualquier labor)** | ✔ | ✖ | ✔ |
| Subir evidencia (`POST /evidencias`, `registrar`) | ✔¹ | ✔ | ✖ | ✔ |
| Registrar riego (`POST /riego`, `registrar`) | ✔ (se fuerza su equipo) | ✔ | ✖ | ✔ |
| Validar/cerrar como paso de supervisión | No existe como acción separada: "cerrar" es un estado más (requiere ejecución y, si es tercerizada, orden: `PuedeCerrar`, `validate.go:154`) | | | |
| Baja lógica / archivar con motivo (`POST …/archivar`, `validar`) | ✖ (pero puede poner `cancelada` vía estado) | ✔ | ✔ | ✔ |
| Solicitudes (listar/crear/editar, `solicitudes`) | ✖ | ✔ | ✔ | ✔ |
| Órdenes de servicio (crear/editar, `registrar`) | ✔ | ✔ | ✖ | ✔ |
| Reportes (`GET /reportes/labores`, `reportes`) | ✖ | ✔ | ✔ | ✔ |
| Catálogo genérico (`POST /catalogos`, desactivar; `catalogos`) | ✖ | ✖ | ✖ | ✔ |
| Catálogos de catastro: zonas, cuadrillas, lugares, especies, ejemplares, recodificar (`registrar`) | ✔ | ✔ | ✖ | ✔ |
| Inventario: tachos, bebederos, puntos, capas; alta, edición y baja lógica (`registrar`) | ✔ | ✔ | ✖ | ✔ |
| Importar, confirmar lote, revertir, ediciones auditadas (`validar`) | ✖ | ✔ | ✔ | ✔ |
| Usuarios: ver cuentas y permisos (`GET /accesos/usuarios`, solo `admin`) | ✖ | ✖ | ✖ | ✔ (solo lectura) |
| Usuarios: crear, desactivar, cambiar clave o rol | No existe para ningún rol | | | |
| Leer labores, timeline, capataces, `/geo/*` | **Sin autenticación** (`handlers/operacion.go:67,81,275`; `handlers/geo.go`) | | | |

Módulos visibles en la PWA (`apps/web/src/App.tsx:55-59`):

| Rol | Módulos |
|---|---|
| capataz | mapa, labores, catastro, inventario |
| jefatura | lo anterior + solicitudes, reportes, importaciones (sin catálogos) |
| coordinacion | lo anterior + catálogos (solo lectura, porque `editable={rol === "admin"}`) |
| admin | todo + "Admin" |

**Lo que exige v2.** Deducido de RF-02, RF-03, RF-01 (observación), RF-06, RF-24, RF-30, RF-31 y RF-32. La matriz formal sigue "pendiente de validación" en RF-03.

| Acción | Capataz | Ingeniería/Coordinación | Jefatura | Admin. del sistema | Diferencia con el código |
|---|---|---|---|---|---|
| Registrar labor (desde el mapa) | ? (RF-12 le da "registrar en campo"; no está claro si crea labores nuevas) | ✔ | ✔ | — | Pregunta abierta P3 |
| Asignar/reasignar | ✖ | ✔ | ✔ | — | Coincide |
| Actualizar avance/estado de lo suyo | ✔ solo lo suyo | ✔ | ✔ | — | El código deja al capataz tocar ficha y avances **ajenos** |
| Validar/cerrar ("supervisión", "conformidad y cierre", RF-31) | ✖ (propuesta) | ✔ | ✔ | — | En el código, el capataz puede **cerrar** y **cancelar** |
| Cancelar/baja lógica con motivo y confirmación | ✖ | ✔ | ✔ | — | Hay que bloquear `cancelada` para el capataz |
| Reportes | ✖ | ✔ | ✔ | — | Coincide |
| Catálogos (RF-24) y zonificación (RF-06) | ✖ | ✔ (zonificación; "Supervisor") | ✔ | ✔ | Hoy solo `admin` edita el catálogo genérico, y el **capataz** puede crear especies, lugares y cuadrillas |
| Importar (RF-07, "Supervisor") | ✖ | ✔ | ✔ | ✔ | Coincide (admin incluido) |
| Usuarios y credenciales (RF-01 obs.: "jefe de sección") | ✖ | ✖ | ✔ | ✔ | No existe la gestión, y `jefatura` no tiene el permiso |
| Evidencias ("usuario autorizado") | ✔ lo suyo | ✔ | ✔ | — | `jefatura` no puede subir |

### 4.4 Mapeo definitivo propuesto y cambios necesarios

**Propuesta:** conservar los **códigos** actuales y cambiar solo las **etiquetas**. Así no hay migración de datos: `usuarios.rol`, `permisos.rol` y el histórico `actividad_eventos.actor_rol` no se tocan.

| Código (estable) | Etiqueta nueva (UI y `roles.nombre`) | Persona v2 |
|---|---|---|
| `capataz` | Capataz | Capataz (campo) |
| `coordinacion` | Ingeniería / Coordinación | Supervisor (oficina) |
| `jefatura` | Jefatura de sección | Supervisor (oficina) + administra credenciales |
| `admin` | Administrador del sistema | Administrador del sistema (usuarios, roles, catálogos) |
| *(reservado)* `proveedor` | Ingeniero de campo de proveedor | RF-27 (Could); **no crear cuentas en el MVP** |
| — | — | Administrador técnico: fuera de la app (secretos, AWS) |

Cambios recomendados (aditivos, sin borrar datos):

1. **Etiquetas.** Hacer un `UPDATE roles SET nombre = …` en una migración nueva y usar `roles.nombre` en la UI en vez de mostrar el código (hoy `AdminPanel` muestra `user.rol` en crudo, `apps/web/src/panel/Modulos.tsx:688`).
2. **Roles configurables (RF-02).** Sustituir el CHECK `usuarios_rol_chk` por una FK `usuarios.rol → roles.codigo` y añadir `roles.activo`. Hoy la FK ya existe para `permisos.rol` (`008_fk_minimas.sql`), pero `usuarios` sigue con el CHECK cerrado.
3. **Permisos configurables (RF-03).** `accesos.Ensure` hace `DELETE FROM permisos` y reescribe la matriz en cada arranque (`accesos.go:99`). Cualquier ajuste en BD se pierde, y además borra filas, lo que choca con la regla de no borrar datos. Propuesta:
   - que `Matriz` solo **siembre si falta** (`INSERT … ON CONFLICT DO NOTHING`);
   - que `Permite` lea de `permisos`, con caché;
   - y registrar los cambios en `cambios`.
4. **Si hiciera falta renombrar códigos** (p. ej. `coordinacion` → `ingenieria`), hacerlo en una migración **aditiva**:
   - insertar el código nuevo en `roles`;
   - ampliar temporalmente el CHECK o FK para aceptar ambos;
   - hacer `UPDATE usuarios` y `UPDATE permisos` en una sola transacción;
   - dejar el código viejo en `roles` como **alias** (`activo = false`, columna `alias_de`), con una función `normalizarRol()` en `accesos` que traduzca el alias en `Login`/`FromToken`;
   - **no reescribir** `actividad_eventos.actor_rol` ni `cambios`, y resolver la etiqueta del histórico al leer;
   - añadir el código nuevo en `types.ts` y en `operacion/model.go` con el alias durante una versión.
5. **Ajustar la matriz a v2:**
   - `jefatura` gana `usuarios` y `registrar` (solo evidencias/observaciones) o una acción `evidencias`;
   - el capataz pierde `cerrada`/`cancelada` en `SetEstado`;
   - añadir la comprobación "solo sus labores" en `Ficha` y `CrearAvance`;
   - separar la acción `catalogos_catastro` de `registrar` para que el capataz no cree especies, lugares ni cuadrillas;
   - exigir sesión en `GET /operacion/*` y `/geo/*`.
6. **Gestión de cuentas (RF-01):** `POST/PATCH /api/v1/accesos/usuarios` (alta, desactivación, cambio de rol y restablecimiento de clave) con permiso `usuarios`, además de `debe_cambiar_password`, como en el esquema objetivo. Hoy las 6 semillas comparten la clave `CAMPUS_DEV_PASSWORD`.
7. **Unificar "capataz".** Hoy conviven `capataces` (equipos `cap-*`, a los que apuntan `usuarios.capataz_id` y `actividades.assigned_capataz_id`) y `cuadrillas` (`cua-*`, a las que apuntan `poligonos_cuadrilla.sector`/`cuadrilla_id` y `actividades.cuadrilla_id`). RF-34 ("sector de capataz") necesita un solo vínculo capataz ↔ sector. Propuesta:
   - añadir `cuadrillas.capataz_usuario_id` o `sector_capataz(capataz, poligono, vigente_desde, vigente_hasta)`;
   - no borrar ninguna de las dos tablas.

---

## 5. Cobertura de cada RF/RNF de v2 en la app actual

| ID v2 | Prioridad | ¿Cubierto? | Dónde | Qué falta |
|---|---|---|---|---|
| RF-01 | Must | Parcial | Cuentas locales con bcrypt; sesión por cookie HttpOnly de 12 h; límite de intentos de login. `accesos/accesos.go` (`Login`, `FromToken`), `POST/GET/DELETE /api/v1/sesion`, pantalla `Login` (`panel/Modulos.tsx:30`) | Gestión de cuentas por el jefe de sección (alta/baja/rol/clave); cambio de clave; claves individuales (hoy todas las semillas usan una sola); auth en endpoints de lectura abiertos |
| RF-02 | Must | Parcial | 4 roles: CHECK en `005_producto.sql`; tabla `roles` (`008`); operario sin cuenta (`personal_labor`) | Roles **configurables** (CHECK cerrado, sin UI); etiquetas v2 |
| RF-03 | Must | Parcial | `accesos.Matriz` + `exige()` en todos los handlers de escritura; `AdminPanel` muestra la matriz | Matriz editable (hoy se reescribe al arrancar); corregir las inconsistencias de §4.3; validar la matriz con el cliente |
| RF-04 | Must | Parcial | Tabla `ejemplares` (código, especie, salud, lat/lon, área, lugar) en `015_especies_ejemplares.sql`; `GET/POST /catastro/ejemplares`; capas "Flora"/"Cafetos" en el mapa (`apps/web/src/inventario.ts`); 521 áreas verdes | UI para consultar/editar ejemplares (hoy solo alta por API o importación); `PATCH` de ejemplar; sector/cuartel del ejemplar (no hay columna; solo `area_verde_id`/`ubicacion_lugar_id`) |
| RF-05 | Should | Parcial | `GET/POST /catastro/ejemplares/:id/codigos` + `codigos_historicos` (`015`); prueba `catastro/recodificar_test.go` | Pantalla de recodificación |
| RF-06 | Must | Parcial | `zonas_supervision` (Z1–Z4), `poligonos_cuadrilla.sector`, `lugares`, `cuadrillas` (`/catastro/*`); edificios (`/geo/edificios`) | Sectores de capataz como catálogo editable (hoy CHECK fijo en `044_sector_poligonos.sql`); referentes "vías"; cuartel histórico (sin tabla); la ficha de labor acepta **`lugar_libre` en texto libre** (`operacion/store.go:201`), contra "evitar escritura libre" |
| RF-07 | Should | Sí | Importación con vista previa, confirmación y reversión para 22 entidades, en CSV/XLSX (`etl/importar.go:15`, `etl/formato.go`, `/api/v1/importaciones`, `/api/v1/lotes/:id/revertir`, `panel/Importaciones.tsx`); exportación CSV de capas (`/inventario/export/:capa`) | Validar con la estructura real del Excel del cliente (pendiente en v2) |
| RF-08 | Must | Parcial | `tipo` (5 valores del catálogo `tipo_actividad`); `clase_actividad` (7 clases, `020_labores.sql`); ficha con clase/fechas/lugar/comentario; tabla `personal_labor`; módulos específicos Poda y Vivero (`021`, `022`) | Taxonomía de 2 niveles alineada con v2 (faltan "manejo fitosanitario" e "inspección y monitoreo" como clase; la clase no se elige al crear); **insumos**; **carga de personal** (sin endpoint para `personal_labor`); campos específicos por tipo (JSONB, decisión pendiente) |
| RF-09 | Must | Sí | `actividades.ejecutor` CHECK `propia`/`tercerizada` (`005`); selector en el formulario (`panel/Labores.tsx:174`) | — |
| RF-10 | Should | **No** | — (no existe la entidad "insumo" en `migrations/` ni en `internal/`) | Todo: catálogo de insumos, consumo por labor, espacio y periodo |
| RF-11 | Must | Parcial | Tabla `solicitudes` (código externo, fuente, prioridad, lugar, cantidad, estado, actividad) en `005`; `GET/POST /solicitudes`, `PATCH /solicitudes/:id`; `SolicitudesPanel` | Cantidad solicitada **y** ejecutada (hay un solo campo); ubicación georreferenciada (hoy texto); `fuente` y `estado` con CHECK fijo (no parametrizables) |
| RF-12 | Must | Parcial | Cola en IndexedDB para altas y cambios de estado (`offline/queue.ts`); cola de evidencias con reintento y drenado al evento `online` (`panel/EvidenciasCampo.tsx:111`); última lista guardada (`App.tsx:157`); PWA con service worker (`vite.config.ts`, `VitePWA`) | La cola de labores/estados solo se vacía al montar la app o con el botón (no escucha `online` ni usa Background Sync); los avances, el riego y la ficha no tienen cola; el capataz no puede crear labores, así que la cola de altas no le sirve |
| RF-26 | Must | Parcial | `riego_registros` (sector, turno, capataz, fecha, zona, ciclo, superficie); `GET/POST /riego`; `RiegoPanel` | Avance y cobertura del ciclo (en la UI figura "definición pendiente", `producto.ts` `HUECOS`) |
| RF-13 | Must | Parcial | `ordenes_servicio` (empresa, referencia, frecuencia, estado, conformidad, periodo) en `005` y `024`; `POST /ordenes`, `PATCH /ordenes/:id`; regla de cierre de tercerizadas (`PuedeCerrar`) | Catálogo de empresas/frecuencias (texto libre); "reporte" como entidad; derivación como evento de la labor |
| RF-14 | Should | Parcial | `evidencias.orden_id`; el formulario acepta `orden_id` (`handlers/evidencias.go`, `EvidenciasCampo.tsx:58`) | UI específica para adjuntar el reporte del proveedor a la orden |
| RF-15 | Should | **No** | Solo un aviso "Métricas de proveedor: definición pendiente" (`producto.ts:90`) | Definir y calcular las métricas |
| RF-27 | Could | **No** | — | Rol `proveedor` y permisos restringidos (fuera del MVP) |
| RF-16 | Must | Parcial | Estados de labor y catálogo `estado`; estados de solicitud (`solicitudes_estado_chk`); regla de tercerizada (`PuedeCerrar`) | Nombres v2 (Por iniciar/Ejecutado/Cerrado/Archivado); **validar el estado contra el catálogo** (`SetEstado` acepta cualquier *slug*; el CHECK se eliminó en `005`); transiciones permitidas; web con `ESTADOS`/`ABIERTOS` en duro (`operacion.ts:21-30`); un solo modelo de estados para labor y solicitud |
| RF-17 | Should | Parcial | El mapa y la lista se recargan tras cada acción; `actividad_avances` + `POST …/avances` | Actualización periódica o en tiempo real; checklist por jardín/área (pendiente en v2) |
| RF-18 | Must | Parcial | Timeline por labor (`GET …/timeline`); historial de cambios (`GET /auditoria/cambios`, `/auditoria/timeline`); reporte con filtros de estado, fechas, zona, cuadrilla y origen | Consulta por **ejemplar** y por **responsable** en la UI; las labores cerradas no se ven en el mapa (`abiertas=1` por defecto) y no hay pantalla de histórico |
| RF-28 | Must | Sí | `solicitudes.codigo_externo` único y nunca generado (`005`, comentario de columna); `podas.codigo_externo` | Llevar el código externo al alta de labor (RF-30) |
| RF-19 | Must | Parcial | `POST /evidencias` (multipart, SHA-256 idempotente, lat/lon, EXIF, PDF/JPG/PNG/WebP, S3 o disco): `evidencias/subir.go`, `040_evidencia_meta.sql`; genera un evento `evidencia` en el timeline | Asociar a **solicitud** (existe la columna `solicitud_id`, pero la API no la recibe); asociar a un **evento concreto** (no hay `evento_id`) |
| RF-20 | Must | Parcial | Reporte "básico": conteos por estado + filas (`GET /reportes/labores`, `ReportesPanel`) | Niveles intermedio y avanzado (contenido por validar) |
| RF-21 | Should | **No** | Aviso "Rendimiento: definición pendiente" | Distribución de personal, cobertura, rendimiento y trazabilidad de atenciones |
| RF-22 | Should | Parcial | CSV y Excel (SpreadsheetML), en `atencion/export.go` (`EscribirExcelXML`) | **PDF** |
| RF-23 | Should | **No** | — | Indicadores configurables |
| RF-24 | Must | Parcial | `catalogos` con 8 clases administrables (`catalogos/store.go:13`): alta y desactivación (baja lógica); `CatalogosPanel` | Clases faltantes: plagas/productos, frecuencias, roles, sedes; `clase_actividad` existe en la BD pero **no está** en `Clases`; edición y orden; varios catálogos siguen como CHECK (ver RNF-05) |
| RF-25 | Should | Parcial | `POST /ia/sugerir-tipo`: reglas locales por palabras clave (`atencion/ia.go`), con confirmación humana en la UI | Un caso de IA con "valor verificable" (el actual es heurístico, no un modelo); definir el caso con el cliente |
| RF-29 | Must | Sí (mayormente) | MapLibre + OSM (`map/CampusMap.tsx`); marcadores de labores abiertas con color por estado y letra por tipo (`actividades-circle`/`actividades-marca`); popup de resumen y detalle en el panel lateral sin salir del mapa | "Vista aérea" (hay plano y relieve, no ortofoto/satélite); icono por **eje** (hoy es una letra por tipo) |
| RF-30 | Must | Parcial | Alta por pin con tipo, título, detalle, capataz y ejecutor; UUID idempotente; alta consecutiva ("Labor creada. Marque el siguiente punto.", `App.tsx:461`) | Eje + subtipo, origen, código externo, unidad solicitante, **nivel de riesgo**, **fecha programada**, cantidad pedida |
| RF-31 | Must | Parcial | `actividad_eventos` con usuario, fecha y nota; baja lógica con motivo del catálogo y **confirmación** (`App.tsx:501-512`); reasignación con evento `reasignada` | Tipos de evento v2 (inicio, actualización, supervisión, derivación, observación/subsanación, conformidad); guardar el capataz **anterior** en la reasignación (hoy solo el nuevo, `operacion/store.go:262-278`); evidencias por evento; los avances no aparecen en el timeline |
| RF-32 | Must | Parcial | Visibilidad del capataz limitada en el servidor (`handlers/operacion.go:91`, `operacion/store.go:67`); la oficina ve todo; filtros de estado, tipo y equipo (`panel/Labores.tsx`) | Filtros de sector, ejecutor, origen, riesgo y fechas en la UI; alternar mapa/lista explícitamente para el capataz; degradar a lista sin conexión (hoy se muestra la lista en caché, pero no hay un modo "solo lista") |
| RF-33 | Should | Parcial | Capas activables: áreas (por uso o sector), zonas, jardines de reserva, xerofítica, edificios, inventario (`types.ts:LAYERS`, `inventario.ts`) | Cuarteles históricos, inventario forestal completo y jardines de préstamo (no se pudo confirmar si equivalen a "jardines de reserva": pregunta P9); carga desde shapes del cliente |
| RF-34 | Should | **No** (solo base de datos) | `poligonos_cuadrilla.sector`; `asignaciones_poligono.vigente` (`012`, `044`) | Vigencia con fechas, reasignación de sectores en la UI, **sugerencia automática** de sector, capataz y referente al marcar el pin (las uniones espaciales solo están en el ETL: `etl/lote.go:238,279`) |
| RF-35 | Could | **No** | — | Vista de concentración (fuera del MVP) |
| RNF-01 | Must | Parcial | Colas en IndexedDB, contador de pendientes (`queueCount` en `Labores`), estados de subida de evidencias (`offline/subida.ts`) | Sincronización automática al recuperar conexión para labores/estados; cobertura offline de avances, riego y ficha |
| RNF-02 | Must | Sí | `vite-plugin-pwa` (manifest, iconos, SW); diseño móvil con *bottom sheet* (`ui/BottomSheet.tsx`); capturas `fase1-mobile-*.png` y `fase1-desktop-*.png` | — |
| RNF-03 | Must | Sí | EC2 + docker compose en AWS Learner Lab (`docs/DEPLOY-AWS.md`); evidencias en S3 (`internal/blobs/s3.go`) | Validar las restricciones institucionales de la PUCP |
| RNF-04 | Must | Parcial | bcrypt, cookie HttpOnly (`accesos/cookie.go`), CORS con lista blanca, límite de login (`handlers/seguridad.go`), secretos por entorno (`internal/config/config.go`), auditoría `cambios` (`007`) y `actividad_eventos.usuario_id` | Lecturas sin autenticación (§4.3); el despliegue sirve **HTTP en el puerto 80** sin TLS (`DEPLOY-AWS.md`); clave compartida por las semillas; cifrado en reposo no verificado |
| RNF-05 | Must | Parcial | Tabla `catalogos` y catálogo `tipo_actividad` validado en el alta (`operacion/store.go:148`) | Valores en duro: `TIPOS`/`ESTADOS` en la web; CHECK de `zonas_supervision.codigo` (Z1–Z4), `poligonos_cuadrilla.sector`, `solicitudes.fuente/estado`, `ordenes.estado`, `usuarios.rol`, `cuadrillas.turno`, `podas.prioridad` |
| RNF-06 | Must | Parcial | Interfaz móvil con controles grandes y mensajes en español (auditorías previas: `docs/UI-AUDIT.md`, `docs/AUDITORIA-UI-SCROLL.md`) | Validación con capataces reales; no se verificó la accesibilidad (no hay pruebas a11y) |
| RNF-07 | Should | Parcial | `LIMIT` fijos (100/200) en solicitudes, órdenes, riego y evidencias (`atencion/store.go`); `Cache-Control` en evidencias; `limit` en `/geo` | Paginación real (parámetros); `GET /operacion/actividades` sin límite |
| RNF-08 | Should | Sí | Capas `handlers/` → `internal/<dominio>/store.go` → `migrations/`; OpenAPI (`apps/api/openapi.yaml`); web por módulos | — |
| RNF-09 | Should | Sí | REST/JSON + OpenAPI; importación CSV/XLSX; exportación CSV/XLS | — |
| RNF-10 | Must | Parcial | 41 archivos `_test.go`; 17 pruebas web (`apps/web/package.json`, script `test`); CI en `.github/workflows/ci.yml` | Pruebas end-to-end y en móvil (la observación de v2 pide "web y móvil") |
| RNF-11 | Must | Parcial | MapLibre + OSM y capas vectoriales de puntos y polígonos | Vista aérea/ortofoto; degradar el mapa a lista sin conexión (no hay caché de teselas ni modo lista forzado) |
| RNF-12 | Must | Sí | UI y mensajes de la API en español | — |
| RNF-13 | Must | Sí (para el caso actual) | La sugerencia es local, sin llamada externa, con `requiere_humano` y confirmación del usuario (`atencion/ia.go`) | Revisar de nuevo cuando se elija el caso definitivo de RF-25 |
| RNF-14 | Must | Parcial | Baja lógica en labores, podas, vivero, inventario y catálogos; eventos y `cambios` inmutables | Quedan borrados físicos: `accesos.Ensure` borra `permisos` al arrancar; la reversión de un lote hace `DELETE FROM medidas_palmera` (`auditoria/store.go:411`); `cmd/etl` hace `TRUNCATE areas_verdes, poligonos_cuadrilla, capas_auxiliares` y `TRUNCATE inventario` (`etl/load.go:47`, `etl/inventario.go:137`), aunque existe la protección `negarSiHayDependientes` y la vía alternativa con upsert `etl-lote`. Falta la política de retención |

**Resumen:** de 35 RF, 4 están en **Sí** (RF-07, RF-09, RF-28, RF-29), 24 en **Parcial** y 7 en **No** (RF-10, RF-15, RF-21, RF-23, RF-27, RF-34, RF-35). De 14 RNF, 6 están en **Sí** (RNF-02, 03, 08, 09, 12, 13) y 8 en **Parcial**.

---

## 6. Brechas priorizadas y preguntas abiertas

### 6.1 Must

| # | Brecha | Requisitos | Acción propuesta |
|---|---|---|---|
| M1 | Lecturas de labores, timeline, capataces y `/geo/*` sin autenticación; despliegue sin TLS | RNF-04, RF-01 | Exigir sesión (`exige(c, "consultar")`) en `handlers/operacion.go` y `handlers/geo.go`; terminar TLS en nginx (`nginx.tls.conf` ya existe) |
| M2 | Permisos del capataz demasiado amplios (cerrar/cancelar; ficha y avances ajenos; crear especies, lugares y cuadrillas) y jefatura sin evidencias | RF-03, RF-31, RF-32 | Ajustar `SetEstado`, `GuardarFicha`, `CrearAvance` y separar la acción `catalogos_catastro` (§4.4, punto 5) |
| M3 | Sin gestión de cuentas por el jefe de sección; roles y permisos no configurables; `DELETE FROM permisos` al arrancar | RF-01, RF-02, RF-03, RNF-14 | CRUD de usuarios con baja lógica; FK `usuarios.rol → roles`; sembrar sin borrar (§4.4, puntos 2, 3 y 6) |
| M4 | Estados no parametrizados, `SetEstado` sin validación contra el catálogo, nombres distintos de v2 | RF-16, RNF-05 | Validar contra `catalogos(clase='estado')`; añadir `ejecutado` y `archivado` y etiquetas v2 **sin renombrar códigos** (mapear `pendiente` → "Por iniciar"); tabla de transiciones; la web debe leer los estados del catálogo |
| M5 | Alta desde el mapa incompleta | RF-30, RF-28, RF-11 | Añadir columnas nullable a `actividades` (`subtipo`, `origen`/`codigo_externo`, `unidad_solicitante`, `nivel_riesgo`, `fecha_programada`, `cantidad_pedida`) y a `createBody`/`CreateInput` y al formulario |
| M6 | Cadena de eventos corta; reasignación sin capataz anterior; avances fuera del timeline; evidencia sin evento | RF-31, RF-19, RNF-14 | Ampliar el CHECK de `actividad_eventos.tipo`; añadir `capataz_anterior_id`; crear un evento por cada avance; añadir `evidencias.evento_id` |
| M7 | Filtros de oficina y modo lista del capataz | RF-32, RNF-11 | Filtros de sector, ejecutor, origen, riesgo y fechas en `Labores.tsx` (la API ya admite parte); conmutador mapa/lista; lista forzada sin conexión |
| M8 | Offline incompleto | RF-12, RNF-01 | Escuchar `online` o usar Background Sync para vaciar `listQueue`/`listEstados`; encolar avances, riego y evidencias de forma homogénea |
| M9 | Taxonomía de 2 niveles, personal e insumos | RF-08 (Must), RF-10 (Should) | Relacionar `clase_actividad` con `tipo_actividad` (añadir fitosanitario e inspección); endpoint para `personal_labor`; entidad de insumos |
| M10 | Catálogos incompletos y valores en duro | RF-24, RNF-05 | Añadir las clases que faltan a `catalogos.Clases` (incluida `clase_actividad`); migrar los CHECK fijos a catálogo de forma aditiva (quitar el CHECK solo después de sembrar el catálogo) |
| M11 | Zonificación v2 (sectores de capataz, referentes, texto libre en la ficha) | RF-06 | Catálogo de sectores editable; `lugar_libre` solo como dato heredado (no editable); unificar `capataces`/`cuadrillas` (§4.4, punto 7) |
| M12 | Reportes solo en nivel básico, sin PDF | RF-20 (Must), RF-22 | Definir los niveles con el cliente; exportación PDF |
| M13 | Riego sin cobertura ni avance del ciclo | RF-26 | Definir el indicador con el cliente y calcularlo por sector y ciclo |
| M14 | Inventario de ejemplares sin UI; historial por ejemplar/responsable | RF-04, RF-18 | Pantalla de ejemplares (consulta, edición con historial, recodificación); vista de histórico con labores cerradas |
| M15 | Pruebas e2e y en móvil | RNF-10 | Playwright en viewport móvil sobre el flujo del capataz |

### 6.2 Should / Could

- **Should:** RF-34 (sectores con vigencia y sugerencia por `ST_Covers` al marcar el pin); RF-33 (capas del cliente, cuarteles históricos); RF-05 (UI de recodificación); RF-13/14/15 (catálogo de empresas, reporte del proveedor, métricas); RF-17 (checklist por jardín, refresco periódico); RF-21/RF-23 (reportes de proceso e indicadores); RF-25 (caso de IA con valor verificable); RNF-07 (paginación).
- **Could:** RF-27 (rol `proveedor`) y RF-35 (concentración).
- **Observabilidad** (antigua HUID 42): hoy solo existe el middleware `Bitacora()` (`internal/server/server.go`). Hay que decidir si vuelve al backlog como RNF.

### 6.3 Preguntas abiertas para el cliente

1. **Matriz de permisos (RF-03):** ¿se aprueba la matriz propuesta en §4.3? En particular: ¿quién **cierra** (capataz o solo la oficina tras la conformidad) y quién **cancela**?
2. **Administración de cuentas (RF-01):** ¿el jefe de sección administra las cuentas directamente (rol `jefatura` con permiso `usuarios`) o existe un "Administrador del sistema" distinto? ¿El "Administrador técnico" necesita cuenta en la app?
3. **¿El capataz crea labores nuevas** en campo (hallazgos) o solo ejecuta las asignadas? RF-12 dice "registrar información en campo" y RF-30 atribuye el alta al Supervisor.
4. **Estados (RF-16):** ¿nombres y transiciones definitivos? ¿"Bloqueada" (existe en la app) se mantiene? ¿"Cancelado" y "Archivado" son estados distintos o uno solo?
5. **Nivel de riesgo (RF-30):** ¿bajo/alto o bajo/medio/alto? ¿Qué criterio se usa?
6. **Taxonomía (RF-08/RF-30):** ¿lista definitiva de ejes y subtipos? ¿Qué campos específicos necesita cada tipo?
7. **Sectores de capataz (RF-06/RF-34):** ¿cuándo llegan los polígonos? ¿La vigencia es por fechas? ¿Un sector puede tener más de un capataz o turno?
8. **Referentes:** ¿qué capa de "vías" se usará? ¿Los edificios OSM bastan como referente?
9. **Capas (RF-33):** ¿los "jardines de préstamo" son los que la app llama "jardines de reserva"? ¿Cuándo llegan los shapes de tipos de uso, inventario forestal y cuarteles?
10. **Reportes (RF-20/21/22):** ¿columnas de los niveles básico, intermedio y avanzado? ¿El PDF debe tener un formato institucional?
11. **Indicadores (RF-23/RF-26):** ¿fórmula oficial de cobertura de riego y de mantenimiento?
12. **Motivos de cancelación y checklist por jardín:** v2 los deja "pendientes de incorporar al Catálogo". ¿Se confirman los 4 motivos que ya existen en la app?
13. **RNF-14:** ¿se incorpora al Catálogo v0.4? ¿Qué política de retención se aplica?
14. **RNF-11:** ¿se confirma OSM/MapLibre y se corrige el texto "proveedor por definir"? ¿Hay ortofoto del cliente para la "vista aérea"?
15. **IA (RF-25):** ¿cuál de los tres candidatos se elige? ¿Qué datos pueden salir del campus?
16. **Catálogo v0.4 de HU:** ¿puede compartirse para verificar a qué historia corresponde cada HU-nn de v2?
