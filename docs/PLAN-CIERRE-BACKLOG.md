# Plan de cierre del backlog v2

> **Nota de conservación (2026-10-03).** Este plan se recuperó tal cual de la rama `docs/plan-cierre-backlog` antes de borrarla. El texto conserva las rutas del momento en que se escribió. Hoy: `apps/web/` es `frontend/`; `apps/api` ya no existe (un solo backend en `backend/`); las migraciones 001 a 078 se consolidaron en `db/migrations/001_esquema_base.sql` y `002_catalogos_base.sql` (la serie vieja está en `db/referencia/migraciones-historicas/`), y las nuevas siguen desde la **079**. Los documentos que el plan cita como `docs/MATRIZ-BACKLOG-V2.md`, `docs/matriz-backlog-v2.csv`, `docs/DECISIONES.md`, `docs/MAPA-DATOS-Y-EDICION.md` y `docs/PROPUESTA-UI-CLARIDAD.md` están ahora en `docs/plan-cierre/`. `docs/AMBIENTES.md`, `docs/DEPLOY-RUNNER.md` y `docs/DESPLIEGUE.md` se reemplazaron por `deploy/README.md`, `infra/README.md` y `docs/CAMBIO-DE-CUENTA-LAB.md`; el contenido anterior sigue en el historial de git.

Fecha: 2026-10-02. Rama de este documento: `docs/plan-cierre-backlog`, salida de `docs/matriz-backlog-v2`.

Solo planificación. Este archivo no cambia `backend/`, `apps/web/`, `db/migrations` ni datos. El cierre se hace después, flujo por flujo, cada uno en su rama.

Lectura obligatoria de quien ejecute un flujo: este plan (su sección), `docs/GLOSARIO-NOMENCLATURA.md`, `docs/MATRIZ-BACKLOG-V2.md` y `docs/MAPA-DATOS-Y-EDICION.md`.

## 1. Corte que hay que cerrar

La matriz (`docs/matriz-backlog-v2.csv`, 201 filas) miró `develop` en `33d6ed2`. Conteos: CA 92, HU 49, RF 35, RNF 14, DEC 11. Por estado de todas las filas: Cumplido 40, Parcial 111, No cumplido 50.

De los 49 RF+RNF, 8 están cumplidos, 34 parciales y 7 no cumplidos (16 % estricto, 51 % de avance). De los 32 Must, 5 están cumplidos y **27 quedan parciales**. Ningún Must está entero en «No cumplido». Los **7 no cumplidos** de requisito son Should o Could.

Los 27 Must parciales: RF-01, RF-02, RF-03, RF-04, RF-06, RF-08, RF-11, RF-12, RF-13, RF-16, RF-18, RF-19, RF-20, RF-24, RF-26, RF-29, RF-30, RF-31, RF-32, RNF-01, RNF-03, RNF-04, RNF-05, RNF-06, RNF-10, RNF-11, RNF-14.

Los 7 sin implementación de requisito: RF-10 (Should), RF-15 (Should), RF-21 (Should), RF-23 (Should), RF-27 (Could), RF-34 (Should), RF-35 (Could).

Pendientes ya encargados, aunque no sean un RF nuevo: DEC-06 cuentas, DEC-07 campos del alta, DEC-08 claridad, DEC-09 DuckDNS/HTTPS, DEC-10 evidencias en S3, DEC-11 PATCH de reserva con una sola hora.

## 2. Reglas comunes

Quien ejecute un flujo las cumple todas. No se rediscuten en el prompt.

### Repositorio y ramas

- Repo de trabajo: **Irico17/Areas-verdes-pucp** (personal). No se empuja al repo del grupo. No se toca `main`. No hay force push.
- La rama de implementación sale de `develop` actualizado (el corte de la matriz es `33d6ed2`; si `develop` avanzó, se usa ese avance). No sale de esta rama de documentación ni de `main`.
- Nombre sugerido: `cierre/<flujo>` (lista en cada flujo). Una rama por flujo. El agente no mezcla dos flujos en la misma rama.
- Promoción: el trabajo entra a `develop`, de ahí a qa (tag `rc-*` o disparo manual) y a producción solo con la aprobación que ya describe `docs/AMBIENTES.md` y `docs/DEPLOY-RUNNER.md`. Este plan no cambia ese orden.
- CI existente (`.github/workflows/ci.yml`, `deploy.yml`) se mantiene. Un flujo de producto no reescribe el workflow salvo el flujo `cierre/o6-duckdns-https`, que solo toca despliegue y documentación de TLS.

### Arquitectura que ya corre

- El proceso es `backend/` (Go, Gin, GORM). `apps/api` no es el servidor. No se portan handlers hacia allá ni se revive ese módulo.
- Rutas de producto en `/areas-verdes/v1`. El alias `/api/v1` sigue respondiendo lo mismo. No se inventa un tercer prefijo.
- Capas: `presentation` → caso de uso → repositorio GORM. El controlador no escribe SQL.
- Migraciones solo en `db/migrations`. La última del corte es `048`. Cada flujo usa únicamente el rango de la tabla de abajo. Numeración de cuatro dígitos no: `049_…sql`, `050_…sql`. Aditivas e idempotentes (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, columnas nulas). No se renumeran, no se edita una migración ya aplicada, no se rellena el hueco de otro flujo.
- No hay `AutoMigrate`. No hay `TRUNCATE`. No hay `DELETE` de datos de negocio (catastro, actividades, ejemplares, evidencias, catálogos, reservas, medidas). Borrar una sesión al cerrar login sí puede quedar. La reversión de un lote que hoy hace `DELETE FROM medidas_palmera` la sustituye el flujo de auditoría por baja lógica.
- `etl-lote` sigue sin truncar. `Load()` con `TRUNCATE` no se llama sobre una base que ya tiene filas.
- Historial: cada alta, edición y baja lógica escribe en `cambios` (antes/después) o en `actividad_eventos`, según el dominio. Toda entidad nueva entra al importador (`panel/importaciones.ts` y el caso de uso de importación) con vista previa, confirmación y reversión que no pisa una edición posterior.
- Nombres de persona: solo ficticios (DEC-03). No se copia `jefes` ni un responsable real. El campo `jefes` de `jefe_de_grupo.json` no se persiste.
- Login: cuentas propias, cookie HttpOnly, bcrypt. **No hay SSO** de la universidad. Se borra de la interfaz y de `usuario.usecase.go` cualquier frase que lo deje pendiente. Las cuentas nuevas no comparten `CAMPUS_DEV_PASSWORD`; cada alta tiene su clave y `debe_cambiar_password`.
- Nomenclatura visible: `docs/GLOSARIO-NOMENCLATURA.md`. Códigos de rol y de estado no se renombran. Etiquetas oficiales de rol: Capataz; Ingeniería / Coordinación; Jefatura de sección; Administrador del sistema.
- Lo que el libro deja «pendiente de validar con el cliente» (sección 6 de la matriz: fórmulas, columnas de reporte, si «Bloqueada» se conserva, si jardín de préstamo es jardín de reserva, retención, nivel de riesgo medio) no se inventa como decisión cerrada. El flujo construye el mecanismo, muestra el número o la columna como **provisional** y lo dice en pantalla. No marca el requisito como fórmula oficial.

### Cómo no pisarse

Tres cortes, y un integrador por ola que es el único que toca el registro.

| Corte | Regla |
|---|---|
| Backend | Un flujo, un dominio. Archivos de otro dominio no se editan. Si el dominio ya tiene `*.group.go`, el flujo añade rutas ahí. Si necesita rutas nuevas y otro flujo de la misma ola ya edita ese group, crea `groups/<flujo>.group.go` y no registra solo: lo deja listo y el integrador añade el campo en `routes.go` y `container.go`. |
| Frontend | Un flujo, una pestaña o un archivo nuevo. `App.tsx` lo editan solo `cierre/o0-desacople` y `cierre/o4-claridad`. El resto exporta un componente y, si hace falta una pestaña, añade **una** entrada al final de `apps/web/src/ui/registroModulos.ts` (ese archivo lo crea la ola 0). |
| OpenAPI | Cada flujo añade paths solo en un fragmento nuevo `apps/api/openapi/<flujo>.yaml`. `UnirContrato` los une. No se reordena `apps/api/openapi.yaml` ni el fragmento de otro flujo. Un path repetido rompe el arranque: no se duplica. |
| Migraciones | Rango reservado de la tabla. Un número que no se usa se deja libre. Nadie lo «aprovecha». |
| Datos | No se carga un nombre real. No se borra catastro para «probar de cero». |

Puntos de choque conocidos (el integrador de la ola los resuelve en este orden, sin reescribir el flujo):

| Archivo | Quién lo toca | Los demás |
|---|---|---|
| `apps/web/src/App.tsx` | Ola 0 y ola 4 (claridad) | Nadie. La selección del mapa vive en `map/seleccionActividad.ts`. La cola offline se engancha en `offline/enganchar.ts`. |
| `apps/web/src/panel/Modulos.tsx` | Ola 0 lo parte y deja de ser el lugar de trabajo | Nadie lo vuelve a crecer. |
| `apps/web/src/panel/Labores.tsx` | Ola 1 nomenclatura (etiquetas), luego ola 2 alta (formulario) | Los flujos posteriores importan un componente nuevo; no reescriben el formulario. |
| `apps/web/src/types.ts` | `ROLES`: flujo de cuentas. `LAYERS`: flujo de nomenclatura, después de cuentas | No se edita en paralelo. |
| `backend/.../routes/routes.go` y `container.go` | Integrador de la ola | Los flujos no añaden campos de dig. |
| `permisos.service.go` | Solo `cierre/o1-cuentas` | Catálogos no lo tocan. |
| `evidencia.repository.go` (`Guardar`) | Solo `cierre/o3-trazabilidad` | S3 no cambia el `INSERT`. |
| `intervencion.repository.go` | Trazabilidad (eventos) y alta (columnas nuevas) van en olas distintas | En la ola 3 los filtros nuevos van en `filtro_actividades.go`, no en el repositorio de eventos. |
| `apps/api/openapi/operacion.yaml` | Nadie en este plan | Los paths nuevos van a un fragmento con el nombre del flujo. |

### Migraciones reservadas

La última aplicada en el corte es `048`. Reservado desde `049`.

