-- =====================================================================
-- Modelo de Datos v1.1 — Sistema de gestión de áreas verdes (PUCP)
-- Grupo 12 · 1INF47 · 2026-09-29 · PostgreSQL 16 + PostGIS 3.5 (imagen postgis/postgis:16-3.5)
-- Base: Catálogo de Requisitos v1.3 (decisiones del equipo del 29/09/2026).
-- Supera a db/schema_v1.0.sql (42 tablas) y a db/schema_nucleo_v0.2.sql.
--
-- 26 tablas + 2 vistas + 1 función.
-- Cambios respecto de v1.0:
--   - Permisos: se eliminan permiso y rol_permiso. usuario → rol; la matriz de
--     acciones por rol la define el equipo en el backend, que valida cada
--     operación (RF-02/03, RNF-04).
--   - Sin registro individual de operarios: se eliminan personal e
--     intervencion_personal (RF-08).
--   - Sin historial de asignaciones: se elimina sector_asignacion; solo
--     sector.capataz_id (RF-34).
--   - Cinco ámbitos en sector: 3 sectores operativos con capataz, bosque húmedo
--     (sin capataz fijo) y canchas concesionadas (gestion_propia = false) (RF-06/34).
--   - Dos flujos separados: intervencion = solo trabajo propio (con
--     actividad_evento); servicio_tercerizado = trabajo tercerizado con su
--     verificación y conformidad. intervencion pierde ejecutor y servicio_id (RF-09/13).
--   - Jardín de préstamo = área verde con nombre: sus atributos pasan de lugar a
--     area_verde, y reserva_jardin apunta a area_verde (RF-06/36).
--   - La pertenencia a un ámbito se deduce de la geometría con ambito_de(); se
--     elimina sector_id de area_verde, lugar, ejemplar, solicitud, servicio e intervención.
--   - Solo geom: se eliminan latitud/longitud duplicadas; superficie calculada.
--   - Tablas eliminadas por no tener requisito o duplicar otra: sede, cuartel
--     (capa GeoJSON fija fuera de la BD), prioridad, plaga, producto_fitosanitario,
--     plaga_especie, producto_plaga, punto_acopio, ficha_tecnica_poda (→ atributos),
--     checklist_item, referencia_externa.
--   - solicitud sin impacto, prioridad ni sector; evidencia con un único destino.
-- Las FK van al final para evitar dependencias de orden.
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid()

-- =====================================================================
-- A. Acceso y roles — Épica 1
-- =====================================================================

-- rol: Perfil de acceso (Jefatura, Ingeniería/Coordinación, Capataz). Solo semilla: lo que puede hacer cada rol lo define el backend (RF-02/03).
CREATE TABLE rol (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(50) NOT NULL UNIQUE,
  descripcion text,
  activo boolean NOT NULL DEFAULT true
);

-- usuario: Cuenta con credenciales propias; la crea, activa y desactiva la Jefatura. Un rol por usuario; sin SSO (RF-01/02).
CREATE TABLE usuario (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombres varchar(120) NOT NULL,
  apellidos varchar(120) NOT NULL,
  email varchar(120) UNIQUE,
  username varchar(60) NOT NULL UNIQUE,
  password_hash varchar(255) NOT NULL,
  rol_id bigint NOT NULL,
  telefono varchar(20),
  debe_cambiar_password boolean NOT NULL DEFAULT true,
  activo boolean NOT NULL DEFAULT true,     -- el backend lo revalida en cada operación (RNF-04)
  ultimo_acceso timestamptz,
  creado_en timestamptz NOT NULL DEFAULT now()
);

-- auditoria: Bitácora de operaciones críticas: login, alta/desactivación de usuarios, bajas, cambios de catálogos y conformidad de servicios (RNF-04).
CREATE TABLE auditoria (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  usuario_id bigint,
  entidad varchar(60) NOT NULL,
  entidad_id bigint,
  accion varchar(20) NOT NULL,             -- crear | editar | baja | login | desactivar | conformar
  detalle jsonb,
  ip varchar(45),
  creado_en timestamptz NOT NULL DEFAULT now()
);

-- =====================================================================
-- B. Geografía y catastro — Épica 2
-- =====================================================================

-- tipo_uso: Clasificación de las áreas verdes por uso; los 6 valores reales del cliente (RF-06).
CREATE TABLE tipo_uso (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(80) NOT NULL UNIQUE,
  descripcion text,
  activo boolean NOT NULL DEFAULT true
);

