# Glosario de nomenclatura — VerdePUCP

Fecha: 2026-10-02. Rama: `docs/plan-cierre-backlog`.

Este glosario fija el vocabulario visible de la interfaz y de la API de cara a la persona. Los códigos internos (slug de rol, slug de estado, nombre de tabla, `feature_id`) no se renombran. Lo que cambia es la etiqueta.

Fuentes, en este orden cuando dicen lo mismo:

1. Textos y propiedades de la web institucional copiada en `data/raw/` (GeoJSON, CSV y `legacy-app/index.html`).
2. Product backlog v2 (`docs/fuente/extraccion/backlog_v2.txt` y `docs/MATRIZ-BACKLOG-V2.md`).
3. Modelo ya cargado por el ETL (`docs/MAPA-DATOS-Y-EDICION.md`, `backend/app/internal/infrastructure/etl`).

Cuando la web institucional y el backlog usan dos palabras para la misma cosa, abajo queda **una** etiqueta oficial. La otra pasa a sinónimo prohibido en pantallas nuevas, para no repetir el choque actual entre «zona», «sector» y «equipo».

Los nombres de persona de la web institucional no se muestran. En su lugar va el ficticio estable del ETL (DEC-03).

## Cómo usarlo

- Pantalla, mensaje de error, leyenda, placeholder y campo JSON que lee una persona: término oficial.
- Código de rol, código de estado, nombre de archivo y nombre de tabla: se quedan.
- Importación: se acepta la cabecera de la fuente (`Uso`, `Proy riego`, `Riego act`, `código`, `Nombre`) y se muestra con la etiqueta oficial.

## Términos oficiales

### Territorio

