# Base de datos

Generado por `scripts/generar-esquema-bd.sh` a partir de un PostGIS vacío con las migraciones de `db/migrations` aplicadas por `cmd/migrate`. La fuente de verdad son esas migraciones. La foto SQL, sin datos, está en `db/esquema.sql`.

Migraciones aplicadas: **2**. No hay filas de negocio en este documento. Se omiten `spatial_ref_sys` y las vistas del catálogo de PostGIS.

## Tablas

| Tabla | Columnas | Clave primaria |
| --- | ---: | --- |
| `actividad_avances` | 7 | `id` |
| `actividad_eventos` | 11 | `id` |
| `actividades` | 31 | `id` |
| `areas_verdes` | 17 | `id` |
| `asignaciones_poligono` | 5 | `id` |
| `bebederos` | 13 | `id` |
| `cambios` | 9 | `id` |
| `capas_auxiliares` | 17 | `id` |
| `capataces` | 4 | `id` |
| `catalogos` | 8 | `id` |
| `codigos_historicos` | 5 | `id` |
| `cuadrillas` | 5 | `id` |
| `cuarteles_historico` | 5 | `id` |
| `ejemplares` | 21 | `id` |
| `especies` | 5 | `id` |
| `evidencias` | 15 | `id` |
| `fauna` | 8 | `id` |
| `inventario` | 9 | `id` |
| `jardines_reserva` | 16 | `id` |
| `lotes_importacion` | 9 | `id` |
| `lugares` | 10 | `id` |
| `medidas_palmera` | 9 | `ejemplar_id` |
| `ordenes_servicio` | 13 | `id` |
| `permisos` | 2 | `accion, rol` |
| `personal_ficticio` | 3 | `id` |
| `personal_labor` | 5 | `id` |
| `playas_estacionamiento` | 8 | `id` |
| `podas` | 25 | `id` |
| `poligonos_cuadrilla` | 19 | `id` |
| `poligonos_sector_ref` | 2 | `source_index` |
| `puertas` | 9 | `id` |
| `puntos_pucp` | 10 | `id` |
| `referentes_edificio` | 4 | `id` |
| `reservas_jardin` | 13 | `id` |
| `riego_registros` | 11 | `id` |
| `roles` | 5 | `codigo` |
| `schema_migrations` | 2 | `version` |
| `sectores_capataz` | 7 | `id` |
| `sesiones` | 3 | `token_hash` |
| `solicitudes` | 19 | `id` |
| `tachos` | 29 | `id` |
| `usuarios` | 10 | `id` |
| `veredas_riesgo` | 8 | `id` |
| `vias` | 8 | `id` |
| `vivero_catalogo` | 4 | `id` |
| `vivero_registros` | 14 | `id` |
| `xerofiticas` | 11 | `id` |
| `zonas_origen` | 14 | `id` |
| `zonas_supervision` | 9 | `id` |

## Columnas

### `actividad_avances`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `actividad_id` | `uuid` | NO | FK |
| `fecha` | `date` | NO | — |
| `nota` | `text` | NO | — |
| `area_feature_id` | `text` | YES | — |
| `ejemplar_ref` | `text` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `actividad_eventos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `actividad_id` | `uuid` | NO | FK |
| `tipo` | `text` | NO | — |
| `estado` | `text` | YES | — |
| `capataz_id` | `text` | YES | — |
| `actor_rol` | `text` | NO | — |
| `nota` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `usuario_id` | `bigint` | YES | FK |
| `uuid_cliente` | `uuid` | YES | — |
| `capataz_anterior` | `text` | YES | — |

### `actividades`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `tipo` | `text` | NO | — |
| `estado` | `text` | NO | — |
| `titulo` | `text` | NO | — |
| `detalle` | `text` | NO | — |
| `area_feature_id` | `text` | YES | — |
| `zona_feature_id` | `text` | YES | — |
| `assigned_capataz_id` | `text` | YES | FK |
| `geom` | `geometry` | YES | — |
| `archivada_en` | `timestamp with time zone` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |
| `ejecutor` | `text` | NO | — |
| `motivo_archivo` | `text` | YES | — |
| `lugar_id` | `bigint` | YES | FK |
| `zona_supervision_id` | `bigint` | YES | FK |
| `fecha_solicitud` | `date` | YES | — |
| `fecha_atencion` | `date` | YES | — |
| `cuadrilla_id` | `text` | YES | FK |
| `clase_codigo` | `text` | YES | — |
| `tipo_codigo` | `text` | YES | — |
| `comentario` | `text` | NO | — |
| `lugar_libre` | `text` | NO | — |
| `origen_ref` | `text` | YES | — |
| `origen` | `text` | NO | — |
| `subtipo` | `text` | YES | — |
| `codigo_externo` | `text` | YES | — |
| `unidad_solicitante` | `text` | YES | — |
| `nivel_riesgo` | `text` | YES | — |
| `fecha_programada` | `date` | YES | — |
| `cantidad` | `numeric` | YES | — |

### `areas_verdes`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `source_index` | `integer` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `uso` | `text` | YES | — |
| `proy_riego` | `text` | YES | — |
| `riego_act` | `text` | YES | — |
| `referencia` | `text` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |
| `zona_supervision_id` | `bigint` | YES | FK |
| `activo` | `boolean` | NO | — |
| `origen_ref` | `text` | YES | — |

### `asignaciones_poligono`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `poligono_id` | `bigint` | NO | FK |
| `cuadrilla_id` | `text` | NO | FK |
| `vigente` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `bebederos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `codigo` | `text` | NO | — |
| `subtipo` | `text` | NO | — |
| `estado` | `text` | NO | — |
| `sede` | `text` | YES | — |
| `lat` | `double precision` | YES | — |
| `lon` | `double precision` | YES | — |
| `foto` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `cambios`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `entidad` | `text` | NO | — |
| `entidad_id` | `text` | NO | — |
| `accion` | `text` | NO | — |
| `antes` | `jsonb` | YES | — |
| `despues` | `jsonb` | YES | — |
| `usuario_id` | `bigint` | YES | FK |
| `lote_id` | `bigint` | YES | FK |
| `created_at` | `timestamp with time zone` | NO | — |