-- sector: Ámbito del campus como polígono. Tres sectores operativos (con capataz), bosque húmedo (gestión propia sin capataz fijo) y canchas concesionadas (sin gestión propia). Solo la asignación vigente (RF-06/34).
CREATE TABLE sector (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(80) NOT NULL UNIQUE,
  descripcion text,
  gestion_propia boolean NOT NULL DEFAULT true,  -- false: concesionado, la oficina no registra labores ahí
  capataz_id bigint,                        -- NULL: sin capataz fijo (se asigna por intervención)
  geom geometry(MultiPolygon,4326),
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT ck_sector_concesionado_sin_capataz CHECK (gestion_propia OR capataz_id IS NULL)
);

-- lugar: Punto de referencia con nombre único (edificio, facultad, puerta, vía). Vocabulario del catastro y de la operación (RF-06/18/34).
CREATE TABLE lugar (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(120) NOT NULL UNIQUE,
  tipo varchar(40),                         -- edificio | facultad | puerta | via | otro
  geom geometry(Point,4326),
  activo boolean NOT NULL DEFAULT true
);

-- area_verde: Polígono de área verde con su tipo de uso. Los jardines son áreas verdes con nombre; los de préstamo llevan dueño y periodo de recuperación (RF-04/06/36/37).
CREATE TABLE area_verde (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(120),                      -- solo los jardines tienen nombre
  codigo_cliente varchar(20),               -- código del cliente (p. ej. 'B 4'); no es único
  tipo_uso_id bigint,
  es_prestamo boolean NOT NULL DEFAULT false,
  dueno varchar(10),                        -- DAF | UNIDAD (jardines de préstamo)
  periodo_recuperacion_dias integer,        -- descanso tras un préstamo (RF-37)
  estado varchar(40),
  geom geometry(MultiPolygon,4326) NOT NULL,
  superficie_m2 numeric(12,2) GENERATED ALWAYS AS (round(ST_Area(geom::geography)::numeric, 2)) STORED,
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT ck_area_prestamo_con_nombre CHECK (NOT es_prestamo OR nombre IS NOT NULL),
  CONSTRAINT ck_area_dueno CHECK (dueno IS NULL OR dueno IN ('DAF','UNIDAD')),
  CONSTRAINT ck_area_recuperacion CHECK (periodo_recuperacion_dias IS NULL OR periodo_recuperacion_dias >= 0)
);

-- especie: Catálogo de especies (nombre común y científico, forma biológica) (RF-04/24).
CREATE TABLE especie (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre_comun varchar(120) NOT NULL,
  nombre_cientifico varchar(160),
  tipo_vegetacion varchar(60),              -- Palmera, Árbol, Arbusto…
  condicion_relevante text,
  activo boolean NOT NULL DEFAULT true
);

-- ejemplar: Árbol, palmera o grupo catastrado, con código único, lugar de referencia, foto y dasometría cuando exista (RF-04/05).
CREATE TABLE ejemplar (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo_actual varchar(60) NOT NULL UNIQUE,
  codigo_anterior varchar(60),              -- recodificación (RF-05)
  especie_id bigint,
  estado_salud varchar(40),
  lugar_id bigint,                          -- columna 'Ubicación' del catastro
  ubicacion_referencia varchar(120),        -- referencia de punto del catastro (p. ej. p01)
  cantidad numeric(10,2) DEFAULT 1,
  diametro_dap numeric(8,2),                -- cm, pendiente de la ingeniera
  altura_m numeric(6,2),
  foto_url text,
  geom geometry(Point,4326),
  fecha_registro date NOT NULL DEFAULT CURRENT_DATE,
  activo boolean NOT NULL DEFAULT true
);

-- =====================================================================
-- C. Catálogos de actividad y de solicitudes — Épicas 3, 5 y 7
-- =====================================================================

-- clase_actividad: Primer nivel de la taxonomía (eje), compartida por intervenciones y servicios (RF-08).
CREATE TABLE clase_actividad (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(80) NOT NULL UNIQUE,
  descripcion text,
  activo boolean NOT NULL DEFAULT true
);

-- tipo_actividad: Segundo nivel (subtipo): ícono del mapa, frecuencia de referencia y esquema de los campos específicos del tipo, p. ej. la ficha de seguridad de poda (RF-08/24/29).
CREATE TABLE tipo_actividad (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(100) NOT NULL,
  clase_actividad_id bigint NOT NULL,
  descripcion text,
  frecuencia_dias integer,                  -- referencial (p. ej. césped 30-45 días)
  ejecutor_default varchar(12),             -- propio | tercerizado: la interfaz propone intervención o servicio
  atributos_schema jsonb,                   -- esquema de los campos específicos del tipo
  icono varchar(60),
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT uq_tipo_actividad UNIQUE (clase_actividad_id, nombre),
  CONSTRAINT ck_tipo_ejecutor CHECK (ejecutor_default IS NULL OR ejecutor_default IN ('propio','tercerizado'))
);

