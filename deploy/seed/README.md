# Semilla ficticia (develop y qa)

`ficticio.sql` inserta jardines y sectores de demostración en el parque oeste del campus, con geometría MultiPolygon EPSG:4326 visible en el mapa y sin tapar los edificios del núcleo. Los nombres son ficticios (Prado Ficticio Institucional, Sector Ficticio Valeria, Equipo Norte). No hay personas.

El entrypoint la aplica en develop y qa, y también cuando `SEED_PROFILE=ficticio`. Cada sentencia es `INSERT ... ON CONFLICT`, y el conflicto solo reescribe la geometría y el nombre de esas filas ficticias. No borra, no trunca y no modifica el catastro del ETL.

`SEED_PROFILE=etl` (develop, qa y producción) carga `data/raw` como en producción. Si la base está vacía corre la carga inicial. Si ya hay filas pero faltan áreas con geometría o sectores (por ejemplo quedó solo el polígono diminuto de una siembra vieja), corre `etl-lote`: upsert, sin `TRUNCATE`. Si el catastro ya trae las 521 áreas y las 534 zonas con sector, no vuelve a cargar.

No restaurar un volcado de producción en develop ni en qa. Si hiciera falta una copia para probar, hay que enmascararla antes (nombres, cuentas, evidencias) en una base desechable y recién después cargarla. Este repositorio no trae un script que copie producción.
