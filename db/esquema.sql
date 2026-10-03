-- Esquema final generado. No editar a mano.
-- Lo escribe scripts/generar-esquema-bd.sh después de aplicar db/migrations
-- con cmd/migrate sobre un PostGIS vacío.
-- La fuente de verdad son esas migraciones. Este archivo es la foto del
-- resultado, sin dueños ni privilegios, para leerla y para el control de deriva.
-- Si difiere de un volcado nuevo, hay que regenerarlo con el mismo script.

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
-- Registro de migraciones
-- ========================================================================
-- schema_migrations
CREATE TABLE public.schema_migrations (
    version text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);
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
