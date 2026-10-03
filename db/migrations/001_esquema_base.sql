-- Esquema base consolidado. Lo aplica cmd/migrate una sola vez.
-- Sustituye la serie histórica 001_postgis.sql … 078_medidas_palmera_baja.sql,
-- archivada en db/referencia/migraciones-historicas/ y que ya no se ejecuta.
-- Una base que ya tenga registrada 078_medidas_palmera_baja.sql no vuelve a
-- ejecutar este archivo: el esquema ya está en el estado final.
-- La foto normalizada, sin datos, es db/esquema.sql. La regenera
-- scripts/generar-esquema-bd.sh. La fuente de verdad de los cambios futuros
-- son las migraciones posteriores a esta baseline.
--
-- Secciones: acceso, catastro, operación, inventario, catálogos, auditoría,
-- evidencias y los datos de referencia que la serie dejaba en una base vacía.
-- Usuarios y sesiones no van aquí: los crea la semilla de accesos al migrar.

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', true);
CREATE EXTENSION IF NOT EXISTS postgis;
SELECT pg_catalog.set_config('search_path', '', true);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
-- ========================================================================
-- Acceso
-- ========================================================================
-- roles
CREATE TABLE public.roles (
    codigo text NOT NULL,
    nombre text NOT NULL,
    orden integer DEFAULT 0 NOT NULL,
    descripcion text,
    activo boolean DEFAULT true NOT NULL
);
COMMENT ON TABLE public.roles IS 'Catálogo semilla de roles. No es un editor de políticas ni el SSO.';
ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (codigo);
-- usuarios
CREATE TABLE public.usuarios (
    id bigint NOT NULL,
    usuario text NOT NULL,
    nombre text NOT NULL,
    rol text NOT NULL,
    capataz_id text,
    password_hash text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    cuadrilla_id text,
    debe_cambiar_password boolean DEFAULT false NOT NULL
);
COMMENT ON TABLE public.usuarios IS 'Cuentas locales de desarrollo. No son el SSO de la PUCP.';
COMMENT ON COLUMN public.usuarios.cuadrilla_id IS 'Cuadrilla que opera la cuenta, si aplica. No identifica a una persona real.';
COMMENT ON COLUMN public.usuarios.debe_cambiar_password IS 'True cuando la jefatura entregó una clave inicial y la persona todavía no la cambió.';
CREATE SEQUENCE public.usuarios_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.usuarios_id_seq OWNED BY public.usuarios.id;
ALTER TABLE ONLY public.usuarios ALTER COLUMN id SET DEFAULT nextval('public.usuarios_id_seq'::regclass);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_usuario_key UNIQUE (usuario);
-- sesiones
CREATE TABLE public.sesiones (
    token_hash text NOT NULL,
    usuario_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL
);
ALTER TABLE ONLY public.sesiones
    ADD CONSTRAINT sesiones_pkey PRIMARY KEY (token_hash);
-- permisos
CREATE TABLE public.permisos (
    rol text NOT NULL,
    accion text NOT NULL
);
ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_pkey PRIMARY KEY (rol, accion);
CREATE UNIQUE INDEX permisos_rol_accion_uidx ON public.permisos USING btree (rol, accion);
-- ========================================================================
-- Catastro
-- ========================================================================
CREATE FUNCTION public.catastro_geom_4326(geojson text) RETURNS public.geometry
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
  g geometry;
BEGIN
  IF geojson IS NULL OR btrim(geojson) = '' OR lower(btrim(geojson)) = 'null' THEN
    RETURN NULL;
  END IF;

  g := ST_CollectionExtract(
    ST_MakeValid(
      ST_Force2D(
        ST_SetSRID(ST_GeomFromGeoJSON(geojson), 4326)
      )
    ),
    3
  );

  IF g IS NULL OR ST_IsEmpty(g) THEN
    RETURN NULL;
  END IF;

  RETURN ST_Multi(g)::geometry(MultiPolygon, 4326);
END;
$$;
COMMENT ON FUNCTION public.catastro_geom_4326(geojson text) IS 'GeoJSON (Polygon o MultiPolygon, lon/lat) → MultiPolygon EPSG:4326. NULL si viene vacío.';
-- areas_verdes
CREATE TABLE public.areas_verdes (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    source_index integer NOT NULL,
    codigo text,
    nombre text,
    uso text,
    proy_riego text,
    riego_act text,
    referencia text,
    perimetro_m double precision,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    zona_supervision_id bigint,
    activo boolean DEFAULT true NOT NULL,
    origen_ref text,
    CONSTRAINT areas_verdes_area_chk CHECK (((area_m2 IS NULL) OR (area_m2 >= (0)::double precision))),
    CONSTRAINT areas_verdes_perimetro_chk CHECK (((perimetro_m IS NULL) OR (perimetro_m >= (0)::double precision))),
    CONSTRAINT areas_verdes_referencia_chk CHECK (((referencia IS NULL) OR (char_length(referencia) <= 500)))
);
COMMENT ON TABLE public.areas_verdes IS 'Áreas verdes de catastro. Semilla: data/raw/areas_verdes.geojson.';
COMMENT ON COLUMN public.areas_verdes.feature_id IS 'Identificador estable AV-NNNN, independiente del id serial.';
COMMENT ON COLUMN public.areas_verdes.geom IS 'MultiPolygon EPSG:4326. NULL = catastro progresivo.';
COMMENT ON COLUMN public.areas_verdes.zona_supervision_id IS 'Zona de supervisión por intersección. Editable.';
COMMENT ON COLUMN public.areas_verdes.origen_ref IS 'Identificador de la feature de catastro (AV-NNNN). etl-lote hace upsert por esta clave, sin TRUNCATE.';
CREATE SEQUENCE public.areas_verdes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.areas_verdes_id_seq OWNED BY public.areas_verdes.id;
ALTER TABLE ONLY public.areas_verdes ALTER COLUMN id SET DEFAULT nextval('public.areas_verdes_id_seq'::regclass);
ALTER TABLE ONLY public.areas_verdes
    ADD CONSTRAINT areas_verdes_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.areas_verdes
    ADD CONSTRAINT areas_verdes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.areas_verdes
    ADD CONSTRAINT areas_verdes_source_index_key UNIQUE (source_index);
CREATE UNIQUE INDEX areas_verdes_codigo_uidx ON public.areas_verdes USING btree (codigo) WHERE ((codigo IS NOT NULL) AND (btrim(codigo) <> ''::text));
CREATE INDEX areas_verdes_geom_gix ON public.areas_verdes USING gist (geom);
CREATE UNIQUE INDEX areas_verdes_origen_ref_uidx ON public.areas_verdes USING btree (origen_ref);
CREATE INDEX areas_verdes_zona_idx ON public.areas_verdes USING btree (zona_supervision_id);
-- zonas_supervision
CREATE TABLE public.zonas_supervision (
    id bigint NOT NULL,
    codigo text NOT NULL,
    nombre text NOT NULL,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    origen_ref text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT zonas_supervision_area_chk CHECK (((area_m2 IS NULL) OR (area_m2 >= (0)::double precision))),
    CONSTRAINT zonas_supervision_codigo_chk CHECK ((codigo = ANY (ARRAY['Z1'::text, 'Z2'::text, 'Z3'::text, 'Z4'::text])))
);
COMMENT ON TABLE public.zonas_supervision IS 'Cuatro zonas de supervisión del campus. No son los polígonos de cuadrilla.';
CREATE SEQUENCE public.zonas_supervision_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.zonas_supervision_id_seq OWNED BY public.zonas_supervision.id;
ALTER TABLE ONLY public.zonas_supervision ALTER COLUMN id SET DEFAULT nextval('public.zonas_supervision_id_seq'::regclass);
ALTER TABLE ONLY public.zonas_supervision
    ADD CONSTRAINT zonas_supervision_codigo_key UNIQUE (codigo);
ALTER TABLE ONLY public.zonas_supervision
    ADD CONSTRAINT zonas_supervision_pkey PRIMARY KEY (id);
CREATE INDEX zonas_supervision_geom_gix ON public.zonas_supervision USING gist (geom);
CREATE UNIQUE INDEX zonas_supervision_origen_ref_uidx ON public.zonas_supervision USING btree (origen_ref);
-- zonas_origen
CREATE TABLE public.zonas_origen (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    source_index integer NOT NULL,
    codigo text,
    nombre text,
    uso text,
    proy_riego text,
    riego_act text,
    referencia text,
    perimetro_m double precision,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.zonas_origen IS 'Copia de la tabla zonas previa al frente 1A. No se borra.';
COMMENT ON COLUMN public.zonas_origen.feature_id IS 'Identificador de zona Z-NNNN. No es un nombre de persona.';
COMMENT ON COLUMN public.zonas_origen.geom IS 'MultiPolygon EPSG:4326. NULL = geometría progresiva.';
CREATE SEQUENCE public.zonas_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.zonas_id_seq OWNED BY public.zonas_origen.id;
ALTER TABLE ONLY public.zonas_origen ALTER COLUMN id SET DEFAULT nextval('public.zonas_id_seq'::regclass);
ALTER TABLE ONLY public.zonas_origen
    ADD CONSTRAINT zonas_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.zonas_origen
    ADD CONSTRAINT zonas_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.zonas_origen
    ADD CONSTRAINT zonas_source_index_key UNIQUE (source_index);
CREATE INDEX zonas_geom_gix ON public.zonas_origen USING gist (geom);
-- lugares
CREATE TABLE public.lugares (
    id bigint NOT NULL,
    nombre text NOT NULL,
    nombre_norm text NOT NULL,
    lat double precision NOT NULL,
    lon double precision NOT NULL,
    zona_supervision_id bigint,
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT lugares_lat_chk CHECK (((lat >= ('-12.20'::numeric)::double precision) AND (lat <= ('-11.90'::numeric)::double precision))),
    CONSTRAINT lugares_lon_chk CHECK (((lon >= ('-77.30'::numeric)::double precision) AND (lon <= ('-76.90'::numeric)::double precision)))
);
COMMENT ON TABLE public.lugares IS 'Diccionario de lugares. nombre_norm es único, en minúsculas y sin tildes.';
COMMENT ON COLUMN public.lugares.nombre_norm IS 'Nombre normalizado para no duplicar el mismo lugar con otra grafía.';
CREATE SEQUENCE public.lugares_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.lugares_id_seq OWNED BY public.lugares.id;
ALTER TABLE ONLY public.lugares ALTER COLUMN id SET DEFAULT nextval('public.lugares_id_seq'::regclass);
ALTER TABLE ONLY public.lugares
    ADD CONSTRAINT lugares_nombre_norm_key UNIQUE (nombre_norm);
ALTER TABLE ONLY public.lugares
    ADD CONSTRAINT lugares_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX lugares_origen_ref_uidx ON public.lugares USING btree (origen_ref);
CREATE INDEX lugares_zona_idx ON public.lugares USING btree (zona_supervision_id);
-- cuadrillas
CREATE TABLE public.cuadrillas (
    id text NOT NULL,
    nombre_ficticio text NOT NULL,
    turno text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cuadrillas_turno_chk CHECK ((turno = ANY (ARRAY['manana'::text, 'tarde'::text])))
);
COMMENT ON TABLE public.cuadrillas IS 'Responsables ficticios. No hay personas reales.';
ALTER TABLE ONLY public.cuadrillas
    ADD CONSTRAINT cuadrillas_pkey PRIMARY KEY (id);
-- poligonos_cuadrilla
CREATE TABLE public.poligonos_cuadrilla (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    source_index integer NOT NULL,
    codigo text,
    nombre text,
    uso text,
    proy_riego text,
    riego_act text,
    referencia text,
    perimetro_m double precision,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326) NOT NULL,
    cuadrilla_id text,
    zona_supervision_id bigint,
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    sector text,
    CONSTRAINT poligonos_area_chk CHECK (((area_m2 IS NULL) OR (area_m2 >= (0)::double precision))),
    CONSTRAINT poligonos_perimetro_chk CHECK (((perimetro_m IS NULL) OR (perimetro_m >= (0)::double precision))),
    CONSTRAINT poligonos_referencia_chk CHECK (((referencia IS NULL) OR (char_length(referencia) <= 500)))
);
COMMENT ON TABLE public.poligonos_cuadrilla IS 'Polígono de cuadrilla (antes zonas). feature_id histórico Z- se conserva. Los nuevos usan PC-.';
COMMENT ON COLUMN public.poligonos_cuadrilla.cuadrilla_id IS 'Cuadrilla ficticia. Sustituye al campo jefes, que no se persiste.';
COMMENT ON COLUMN public.poligonos_cuadrilla.sector IS 'Sector operativo ficticio o rótulo de lugar. NULL = sin sector. No identifica personas.';
CREATE SEQUENCE public.poligonos_cuadrilla_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.poligonos_cuadrilla_id_seq OWNED BY public.poligonos_cuadrilla.id;
ALTER TABLE ONLY public.poligonos_cuadrilla ALTER COLUMN id SET DEFAULT nextval('public.poligonos_cuadrilla_id_seq'::regclass);
ALTER TABLE ONLY public.poligonos_cuadrilla
    ADD CONSTRAINT poligonos_cuadrilla_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.poligonos_cuadrilla
    ADD CONSTRAINT poligonos_cuadrilla_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.poligonos_cuadrilla
    ADD CONSTRAINT poligonos_cuadrilla_source_index_key UNIQUE (source_index);
CREATE INDEX poligonos_cuadrilla_cuadrilla_idx ON public.poligonos_cuadrilla USING btree (cuadrilla_id);
CREATE INDEX poligonos_cuadrilla_geom_gix ON public.poligonos_cuadrilla USING gist (geom);
CREATE UNIQUE INDEX poligonos_cuadrilla_origen_ref_uidx ON public.poligonos_cuadrilla USING btree (origen_ref);
CREATE INDEX poligonos_cuadrilla_sector_idx ON public.poligonos_cuadrilla USING btree (sector);
-- poligonos_sector_ref
CREATE TABLE public.poligonos_sector_ref (
    source_index integer NOT NULL,
    sector text NOT NULL,
    CONSTRAINT poligonos_sector_ref_chk CHECK ((sector = ANY (ARRAY['cua-valeria'::text, 'cua-mateo'::text, 'cua-renato'::text, 'campo-deportivo'::text, 'bosque-humedo'::text])))
);
COMMENT ON TABLE public.poligonos_sector_ref IS 'Sector por source_index de la fuente de polígonos. Sin FK: el TRUNCATE del ETL no la vacía.';
ALTER TABLE ONLY public.poligonos_sector_ref
    ADD CONSTRAINT poligonos_sector_ref_pkey PRIMARY KEY (source_index);
