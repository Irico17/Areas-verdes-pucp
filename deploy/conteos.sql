-- Conteos de tablas public (el SQL del §4.4). No modifica datos.
-- En la EC2 o local cuenta todas las tablas base del esquema public.
-- deploy/comparar_conteos.py y deploy/deploy.sh excluyen las tablas
-- técnicas definidas en deploy/conteos.excluir (sesiones, schema_migrations).
SELECT table_name,
       (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint AS filas
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY 1;