| Término oficial | Qué es | De dónde sale | Sinónimos prohibidos en la UI | Dónde está hoy |
|---|---|---|---|---|
| Área verde | Polígono del catastro. Propiedades de fuente: `Nombre`, `código`, `Uso`, `Proy riego`, `Riego act`, `Referenc_1`, `Perimetro`, `Área`. | `data/raw/areas_verdes.geojson` (521). Leyenda institucional: «Áreas Verdes». | Zona (para este polígono), feature, parcela | Capa «Áreas verdes» en `apps/web/src/types.ts` (`LAYERS`, id `areas`). Correcto. |
| Uso | Clasificación del área. Valores de la fuente, tal cual: «Uso Institucional» (1), «Áreas de uso administrativo» (351), «Áreas de uso recreativo/descanso» (70), «Áreas de manejo sostenible y reducción de consumo de agua» (96), «Áreas deportivas y recreación activa» (2), «Áreas de conservación» (1). | Propiedad `Uso` del GeoJSON. La leyenda del visor original acorta algunas («Áreas de Uso Administrativo», «Áreas Recreativas / Descanso», «Manejo Sostenible y Ahorro Agua», «Otros Usos»). | Categoría, tipo de área, rubro | `apps/web/src/map/categorias.ts`: «Institucional», «Administrativo», «Recreativo / descanso», «Manejo sostenible», «Otros». Hay que mostrar el texto de la fuente, no el acortado. |
| Proyecto de riego | Intención de riego del área. Valores de fuente en `Proy riego`: «Por validar con unidad», «Goteo», «Falta aspersión», «Cuenta con aspersión». | Misma capa. | «Proy», proyecto a secas | El editor de catastro no rotula «Proyecto de riego» con esos cuatro valores. |
| Riego actual | Riego que hay hoy. Valores de fuente en `Riego act`: «Sin riego tecnificado», «Riego por aspersión», «Riego por goteo». La leyenda original también nombra «Cisterna / Camión» y «Riego Manual»; no están en el GeoJSON y no se inventan. | Propiedad `Riego act`. | «riego_act» visible, «sistema» | Campo técnico `riego_act` en `apps/web/src/producto.ts` (`Ficha`). |
| Referencia | Texto `Referenc_1` (edificio o lugar cercano). Ejemplo en fuente: «Exterior de Facultad de Artes Escénicas» en Jardín Rosales. | GeoJSON. | Dirección, comentario de mapa | Ficha de catastro, campo referencia. El nombre de columna no se muestra. |
| Zona de supervisión | Una de cuatro zonas del campus. En la fuente la propiedad `ZONA` vale `Zona1` … `Zona4`. En pantalla: «Zona 1» … «Zona 4», códigos `Z1`–`Z4`. Sirve para filtrar tachos, como el visor original. | `data/raw/lote/supervisoress.geojson`. Backlog: zona. | Sector, cuadrilla, polígono de cuadrilla | `CatastroEditor` y el riego dicen «zona de supervisión». El resumen del mapa dice «534 zonas» para otra cosa (ver polígono de cuadrilla). |
| Sector de capataz | Polígono operativo de una cuadrilla. En el visor original el grupo se llama «Zonas de Áreas Verdes» y cada casilla «Sectores de responsable», más «Campo deportivo» y «Bosque húmedo». El backlog v2 lo llama sector de capataz (RF-06, RF-34). Esa es la etiqueta de pantalla. | `jefe_de_grupo.json` (534) + columna `poligonos_cuadrilla.sector`. Rótulos de lugar que se conservan: «Campo deportivo», «Bosque húmedo». | Zona, zonas, jefe de grupo, sector operativo, equipo | Capa «Zonas» con pista «Sectores operativos, sin nombres de personas» (`types.ts`). Leyenda «Sector operativo» en el mapa. Códigos internos `cua-valeria`, `cua-mateo`, `cua-renato` no se muestran. |
| Polígono de cuadrilla | Nombre técnico de esa misma geometría en base, ETL e importación. No es una pestaña. | Tabla `poligonos_cuadrilla`. El ETL escribe `feature_id` `PC-`. | Llamarlo «zona» en un formulario | Importación: «Polígonos de cuadrilla» (`panel/importaciones.ts`). Correcto en esa pantalla técnica. La capa del mapa no debe usar este nombre. |
| Cuadrilla | Equipo de campo, con nombre ficticio. Sustituye al campo `jefes` de la fuente, que no se guarda. Las de demostración se dicen «Cuadrilla Norte», «Cuadrilla Sur» y «Cuadrilla Riego», no «Equipo Norte». | ETL y `docs/MAPA-DATOS-Y-EDICION.md` §4.4. | Equipo, jefe, responsable (como si fuera el nombre real) | Filtro y alta dicen «Equipo» (`panel/Labores.tsx`). Reportes dicen «Cuadrilla» (`panel/Modulos.tsx`). Hay que dejar una sola palabra: cuadrilla. |
| Lugar | Punto con nombre del diccionario (`data/raw/sheets/lugares.csv`, columnas `lugar`, `latitud`, `longitud`). | 76 filas. Backlog RF-06. | Lugar libre, sitio, texto suelto | La ficha de la actividad tiene un input «Lugar» de texto libre (`Labores.tsx`). Pasa a selector del catálogo. Lo ya guardado en `lugar_libre` se muestra como dato histórico, no como campo editable. |
| Jardín | Área verde con nombre de jardín en la fuente (ejemplo: «Jardín Rosales», código `B 10`). | Propiedad `Nombre` cuando trae «Jardín …». | Macizo, cantero | No hay etiqueta distinta de área verde. No hace falta una pestaña «Jardín» para el polígono general. |
| Jardín de reserva | Capa de 21 polígonos. Propiedad extra `Pertenecen` (ejemplo `DAF`). | `data/raw/jardines_reserva.geojson`. | Jardín de préstamo. El backlog lo deja sin confirmar (RF-33). Hasta que el cliente lo diga, no se renombra esta capa. | «Jardines de reserva» en `types.ts` y en `panel/inventarioCapas.ts`. Se mantiene. |
| Cuartel | Referencia histórica forestal, solo lectura, cuando exista el shape. No es un sector de capataz. | Backlog RF-06 y RF-33. No está en `data/raw`. | Sector, zona | No aparece en la UI. Al cargarlo, la capa se llama «Cuarteles (histórico)». |
| Vía | Referente lineal. Edificios ya existen como capa. | Backlog RF-06. El visor no trae un GeoJSON de vías. | Calle a secas si se mezcla con lugar | No hay capa. La nueva se llama «Vías». La de edificios se llama «Edificios», no «Huellas del recinto». |
| Ejemplar | Un individuo del inventario de flora (árbol, palmera, arbusto, herbácea, trepadora, suculenta), con `N°`, código, nombre común y nombre científico. | Hoja de flora descrita en `docs/MAPA-DATOS-Y-EDICION.md` §4.6. | Árbol (como si todo fuera árbol), item, feature | No hay pestaña. La capa se llama «Flora». Al editar la ficha, el título es «Ejemplar». |
| Especie | Nombre científico (y común) que agrupa ejemplares. | Misma hoja. | Tipo de planta | Catálogo `especie`. Correcto como clase de catálogo. |