| Flujo | Números |
|---|---|
| `cierre/o1-cuentas` | 049–051 |
| `cierre/o1-catalogos-estados` | 052–055 |
| `cierre/o1-reserva-patch` | ninguno |
| `cierre/o1-nomenclatura` | ninguno (las etiquetas de rol las escribe 049) |
| `cierre/o2-zonificacion` | 056–059 |
| `cierre/o2-ejemplares` | 060–062 |
| `cierre/o2-alta-actividad` | 063–066 |
| `cierre/o3-offline` | ninguno |
| `cierre/o3-trazabilidad` | 067–069 |
| `cierre/o3-mapa-filtros` | ninguno (lee columnas de 063–066) |
| `cierre/o3-evidencias-s3` | ninguno |
| `cierre/o3-solicitudes` | 071–073 |
| `cierre/o3-riego` | 074, solo si hace falta una columna; si el cálculo es en lectura, se deja libre |
| `cierre/o4-auditoria` | 075–076 |
| `cierre/o5-reportes` | 077–078 |
| `cierre/o5-indicadores` | 079 |
| `cierre/o6-duckdns-https` | ninguno |
| `cierre/o7-insumos` | 080–082 |
| `cierre/o7-vigencia-sector` | 083–085 |
| `cierre/o7-capas` | 086 |
| `cierre/o8-portal-proveedor` | 087 |
| `cierre/o8-concentracion` | 088–089 |

El 070 queda libre a propósito, entre trazabilidad y solicitudes, por si el integrador de la ola 3 necesita un índice y nada más.

## 3. Olas

Must primero, luego Should, luego Could. Dentro de una ola, los flujos marcados «paralelo» no comparten archivos. El que dice «después» espera a que el anterior esté en `develop`.

```text
Ola 0   desacople                          (bloquea a todas)
Ola 1   cuentas ∥ catálogos ∥ reserva
        después: nomenclatura
        después: integrador 1
Ola 2   zonificación ∥ ejemplares
        después: alta de actividad
        después: integrador 2
Ola 3   offline ∥ trazabilidad ∥ mapa ∥ S3 ∥ solicitudes ∥ riego
        después: integrador 3
Ola 4   auditoría
        después: claridad
        después: Playwright
Ola 5   reportes y PDF
        después: indicadores y proceso
Ola 6   DuckDNS y HTTPS                    (puede arrancar tras la ola 0; no espera a la 5)
Ola 7   Should restante                    (cada uno espera su dependencia)
Ola 8   Could
```

La ola 6 no comparte archivos con las demás. Puede correr al mismo tiempo que las olas 1 a 5. Se promueve a producción cuando el certificado existe, no antes de que el login de cuentas (ola 1) ya no anuncie un SSO.

Estimación, la misma de la matriz: **S** un ajuste localizado, **M** API y pantalla de un flujo, **L** módulo nuevo o una definición que el cliente aún no cierra.

## 4. Ola 0 — Desacople

Un solo agente. Sin él, los flujos de UI se pisan en `Modulos.tsx` y en `App.tsx`.

### `cierre/o0-desacople`

- Estimación: **S**. Depende de: nada. Bloquea: todas las olas siguientes.
- Cierra: ningún RF. Prepara el terreno.
- Archivos: `apps/web/src/panel/Modulos.tsx` (se vacía en reexportaciones), archivos nuevos `panel/Login.tsx`, `panel/Reportes.tsx`, `panel/Solicitudes.tsx`, `panel/Riego.tsx`, `panel/Admin.tsx`, `panel/Catalogos.tsx` (mover lo que hoy está en `Modulos.tsx`), `ui/registroModulos.ts` (la lista `MODULOS`), `map/seleccionActividad.ts` (el callback que hoy hace `setModulo('labores')`, con el mismo comportamiento), `offline/enganchar.ts` (el vaciado de cola que hoy está en el montaje de `App.tsx`). `App.tsx` solo importa esos módulos. Backend: no se toca. Migraciones: ninguna.
- Tareas: mover sin cambiar textos, rutas ni permisos. Los tests de `panel/` siguen importando los mismos nombres exportados (`Login`, `Labores` no se mueve). `npm test` en `apps/web` en verde. Ningún cambio visual.
- Aceptación: `node --test` de la web igual que antes. Un diff de comportamiento no existe: las pestañas, el login y el pin siguen iguales. No aparece una migración nueva.

Prompt:

```text
Eres el agente cierre/o0-desacople de VerdePUCP, repo Irico17/Areas-verdes-pucp. Rama desde develop: cierre/o0-desacople. No toques main, no hagas force push, no uses el repo del grupo.

Lee docs/PLAN-CIERRE-BACKLOG.md sección «Ola 0» y docs/GLOSARIO-NOMENCLATURA.md. No cambies etiquetas ni comportamiento.

Parte apps/web/src/panel/Modulos.tsx en Login.tsx, Reportes.tsx, Solicitudes.tsx, Riego.tsx, Admin.tsx y Catalogos.tsx. Modulos.tsx reexporta esos nombres para no romper imports. Mueve la lista MODULOS a apps/web/src/ui/registroModulos.ts. Extrae el callback de selección de actividad a map/seleccionActividad.ts (sigue cambiando a la pestaña labores) y el vaciado de la cola al montar a offline/enganchar.ts. App.tsx solo importa. No edites backend/, db/migrations ni apps/api.

Hecho cuando: npm test en apps/web pasa y la UI hace lo mismo que antes.
```

## 5. Ola 1 — Cuentas, catálogos, reserva y nombres

Paralelo: cuentas, catálogos y reserva. Nomenclatura después de cuentas. Integrador al final.

### `cierre/o1-cuentas` — M — paralelo

Cierra: RF-01, RF-02, RF-03, HU-01, HU-02, RF-01-CA2, RF-01-CA3, RF-02-CA2, RF-03-CA1, RF-03-CA2, DEC-01, DEC-02, DEC-06, y la parte de claves y de texto de RNF-04 y RNF-04-CA2. La parte HTTPS de RNF-04 queda en la ola 6. La parte «leer catálogos desde la base» de RNF-05 queda en catálogos; este flujo deja la matriz de permisos leída desde `permisos`.

Archivos: `groups/accesos.group.go`, controlador y caso de uso de usuario, `permisos.service.go`, `usuario.usecase.go` (borrar la frase del SSO), `panel/Admin.tsx`, `panel/Login.tsx` (la frase del SSO), `types.ts` solo el arreglo `ROLES`, tests de esos paquetes, fragmento `apps/api/openapi/o1-cuentas.yaml`. Migraciones 049–051: `debe_cambiar_password`, etiquetas exactas de rol con `UPDATE` auditado en `cambios`, semilla de permisos de jefatura para `usuarios` con `ON CONFLICT DO NOTHING`. No se renombran códigos.

Tareas:

1. `POST /areas-verdes/v1/accesos/usuarios` (alta), `PATCH` (rol, nombre, activo, clave), baja lógica. Jefatura y admin. Capataz e Ingeniería / Coordinación reciben 403.
2. Alta obliga clave propia y `debe_cambiar_password`. El login pide el cambio antes de entrar al mapa.
3. `Permite()` lee `permisos` y `roles.activo`. Un rol inactivo no entra. Editar la matriz desde Administración escribe `cambios` y no exige redeploy. El arranque sigue sin `DELETE`.
4. Capataz ya no crea especies (deja de llegar con `registrar` a catastro de especies). Jefatura administra cuentas. La matriz que el libro aún no valida con el cliente se deja editable, con la semilla actual como punto de partida, no como una matriz inventada.
5. Etiquetas: «Ingeniería / Coordinación» y «Jefatura de sección». Códigos `coordinacion` y `jefatura`.
6. Textos de SSO fuera del login, de Administración y del JSON de `GET /accesos/usuarios`.

Aceptación:

- Test Go: alta por jefatura 201; alta por capataz 403; `POST` que antes era 404 ahora existe; usuario inactivo no abre sesión; `Permite` cambia al actualizar una fila de `permisos` sin reiniciar el proceso (el test recarga).
- Curl: jefatura crea una cuenta, esa cuenta entra con su clave, el JSON de listado no contiene «SSO». Login viejo de `coordinacion` sigue 200.
- Playwright: el chip dice «Ingeniería / Coordinación» o «Jefatura de sección»; Administración muestra el formulario de alta; el login no menciona la universidad.

Prompt:

```text
Eres el agente cierre/o1-cuentas de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop, que ya tiene el desacople de cierre/o0-desacople: cierre/o1-cuentas. Sin main, sin force push, sin repo del grupo.

Lee docs/PLAN-CIERRE-BACKLOG.md (reglas comunes y este flujo), docs/GLOSARIO-NOMENCLATURA.md y la fila RF-01 de docs/MATRIZ-BACKLOG-V2.md.

Implementa la gestión de cuentas de la jefatura: POST y PATCH /areas-verdes/v1/accesos/usuarios, baja lógica, clave propia, debe_cambiar_password. Jefatura y admin; el resto 403. Permite() lee la tabla permisos y roles.activo; el editor queda en panel/Admin.tsx y audita en cambios. El arranque no borra permisos. Quita toda frase de SSO pendiente (login, admin, usuario.usecase.go). Etiquetas exactas, sin cambiar códigos: Capataz; Ingeniería / Coordinación; Jefatura de sección; Administrador del sistema. Migraciones solo 049-051, aditivas e idempotentes. No edites catalogos, reservas, Labores.tsx, App.tsx ni tipos de LAYERS. OpenAPI solo en apps/api/openapi/o1-cuentas.yaml. Nombres ficticios. No borres datos.

Hecho cuando: los tests Go del flujo pasan; curl de jefatura crea una cuenta y capataz recibe 403; Playwright muestra el alta y no la palabra SSO.
```

### `cierre/o1-catalogos-estados` — L — paralelo con cuentas y con la reserva

Cierra: RF-24, RF-16, HU-19 (catálogos), HU-13, HU-29 en lo de estados, RF-16-CA1, RF-16-CA3, RF-24-CA1, RF-24-CA2, RNF-05 en lo de estados, tipos y clases (los permisos los cierra cuentas).

Archivos: dominio catálogo (`catalogo.group.go`, enum de clases, caso de uso), `intervencion.usecase.go` solo la validación de `SetEstado` y transiciones, `panel/Catalogos.tsx`, `operacion.ts` (deja de ser la lista maestra: carga el catálogo y traduce), tests. No toca `permisos.service.go` ni `Admin.tsx`. Migraciones 052–055: clases que faltan en el enum de la API (`clase_actividad`, plaga, producto fitosanitario, frecuencia, sede, cuartel, sector de capataz como catálogo de nombres), ítems de estado con las etiquetas del libro, `clase_actividad` con manejo fitosanitario e inspección y monitoreo. Los códigos de estado viejos se quedan; se añade la etiqueta. «Bloqueada» no se borra: queda inactiva o marcada provisional hasta que el cliente decida (pregunta abierta de la matriz).