### `capas_auxiliares`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `capa` | `text` | NO | — |
| `feature_id` | `text` | NO | — |
| `source_index` | `integer` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `uso` | `text` | YES | — |
| `proy_riego` | `text` | YES | — |
| `clase` | `text` | YES | — |
| `riego_act` | `text` | YES | — |
| `referencia` | `text` | YES | — |
| `pertenecen` | `text` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `capataces`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `text` | NO | PK |
| `equipo` | `text` | NO | — |
| `turno` | `text` | NO | — |
| `activo` | `boolean` | NO | — |

### `catalogos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `clase` | `text` | NO | — |
| `codigo` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `orden` | `integer` | NO | — |
| `provisional` | `boolean` | NO | — |
| `padre_codigo` | `text` | YES | — |

### `codigos_historicos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `ejemplar_id` | `bigint` | NO | FK |
| `codigo_anterior` | `text` | NO | — |
| `codigo_nuevo` | `text` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `cuadrillas`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `text` | NO | PK |
| `nombre_ficticio` | `text` | NO | — |
| `turno` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `cuarteles_historico`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `codigo` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `geom` | `geometry` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `ejemplares`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `numero_origen` | `integer` | YES | — |
| `codigo` | `text` | YES | — |
| `especie_id` | `bigint` | YES | FK |
| `nombre_comun` | `text` | YES | — |
| `tipo_vegetacion` | `text` | YES | — |
| `cantidad` | `integer` | NO | — |
| `area_verde_id` | `bigint` | YES | FK |
| `ubicacion_lugar_id` | `bigint` | YES | FK |
| `referencia` | `text` | YES | — |
| `lat` | `double precision` | YES | — |
| `lon` | `double precision` | YES | — |
| `observacion_fen_2026` | `text` | YES | — |
| `foto_id` | `uuid` | YES | FK |
| `salud` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |
| `sector_cuartel_id` | `bigint` | YES | FK |

### `especies`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `nombre_cientifico` | `text` | NO | — |
| `nombre_comun` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `evidencias`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `actividad_id` | `uuid` | YES | FK |
| `solicitud_id` | `uuid` | YES | FK |
| `orden_id` | `uuid` | YES | FK |
| `nombre` | `text` | NO | — |
| `mime` | `text` | NO | — |
| `bytes` | `integer` | NO | — |
| `ruta` | `text` | NO | — |
| `nota` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `sha256` | `text` | YES | — |
| `lat` | `double precision` | YES | — |
| `lon` | `double precision` | YES | — |
| `exif` | `jsonb` | YES | — |
| `evento_id` | `bigint` | YES | FK |

### `fauna`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `nombre` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `inventario`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `capa` | `text` | NO | — |
| `feature_id` | `text` | NO | — |
| `nombre` | `text` | YES | — |
| `subtipo` | `text` | YES | — |
| `detalle` | `text` | YES | — |
| `lugar` | `text` | YES | — |
| `foto` | `text` | YES | — |
| `geom` | `geometry` | YES | — |

