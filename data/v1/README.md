# data/v1

Salida del ETL (`make etl` / `go run ./cmd/etl`). Se regenera desde `data/raw/`; no editar a mano.

| Archivo | Features | Notas |
|---------|----------|--------|
| `areas_verdes.geojson` | 521 | ids `AV-NNNN` |
| `zonas.geojson` | 534 | ids `Z-NNNN`. Sin el campo `jefes` |
| `jardines_reserva.geojson` | 21 | ids `JR-NNNN` |
| `xerofitica.geojson` | 10 | ids `XE-NNNN` |
| `manifest.json` | — | CRS, conteos, SHA-256 de la fuente, nota de PII |

CRS: **EPSG:4326** (lon/lat). El miembro `crs_nota` solo documenta el sistema; las coordenadas siguen GeoJSON RFC 7946.

Atributos vacíos del origen se omiten. Medidas `perimetro_m` y `area_m2` conservan el texto numérico de la fuente.