Tareas:

1. Editar el nombre de un ítem de catálogo (hoy solo alta y desactivar) con historial. Desactivar no borra.
2. `clase_actividad` entra a las clases que `POST /catalogos` acepta. `GET /catalogos` la devuelve.
3. `PATCH` de estado rechaza un slug que no esté activo en `catalogos` clase `estado`. El curl que hoy acepta `ejecutado` fuera de catálogo pasa a 400. Si `ejecutado` se da de alta en el catálogo, 200, y la UI muestra «Ejecutado», no el slug.
4. Transiciones mínimas documentadas en código como tabla de datos, no como `if` sueltos: por iniciar → en proceso → ejecutado → cerrado; cancelado y archivado desde los abiertos. Una tercerizada sigue sin cerrarse sin orden (eso ya existe; no se afloja).
5. La web pinta los estados desde el API.

Aceptación:

- Test Go: estado desconocido 400; estado del catálogo 200; no se puede cerrar saltando desde por iniciar si la tabla de transiciones lo niega; editar nombre de catálogo deja fila en `cambios` y no borra la anterior.
- Curl: `GET /catalogos` incluye `clase_actividad`. `PATCH estado=ejecutado` 400 si no está en catálogo y 200 si se creó el ítem, con etiqueta «Ejecutado» en el JSON.
- Playwright: el selector de estado no muestra la palabra cruda; Catálogos permite corregir un nombre.

Prompt:

```text
Eres el agente cierre/o1-catalogos-estados de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop (con o0): cierre/o1-catalogos-estados. Sin main, sin force push, sin repo del grupo. No edites permisos.service.go, accesos, reservas ni App.tsx: otro agente los lleva en paralelo.

Lee docs/PLAN-CIERRE-BACKLOG.md, docs/GLOSARIO-NOMENCLATURA.md y las filas RF-16 y RF-24 de la matriz.

Estados y catálogos salen de la base. SetEstado solo acepta un código activo de catalogos clase estado y una transición declarada en datos. Etiquetas visibles: Por iniciar, En proceso, Ejecutado, Cerrado, Cancelado, Archivado. No renombres códigos. No borres «bloqueada»: déjala inactiva o provisional. clase_actividad entra al alta de catálogos, con Manejo fitosanitario e Inspección y monitoreo, sin borrar las clases ya cargadas. Se puede corregir el nombre de un ítem, con cambios, sin DELETE. La web (operacion.ts y panel/Catalogos.tsx) lee el catálogo. Migraciones solo 052-055, aditivas. OpenAPI en apps/api/openapi/o1-catalogos.yaml. Nombres ficticios.

Hecho cuando: curl PATCH estado fuera de catálogo responde 400; con el ítem creado, 200 y la etiqueta «Ejecutado»; GET /catalogos incluye clase_actividad; hay test Go de la transición.
```

### `cierre/o1-reserva-patch` — S — paralelo

Cierra: DEC-11.

Archivos: `inventario_campo.repository.go` (el `PATCH` de reserva) y su test. Nada más. Sin migración. Sin frontend, salvo que un test de `calendario` cubra el cuerpo parcial; si no existe, el test es de Go.

Tareas: si el JSON trae solo `hora_inicio`, se conserva `hora_fin` de la fila, y al revés. Se valida el intervalo ya fusionado. Si el resultado queda invertido, 400 y la fila no cambia.

Aceptación: test de repositorio. Cuerpo `{hora_inicio}` sobre una reserva 09:00–11:00 responde 200 y la fila queda 10:00–11:00 (o la hora enviada). Cuerpo con fin anterior al inicio guardado responde 400 y la fila sigue igual. El curl de la matriz (solo `hora_inicio` → 400) deja de cumplirse.

Prompt:

```text
Eres el agente cierre/o1-reserva-patch de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop (con o0): cierre/o1-reserva-patch. Sin main, sin force push, sin repo del grupo.

Lee DEC-11 en docs/MATRIZ-BACKLOG-V2.md. El PATCH de reserva hoy compara hora_inicio y hora_fin del cuerpo y rechaza si falta una. Fusionalas con la fila guardada y valida el intervalo resultante. Solo edita el repositorio de inventario de campo y su test. Sin migraciones, sin frontend, sin otros dominios. No borres reservas.

Hecho cuando: un test demuestra 200 con solo hora_inicio y 400 si la hora resultante invierte el intervalo, sin cambiar la fila en el 400.
```

### `cierre/o1-nomenclatura` — S — después de `o1-cuentas`

Cierra: los renombres 1–8 y 15–16 de `docs/GLOSARIO-NOMENCLATURA.md`. No cierra un RF solo; deja la UI alineada antes de que la ola 2 escriba formularios nuevos. Los renombres 9, 11 y 12 (roles y SSO) ya los hizo cuentas. El 10 lo hace claridad. El 13 lo hace zonificación. El 14 lo hace catálogos. El 17 lo hacen los reportes.

Archivos: `types.ts` solo `LAYERS` (cuentas ya terminó `ROLES`), `map/categorias.ts`, etiquetas en `Labores.tsx` (título y «Equipo»), `importaciones.ts`, `inventarioCapas.ts`, pista de edificios. No cambia ids de capa, rutas ni lógica. No edita `ROLES`.

Aceptación: `npm test` de categorias y de labores. Playwright o el test de render: la pestaña dice «Actividades», el filtro dice «Cuadrilla», la capa ya no dice «Zonas», los usos muestran el texto de la propiedad `Uso` del GeoJSON. Búsqueda en `apps/web/src` de la etiqueta visible «Equipo» y del título «Labores»: cero en JSX (el nombre de archivo puede quedar).

Prompt:

```text
Eres el agente cierre/o1-nomenclatura de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop, que ya tiene cierre/o1-cuentas: cierre/o1-nomenclatura. Sin main, sin force push, sin repo del grupo.

Lee docs/GLOSARIO-NOMENCLATURA.md y aplica solo los renombres 1 a 8, 15 y 16 de la tabla «Renombres necesarios». No toques ROLES ni las frases de SSO (ya cerradas). No cambies ids, rutas ni comportamiento. No edites backend ni migraciones. La pestaña y el título pasan de Labores a Actividades; Equipo pasa a Cuadrilla; la capa zonas se etiqueta «Sectores de capataz»; los usos usan el texto de la fuente, no el acortado.

Hecho cuando: los tests de la web pasan y un render muestra Actividades, Cuadrilla y Sectores de capataz.
```

### Integrador de la ola 1 — S — después de los cuatro

No implementa reglas. Revisa que `routes.go` y `container.go` no hayan sido editados por los tres flujos de backend (no deberían). Si cuentas añadió rutas dentro de `accesos.group.go`, no hay nada que registrar. Corre `go test ./...` en `backend/app` y `npm test` en `apps/web`. Si un fragmento OpenAPI repite un path, lo corrige el integrador sin mover paths viejos.

Prompt:

```text
Eres el integrador cierre/o1-integracion de VerdePUCP, rama desde develop ya con o1-cuentas, o1-catalogos-estados, o1-reserva-patch y o1-nomenclatura. No añadas funcionalidad. Corre go test en backend/app y npm test en apps/web. Si hay conflicto de registro en routes.go, container.go o un path OpenAPI duplicado, resuélvelo sin cambiar la regla de negocio. No toques main ni hagas force push.
```

## 6. Ola 2 — Territorio y alta de la actividad

Paralelo: zonificación y ejemplares. El alta espera a zonificación (selector de lugar) y a catálogos (clase y tipo).

### `cierre/o2-zonificacion` — L — paralelo con ejemplares

Cierra: RF-06, HU-19 en lo de sectores y lugares, RF-06-CA1, RF-06-CA3, RF-06-CA4, RF-06-CA5 en la medida en que el shape de cuarteles no existe: la capa queda preparada, vacía, de solo lectura, y el CA5 se marca cubierto en mecanismo. Si el cliente no entregó el shape, la pantalla dice «sin archivo de cuarteles», no inventa geometría. RF-34 (vigencia y sugerencia) no entra aquí.

Archivos nuevos: `groups/zonificacion.group.go`, controlador y caso de uso propios, `panel/SelectorLugar.tsx`, `panel/SectoresCapataz.tsx`. Puede editar `CatastroEditor.tsx` y `map/categorias.ts` solo para colgar el selector. No edita el bloque de ejemplares de `catastro.group.go` ni `Labores.tsx` (el alta lo hará el flujo siguiente importando `SelectorLugar`). Migraciones 056–059: catálogo de sector de capataz con nombre (sin CHECK nuevo), vías como capa vacía importable, cuartel histórico sin geometría obligatoria. No se borra `lugar_libre`.

Tareas:

1. CRUD de sector de capataz (alta, nombre, activo) con historial e importación. El color del mapa sale de ese catálogo, no de la lista fija.
2. Lugares siguen en catálogo. Componente `SelectorLugar` para que el alta deje de usar texto libre. Los valores ya guardados en `lugar_libre` se leen.
3. Capa «Vías» apagada por defecto, importable por GeoJSON, sin fila inventada.
4. Capa «Cuarteles (histórico)» de solo lectura. Sin shape, estado vacío explicado.
5. Edificios se eligen como referente al marcar un punto (id de edificio, no texto suelto).

Aceptación:

- Test Go del CRUD de sector y de que un lugar desconocido no se inserta como texto nuevo en el catálogo.
- Curl: `POST` de un sector 201; segundo `POST` con el mismo código 409 o actualización, sin duplicar; `DELETE` no existe.
- Playwright: el mapa dice «Sectores de capataz»; el editor no ofrece un input libre de lugar cuando se usa `SelectorLugar` en catastro.

Prompt:

```text
Eres el agente cierre/o2-zonificacion de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 1 integrada: cierre/o2-zonificacion. Sin main, sin force push, sin repo del grupo.

Lee RF-06 en la matriz, docs/GLOSARIO-NOMENCLATURA.md y docs/MAPA-DATOS-Y-EDICION.md secciones 4.2 a 4.5. Otro agente edita ejemplares en paralelo: no toques catastro.group.go en el bloque de ejemplares, ni ejemplar.repository.go, ni Labores.tsx.

Crea groups/zonificacion.group.go (no lo registres en routes.go: lo hará el integrador) con CRUD de sector de capataz, capa de vías vacía e importable, y cuarteles históricos de solo lectura sin inventar geometría. Exporta panel/SelectorLugar.tsx. lugar_libre queda solo lectura. Migraciones solo 056-059, aditivas, sin DELETE ni TRUNCATE. OpenAPI en apps/api/openapi/o2-zonificacion.yaml. Nombres ficticios. Etiqueta visible: sector de capataz, nunca «zona» para ese polígono.

Hecho cuando: curl crea un sector y no hay ruta DELETE; el mapa no usa una lista fija de sectores; los tests Go del paquete pasan.
```