### `jardines_reserva`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `uso` | `text` | YES | — |
| `proy_riego` | `text` | YES | — |
| `riego_act` | `text` | YES | — |
| `referencia` | `text` | YES | — |
| `pertenecen` | `text` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `lotes_importacion`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `entidad` | `text` | NO | — |
| `estado` | `text` | NO | — |
| `usuario_id` | `bigint` | NO | FK |
| `filas` | `integer` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `revertido_en` | `timestamp with time zone` | YES | — |
| `contenido` | `bytea` | YES | — |
| `nombre_archivo` | `text` | NO | — |

### `lugares`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `nombre` | `text` | NO | — |
| `nombre_norm` | `text` | NO | — |
| `lat` | `double precision` | NO | — |
| `lon` | `double precision` | NO | — |
| `zona_supervision_id` | `bigint` | YES | FK |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `medidas_palmera`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `ejemplar_id` | `bigint` | NO | PK FK |
| `utm_norte` | `double precision` | YES | — |
| `utm_este` | `double precision` | YES | — |
| `altura` | `double precision` | YES | — |
| `altura_fuste` | `double precision` | YES | — |
| `dap` | `double precision` | YES | — |
| `radio` | `double precision` | YES | — |
| `zunchado` | `boolean` | YES | — |
| `baja_en` | `timestamp with time zone` | YES | — |

### `ordenes_servicio`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `actividad_id` | `uuid` | NO | FK |
| `empresa` | `text` | NO | — |
| `referencia` | `text` | NO | — |
| `frecuencia` | `text` | NO | — |
| `estado` | `text` | NO | — |
| `conformidad` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `periodo_inicio` | `date` | YES | — |
| `periodo_fin` | `date` | YES | — |
| `reporte_proveedor` | `text` | NO | — |
| `empresa_id` | `bigint` | YES | FK |
| `frecuencia_id` | `bigint` | YES | FK |

### `permisos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `rol` | `text` | NO | PK FK |
| `accion` | `text` | NO | PK |

### `personal_ficticio`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `text` | NO | PK |
| `nombre_ficticio` | `text` | NO | — |
| `activo` | `boolean` | NO | — |

### `personal_labor`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `actividad_id` | `uuid` | NO | FK |
| `nombre_ficticio` | `text` | NO | — |
| `rol_campo` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `playas_estacionamiento`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `codigo` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `podas`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `codigo` | `text` | NO | — |
| `codigo_externo` | `text` | YES | — |
| `tipo` | `text` | NO | — |
| `tipo_actividad` | `text` | NO | — |
| `fecha_reporte` | `date` | YES | — |
| `fecha_ejecucion` | `date` | YES | — |
| `personal_ficticio` | `text` | NO | — |
| `ubicacion` | `text` | NO | — |
| `lugar_id` | `text` | YES | — |
| `unidad` | `text` | NO | — |
| `cantidad_pedida` | `numeric` | NO | — |
| `cantidad_ejecutada` | `numeric` | NO | — |
| `tipo_vegetacion` | `text` | NO | — |
| `nombre_comun` | `text` | NO | — |
| `nombre_cientifico` | `text` | NO | — |
| `especie_id` | `text` | YES | — |
| `prioridad` | `text` | NO | — |
| `comentario` | `text` | NO | — |
| `solicitud_id` | `uuid` | YES | FK |
| `actividad_id` | `uuid` | YES | FK |
| `origen_ref` | `text` | YES | — |
| `archivada_en` | `timestamp with time zone` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `poligonos_cuadrilla`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `source_index` | `integer` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `uso` | `text` | YES | — |
| `proy_riego` | `text` | YES | — |
| `riego_act` | `text` | YES | — |
| `referencia` | `text` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | NO | — |
| `cuadrilla_id` | `text` | YES | FK |
| `zona_supervision_id` | `bigint` | YES | FK |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |
| `sector` | `text` | YES | — |

### `poligonos_sector_ref`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `source_index` | `integer` | NO | PK |
| `sector` | `text` | NO | — |

### `puertas`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `puntos_pucp`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `titulo` | `text` | NO | — |
| `lat` | `double precision` | NO | — |
| `lon` | `double precision` | NO | — |
| `url` | `text` | YES | — |
| `geom` | `geometry` | NO | — |
| `origen_ref` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `referentes_edificio`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `lugar_id` | `bigint` | NO | FK |
| `edificio_id` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |

