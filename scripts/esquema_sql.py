#!/usr/bin/env python3
"""Normaliza un pg_dump --schema-only y arma la baseline legible.

La entrada es el volcado de public tras aplicar db/migrations. Se quitan
dueños, privilegios, el token \\restrict y el esquema public (ya existe).
El orden queda fijo por dominio para que el diff de deriva no dependa
del orden interno de pg_dump.
"""

from __future__ import annotations

import re
import sys

PREAMBULO = """\
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
"""

# nombre interno, título de sección, tablas
DOMINIOS = [
    ("acceso", "Acceso", ["roles", "usuarios", "sesiones", "permisos"]),
    (
        "catastro",
        "Catastro",
        [
            "areas_verdes",
            "zonas_supervision",
            "zonas_origen",
            "lugares",
            "cuadrillas",
            "poligonos_cuadrilla",
            "poligonos_sector_ref",
            "asignaciones_poligono",
            "sectores_capataz",
            "vias",
            "cuarteles_historico",
            "referentes_edificio",
            "capas_auxiliares",
        ],
    ),
    (
        "operacion",
        "Operación",
        [
            "capataces",
            "actividades",
            "actividad_eventos",
            "actividad_avances",
            "personal_labor",
            "personal_ficticio",
            "podas",
            "vivero_catalogo",
            "vivero_registros",
            "riego_registros",
            "solicitudes",
            "ordenes_servicio",
        ],
    ),
    (
        "inventario",
        "Inventario",
        [
            "inventario",
            "especies",
            "ejemplares",
            "codigos_historicos",
            "medidas_palmera",
            "fauna",
            "puertas",
            "playas_estacionamiento",
            "veredas_riesgo",
            "xerofiticas",
            "jardines_reserva",
            "tachos",
            "bebederos",
            "puntos_pucp",
            "reservas_jardin",
        ],
    ),
    ("catalogos", "Catálogos", ["catalogos"]),
    ("auditoria", "Auditoría", ["cambios", "lotes_importacion"]),
    ("evidencias", "Evidencias", ["evidencias"]),
    ("interno", "Registro de migraciones", ["schema_migrations"]),
]

FUNCIONES = {
    "catastro_geom_4326": "catastro",
    "inventario_geom_4326": "inventario",
    "ejemplares_sector_cuartel_clase": "inventario",
}

VISTAS = {"zonas": "catastro"}

BLOQUE = re.compile(
    r"--\n-- Name: (?P<nombre>[^\n;]*); Type: (?P<tipo>[^\n;]*); Schema: [^\n;]*; Owner:[^\n]*\n--\n\n"
    r"(?P<cuerpo>.*?)(?=\n--\n-- Name: |\n--\n-- PostgreSQL database dump complete|\Z)",
    re.S,
)

TABLA_DE = {
    "SEQUENCE OWNED BY": re.compile(
        r"ALTER SEQUENCE public\.(\S+) OWNED BY public\.(\w+)\."
    ),
    "DEFAULT": re.compile(r"ALTER TABLE ONLY public\.(\w+) "),
    "INDEX": re.compile(r"\bON public\.(\w+)\b"),
    "CONSTRAINT": re.compile(r"ALTER TABLE ONLY public\.(\w+)\s"),
    "FK CONSTRAINT": re.compile(r"ALTER TABLE ONLY public\.(\w+)\s"),
    "TRIGGER": re.compile(r"\bON public\.(\w+)\b"),
}


def _dominio_de_tabla() -> dict[str, str]:
    out: dict[str, str] = {}
    for clave, _titulo, tablas in DOMINIOS:
        for tabla in tablas:
            out[tabla] = clave
    return out


def _cuerpo(texto: str) -> str:
    lineas = [ln.rstrip() for ln in texto.strip().splitlines()]
    return "\n".join(lineas).strip() + "\n"


def _parsear(dump: str) -> list[dict[str, str]]:
    bloques = []
    for m in BLOQUE.finditer(dump):
        bloques.append(
            {
                "nombre": m.group("nombre").strip(),
                "tipo": m.group("tipo").strip(),
                "cuerpo": _cuerpo(m.group("cuerpo")),
            }
        )
    if not bloques:
        raise SystemExit("el volcado no trae objetos de public")
    return bloques