### Trabajo de campo

| Término oficial | Qué es | De dónde sale | Sinónimos prohibidos en la UI | Dónde está hoy |
|---|---|---|---|---|
| Actividad | Registro operativo. La hoja de monitoreo tiene la columna `Actividad`. El backlog titula la épica «Registro de actividades e intervenciones». La API es `GET/POST /areas-verdes/v1/operacion/actividades`. | `data/raw/sheets/monitoreo_2026.csv`, `actividades.csv`, backlog v2. | Labor, labores (como título de pantalla o de importación) | Pestaña y `<h2>` «Labores» (`App.tsx`, `Labores.tsx`). Importación etiqueta «Labores» (`importaciones.ts`). El archivo puede seguir llamándose `Labores.tsx`; la etiqueta visible no. |
| Intervención | La misma actividad, cuando el texto del backlog habla del trabajo ya clasificado. No es una segunda entidad ni una segunda pestaña. | RF-08, RF-16, RF-19. | Un módulo aparte llamado Intervenciones | No hay pestaña. No se crea. |
| Clase de actividad | Primer nivel. Las nueve de la hoja (`actividades.csv`, bloque derecho): Habilitación de jardines, Rehabilitación de jardines, Mantenimiento de jardines, Poda, Propagación y plantación, Riego, Manejo fitosanitario, Manejo de residuos vegetales, Inspección y monitoreo. | `data/raw/sheets/actividades.csv`. RF-08 pide esas clases. La base tiene siete y le faltan manejo fitosanitario e inspección y monitoreo. | Eje, tipo (para este nivel), categoría | El alta pide un solo nivel «Tipo» (`Labores.tsx`). `GET /catalogos` no publica `clase_actividad`. |
| Tipo de actividad | Segundo nivel, hijo de la clase. La hoja trae 45 (Preparación del terreno, Poda de mantenimiento, Riego manual, …). | Misma hoja, columna «Tipo de actividad». | Subtipo suelto sin clase, código crudo | `TIPOS` fijos en `operacion.ts`. |
| Personal de la actividad | Nombres ficticios de quien trabajó, sin cuenta. | RF-08. `personal_labor.nombre_ficticio`. | Operario como rol con login, DNI, nombre real | No hay campo en el alta. |
| Insumo | Material consumido por la actividad, el espacio y el periodo. | RF-10. No está en la web institucional. | Producto (a secas; «producto fitosanitario» sí es catálogo, RF-24) | No hay pantalla. |
| Solicitud | Pedido con fuente, código externo, prioridad, lugar, cantidad solicitada y cantidad ejecutada. | RF-11. | Incidencia como entidad distinta, ticket | Pestaña «Solicitudes». El lugar es texto, no un punto. |
| Código externo | Código de Centuria u OSG. El sistema no lo inventa. | RF-28, RF-30. | Código interno, id | Campo en solicitudes. Falta en el alta de la actividad. |
| Orden de servicio | Contratación tercerizada: empresa, referencia, frecuencia, conformidad. | RF-13. | Orden a secas si se confunde con una actividad | Formulario dentro de Solicitudes (`Modulos.tsx`). |
| Servicio tercerizado | Ejecutor `tercerizada`. La etiqueta visible es «Servicio tercerizado». El código sigue siendo `tercerizada`. | RF-09. | Proveedor como rol del MVP, outsourcing | Selector «Quién ejecuta» en `Labores.tsx`. |
| Riego | Registro por sector de capataz, turno y cuadrilla, más avance y cobertura del ciclo. | RF-26. Clase de actividad «Riego». | Un «sector» de texto libre distinto del sector de capataz | `RiegoPanel` en `Modulos.tsx`: campo sector en texto («Eje central») y otro campo de zona. |
| Poda | Registro de poda sobre un ejemplar. | Hoja y tabla `podas`. | Actividad de tipo poda duplicada sin vínculo | Panel al final de Labores. |
| Vivero | Registro de propagación. | Hoja «Propagación y plantación». | Flora como título de ese formulario | Panel al final de Labores. |
| Evidencia | Foto, PDF o coordenada colgada de una actividad, una solicitud, una orden o un evento. | RF-19. | Adjunto, archivo (como título) | `EvidenciasCampo.tsx`, solo a la actividad. |
| Reserva de jardín | Turno ficticio sobre un jardín de reserva (`hora_inicio`, `hora_fin`). La hoja real respondió 401. | `docs/MAPA-DATOS-Y-EDICION.md` §4.14. DEC-11. | Agenda institucional, reserva real | `CalendarioReservas.tsx` e inventario. |
| Estado de la actividad | Etiquetas del backlog: Por iniciar, En proceso, Ejecutado, Cerrado, Cancelado, Archivado. Los códigos (`pendiente`, `en_proceso`, …) no se renombran. «Bloqueada» sigue en datos hasta que el cliente decida; no se ofrece como nombre nuevo. | RF-16. | Mostrar el slug (`ejecutado`) | `operacion.ts` usa Pendiente, En proceso, Bloqueada, Cerrada, Cancelada. |