### `reservas_jardin`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `jardin_id` | `bigint` | YES | FK |
| `fecha` | `date` | NO | — |
| `hora_inicio` | `time without time zone` | NO | — |
| `hora_fin` | `time without time zone` | NO | — |
| `estado` | `text` | NO | — |
| `evento` | `text` | NO | — |
| `unidad` | `text` | YES | — |
| `origen` | `text` | NO | — |
| `origen_ref` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `riego_registros`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `sector` | `text` | NO | — |
| `turno` | `text` | NO | — |
| `capataz_id` | `text` | YES | FK |
| `fecha` | `date` | NO | — |
| `nota` | `text` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `zona_supervision_id` | `bigint` | YES | FK |
| `ciclo` | `text` | NO | — |
| `superficie_m2` | `numeric` | YES | — |
| `sector_id` | `bigint` | YES | FK |

### `roles`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `codigo` | `text` | NO | PK |
| `nombre` | `text` | NO | — |
| `orden` | `integer` | NO | — |
| `descripcion` | `text` | YES | — |
| `activo` | `boolean` | NO | — |

### `schema_migrations`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `version` | `text` | NO | PK |
| `applied_at` | `timestamp with time zone` | NO | — |

### `sectores_capataz`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `codigo` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `color` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `sesiones`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `token_hash` | `text` | NO | PK |
| `usuario_id` | `bigint` | NO | FK |
| `expires_at` | `timestamp with time zone` | NO | — |

### `solicitudes`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `codigo_externo` | `text` | YES | — |
| `fuente` | `text` | NO | — |
| `titulo` | `text` | NO | — |
| `detalle` | `text` | NO | — |
| `prioridad` | `text` | NO | — |
| `estado` | `text` | NO | — |
| `lugar` | `text` | YES | — |
| `cantidad` | `integer` | YES | — |
| `actividad_id` | `uuid` | YES | FK |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |
| `origen_ref` | `text` | YES | — |
| `archivada_en` | `timestamp with time zone` | YES | — |
| `cantidad_solicitada` | `integer` | YES | — |
| `cantidad_ejecutada` | `integer` | YES | — |
| `lugar_id` | `bigint` | YES | FK |
| `lat` | `double precision` | YES | — |
| `lon` | `double precision` | YES | — |

### `tachos`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `codigo` | `text` | NO | — |
| `lat` | `double precision` | YES | — |
| `lon` | `double precision` | YES | — |
| `nota` | `text` | YES | — |
| `lugar` | `text` | YES | — |
| `espacios` | `text` | YES | — |
| `accion` | `text` | YES | — |
| `tacho_actual` | `text` | YES | — |
| `tacho_nuevo` | `text` | YES | — |
| `recomendaciones` | `text` | YES | — |
| `no_aprovechables` | `integer` | NO | — |
| `papel_carton` | `integer` | NO | — |
| `plastico` | `integer` | NO | — |
| `vidrio` | `integer` | NO | — |
| `pilas` | `integer` | NO | — |
| `peligrosos` | `integer` | NO | — |
| `raee` | `integer` | NO | — |
| `metales` | `integer` | NO | — |
| `aniquem` | `integer` | NO | — |
| `intermedios_plastico` | `integer` | NO | — |
| `intermedios_metal` | `integer` | NO | — |
| `foto` | `text` | YES | — |
| `zona_supervision_id` | `bigint` | YES | FK |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `usuarios`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `usuario` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `rol` | `text` | NO | FK |
| `capataz_id` | `text` | YES | FK |
| `password_hash` | `text` | NO | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `cuadrilla_id` | `text` | YES | FK |
| `debe_cambiar_password` | `boolean` | NO | — |

### `veredas_riesgo`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `nota` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `vias`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `nombre` | `text` | YES | — |
| `geom` | `geometry` | YES | — |
| `activo` | `boolean` | NO | — |
| `origen_ref` | `text` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `vivero_catalogo`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `clase` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `activo` | `boolean` | NO | — |

### `vivero_registros`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `uuid` | NO | PK |
| `fecha` | `date` | YES | — |
| `area` | `text` | NO | — |
| `subproceso` | `text` | NO | — |
| `etapa` | `text` | NO | — |
| `descripcion` | `text` | NO | — |
| `observaciones` | `text` | NO | — |
| `responsables` | `text` | NO | — |
| `lugar_id` | `text` | YES | — |
| `lugar_libre` | `text` | NO | — |
| `origen_ref` | `text` | YES | — |
| `archivada_en` | `timestamp with time zone` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `xerofiticas`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `clase` | `text` | YES | — |
| `riego` | `text` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `geom` | `geometry` | YES | — |
| `origen_ref` | `text` | YES | — |
| `activo` | `boolean` | NO | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `zonas_origen`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `feature_id` | `text` | NO | — |
| `source_index` | `integer` | NO | — |
| `codigo` | `text` | YES | — |
| `nombre` | `text` | YES | — |
| `uso` | `text` | YES | — |
| `proy_riego` | `text` | YES | — |
| `riego_act` | `text` | YES | — |
| `referencia` | `text` | YES | — |
| `perimetro_m` | `double precision` | YES | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

