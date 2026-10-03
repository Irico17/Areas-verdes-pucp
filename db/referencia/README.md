# Referencia de esquemas ajenos

Esta carpeta no se monta en Postgres y no entra en `docker-entrypoint-initdb.d`.

`db/schema_nucleo_v0.2.sql` no está en este repositorio. El esquema del equipo (`schema_nucleo_v0.2.sql`, luego `schema_v1.1.sql`) no se adopta y no se aplica sobre la base con datos. Copiarlo aquí, si alguna vez hace falta compararlo, no autoriza ejecutarlo.

La definición del esquema de VerdePUCP es `db/migrations/`: `001_esquema_base.sql` y `002_catalogos_base.sql`. La serie anterior está en `migraciones-historicas/` y no se aplica. La foto del resultado es `db/esquema.sql`.