### `cierre/o2-ejemplares` — M — paralelo con zonificación

Cierra: RF-04, HU-03, HU-04, RF-04-CA1, RF-04-CA3, RF-05, HU-05, RF-05-CA1.

Archivos: repositorio y caso de uso de ejemplar, las líneas de ejemplar en `catastro.group.go` (PATCH y la ficha), `panel/Ejemplares.tsx` nuevo, una entrada al final de `registroModulos.ts`. No edita `CatastroEditor.tsx`, `zonificacion.group.go` ni `Labores.tsx`. Migraciones 060–062: columna nullable de sector de capataz o cuartel en `ejemplares` (FK, no texto libre), sin rellenar a mano las 1081 filas con un valor inventado. Salud editable. Historial en `cambios`. Importación de la ficha ya existe: el PATCH no debe impedir el upsert por `origen_ref`.

Tareas:

1. `PATCH /catastro/ejemplares/:id` con código, especie, salud, coordenadas, lugar y sector o cuartel.
2. Pantalla de consulta y edición, más el historial de códigos que la API ya guarda (RF-05).
3. Capataz no edita ejemplares (el flujo de cuentas le quitó ese permiso; este flujo no se lo devuelve). Ingeniería / Coordinación sí, si tiene `registrar`.

Aceptación:

- Test Go del PATCH y de la recodificación que conserva el código anterior.
- Curl: PATCH de salud 200; sin sesión 401; capataz 403.
- Playwright: se abre un ejemplar, se cambia la salud, se ve el código anterior después de recodificar.

Prompt:

```text
Eres el agente cierre/o2-ejemplares de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 1: cierre/o2-ejemplares. Sin main, sin force push, sin repo del grupo.

Lee RF-04 y RF-05 en la matriz y el glosario (término Ejemplar). Otro agente hace sectores en paralelo: no edites CatastroEditor.tsx, Labores.tsx ni crees zonificacion.group.go. Tú solo añades PATCH de ejemplar en el bloque de ejemplares de catastro.group.go y el paquete de ejemplar. Pantalla nueva panel/Ejemplares.tsx y una línea al final de registroModulos.ts. Migraciones solo 060-062: FK nullable de sector o cuartel, sin inventar datos para las filas ya cargadas. Historial en cambios. Recodificar muestra el código viejo. Capataz 403. OpenAPI en apps/api/openapi/o2-ejemplares.yaml. No borres ejemplares.

Hecho cuando: curl PATCH de salud responde 200 con sesión de coordinación y 403 con capataz; el test de recodificación sigue en verde; la pantalla enseña el código anterior.
```

### `cierre/o2-alta-actividad` — L — después de zonificación y de catálogos

Cierra: RF-30, HU-27, HU-28, RF-30-CA1 (segundo nivel), RF-30-CA2, DEC-07, RF-08 en taxonomía y personal (RF-08-CA1, RF-08-CA2, la mitad de personal de RF-08-CA3). Los insumos de RF-08-CA3 y RF-10 quedan en la ola 7. RF-09 ya está cumplido: no se reabre.

Archivos: DTO y repositorio de intervención (columnas nuevas), `panel/Labores.tsx` solo el formulario de alta, que debe importar `SelectorLugar` y los catálogos de clase y tipo. No reescribe filtros ni la bitácora. Migraciones 063–066: columnas nulas `subtipo`, `codigo_externo`, `unidad_solicitante`, `nivel_riesgo`, `fecha_programada`, `cantidad`, más `personal_labor` con endpoint. `origen` ya existe: se persiste el valor que hoy se ignora. Riesgo: bajo y alto. «Medio» no se ofrece hasta que el cliente lo confirme (pregunta abierta); el catálogo puede recibir el ítem después sin migración de CHECK.

Tareas:

1. El POST guarda origen, código externo, unidad solicitante, nivel de riesgo, fecha programada, cantidad y subtipo. El GET de la feature los devuelve. Un campo desconocido sigue sin colarse en silencio: los de esta lista sí se mapean.
2. El formulario pide clase y tipo (dos niveles), en ese orden, desde el catálogo. Personal de la actividad: uno o más nombres ficticios, sin cuenta.
3. Zona de supervisión seleccionable. Lugar por `SelectorLugar`, no por texto libre.
4. Altas seguidas sin salir del mapa: se conserva el gesto que ya existe.

Aceptación:

- Test Go: el POST con esos campos responde 201 y el GET los trae. Fuera del campus, 400. Capataz, 403 al crear.
- Curl: el mismo cuerpo que la matriz (nivel_riesgo, fecha_programada, cantidad, origen) ya no desaparece.
- Playwright: el alta muestra clase, tipo, riesgo, fecha programada, cantidad, unidad, código externo y personal. No hay input de lugar libre.

Prompt:

```text
Eres el agente cierre/o2-alta-actividad de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o1-catalogos-estados y o2-zonificacion ya integrados: cierre/o2-alta-actividad. Sin main, sin force push, sin repo del grupo.

Lee RF-30, RF-08 y DEC-07 en la matriz, y el glosario (Actividad, Clase de actividad, Tipo de actividad, Personal). Cierra el alta: el POST /areas-verdes/v1/operacion/actividades persiste origen, codigo_externo, unidad_solicitante, nivel_riesgo (bajo|alto, sin medio), fecha_programada, cantidad y subtipo. Clase y tipo salen del catálogo, dos niveles. Personal con nombres ficticios, sin cuenta de operario. Usa SelectorLugar; no aceptes lugar libre nuevo. Migraciones solo 063-066, columnas nulas, sin borrar actividades. Edita el formulario de alta en Labores.tsx, no los filtros ni la bitácora. No implementes insumos (eso es o7-insumos). OpenAPI en apps/api/openapi/o2-alta-actividad.yaml.

Hecho cuando: el curl de la matriz, que antes ignoraba esos campos, los devuelve en el GET; Playwright muestra los campos; capataz recibe 403 al crear.
```

### Integrador de la ola 2 — S

Registra `zonificacion.group.go` en `routes.go` y `container.go` (una línea cada uno). Confirma que ejemplares y el alta no chocaron en `catastro.group.go`. Corre los tests.

Prompt:

```text
Eres el integrador cierre/o2-integracion. develop ya tiene o2-zonificacion, o2-ejemplares y o2-alta-actividad. Registra zonificacion.group.go en routes.go y container.go si el flujo no pudo. No cambies reglas. go test en backend/app y npm test en apps/web. Sin main y sin force push.
```

## 7. Ola 3 — Campo, mapa, evidencias y terceros

Los seis flujos son paralelos. El alta de la ola 2 ya está en `develop`, así que nadie reescribe ese formulario.

### `cierre/o3-offline` — M — paralelo

Cierra: RF-12, HU-07/HU-09/HU-21/HU-31 en lo offline, RF-12-CA1, RF-12-CA2, RNF-01, RNF-01-CA2.

Archivos: `offline/queue.ts`, `offline/enganchar.ts`, `offline/cola.test.ts`, y el encolado en `Riego.tsx`, `Poda.tsx`, `Vivero.tsx` y el guardado de ficha. No edita el formulario de alta ni `App.tsx`. Sin migración. El capataz no crea actividades: la cola de altas no se le ofrece. Sí encola cambios de estado, avances, riego y ficha de lo que ya es suyo.

Tareas: `window` `online` vacía la cola de estados y de registros, igual que las evidencias. El botón manual se queda. 409 sigue avisando sin duplicar.

Aceptación: test de la cola (ya existe) más un test del enganche: al evento `online` se llama a `flush`. Playwright con red cortada: un cambio de estado queda «en cola» y, al recuperar la red, desaparece de la cola sin pulsar el botón.

Prompt:

```text
Eres el agente cierre/o3-offline de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 2 integrada: cierre/o3-offline. Sin main, sin force push, sin repo del grupo.

Lee RF-12 y RNF-01. La API ya es idempotente por UUID: no la cambies. Engancha el vaciado de la cola de actividades y de estados al evento online, en offline/enganchar.ts. Encola también avances, riego y ficha. No ofrezcas alta al capataz. No edites App.tsx, el formulario de alta, ni evidencia.repository.go. Sin migraciones. Tests en offline/cola.test.ts.

Hecho cuando: un test dispara online y la cola se vacía; Playwright corta la red, encola un estado y lo envía al volver, sin el botón.
```

### `cierre/o3-trazabilidad` — L — paralelo

Cierra: RF-31, HU-26/HU-29/HU-31/HU-32 en la cadena, RF-31-CA1, RF-31-CA2, RF-31-CA4, RF-19-CA3 (el vínculo; el cubo S3 es el otro flujo).

Archivos: inserción de eventos en el repositorio de intervención, `Guardar` de evidencias solo para escribir `evento_id` y aceptar `solicitud_id`, `panel/Bitacora.tsx` nuevo. No edita el almacenamiento S3 ni `EvidenciasCampo.tsx` (el otro flujo). No edita el listado de filtros. Migraciones 067–069: ampliar el CHECK de tipos de evento de forma aditiva (inicio, supervision, derivacion, observacion, conformidad; los siete actuales se conservan), columna o nota `capataz_anterior`. No se borra un evento.

Tareas:

1. Cada hito del libro genera un evento con usuario, fecha y texto.
2. Reasignar guarda el capataz anterior y el nuevo.
3. Un avance crea evento.
4. La evidencia puede ir con `evento_id`. El `INSERT` lo persiste. La bitácora muestra la miniatura en ese evento.
5. Archivar sigue siendo baja lógica con confirmación (ya cumplido): no se toca esa regla para aflojarla.

Aceptación:

- Test Go: reasignar deja `capataz_anterior`; evidencia con `evento_id` se lee en el timeline; un tipo de evento nuevo pasa el CHECK.
- Curl: `POST` de evidencia multipart con `evento_id` 201 y el GET del timeline lo trae.
- Playwright: la bitácora muestra «Reasignada» con los dos nombres ficticios y la foto bajo ese evento.

Prompt:

