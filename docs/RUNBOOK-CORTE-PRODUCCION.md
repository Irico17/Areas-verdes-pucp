# Runbook del corte en producción

Este documento es el checklist del §4.4 de `docs/PLAN-MIGRACION-BACKEND.md`. **No se ejecutó** al dejar el lote 23 en el repositorio. Lo corre el responsable, con orden expresa, sobre la EC2. No forma parte del arranque automático ni de un job de CI.

La imagen nueva ya trae `db/migrations` en `/opt/campus/migrations` (`MIGRATIONS_DIR`). El `docker-entrypoint.sh` ejecuta el binario `migrate` (el binario `api` no migra) y aplica solo las versiones que falten en `schema_migrations`. Requiere `CAMPUS_DEV_PASSWORD` inyectada; sin ella el entrypoint sale con error. Hoy esas versiones nuevas, respecto de un despliegue que se quedó en `046`, son:

- `047_evidencia_evento.sql` (`evidencias.evento_id`, nulo, con FK e índice)
- `048_uuid_cliente_eventos.sql` (`actividad_eventos.uuid_cliente`, índice único parcial)

No hay migración `049`. `areas_verdes.activo` y `zonas_supervision.activo` ya existían.

## Antes de tocar la instancia

- [ ] Confirmar el tag que está sirviendo ahora y anotarlo como **tag de rollback**. Ejemplo: la imagen de `apps/api` que responde hoy en `/health`.
- [ ] Confirmar que no hay otro `apply` de Terraform ni otro despliegue en curso.
- [ ] No usar credenciales de AWS en el repositorio ni en el chat. El acceso a la instancia es el que ya usa el responsable (SSM).

## 1. Congelar el esquema

En la base que está en servicio:

```sql
SELECT version FROM schema_migrations ORDER BY 1;
```

Guardar la salida como `antes-versiones.txt`. Tiene que coincidir con los nombres de `db/migrations/` ya aplicados (hasta `046` si el corte aún no pasó). No editar esos archivos: el runner identifica cada uno por el nombre.

## 2. Backup y snapshot

En la EC2, con la `DATABASE_URL` de esa base (no pegar la clave en el repo):

```bash
bash scripts/backup-postgis.sh /opt/campus/data/backups/pre-backend-v2-FECHA.dump
```

El volcado es `pg_dump -Fc --no-owner --no-acl` e incluye `CREATE EXTENSION`. Copiarlo fuera de la instancia (cubo privado o descarga) y tomar un **snapshot EBS** del volumen de datos. Si las evidencias siguen en disco y no hay bucket, respaldar también `/opt/campus/data/app/evidencias`. El procedimiento de backup y de snapshot está en `docs/OPERACION.md`.

- [ ] Dump con fecha, fuera de la instancia
- [ ] Snapshot EBS anotado (id y fecha)
- [ ] Evidencias respaldadas si no hay bucket

## 3. Conteos antes

```sql
SELECT table_name,
       (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint AS filas
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY 1;
```

Guardar el resultado como `conteos-antes.tsv`. Complementar con `scripts/counts.sh` (geometrías en SRID 4326).

## 4. Ensayo en una copia

Restaurar el dump en un PostGIS desechable (`scripts/restore-postgis.sh` o `scripts/probar-restauracion.sh`). Ahí correr el `migrate` de la **imagen nueva**, no contra la base de servicio. Volver a contar.

- [ ] Cada tabla que ya existía tiene **exactamente** las mismas filas
- [ ] Solo pueden aparecer columnas nuevas (`evento_id`, `uuid_cliente`)
- [ ] `schema_migrations` de la copia = versiones de antes + `047_evidencia_evento.sql` + `048_uuid_cliente_eventos.sql`

## 5. Proteger la carga inicial

Confirmar que `/opt/campus/data/app/.etl-done` existe. El entrypoint nuevo no ejecuta `etl` si `areas_verdes` ya tiene filas. No borrar ese archivo para «forzar» una recarga: el ETL histórico hace `TRUNCATE`.

## 6. Desplegar

Subir la imagen nueva con tag inmutable (el SHA del commit) y reiniciar el servicio como ya se hace en `docs/DEPLOY-AWS.md`. En el script automatizado (`deploy/deploy.sh produccion --aws`), el snapshot EBS del volumen de datos, el backup `pg_dump` y los conteos «antes» se toman sobre la instancia existente en el estado antes de ejecutar `terraform apply` (requiere `DEPLOY_PRIMERA_VEZ=1` si es la creación inicial). El entrypoint aplica solo lo pendiente, cada archivo en su transacción.

- [ ] Tag desplegado anotado
- [ ] Tag de rollback (paso inicial) sigue disponible en el registro de imágenes

## 7. Verificar

- [ ] `GET /health` responde `"status":"ok"` (lo que mira `campus-healthcheck`)
- [ ] `GET /areas-verdes/v1/health` responde
- [ ] Login con una cuenta semilla
- [ ] Conteos después = conteos antes en todas las tablas de negocio existentes (validado por `deploy/comparar_conteos.py` con `deploy/conteos.excluir`)
- [ ] `schema_migrations` = lista de antes + `047` y `048`
- [ ] Mapa, una labor y una evidencia se abren
- [ ] `cambios` reciente no muestra borrados de filas cargadas

## 8. Rollback

**Aplicación.** Volver a desplegar el tag anotado en el paso inicial (la imagen anterior de `apps/api`). Las migraciones `047` y `048` solo agregan columnas nulas: el binario anterior sigue leyendo las columnas que ya conocía.

Aviso: el `accesos.Ensure` de esa imagen anterior puede reescribir `permisos` desde su matriz. Un permiso cambiado con el backend nuevo se perdería; sigue estando en el dump del paso 2.

**Datos, solo si hay corrupción.** Restaurar el dump en una base nueva y apuntar `DATABASE_URL` a esa base. Se pierde lo registrado después del backup. Lo decide el responsable. No hay `DROP DATABASE` de la base en servicio como primer paso.

Las migraciones nuevas no tienen script «down». Quitar una columna sería un `DROP COLUMN`, y eso no se hace. Si una migración quedó mal, se corrige con otra migración aditiva.

## Qué no hacer

- No montar un `.sql` en `docker-entrypoint-initdb.d`.
- No aplicar `schema_nucleo_v0.2.sql` ni `schema_v1.1.sql`.
- No hacer `DROP`, `TRUNCATE` ni `DELETE FROM` para «limpiar» el corte.
- No hacer `docker compose down -v` en la instancia: eso borra el volumen.