def _nombre_funcion(cuerpo: str) -> str:
    m = re.search(r"CREATE FUNCTION public\.(\w+)\(", cuerpo)
    if not m:
        raise SystemExit(f"función sin nombre:\n{cuerpo[:200]}")
    return m.group(1)


def _tabla_de(tipo: str, cuerpo: str) -> str:
    rx = TABLA_DE.get(tipo)
    if rx is None:
        raise SystemExit(f"no sé extraer la tabla de un bloque {tipo}")
    m = rx.search(cuerpo)
    if not m:
        raise SystemExit(f"bloque {tipo} sin tabla:\n{cuerpo[:240]}")
    if tipo == "SEQUENCE OWNED BY":
        return m.group(2)
    return m.group(1)


def _indice_comentarios(bloques: list[dict[str, str]]) -> dict[tuple[str, str], list[str]]:
    """Clave (clase, identificador) → sentencias COMMENT."""
    out: dict[tuple[str, str], list[str]] = {}
    for b in bloques:
        if b["tipo"] != "COMMENT":
            continue
        cuerpo = b["cuerpo"]
        if cuerpo.startswith("COMMENT ON SCHEMA "):
            b["usado"] = True
            continue
        m = re.match(
            r"COMMENT ON (TABLE|VIEW|COLUMN|FUNCTION|CONSTRAINT) public\.(\w+)",
            cuerpo,
        )
        if not m:
            m = re.match(
                r"COMMENT ON CONSTRAINT (\w+) ON public\.(\w+)",
                cuerpo,
            )
            if not m:
                raise SystemExit(f"comentario no reconocido:\n{cuerpo[:240]}")
            out.setdefault(("constraint", m.group(1)), []).append(cuerpo)
            b["usado"] = True
            continue
        clase, ident = m.group(1), m.group(2)
        if clase == "COLUMN":
            out.setdefault(("column", ident), []).append(cuerpo)
        elif clase == "FUNCTION":
            out.setdefault(("function", ident), []).append(cuerpo)
        elif clase == "TABLE":
            out.setdefault(("table", ident), []).append(cuerpo)
        elif clase == "VIEW":
            out.setdefault(("view", ident), []).append(cuerpo)
        else:
            out.setdefault(("constraint", ident), []).append(cuerpo)
        b["usado"] = True
    return out


def _tomar(comentarios: dict[tuple[str, str], list[str]], clave: tuple[str, str]) -> list[str]:
    return comentarios.pop(clave, [])


def _emitir(partes: list[str], sentencias: list[str]) -> None:
    for s in sentencias:
        partes.append(s.rstrip() + "\n")