```text
Eres el agente cierre/o3-trazabilidad de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 2: cierre/o3-trazabilidad. Sin main, sin force push, sin repo del grupo.

Lee RF-31 y RF-19-CA3. Amplía los tipos de evento (inicio, supervisión, derivación, observación, conformidad) sin quitar los siete actuales. Guarda el capataz anterior al reasignar. Un avance crea evento. El INSERT de evidencias persiste evento_id y acepta solicitud_id. No cambies el almacenamiento en disco o S3: otro agente lo hace y no debes editar s3.storage.go ni config de bucket. Pantalla nueva panel/Bitacora.tsx. Migraciones solo 067-069. OpenAPI en apps/api/openapi/o3-trazabilidad.yaml. No borres eventos ni actividades.

Hecho cuando: el test de reasignación guarda el capataz anterior y un curl de evidencia con evento_id lo devuelve en el timeline.
```

### `cierre/o3-mapa-filtros` — M — paralelo

Cierra: RF-29, RF-32, HU-23, HU-24, HU-25, HU-30 en la vista, RF-29-CA1 en lo que se puede sin ortofoto, RF-29-CA2, RF-29-CA3, RF-32-CA1, RF-32-CA2, RNF-11, RNF-11-CA2. La ortofoto no está en `data/raw`: el flujo escribe en `docs/DECISIONES.md` que el mapa base es OSM (MapLibre), que la ortofoto queda pendiente del archivo del cliente, y no simula una. El modo sin teselas es la lista.

Archivos: `map/CampusMap.tsx`, `map/seleccionActividad.ts`, `panel/FiltrosActividad.tsx` nuevo (Labores.tsx solo lo importa), archivo nuevo `filtro_actividades.go` en el dominio de operación (no el repositorio de eventos). Fragmento OpenAPI propio. Sin migración.

Tareas:

1. Tocar un marcador abre el resumen en el popup y el detalle en el panel, sin `setModulo`.
2. Color por estado, icono por clase de actividad (no una letra de tipo).
3. Filtros de oficina: estado, tipo, cuadrilla, sector de capataz, ejecutor, origen, riesgo, fechas. El capataz no ve el filtro de cuadrilla ajena: el servidor ya fuerza la suya.
4. Conmutador mapa / lista. Sin teselas o sin red, la lista queda forzada y se dice por qué.
5. Plano OSM se queda. No se añade Street View.

Aceptación:

- Test Go de la query: combinación de riesgo y fechas; el capataz no recibe filas de otra cuadrilla aunque mande el parámetro.
- Curl: `GET /operacion/actividades?nivel_riesgo=alto&desde=2026-01-01` 200 y no incluye cerradas si `abiertas=1`; con el parámetro de histórico, sí.
- Playwright: el pin no cambia la pestaña; a 390 px el conmutador lista está visible.

Prompt:

```text
Eres el agente cierre/o3-mapa-filtros de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 2: cierre/o3-mapa-filtros. Sin main, sin force push, sin repo del grupo.

Lee RF-29, RF-32 y RNF-11, y el glosario. El resumen del marcador se queda en el mapa: edita map/seleccionActividad.ts y CampusMap.tsx para no cambiar de pestaña. Icono por clase de actividad, color por estado. Filtros en un componente nuevo panel/FiltrosActividad.tsx (Labores.tsx solo importa) y query en un archivo nuevo filtro_actividades.go. No edites el INSERT de eventos ni evidencia.repository.go. Criterios: estado, tipo, cuadrilla, sector de capataz, ejecutor, origen, riesgo, fechas. Conmutador mapa/lista; sin red, lista forzada. Mapa base OSM; no inventes ortofoto; anota la decisión OSM en docs/DECISIONES.md. Sin migración. OpenAPI en apps/api/openapi/o3-mapa-filtros.yaml.

Hecho cuando: curl filtra por riesgo y fechas; Playwright en 390 px abre el pin sin cambiar de pestaña y puede pasar a lista.
```

### `cierre/o3-evidencias-s3` — M — paralelo

Cierra: DEC-10 y la parte de almacén de RF-19 / RF-19-CA1. El `evento_id` lo escribe el flujo de trazabilidad: este flujo no toca el `INSERT`.

Archivos: `s3.storage.go`, `config.go` en lo de `EVIDENCIAS_BUCKET`, `deploy/env/*.env.example` (variable vacía documentada, sin secretos), `docs/DESPLIEGUE.md` o `docs/DEPLOY-RUNNER.md` donde hoy dice que S3 es fase posterior, test del almacenamiento con un falso cliente. No edita `evidencia.repository.go`. El archivo en disco sigue siendo el camino si la variable no está (develop local). Qa y producción documentan el cubo privado. No se sube un secreto al repo.

Tareas:

1. Con bucket configurado, el byte va al cubo y la base guarda la clave, no la URL pública.
2. Sin bucket, disco, como hoy, y el log lo dice una vez.
3. Prueba de subida contra un doble del cliente S3 (no hace falta un cubo real en CI).
4. La UI no promete S3 si el API no lo confirma. Si hace falta un campo `almacen` en la respuesta, se añade en el DTO de salida sin cambiar el `INSERT` de columnas de negocio: coordinar en el integrador si trazabilidad también toca el DTO. Para no chocar, el campo lo añade este flujo en un mapper de salida nuevo, no en `Guardar`.

Aceptación: test de unidad del storage (bucket llama a PutObject; sin bucket escribe en el directorio temporal). `grep` de claves AWS en el diff: cero. Documentación de las variables `EVIDENCIAS_BUCKET` y credenciales por entorno.

Prompt:

```text
Eres el agente cierre/o3-evidencias-s3 de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 2: cierre/o3-evidencias-s3. Sin main, sin force push, sin repo del grupo.

Lee DEC-10 y RF-19. Otro agente escribe evento_id en evidencia.repository.go: no edites ese archivo ni el INSERT. Tú cierras el cubo: si EVIDENCIAS_BUCKET está definido, los bytes van a S3 privado y la base guarda la clave; si no, disco, como hoy. Test con un cliente falso. Documenta la variable en los env.example sin poner secretos. No inventes un bucket real. Sin migraciones.

Hecho cuando: el test cubre las dos ramas y el diff no contiene una clave de AWS.
```

### `cierre/o3-solicitudes` — M — paralelo

Cierra: RF-11, HU-11, RF-11-CA1, RF-11-CA2, RF-13, HU-12 de tercerizados, RF-13-CA1. El archivo del proveedor (RF-14, RF-13-CA2) es la ola 7, para no compartir la pantalla de la orden con otro flujo de esta ola. Este flujo deja el gancho: la orden ya muestra sus evidencias si `orden_id` viene.

Archivos: dominio de atención (`atencion.group.go`, solicitud, orden), `panel/Solicitudes.tsx`. Migraciones 071–073: `cantidad_solicitada` y `cantidad_ejecutada` (se copia `cantidad` a solicitada sin perderla), `lugar_id` o punto nullable, catálogo de fuente y de estado de solicitud en datos (se deja el CHECK hasta que el catálogo tenga los mismos códigos, y el flujo de catálogos ya permitió altas; aquí se migra la validación a catálogo sin `DELETE`). Empresa y frecuencia pasan a FK de catálogo, no a texto nuevo. Lo ya escrito en texto se conserva y se muestra.

Tareas:

1. Dos cantidades. Ubicación: pin o lugar de catálogo.
2. Empresa y frecuencia desde catálogo. Conformidad se sigue guardando.
3. No hay portal de proveedor.

Aceptación:

- Test Go: cantidad ejecutada distinta de la pedida; fuente fuera de catálogo 400.
- Curl: POST con lat/lon dentro del campus 201; fuera 400.
- Playwright: el formulario de solicitud no es solo un texto de lugar; la orden elige empresa de una lista.

Prompt:

```text
Eres el agente cierre/o3-solicitudes de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con las olas 1 y 2: cierre/o3-solicitudes. Sin main, sin force push, sin repo del grupo.

Lee RF-11 y RF-13. En panel/Solicitudes.tsx y en el dominio de atención: cantidad solicitada y cantidad ejecutada, punto en el campus o lugar de catálogo, empresa y frecuencia de catálogo. Conserva el texto ya cargado. No construyas el portal del proveedor ni la pantalla de adjuntar el reporte (eso es o7-adjunto-proveedor). Migraciones solo 071-073, aditivas, sin DELETE. OpenAPI en apps/api/openapi/o3-solicitudes.yaml. Nombres de empresa ficticios en tests.

Hecho cuando: curl crea una solicitud con las dos cantidades y un punto dentro del campus; un punto fuera responde 400; Playwright muestra el selector de empresa.
```

### `cierre/o3-riego` — M — paralelo

Cierra: RF-26, HU-21, RF-26-CA2. RF-26-CA1 ya está cumplido. La fórmula oficial no existe (pregunta abierta, junto a RF-23). Se publica una fórmula **provisional** y la pantalla lo dice.

Definición provisional, solo para poder ver un número, no como acuerdo con el cliente: cobertura del ciclo = sectores de capataz con al menos un riego en el ciclo / sectores de capataz activos. El ciclo es el mes calendario (America/Lima) hasta que el cliente diga otra cosa. El texto de la UI: «Cobertura provisional: N %. Pendiente de validar con la jefatura de sección.»

Archivos: caso de uso de riego, `panel/Riego.tsx`. El sector deja de ser texto libre: usa el catálogo de sector de capataz. Migración 074 solo si hace falta guardar el sector como FK en `riego_registros`; si la columna actual es texto, se añade `sector_id` nullable y no se borra el texto viejo. Si no hace falta, 074 queda sin archivo.

Aceptación: test del cálculo con dos sectores, uno regado, resultado 50 y la leyenda de provisional. Curl `GET` del resumen de riego trae `provisional: true`. Playwright ya no muestra solo «definición pendiente» sin número: muestra el número y la advertencia.

Prompt:

```text
Eres el agente cierre/o3-riego de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o2-zonificacion: cierre/o3-riego. Sin main, sin force push, sin repo del grupo.

Lee RF-26. El sector del formulario es el sector de capataz (catálogo), no un texto libre. Calcula una cobertura provisional: sectores con al menos un riego en el mes (America/Lima) dividido por sectores activos. La pantalla dice que es provisional y que la jefatura de sección no la ha validado. No la presentes como fórmula oficial. No implementes el tablero de indicadores (o5-indicadores). Migración 074 solo si añades sector_id nullable; no borres el texto ya guardado. OpenAPI en apps/api/openapi/o3-riego.yaml.

Hecho cuando: un test da 50 con dos sectores y uno regado, y el JSON lleva provisional true.
```

### Integrador de la ola 3 — S

Orden de merge si Git choca: evidencias S3, trazabilidad, solicitudes, riego, mapa, offline. Registra grupos nuevos. Corre los tests. El 070 sigue libre.

