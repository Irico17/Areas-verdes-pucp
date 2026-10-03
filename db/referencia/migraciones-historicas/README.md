# Serie histórica de migraciones

Estos archivos son la serie `001_postgis.sql` … `078_medidas_palmera_baja.sql`.
Ya no los aplica `cmd/migrate`. El esquema que dejaban está en
`db/migrations/001_esquema_base.sql` y los catálogos de una base vacía en
`db/migrations/002_catalogos_base.sql`.

Se conservan fuera del directorio que ejecuta el runner: hacen falta para
reconocer una base que ya aplicó esta serie y para las pruebas de transición,
códigos duplicados y copia de polígonos. Si `schema_migrations` tiene
`078_medidas_palmera_baja.sql`, `001` y `002` se registran sin ejecutarse.

Los cambios nuevos van en `db/migrations/`, desde `003_*.sql`.
No se reordenan ni se renombran. La foto del resultado es `db/esquema.sql`.