-- estado_atencion: Estados con ciclo propio por tipo de registro. El código lo usa el backend (máquina de estados); el nombre y el color son configurables (RF-16/29).
CREATE TABLE estado_atencion (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  aplica_a varchar(20) NOT NULL,            -- intervencion | solicitud | servicio
  codigo varchar(30) NOT NULL,              -- estable: lo referencia el backend
  nombre varchar(60) NOT NULL,
  descripcion text,
  orden integer,
  es_final boolean NOT NULL DEFAULT false,
  color varchar(9),
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT uq_estado UNIQUE (aplica_a, codigo),
  CONSTRAINT ck_estado_aplica CHECK (aplica_a IN ('intervencion','solicitud','servicio'))
);

-- severidad: Alta/Media/Baja de la matriz OSG; también es el nivel de riesgo de la intervención (RF-11/30).
CREATE TABLE severidad (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(20) NOT NULL UNIQUE,
  nivel integer NOT NULL,
  activo boolean NOT NULL DEFAULT true
);

-- tipo_evento_incidencia: Incidencia, Reclamo, No conformidad, Hallazgo de supervisión o de auditoría (RF-11).
CREATE TABLE tipo_evento_incidencia (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(60) NOT NULL UNIQUE,
  activo boolean NOT NULL DEFAULT true
);

-- causa_raiz: Causas raíz de áreas verdes (AV-01…) del catálogo del cliente (RF-11/24).
CREATE TABLE causa_raiz (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo varchar(20) NOT NULL UNIQUE,
  descripcion text NOT NULL,
  activo boolean NOT NULL DEFAULT true
);

-- motivo_cancelacion: Lista predefinida de motivos de baja lógica (RF-31/34).
CREATE TABLE motivo_cancelacion (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(80) NOT NULL UNIQUE,
  activo boolean NOT NULL DEFAULT true
);

-- fuente_solicitud: Canal de ingreso de la solicitud (WhatsApp, correo, Centuria…) (RF-11/28).
CREATE TABLE fuente_solicitud (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(60) NOT NULL UNIQUE,
  descripcion text,
  activo boolean NOT NULL DEFAULT true
);

-- insumo: Consumibles del trabajo propio. Los productos de los proveedores no se registran aquí (RF-10).
CREATE TABLE insumo (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre varchar(120) NOT NULL,
  unidad_medida varchar(20) NOT NULL,
  tipo varchar(40),
  activo boolean NOT NULL DEFAULT true
);

-- =====================================================================
-- D. Trabajo tercerizado — Épica 4
-- =====================================================================

-- empresa: Proveedor tercerizado (RF-13).
CREATE TABLE empresa (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  razon_social varchar(160) NOT NULL,
  ruc varchar(11),
  contacto varchar(120),
  telefono varchar(20),
  email varchar(120),
  activo boolean NOT NULL DEFAULT true
);

-- servicio_tercerizado: Trabajo de un proveedor, separado de las intervenciones. Ciclo: orden de compra → ejecución → informado → verificación → conforme u observado. No registra la ejecución interna del proveedor (RF-09/13/14/16).
CREATE TABLE servicio_tercerizado (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id bigint NOT NULL,
  tipo_actividad_id bigint NOT NULL,
  solicitud_id bigint,                      -- solicitud de origen, si la hay
  referencia_oc varchar(60),                -- orden de compra (documento de Logística)
  requerimiento varchar(80),
  lugar_id bigint,                          -- lugar de referencia
  geom geometry(Point,4326),                -- pin en el mapa; NULL si abarca todo el campus
  fecha_inicio date,                        -- periodo planificado
  fecha_fin date,
  estado_id bigint NOT NULL,                -- estados con aplica_a = 'servicio'
  resultado_conformidad varchar(12),        -- conforme | observado | NULL (sin verificar)
  fecha_conformidad date,
  conformidad_por bigint,                   -- quién dio la conformidad (habilita el pago)
  reporte_url text,
  observaciones text,
  creado_por bigint,
  creado_en timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_servicio_fechas CHECK (fecha_fin IS NULL OR fecha_inicio IS NULL OR fecha_fin >= fecha_inicio),
  CONSTRAINT ck_servicio_resultado CHECK (resultado_conformidad IS NULL OR resultado_conformidad IN ('conforme','observado')),
  CONSTRAINT ck_servicio_conforme CHECK (resultado_conformidad IS DISTINCT FROM 'conforme'
                                         OR (fecha_conformidad IS NOT NULL AND conformidad_por IS NOT NULL))
);

-- =====================================================================
-- E. Solicitudes e incidencias — Épica 3
-- =====================================================================

