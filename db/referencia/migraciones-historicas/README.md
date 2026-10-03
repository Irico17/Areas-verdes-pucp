# Serie histórica de migraciones

Estos archivos son la serie `001_postgis.sql` … `078_medidas_palmera_baja.sql`.
Ya no los aplica `cmd/migrate`. El esquema que dejaban, más los datos de
referencia de una base vacía, está consolidado en `db/migrations/001_esquema_base.sql`.

Se conservan para leer el camino de cada cambio y para la prueba que recorre
una base a medias (códigos duplicados y la copia de polígonos). Una base que
ya registró `078_medidas_palmera_baja.sql` no vuelve a ejecutar la baseline.

Los cambios nuevos van en `db/migrations/`, después de `001_esquema_base.sql`.
No se reordenan ni se renombran. La foto del resultado es `db/esquema.sql`.