Prompt:

```text
Eres el integrador cierre/o3-integracion. develop tiene los seis flujos o3. Si Git choca, integra en este orden: evidencias-s3, trazabilidad, solicitudes, riego, mapa-filtros, offline. Registra groups nuevos en routes.go y container.go. No cambies reglas. go test y npm test. Sin main y sin force push.
```

## 8. Ola 4 — Auditoría, claridad y prueba de punta a punta

### `cierre/o4-auditoria` — M — primero en la ola

Cierra: RNF-14, RNF-14-CA1, DEC-04 (pantalla de historial), DEC-05 (el `DELETE` de medidas). RNF-14-CA2 (retención) no se implementa como borrado: la matriz dice que el cliente no fijó el plazo. La pantalla explica que no hay plazo y que no se borra por fecha. RF-18-CA2 (histórico de actividades cerradas por ejemplar y responsable) se cierra aquí en la consulta, apoyada en los filtros que la ola 3 ya expuso.

Archivos: `lote.repository.go` (la reversión de `medidas_palmera`), caso de uso de auditoría, `panel/Auditoria.tsx` nuevo, una línea en `registroModulos.ts`. Migraciones 075–076 solo si la baja lógica de la medida necesita una columna `baja_en`. No se borra la medida ya cargada para «migrar».

Tareas:

1. Revertir un lote deja la medida inactiva en vez de `DELETE`.
2. Pantalla de `GET /auditoria/cambios` para Ingeniería / Coordinación, jefatura y admin. Capataz 403.
3. Consulta de actividades cerradas por ejemplar, responsable, origen y fecha, usando los parámetros de la ola 3 más `ejemplar_id` y `responsable` si faltan. Esos dos parámetros se añaden en `filtro_actividades.go` **después** de que la ola 3 esté integrada; este flujo es el dueño de ese archivo en la ola 4.
4. `Load()` no trunca si hay filas. El camino de carga con datos sigue siendo `etl-lote`.

Aceptación: test de reversión que cuenta filas de `medidas_palmera` antes y después (iguales, con baja lógica). Curl de auditoría 200 para jefatura y 403 para capataz. Playwright abre «Historial» y ve un cambio de la semilla.

Prompt:

```text
Eres el agente cierre/o4-auditoria de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 3 integrada: cierre/o4-auditoria. Sin main, sin force push, sin repo del grupo.

Lee RNF-14, DEC-04 y DEC-05. Sustituye el DELETE FROM medidas_palmera de la reversión de lote por baja lógica. No implementes un job que borre por antigüedad: la retención no está decidida y la pantalla lo dice. Añade panel/Auditoria.tsx sobre GET /auditoria/cambios (capataz 403). Completa el histórico de actividades cerradas por ejemplar y responsable en filtro_actividades.go. Load() no trunca si ya hay filas. Migraciones solo 075-076, y solo si hace falta la columna de baja. OpenAPI en apps/api/openapi/o4-auditoria.yaml.

Hecho cuando: el test de reversión conserva la fila de medida; curl de auditoría distingue jefatura y capataz.
```

### `cierre/o4-claridad` — L — después de auditoría

Cierra: DEC-08, RNF-06, HU-30 en la interfaz de campo, RNF-06-CA1, RNF-06-CA2. También el renombre 10 del glosario (pestaña Admin → Administración). Sigue `docs/PROPUESTA-UI-CLARIDAD.md` en lo que no choque con el glosario: el vocabulario de esta ola es el glosario (Actividad, sector de capataz, cuadrilla), no la palabra «actividad» mezclada con «labor».

Archivos: `App.tsx`, `ui/navegacion.ts`, `ui/registroModulos.ts`, vacíos de los paneles, estilos que hagan falta en los archivos que la propuesta ya usa. No cambia contratos de API. No esconde un 403 dejando el botón activo: el botón que el rol no puede usar no se renderiza.

Tareas, las cinco de la propuesta:

1. Entrada por rol: «Hoy» para capataz (sus actividades y la cola), «Resumen» para oficina.
2. Pestañas según permisos de la sesión, no según `modulosDe` fijo. Capataz no ve Catastro ni Inventario. Nombres de tarea, no de tabla.
3. Riego, poda y vivero salen de debajo de la lista de actividades a «Registros de campo».
4. El mapa queda limpio (el popup ya no cambia de módulo desde la ola 3; aquí se ordena el panel).
5. Vacíos que dicen el siguiente paso. Contraste y nombre accesible en los controles nuevos. Una prueba mínima de accesibilidad (rol, nombre, foco), no un lector de pantalla completo.

Aceptación: Playwright a 390×844 con capataz: ve Hoy, Actividades y el mapa; no ve Catastro. Con jefatura: no hay botón Guardar en una ficha que el servidor niega. `npm test` de navegación actualizado. La pestaña dice «Administración», no «Admin».

Prompt:

```text
Eres el agente cierre/o4-claridad de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o4-auditoria: cierre/o4-claridad. Sin main, sin force push, sin repo del grupo. Eres el único flujo que puede reescribir App.tsx en esta ola.

Lee docs/PROPUESTA-UI-CLARIDAD.md y docs/GLOSARIO-NOMENCLATURA.md. Aplica la propuesta con el vocabulario del glosario (Actividad, no Labor; sector de capataz; cuadrilla; Administración). Entrada por rol, pestañas según permisos reales, riego/poda/vivero en Registros de campo, vacíos que explican el siguiente paso. El capataz en 390 px no ve Catastro ni Inventario. No muestres un botón que el API responde 403. No cambies contratos ni migraciones. El popup del mapa ya no debe volver a cambiar de pestaña.

Hecho cuando: Playwright de capataz a 390×844 no encuentra la pestaña Catastro y sí encuentra Hoy; jefatura no ve un Guardar de catastro.
```

### `cierre/o4-e2e` — M — después de claridad

Cierra: RNF-10, RNF-10-CA2. RNF-10-CA1 ya está cumplido.

Archivos: `apps/web/e2e/` (nuevo), dependencia de Playwright en `apps/web` como devDependency, script `test:e2e`. No cambia producción salvo un `data-testid` si el spec no puede nombrar el control de otra forma; esos id se acuerdan con el texto visible del glosario. CI: un job que corre el spec contra la web de prueba, sin sustituir `go test`.

Recorrido mínimo: login de capataz, viewport 390×844, ve sus actividades, cambia un estado, ve la cola si se corta la red, no entra a Administración.

Aceptación: el spec corre en local y queda invocado desde `package.json`. El README de `apps/web` dice el comando en un párrafo, no un manual nuevo.

Prompt:

```text
Eres el agente cierre/o4-e2e de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o4-claridad: cierre/o4-e2e. Sin main, sin force push, sin repo del grupo.

Añade Playwright en apps/web (devDependency) y un spec del capataz a 390×844: login, sus actividades, cambio de estado, sin la pantalla de administración. Engánchalo a package.json como test:e2e y a CI sin quitar go test. No cambies reglas de negocio. Si necesitas un data-testid, que el texto visible siga el glosario.

Hecho cuando: el comando del spec pasa contra la web de prueba.
```

## 9. Ola 5 — Reportes

### `cierre/o5-reportes` — M — primero

Cierra: RF-20, HU-16 de niveles, RF-20-CA2, RF-22, HU-17, RF-22-CA2. El nivel básico ya está (RF-20-CA1, RF-22-CA1).

Las columnas intermedias y avanzadas no están cerradas con el cliente. Propuesta provisional, etiquetada en el JSON (`provisional: true`) y en la pantalla:

- Intermedio: lo del básico más clase, lugar, cuadrilla, código externo, fechas de solicitud y de atención, cantidad.
- Avanzado: lo anterior más el conteo de eventos del timeline y el nombre ficticio del personal. No se calcula rendimiento (eso es el flujo siguiente).

PDF: `formato=pdf` devuelve `application/pdf`, no el JSON. Un generador en el backend (biblioteca ya permitida por el módulo; si no hay una, se elige una de PDF en puro Go, sin servicio externo).

Archivos: `reporte.group.go` y el caso de uso, `panel/Reportes.tsx`. Migraciones 077–078 solo si hace falta una vista; si el SELECT alcanza, se dejan libres. OpenAPI del flujo.

Aceptación: curl `formato=pdf` con cabecera `%PDF`. Curl `nivel=intermedio` trae cantidad y `provisional: true`. Playwright: tres niveles en el selector y un enlace PDF. Excel y CSV siguen.

Prompt:

```text
Eres el agente cierre/o5-reportes de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 4: cierre/o5-reportes. Sin main, sin force push, sin repo del grupo.

Lee RF-20 y RF-22. formato=pdf debe responder application/pdf. Niveles intermedio y avanzado con las columnas listadas en docs/PLAN-CIERRE-BACKLOG.md (ola 5), marcadas provisional: true, porque el cliente no las cerró. No calcules rendimiento ni indicadores oficiales. Conserva csv y xls. Migraciones 077-078 solo si hacen falta; si no, no las crees. OpenAPI en apps/api/openapi/o5-reportes.yaml. No borres reportes ni actividades.

Hecho cuando: curl formato=pdf empieza por %PDF y nivel=intermedio incluye cantidad y provisional true.
```

### `cierre/o5-indicadores` — L — después de reportes

Cierra: RF-21, RF-23, y deja a RF-15 el número de proveedor para el flujo 7C (este flujo expone el catálogo; 7C lo usa). Fórmulas provisionales, mismas reglas que el riego:

- Cobertura de mantenimiento: actividades no canceladas con estado ejecutado o cerrado en el mes / actividades no canceladas del mes, por zona de supervisión.
- Avance del ciclo de riego: el número que ya calcula `o3-riego`, citado, no recalculado con otra fórmula.
- Frecuencia: actividades por sector de capataz y por mes.
- Atenciones por zona de supervisión.
- Rendimiento: no hay horas en los datos. El indicador existe en el catálogo con estado «sin dato de horas» y no muestra un porcentaje inventado.

Archivos: endpoint nuevo `/reportes/indicadores` en archivos nuevos, `panel/Indicadores.tsx`, migración 079 para la tabla de indicadores (código, nombre, formula_nota, provisional). Importable. Sin borrar.

Aceptación: test de la cobertura de mantenimiento con tres actividades no canceladas, dos en estado cerrado y una por iniciar: el valor es 2/3 y `provisional: true`. Rendimiento no devuelve un número. Playwright muestra la advertencia.

Prompt:

