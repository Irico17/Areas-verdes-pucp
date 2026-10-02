# Semilla ficticia (develop y qa)

`ficticio.sql` inserta un jardín y un sector de demostración. Los nombres son ficticios (Jardín Ficticio Norte, Sector Ficticio Norte, Equipo Norte). No hay personas.

El entrypoint la aplica solo cuando `SEED_PROFILE=ficticio` y el catastro está vacío (`areas_verdes`, `poligonos_cuadrilla`, `capas_auxiliares`, `inventario`, `ejemplares`, `asignaciones_poligono`). Si alguna de esas tablas ya tiene filas, no se ejecuta. No borra, no trunca y no actualiza filas cargadas.

Producción usa `SEED_PROFILE=etl`: la carga inicial de `data/raw` corre solo si ese mismo catastro está vacío. Tampoco borra datos ya cargados.

No restaurar un volcado de producción en develop ni en qa. Si hiciera falta una copia para probar, hay que enmascararla antes (nombres, cuentas, evidencias) en una base desechable y recién después cargarla. Este repositorio no trae un script que copie producción.