### Inventario del visor

Etiquetas tomadas del panel del visor original (`data/raw/legacy-app/index.html`) y de los archivos de `data/raw/`.

| Término oficial | Sinónimo prohibido | Dónde está hoy |
|---|---|---|
| Tachos | Contenedores, basureros | Inventario. Los conteos se muestran con el nombre de columna (`no_aprovechables`). Pasan a: No aprovechables, Papel y cartón, Plástico, Vidrio, Pilas, Peligrosos, RAEE, Metales, Aniquem, Intermedios plástico, Intermedios metal. |
| Bebederos | Fuentes (como título de la capa). El tipo sí se dice «Tipo fuente» o «Tipo llenador de botella», como el archivo. | Inventario. |
| Fauna | Animales | `inventarioCapas.ts`. Correcto. |
| Puertas y entradas | Accesos | El archivo es `puertas_entradas.geojson`. La UI dice «Puertas». Pasa a «Puertas y entradas». |
| Playas de estacionamiento | Playas, estacionamientos | La UI dice «Playas». |
| Vereda en riesgo | Vereda | Se mantiene «Vereda en riesgo». |
| Xerofítica | Jardín seco | Se mantiene. |
| Puntos del campus | Puntos PUCP, POI | Capa de `puntos_pucp.csv`. Título visible: «Puntos del campus». |
| Edificios | Huellas del recinto, OSM (como título) | Pista actual «Huellas del recinto». |

### Roles

Etiquetas ya decididas. Los códigos no cambian: `capataz`, `coordinacion`, `jefatura`, `admin`.

| Código | Etiqueta oficial | Hoy en base (`045_roles_v2_etiquetas.sql`) y en `types.ts` |
|---|---|---|
| `capataz` | Capataz | Capataz. Correcto. |
| `coordinacion` | Ingeniería / Coordinación | Ingeniería/Coordinación (sin espacios alrededor de la barra). |
| `jefatura` | Jefatura de sección | Jefatura / Jefe de sección. |
| `admin` | Administrador del sistema | Administrador del sistema. Correcto en el chip. La pestaña dice «Admin». |

Prohibido en pantalla: Supervisor (como rol), Jefe de sección (suelto), Operario (como cuenta), SSO, «pendiente de la universidad».

«Como supervisor» en las historias del libro es la voz de la historia, no un quinto rol. En la interfaz esa persona es Ingeniería / Coordinación o Jefatura de sección, según el permiso.

### Palabras que se quedan en código

| Se ve en código | No se muestra así |
|---|---|
| `labor`, `Labores.tsx`, cola offline `QueuedLabor` | Actividad |
| `equipo`, `formEquipo`, `readEquipo` | Cuadrilla |
| `zonas` como id de capa (`LAYERS` id `zonas`) | Sector de capataz. El id de capa puede quedar para no romper el cliente; la etiqueta no. |
| `ejecutor=propia\|tercerizada` | Personal propio / Servicio tercerizado |
| `lugar_libre` | Solo lectura, «Lugar (dato ya cargado)» |