```text
Eres el agente cierre/o5-indicadores de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o5-reportes: cierre/o5-indicadores. Sin main, sin force push, sin repo del grupo.

Lee RF-21 y RF-23. Crea GET /areas-verdes/v1/reportes/indicadores y panel/Indicadores.tsx. Fórmulas provisionales descritas en el plan (cobertura de mantenimiento, frecuencia, atenciones por zona). El riego se cita del flujo o3, no se reescribe. Rendimiento no inventa horas: responde sin dato. Todo lleva provisional true. Migración solo 079. No edites el PDF de o5-reportes. OpenAPI en apps/api/openapi/o5-indicadores.yaml.

Hecho cuando: el test de cobertura usa tres actividades y no presenta la fórmula como oficial.
```

## 10. Ola 6 — DuckDNS y HTTPS

Puede empezar después de la ola 0. No comparte archivos de producto. En producción se cruza con la ola 1: la cookie `Secure` solo se enciende donde hay TLS.

### `cierre/o6-duckdns-https` — M

Cierra: DEC-09, RNF-03, RNF-03-CA1 en el tramo público, y la parte de tránsito de RNF-04 / RNF-04-CA2. No cierra la validación institucional con la PUCP: la observación del libro sigue abierta. El flujo deja el dominio y el certificado repetibles, y un párrafo en `docs/AMBIENTES.md` que dice que la restricción de la universidad no está firmada.

Archivos: `deploy/` (compose, nginx, entrypoint de la web que ya mira los PEM), `docs/DESPLIEGUE.md`, `docs/AMBIENTES.md`, ejemplo de entorno. No se commitea la clave privada ni el token de DuckDNS. `CAMPUS_COOKIE_SECURE=true` solo en el ejemplo del ambiente que tiene certificado.

Tareas:

1. Documentar el nombre DuckDNS, el puerto 443 y la renovación del certificado (el mecanismo concreto: certbot o Caddy, el que ya encaje con `apps/web/docker-entrypoint.sh`; no se añade un segundo reverse proxy si nginx ya elige TLS al ver los PEM).
2. Sin PEM, el compose de develop sigue en HTTP y la cookie no es Secure. Con PEM, HTTPS y cookie Secure.
3. Un ensayo local con un certificado de prueba (autofirmado) demuestra la redirección o el servidor TLS. No se exige que el agente compre un dominio.

Aceptación: `deploy/` no contiene secretos (`git grep` de `BEGIN PRIVATE KEY` vacío). La documentación dice el comando para poner los PEM y el valor de la cookie. Un test o un script de humo, si ya existe `deploy/smoke.sh`, gana un caso «con certificado de prueba la cookie lleva Secure» o la documentación de una prueba manual reproducible. No se abre el puerto de Postgres.

Prompt:

```text
Eres el agente cierre/o6-duckdns-https de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop: cierre/o6-duckdns-https. Sin main, sin force push, sin repo del grupo. No edites backend de negocio ni migraciones.

Lee DEC-09, RNF-03 y RNF-04. Prepara DuckDNS y HTTPS sobre el nginx que ya usa los PEM (apps/web/docker-entrypoint.sh). Sin PEM, develop sigue en HTTP. Con PEM, TLS y CAMPUS_COOKIE_SECURE=true. No commitees llaves ni el token de DuckDNS. Documenta en docs/AMBIENTES.md y docs/DESPLIEGUE.md, incluida la frase de que la validación con la PUCP sigue abierta. No abras Postgres a la red.

Hecho cuando: el diff no tiene una clave privada y la documentación permite repetir el alta del certificado.
```

## 11. Ola 7 — Should que falta

Cada flujo espera la dependencia indicada. Entre ellos son paralelos si sus archivos no se cruzan: insumos, capas, paginación y refresco pueden ir juntos después de sus dependencias. Adjunto espera a solicitudes. Métricas espera a indicadores. Vigencia espera a zonificación y al alta. IA puede ir en cualquier momento después de la ola 2.

### `cierre/o7-insumos` — L — después de `o2-alta-actividad`

Cierra: RF-10, HU-10, RF-10-CA1 y la mitad de insumos de RF-08-CA3.

Migraciones 080–082: catálogo de insumo, consumo por actividad, lugar o área, y periodo. Importable. Nombres de insumo de prueba ficticios o genéricos («abono de demostración»), no marcas inventadas como si fueran del campus. Pantalla en el alta y en el detalle, componente nuevo `panel/Insumos.tsx` para no reescribir el formulario entero: `Labores.tsx` solo importa. Consulta por actividad, por lugar y por mes.

Aceptación: test del consumo; curl de consulta por mes; Playwright añade un insumo a una actividad y lo ve en el detalle. No hay `DELETE` del catálogo, sí desactivar.

Prompt:

```text
Eres el agente cierre/o7-insumos de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o2-alta-actividad: cierre/o7-insumos. Sin main, sin force push, sin repo del grupo.

Lee RF-10. Catálogo de insumos y consumo ligado a la actividad, al lugar o área y al periodo. Importable, con historial, baja lógica. Componente panel/Insumos.tsx importado por el alta; no reescribas el formulario. Migraciones solo 080-082. OpenAPI en apps/api/openapi/o7-insumos.yaml. Datos de prueba genéricos, sin marcas reales ni nombres de persona.

Hecho cuando: curl lista el consumo de un mes y Playwright lo agrega a una actividad.
```

### `cierre/o7-adjunto-proveedor` — S — después de `o3-solicitudes`

Cierra: RF-14, HU-12 de adjunto, RF-14-CA1, RF-13-CA2.

Archivos: `panel/Solicitudes.tsx` solo el bloque de la orden (después de que solicitudes está integrado), reutilizando el `POST /evidencias` con `orden_id`. Sin migración si la columna ya existe. Sin portal de proveedor.

Aceptación: Playwright adjunta un PDF a la orden y lo vuelve a ver en esa orden. Curl multipart con `orden_id` 201.

Prompt:

```text
Eres el agente cierre/o7-adjunto-proveedor de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o3-solicitudes: cierre/o7-adjunto-proveedor. Sin main, sin force push, sin repo del grupo.

Lee RF-14. En la orden de servicio, adjunta y muestra el archivo del proveedor reutilizando POST /evidencias con orden_id. No crees rol proveedor ni migración si la columna ya existe. No edites el almacenamiento S3.

Hecho cuando: Playwright adjunta un PDF a la orden y lo lista ahí.
```

### `cierre/o7-metricas-proveedor` — L — después de `o5-indicadores` y de solicitudes

Cierra: RF-15, HU-14 de métricas, RF-15-CA1. Fórmula provisional: órdenes en estado conforme / órdenes del periodo, por empresa. Pantalla en Solicitudes o en Indicadores, con `provisional: true`. No es el acuerdo con el cliente.

Aceptación: test con dos órdenes, una conforme, 50 %. La UI lo dice provisional. Sin migración: el cálculo es en lectura.

Prompt:

```text
Eres el agente cierre/o7-metricas-proveedor de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o5-indicadores y o3-solicitudes: cierre/o7-metricas-proveedor. Sin main, sin force push, sin repo del grupo.

Lee RF-15. Métrica provisional: órdenes conformes / órdenes del periodo, por empresa, provisional true. No la presentes como fórmula acordada. Sin portal de proveedor. Sin migración: calcúlalo en lectura. No tomes un número de migración de otro flujo.

Hecho cuando: el test da 50 con dos órdenes y una conforme.
```

### `cierre/o7-vigencia-sector` — L — después de `o2-zonificacion` y `o2-alta-actividad`

Cierra: RF-34, HU-27/HU-33 de sugerencia, RF-34-CA1, RF-34-CA2.

Migraciones 083–085: `vigente_desde` y `vigente_hasta` en la asignación del polígono. No se borra la asignación vieja: se cierra el intervalo. Al clavar el pin, `ST_Covers` sugiere sector de capataz, cuadrilla y edificio referente. La persona confirma. Si el punto no cae en ningún polígono, no se inventa un sector.

Archivos nuevos: caso de uso de sugerencia, `panel/SugerenciaSector.tsx` importado por el alta con una línea.

Aceptación: test espacial con un polígono de prueba y un punto dentro (sugiere) y uno fuera (no sugiere). Playwright muestra la sugerencia antes de guardar.

Prompt:

```text
Eres el agente cierre/o7-vigencia-sector de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o2-zonificacion y o2-alta-actividad: cierre/o7-vigencia-sector. Sin main, sin force push, sin repo del grupo.

Lee RF-34. Vigencia por fechas en la asignación del polígono, sin borrar la asignación anterior. Al marcar el punto, sugiere sector de capataz, cuadrilla y edificio si ST_Covers lo encuentra; si no, no inventes. Componente panel/SugerenciaSector.tsx, una línea de import en el alta. Migraciones solo 083-085. OpenAPI en apps/api/openapi/o7-vigencia-sector.yaml.

Hecho cuando: un test espacial distingue punto dentro y punto fuera.
```

### `cierre/o7-capas` — L — después de `o2-zonificacion`

Cierra: RF-33, HU-33, RF-33-CA1, RF-33-CA2. Jardín de préstamo no se renombra a jardín de reserva (pregunta abierta). La capa sigue diciendo «Jardines de reserva».

Tareas: cargador guiado de GeoJSON o shapefile que entrega la sección (sectores, usos, jardines, inventario forestal, cuarteles), con vista previa y lote reversible. Sin archivo, el vacío lo dice. No se pisan las capas que el ETL ya cargó: el cargador hace upsert por `origen_ref`.

Migración 086 solo si falta una tabla de capa genérica. Si el importador actual ya cubre la entidad, no se duplica: se añade la entidad que falte (cuartel, inventario forestal) al caso de uso existente **en archivos nuevos** para no reescribir `importacion.group.go` entero. Si hay que tocar la lista de entidades, se añade al final.

Aceptación: test de importación de un GeoJSON mínimo de cuartel; la fila queda; revertir no hace `DELETE` físico si el flujo de auditoría ya cambió esa regla (si este flujo corre antes, no introduce un `DELETE`). Playwright: el selector de importar agrupa «Capas del campus» y nombra jardín de reserva, no jardín de préstamo.

Prompt:

```text
Eres el agente cierre/o7-capas de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o2-zonificacion y, si ya está, o4-auditoria: cierre/o7-capas. Sin main, sin force push, sin repo del grupo.

Lee RF-33. Cargador con vista previa de shapes o GeoJSON de sectores, usos, jardines, inventario forestal y cuarteles. Upsert por origen_ref. No llames «jardín de préstamo» al jardín de reserva. No inventes geometría. No TRUNCATE. Migración solo 086 si falta tabla. OpenAPI en apps/api/openapi/o7-capas.yaml.

Hecho cuando: un test importa un GeoJSON de un cuartel y la fila se puede revertir sin DELETE físico.
```

