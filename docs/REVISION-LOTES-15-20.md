# Revisión de los lotes 15–20

Fecha: 2026-10-02. Alcance: `git diff d35c28f..HEAD -- backend` (lotes 15 evidencias, 16 reportes/IA, 17 auditoría, 18–19 ETL, 20 swagger) más el arreglo parcial `cdeafd4`.

Se comparó con `apps/api` (códigos HTTP, textos en español, JSON, `ORDER BY`, permisos por rol y escrituras en `cambios`). La base de datos de las migraciones `001`–`046` sigue siendo la única fuente de verdad.

## Hallazgos y corrección

| # | Severidad | Hallazgo | Corrección |
|---|---|---|---|
| 1 | Alta | `POST /importaciones`: un fallo al insertar `lotes_importacion` que no traía `SQLSTATE` ni `pq:` (por ejemplo `dial tcp`) respondía **400** con el texto del driver. La API anterior responde siempre **500** `no se pudo guardar la vista previa`. | El caso de uso envuelve el error con `ErrGuardarVistaPrevia` y el controller responde 500 con ese texto. |
| 2 | Media | `VistaPreviaResponseDTO` tenía `omitempty` en `columnas_omitidas`, `aviso_omitidas` y `avisos`. La API anterior arma un `gin.H` y esas claves siempre van (lista nula o cadena vacía). | Se quitó `omitempty` en esas tres claves del DTO de respuesta. |
| 3 | Baja | La subida de evidencia recortaba `assigned_capataz_id` antes de compararlo con el capataz de la sesión. `evidencias/subir.go` de la API anterior no recorta el valor de la base: un id con espacios podía pasar aquí y recibir 403 allí. | La comparación vuelve a ser `asignado != strings.TrimSpace(in.CapatazID)`. |
| 4 | Media | `TestSwaggerRouteCoverage` exigía `>= 62` operaciones. Ese piso no se mueve si se borran rutas (el archivo ya documenta más de cien) y no describe el router. | El test exige igualdad de conjuntos entre las rutas Gin de `/areas-verdes/v1` (salvo `/swagger`) y los verbos HTTP de `swagger.json`. |
| 5 | Baja | Código muerto: mappers de evidencia y de lote que nadie llamaba; entidad `LoteImportacion` solo usada por ese mapper; campo `GuardarEvidencia.Ruta` sin lecturas ni escrituras; `RevertirLoteDTO` (el cuerpo es `requests.RevertirLoteRequest`); `ErrLoteYaRevertido` (la reversión sigue en `InputError` con el texto `el lote ya fue revertido`, que el controller mapea a 400); `iaController.logger` guardado y nunca leído; `if rep == nil` en `Revertir` (el repositorio siempre devuelve un reporte); `if u.lector != nil` en `Entidades` (el contenedor siempre inyecta el lector y, si faltara, devolver `null` ocultaba el fallo). | Eliminados. Los modelos GORM `EvidenciaModel` y `LoteImportacionModel` se quedan: los usa el registro de `modelgen`. El listado de evidencias sigue leyendo por SQL las columnas que devuelve la API anterior; no pasa por el mapper de fila completa. |

Los puntos 1–4 quedaron en `fix(backend): paridad de importaciones y conteo real de swagger`. El punto 5, en el commit de esta revisión.

`cdeafd4` ya había hecho: sufijo aleatorio de la BD de prueba, cobertura de swagger en las dos direcciones, comentario de la lista cerrada en `load.go` y comentario del `DELETE` de `medidas_palmera` al revertir (igual que la API anterior).

## Revisado y se conserva

- **Capas.** `persistence` no importa `application/dto` ni `presentation`. `application` no importa `persistence`, `infrastructure`, `presentation` ni `gorm`. Lo cubre `TestArchitectureLayers`.
- **SQL.** Los nombres de tabla o columna interpolados salen de mapas o `switch` cerrados (`bajaPorActivo`, `tablaPermitida`, `TablasETL`, `TablasGuardLoad`, `upsertCapaPunto`). Filtros de reportes, evidencias y auditoría van en placeholders. `ORDER BY` de reportes y de evidencias está fijo.
- **Archivos.** `disco.storage` usa `filepath.Base` al guardar y, al abrir, exige que la ruta limpia siga bajo el directorio. El id de la evidencia es parámetro SQL; el nombre subido no entra en `Content-Disposition` (`inline`) ni en la ruta. La importación no extrae el zip al disco con el nombre del cliente.
- **Permisos.** Evidencias (`consultar` / `registrar` o `evidencias`), reportes (`reportes`), IA (`consultar`), auditoría y lotes (`validar` / `consultar`) e importaciones (`consultar` / `validar`) coinciden con `exige` / `PermiteAlguno` de `apps/api`. Hay tests de 401 y de rol en `routes_test.go`.
- **Evidencias.** Tope 8 MB, MIME por contenido, hash distinto del pedido en 400, mismo id con otro hash en 409, EXIF filtrado a `fecha`/`lat`/`lon`/`orientacion`. Cabeceras de descarga: `Content-Type` de la fila, `Cache-Control: private, max-age=86400`, `nosniff`, `inline`.
- **Reportes e IA.** Mismos errores de fecha, mismos nombres JSON, CSV y XLS con el mismo `Content-Disposition`, mismas cinco reglas de sugerencia.
- **Auditoría.** Mismos 400/404/409/500, incluido 409 `hay filas editadas después del lote`. `cambios` solo donde la API anterior los escribe. El `DELETE` de `medidas_palmera` al revertir es paridad, no un borrado de datos cargados por migración.
- **`IAuditoriaService.RegistrarCambio`.** Está registrado y cubierto por test (omite ediciones sin cambio). Las rutas HTTP escriben `cambios` en SQL del repositorio, igual que la API anterior. No se elimina: es el servicio de la arquitectura y no cambia respuestas.
- **Índice estático `RutasV1`.** Sigue incompleto (no lista auditoría, importaciones ni `GET .../archivo`) porque el de `apps/api` también lo está. Cambiarlo rompería la paridad del `GET /api/v1`.
- **503 `base de datos no disponible`.** En la API anterior solo ocurre si el puntero de BD del handler es nil. El proceso nuevo no arranca sin BD (R-13 del plan).

## Pruebas de esta revisión

`gofmt`, `go vet` y `go test` de `backend/app` se ejecutan al cerrar el lote 22 (el informe final dice qué corrió y qué no).