### `zonas_supervision`

| Columna | Tipo | Nulo | Clave |
| --- | --- | --- | --- |
| `id` | `bigint` | NO | PK |
| `codigo` | `text` | NO | — |
| `nombre` | `text` | NO | — |
| `area_m2` | `double precision` | YES | — |
| `geom` | `geometry` | NO | — |
| `activo` | `boolean` | NO | — |
| `origen_ref` | `text` | YES | — |
| `created_at` | `timestamp with time zone` | NO | — |
| `updated_at` | `timestamp with time zone` | NO | — |

## Relaciones

| Tabla | Columna | Referencia | Nulo |
| --- | --- | --- | --- |
| `actividad_avances` | `actividad_id` | `actividades.id` | NO |
| `actividad_eventos` | `actividad_id` | `actividades.id` | NO |
| `actividad_eventos` | `usuario_id` | `usuarios.id` | YES |
| `actividades` | `assigned_capataz_id` | `capataces.id` | YES |
| `actividades` | `cuadrilla_id` | `cuadrillas.id` | YES |
| `actividades` | `lugar_id` | `lugares.id` | YES |
| `actividades` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `areas_verdes` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `asignaciones_poligono` | `cuadrilla_id` | `cuadrillas.id` | NO |
| `asignaciones_poligono` | `poligono_id` | `poligonos_cuadrilla.id` | NO |
| `cambios` | `lote_id` | `lotes_importacion.id` | YES |
| `cambios` | `usuario_id` | `usuarios.id` | YES |
| `codigos_historicos` | `ejemplar_id` | `ejemplares.id` | NO |
| `ejemplares` | `area_verde_id` | `areas_verdes.id` | YES |
| `ejemplares` | `especie_id` | `especies.id` | YES |
| `ejemplares` | `foto_id` | `evidencias.id` | YES |
| `ejemplares` | `sector_cuartel_id` | `catalogos.id` | YES |
| `ejemplares` | `ubicacion_lugar_id` | `lugares.id` | YES |
| `evidencias` | `actividad_id` | `actividades.id` | YES |
| `evidencias` | `evento_id` | `actividad_eventos.id` | YES |
| `evidencias` | `orden_id` | `ordenes_servicio.id` | YES |
| `evidencias` | `solicitud_id` | `solicitudes.id` | YES |
| `lotes_importacion` | `usuario_id` | `usuarios.id` | NO |
| `lugares` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `medidas_palmera` | `ejemplar_id` | `ejemplares.id` | NO |
| `ordenes_servicio` | `actividad_id` | `actividades.id` | NO |
| `ordenes_servicio` | `empresa_id` | `catalogos.id` | YES |
| `ordenes_servicio` | `frecuencia_id` | `catalogos.id` | YES |
| `permisos` | `rol` | `roles.codigo` | NO |
| `personal_labor` | `actividad_id` | `actividades.id` | NO |
| `podas` | `actividad_id` | `actividades.id` | YES |
| `podas` | `solicitud_id` | `solicitudes.id` | YES |
| `poligonos_cuadrilla` | `cuadrilla_id` | `cuadrillas.id` | YES |
| `poligonos_cuadrilla` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `referentes_edificio` | `lugar_id` | `lugares.id` | NO |
| `reservas_jardin` | `jardin_id` | `jardines_reserva.id` | YES |
| `riego_registros` | `capataz_id` | `capataces.id` | YES |
| `riego_registros` | `sector_id` | `sectores_capataz.id` | YES |
| `riego_registros` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `sesiones` | `usuario_id` | `usuarios.id` | NO |
| `solicitudes` | `actividad_id` | `actividades.id` | YES |
| `solicitudes` | `lugar_id` | `lugares.id` | YES |
| `tachos` | `zona_supervision_id` | `zonas_supervision.id` | YES |
| `usuarios` | `capataz_id` | `capataces.id` | YES |
| `usuarios` | `cuadrilla_id` | `cuadrillas.id` | YES |
| `usuarios` | `rol` | `roles.codigo` | NO |