## Renombres necesarios en el frontend

Solo etiquetas. Ningún renombre de ruta, de código de rol ni de id de capa.

| # | Dónde | Texto de hoy | Texto oficial | Flujo que lo hace |
|---|---|---|---|---|
| 1 | `App.tsx` pestaña `labores` | Labores | Actividades | `cierre/o1-nomenclatura` |
| 2 | `panel/Labores.tsx` título | Labores | Actividades | `cierre/o1-nomenclatura` |
| 3 | `panel/importaciones.ts` entidad `labores` | Labores | Actividades | `cierre/o1-nomenclatura` |
| 4 | `panel/Labores.tsx` filtros y alta | Equipo | Cuadrilla | `cierre/o1-nomenclatura` |
| 5 | `panel/Modulos.tsx` (`RiegoPanel`) | Sector (texto libre) y «Equipo» si aparece | Sector de capataz y Cuadrilla | `cierre/o1-nomenclatura` el rótulo; `cierre/o3-riego` el selector |
| 6 | `types.ts` `LAYERS` id `zonas` | Zonas / «Sectores operativos…» | Sectores de capataz / «Polígono de la cuadrilla, sin nombres reales» | `cierre/o1-nomenclatura` |
| 7 | `map/categorias.ts` `USOS` | Institucional, Administrativo, Recreativo / descanso, Manejo sostenible, Otros | Los seis textos de la propiedad `Uso` (los dos minoritarios no se funden en «Otros» sin poder verlos) | `cierre/o1-nomenclatura` |
| 8 | `map/categorias.ts` `SECTORES` | Cuadrilla Valeria Quispe, … | Sector de capataz — Valeria Quispe (ficticio), y igual con Mateo Salazar y Renato Cárdenas. Campo deportivo y Bosque húmedo se quedan. | `cierre/o1-nomenclatura` |
| 9 | `types.ts` `ROLES` y chip | Ingeniería/Coordinación; Jefatura / Jefe de sección | Ingeniería / Coordinación; Jefatura de sección | `cierre/o1-cuentas` (también el `UPDATE` de `roles.nombre`) |
| 10 | `App.tsx` pestaña `admin` | Admin | Administración | `cierre/o4-claridad` |
| 11 | `panel/Modulos.tsx` login, línea del SSO | «el SSO de la universidad no está conectado» | «Cuentas de la jefatura. No hay SSO.» | `cierre/o1-cuentas` |
| 12 | `panel/Modulos.tsx` `AdminPanel` | «El SSO institucional no forma parte de este piloto.» | «La jefatura de sección administra las cuentas.» | `cierre/o1-cuentas` |
| 13 | `panel/Labores.tsx` ficha, input Lugar | Texto libre | Lugar (catálogo). El valor viejo, si no tiene `lugar_id`, se lee y no se edita. | `cierre/o2-zonificacion` |
| 14 | `operacion.ts` `ESTADOS` | Pendiente, Bloqueada, Cerrada, Cancelada, slug crudo | Por iniciar, En proceso, Ejecutado, Cerrado, Cancelado, Archivado | `cierre/o1-catalogos-estados` |
| 15 | `panel/inventarioCapas.ts` | Playas; Puertas; conteos en snake | Playas de estacionamiento; Puertas y entradas; etiquetas del visor | `cierre/o1-nomenclatura` |
| 16 | Capa de edificios | Huellas del recinto | Edificios | `cierre/o1-nomenclatura` |
| 17 | `producto.ts` `HUECOS` y reportes | «Jefatura no ha acordado la fórmula» | Se mantiene la honestidad, con el rol «Jefatura de sección» | `cierre/o5-reportes` al reescribir la pantalla |

Lo que no se renombra: «Jardines de reserva», «Xerofítica», «Fauna», «Tachos», «Bebederos», «Capataz», «Administrador del sistema», «Catastro», «Solicitudes», «Reportes», «Catálogos», «Importar». «Catastro» e «Importar» los vuelve a mirar el flujo de claridad (`cierre/o4-claridad`) como nombres de tarea, sin cambiar el significado de estos términos.