def construir(dump: str, *, excluir: set[str] | None = None) -> str:
    excluir = excluir or set()
    dom_tabla = _dominio_de_tabla()
    bloques = _parsear(dump)
    for b in bloques:
        b["usado"] = False
        if b["tipo"] == "SCHEMA":
            b["usado"] = True
    comentarios = _indice_comentarios(bloques)

    tablas: dict[str, str] = {}
    secuencias: dict[str, list[str]] = {}
    owned: dict[str, str] = {}
    defaults: dict[str, list[str]] = {}
    constraints: dict[str, list[tuple[str, str]]] = {}
    indices: dict[str, list[tuple[str, str]]] = {}
    fks: list[tuple[str, str, str]] = []
    funciones: dict[str, str] = {}
    vistas: dict[str, str] = {}
    triggers: dict[str, list[str]] = {}

    for b in bloques:
        if b["usado"]:
            continue
        tipo = b["tipo"]
        cuerpo = b["cuerpo"]
        if tipo == "TABLE":
            m = re.search(r"CREATE TABLE public\.(\w+)", cuerpo)
            if not m:
                raise SystemExit(f"tabla sin nombre:\n{cuerpo[:200]}")
            tablas[m.group(1)] = cuerpo
        elif tipo == "SEQUENCE":
            m = re.search(r"CREATE SEQUENCE public\.(\w+)", cuerpo)
            if not m:
                raise SystemExit(f"secuencia sin nombre:\n{cuerpo[:200]}")
            secuencias.setdefault(m.group(1), []).append(cuerpo)
        elif tipo == "SEQUENCE OWNED BY":
            m = TABLA_DE[tipo].search(cuerpo)
            if not m:
                raise SystemExit(cuerpo[:200])
            owned[m.group(1)] = m.group(2)
            secuencias.setdefault(m.group(1), []).append(cuerpo)
        elif tipo == "DEFAULT":
            tabla = _tabla_de(tipo, cuerpo)
            defaults.setdefault(tabla, []).append(cuerpo)
        elif tipo == "CONSTRAINT":
            tabla = _tabla_de(tipo, cuerpo)
            m = re.search(r"ADD CONSTRAINT (\w+)", cuerpo)
            nombre = m.group(1) if m else cuerpo
            constraints.setdefault(tabla, []).append((nombre, cuerpo))
        elif tipo == "INDEX":
            tabla = _tabla_de(tipo, cuerpo)
            m = re.search(r"CREATE (?:UNIQUE )?INDEX (\w+)", cuerpo)
            nombre = m.group(1) if m else cuerpo
            indices.setdefault(tabla, []).append((nombre, cuerpo))
        elif tipo == "FK CONSTRAINT":
            tabla = _tabla_de(tipo, cuerpo)
            m = re.search(r"ADD CONSTRAINT (\w+)", cuerpo)
            nombre = m.group(1) if m else cuerpo
            fks.append((tabla, nombre, cuerpo))
        elif tipo == "FUNCTION":
            funciones[_nombre_funcion(cuerpo)] = cuerpo
        elif tipo == "VIEW":
            m = re.search(r"CREATE VIEW public\.(\w+)", cuerpo)
            if not m:
                raise SystemExit(cuerpo[:200])
            vistas[m.group(1)] = cuerpo
        elif tipo == "TRIGGER":
            tabla = _tabla_de(tipo, cuerpo)
            triggers.setdefault(tabla, []).append(cuerpo)
        else:
            raise SystemExit(f"tipo de objeto no previsto: {tipo} ({b['nombre']})")
        b["usado"] = True

    sin_usar = [f"{b['tipo']} {b['nombre']}" for b in bloques if not b["usado"]]
    if sin_usar:
        raise SystemExit("objetos sin clasificar: " + ", ".join(sin_usar))

    faltan = sorted(set(tablas) - set(dom_tabla))
    if faltan:
        raise SystemExit("tablas sin dominio: " + ", ".join(faltan))
    for nombre in funciones:
        if nombre not in FUNCIONES:
            raise SystemExit(f"función sin dominio: {nombre}")
    for nombre in vistas:
        if nombre not in VISTAS:
            raise SystemExit(f"vista sin dominio: {nombre}")
    for seq, tabla in owned.items():
        if tabla not in tablas:
            raise SystemExit(f"la secuencia {seq} pertenece a {tabla}, que no existe")
    huerfanas = sorted(set(secuencias) - set(owned))
    if huerfanas:
        raise SystemExit("secuencias sin dueño: " + ", ".join(huerfanas))

    partes: list[str] = [PREAMBULO]
    for clave, titulo, lista in DOMINIOS:
        tablas_seccion = [t for t in lista if t in tablas and t not in excluir]
        funcs = sorted(n for n, d in FUNCIONES.items() if d == clave and n in funciones)
        vis = sorted(n for n, d in VISTAS.items() if d == clave and n in vistas)
        if not tablas_seccion and not funcs and not vis:
            continue
        partes.append(f"-- {'=' * 72}\n")
        partes.append(f"-- {titulo}\n")
        partes.append(f"-- {'=' * 72}\n")
        for nombre in funcs:
            _emitir(partes, [funciones[nombre]])
            _emitir(partes, _tomar(comentarios, ("function", nombre)))
        for tabla in tablas_seccion:
            partes.append(f"-- {tabla}\n")
            _emitir(partes, [tablas[tabla]])
            for m in re.finditer(r"CONSTRAINT (\w+)", tablas[tabla]):
                _emitir(partes, _tomar(comentarios, ("constraint", m.group(1))))
            _emitir(partes, _tomar(comentarios, ("table", tabla)))
            _emitir(partes, _tomar(comentarios, ("column", tabla)))
            for seq, dueño in sorted(owned.items()):
                if dueño != tabla:
                    continue
                _emitir(partes, secuencias[seq])
            _emitir(partes, defaults.get(tabla, []))
            for _nombre, cuerpo in sorted(constraints.get(tabla, [])):
                _emitir(partes, [cuerpo])
                m = re.search(r"ADD CONSTRAINT (\w+)", cuerpo)
                if m:
                    _emitir(partes, _tomar(comentarios, ("constraint", m.group(1))))
            for _nombre, cuerpo in sorted(indices.get(tabla, [])):
                _emitir(partes, [cuerpo])
            _emitir(partes, triggers.get(tabla, []))
        for nombre in vis:
            _emitir(partes, [vistas[nombre]])
            _emitir(partes, _tomar(comentarios, ("view", nombre)))

    fks_vis = [(t, n, c) for t, n, c in fks if t not in excluir]
    if fks_vis:
        partes.append(f"-- {'=' * 72}\n")
        partes.append("-- Claves foráneas\n")
        partes.append(f"-- {'=' * 72}\n")
        for clave, titulo, lista in DOMINIOS:
            grupo = [(t, n, c) for t, n, c in fks_vis if t in lista]
            if not grupo:
                continue
            partes.append(f"-- {titulo}\n")
            for _t, nombre, cuerpo in sorted(grupo, key=lambda x: x[1]):
                _emitir(partes, [cuerpo])
                _emitir(partes, _tomar(comentarios, ("constraint", nombre)))

    if comentarios:
        sobra = ", ".join(f"{k[0]}:{k[1]}" for k in sorted(comentarios))
        raise SystemExit("comentarios sin objeto: " + sobra)
    return "".join(partes)