-- solicitud: Solicitud o incidencia recibida (formato de la matriz OSG). Se atiende con una intervención propia o un servicio tercerizado (RF-11/16/28).
CREATE TABLE solicitud (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo_interno varchar(40) NOT NULL UNIQUE,
  codigo_externo varchar(80),               -- p. ej. OSG-0001; uno por solicitud; el sistema no lo genera
  fuente_id bigint NOT NULL,                -- canal de ingreso
  tipo_evento_id bigint,
  fecha_solicitud date NOT NULL,
  reportado_por varchar(160),
  unidad varchar(120),
  severidad_id bigint,
  causa_raiz_id bigint,
  tipo_actividad_id bigint,                 -- tipo de trabajo sugerido
  lugar_id bigint,                          -- lugar de referencia (el ámbito se deduce de él)
  descripcion text NOT NULL,
  cantidad_solicitada numeric(10,2),
  requiere_tercerizacion boolean NOT NULL DEFAULT false,
  estado_id bigint NOT NULL,                -- estados con aplica_a = 'solicitud'
  fecha_recepcion date,
  fecha_cierre date,
  observaciones text,
  creado_por bigint,
  creado_en timestamptz NOT NULL DEFAULT now(),
  actualizado_en timestamptz
);

-- =====================================================================
-- F. Trabajo propio: intervenciones y trazabilidad — Épicas 3, 5 y 9
-- =====================================================================

-- intervencion: Trabajo propio (la actividad): pin del mapa, capataz responsable, fechas planificadas y reales. Registrable sin conexión (uuid_cliente). Estado actual = copia del último evento (RF-08/16/29-32/38).
CREATE TABLE intervencion (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  uuid_cliente uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
  tipo_actividad_id bigint NOT NULL,
  solicitud_id bigint,                      -- si atiende una solicitud
  lugar_id bigint,                          -- lugar de referencia sugerido
  area_verde_id bigint,                     -- área trabajada, si aplica
  ejemplar_id bigint,                       -- árbol trabajado, si aplica
  descripcion text,
  capataz_id bigint,                        -- capataz responsable
  severidad_id bigint,                      -- nivel de riesgo (se hereda de la solicitud si la hay)
  geom geometry(Point,4326) NOT NULL,       -- ubicación marcada en el mapa (el ámbito se deduce de aquí)
  fecha_planificada_inicio date,
  fecha_planificada_fin date,
  fecha_inicio_real date,                   -- copia del evento 'inicio'
  fecha_fin_real date,                      -- copia del evento 'cierre'
  cantidad_ejecutada numeric(10,2),
  estado_id bigint NOT NULL,                -- estados con aplica_a = 'intervencion'; copia del último evento
  atributos jsonb,                          -- campos específicos del tipo (validados con tipo_actividad.atributos_schema)
  observaciones text,
  motivo_cancelacion_id bigint,             -- motivo de la baja lógica
  creado_por bigint,
  creado_en timestamptz NOT NULL DEFAULT now(),
  actualizado_en timestamptz,
  CONSTRAINT ck_interv_plan CHECK (fecha_planificada_fin IS NULL OR fecha_planificada_inicio IS NULL OR fecha_planificada_fin >= fecha_planificada_inicio),
  CONSTRAINT ck_interv_real CHECK (fecha_fin_real IS NULL OR fecha_inicio_real IS NULL OR fecha_fin_real >= fecha_inicio_real)
);

-- actividad_evento: Cadena cronológica de la intervención; lo que haga falta precisar va en la descripción. 'fecha' es la hora en que ocurrió (la envía el dispositivo) (RF-12/31).
CREATE TABLE actividad_evento (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  intervencion_id bigint NOT NULL,
  tipo_evento varchar(20) NOT NULL,
  estado_id bigint,                         -- estado resultante, si el evento lo cambia
  usuario_id bigint NOT NULL,
  descripcion text,
  fecha timestamptz NOT NULL DEFAULT now(),
  uuid_cliente uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
  CONSTRAINT ck_evento_tipo CHECK (tipo_evento IN ('registro','inicio','actualizacion','supervision','reasignacion','derivacion','cierre','baja'))
);

-- consumo_insumo: Insumos usados en una intervención propia; la cantidad depende de ambos (RF-10).
CREATE TABLE consumo_insumo (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  intervencion_id bigint NOT NULL,
  insumo_id bigint NOT NULL,
  cantidad numeric(10,2) NOT NULL,
  observacion text,
  CONSTRAINT ck_consumo_cantidad CHECK (cantidad > 0)
);

-- =====================================================================
-- G. Evidencias — transversal
-- =====================================================================

-- evidencia: Foto o documento asociado a UN solo registro: un evento de intervención, una solicitud o un servicio. Idempotente offline (RF-12/14/19).
CREATE TABLE evidencia (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  uuid_cliente uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
  evento_id bigint,
  solicitud_id bigint,
  servicio_id bigint,
  tipo varchar(20) NOT NULL DEFAULT 'foto',
  url_archivo text NOT NULL,
  latitud numeric(9,6),                     -- metadato de la foto (no hay geom aquí)
  longitud numeric(9,6),
  fecha timestamptz NOT NULL DEFAULT now(),
  creado_por bigint,
  CONSTRAINT ck_evidencia_un_destino CHECK (num_nonnulls(evento_id, solicitud_id, servicio_id) = 1),
  CONSTRAINT ck_evidencia_tipo CHECK (tipo IN ('foto','documento'))
);

