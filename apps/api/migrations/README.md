# Migraciones movidas

Los archivos `NNN_*.sql` están en `db/migrations/` con los mismos nombres. Esta carpeta queda vacía de SQL a propósito: `apps/api` es referencia de código y lee el esquema desde `db/migrations` (en la imagen, `MIGRATIONS_DIR=/opt/campus/migrations`).