def datos_referencia(dump_datos: str, excluir: set[str]) -> str:
    lineas = []
    for ln in dump_datos.splitlines():
        s = ln.strip()
        if s.startswith("INSERT INTO public."):
            m = re.match(r"INSERT INTO public\.(\w+)", s)
            if m and m.group(1) in excluir:
                continue
            lineas.append(s)
        elif s.startswith("SELECT pg_catalog.setval("):
            lineas.append(s)
    if not lineas:
        raise SystemExit("el volcado de datos no trae INSERT ni setval")
    return "\n".join(lineas) + "\n"


ENCABEZADO_ESQUEMA = """\
-- Esquema final generado. No editar a mano.
-- Lo escribe scripts/generar-esquema-bd.sh después de aplicar db/migrations
-- con cmd/migrate sobre un PostGIS vacío.
-- La fuente de verdad son esas migraciones. Este archivo es la foto del
-- resultado, sin dueños ni privilegios, para leerla y para el control de deriva.
-- Si difiere de un volcado nuevo, hay que regenerarlo con el mismo script.

"""

ENCABEZADO_BASELINE = """\
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

"""


def esquema(dump: str) -> str:
    return ENCABEZADO_ESQUEMA + construir(dump)


def baseline(dump: str, dump_datos: str) -> str:
    excluir = {"schema_migrations", "usuarios", "sesiones"}
    ddl = construir(dump, excluir={"schema_migrations"})
    datos = datos_referencia(dump_datos, excluir)
    return (
        ENCABEZADO_BASELINE
        + ddl
        + f"-- {'=' * 72}\n"
        + "-- Datos de referencia\n"
        + "-- Filas que la serie histórica dejaba en una base vacía.\n"
        + "-- Sin usuarios ni sesiones: los crea la semilla de accesos.\n"
        + f"-- {'=' * 72}\n"
        + datos
        + "SELECT pg_catalog.set_config('search_path', 'public', true);\n"
    )


def main(argv: list[str]) -> None:
    if len(argv) < 3 or argv[1] not in {"normalizar", "baseline"}:
        raise SystemExit("uso: esquema_sql.py normalizar VOLCADO | baseline VOLCADO DATOS")
    dump = open(argv[2], encoding="utf-8").read()
    if argv[1] == "normalizar":
        sys.stdout.write(esquema(dump))
        return
    if len(argv) != 4:
        raise SystemExit("baseline necesita el volcado de datos")
    datos = open(argv[3], encoding="utf-8").read()
    sys.stdout.write(baseline(dump, datos))


if __name__ == "__main__":
    main(sys.argv)