-- =====================================================================
-- H. Reservas de jardines — módulo de reservas
-- =====================================================================

-- reserva_jardin: Reserva de un jardín de préstamo (área verde con es_prestamo). Solo Reservado/Cancelado; 'en uso', 'en recuperación' y 'disponible' se calculan (RF-36/37).
CREATE TABLE reserva_jardin (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  area_verde_id bigint NOT NULL,
  evento varchar(200) NOT NULL,
  fecha_inicio timestamptz NOT NULL,
  fecha_fin timestamptz NOT NULL,
  unidad_solicitante varchar(160),
  contacto_nombre varchar(160),             -- dato personal (RNF-04)
  contacto_telefono varchar(40),
  fecha_solicitud date,
  estado varchar(12) NOT NULL DEFAULT 'reservado',
  observacion text,
  creado_por bigint,
  creado_en timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_reserva_fechas CHECK (fecha_fin >= fecha_inicio),
  CONSTRAINT ck_reserva_estado CHECK (estado IN ('reservado','cancelado'))
);

-- =====================================================================
-- Claves foráneas
-- =====================================================================
ALTER TABLE usuario ADD CONSTRAINT fk_usuario_rol FOREIGN KEY (rol_id) REFERENCES rol(id);
ALTER TABLE auditoria ADD CONSTRAINT fk_auditoria_usuario FOREIGN KEY (usuario_id) REFERENCES usuario(id);
ALTER TABLE sector ADD CONSTRAINT fk_sector_capataz FOREIGN KEY (capataz_id) REFERENCES usuario(id);
ALTER TABLE area_verde ADD CONSTRAINT fk_area_tipouso FOREIGN KEY (tipo_uso_id) REFERENCES tipo_uso(id);
ALTER TABLE ejemplar ADD CONSTRAINT fk_ejemplar_especie FOREIGN KEY (especie_id) REFERENCES especie(id);
ALTER TABLE ejemplar ADD CONSTRAINT fk_ejemplar_lugar FOREIGN KEY (lugar_id) REFERENCES lugar(id);
ALTER TABLE tipo_actividad ADD CONSTRAINT fk_tipoact_clase FOREIGN KEY (clase_actividad_id) REFERENCES clase_actividad(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_empresa FOREIGN KEY (empresa_id) REFERENCES empresa(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_tipoact FOREIGN KEY (tipo_actividad_id) REFERENCES tipo_actividad(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_solicitud FOREIGN KEY (solicitud_id) REFERENCES solicitud(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_lugar FOREIGN KEY (lugar_id) REFERENCES lugar(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_estado FOREIGN KEY (estado_id) REFERENCES estado_atencion(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_conformidadpor FOREIGN KEY (conformidad_por) REFERENCES usuario(id);
ALTER TABLE servicio_tercerizado ADD CONSTRAINT fk_serv_creadopor FOREIGN KEY (creado_por) REFERENCES usuario(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_fuente FOREIGN KEY (fuente_id) REFERENCES fuente_solicitud(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_tipoevento FOREIGN KEY (tipo_evento_id) REFERENCES tipo_evento_incidencia(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_severidad FOREIGN KEY (severidad_id) REFERENCES severidad(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_causaraiz FOREIGN KEY (causa_raiz_id) REFERENCES causa_raiz(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_tipoact FOREIGN KEY (tipo_actividad_id) REFERENCES tipo_actividad(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_lugar FOREIGN KEY (lugar_id) REFERENCES lugar(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_estado FOREIGN KEY (estado_id) REFERENCES estado_atencion(id);
ALTER TABLE solicitud ADD CONSTRAINT fk_sol_creadopor FOREIGN KEY (creado_por) REFERENCES usuario(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_tipoact FOREIGN KEY (tipo_actividad_id) REFERENCES tipo_actividad(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_solicitud FOREIGN KEY (solicitud_id) REFERENCES solicitud(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_lugar FOREIGN KEY (lugar_id) REFERENCES lugar(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_area FOREIGN KEY (area_verde_id) REFERENCES area_verde(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_ejemplar FOREIGN KEY (ejemplar_id) REFERENCES ejemplar(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_capataz FOREIGN KEY (capataz_id) REFERENCES usuario(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_severidad FOREIGN KEY (severidad_id) REFERENCES severidad(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_estado FOREIGN KEY (estado_id) REFERENCES estado_atencion(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_motivobaja FOREIGN KEY (motivo_cancelacion_id) REFERENCES motivo_cancelacion(id);
ALTER TABLE intervencion ADD CONSTRAINT fk_int_creadopor FOREIGN KEY (creado_por) REFERENCES usuario(id);
ALTER TABLE actividad_evento ADD CONSTRAINT fk_evt_interv FOREIGN KEY (intervencion_id) REFERENCES intervencion(id);
ALTER TABLE actividad_evento ADD CONSTRAINT fk_evt_estado FOREIGN KEY (estado_id) REFERENCES estado_atencion(id);
ALTER TABLE actividad_evento ADD CONSTRAINT fk_evt_usuario FOREIGN KEY (usuario_id) REFERENCES usuario(id);
ALTER TABLE consumo_insumo ADD CONSTRAINT fk_consumo_interv FOREIGN KEY (intervencion_id) REFERENCES intervencion(id);
ALTER TABLE consumo_insumo ADD CONSTRAINT fk_consumo_insumo FOREIGN KEY (insumo_id) REFERENCES insumo(id);
ALTER TABLE evidencia ADD CONSTRAINT fk_evid_evento FOREIGN KEY (evento_id) REFERENCES actividad_evento(id);
ALTER TABLE evidencia ADD CONSTRAINT fk_evid_sol FOREIGN KEY (solicitud_id) REFERENCES solicitud(id);
ALTER TABLE evidencia ADD CONSTRAINT fk_evid_serv FOREIGN KEY (servicio_id) REFERENCES servicio_tercerizado(id);
ALTER TABLE evidencia ADD CONSTRAINT fk_evid_creadopor FOREIGN KEY (creado_por) REFERENCES usuario(id);
ALTER TABLE reserva_jardin ADD CONSTRAINT fk_reserva_area FOREIGN KEY (area_verde_id) REFERENCES area_verde(id);
ALTER TABLE reserva_jardin ADD CONSTRAINT fk_reserva_creadopor FOREIGN KEY (creado_por) REFERENCES usuario(id);

-- =====================================================================
-- Índices
-- =====================================================================
CREATE INDEX ix_intervencion_tipo      ON intervencion(tipo_actividad_id);
CREATE INDEX ix_intervencion_estado    ON intervencion(estado_id);
CREATE INDEX ix_intervencion_capataz   ON intervencion(capataz_id);
CREATE INDEX ix_intervencion_solicitud ON intervencion(solicitud_id);
CREATE INDEX ix_intervencion_ejemplar  ON intervencion(ejemplar_id);
CREATE INDEX ix_intervencion_area      ON intervencion(area_verde_id);
CREATE INDEX ix_intervencion_planfin   ON intervencion(fecha_planificada_fin);
CREATE INDEX ix_evento_interv          ON actividad_evento(intervencion_id, fecha);
CREATE INDEX ix_evento_usuario         ON actividad_evento(usuario_id);
CREATE INDEX ix_evidencia_evento       ON evidencia(evento_id);
CREATE INDEX ix_evidencia_solicitud    ON evidencia(solicitud_id);
CREATE INDEX ix_evidencia_servicio     ON evidencia(servicio_id);
CREATE INDEX ix_solicitud_estado       ON solicitud(estado_id);
CREATE INDEX ix_solicitud_severidad    ON solicitud(severidad_id);
CREATE INDEX ix_solicitud_lugar        ON solicitud(lugar_id);
CREATE INDEX ix_servicio_estado        ON servicio_tercerizado(estado_id);
CREATE INDEX ix_servicio_empresa       ON servicio_tercerizado(empresa_id);
CREATE INDEX ix_servicio_solicitud     ON servicio_tercerizado(solicitud_id);
CREATE INDEX ix_consumo_interv         ON consumo_insumo(intervencion_id);
CREATE INDEX ix_reserva_area_fechas    ON reserva_jardin(area_verde_id, fecha_inicio, fecha_fin);
-- Únicos parciales (catálogos controlados, RF-06)
CREATE UNIQUE INDEX ux_area_verde_nombre   ON area_verde(nombre) WHERE nombre IS NOT NULL;
CREATE UNIQUE INDEX ux_especie_cientifico  ON especie(nombre_cientifico) WHERE nombre_cientifico IS NOT NULL;
CREATE UNIQUE INDEX ux_empresa_ruc         ON empresa(ruc) WHERE ruc IS NOT NULL;
CREATE UNIQUE INDEX ux_solicitud_codext    ON solicitud(codigo_externo) WHERE codigo_externo IS NOT NULL;
-- Geoespaciales (GIST)
CREATE INDEX gx_sector_geom       ON sector USING GIST (geom);
CREATE INDEX gx_lugar_geom        ON lugar USING GIST (geom);
CREATE INDEX gx_area_geom         ON area_verde USING GIST (geom);
CREATE INDEX gx_ejemplar_geom     ON ejemplar USING GIST (geom);
CREATE INDEX gx_intervencion_geom ON intervencion USING GIST (geom);
CREATE INDEX gx_servicio_geom     ON servicio_tercerizado USING GIST (geom);

-- =====================================================================
-- Función y vistas derivadas
-- =====================================================================

-- ambito_de: ámbito (fila de sector) que contiene una geometría. Reemplaza a las columnas sector_id:
-- la pertenencia se deduce de la ubicación (RF-06) y sirve para sugerir ámbito y capataz (RF-34).
-- Para una solicitud se usa la geometría de su lugar.
CREATE FUNCTION ambito_de(g geometry) RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT s.id
  FROM sector s
  WHERE s.activo AND s.geom IS NOT NULL AND ST_Contains(s.geom, ST_PointOnSurface(g))
  ORDER BY s.id
  LIMIT 1
$$;

-- vw_actividad_retraso: marca de retraso de las intervenciones (RF-38). En retraso si el fin real
-- superó al planificado, o si sigue abierta y ya pasó el día de fin planificado. Excluye las bajas.
CREATE VIEW vw_actividad_retraso AS
SELECT i.id AS intervencion_id,
       i.tipo_actividad_id,
       i.capataz_id,
       i.estado_id,
       i.fecha_planificada_fin,
       i.fecha_fin_real,
       CASE
         WHEN i.fecha_planificada_fin IS NULL THEN false
         WHEN i.fecha_fin_real IS NOT NULL THEN (i.fecha_fin_real > i.fecha_planificada_fin)
         ELSE (CURRENT_DATE > i.fecha_planificada_fin)
       END AS en_retraso,
       CASE
         WHEN i.fecha_planificada_fin IS NULL THEN NULL
         WHEN i.fecha_fin_real IS NOT NULL THEN (i.fecha_fin_real - i.fecha_planificada_fin)
         ELSE (CURRENT_DATE - i.fecha_planificada_fin)
       END AS dias_retraso
FROM intervencion i
JOIN estado_atencion e ON e.id = i.estado_id
WHERE e.codigo NOT IN ('CANCELADO','ARCHIVADO');

-- vw_jardin_disponibilidad: situación de cada jardín de préstamo (RF-37): en uso (hay una reserva
-- vigente ahora), en recuperación (terminó hace menos días que su periodo) o disponible.
-- Solo cuentan reservas no canceladas que ya terminaron para calcular el descanso.
CREATE VIEW vw_jardin_disponibilidad AS
SELECT a.id AS area_verde_id,
       a.nombre,
       a.periodo_recuperacion_dias,
       ult.ultimo_uso,
       (ult.ultimo_uso::date + COALESCE(a.periodo_recuperacion_dias, 0)) AS recuperacion_hasta,
       CASE
         WHEN EXISTS (SELECT 1 FROM reserva_jardin r
                      WHERE r.area_verde_id = a.id AND r.estado = 'reservado'
                        AND now() BETWEEN r.fecha_inicio AND r.fecha_fin) THEN 'en_uso'
         WHEN ult.ultimo_uso IS NOT NULL
              AND CURRENT_DATE <= ult.ultimo_uso::date + COALESCE(a.periodo_recuperacion_dias, 0) THEN 'en_recuperacion'
         ELSE 'disponible'
       END AS situacion
FROM area_verde a
LEFT JOIN LATERAL (
  SELECT max(r.fecha_fin) AS ultimo_uso
  FROM reserva_jardin r
  WHERE r.area_verde_id = a.id AND r.estado = 'reservado' AND r.fecha_fin <= now()
) ult ON true
WHERE a.es_prestamo;

-- =====================================================================
-- Semillas mínimas (catálogos configurables)
-- =====================================================================
INSERT INTO rol (nombre, descripcion) VALUES
  ('Jefatura','Supervisión integral, reportes y administración de usuarios'),
  ('Ingeniería/Coordinación','Planificación, registro, solicitudes, servicios y catálogos'),
  ('Capataz','Registro en campo de las intervenciones a su cargo (sin conexión)');

-- Valores reales del atributo 'Uso' de la capa de áreas verdes del cliente.
INSERT INTO tipo_uso (nombre, descripcion) VALUES
  ('Áreas de uso administrativo','Entornos de edificios y áreas de uso administrativo'),
  ('Áreas de manejo sostenible y reducción de consumo de agua','Especies xerofíticas y de bajo consumo de agua'),
  ('Áreas de uso recreativo/descanso','Mayor carga de mantenimiento; incluye la mayoría de jardines de préstamo'),
  ('Áreas deportivas y recreación activa','Canchas deportivas (concesionadas)'),
  ('Uso Institucional','Valor paisajístico e institucional'),
  ('Áreas de conservación','Bosque húmedo y fauna');

INSERT INTO estado_atencion (aplica_a, codigo, nombre, descripcion, orden, es_final, color) VALUES
  ('intervencion','POR_INICIAR','Por iniciar','Creada y asignada, aún no iniciada',1,false,'#6B7280'),
  ('intervencion','EN_PROCESO','En proceso','En ejecución',2,false,'#2563EB'),
  ('intervencion','EJECUTADO','Ejecutado','Terminada, por cerrar con la evidencia del capataz',3,false,'#10B981'),
  ('intervencion','CERRADO','Cerrado','Cerrada (no aparece en el mapa)',4,true,'#374151'),
  ('intervencion','CANCELADO','Cancelado','Baja lógica con motivo',5,true,'#B23A3A'),
  ('intervencion','ARCHIVADO','Archivado','Baja lógica con motivo',6,true,'#9CA3AF'),
  ('servicio','REGISTRADO','Registrado','Registrado; orden de compra en trámite',1,false,'#6B7280'),
  ('servicio','EN_EJECUCION','En ejecución','El proveedor está ejecutando',2,false,'#2563EB'),
  ('servicio','INFORMADO','Informado','El proveedor informó que terminó; falta verificar',3,false,'#D97706'),
  ('servicio','OBSERVADO','Observado','Verificado con observaciones; el proveedor debe corregir',4,false,'#B23A3A'),
  ('servicio','CONFORME','Conforme','Conformidad otorgada; habilita el pago',5,true,'#10B981'),
  ('servicio','CANCELADO','Cancelado','Servicio anulado',6,true,'#9CA3AF'),
  ('solicitud','ABIERTO','Abierto','Recibida, sin atención iniciada',1,false,'#6B7280'),
  ('solicitud','EN_PROCESO','En proceso','En atención (intervención o servicio)',2,false,'#2563EB'),
  ('solicitud','CERRADO','Cerrado','Atendida',3,true,'#374151');

INSERT INTO severidad (nombre, nivel) VALUES ('Alta',1),('Media',2),('Baja',3);

INSERT INTO tipo_evento_incidencia (nombre) VALUES
  ('Incidencia'),('Reclamo'),('No conformidad'),('Hallazgo de supervisión'),('Hallazgo de auditoría');

INSERT INTO motivo_cancelacion (nombre) VALUES
  ('Registrada por error'),('Duplicada'),('Ya no corresponde'),('Otro');

INSERT INTO fuente_solicitud (nombre, descripcion) VALUES
  ('Centuria','Plataforma institucional (se conserva su código)'),
  ('Correo institucional','Solicitud por correo'),
  ('WhatsApp','Aviso operativo'),
  ('Supervisión','Detectada en campo'),
  ('Auditoría','Hallazgo de auditoría'),
  ('Encuesta','Reportada vía encuesta'),
  ('Presencial','Reportada presencialmente'),
  ('Formulario','Formulario web');

-- Muestra del catálogo de causas raíz de la matriz OSG (el catálogo completo se importa).
INSERT INTO causa_raiz (codigo, descripcion) VALUES
  ('AV-01','Caída y/o acumulación de hojas de palmera'),
  ('AV-02','Caída de ramas por sobrecrecimiento / falta de poda'),
  ('AV-03','Caída estacional de hojarasca en exceso'),
  ('AV-07','Césped deteriorado por sobrecarga de uso / alto tránsito'),
  ('AV-09','Aspersor mal colocado / rotación incorrecta (riego)'),
  ('AV-11','Árbol con falta de tutor / tallo inestable');

INSERT INTO clase_actividad (nombre, descripcion) VALUES
  ('Mantenimiento de jardines','Resiembra, fitodecoración, remoción, etc.'),
  ('Poda','Poda propia (baja, con escalera) y poda de altura (tercerizada)'),
  ('Control fitosanitario','Prevención y control de plagas (tercerizado)'),
  ('Corte de césped','Tercerizado, por frecuencia'),
  ('Riego','Riego por sectores (tipo de actividad, no módulo aparte)'),
  ('Propagación y plantación','Esquejes, siembra, plantación'),
  ('Rehabilitación y habilitación','Habilitación y rediseño de áreas'),
  ('Manejo de residuos vegetales','Retiro de residuos'),
  ('Extracción y traslado','Extracción, traslado y reubicación de arbolado (tercerizado)'),
  ('Inspección y monitoreo','Recorridos y evaluación');

-- Primer usuario (Jefatura): crearlo por migración con un hash real y debe_cambiar_password = true.
-- INSERT INTO usuario (nombres, apellidos, email, username, password_hash, rol_id)
--   VALUES ('Usuario','Ejemplo','usuario@ejemplo.com','uejemplo','<hash>', (SELECT id FROM rol WHERE nombre='Jefatura'));
-- Los cinco ámbitos (sector) se cargan con sus polígonos cuando el cliente confirme las geometrías.
