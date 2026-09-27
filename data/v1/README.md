# data/v1

Salida del ETL (`make etl` / `go run ./cmd/etl`). Se regenera desde `data/raw/`; no editar a mano.

| Archivo | Features | Notas |
|---------|----------|--------|
| `areas_verdes.geojson` | 521 | ids `AV-NNNN` |
| `zonas.geojson` | 534 | ids `Z-NNNN`. Sin el campo `jefes` |
| `jardines_reserva.geojson` | 21 | ids `JR-NNNN` |
| `xerofitica.geojson` | 10 | ids `XE-NNNN` |
| `zonas_sector.json` | 534 | `source_index` → sector operativo (`cua-*` o rótulo de lugar). `make sectores` / `go run ./cmd/sectores` |
| `manifest.json` | — | CRS, conteos, SHA-256 de la fuente, nota de PII |

CRS: **EPSG:4326** (lon/lat). El miembro `crs_nota` solo documenta el sistema; las coordenadas siguen GeoJSON RFC 7946.

Atributos vacíos del origen se omiten. Medidas `perimetro_m` y `area_m2` conservan el texto numérico de la fuente.

`zonas_sector.json` no lo escribe `make etl`: sale de `go run ./cmd/sectores`, que agrupa
`data/raw/lote/jefe_de_grupo.json` (copia ya anonimizada, sin red por defecto) por hash de
jefe de grupo y asigna `cua-valeria`/`cua-mateo`/`cua-renato` por frecuencia (el más numeroso
primero), más `campo-deportivo` y `bosque-humedo` para los dos rótulos que no son personas.
No lleva nombres, hashes ni el campo `jefes`. `go run ./cmd/sectores -sql` imprime los
`INSERT` de la migración `044_sector_poligonos.sql`, que completa la columna
`poligonos_cuadrilla.sector` (y la deja disponible en la vista `zonas`).