### `cierre/o7-paginacion` — M — después de `o3-mapa-filtros`

Cierra: RNF-07, RNF-07-CA1.

`GET /operacion/actividades` acepta `limit` y `offset` (tope máximo, por defecto el comportamiento actual para no romper la web). La lista de la UI pide páginas. Sin migración.

Aceptación: test de limit/offset; Playwright pasa de página.

Prompt:

```text
Eres el agente cierre/o7-paginacion de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o3-mapa-filtros: cierre/o7-paginacion. Sin main, sin force push, sin repo del grupo.

Lee RNF-07. Añade limit y offset a GET /areas-verdes/v1/operacion/actividades, con tope, sin cambiar el default que ya consume la web. La lista pagina. No edites eventos ni evidencias. Sin migración.

Hecho cuando: un test pide limit=1 y recibe una fila, y la UI avanza de página.
```

### `cierre/o7-refresco` — S — después de `o4-claridad`

Cierra: RF-17, HU-14 de avance, RF-17-CA1.

Al volver el foco o cada minuto corto, la oficina relee actividades abiertas. No es un canal en tiempo real. El checklist por jardín no se construye (la matriz lo deja fuera del texto del RF).

Aceptación: test del hook (foco dispara la recarga). Playwright no es obligatorio si el hook está cubierto.

Prompt:

```text
Eres el agente cierre/o7-refresco de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o4-claridad: cierre/o7-refresco. Sin main, sin force push, sin repo del grupo.

Lee RF-17. Al recuperar el foco, y con un intervalo corto, la oficina vuelve a leer las actividades abiertas. Sin websocket. Sin checklist por jardín. Sin migración. Un test del hook basta.

Hecho cuando: el test simula el foco y cuenta una lectura nueva.
```

### `cierre/o7-ia` — L — después de la ola 2

Cierra el mecanismo de RF-25 / RF-25-CA1 en lo que el libro permite cerrar. El caso (imagen, reporte asistido o lenguaje natural) **sigue pendiente**. Este flujo no llama a un servicio externo, no envía el título fuera de la máquina y no marca el RF como caso de valor elegido. Mantiene `requiere_humano: true` y la confirmación. Documenta en `docs/DECISIONES.md` los tres candidatos y que no hay elección. Si se añade algo, es una prueba más de la heurística local ya existente, no un modelo nuevo.

Aceptación: el test actual de IA sigue; uno nuevo asegura que el caso de uso no abre un cliente HTTP. La UI sigue exigiendo Confirmar.

Prompt:

```text
Eres el agente cierre/o7-ia de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con la ola 2: cierre/o7-ia. Sin main, sin force push, sin repo del grupo.

Lee RF-25 y RNF-13. No elijas tú el caso de IA: el libro no lo cerró. No llames a un servicio externo ni envíes el título. Conserva requiere_humano y el botón Confirmar. Anota los tres candidatos como pendientes en docs/DECISIONES.md. Un test debe fallar si el caso de uso importa un cliente HTTP. Sin migración.

Hecho cuando: ese test pasa y la sugerencia sigue siendo local.
```

## 12. Ola 8 — Could

Fuera del MVP. Van al final para no adelantarles permisos ni datos.

### `cierre/o8-portal-proveedor` — L — después de `o7-adjunto-proveedor` y de cuentas

Cierra: RF-27, HU-22, RF-27-CA1.

Rol nuevo con código `proveedor` (no se renombran los cuatro códigos actuales). Etiqueta visible: «Proveedor». Solo ve sus órdenes y puede dejar atención o reporte. No ve catastro, cuentas ni reportes de la jefatura. Cuenta ficticia de prueba. Migración 087: el rol y sus permisos, `ON CONFLICT DO NOTHING`.

Aceptación: curl con esa cuenta lee su orden y recibe 403 en `/accesos/usuarios` y en catastro. Playwright entra y no ve Administración.

Prompt:

```text
Eres el agente cierre/o8-portal-proveedor de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o7-adjunto-proveedor y o1-cuentas: cierre/o8-portal-proveedor. Sin main, sin force push, sin repo del grupo.

Lee RF-27. Rol código proveedor, etiqueta Proveedor, sin renombrar capataz, coordinacion, jefatura ni admin. Solo sus órdenes y el reporte. Migración solo 087. Cuenta de prueba ficticia. OpenAPI en apps/api/openapi/o8-portal-proveedor.yaml.

Hecho cuando: curl de ese rol lee su orden y recibe 403 en cuentas y en catastro.
```

### `cierre/o8-concentracion` — L — después de `o3-mapa-filtros` y de `o7-vigencia-sector`

Cierra: RF-35, HU-34, RF-35-CA1.

Conteo de actividades por sector de capataz y por jardín de reserva, en un periodo, **incluidas las cerradas**. Capa o lista, no un mapa de calor decorativo sin números. Migraciones 088–089 solo si hace falta una vista materializada; si un `GROUP BY` alcanza, no se crean. No se borra el histórico para «acelerar».

Aceptación: test con dos actividades cerradas en un sector y cero en otro. Curl del periodo las cuenta. Playwright muestra el conteo.

Prompt:

```text
Eres el agente cierre/o8-concentracion de VerdePUCP (Irico17/Areas-verdes-pucp). Rama desde develop con o3-mapa-filtros y o7-vigencia-sector: cierre/o8-concentracion. Sin main, sin force push, sin repo del grupo.

Lee RF-35. Cuenta actividades por sector de capataz y por jardín de reserva, cerradas incluidas, en un periodo. Muestra el número. No borres histórico. Migraciones 088-089 solo si hacen falta; si no, no las crees. OpenAPI en apps/api/openapi/o8-concentracion.yaml.

Hecho cuando: un test cuenta dos actividades cerradas en un sector y cero en el otro.
```

## 13. Orden de integración a `develop`

1. `o0-desacople`
2. En paralelo `o1-cuentas`, `o1-catalogos-estados`, `o1-reserva-patch`. Luego `o1-nomenclatura`. Luego `o1-integracion`.
3. En paralelo `o2-zonificacion` y `o2-ejemplares`. Luego `o2-alta-actividad`. Luego `o2-integracion`.
4. En paralelo los seis de la ola 3. Merge en el orden del integrador 3.
5. `o4-auditoria`, `o4-claridad`, `o4-e2e`.
6. `o5-reportes`, `o5-indicadores`.
7. `o6-duckdns-https` cuando haya PEM. Puede haberse desarrollado desde el paso 1.
8. Ola 7 en el orden de sus dependencias. Paralelo posible: insumos, paginación, refresco, IA, capas (si zonificación ya está).
9. Ola 8.

Después de cada integrador: `go test ./...` en `backend/app`, `npm test` en `apps/web`, y el humo de `docs/AMBIENTES.md` (`make smoke ENV=develop`) si el entorno lo permite. Qa recibe un `rc-*`. Producción espera la aprobación. Ningún paso hace force push ni toca `main`.

## 14. Cobertura

| ID | Flujo que lo cierra |
|---|---|
| RF-01, RF-02, RF-03, DEC-01, DEC-02, DEC-06 | `o1-cuentas` |
| RF-16, RF-24, RNF-05 (catálogos) | `o1-catalogos-estados` |
| RNF-05 (permisos) | `o1-cuentas` |
| DEC-11 | `o1-reserva-patch` |
| Etiquetas del glosario salvo roles, SSO, estados y claridad | `o1-nomenclatura` |
| RF-06 | `o2-zonificacion` |
| RF-04, RF-05 | `o2-ejemplares` |
| RF-08 (sin insumos), RF-30, DEC-07 | `o2-alta-actividad` |
| RF-12, RNF-01 | `o3-offline` |
| RF-31, RF-19-CA3 | `o3-trazabilidad` |
| RF-29, RF-32, RNF-11 | `o3-mapa-filtros` |
| RF-19 almacén, DEC-10 | `o3-evidencias-s3` |
| RF-11, RF-13 | `o3-solicitudes` |
| RF-26 | `o3-riego` |
| RNF-14, DEC-04, DEC-05, RF-18 | `o4-auditoria` |
| DEC-08, RNF-06 | `o4-claridad` |
| RNF-10 | `o4-e2e` |
| RF-20, RF-22 | `o5-reportes` |
| RF-21, RF-23 | `o5-indicadores` |
| DEC-09, RNF-03, RNF-04 (tránsito) | `o6-duckdns-https` |
| RNF-04 (claves y texto) | `o1-cuentas` |
| RF-10 | `o7-insumos` |
| RF-14 | `o7-adjunto-proveedor` |
| RF-15 | `o7-metricas-proveedor` |
| RF-34 | `o7-vigencia-sector` |
| RF-33 | `o7-capas` |
| RNF-07 | `o7-paginacion` |
| RF-17 | `o7-refresco` |
| RF-25 | `o7-ia` (mecanismo local; el caso queda abierto) |
| RF-27 | `o8-portal-proveedor` |
| RF-35 | `o8-concentracion` |

Ya cumplidos, no se reabren: RF-07, RF-09, RF-28, RNF-02, RNF-08, RNF-09, RNF-12, RNF-13, DEC-03, y los CA marcados Cumplido en la matriz.

## 15. Lo que ningún flujo debe inventar

Queda escrito para no convertirlo en código «porque faltaba»:

1. Matriz de permisos definitiva del cliente (RF-03). Se entrega editable, con la semilla actual.
2. Estructura real del Excel del cliente (RF-07). El importador ya existe.
3. Campos específicos por tipo en JSONB (RF-08, observación).
4. Si «Bloqueada» se conserva (RF-16).
5. Checklist por jardín (RF-17, observación).
6. Columnas oficiales de reporte intermedio y avanzado (RF-20, RF-21). Hay propuesta provisional.
7. Fórmulas oficiales de cobertura, riego, rendimiento y proveedor (RF-15, RF-23, RF-26). Hay fórmula provisional etiquetada.
8. Motivos de cancelación: la app ya tiene cuatro; el cliente no los ha confirmado. No se borran ni se multiplican.
9. Caso de IA (RF-25).
10. Si el riesgo admite «medio» (RF-30).
11. Si jardín de préstamo es jardín de reserva (RF-33).
12. Shapes que la sección aún no entrega (RF-33, RF-34).
13. Restricciones institucionales de la PUCP y el dominio definitivo frente a DuckDNS (RNF-03).
14. Plazo de retención (RNF-14). No hay borrado por fecha.
15. Ortofoto. No está en `data/raw`.
