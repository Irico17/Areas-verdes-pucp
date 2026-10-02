# Semilla ficticia (develop y qa)

`ficticio.sql` inserta jardines y sectores de demostración dentro del campus, con geometría MultiPolygon EPSG:4326 lo bastante grande para verse en el mapa. Los nombres son ficticios (Prado Ficticio Institucional, Sector Ficticio Valeria, Equipo Norte). No hay personas.

El entrypoint la aplica en develop y qa, y también cuando `SEED_PROFILE=ficticio`. Cada sentencia es `INSERT ... ON CONFLICT DO NOTHING`: si el catastro ya tiene filas (ETL o una siembra anterior), solo agrega los identificadores que faltan. No borra, no trunca y no actualiza filas cargadas.

`SEED_PROFILE=etl` (develop, qa y producción) carga `data/raw` como en producción. Si la base está vacía corre la carga inicial. Si ya hay filas pero faltan áreas con geometría o sectores (por ejemplo quedó solo el polígono diminuto de una siembra vieja), corre `etl-lote`: upsert, sin `TRUNCATE`. Si el catastro ya trae las 521 áreas y las 534 zonas con sector, no vuelve a cargar.

No restaurar un volcado de producción en develop ni en qa. Si hiciera falta una copia para probar, hay que enmascararla antes (nombres, cuentas, evidencias) en una base desechable y recién después cargarla. Este repositorio no trae un script que copie producción.
