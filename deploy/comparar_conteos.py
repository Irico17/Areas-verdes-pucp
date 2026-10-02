#!/usr/bin/env python3
"""Compara conteos de tablas antes y después del despliegue en producción.

Compara únicamente las tablas de negocio y excluye tablas técnicas que cambian
legítimamente durante el despliegue (migraciones y sesiones de smoke test).
"""

from __future__ import annotations

import argparse
import io
import sys
from pathlib import Path


def leer_exclusiones(path: str | Path | None) -> set[str]:
    """Lee el archivo de exclusiones, ignorando comentarios (#) y líneas vacías."""
    if not path:
        return set()
    p = Path(path)
    if not p.is_file():
        return set()
    excluidas = set()
    for line in p.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        # En caso de comentarios inline o espacios adicionales
        tabla = line.split("#", 1)[0].strip()
        if tabla:
            excluidas.add(tabla)
    return excluidas


def leer_conteos(contenido_o_path: str | Path) -> dict[str, int]:
    """Parsea el resultado del conteo TSV (table_name\\tfilas)."""
    if isinstance(contenido_o_path, Path) or (
        isinstance(contenido_o_path, str) and "\n" not in contenido_o_path and Path(contenido_o_path).is_file()
    ):
        text = Path(contenido_o_path).read_text(encoding="utf-8")
    else:
        text = str(contenido_o_path)

    out: dict[str, int] = {}
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("table_name"):
            continue
        parts = line.split("\t")
        if len(parts) < 2:
            continue
        try:
            out[parts[0]] = int(parts[1])
        except ValueError:
            continue
    return out


def comparar(
    antes: dict[str, int],
    despues: dict[str, int],
    excluidas: set[str] | None = None,
) -> tuple[list[str], list[str], int]:
    """Compara los conteos excluyendo las tablas indicadas.

    Devuelve (diferencias, nuevas, total_comparadas).
    """
    if excluidas is None:
        excluidas = set()

    diferencias: list[str] = []
    comparadas = 0

    for tabla, filas in sorted(antes.items()):
        if tabla in excluidas:
            continue
        comparadas += 1
        if despues.get(tabla) != filas:
            despues_val = despues.get(tabla)
            diferencias.append(f"{tabla}: antes {filas}, después {despues_val}")

    nuevas = sorted(set(despues) - set(antes))
    return diferencias, nuevas, comparadas


def filtrar_conteos(texto: str, excluidas: set[str]) -> str:
    """Filtra líneas TSV excluyendo las tablas presentes en excluidas."""
    lineas_out = []
    for line in texto.splitlines():
        line_clean = line.strip()
        if not line_clean or line_clean.startswith("table_name"):
            lineas_out.append(line)
            continue
        parts = line_clean.split("\t")
        tabla = parts[0]
        if tabla in excluidas:
            continue
        lineas_out.append(line)
    return "\n".join(lineas_out) + ("\n" if texto.endswith("\n") else "")