## Diagrama

```mermaid
erDiagram
    actividad_avances {
        uuid id PK
        uuid actividad_id FK
        date fecha
        text nota
        text area_feature_id
        text ejemplar_ref
        timestamptz created_at
    }
    actividad_eventos {
        bigint id PK
        uuid actividad_id FK
        text tipo
        text estado
        text capataz_id
        text actor_rol
        text nota
        timestamptz created_at
        bigint usuario_id FK
        uuid uuid_cliente
        text capataz_anterior
    }
    actividades {
        uuid id PK
        text tipo
        text estado
        text titulo
        text detalle
        text area_feature_id
        text zona_feature_id
        text assigned_capataz_id FK
        geometry geom
        timestamptz archivada_en
        timestamptz created_at
        timestamptz updated_at
        text ejecutor
        text motivo_archivo
        bigint lugar_id FK
        bigint zona_supervision_id FK
        date fecha_solicitud
        date fecha_atencion
        text cuadrilla_id FK
        text clase_codigo
        text tipo_codigo
        text comentario
        text lugar_libre
        text origen_ref
        text origen
        text subtipo
        text codigo_externo
        text unidad_solicitante
        text nivel_riesgo
        date fecha_programada
        numeric cantidad
    }
    areas_verdes {
        bigint id PK
        text feature_id
        integer source_index
        text codigo
        text nombre
        text uso
        text proy_riego
        text riego_act
        text referencia
        float8 perimetro_m
        float8 area_m2
        geometry geom
        timestamptz created_at
        timestamptz updated_at
        bigint zona_supervision_id FK
        boolean activo
        text origen_ref
    }
    asignaciones_poligono {
        bigint id PK
        bigint poligono_id FK
        text cuadrilla_id FK
        boolean vigente
        timestamptz created_at
    }
    bebederos {
        bigint id PK
        text codigo
        text subtipo
        text estado
        text sede
        float8 lat
        float8 lon
        text foto
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    cambios {
        bigint id PK
        text entidad
        text entidad_id
        text accion
        jsonb antes
        jsonb despues
        bigint usuario_id FK
        bigint lote_id FK
        timestamptz created_at
    }
    capas_auxiliares {
        bigint id PK
        text capa
        text feature_id
        integer source_index
        text codigo
        text nombre
        text uso
        text proy_riego
        text clase
        text riego_act
        text referencia
        text pertenecen
        float8 perimetro_m
        float8 area_m2
        geometry geom
        timestamptz created_at
        timestamptz updated_at
    }
    capataces {
        text id PK
        text equipo
        text turno
        boolean activo
    }
    catalogos {
        bigint id PK
        text clase
        text codigo
        text nombre
        boolean activo
        integer orden
        boolean provisional
        text padre_codigo
    }
    codigos_historicos {
        bigint id PK
        bigint ejemplar_id FK
        text codigo_anterior
        text codigo_nuevo
        timestamptz created_at
    }
    cuadrillas {
        text id PK
        text nombre_ficticio
        text turno
        boolean activo
        timestamptz created_at
    }
    cuarteles_historico {
        bigint id PK
        text codigo
        text nombre
        geometry geom
        timestamptz created_at
    }
    ejemplares {
        bigint id PK
        integer numero_origen
        text codigo
        bigint especie_id FK
        text nombre_comun
        text tipo_vegetacion
        integer cantidad
        bigint area_verde_id FK
        bigint ubicacion_lugar_id FK
        text referencia
        float8 lat
        float8 lon
        text observacion_fen_2026
        uuid foto_id FK
        text salud
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
        bigint sector_cuartel_id FK
    }
    especies {
        bigint id PK
        text nombre_cientifico
        text nombre_comun
        boolean activo
        timestamptz created_at
    }
    evidencias {
        uuid id PK
        uuid actividad_id FK
        uuid solicitud_id FK
        uuid orden_id FK
        text nombre
        text mime
        integer bytes
        text ruta
        text nota
        timestamptz created_at
        text sha256
        float8 lat
        float8 lon
        jsonb exif
        bigint evento_id FK
    }
    fauna {
        bigint id PK
        text feature_id
        text nombre
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    inventario {
        bigint id PK
        text capa
        text feature_id
        text nombre
        text subtipo
        text detalle
        text lugar
        text foto
        geometry geom
    }
    jardines_reserva {
        bigint id PK
        text feature_id
        text codigo
        text nombre
        text uso
        text proy_riego
        text riego_act
        text referencia
        text pertenecen
        float8 perimetro_m
        float8 area_m2
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    lotes_importacion {
        bigint id PK
        text entidad
        text estado
        bigint usuario_id FK
        integer filas
        timestamptz created_at
        timestamptz revertido_en
        bytea contenido
        text nombre_archivo
    }
    lugares {
        bigint id PK
        text nombre
        text nombre_norm
        float8 lat
        float8 lon
        bigint zona_supervision_id FK
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    medidas_palmera {
        bigint ejemplar_id PK FK
        float8 utm_norte
        float8 utm_este
        float8 altura
        float8 altura_fuste
        float8 dap
        float8 radio
        boolean zunchado
        timestamptz baja_en
    }
    ordenes_servicio {
        uuid id PK
        uuid actividad_id FK
        text empresa
        text referencia
        text frecuencia
        text estado
        text conformidad
        timestamptz created_at
        date periodo_inicio
        date periodo_fin
        text reporte_proveedor
        bigint empresa_id FK
        bigint frecuencia_id FK
    }
    permisos {
        text rol PK FK
        text accion PK
    }
    personal_ficticio {
        text id PK
        text nombre_ficticio
        boolean activo
    }
    personal_labor {
        uuid id PK
        uuid actividad_id FK
        text nombre_ficticio
        text rol_campo
        timestamptz created_at
    }
    playas_estacionamiento {
        bigint id PK
        text feature_id
        text codigo
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    podas {
        uuid id PK
        text codigo
        text codigo_externo
        text tipo
        text tipo_actividad
        date fecha_reporte
        date fecha_ejecucion
        text personal_ficticio
        text ubicacion
        text lugar_id
        text unidad
        numeric cantidad_pedida
        numeric cantidad_ejecutada
        text tipo_vegetacion
        text nombre_comun
        text nombre_cientifico
        text especie_id
        text prioridad
        text comentario
        uuid solicitud_id FK
        uuid actividad_id FK
        text origen_ref
        timestamptz archivada_en
        timestamptz created_at
        timestamptz updated_at
    }
    poligonos_cuadrilla {
        bigint id PK
        text feature_id
        integer source_index
        text codigo
        text nombre
        text uso
        text proy_riego
        text riego_act
        text referencia
        float8 perimetro_m
        float8 area_m2
        geometry geom
        text cuadrilla_id FK
        bigint zona_supervision_id FK
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
        text sector
    }
    poligonos_sector_ref {
        integer source_index PK
        text sector
    }
    puertas {
        bigint id PK
        text feature_id
        text codigo
        text nombre
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    puntos_pucp {
        bigint id PK
        text titulo
        float8 lat
        float8 lon
        text url
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    referentes_edificio {
        bigint id PK
        bigint lugar_id FK
        text edificio_id
        timestamptz created_at
    }
    reservas_jardin {
        bigint id PK
        bigint jardin_id FK
        date fecha
        time_without_time_zone hora_inicio
        time_without_time_zone hora_fin
        text estado
        text evento
        text unidad
        text origen
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    riego_registros {
        uuid id PK
        text sector
        text turno
        text capataz_id FK
        date fecha
        text nota
        timestamptz created_at
        bigint zona_supervision_id FK
        text ciclo
        numeric superficie_m2
        bigint sector_id FK
    }
    roles {
        text codigo PK
        text nombre
        integer orden
        text descripcion
        boolean activo
    }
    schema_migrations {
        text version PK
        timestamptz applied_at
    }
    sectores_capataz {
        bigint id PK
        text codigo
        text nombre
        text color
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    sesiones {
        text token_hash PK
        bigint usuario_id FK
        timestamptz expires_at
    }
    solicitudes {
        uuid id PK
        text codigo_externo
        text fuente
        text titulo
        text detalle
        text prioridad
        text estado
        text lugar
        integer cantidad
        uuid actividad_id FK
        timestamptz created_at
        timestamptz updated_at
        text origen_ref
        timestamptz archivada_en
        integer cantidad_solicitada
        integer cantidad_ejecutada
        bigint lugar_id FK
        float8 lat
        float8 lon
    }
    tachos {
        bigint id PK
        text codigo
        float8 lat
        float8 lon
        text nota
        text lugar
        text espacios
        text accion
        text tacho_actual
        text tacho_nuevo
        text recomendaciones
        integer no_aprovechables
        integer papel_carton
        integer plastico
        integer vidrio
        integer pilas
        integer peligrosos
        integer raee
        integer metales
        integer aniquem
        integer intermedios_plastico
        integer intermedios_metal
        text foto
        bigint zona_supervision_id FK
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    usuarios {
        bigint id PK
        text usuario
        text nombre
        text rol FK
        text capataz_id FK
        text password_hash
        boolean activo
        timestamptz created_at
        text cuadrilla_id FK
        boolean debe_cambiar_password
    }
    veredas_riesgo {
        bigint id PK
        text feature_id
        text nota
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    vias {
        bigint id PK
        text feature_id
        text nombre
        geometry geom
        boolean activo
        text origen_ref
        timestamptz created_at
        timestamptz updated_at
    }
    vivero_catalogo {
        bigint id PK
        text clase
        text nombre
        boolean activo
    }
    vivero_registros {
        uuid id PK
        date fecha
        text area
        text subproceso
        text etapa
        text descripcion
        text observaciones
        text responsables
        text lugar_id
        text lugar_libre
        text origen_ref
        timestamptz archivada_en
        timestamptz created_at
        timestamptz updated_at
    }
    xerofiticas {
        bigint id PK
        text feature_id
        text clase
        text riego
        float8 area_m2
        float8 perimetro_m
        geometry geom
        text origen_ref
        boolean activo
        timestamptz created_at
        timestamptz updated_at
    }
    zonas_origen {
        bigint id PK
        text feature_id
        integer source_index
        text codigo
        text nombre
        text uso
        text proy_riego
        text riego_act
        text referencia
        float8 perimetro_m
        float8 area_m2
        geometry geom
        timestamptz created_at
        timestamptz updated_at
    }
    zonas_supervision {
        bigint id PK
        text codigo
        text nombre
        float8 area_m2
        geometry geom
        boolean activo
        text origen_ref
        timestamptz created_at
        timestamptz updated_at
    }
    actividades ||--|{ actividad_avances : actividad_id
    actividades ||--|{ actividad_eventos : actividad_id
    usuarios ||--o{ actividad_eventos : usuario_id
    capataces ||--o{ actividades : assigned_capataz_id
    cuadrillas ||--o{ actividades : cuadrilla_id
    lugares ||--o{ actividades : lugar_id
    zonas_supervision ||--o{ actividades : zona_supervision_id
    zonas_supervision ||--o{ areas_verdes : zona_supervision_id
    cuadrillas ||--|{ asignaciones_poligono : cuadrilla_id
    poligonos_cuadrilla ||--|{ asignaciones_poligono : poligono_id
    lotes_importacion ||--o{ cambios : lote_id
    usuarios ||--o{ cambios : usuario_id
    ejemplares ||--|{ codigos_historicos : ejemplar_id
    areas_verdes ||--o{ ejemplares : area_verde_id
    especies ||--o{ ejemplares : especie_id
    evidencias ||--o{ ejemplares : foto_id
    catalogos ||--o{ ejemplares : sector_cuartel_id
    lugares ||--o{ ejemplares : ubicacion_lugar_id
    actividades ||--o{ evidencias : actividad_id
    actividad_eventos ||--o{ evidencias : evento_id
    ordenes_servicio ||--o{ evidencias : orden_id
    solicitudes ||--o{ evidencias : solicitud_id
    usuarios ||--|{ lotes_importacion : usuario_id
    zonas_supervision ||--o{ lugares : zona_supervision_id
    ejemplares ||--|{ medidas_palmera : ejemplar_id
    actividades ||--|{ ordenes_servicio : actividad_id
    catalogos ||--o{ ordenes_servicio : empresa_id
    catalogos ||--o{ ordenes_servicio : frecuencia_id
    roles ||--|{ permisos : rol
    actividades ||--|{ personal_labor : actividad_id
    actividades ||--o{ podas : actividad_id
    solicitudes ||--o{ podas : solicitud_id
    cuadrillas ||--o{ poligonos_cuadrilla : cuadrilla_id
    zonas_supervision ||--o{ poligonos_cuadrilla : zona_supervision_id
    lugares ||--|{ referentes_edificio : lugar_id
    jardines_reserva ||--o{ reservas_jardin : jardin_id
    capataces ||--o{ riego_registros : capataz_id
    sectores_capataz ||--o{ riego_registros : sector_id
    zonas_supervision ||--o{ riego_registros : zona_supervision_id
    usuarios ||--|{ sesiones : usuario_id
    actividades ||--o{ solicitudes : actividad_id
    lugares ||--o{ solicitudes : lugar_id
    zonas_supervision ||--o{ tachos : zona_supervision_id
    capataces ||--o{ usuarios : capataz_id
    cuadrillas ||--o{ usuarios : cuadrilla_id
    roles ||--|{ usuarios : rol
```