-- asignaciones_poligono
CREATE TABLE public.asignaciones_poligono (
    id bigint NOT NULL,
    poligono_id bigint NOT NULL,
    cuadrilla_id text NOT NULL,
    vigente boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.asignaciones_poligono IS 'Asignación de un polígono a una cuadrilla ficticia.';
CREATE SEQUENCE public.asignaciones_poligono_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.asignaciones_poligono_id_seq OWNED BY public.asignaciones_poligono.id;
ALTER TABLE ONLY public.asignaciones_poligono ALTER COLUMN id SET DEFAULT nextval('public.asignaciones_poligono_id_seq'::regclass);
ALTER TABLE ONLY public.asignaciones_poligono
    ADD CONSTRAINT asignaciones_poligono_pkey PRIMARY KEY (id);
CREATE INDEX asignaciones_poligono_poligono_idx ON public.asignaciones_poligono USING btree (poligono_id);
-- sectores_capataz
CREATE TABLE public.sectores_capataz (
    id bigint NOT NULL,
    codigo text NOT NULL,
    nombre text NOT NULL,
    color text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.sectores_capataz IS 'Catálogo editable del sector de capataz. El color del mapa sale de aquí. activo = false es la baja y no se borra la fila.';
COMMENT ON COLUMN public.sectores_capataz.color IS 'Hexadecimal #rrggbb que pinta el polígono. No es una lista cerrada en código.';
CREATE SEQUENCE public.sectores_capataz_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.sectores_capataz_id_seq OWNED BY public.sectores_capataz.id;
ALTER TABLE ONLY public.sectores_capataz ALTER COLUMN id SET DEFAULT nextval('public.sectores_capataz_id_seq'::regclass);
ALTER TABLE ONLY public.sectores_capataz
    ADD CONSTRAINT sectores_capataz_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX sectores_capataz_codigo_uidx ON public.sectores_capataz USING btree (codigo);
-- vias
CREATE TABLE public.vias (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    nombre text,
    geom public.geometry(Geometry,4326),
    activo boolean DEFAULT true NOT NULL,
    origen_ref text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.vias IS 'Referente lineal. Nace vacía. La capa del mapa permanece apagada hasta que se importe un GeoJSON.';
CREATE SEQUENCE public.vias_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.vias_id_seq OWNED BY public.vias.id;
ALTER TABLE ONLY public.vias ALTER COLUMN id SET DEFAULT nextval('public.vias_id_seq'::regclass);
ALTER TABLE ONLY public.vias
    ADD CONSTRAINT vias_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX vias_feature_id_uidx ON public.vias USING btree (feature_id);
CREATE INDEX vias_geom_gix ON public.vias USING gist (geom);
-- cuarteles_historico
CREATE TABLE public.cuarteles_historico (
    id bigint NOT NULL,
    codigo text NOT NULL,
    nombre text NOT NULL,
    geom public.geometry(MultiPolygon,4326),
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.cuarteles_historico IS 'Referencia histórica de solo lectura. geom puede ser NULL. Sin archivo de cuarteles no hay filas.';
COMMENT ON COLUMN public.cuarteles_historico.geom IS 'Polígono del shape del cliente, si llega. Nunca se rellena con una geometría de ejemplo.';
CREATE SEQUENCE public.cuarteles_historico_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.cuarteles_historico_id_seq OWNED BY public.cuarteles_historico.id;
ALTER TABLE ONLY public.cuarteles_historico ALTER COLUMN id SET DEFAULT nextval('public.cuarteles_historico_id_seq'::regclass);
ALTER TABLE ONLY public.cuarteles_historico
    ADD CONSTRAINT cuarteles_historico_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX cuarteles_historico_codigo_uidx ON public.cuarteles_historico USING btree (codigo);
-- referentes_edificio
CREATE TABLE public.referentes_edificio (
    id bigint NOT NULL,
    lugar_id bigint NOT NULL,
    edificio_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.referentes_edificio IS 'Par lugar de catálogo + id de edificio. No acepta un nombre de edificio escrito a mano.';
COMMENT ON COLUMN public.referentes_edificio.edificio_id IS 'id del GeoJSON de edificios (por ejemplo osm-way-…). No es un texto libre.';
CREATE SEQUENCE public.referentes_edificio_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.referentes_edificio_id_seq OWNED BY public.referentes_edificio.id;
ALTER TABLE ONLY public.referentes_edificio ALTER COLUMN id SET DEFAULT nextval('public.referentes_edificio_id_seq'::regclass);
ALTER TABLE ONLY public.referentes_edificio
    ADD CONSTRAINT referentes_edificio_par_uidx UNIQUE (lugar_id, edificio_id);
ALTER TABLE ONLY public.referentes_edificio
    ADD CONSTRAINT referentes_edificio_pkey PRIMARY KEY (id);
CREATE INDEX referentes_edificio_lugar_idx ON public.referentes_edificio USING btree (lugar_id);
-- capas_auxiliares
CREATE TABLE public.capas_auxiliares (
    id bigint NOT NULL,
    capa text NOT NULL,
    feature_id text NOT NULL,
    source_index integer NOT NULL,
    codigo text,
    nombre text,
    uso text,
    proy_riego text,
    clase text,
    riego_act text,
    referencia text,
    pertenecen text,
    perimetro_m double precision,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.capas_auxiliares IS 'Capas opcionales de catastro: jardines_reserva y xerofitica.';
COMMENT ON COLUMN public.capas_auxiliares.pertenecen IS 'Unidad organizacional (p. ej. DAF), no una persona.';
CREATE SEQUENCE public.capas_auxiliares_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.capas_auxiliares_id_seq OWNED BY public.capas_auxiliares.id;
ALTER TABLE ONLY public.capas_auxiliares ALTER COLUMN id SET DEFAULT nextval('public.capas_auxiliares_id_seq'::regclass);
ALTER TABLE ONLY public.capas_auxiliares
    ADD CONSTRAINT capas_auxiliares_capa_source_unique UNIQUE (capa, source_index);
ALTER TABLE ONLY public.capas_auxiliares
    ADD CONSTRAINT capas_auxiliares_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.capas_auxiliares
    ADD CONSTRAINT capas_auxiliares_pkey PRIMARY KEY (id);
CREATE INDEX capas_auxiliares_capa_idx ON public.capas_auxiliares USING btree (capa);
CREATE INDEX capas_auxiliares_geom_gix ON public.capas_auxiliares USING gist (geom);
CREATE VIEW public.zonas AS
 SELECT id,
    feature_id,
    source_index,
    codigo,
    nombre,
    uso,
    proy_riego,
    riego_act,
    referencia,
    perimetro_m,
    area_m2,
    geom,
    created_at,
    updated_at,
    sector
   FROM public.poligonos_cuadrilla;
COMMENT ON VIEW public.zonas IS 'Compatibilidad de lectura. Los polígonos viven en poligonos_cuadrilla.';
-- ========================================================================
-- Operación
-- ========================================================================
-- capataces
CREATE TABLE public.capataces (
    id text NOT NULL,
    equipo text NOT NULL,
    turno text NOT NULL,
    activo boolean DEFAULT true NOT NULL
);
COMMENT ON TABLE public.capataces IS 'Equipos de campo de demostración. No identifican a una persona.';
ALTER TABLE ONLY public.capataces
    ADD CONSTRAINT capataces_pkey PRIMARY KEY (id);
-- actividades
CREATE TABLE public.actividades (
    id uuid NOT NULL,
    tipo text NOT NULL,
    estado text NOT NULL,
    titulo text NOT NULL,
    detalle text DEFAULT ''::text NOT NULL,
    area_feature_id text,
    zona_feature_id text,
    assigned_capataz_id text,
    geom public.geometry(Point,4326),
    archivada_en timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    ejecutor text DEFAULT 'propia'::text NOT NULL,
    motivo_archivo text,
    lugar_id bigint,
    zona_supervision_id bigint,
    fecha_solicitud date,
    fecha_atencion date,
    cuadrilla_id text,
    clase_codigo text,
    tipo_codigo text,
    comentario text DEFAULT ''::text NOT NULL,
    lugar_libre text DEFAULT ''::text NOT NULL,
    origen_ref text,
    origen text DEFAULT 'interna'::text NOT NULL,
    subtipo text,
    codigo_externo text,
    unidad_solicitante text,
    nivel_riesgo text,
    fecha_programada date,
    cantidad numeric,
    CONSTRAINT actividades_ejecutor_chk CHECK ((ejecutor = ANY (ARRAY['propia'::text, 'tercerizada'::text]))),
    CONSTRAINT actividades_fechas_chk CHECK (((fecha_solicitud IS NULL) OR (fecha_atencion IS NULL) OR (fecha_atencion >= fecha_solicitud))),
    CONSTRAINT actividades_ubicacion_chk CHECK (((geom IS NOT NULL) OR (lugar_id IS NOT NULL) OR (zona_supervision_id IS NOT NULL) OR ((origen_ref IS NOT NULL) AND (origen_ref <> ''::text))))
);
COMMENT ON TABLE public.actividades IS 'Labores de supervisión. geom es Point EPSG:4326. archivada_en es la baja lógica.';
COMMENT ON COLUMN public.actividades.estado IS 'Incluye sin_estado cuando la hoja de monitoreo trae el estado vacío.';
COMMENT ON COLUMN public.actividades.assigned_capataz_id IS 'Equipo stub (capataces.id), no un usuario institucional.';
COMMENT ON COLUMN public.actividades.geom IS 'Point EPSG:4326. NULL solo si hay lugar_id o zona_supervision_id.';
COMMENT ON COLUMN public.actividades.lugar_id IS 'BIGINT con FK a lugares (016). No es TEXT.';
COMMENT ON COLUMN public.actividades.zona_supervision_id IS 'BIGINT con FK a zonas_supervision (016). No es TEXT.';
COMMENT ON COLUMN public.actividades.origen_ref IS 'Id de la fila fuente para reimportar. No guarda datos personales.';
COMMENT ON COLUMN public.actividades.subtipo IS 'Segundo nivel de la actividad, código del catálogo subtipo_actividad. Nulo si el alta no lo trae.';
COMMENT ON COLUMN public.actividades.codigo_externo IS 'Código de Centuria u OSG si ya viene. El sistema no lo inventa.';
COMMENT ON COLUMN public.actividades.unidad_solicitante IS 'Unidad que pide la actividad. Nulo si no se informa.';
COMMENT ON COLUMN public.actividades.nivel_riesgo IS 'Código de catalogos clase nivel_riesgo. Sin CHECK: un valor nuevo se agrega al catálogo.';
COMMENT ON COLUMN public.actividades.fecha_programada IS 'Fecha pedida para la actividad. No es la fecha de atención de la ficha.';
COMMENT ON COLUMN public.actividades.cantidad IS 'Cantidad pedida. Nula si el alta no la trae.';
ALTER TABLE ONLY public.actividades
    ADD CONSTRAINT actividades_pkey PRIMARY KEY (id);
CREATE INDEX actividades_asignacion_idx ON public.actividades USING btree (assigned_capataz_id);
CREATE INDEX actividades_geom_gix ON public.actividades USING gist (geom);
CREATE INDEX actividades_lugar_idx ON public.actividades USING btree (lugar_id);
CREATE UNIQUE INDEX actividades_origen_ref_uidx ON public.actividades USING btree (origen_ref) WHERE ((origen_ref IS NOT NULL) AND (origen_ref <> ''::text));
CREATE INDEX actividades_zona_sup_idx ON public.actividades USING btree (zona_supervision_id);
-- actividad_eventos
CREATE TABLE public.actividad_eventos (
    id bigint NOT NULL,
    actividad_id uuid NOT NULL,
    tipo text NOT NULL,
    estado text,
    capataz_id text,
    actor_rol text NOT NULL,
    nota text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    usuario_id bigint,
    uuid_cliente uuid,
    capataz_anterior text,
    CONSTRAINT actividad_eventos_tipo_chk CHECK ((tipo = ANY (ARRAY['creada'::text, 'asignada'::text, 'reasignada'::text, 'estado'::text, 'cancelada'::text, 'archivada'::text, 'evidencia'::text, 'inicio'::text, 'supervision'::text, 'derivacion'::text, 'observacion'::text, 'conformidad'::text, 'avance'::text])))
);
COMMENT ON CONSTRAINT actividad_eventos_tipo_chk ON public.actividad_eventos IS 'Cadena de la actividad. Los siete tipos originales siguen valiendo y se suman los hitos del libro y el avance.';
COMMENT ON COLUMN public.actividad_eventos.usuario_id IS 'Usuario de sesión que registró el evento. El timeline muestra esta cuenta, no actor_rol.';
COMMENT ON COLUMN public.actividad_eventos.uuid_cliente IS 'Identificador generado por el cliente para reintentos offline. Nulo en eventos anteriores.';
COMMENT ON COLUMN public.actividad_eventos.capataz_anterior IS 'Cuadrilla responsable antes de una reasignación. La nueva queda en capataz_id. Nulo en el resto de hitos.';
CREATE SEQUENCE public.actividad_eventos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.actividad_eventos_id_seq OWNED BY public.actividad_eventos.id;
ALTER TABLE ONLY public.actividad_eventos ALTER COLUMN id SET DEFAULT nextval('public.actividad_eventos_id_seq'::regclass);
ALTER TABLE ONLY public.actividad_eventos
    ADD CONSTRAINT actividad_eventos_pkey PRIMARY KEY (id);
CREATE INDEX actividad_eventos_act_idx ON public.actividad_eventos USING btree (actividad_id, id);
CREATE INDEX actividad_eventos_capataz_anterior_idx ON public.actividad_eventos USING btree (capataz_anterior) WHERE (capataz_anterior IS NOT NULL);
CREATE INDEX actividad_eventos_usuario_idx ON public.actividad_eventos USING btree (usuario_id) WHERE (usuario_id IS NOT NULL);
CREATE UNIQUE INDEX actividad_eventos_uuid_cliente_uidx ON public.actividad_eventos USING btree (uuid_cliente) WHERE (uuid_cliente IS NOT NULL);
-- actividad_avances
CREATE TABLE public.actividad_avances (
    id uuid NOT NULL,
    actividad_id uuid NOT NULL,
    fecha date NOT NULL,
    nota text DEFAULT ''::text NOT NULL,
    area_feature_id text,
    ejemplar_ref text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.actividad_avances IS 'Avance de varios días ligado al área o al ejemplar. No hay borrado físico.';
ALTER TABLE ONLY public.actividad_avances
    ADD CONSTRAINT actividad_avances_pkey PRIMARY KEY (id);
CREATE INDEX actividad_avances_act_idx ON public.actividad_avances USING btree (actividad_id, fecha);
-- personal_labor
CREATE TABLE public.personal_labor (
    id uuid NOT NULL,
    actividad_id uuid NOT NULL,
    nombre_ficticio text NOT NULL,
    rol_campo text DEFAULT 'operario de cuadrilla'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.personal_labor IS 'Participantes de la labor. Solo nombres ficticios.';
ALTER TABLE ONLY public.personal_labor
    ADD CONSTRAINT personal_labor_pkey PRIMARY KEY (id);
CREATE INDEX personal_labor_actividad_idx ON public.personal_labor USING btree (actividad_id);
CREATE UNIQUE INDEX personal_labor_nombre_uidx ON public.personal_labor USING btree (actividad_id, nombre_ficticio);
-- personal_ficticio
CREATE TABLE public.personal_ficticio (
    id text NOT NULL,
    nombre_ficticio text NOT NULL,
    activo boolean DEFAULT true NOT NULL
);
COMMENT ON TABLE public.personal_ficticio IS 'Nombres ficticios que se pueden asignar a una actividad. No son cuentas ni personas reales.';
ALTER TABLE ONLY public.personal_ficticio
    ADD CONSTRAINT personal_ficticio_pkey PRIMARY KEY (id);
-- podas
CREATE TABLE public.podas (
    id uuid NOT NULL,
    codigo text NOT NULL,
    codigo_externo text,
    tipo text DEFAULT ''::text NOT NULL,
    tipo_actividad text DEFAULT ''::text NOT NULL,
    fecha_reporte date,
    fecha_ejecucion date,
    personal_ficticio text DEFAULT ''::text NOT NULL,
    ubicacion text DEFAULT ''::text NOT NULL,
    lugar_id text,
    unidad text DEFAULT ''::text NOT NULL,
    cantidad_pedida numeric DEFAULT 0 NOT NULL,
    cantidad_ejecutada numeric DEFAULT 0 NOT NULL,
    tipo_vegetacion text DEFAULT ''::text NOT NULL,
    nombre_comun text DEFAULT ''::text NOT NULL,
    nombre_cientifico text DEFAULT ''::text NOT NULL,
    especie_id text,
    prioridad text DEFAULT 'media'::text NOT NULL,
    comentario text DEFAULT ''::text NOT NULL,
    solicitud_id uuid,
    actividad_id uuid,
    origen_ref text,
    archivada_en timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT podas_cantidades_chk CHECK (((cantidad_pedida >= (0)::numeric) AND (cantidad_ejecutada >= (0)::numeric))),
    CONSTRAINT podas_fechas_chk CHECK (((fecha_reporte IS NULL) OR (fecha_ejecucion IS NULL) OR (fecha_ejecucion >= fecha_reporte))),
    CONSTRAINT podas_prioridad_chk CHECK ((prioridad = ANY (ARRAY['baja'::text, 'media'::text, 'alta'::text])))
);
COMMENT ON TABLE public.podas IS 'Registros de poda. La baja es lógica (archivada_en).';
COMMENT ON COLUMN public.podas.codigo_externo IS 'OSG-… solo si la fuente lo trae. El sistema no lo genera.';
ALTER TABLE ONLY public.podas
    ADD CONSTRAINT podas_codigo_key UNIQUE (codigo);
ALTER TABLE ONLY public.podas
    ADD CONSTRAINT podas_pkey PRIMARY KEY (id);
CREATE INDEX podas_codigo_externo_idx ON public.podas USING btree (codigo_externo) WHERE ((codigo_externo IS NOT NULL) AND (codigo_externo <> ''::text));
CREATE UNIQUE INDEX podas_origen_ref_uidx ON public.podas USING btree (origen_ref) WHERE ((origen_ref IS NOT NULL) AND (origen_ref <> ''::text));
-- vivero_catalogo
CREATE TABLE public.vivero_catalogo (
    id bigint NOT NULL,
    clase text NOT NULL,
    nombre text NOT NULL,
    activo boolean DEFAULT true NOT NULL
);
COMMENT ON TABLE public.vivero_catalogo IS 'Área, subproceso y etapa. Después de importar no se escribe texto libre.';
CREATE SEQUENCE public.vivero_catalogo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.vivero_catalogo_id_seq OWNED BY public.vivero_catalogo.id;
ALTER TABLE ONLY public.vivero_catalogo ALTER COLUMN id SET DEFAULT nextval('public.vivero_catalogo_id_seq'::regclass);
ALTER TABLE ONLY public.vivero_catalogo
    ADD CONSTRAINT vivero_catalogo_clase_nombre_key UNIQUE (clase, nombre);
ALTER TABLE ONLY public.vivero_catalogo
    ADD CONSTRAINT vivero_catalogo_pkey PRIMARY KEY (id);
-- vivero_registros
CREATE TABLE public.vivero_registros (
    id uuid NOT NULL,
    fecha date,
    area text DEFAULT ''::text NOT NULL,
    subproceso text DEFAULT ''::text NOT NULL,
    etapa text DEFAULT ''::text NOT NULL,
    descripcion text DEFAULT ''::text NOT NULL,
    observaciones text DEFAULT ''::text NOT NULL,
    responsables text DEFAULT ''::text NOT NULL,
    lugar_id text,
    lugar_libre text DEFAULT ''::text NOT NULL,
    origen_ref text,
    archivada_en timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON COLUMN public.vivero_registros.responsables IS 'Lista de nombres ficticios. No hay personas reales.';
COMMENT ON COLUMN public.vivero_registros.lugar_libre IS 'Texto de lugar cuando el nombre no calza con un lugar conocido.';
ALTER TABLE ONLY public.vivero_registros
    ADD CONSTRAINT vivero_registros_pkey PRIMARY KEY (id);
CREATE INDEX vivero_fecha_idx ON public.vivero_registros USING btree (fecha);
CREATE UNIQUE INDEX vivero_origen_ref_uidx ON public.vivero_registros USING btree (origen_ref) WHERE ((origen_ref IS NOT NULL) AND (origen_ref <> ''::text));
-- riego_registros
CREATE TABLE public.riego_registros (
    id uuid NOT NULL,
    sector text NOT NULL,
    turno text NOT NULL,
    capataz_id text,
    fecha date NOT NULL,
    nota text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    zona_supervision_id bigint,
    ciclo text DEFAULT ''::text NOT NULL,
    superficie_m2 numeric,
    sector_id bigint,
    CONSTRAINT riego_superficie_chk CHECK (((superficie_m2 IS NULL) OR (superficie_m2 >= (0)::numeric)))
);
COMMENT ON TABLE public.riego_registros IS 'Cobertura mínima de riego. No calcula un indicador oficial.';
COMMENT ON COLUMN public.riego_registros.zona_supervision_id IS 'Sector de supervisión. El alta nueva exige esta FK. El indicador oficial queda pendiente.';
COMMENT ON COLUMN public.riego_registros.sector_id IS 'Sector de capataz del catálogo. Nulo en registros anteriores, que conservan solo el texto de sector.';
ALTER TABLE ONLY public.riego_registros
    ADD CONSTRAINT riego_registros_pkey PRIMARY KEY (id);
CREATE INDEX riego_registros_sector_id_idx ON public.riego_registros USING btree (sector_id);
CREATE UNIQUE INDEX riego_zona_turno_fecha_uidx ON public.riego_registros USING btree (zona_supervision_id, turno, fecha) WHERE (zona_supervision_id IS NOT NULL);
-- solicitudes
CREATE TABLE public.solicitudes (
    id uuid NOT NULL,
    codigo_externo text,
    fuente text NOT NULL,
    titulo text NOT NULL,
    detalle text DEFAULT ''::text NOT NULL,
    prioridad text DEFAULT 'media'::text NOT NULL,
    estado text DEFAULT 'por_iniciar'::text NOT NULL,
    lugar text,
    cantidad integer,
    actividad_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    origen_ref text,
    archivada_en timestamp with time zone,
    cantidad_solicitada integer,
    cantidad_ejecutada integer,
    lugar_id bigint,
    lat double precision,
    lon double precision,
    CONSTRAINT solicitudes_estado_chk CHECK ((estado = ANY (ARRAY['por_iniciar'::text, 'en_proceso'::text, 'ejecutado'::text, 'cerrado'::text, 'cancelado'::text]))),
    CONSTRAINT solicitudes_fuente_chk CHECK ((fuente = ANY (ARRAY['centuria'::text, 'osg'::text, 'correo'::text, 'interna'::text]))),
    CONSTRAINT solicitudes_punto_par_chk CHECK ((((lat IS NULL) AND (lon IS NULL)) OR ((lat IS NOT NULL) AND (lon IS NOT NULL))))
);
COMMENT ON CONSTRAINT solicitudes_estado_chk ON public.solicitudes IS 'Se deja hasta que el catálogo tenga exactamente estos códigos. La API valida contra catalogos (clase estado_solicitud).';
COMMENT ON CONSTRAINT solicitudes_fuente_chk ON public.solicitudes IS 'Se deja hasta que el catálogo fuente tenga exactamente estos códigos. La API ya valida contra catalogos (clase fuente).';
COMMENT ON COLUMN public.solicitudes.codigo_externo IS 'Código de Centuria u OSG si la fuente lo trae. El sistema no lo inventa ni llama a Centuria.';
COMMENT ON COLUMN public.solicitudes.lugar IS 'Texto de lugar ya cargado. Se conserva y se muestra. El alta nueva usa lugar_id o el punto.';
COMMENT ON COLUMN public.solicitudes.cantidad IS 'Cantidad histórica. No se borra. cantidad_solicitada la copia si aún no tenía valor.';
COMMENT ON COLUMN public.solicitudes.cantidad_solicitada IS 'Cantidad pedida. Puede diferir de cantidad_ejecutada.';
COMMENT ON COLUMN public.solicitudes.cantidad_ejecutada IS 'Cantidad atendida. Nula si todavía no hay ejecución. Puede diferir de la pedida.';
COMMENT ON COLUMN public.solicitudes.lugar_id IS 'Lugar del catálogo. Nulo si la ubicación es un punto o solo texto ya cargado.';
COMMENT ON COLUMN public.solicitudes.lat IS 'Latitud del pin. Nula si no hay punto. Va junto con lon.';
COMMENT ON COLUMN public.solicitudes.lon IS 'Longitud del pin. Nula si no hay punto. Va junto con lat.';
ALTER TABLE ONLY public.solicitudes
    ADD CONSTRAINT solicitudes_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX solicitudes_codigo_externo_uidx ON public.solicitudes USING btree (codigo_externo) WHERE ((codigo_externo IS NOT NULL) AND (codigo_externo <> ''::text));
CREATE INDEX solicitudes_lugar_id_idx ON public.solicitudes USING btree (lugar_id);
CREATE UNIQUE INDEX solicitudes_origen_ref_uidx ON public.solicitudes USING btree (origen_ref) WHERE ((origen_ref IS NOT NULL) AND (origen_ref <> ''::text));
-- ordenes_servicio
CREATE TABLE public.ordenes_servicio (
    id uuid NOT NULL,
    actividad_id uuid NOT NULL,
    empresa text NOT NULL,
    referencia text NOT NULL,
    frecuencia text DEFAULT ''::text NOT NULL,
    estado text DEFAULT 'en_proceso'::text NOT NULL,
    conformidad text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    periodo_inicio date,
    periodo_fin date,
    reporte_proveedor text DEFAULT ''::text NOT NULL,
    empresa_id bigint,
    frecuencia_id bigint,
    CONSTRAINT ordenes_estado_chk CHECK ((estado = ANY (ARRAY['en_proceso'::text, 'ejecutada'::text, 'conforme'::text]))),
    CONSTRAINT ordenes_periodo_chk CHECK (((periodo_inicio IS NULL) OR (periodo_fin IS NULL) OR (periodo_fin >= periodo_inicio)))
);
COMMENT ON TABLE public.ordenes_servicio IS 'La orden no cierra la labor ni la solicitud. El cierre de una labor tercerizada exige que la orden exista, y además una ejecución registrada.';
COMMENT ON COLUMN public.ordenes_servicio.empresa IS 'Texto ya cargado o nombre del catálogo al crear. No es la fuente de las altas nuevas: esa es empresa_id.';
COMMENT ON COLUMN public.ordenes_servicio.frecuencia IS 'Texto ya cargado o nombre del catálogo al crear. No es la fuente de las altas nuevas: esa es frecuencia_id.';
COMMENT ON COLUMN public.ordenes_servicio.conformidad IS 'Conformidad de la orden. Se sigue guardando. No cierra la solicitud ni adjunta el reporte del proveedor.';
COMMENT ON COLUMN public.ordenes_servicio.empresa_id IS 'FK a catalogos clase empresa. Nula en filas anteriores al catálogo.';
COMMENT ON COLUMN public.ordenes_servicio.frecuencia_id IS 'FK a catalogos clase frecuencia. Nula en filas anteriores al catálogo.';
ALTER TABLE ONLY public.ordenes_servicio
    ADD CONSTRAINT ordenes_servicio_pkey PRIMARY KEY (id);
CREATE INDEX ordenes_empresa_id_idx ON public.ordenes_servicio USING btree (empresa_id);
CREATE INDEX ordenes_frecuencia_id_idx ON public.ordenes_servicio USING btree (frecuencia_id);
-- ========================================================================
-- Inventario
-- ========================================================================
CREATE FUNCTION public.ejemplares_sector_cuartel_clase() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF NEW.sector_cuartel_id IS NULL THEN
    RETURN NEW;
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM catalogos c
    WHERE c.id = NEW.sector_cuartel_id
      AND c.activo
      AND c.clase IN ('sector_capataz', 'cuartel')
  ) THEN
    RAISE EXCEPTION 'sector o cuartel desconocido'
      USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END;
$$;
COMMENT ON FUNCTION public.ejemplares_sector_cuartel_clase() IS 'Rechaza un sector_cuartel_id que no sea sector de capataz o cuartel activo.';
CREATE FUNCTION public.inventario_geom_4326(geojson text) RETURNS public.geometry
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
  g geometry;
BEGIN
  IF geojson IS NULL OR btrim(geojson) = '' OR lower(btrim(geojson)) = 'null' THEN
    RETURN NULL;
  END IF;

  g := ST_MakeValid(ST_Force2D(ST_SetSRID(ST_GeomFromGeoJSON(geojson), 4326)));
  IF g IS NULL OR ST_IsEmpty(g) THEN
    RETURN NULL;
  END IF;
  RETURN g;
END;
$$;
COMMENT ON FUNCTION public.inventario_geom_4326(geojson text) IS 'GeoJSON (punto o polígono, lon/lat, con o sin Z) → geometría EPSG:4326.';


SET default_table_access_method = heap;
-- inventario
CREATE TABLE public.inventario (
    id bigint NOT NULL,
    capa text NOT NULL,
    feature_id text NOT NULL,
    nombre text,
    subtipo text,
    detalle text,
    lugar text,
    foto text,
    geom public.geometry(Geometry,4326)
);
COMMENT ON TABLE public.inventario IS 'Overlays legacy: bebederos, fauna, playas, puertas, vereda, flora, cafetos, tachos.';
COMMENT ON COLUMN public.inventario.foto IS 'Nombre de archivo JPEG local si existe bajo data/raw/drive_fotos. Nunca un id de Drive.';
CREATE SEQUENCE public.inventario_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.inventario_id_seq OWNED BY public.inventario.id;
ALTER TABLE ONLY public.inventario ALTER COLUMN id SET DEFAULT nextval('public.inventario_id_seq'::regclass);
ALTER TABLE ONLY public.inventario
    ADD CONSTRAINT inventario_capa_feature_id_key UNIQUE (capa, feature_id);
ALTER TABLE ONLY public.inventario
    ADD CONSTRAINT inventario_pkey PRIMARY KEY (id);
CREATE INDEX inventario_capa_idx ON public.inventario USING btree (capa);
CREATE INDEX inventario_geom_gix ON public.inventario USING gist (geom);
-- especies
CREATE TABLE public.especies (
    id bigint NOT NULL,
    nombre_cientifico text NOT NULL,
    nombre_comun text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.especies IS 'Especie de un ejemplar. Se da de alta cuando el nombre científico es nuevo.';
CREATE SEQUENCE public.especies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.especies_id_seq OWNED BY public.especies.id;
ALTER TABLE ONLY public.especies ALTER COLUMN id SET DEFAULT nextval('public.especies_id_seq'::regclass);
ALTER TABLE ONLY public.especies
    ADD CONSTRAINT especies_nombre_cientifico_key UNIQUE (nombre_cientifico);
ALTER TABLE ONLY public.especies
    ADD CONSTRAINT especies_pkey PRIMARY KEY (id);
-- ejemplares
CREATE TABLE public.ejemplares (
    id bigint NOT NULL,
    numero_origen integer,
    codigo text,
    especie_id bigint,
    nombre_comun text,
    tipo_vegetacion text,
    cantidad integer DEFAULT 1 NOT NULL,
    area_verde_id bigint,
    ubicacion_lugar_id bigint,
    referencia text,
    lat double precision,
    lon double precision,
    observacion_fen_2026 text,
    foto_id uuid,
    salud text,
    geom public.geometry(Point,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    sector_cuartel_id bigint,
    CONSTRAINT ejemplares_cantidad_chk CHECK ((cantidad >= 1)),
    CONSTRAINT ejemplares_lat_chk CHECK (((lat IS NULL) OR ((lat >= ('-12.20'::numeric)::double precision) AND (lat <= ('-11.90'::numeric)::double precision)))),
    CONSTRAINT ejemplares_lon_chk CHECK (((lon IS NULL) OR ((lon >= ('-77.30'::numeric)::double precision) AND (lon <= ('-76.90'::numeric)::double precision)))),
    CONSTRAINT ejemplares_tipo_chk CHECK (((tipo_vegetacion IS NULL) OR (tipo_vegetacion = ANY (ARRAY['Árbol'::text, 'Palmera'::text, 'Arbusto'::text, 'Herbácea'::text, 'Trepadora'::text, 'Suculenta'::text, 'cafeto'::text]))))
);
COMMENT ON TABLE public.ejemplares IS 'Ejemplar de flora. salud permanece NULL: la fuente no trae ese dato.';
COMMENT ON COLUMN public.ejemplares.numero_origen IS 'N° de la hoja de origen. Duplicado = el mismo número.';
COMMENT ON COLUMN public.ejemplares.foto_id IS 'Archivo propio. No guarda un id de Drive.';
COMMENT ON COLUMN public.ejemplares.sector_cuartel_id IS 'Catálogo de sector de capataz o de cuartel. NULL hasta que alguien lo asigne. No es texto libre.';
CREATE SEQUENCE public.ejemplares_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.ejemplares_id_seq OWNED BY public.ejemplares.id;
ALTER TABLE ONLY public.ejemplares ALTER COLUMN id SET DEFAULT nextval('public.ejemplares_id_seq'::regclass);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_numero_origen_key UNIQUE (numero_origen);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_pkey PRIMARY KEY (id);
CREATE INDEX ejemplares_especie_idx ON public.ejemplares USING btree (especie_id);
CREATE INDEX ejemplares_geom_gix ON public.ejemplares USING gist (geom);
CREATE INDEX ejemplares_lugar_idx ON public.ejemplares USING btree (ubicacion_lugar_id);
CREATE UNIQUE INDEX ejemplares_origen_ref_uidx ON public.ejemplares USING btree (origen_ref);
CREATE INDEX ejemplares_sector_cuartel_idx ON public.ejemplares USING btree (sector_cuartel_id) WHERE (sector_cuartel_id IS NOT NULL);
CREATE TRIGGER ejemplares_sector_cuartel_clase_trg BEFORE INSERT OR UPDATE OF sector_cuartel_id ON public.ejemplares FOR EACH ROW EXECUTE FUNCTION public.ejemplares_sector_cuartel_clase();
-- codigos_historicos
CREATE TABLE public.codigos_historicos (
    id bigint NOT NULL,
    ejemplar_id bigint NOT NULL,
    codigo_anterior text NOT NULL,
    codigo_nuevo text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.codigos_historicos IS 'Código anterior de un ejemplar. La recodificación masiva no vive aquí.';
CREATE SEQUENCE public.codigos_historicos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.codigos_historicos_id_seq OWNED BY public.codigos_historicos.id;
ALTER TABLE ONLY public.codigos_historicos ALTER COLUMN id SET DEFAULT nextval('public.codigos_historicos_id_seq'::regclass);
ALTER TABLE ONLY public.codigos_historicos
    ADD CONSTRAINT codigos_historicos_pkey PRIMARY KEY (id);
CREATE INDEX codigos_historicos_ejemplar_idx ON public.codigos_historicos USING btree (ejemplar_id, id);
-- medidas_palmera
CREATE TABLE public.medidas_palmera (
    ejemplar_id bigint NOT NULL,
    utm_norte double precision,
    utm_este double precision,
    altura double precision,
    altura_fuste double precision,
    dap double precision,
    radio double precision,
    zunchado boolean,
    baja_en timestamp with time zone,
    CONSTRAINT medidas_palmera_altura_chk CHECK (((altura IS NULL) OR (altura >= (0)::double precision))),
    CONSTRAINT medidas_palmera_dap_chk CHECK (((dap IS NULL) OR (dap >= (0)::double precision))),
    CONSTRAINT medidas_palmera_fuste_chk CHECK (((altura_fuste IS NULL) OR (altura_fuste >= (0)::double precision))),
    CONSTRAINT medidas_palmera_radio_chk CHECK (((radio IS NULL) OR (radio >= (0)::double precision)))
);
COMMENT ON TABLE public.medidas_palmera IS 'Medidas 1:1 de una palmera. La fila se rechaza en la carga si la coordenada no cae en el campus.';
COMMENT ON COLUMN public.medidas_palmera.baja_en IS 'Baja lógica al revertir el lote que cargó la medida. La fila permanece. No hay plazo de retención: no se borra por antigüedad.';
ALTER TABLE ONLY public.medidas_palmera
    ADD CONSTRAINT medidas_palmera_pkey PRIMARY KEY (ejemplar_id);
-- fauna
CREATE TABLE public.fauna (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    nombre text,
    geom public.geometry(Geometry,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.fauna IS 'Avistamiento. nombre es el del animal, no el de una persona.';
CREATE SEQUENCE public.fauna_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.fauna_id_seq OWNED BY public.fauna.id;
ALTER TABLE ONLY public.fauna ALTER COLUMN id SET DEFAULT nextval('public.fauna_id_seq'::regclass);
ALTER TABLE ONLY public.fauna
    ADD CONSTRAINT fauna_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.fauna
    ADD CONSTRAINT fauna_pkey PRIMARY KEY (id);
CREATE INDEX fauna_geom_gix ON public.fauna USING gist (geom);
-- puertas
CREATE TABLE public.puertas (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    codigo text,
    nombre text,
    geom public.geometry(Geometry,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.puertas IS 'Acceso del campus. nombre queda vacío hasta que lo entreguen.';
CREATE SEQUENCE public.puertas_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.puertas_id_seq OWNED BY public.puertas.id;
ALTER TABLE ONLY public.puertas ALTER COLUMN id SET DEFAULT nextval('public.puertas_id_seq'::regclass);
ALTER TABLE ONLY public.puertas
    ADD CONSTRAINT puertas_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.puertas
    ADD CONSTRAINT puertas_pkey PRIMARY KEY (id);
CREATE INDEX puertas_geom_gix ON public.puertas USING gist (geom);
-- playas_estacionamiento
CREATE TABLE public.playas_estacionamiento (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    codigo text,
    geom public.geometry(Geometry,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.playas_estacionamiento IS 'Playa de estacionamiento.';
CREATE SEQUENCE public.playas_estacionamiento_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.playas_estacionamiento_id_seq OWNED BY public.playas_estacionamiento.id;
ALTER TABLE ONLY public.playas_estacionamiento ALTER COLUMN id SET DEFAULT nextval('public.playas_estacionamiento_id_seq'::regclass);
ALTER TABLE ONLY public.playas_estacionamiento
    ADD CONSTRAINT playas_estacionamiento_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.playas_estacionamiento
    ADD CONSTRAINT playas_estacionamiento_pkey PRIMARY KEY (id);
CREATE INDEX playas_estacionamiento_geom_gix ON public.playas_estacionamiento USING gist (geom);
-- veredas_riesgo
CREATE TABLE public.veredas_riesgo (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    nota text,
    geom public.geometry(Geometry,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
COMMENT ON TABLE public.veredas_riesgo IS 'Tramo de vereda en riesgo.';
CREATE SEQUENCE public.veredas_riesgo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.veredas_riesgo_id_seq OWNED BY public.veredas_riesgo.id;
ALTER TABLE ONLY public.veredas_riesgo ALTER COLUMN id SET DEFAULT nextval('public.veredas_riesgo_id_seq'::regclass);
ALTER TABLE ONLY public.veredas_riesgo
    ADD CONSTRAINT veredas_riesgo_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.veredas_riesgo
    ADD CONSTRAINT veredas_riesgo_pkey PRIMARY KEY (id);
CREATE INDEX veredas_riesgo_geom_gix ON public.veredas_riesgo USING gist (geom);
-- xerofiticas
CREATE TABLE public.xerofiticas (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    clase text,
    riego text,
    area_m2 double precision,
    perimetro_m double precision,
    geom public.geometry(MultiPolygon,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT xerofiticas_area_chk CHECK (((area_m2 IS NULL) OR (area_m2 >= (0)::double precision))),
    CONSTRAINT xerofiticas_perimetro_chk CHECK (((perimetro_m IS NULL) OR (perimetro_m >= (0)::double precision)))
);
COMMENT ON TABLE public.xerofiticas IS 'Área xerofítica.';
CREATE SEQUENCE public.xerofiticas_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.xerofiticas_id_seq OWNED BY public.xerofiticas.id;
ALTER TABLE ONLY public.xerofiticas ALTER COLUMN id SET DEFAULT nextval('public.xerofiticas_id_seq'::regclass);
ALTER TABLE ONLY public.xerofiticas
    ADD CONSTRAINT xerofiticas_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.xerofiticas
    ADD CONSTRAINT xerofiticas_pkey PRIMARY KEY (id);
CREATE INDEX xerofiticas_geom_gix ON public.xerofiticas USING gist (geom);
-- jardines_reserva
CREATE TABLE public.jardines_reserva (
    id bigint NOT NULL,
    feature_id text NOT NULL,
    codigo text,
    nombre text,
    uso text,
    proy_riego text,
    riego_act text,
    referencia text,
    pertenecen text,
    perimetro_m double precision,
    area_m2 double precision,
    geom public.geometry(MultiPolygon,4326),
    origen_ref text,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT jardines_reserva_area_chk CHECK (((area_m2 IS NULL) OR (area_m2 >= (0)::double precision))),
    CONSTRAINT jardines_reserva_perimetro_chk CHECK (((perimetro_m IS NULL) OR (perimetro_m >= (0)::double precision))),
    CONSTRAINT jardines_reserva_referencia_chk CHECK (((referencia IS NULL) OR (char_length(referencia) <= 500)))
);
COMMENT ON TABLE public.jardines_reserva IS 'Jardín reservable. pertenecen es una unidad organizativa, no una persona.';
CREATE SEQUENCE public.jardines_reserva_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.jardines_reserva_id_seq OWNED BY public.jardines_reserva.id;
ALTER TABLE ONLY public.jardines_reserva ALTER COLUMN id SET DEFAULT nextval('public.jardines_reserva_id_seq'::regclass);
ALTER TABLE ONLY public.jardines_reserva
    ADD CONSTRAINT jardines_reserva_feature_id_key UNIQUE (feature_id);
ALTER TABLE ONLY public.jardines_reserva
    ADD CONSTRAINT jardines_reserva_pkey PRIMARY KEY (id);
CREATE INDEX jardines_reserva_geom_gix ON public.jardines_reserva USING gist (geom);
-- tachos
CREATE TABLE public.tachos (
    id bigint NOT NULL,
    codigo text NOT NULL,
    lat double precision,
    lon double precision,
    nota text,
    lugar text,
    espacios text,
    accion text,
    tacho_actual text,
    tacho_nuevo text,
    recomendaciones text,
    no_aprovechables integer DEFAULT 0 NOT NULL,
    papel_carton integer DEFAULT 0 NOT NULL,
    plastico integer DEFAULT 0 NOT NULL,
    vidrio integer DEFAULT 0 NOT NULL,
    pilas integer DEFAULT 0 NOT NULL,
    peligrosos integer DEFAULT 0 NOT NULL,
    raee integer DEFAULT 0 NOT NULL,
    metales integer DEFAULT 0 NOT NULL,
    aniquem integer DEFAULT 0 NOT NULL,
    intermedios_plastico integer DEFAULT 0 NOT NULL,
    intermedios_metal integer DEFAULT 0 NOT NULL,
    foto text,
    zona_supervision_id bigint,
    geom public.geometry(Point,4326),
    origen_ref text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tachos_codigo_chk CHECK ((codigo ~ '^PT'::text)),
    CONSTRAINT tachos_conteos_chk CHECK (((no_aprovechables >= 0) AND (papel_carton >= 0) AND (plastico >= 0) AND (vidrio >= 0) AND (pilas >= 0) AND (peligrosos >= 0) AND (raee >= 0) AND (metales >= 0) AND (aniquem >= 0) AND (intermedios_plastico >= 0) AND (intermedios_metal >= 0))),
    CONSTRAINT tachos_lat_chk CHECK (((lat IS NULL) OR ((lat >= ('-12.20'::numeric)::double precision) AND (lat <= ('-11.90'::numeric)::double precision)))),
    CONSTRAINT tachos_lon_chk CHECK (((lon IS NULL) OR ((lon >= ('-77.30'::numeric)::double precision) AND (lon <= ('-76.90'::numeric)::double precision))))
);
COMMENT ON TABLE public.tachos IS 'Punto de residuos. Los 11 conteos son columnas propias. foto es un archivo local, nunca una URL de Drive.';
CREATE SEQUENCE public.tachos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.tachos_id_seq OWNED BY public.tachos.id;
ALTER TABLE ONLY public.tachos ALTER COLUMN id SET DEFAULT nextval('public.tachos_id_seq'::regclass);
ALTER TABLE ONLY public.tachos
    ADD CONSTRAINT tachos_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX tachos_codigo_activo_uidx ON public.tachos USING btree (codigo) WHERE activo;
CREATE INDEX tachos_geom_gix ON public.tachos USING gist (geom);
CREATE UNIQUE INDEX tachos_origen_ref_uidx ON public.tachos USING btree (origen_ref);
-- bebederos
CREATE TABLE public.bebederos (
    id bigint NOT NULL,
    codigo text NOT NULL,
    subtipo text NOT NULL,
    estado text NOT NULL,
    sede text,
    lat double precision,
    lon double precision,
    foto text,
    geom public.geometry(Point,4326),
    origen_ref text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT bebederos_codigo_chk CHECK ((codigo ~ '^PT_'::text)),
    CONSTRAINT bebederos_lat_chk CHECK (((lat IS NULL) OR ((lat >= ('-12.20'::numeric)::double precision) AND (lat <= ('-11.90'::numeric)::double precision)))),
    CONSTRAINT bebederos_lon_chk CHECK (((lon IS NULL) OR ((lon >= ('-77.30'::numeric)::double precision) AND (lon <= ('-76.90'::numeric)::double precision)))),
    CONSTRAINT bebederos_subtipo_chk CHECK ((subtipo = ANY (ARRAY['fuente'::text, 'llenador'::text, 'nuevo'::text, 'deterioro'::text, 'baja'::text])))
);
COMMENT ON TABLE public.bebederos IS 'Bebedero. estado viene de la columna del archivo, no solo del nombre. sede es la sede KML.';
COMMENT ON COLUMN public.bebederos.foto IS 'Nombre JPEG local si el índice de 70 códigos lo encuentra. No es un id de Drive.';
CREATE SEQUENCE public.bebederos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.bebederos_id_seq OWNED BY public.bebederos.id;
ALTER TABLE ONLY public.bebederos ALTER COLUMN id SET DEFAULT nextval('public.bebederos_id_seq'::regclass);
ALTER TABLE ONLY public.bebederos
    ADD CONSTRAINT bebederos_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX bebederos_codigo_activo_uidx ON public.bebederos USING btree (codigo) WHERE activo;
CREATE INDEX bebederos_geom_gix ON public.bebederos USING gist (geom);
CREATE UNIQUE INDEX bebederos_origen_ref_uidx ON public.bebederos USING btree (origen_ref);
-- puntos_pucp
CREATE TABLE public.puntos_pucp (
    id bigint NOT NULL,
    titulo text NOT NULL,
    lat double precision NOT NULL,
    lon double precision NOT NULL,
    url text,
    geom public.geometry(Point,4326) NOT NULL,
    origen_ref text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT puntos_pucp_lat_chk CHECK (((lat >= ('-12.20'::numeric)::double precision) AND (lat <= ('-11.90'::numeric)::double precision))),
    CONSTRAINT puntos_pucp_lon_chk CHECK (((lon >= ('-77.30'::numeric)::double precision) AND (lon <= ('-76.90'::numeric)::double precision))),
    CONSTRAINT puntos_pucp_url_chk CHECK (((url IS NULL) OR ((url !~~ '%place_id=%'::text) AND (url !~* 'phone'::text) AND (POSITION(('placeId'::text) IN (url)) = 0))))
);
COMMENT ON TABLE public.puntos_pucp IS 'Lugar público del campus. Solo título, coordenada y URL de mapa. No hay columnas de contacto.';
CREATE SEQUENCE public.puntos_pucp_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.puntos_pucp_id_seq OWNED BY public.puntos_pucp.id;
ALTER TABLE ONLY public.puntos_pucp ALTER COLUMN id SET DEFAULT nextval('public.puntos_pucp_id_seq'::regclass);
ALTER TABLE ONLY public.puntos_pucp
    ADD CONSTRAINT puntos_pucp_pkey PRIMARY KEY (id);
CREATE INDEX puntos_pucp_geom_gix ON public.puntos_pucp USING gist (geom);
CREATE UNIQUE INDEX puntos_pucp_origen_ref_uidx ON public.puntos_pucp USING btree (origen_ref);
CREATE INDEX puntos_pucp_titulo_idx ON public.puntos_pucp USING btree (lower(titulo));
-- reservas_jardin
CREATE TABLE public.reservas_jardin (
    id bigint NOT NULL,
    jardin_id bigint,
    fecha date NOT NULL,
    hora_inicio time without time zone NOT NULL,
    hora_fin time without time zone NOT NULL,
    estado text NOT NULL,
    evento text NOT NULL,
    unidad text,
    origen text DEFAULT 'ficticio'::text NOT NULL,
    origen_ref text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reservas_jardin_estado_chk CHECK ((estado = ANY (ARRAY['reservado'::text, 'realizado'::text, 'cancelado'::text]))),
    CONSTRAINT reservas_jardin_hora_chk CHECK ((hora_fin > hora_inicio)),
    CONSTRAINT reservas_jardin_origen_chk CHECK ((origen = 'ficticio'::text))
);
COMMENT ON TABLE public.reservas_jardin IS 'Agenda ficticia. No proviene de la hoja de reservas (HTTP 401).';
COMMENT ON COLUMN public.reservas_jardin.unidad IS 'Unidad organizativa, no una persona.';
COMMENT ON COLUMN public.reservas_jardin.origen IS 'Siempre ficticio mientras la hoja institucional responda 401.';
CREATE SEQUENCE public.reservas_jardin_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.reservas_jardin_id_seq OWNED BY public.reservas_jardin.id;
ALTER TABLE ONLY public.reservas_jardin ALTER COLUMN id SET DEFAULT nextval('public.reservas_jardin_id_seq'::regclass);
ALTER TABLE ONLY public.reservas_jardin
    ADD CONSTRAINT reservas_jardin_pkey PRIMARY KEY (id);
CREATE INDEX reservas_jardin_fecha_idx ON public.reservas_jardin USING btree (fecha, hora_inicio);
CREATE UNIQUE INDEX reservas_jardin_origen_ref_uidx ON public.reservas_jardin USING btree (origen_ref);
-- ========================================================================
-- Catálogos
-- ========================================================================
-- catalogos
CREATE TABLE public.catalogos (
    id bigint NOT NULL,
    clase text NOT NULL,
    codigo text NOT NULL,
    nombre text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    orden integer DEFAULT 0 NOT NULL,
    provisional boolean DEFAULT false NOT NULL,
    padre_codigo text
);
COMMENT ON TABLE public.catalogos IS 'Valores parametrizables. La baja es lógica (activo = false).';
COMMENT ON COLUMN public.catalogos.provisional IS 'El cliente todavía no cierra este valor. Se muestra, no se trata como decisión oficial.';
COMMENT ON COLUMN public.catalogos.padre_codigo IS 'Código de la clase de actividad cuando el ítem es un tipo de segundo nivel. Nulo en el resto.';
CREATE SEQUENCE public.catalogos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.catalogos_id_seq OWNED BY public.catalogos.id;
ALTER TABLE ONLY public.catalogos ALTER COLUMN id SET DEFAULT nextval('public.catalogos_id_seq'::regclass);
ALTER TABLE ONLY public.catalogos
    ADD CONSTRAINT catalogos_clase_codigo_key UNIQUE (clase, codigo);
ALTER TABLE ONLY public.catalogos
    ADD CONSTRAINT catalogos_pkey PRIMARY KEY (id);
-- ========================================================================
-- Auditoría
-- ========================================================================
-- cambios
CREATE TABLE public.cambios (
    id bigint NOT NULL,
    entidad text NOT NULL,
    entidad_id text NOT NULL,
    accion text NOT NULL,
    antes jsonb,
    despues jsonb,
    usuario_id bigint,
    lote_id bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cambios_accion_chk CHECK ((accion = ANY (ARRAY['alta'::text, 'edicion'::text, 'baja'::text, 'importacion'::text, 'reversion'::text])))
);
COMMENT ON TABLE public.cambios IS 'Antes y después de una edición. El actor es el usuario de sesión cuando exista.';
CREATE SEQUENCE public.cambios_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.cambios_id_seq OWNED BY public.cambios.id;
ALTER TABLE ONLY public.cambios ALTER COLUMN id SET DEFAULT nextval('public.cambios_id_seq'::regclass);
ALTER TABLE ONLY public.cambios
    ADD CONSTRAINT cambios_pkey PRIMARY KEY (id);
CREATE INDEX cambios_entidad_idx ON public.cambios USING btree (entidad, entidad_id, id);
CREATE INDEX cambios_lote_idx ON public.cambios USING btree (lote_id) WHERE (lote_id IS NOT NULL);
-- lotes_importacion
CREATE TABLE public.lotes_importacion (
    id bigint NOT NULL,
    entidad text NOT NULL,
    estado text DEFAULT 'confirmado'::text NOT NULL,
    usuario_id bigint NOT NULL,
    filas integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revertido_en timestamp with time zone,
    contenido bytea,
    nombre_archivo text DEFAULT ''::text NOT NULL,
    CONSTRAINT lotes_estado_chk CHECK ((estado = ANY (ARRAY['vista_previa'::text, 'confirmado'::text, 'revertido'::text])))
);
COMMENT ON TABLE public.lotes_importacion IS 'Lote de importación. La reversión aplica el antes de cambios y no pisa una edición posterior.';
COMMENT ON COLUMN public.lotes_importacion.contenido IS 'Archivo de la vista previa. Confirmar lo vuelve a leer. Revertir no lo necesita.';
CREATE SEQUENCE public.lotes_importacion_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.lotes_importacion_id_seq OWNED BY public.lotes_importacion.id;
ALTER TABLE ONLY public.lotes_importacion ALTER COLUMN id SET DEFAULT nextval('public.lotes_importacion_id_seq'::regclass);
ALTER TABLE ONLY public.lotes_importacion
    ADD CONSTRAINT lotes_importacion_pkey PRIMARY KEY (id);
-- ========================================================================
-- Evidencias
-- ========================================================================
-- evidencias
CREATE TABLE public.evidencias (
    id uuid NOT NULL,
    actividad_id uuid,
    solicitud_id uuid,
    orden_id uuid,
    nombre text NOT NULL,
    mime text NOT NULL,
    bytes integer NOT NULL,
    ruta text NOT NULL,
    nota text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    sha256 text,
    lat double precision,
    lon double precision,
    exif jsonb,
    evento_id bigint
);
COMMENT ON COLUMN public.evidencias.solicitud_id IS 'Solicitud a la que puede colgar la evidencia, además de la actividad. Nulo si no aplica.';
COMMENT ON COLUMN public.evidencias.sha256 IS 'SHA-256 hexadecimal del archivo. El mismo id con otro hash responde 409.';
COMMENT ON COLUMN public.evidencias.lat IS 'Latitud opcional (EXIF o permiso del teléfono). NULL si no hay punto.';
COMMENT ON COLUMN public.evidencias.lon IS 'Longitud opcional, en el mismo sistema que lat.';
COMMENT ON COLUMN public.evidencias.exif IS 'Solo fecha, lat, lon y orientacion. El resto se descarta.';
COMMENT ON COLUMN public.evidencias.evento_id IS 'Hito de la cadena al que cuelga la evidencia. Nulo en archivos anteriores a este vínculo.';
ALTER TABLE ONLY public.evidencias
    ADD CONSTRAINT evidencias_pkey PRIMARY KEY (id);
CREATE INDEX evidencias_evento_id_idx ON public.evidencias USING btree (evento_id);
CREATE INDEX evidencias_solicitud_id_idx ON public.evidencias USING btree (solicitud_id) WHERE (solicitud_id IS NOT NULL);
-- ========================================================================
-- Claves foráneas
-- ========================================================================
-- Acceso
ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_rol_fkey FOREIGN KEY (rol) REFERENCES public.roles(codigo);
ALTER TABLE ONLY public.sesiones
    ADD CONSTRAINT sesiones_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_capataz_id_fkey FOREIGN KEY (capataz_id) REFERENCES public.capataces(id);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_cuadrilla_id_fkey FOREIGN KEY (cuadrilla_id) REFERENCES public.cuadrillas(id);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_rol_fkey FOREIGN KEY (rol) REFERENCES public.roles(codigo);
-- Catastro
ALTER TABLE ONLY public.areas_verdes
    ADD CONSTRAINT areas_verdes_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
ALTER TABLE ONLY public.asignaciones_poligono
    ADD CONSTRAINT asignaciones_poligono_cuadrilla_id_fkey FOREIGN KEY (cuadrilla_id) REFERENCES public.cuadrillas(id);
ALTER TABLE ONLY public.asignaciones_poligono
    ADD CONSTRAINT asignaciones_poligono_poligono_id_fkey FOREIGN KEY (poligono_id) REFERENCES public.poligonos_cuadrilla(id);
ALTER TABLE ONLY public.lugares
    ADD CONSTRAINT lugares_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
ALTER TABLE ONLY public.poligonos_cuadrilla
    ADD CONSTRAINT poligonos_cuadrilla_cuadrilla_id_fkey FOREIGN KEY (cuadrilla_id) REFERENCES public.cuadrillas(id);
ALTER TABLE ONLY public.poligonos_cuadrilla
    ADD CONSTRAINT poligonos_cuadrilla_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
ALTER TABLE ONLY public.referentes_edificio
    ADD CONSTRAINT referentes_edificio_lugar_id_fkey FOREIGN KEY (lugar_id) REFERENCES public.lugares(id);
-- Operación
ALTER TABLE ONLY public.actividad_avances
    ADD CONSTRAINT actividad_avances_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.actividad_eventos
    ADD CONSTRAINT actividad_eventos_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.actividad_eventos
    ADD CONSTRAINT actividad_eventos_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id);
ALTER TABLE ONLY public.actividades
    ADD CONSTRAINT actividades_assigned_capataz_id_fkey FOREIGN KEY (assigned_capataz_id) REFERENCES public.capataces(id);
ALTER TABLE ONLY public.actividades
    ADD CONSTRAINT actividades_cuadrilla_fk FOREIGN KEY (cuadrilla_id) REFERENCES public.cuadrillas(id);
ALTER TABLE ONLY public.actividades
    ADD CONSTRAINT actividades_lugar_id_fkey FOREIGN KEY (lugar_id) REFERENCES public.lugares(id);
ALTER TABLE ONLY public.actividades
    ADD CONSTRAINT actividades_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
ALTER TABLE ONLY public.ordenes_servicio
    ADD CONSTRAINT ordenes_empresa_id_fkey FOREIGN KEY (empresa_id) REFERENCES public.catalogos(id);
ALTER TABLE ONLY public.ordenes_servicio
    ADD CONSTRAINT ordenes_frecuencia_id_fkey FOREIGN KEY (frecuencia_id) REFERENCES public.catalogos(id);
ALTER TABLE ONLY public.ordenes_servicio
    ADD CONSTRAINT ordenes_servicio_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.personal_labor
    ADD CONSTRAINT personal_labor_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.podas
    ADD CONSTRAINT podas_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.podas
    ADD CONSTRAINT podas_solicitud_id_fkey FOREIGN KEY (solicitud_id) REFERENCES public.solicitudes(id);
ALTER TABLE ONLY public.riego_registros
    ADD CONSTRAINT riego_registros_capataz_id_fkey FOREIGN KEY (capataz_id) REFERENCES public.capataces(id);
ALTER TABLE ONLY public.riego_registros
    ADD CONSTRAINT riego_registros_sector_id_fkey FOREIGN KEY (sector_id) REFERENCES public.sectores_capataz(id);
ALTER TABLE ONLY public.riego_registros
    ADD CONSTRAINT riego_registros_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
ALTER TABLE ONLY public.solicitudes
    ADD CONSTRAINT solicitudes_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.solicitudes
    ADD CONSTRAINT solicitudes_lugar_id_fkey FOREIGN KEY (lugar_id) REFERENCES public.lugares(id);
-- Inventario
ALTER TABLE ONLY public.codigos_historicos
    ADD CONSTRAINT codigos_historicos_ejemplar_id_fkey FOREIGN KEY (ejemplar_id) REFERENCES public.ejemplares(id);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_area_verde_id_fkey FOREIGN KEY (area_verde_id) REFERENCES public.areas_verdes(id);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_especie_id_fkey FOREIGN KEY (especie_id) REFERENCES public.especies(id);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_foto_id_fkey FOREIGN KEY (foto_id) REFERENCES public.evidencias(id);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_sector_cuartel_fk FOREIGN KEY (sector_cuartel_id) REFERENCES public.catalogos(id);
ALTER TABLE ONLY public.ejemplares
    ADD CONSTRAINT ejemplares_ubicacion_lugar_id_fkey FOREIGN KEY (ubicacion_lugar_id) REFERENCES public.lugares(id);
ALTER TABLE ONLY public.medidas_palmera
    ADD CONSTRAINT medidas_palmera_ejemplar_id_fkey FOREIGN KEY (ejemplar_id) REFERENCES public.ejemplares(id);
ALTER TABLE ONLY public.reservas_jardin
    ADD CONSTRAINT reservas_jardin_jardin_id_fkey FOREIGN KEY (jardin_id) REFERENCES public.jardines_reserva(id);
ALTER TABLE ONLY public.tachos
    ADD CONSTRAINT tachos_zona_supervision_id_fkey FOREIGN KEY (zona_supervision_id) REFERENCES public.zonas_supervision(id);
-- Auditoría
ALTER TABLE ONLY public.cambios
    ADD CONSTRAINT cambios_lote_fkey FOREIGN KEY (lote_id) REFERENCES public.lotes_importacion(id);
ALTER TABLE ONLY public.cambios
    ADD CONSTRAINT cambios_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id);
ALTER TABLE ONLY public.lotes_importacion
    ADD CONSTRAINT lotes_importacion_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id);
-- Evidencias
ALTER TABLE ONLY public.evidencias
    ADD CONSTRAINT evidencias_actividad_id_fkey FOREIGN KEY (actividad_id) REFERENCES public.actividades(id);
ALTER TABLE ONLY public.evidencias
    ADD CONSTRAINT evidencias_evento_id_fkey FOREIGN KEY (evento_id) REFERENCES public.actividad_eventos(id);
ALTER TABLE ONLY public.evidencias
    ADD CONSTRAINT evidencias_orden_id_fkey FOREIGN KEY (orden_id) REFERENCES public.ordenes_servicio(id);
ALTER TABLE ONLY public.evidencias
    ADD CONSTRAINT evidencias_solicitud_id_fkey FOREIGN KEY (solicitud_id) REFERENCES public.solicitudes(id);
-- ========================================================================
-- Datos de referencia
-- Filas que la serie histórica dejaba en una base vacía.
-- Sin usuarios ni sesiones: los crea la semilla de accesos.
-- ========================================================================
INSERT INTO public.capataces (id, equipo, turno, activo) VALUES ('cap-norte', 'Cuadrilla Norte', 'mañana', true);
INSERT INTO public.capataces (id, equipo, turno, activo) VALUES ('cap-sur', 'Cuadrilla Sur', 'mañana', true);
INSERT INTO public.capataces (id, equipo, turno, activo) VALUES ('cap-riego', 'Cuadrilla Riego', 'tarde', true);
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-valeria', 'Valeria Quispe', 'manana', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-mateo', 'Mateo Salazar', 'manana', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-renato', 'Renato Cárdenas', 'tarde', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-nora', 'Nora Beltrán', 'manana', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-ivan', 'Iván Paredes', 'tarde', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cua-lucia', 'Lucía Mendoza', 'manana', true, '2026-10-03 04:52:21.607609+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cap-norte', 'Cuadrilla Norte', 'manana', true, '2026-10-03 04:52:21.479777+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cap-sur', 'Cuadrilla Sur', 'manana', true, '2026-10-03 04:52:21.479777+00');
INSERT INTO public.cuadrillas (id, nombre_ficticio, turno, activo, created_at) VALUES ('cap-riego', 'Cuadrilla Riego', 'tarde', true, '2026-10-03 04:52:21.479777+00');
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('11111111-1111-4111-8111-111111111111', 'riego', 'pendiente', 'Riego del eje central', 'Revisar aspersores de la explanada central.', NULL, NULL, 'cap-norte', '0101000020E61000006519E258174553C0A4DFBE0E9C2328C0', NULL, '2026-09-24 13:10:00+00', '2026-09-24 13:10:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('22222222-2222-4222-8222-222222222222', 'poda', 'en_proceso', 'Poda de setos sur', 'Setos del borde sur, altura de mantenimiento.', NULL, NULL, 'cap-sur', '0101000020E61000001361C3D32B4553C0DC4603780B2428C0', NULL, '2026-09-24 13:20:00+00', '2026-09-24 14:05:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('33333333-3333-4333-8333-333333333333', 'limpieza', 'pendiente', 'Limpieza de caminería norte', 'Hojas y residuos en la caminería del sector norte.', NULL, NULL, 'cap-norte', '0101000020E6100000E2E995B20C4553C0A52C431CEB2228C0', NULL, '2026-09-24 13:30:00+00', '2026-09-24 13:30:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('44444444-4444-4444-8444-444444444444', 'incidencia', 'bloqueada', 'Fuga en línea de riego', 'Espera de corte de agua antes de intervenir.', NULL, NULL, 'cap-riego', '0101000020E6100000DA1B7C61324553C0F853E3A59B2428C0', NULL, '2026-09-24 13:40:00+00', '2026-09-24 15:00:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('55555555-5555-4555-8555-555555555555', 'inspeccion', 'pendiente', 'Inspección bosque húmedo', 'Recorrido de control del sector bosque húmedo.', NULL, NULL, 'cap-sur', '0101000020E6100000F775E09C114553C0304CA60A462528C0', NULL, '2026-09-24 13:50:00+00', '2026-09-24 13:50:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('66666666-6666-4666-8666-666666666666', 'riego', 'cerrada', 'Riego de losas, turno temprano', 'Cerrada en la semilla: no debe salir en el listado abierto.', NULL, NULL, 'cap-riego', '0101000020E6100000E9482EFF214553C0E0BE0E9C332228C0', NULL, '2026-09-23 12:00:00+00', '2026-09-23 16:00:00+00', 'propia', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividades (id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id, assigned_capataz_id, geom, archivada_en, created_at, updated_at, ejecutor, motivo_archivo, lugar_id, zona_supervision_id, fecha_solicitud, fecha_atencion, cuadrilla_id, clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen, subtipo, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada, cantidad) VALUES ('99999999-9999-4999-8999-999999999999', 'poda', 'pendiente', 'Poda contratada del borde sur', 'Servicio externo de setos. No se cierra sin una orden.', NULL, NULL, 'cap-sur', '0101000020E6100000DA1B7C61324553C0151DC9E53F2428C0', NULL, '2026-10-03 04:52:21.450843+00', '2026-10-03 04:52:21.450843+00', 'tercerizada', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', '', NULL, 'interna', NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (1, '11111111-1111-4111-8111-111111111111', 'creada', 'pendiente', 'cap-norte', 'coordinacion', 'Carga de demostración', '2026-09-24 13:10:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (2, '22222222-2222-4222-8222-222222222222', 'creada', 'en_proceso', 'cap-sur', 'coordinacion', 'Carga de demostración', '2026-09-24 13:20:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (3, '33333333-3333-4333-8333-333333333333', 'creada', 'pendiente', 'cap-norte', 'coordinacion', 'Carga de demostración', '2026-09-24 13:30:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (4, '44444444-4444-4444-8444-444444444444', 'creada', 'bloqueada', 'cap-riego', 'coordinacion', 'Carga de demostración', '2026-09-24 13:40:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (5, '55555555-5555-4555-8555-555555555555', 'creada', 'pendiente', 'cap-sur', 'coordinacion', 'Carga de demostración', '2026-09-24 13:50:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (6, '66666666-6666-4666-8666-666666666666', 'creada', 'cerrada', 'cap-riego', 'coordinacion', 'Carga de demostración', '2026-09-23 12:00:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (7, '11111111-1111-4111-8111-111111111111', 'asignada', 'pendiente', 'cap-norte', 'coordinacion', 'Asignación inicial de demostración', '2026-09-24 13:10:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (8, '22222222-2222-4222-8222-222222222222', 'asignada', 'en_proceso', 'cap-sur', 'coordinacion', 'Asignación inicial de demostración', '2026-09-24 13:20:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (9, '33333333-3333-4333-8333-333333333333', 'asignada', 'pendiente', 'cap-norte', 'coordinacion', 'Asignación inicial de demostración', '2026-09-24 13:30:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (10, '44444444-4444-4444-8444-444444444444', 'asignada', 'bloqueada', 'cap-riego', 'coordinacion', 'Asignación inicial de demostración', '2026-09-24 13:40:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (11, '55555555-5555-4555-8555-555555555555', 'asignada', 'pendiente', 'cap-sur', 'coordinacion', 'Asignación inicial de demostración', '2026-09-24 13:50:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (12, '66666666-6666-4666-8666-666666666666', 'asignada', 'cerrada', 'cap-riego', 'coordinacion', 'Asignación inicial de demostración', '2026-09-23 12:00:00+00', NULL, NULL, NULL);
INSERT INTO public.actividad_eventos (id, actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at, usuario_id, uuid_cliente, capataz_anterior) VALUES (13, '99999999-9999-4999-8999-999999999999', 'creada', 'pendiente', 'cap-sur', 'coordinacion', 'Alta de demostración', '2026-10-03 04:52:21.450843+00', NULL, NULL, NULL);
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (1, 'roles', 'admin', 'edicion', '{"nombre": "Administración"}', '{"nombre": "Administrador del sistema"}', NULL, NULL, '2026-10-03 04:52:21.759993+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (2, 'roles', 'coordinacion', 'edicion', '{"nombre": "Coordinación"}', '{"nombre": "Ingeniería/Coordinación"}', NULL, NULL, '2026-10-03 04:52:21.759993+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (3, 'roles', 'jefatura', 'edicion', '{"nombre": "Jefatura"}', '{"nombre": "Jefatura / Jefe de sección"}', NULL, NULL, '2026-10-03 04:52:21.759993+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (4, 'catalogos', '6', 'edicion', '{"clase": "estado", "orden": 1, "activo": true, "codigo": "pendiente", "nombre": "Pendiente", "provisional": false}', '{"clase": "estado", "orden": 1, "activo": true, "codigo": "pendiente", "nombre": "Por iniciar", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (5, 'catalogos', '8', 'edicion', '{"clase": "estado", "orden": 3, "activo": true, "codigo": "bloqueada", "nombre": "Bloqueada", "provisional": false}', '{"clase": "estado", "orden": 90, "activo": false, "codigo": "bloqueada", "nombre": "Bloqueada", "provisional": true}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (6, 'catalogos', '9', 'edicion', '{"clase": "estado", "orden": 4, "activo": true, "codigo": "cerrada", "nombre": "Cerrada", "provisional": false}', '{"clase": "estado", "orden": 4, "activo": true, "codigo": "cerrada", "nombre": "Cerrado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (7, 'catalogos', '10', 'edicion', '{"clase": "estado", "orden": 5, "activo": true, "codigo": "cancelada", "nombre": "Cancelada", "provisional": false}', '{"clase": "estado", "orden": 5, "activo": true, "codigo": "cancelada", "nombre": "Cancelado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (8, 'catalogos', '46', 'alta', NULL, '{"clase": "estado", "activo": true, "codigo": "ejecutado", "nombre": "Ejecutado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (9, 'catalogos', '47', 'alta', NULL, '{"clase": "estado", "activo": true, "codigo": "archivada", "nombre": "Archivado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.779564+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (10, 'catalogos', '48', 'alta', NULL, '{"clase": "clase_actividad", "activo": true, "codigo": "fitosanitario", "nombre": "Manejo fitosanitario", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781093+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (11, 'catalogos', '49', 'alta', NULL, '{"clase": "clase_actividad", "activo": true, "codigo": "inspeccion_monitoreo", "nombre": "Inspección y monitoreo", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781093+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (12, 'catalogos', '50', 'alta', NULL, '{"clase": "plaga", "activo": true, "codigo": "demo_hoja", "nombre": "Plaga de demostración", "provisional": true}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (13, 'catalogos', '51', 'alta', NULL, '{"clase": "producto_fitosanitario", "activo": true, "codigo": "demo_producto", "nombre": "Producto fitosanitario de demostración", "provisional": true}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (14, 'catalogos', '52', 'alta', NULL, '{"clase": "frecuencia", "activo": true, "codigo": "semanal", "nombre": "Semanal", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (15, 'catalogos', '53', 'alta', NULL, '{"clase": "frecuencia", "activo": true, "codigo": "quincenal", "nombre": "Quincenal", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (16, 'catalogos', '54', 'alta', NULL, '{"clase": "sede", "activo": true, "codigo": "sede_demo", "nombre": "Sede de demostración", "provisional": true}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (17, 'catalogos', '55', 'alta', NULL, '{"clase": "cuartel", "activo": true, "codigo": "sin_archivo", "nombre": "Sin archivo de cuarteles", "provisional": true}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (18, 'catalogos', '56', 'alta', NULL, '{"clase": "sector_capataz", "activo": true, "codigo": "sector_norte", "nombre": "Sector norte (ficticio)", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (19, 'catalogos', '57', 'alta', NULL, '{"clase": "sector_capataz", "activo": true, "codigo": "campo_deportivo", "nombre": "Campo deportivo", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (20, 'catalogos', '58', 'alta', NULL, '{"clase": "sector_capataz", "activo": true, "codigo": "bosque_humedo", "nombre": "Bosque húmedo", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.781892+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (21, 'capataces', 'cap-norte', 'edicion', '{"equipo": "Equipo Norte"}', '{"equipo": "Cuadrilla Norte"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (22, 'capataces', 'cap-sur', 'edicion', '{"equipo": "Equipo Sur"}', '{"equipo": "Cuadrilla Sur"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (23, 'capataces', 'cap-riego', 'edicion', '{"equipo": "Equipo Riego"}', '{"equipo": "Cuadrilla Riego"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (24, 'cuadrillas', 'cap-norte', 'edicion', '{"nombre_ficticio": "Equipo Norte"}', '{"nombre_ficticio": "Cuadrilla Norte"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (25, 'cuadrillas', 'cap-sur', 'edicion', '{"nombre_ficticio": "Equipo Sur"}', '{"nombre_ficticio": "Cuadrilla Sur"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (26, 'cuadrillas', 'cap-riego', 'edicion', '{"nombre_ficticio": "Equipo Riego"}', '{"nombre_ficticio": "Cuadrilla Riego"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (27, 'catalogos', '4', 'edicion', '{"clase": "tipo_actividad", "codigo": "incidencia", "nombre": "Incidencia"}', '{"clase": "tipo_actividad", "codigo": "incidencia", "nombre": "Novedad de campo"}', NULL, NULL, '2026-10-03 04:52:21.78295+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (28, 'sectores_capataz', '1', 'alta', NULL, '{"color": "#6b5596", "activo": true, "codigo": "cua-valeria", "nombre": "Sector de capataz — Valeria Quispe (ficticio)"}', NULL, NULL, '2026-10-03 04:52:21.784686+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (29, 'sectores_capataz', '2', 'alta', NULL, '{"color": "#3f73b0", "activo": true, "codigo": "cua-mateo", "nombre": "Sector de capataz — Mateo Salazar (ficticio)"}', NULL, NULL, '2026-10-03 04:52:21.784686+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (30, 'sectores_capataz', '3', 'alta', NULL, '{"color": "#c27c2c", "activo": true, "codigo": "cua-renato", "nombre": "Sector de capataz — Renato Cárdenas (ficticio)"}', NULL, NULL, '2026-10-03 04:52:21.784686+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (31, 'sectores_capataz', '4', 'alta', NULL, '{"color": "#9aab3e", "activo": true, "codigo": "campo-deportivo", "nombre": "Campo deportivo"}', NULL, NULL, '2026-10-03 04:52:21.784686+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (32, 'sectores_capataz', '5', 'alta', NULL, '{"color": "#3f9a82", "activo": true, "codigo": "bosque-humedo", "nombre": "Bosque húmedo"}', NULL, NULL, '2026-10-03 04:52:21.784686+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (33, 'catalogos', '106', 'alta', NULL, '{"clase": "estado_solicitud", "activo": true, "codigo": "por_iniciar", "nombre": "Por iniciar", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.853983+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (34, 'catalogos', '107', 'alta', NULL, '{"clase": "estado_solicitud", "activo": true, "codigo": "en_proceso", "nombre": "En proceso", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.853983+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (35, 'catalogos', '108', 'alta', NULL, '{"clase": "estado_solicitud", "activo": true, "codigo": "ejecutado", "nombre": "Ejecutado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.853983+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (36, 'catalogos', '109', 'alta', NULL, '{"clase": "estado_solicitud", "activo": true, "codigo": "cerrado", "nombre": "Cerrado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.853983+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (37, 'catalogos', '110', 'alta', NULL, '{"clase": "estado_solicitud", "activo": true, "codigo": "cancelado", "nombre": "Cancelado", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.853983+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (38, 'catalogos', '111', 'alta', NULL, '{"clase": "empresa", "activo": true, "codigo": "taller_verde_andino", "nombre": "Taller Verde Andino (ficticia)", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (39, 'catalogos', '112', 'alta', NULL, '{"clase": "empresa", "activo": true, "codigo": "jardines_del_rimac", "nombre": "Jardines del Rímac (ficticia)", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (40, 'catalogos', '113', 'alta', NULL, '{"clase": "empresa", "activo": true, "codigo": "poda_litoral", "nombre": "Poda Litoral (ficticia)", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (41, 'catalogos', '114', 'alta', NULL, '{"clase": "empresa", "activo": true, "codigo": "servicios_forestales_demo", "nombre": "Servicios Forestales del Perú", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (42, 'catalogos', '115', 'alta', NULL, '{"clase": "frecuencia", "activo": true, "codigo": "mensual", "nombre": "Mensual", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.cambios (id, entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at) VALUES (43, 'catalogos', '116', 'alta', NULL, '{"clase": "frecuencia", "activo": true, "codigo": "unica", "nombre": "Única", "provisional": false}', NULL, NULL, '2026-10-03 04:52:21.855335+00');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (1, 'tipo_actividad', 'riego', 'Riego', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (2, 'tipo_actividad', 'poda', 'Poda', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (3, 'tipo_actividad', 'limpieza', 'Limpieza', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (5, 'tipo_actividad', 'inspeccion', 'Inspección', true, 5, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (7, 'estado', 'en_proceso', 'En proceso', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (11, 'prioridad', 'baja', 'Baja', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (12, 'prioridad', 'media', 'Media', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (13, 'prioridad', 'alta', 'Alta', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (14, 'lugar', 'norte', 'Sector norte', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (15, 'lugar', 'sur', 'Sector sur', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (16, 'lugar', 'eje', 'Eje central', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (17, 'lugar', 'bosque', 'Bosque húmedo', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (18, 'especie', 'tipuana', 'Tipuana tipu', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (19, 'especie', 'ficus', 'Ficus benjamina', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (20, 'especie', 'schinus', 'Schinus molle', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (21, 'motivo_archivo', 'error', 'Registrada por error', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (22, 'motivo_archivo', 'duplicada', 'Duplicada', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (23, 'motivo_archivo', 'no_corresponde', 'Ya no corresponde', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (24, 'motivo_archivo', 'otro', 'Otro', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (25, 'turno', 'manana', 'Mañana', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (26, 'turno', 'tarde', 'Tarde', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (27, 'fuente', 'centuria', 'Centuria', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (28, 'fuente', 'osg', 'Matriz OSG', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (29, 'fuente', 'correo', 'Correo', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (30, 'fuente', 'interna', 'Interna', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (31, 'tipo_vegetacion', 'arbol', 'Árbol', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (32, 'tipo_vegetacion', 'palmera', 'Palmera', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (33, 'tipo_vegetacion', 'arbusto', 'Arbusto', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (34, 'tipo_vegetacion', 'herbacea', 'Herbácea', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (35, 'tipo_vegetacion', 'trepadora', 'Trepadora', true, 5, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (36, 'tipo_vegetacion', 'suculenta', 'Suculenta', true, 6, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (37, 'tipo_vegetacion', 'cafeto', 'cafeto', true, 7, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (38, 'estado', 'sin_estado', 'Sin estado', true, 0, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (39, 'clase_actividad', 'habilitacion', 'Habilitación de jardines', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (40, 'clase_actividad', 'rehabilitacion', 'Rehabilitación y rediseño de jardines', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (41, 'clase_actividad', 'mantenimiento', 'Mantenimiento de jardines', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (42, 'clase_actividad', 'poda', 'Poda', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (43, 'clase_actividad', 'propagacion', 'Propagación y plantación', true, 5, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (44, 'clase_actividad', 'riego', 'Riego', true, 6, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (45, 'clase_actividad', 'residuos', 'Manejo de residuos vegetales', true, 7, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (6, 'estado', 'pendiente', 'Por iniciar', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (8, 'estado', 'bloqueada', 'Bloqueada', false, 90, true, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (9, 'estado', 'cerrada', 'Cerrado', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (10, 'estado', 'cancelada', 'Cancelado', true, 5, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (46, 'estado', 'ejecutado', 'Ejecutado', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (47, 'estado', 'archivada', 'Archivado', true, 6, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (48, 'clase_actividad', 'fitosanitario', 'Manejo fitosanitario', true, 8, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (49, 'clase_actividad', 'inspeccion_monitoreo', 'Inspección y monitoreo', true, 9, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (50, 'plaga', 'demo_hoja', 'Plaga de demostración', true, 1, true, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (51, 'producto_fitosanitario', 'demo_producto', 'Producto fitosanitario de demostración', true, 1, true, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (52, 'frecuencia', 'semanal', 'Semanal', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (53, 'frecuencia', 'quincenal', 'Quincenal', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (54, 'sede', 'sede_demo', 'Sede de demostración', true, 1, true, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (55, 'cuartel', 'sin_archivo', 'Sin archivo de cuarteles', true, 1, true, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (56, 'sector_capataz', 'sector_norte', 'Sector norte (ficticio)', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (57, 'sector_capataz', 'campo_deportivo', 'Campo deportivo', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (58, 'sector_capataz', 'bosque_humedo', 'Bosque húmedo', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (4, 'tipo_actividad', 'incidencia', 'Novedad de campo', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (59, 'nivel_riesgo', 'bajo', 'Bajo', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (60, 'nivel_riesgo', 'alto', 'Alto', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (61, 'subtipo_actividad', 'preparacion_terreno', 'Preparación del terreno', true, 1, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (62, 'subtipo_actividad', 'incorporacion_sustrato', 'Incorporación de sustrato', true, 2, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (63, 'subtipo_actividad', 'instalacion_plantas', 'Instalación de plantas', true, 3, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (64, 'subtipo_actividad', 'instalacion_cesped', 'Instalación de césped', true, 4, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (65, 'subtipo_actividad', 'cobertura_ornamental', 'Colocación de cobertura ornamental', true, 5, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (66, 'subtipo_actividad', 'instalacion_tutores', 'Instalación de tutores', true, 6, false, 'habilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (67, 'subtipo_actividad', 'reposicion_plantas', 'Reposición de plantas', true, 1, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (68, 'subtipo_actividad', 'recuperacion_areas', 'Recuperación de áreas verdes', true, 2, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (69, 'subtipo_actividad', 'renovacion_jardineras', 'Renovación de jardineras', true, 3, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (70, 'subtipo_actividad', 'mejoramiento_suelo', 'Mejoramiento del suelo', true, 4, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (71, 'subtipo_actividad', 'renovacion_cobertura', 'Renovación de cobertura ornamental', true, 5, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (72, 'subtipo_actividad', 'resiembra', 'Resiembra', true, 6, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (73, 'subtipo_actividad', 'reubicacion_macetas', 'Reubicación de macetas', true, 7, false, 'rehabilitacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (74, 'subtipo_actividad', 'canteo', 'Canteo', true, 1, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (75, 'subtipo_actividad', 'deshierbo', 'Deshierbo', true, 2, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (76, 'subtipo_actividad', 'escarda', 'Escarda', true, 3, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (77, 'subtipo_actividad', 'limpieza_hojarasca', 'Limpieza de hojarasca', true, 4, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (78, 'subtipo_actividad', 'limpieza_plantas', 'Limpieza de plantas', true, 5, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (79, 'subtipo_actividad', 'aireacion_suelo', 'Aireación del suelo', true, 6, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (80, 'subtipo_actividad', 'limpieza_jardineras', 'Limpieza integral de jardineras', true, 7, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (81, 'subtipo_actividad', 'fertilizacion', 'Fertilización', true, 8, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (82, 'subtipo_actividad', 'aplicacion_enmiendas', 'Aplicación de enmiendas', true, 9, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (83, 'subtipo_actividad', 'traslado_macetas', 'Traslado de macetas', true, 10, false, 'mantenimiento');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (84, 'subtipo_actividad', 'poda_mantenimiento', 'Poda de mantenimiento', true, 1, false, 'poda');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (85, 'subtipo_actividad', 'poda_formacion', 'Poda de formación', true, 2, false, 'poda');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (86, 'subtipo_actividad', 'poda_sanitaria', 'Poda sanitaria', true, 3, false, 'poda');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (87, 'subtipo_actividad', 'poda_despeje', 'Poda de despeje/reducción', true, 4, false, 'poda');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (88, 'subtipo_actividad', 'division_matas', 'Propagación por división de matas', true, 1, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (89, 'subtipo_actividad', 'esquejes', 'Propagación por esquejes', true, 2, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (90, 'subtipo_actividad', 'trasplante', 'Trasplante', true, 3, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (91, 'subtipo_actividad', 'siembra_plantas', 'Siembra de plantas', true, 4, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (92, 'subtipo_actividad', 'plantacion_arboles', 'Plantación de árboles', true, 5, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (93, 'subtipo_actividad', 'plantacion_arbustos', 'Plantación de arbustos', true, 6, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (94, 'subtipo_actividad', 'plantacion_cubresuelos', 'Plantación de cubresuelos', true, 7, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (95, 'subtipo_actividad', 'plantacion_macetas', 'Plantación en macetas', true, 8, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (96, 'subtipo_actividad', 'trasplante_macetas', 'Trasplante a macetas', true, 9, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (97, 'subtipo_actividad', 'cambio_maceta', 'Cambio de maceta', true, 10, false, 'propagacion');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (98, 'subtipo_actividad', 'riego_manual', 'Riego manual', true, 1, false, 'riego');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (99, 'subtipo_actividad', 'riego_nocturno', 'Riego nocturno', true, 2, false, 'riego');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (100, 'subtipo_actividad', 'riego_establecimiento', 'Riego de establecimiento', true, 3, false, 'riego');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (101, 'subtipo_actividad', 'verificacion_riego', 'Verificación del sistema de riego', true, 4, false, 'riego');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (102, 'subtipo_actividad', 'recoleccion_hojarasca', 'Recolección de hojarasca', true, 1, false, 'residuos');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (103, 'subtipo_actividad', 'recoleccion_ramas', 'Recolección de ramas', true, 2, false, 'residuos');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (104, 'subtipo_actividad', 'triturado_residuos', 'Triturado de residuos', true, 3, false, 'residuos');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (105, 'subtipo_actividad', 'aprovechamiento_residuos', 'Disposición o aprovechamiento de residuos', true, 4, false, 'residuos');
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (106, 'estado_solicitud', 'por_iniciar', 'Por iniciar', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (107, 'estado_solicitud', 'en_proceso', 'En proceso', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (108, 'estado_solicitud', 'ejecutado', 'Ejecutado', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (109, 'estado_solicitud', 'cerrado', 'Cerrado', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (110, 'estado_solicitud', 'cancelado', 'Cancelado', true, 5, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (111, 'empresa', 'taller_verde_andino', 'Taller Verde Andino (ficticia)', true, 1, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (112, 'empresa', 'jardines_del_rimac', 'Jardines del Rímac (ficticia)', true, 2, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (113, 'empresa', 'poda_litoral', 'Poda Litoral (ficticia)', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (114, 'empresa', 'servicios_forestales_demo', 'Servicios Forestales del Perú', true, 4, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (115, 'frecuencia', 'mensual', 'Mensual', true, 3, false, NULL);
INSERT INTO public.catalogos (id, clase, codigo, nombre, activo, orden, provisional, padre_codigo) VALUES (116, 'frecuencia', 'unica', 'Única', true, 4, false, NULL);
INSERT INTO public.ordenes_servicio (id, actividad_id, empresa, referencia, frecuencia, estado, conformidad, created_at, periodo_inicio, periodo_fin, reporte_proveedor, empresa_id, frecuencia_id) VALUES ('99999999-9999-4999-8999-999999999991', '99999999-9999-4999-8999-999999999999', 'Jardines del Rímac S.A.C.', 'OS-2026-018', 'quincenal', 'en_proceso', '', '2026-10-03 04:52:21.450843+00', NULL, NULL, '', NULL, NULL);
INSERT INTO public.solicitudes (id, codigo_externo, fuente, titulo, detalle, prioridad, estado, lugar, cantidad, actividad_id, created_at, updated_at, origen_ref, archivada_en, cantidad_solicitada, cantidad_ejecutada, lugar_id, lat, lon) VALUES ('77777777-7777-4777-8777-777777777777', 'OSG-2026-0142', 'osg', 'Fuga reportada en línea de riego', 'Aviso manual. No hay integración con Centuria.', 'alta', 'en_proceso', 'Eje central', NULL, '44444444-4444-4444-8444-444444444444', '2026-10-03 04:52:21.419737+00', '2026-10-03 04:52:21.419737+00', NULL, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.roles (codigo, nombre, orden, descripcion, activo) VALUES ('capataz', 'Capataz', 1, NULL, true);
INSERT INTO public.roles (codigo, nombre, orden, descripcion, activo) VALUES ('admin', 'Administrador del sistema', 4, NULL, true);
INSERT INTO public.roles (codigo, nombre, orden, descripcion, activo) VALUES ('coordinacion', 'Ingeniería / Coordinación', 2, NULL, true);
INSERT INTO public.roles (codigo, nombre, orden, descripcion, activo) VALUES ('jefatura', 'Jefatura de sección', 3, NULL, true);
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'evidencias');
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'usuarios');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'usuarios');
INSERT INTO public.permisos (rol, accion) VALUES ('coordinacion', 'consultar');
INSERT INTO public.permisos (rol, accion) VALUES ('coordinacion', 'registrar');
INSERT INTO public.permisos (rol, accion) VALUES ('coordinacion', 'validar');
INSERT INTO public.permisos (rol, accion) VALUES ('coordinacion', 'solicitudes');
INSERT INTO public.permisos (rol, accion) VALUES ('coordinacion', 'reportes');
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'consultar');
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'validar');
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'reportes');
INSERT INTO public.permisos (rol, accion) VALUES ('jefatura', 'solicitudes');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'consultar');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'registrar');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'validar');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'reportes');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'catalogos');
INSERT INTO public.permisos (rol, accion) VALUES ('admin', 'solicitudes');
INSERT INTO public.permisos (rol, accion) VALUES ('capataz', 'consultar');
INSERT INTO public.permisos (rol, accion) VALUES ('capataz', 'registrar');
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-elsa', 'Elsa Mamani', true);
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-marco', 'Marco Huamán', true);
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-nelida', 'Nélida Choque', true);
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-pedro', 'Pedro Salas', true);
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-rita', 'Rita Ccopa', true);
INSERT INTO public.personal_ficticio (id, nombre_ficticio, activo) VALUES ('pf-tomas', 'Tomás Alegre', true);
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (2, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (4, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (12, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (13, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (14, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (23, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (24, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (25, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (26, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (27, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (28, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (29, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (30, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (31, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (32, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (33, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (34, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (35, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (36, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (37, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (38, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (39, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (40, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (41, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (42, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (43, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (44, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (45, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (46, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (57, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (58, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (59, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (60, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (61, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (62, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (63, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (64, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (65, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (66, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (67, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (68, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (69, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (70, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (71, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (72, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (73, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (74, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (75, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (76, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (77, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (78, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (79, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (80, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (81, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (82, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (83, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (84, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (85, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (86, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (87, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (88, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (89, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (90, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (91, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (92, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (93, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (94, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (95, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (96, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (97, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (98, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (99, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (100, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (101, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (102, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (103, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (104, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (105, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (106, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (107, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (108, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (109, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (110, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (111, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (112, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (113, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (114, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (115, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (116, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (117, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (118, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (119, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (120, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (121, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (122, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (123, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (124, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (125, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (126, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (127, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (128, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (129, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (130, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (131, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (132, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (133, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (134, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (135, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (136, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (182, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (183, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (191, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (193, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (204, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (205, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (206, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (207, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (208, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (231, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (255, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (270, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (275, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (276, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (277, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (278, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (279, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (280, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (281, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (288, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (292, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (293, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (294, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (295, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (305, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (306, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (307, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (333, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (334, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (335, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (336, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (337, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (338, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (339, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (340, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (341, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (342, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (343, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (344, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (345, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (346, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (347, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (348, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (349, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (350, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (351, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (352, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (353, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (354, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (355, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (356, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (384, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (386, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (387, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (388, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (389, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (390, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (391, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (392, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (393, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (394, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (395, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (398, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (399, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (400, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (401, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (402, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (403, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (404, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (405, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (406, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (407, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (408, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (409, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (410, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (412, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (422, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (423, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (424, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (425, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (426, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (427, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (428, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (429, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (430, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (431, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (432, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (433, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (434, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (435, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (436, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (437, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (438, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (439, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (440, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (441, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (442, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (443, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (444, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (445, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (446, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (447, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (450, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (451, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (452, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (453, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (455, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (456, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (457, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (458, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (459, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (460, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (461, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (462, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (463, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (464, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (465, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (466, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (467, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (468, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (484, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (485, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (486, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (487, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (488, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (489, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (490, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (491, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (495, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (496, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (497, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (503, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (504, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (505, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (506, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (507, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (508, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (509, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (510, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (514, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (515, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (518, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (519, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (520, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (521, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (522, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (529, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (530, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (532, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (533, 'cua-valeria');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (1, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (3, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (5, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (6, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (8, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (9, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (10, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (11, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (15, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (16, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (17, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (18, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (19, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (20, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (21, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (22, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (47, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (48, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (49, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (50, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (52, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (53, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (54, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (55, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (137, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (138, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (139, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (140, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (141, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (142, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (143, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (144, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (145, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (146, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (147, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (148, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (149, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (150, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (151, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (152, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (153, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (154, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (155, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (156, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (157, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (158, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (159, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (160, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (161, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (162, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (163, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (164, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (165, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (166, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (167, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (168, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (169, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (170, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (171, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (172, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (173, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (174, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (175, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (176, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (177, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (211, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (212, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (213, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (214, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (215, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (216, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (217, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (232, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (233, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (234, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (235, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (236, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (237, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (238, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (239, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (240, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (241, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (242, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (243, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (266, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (271, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (274, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (282, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (283, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (284, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (296, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (297, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (298, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (299, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (300, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (301, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (302, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (303, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (304, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (357, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (358, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (359, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (360, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (361, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (362, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (363, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (364, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (365, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (366, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (367, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (368, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (369, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (370, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (371, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (372, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (373, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (374, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (375, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (376, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (377, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (378, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (379, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (381, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (382, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (383, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (385, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (396, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (397, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (411, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (413, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (414, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (415, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (416, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (417, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (418, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (419, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (420, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (421, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (448, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (449, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (454, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (469, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (470, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (471, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (472, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (473, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (474, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (475, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (476, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (477, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (478, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (479, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (480, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (481, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (482, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (483, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (492, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (493, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (494, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (512, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (513, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (516, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (517, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (523, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (524, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (525, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (526, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (527, 'cua-mateo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (7, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (51, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (56, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (178, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (179, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (180, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (181, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (184, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (185, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (186, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (187, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (188, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (189, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (190, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (192, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (194, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (195, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (196, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (197, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (198, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (199, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (200, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (201, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (202, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (203, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (209, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (210, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (218, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (219, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (220, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (221, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (222, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (223, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (224, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (225, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (226, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (227, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (228, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (229, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (230, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (244, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (245, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (246, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (247, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (248, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (249, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (250, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (251, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (252, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (253, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (254, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (256, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (257, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (258, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (259, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (260, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (261, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (262, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (263, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (264, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (265, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (267, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (268, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (269, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (272, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (273, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (285, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (286, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (287, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (289, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (290, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (291, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (308, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (309, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (310, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (311, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (312, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (313, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (314, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (315, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (316, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (317, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (318, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (319, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (320, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (321, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (322, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (323, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (324, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (325, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (326, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (327, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (328, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (329, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (330, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (331, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (332, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (380, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (498, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (499, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (500, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (511, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (528, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (531, 'cua-renato');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (501, 'campo-deportivo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (502, 'campo-deportivo');
INSERT INTO public.poligonos_sector_ref (source_index, sector) VALUES (0, 'bosque-humedo');
INSERT INTO public.sectores_capataz (id, codigo, nombre, color, activo, created_at, updated_at) VALUES (1, 'cua-valeria', 'Sector de capataz — Valeria Quispe (ficticio)', '#6b5596', true, '2026-10-03 04:52:21.784686+00', '2026-10-03 04:52:21.784686+00');
INSERT INTO public.sectores_capataz (id, codigo, nombre, color, activo, created_at, updated_at) VALUES (2, 'cua-mateo', 'Sector de capataz — Mateo Salazar (ficticio)', '#3f73b0', true, '2026-10-03 04:52:21.784686+00', '2026-10-03 04:52:21.784686+00');
INSERT INTO public.sectores_capataz (id, codigo, nombre, color, activo, created_at, updated_at) VALUES (3, 'cua-renato', 'Sector de capataz — Renato Cárdenas (ficticio)', '#c27c2c', true, '2026-10-03 04:52:21.784686+00', '2026-10-03 04:52:21.784686+00');
INSERT INTO public.sectores_capataz (id, codigo, nombre, color, activo, created_at, updated_at) VALUES (4, 'campo-deportivo', 'Campo deportivo', '#9aab3e', true, '2026-10-03 04:52:21.784686+00', '2026-10-03 04:52:21.784686+00');
INSERT INTO public.sectores_capataz (id, codigo, nombre, color, activo, created_at, updated_at) VALUES (5, 'bosque-humedo', 'Bosque húmedo', '#3f9a82', true, '2026-10-03 04:52:21.784686+00', '2026-10-03 04:52:21.784686+00');
INSERT INTO public.riego_registros (id, sector, turno, capataz_id, fecha, nota, created_at, zona_supervision_id, ciclo, superficie_m2, sector_id) VALUES ('88888888-8888-4888-8888-888888888881', 'Eje central', 'manana', 'cap-norte', '2026-09-24', 'Aspersores del eje.', '2026-10-03 04:52:21.450843+00', NULL, '', NULL, NULL);
INSERT INTO public.riego_registros (id, sector, turno, capataz_id, fecha, nota, created_at, zona_supervision_id, ciclo, superficie_m2, sector_id) VALUES ('88888888-8888-4888-8888-888888888882', 'Bosque húmedo', 'tarde', 'cap-riego', '2026-09-24', 'Ronda de la tarde.', '2026-10-03 04:52:21.450843+00', NULL, '', NULL, NULL);
INSERT INTO public.vivero_catalogo (id, clase, nombre, activo) VALUES (1, 'area', 'Fauna', true);
INSERT INTO public.vivero_catalogo (id, clase, nombre, activo) VALUES (2, 'area', 'Flora', true);
INSERT INTO public.vivero_catalogo (id, clase, nombre, activo) VALUES (3, 'area', 'Ambiental', true);
INSERT INTO public.vivero_catalogo (id, clase, nombre, activo) VALUES (4, 'area', 'Otros', true);
SELECT pg_catalog.setval('public.actividad_eventos_id_seq', 13, true);
SELECT pg_catalog.setval('public.areas_verdes_id_seq', 1, false);
SELECT pg_catalog.setval('public.asignaciones_poligono_id_seq', 1, false);
SELECT pg_catalog.setval('public.bebederos_id_seq', 1, false);
SELECT pg_catalog.setval('public.cambios_id_seq', 43, true);
SELECT pg_catalog.setval('public.capas_auxiliares_id_seq', 1, false);
SELECT pg_catalog.setval('public.catalogos_id_seq', 116, true);
SELECT pg_catalog.setval('public.codigos_historicos_id_seq', 1, false);
SELECT pg_catalog.setval('public.cuarteles_historico_id_seq', 1, false);
SELECT pg_catalog.setval('public.ejemplares_id_seq', 1, false);
SELECT pg_catalog.setval('public.especies_id_seq', 1, false);
SELECT pg_catalog.setval('public.fauna_id_seq', 1, false);
SELECT pg_catalog.setval('public.inventario_id_seq', 1, false);
SELECT pg_catalog.setval('public.jardines_reserva_id_seq', 1, false);
SELECT pg_catalog.setval('public.lotes_importacion_id_seq', 1, false);
SELECT pg_catalog.setval('public.lugares_id_seq', 1, false);
SELECT pg_catalog.setval('public.playas_estacionamiento_id_seq', 1, false);
SELECT pg_catalog.setval('public.poligonos_cuadrilla_id_seq', 1, false);
SELECT pg_catalog.setval('public.puertas_id_seq', 1, false);
SELECT pg_catalog.setval('public.puntos_pucp_id_seq', 1, false);
SELECT pg_catalog.setval('public.referentes_edificio_id_seq', 1, false);
SELECT pg_catalog.setval('public.reservas_jardin_id_seq', 1, false);
SELECT pg_catalog.setval('public.sectores_capataz_id_seq', 5, true);
SELECT pg_catalog.setval('public.tachos_id_seq', 1, false);
SELECT pg_catalog.setval('public.usuarios_id_seq', 6, true);
SELECT pg_catalog.setval('public.veredas_riesgo_id_seq', 1, false);
SELECT pg_catalog.setval('public.vias_id_seq', 1, false);
SELECT pg_catalog.setval('public.vivero_catalogo_id_seq', 4, true);
SELECT pg_catalog.setval('public.xerofiticas_id_seq', 1, false);
SELECT pg_catalog.setval('public.zonas_id_seq', 1, false);
SELECT pg_catalog.setval('public.zonas_supervision_id_seq', 1, false);
SELECT pg_catalog.set_config('search_path', 'public', true);