def _selftest() -> None:
    # 1. Caso base: conteos coinciden en negocio, pero sesiones y schema_migrations cambiaron
    antes_tsv = """table_name\tfilas
areas_verdes\t521
schema_migrations\t36
sesiones\t0
usuarios\t6
zonas_supervision\t534
"""
    despues_tsv = """table_name\tfilas
areas_verdes\t521
schema_migrations\t38
sesiones\t1
usuarios\t6
zonas_supervision\t534
"""
    excluir_txt = """# Archivo de prueba
schema_migrations # migración aplicada
sesiones
"""
    excluidas = leer_exclusiones(Path("/dev/null"))  # archivo inexistente
    if excluidas != set():
        raise SystemExit(f"archivo inexistente debió devolver set vacío: {excluidas}")

    # Parsear exclusiones desde texto
    excluidas_test = set()
    for l in excluir_txt.splitlines():
        l = l.strip()
        if l and not l.startswith("#"):
            excluidas_test.add(l.split("#")[0].strip())

    antes = leer_conteos(antes_tsv)
    despues = leer_conteos(despues_tsv)

    # Con exclusiones: debe pasar sin diferencias
    difs, nuevas, n_comp = comparar(antes, despues, excluidas_test)
    if difs:
        raise SystemExit(f"No debió haber diferencias con exclusiones: {difs}")
    if n_comp != 3:
        raise SystemExit(f"Debieron compararse 3 tablas de negocio, se compararon: {n_comp}")
    if nuevas:
        raise SystemExit(f"No debió haber tablas nuevas: {nuevas}")

    # Sin exclusiones: debe fallar reportando sesiones y schema_migrations
    difs_sin_excl, _, _ = comparar(antes, despues, set())
    if len(difs_sin_excl) != 2:
        raise SystemExit(f"Debieron fallar 2 tablas sin exclusiones: {difs_sin_excl}")

    # 2. Caso fallo: tabla de negocio con conteo distinto debe fallar
    despues_negocio_cambio = """table_name\tfilas
areas_verdes\t522
schema_migrations\t38
sesiones\t1
usuarios\t6
zonas_supervision\t534
"""
    despues_neg = leer_conteos(despues_negocio_cambio)
    difs_neg, _, _ = comparar(antes, despues_neg, excluidas_test)
    if not difs_neg or "areas_verdes" not in difs_neg[0]:
        raise SystemExit(f"Cambio en tabla de negocio 'areas_verdes' no fue detectado: {difs_neg}")

    # 3. Caso tablas nuevas: se reportan y no provocan fallo
    despues_con_nueva = """table_name\tfilas
areas_verdes\t521
nueva_tabla_negocio\t10
schema_migrations\t38
sesiones\t1
usuarios\t6
zonas_supervision\t534
"""
    despues_nueva = leer_conteos(despues_con_nueva)
    difs_nueva, nuevas_tabla, _ = comparar(antes, despues_nueva, excluidas_test)
    if difs_nueva:
        raise SystemExit(f"Tabla nueva no debe considerarse diferencia destructiva: {difs_nueva}")
    if nuevas_tabla != ["nueva_tabla_negocio"]:
        raise SystemExit(f"No se reportó la tabla nueva: {nuevas_tabla}")

    # 4. Caso filtro de texto
    filtrado = filtrar_conteos(antes_tsv, excluidas_test)
    if "schema_migrations" in filtrado or "sesiones" in filtrado:
        raise SystemExit(f"Filtro no removió tablas excluidas: {filtrado}")
    if "areas_verdes" not in filtrado:
        raise SystemExit(f"Filtro eliminó tabla de negocio: {filtrado}")

    print("comparar_conteos ok")


def main() -> None:
    if len(sys.argv) == 2 and sys.argv[1] == "--selftest":
        _selftest()
        return

    parser = argparse.ArgumentParser(
        description="Compara conteos de tablas de negocio antes y después del despliegue"
    )
    parser.add_argument("antes", nargs="?", default="", help="Archivo TSV con conteos antes")
    parser.add_argument("despues", nargs="?", default="", help="Archivo TSV con conteos después")
    parser.add_argument(
        "--excluir",
        "-e",
        default=None,
        help="Archivo con lista de tablas a excluir (una por línea)",
    )
    parser.add_argument(
        "--filtrar",
        metavar="EXCLUIR_FILE",
        default=None,
        help="Filtra conteos TSV (desde archivo o stdin) omitiendo tablas excluidas",
    )
    parser.add_argument(
        "excluir_posicional",
        nargs="?",
        default=None,
        help="Archivo con tablas a excluir (argumento posicional opcional)",
    )

    args = parser.parse_args()

    # Modo filtrar (usado por conteos.sh)
    if args.filtrar:
        excluidas = leer_exclusiones(args.filtrar)
        if args.antes and Path(args.antes).is_file():
            entrada = Path(args.antes).read_text(encoding="utf-8")
        else:
            entrada = sys.stdin.read()
        sys.stdout.write(filtrar_conteos(entrada, excluidas))
        return

    if not args.antes or not args.despues:
        parser.print_usage(sys.stderr)
        sys.exit(2)

    excluir_path = args.excluir or args.excluir_posicional
    excluidas = leer_exclusiones(excluir_path)

    antes = leer_conteos(Path(args.antes))
    despues = leer_conteos(Path(args.despues))

    diferencias, nuevas, comparadas = comparar(antes, despues, excluidas)

    if diferencias:
        sys.stderr.write("conteos distintos en tablas que ya existían:\n")
        for d in diferencias:
            sys.stderr.write(f"{d}\n")
        sys.exit(1)

    print(f"conteos iguales en {comparadas} tablas existentes")
    if nuevas:
        print("tablas nuevas: " + ", ".join(nuevas))
    if excluidas:
        excluidas_presentes = sorted(excluidas & (set(antes) | set(despues)))
        if excluidas_presentes:
            print(f"tablas excluidas de la comparación: {', '.join(excluidas_presentes)}")


if __name__ == "__main__":
    main()
