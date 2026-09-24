# Product

<!-- impeccable:product-schema 1 -->

Supuestos de esta sesión, porque el encargo prohibió preguntas y pidió decidir: marcados con «(supuesto)». El resto sale del repositorio y del encargo.

## Platform

web

## Stack

Monolito ya existente: API Go (Gin, GORM) y visor React + Vite + MapLibre. Postgres + PostGIS. No se cambia de framework. (El encargo fija el stack; no es una decisión abierta.)

## Users

- Capataz de campo (cuentas norte, sur, riego). Está en el campus, a menudo con el teléfono, y registra o consulta las labores de su equipo. No ve el trabajo de otros equipos.
- Coordinación de áreas verdes. En oficina, sobre el mapa, arma labores, catastro, solicitudes y catálogos de consulta.
- Jefatura. Revisa, valida y exporta. No registra labores ni administra catálogos.
- Administración local. Cuentas y permisos de esta instalación. El SSO de la PUCP no está conectado.

(supuesto) El trabajo real ocurre de día, entre oficina y campo, con luz ambiente variable y prisa. La interfaz tiene que escanearse, no contemplarse.

## Product Purpose

Campus Verde es la mesa de guardia de las áreas verdes de PUCP Pando: ver el campus, saber qué labor está abierta, de quién es, y dejar constancia. El éxito es que un capataz encuentre sus labores y que coordinación cierre el día sin una hoja de cálculo paralela.

## Positioning

El mapa es el expediente, no un adorno. Cada labor, área y solicitud se apoya en el catastro de Pando (PostGIS). Un tablero genérico de tickets no puede decir eso sin mentir.

## Operating Context

Módulos: Mapa (plano y relieve por extrusión), Labores, Catastro, Solicitudes y órdenes, Riego por sector y turno, Reportes, Catálogos, Admin. Sesión por cookie `cv_sesion`. Cola offline del capataz en el navegador. Evidencias como archivo (disco local; S3 cuando hay cubo). Fechas de Lima, calendario es-PE. Idioma de la interfaz: español del Perú.

## Capabilities and Constraints

- Roles y permisos ya sembrados. El capataz consulta y registra; coordinación además valida, solicita y reporta; jefatura no registra; admin tiene catálogos y cuentas.
- Una labor tercerizada no se cierra sin orden de servicio.
- La sugerencia de tipo es una regla local sobre el título. No es un modelo externo y exige confirmación humana.
- Los conteos del reporte no son el indicador oficial de cobertura. Esa definición sigue pendiente.
- No hay SSO real, ni Centuria, ni ortofoto, ni portal de proveedores.
- MapLibre sigue siendo el mapa. El relieve es `fill-extrusion`, no una maqueta.
- Credencial local de desarrollo: `pando-local`. No es un secreto de producción.
- (supuesto) Las 501 áreas sin nombre en el catastro fuente no se inventan. Se muestran con su código (`AV-0004`, `B 4`) y su uso, y las fichas con nombre real van primero.

## Brand Commitments

Nombre: VerdePUCP. Lugar: PUCP Pando. Voz: operativa, concreta, en español del Perú, sin eslóganes. El encargo rechaza el aspecto genérico de “IA” y el estado actual, descrito como feo: formulario de acceso pegado arriba a la izquierda, fechas en mm/dd/yyyy, lista de catastro dominada por “Sin nombre”.

## Evidence on Hand

- Código en `apps/web` y `apps/api`. Plan previo en `docs/PLAN-PRODUCTO-Y-UI.md`.
- Catastro real: 521 áreas, 534 zonas. Unas 20 áreas tienen nombre de jardín (Jardín Tinkuy, Jardín Humanidades, y otras). El resto trae código y uso, no nombre.
- Labores de demostración en la migración 003 y la solicitud `OSG-2026-0142`.
- No hay fotos de equipo, testimonios ni manual de marca PUCP en el repo. No se fabrican.

## Product Principles

- El mapa manda; el panel responde.
- Cada rol ve solo el trabajo que le corresponde. El filtro vive en el servidor.
- La copia dice el límite del sistema (SSO pendiente, cobertura no definida, regla local de tipo).
- Densidad de oficina, no de afiche. Lo que se repite se alinea.
- Lo que no está en el catastro no se inventa: se nombra por su código.

## Accessibility & Inclusion

Interfaz de trabajo de día completo: contraste de texto AA, foco visible, objetivos táctiles de al menos 44 px en el teléfono, y movimiento que se retira con `prefers-reduced-motion` sin esconder el cambio de estado. (supuesto) No hay un requisito formal WCAG más allá de eso.
