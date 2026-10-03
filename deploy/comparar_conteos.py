#!/usr/bin/env python3
"""Compara conteos de tablas antes y después del despliegue en producción.

Compara únicamente las tablas de negocio y excluye tablas técnicas que cambian
legítimamente durante el despliegue (migraciones y sesiones de smoke test).

Solo una BAJA de conteo (o una tabla que existía antes y ya no está) es pérdida
de datos y hace fallar. Los aumentos son válidos (las migraciones y el smoke
añaden filas: catálogos, permisos, auditoría) y las tablas nuevas no cuentan:
se reportan en el log sin fallar.
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
) -> tuple[list[str], list[str], list[str], int]:
    """Compara los conteos excluyendo las tablas indicadas.

    Devuelve (bajas, aumentos, nuevas, total_comparadas):
    - bajas: tablas que bajaron de conteo o desaparecieron (pérdida: falla).
    - aumentos: tablas que subieron (válido, solo se reporta).
    - nuevas: tablas que no existían antes (válido, solo se reporta).
    """
    if excluidas is None:
        excluidas = set()

    bajas: list[str] = []
    aumentos: list[str] = []
    comparadas = 0

    for tabla, filas in sorted(antes.items()):
        if tabla in excluidas:
            continue
        comparadas += 1
        despues_val = despues.get(tabla)
        if despues_val is None:
            bajas.append(f"{tabla}: antes {filas}, después ausente")
        elif despues_val < filas:
            bajas.append(f"{tabla}: antes {filas}, después {despues_val} (bajó {filas - despues_val})")
        elif despues_val > filas:
            aumentos.append(f"{tabla}: antes {filas}, después {despues_val} (+{despues_val - filas})")

    nuevas = sorted(t for t in set(despues) - set(antes) if t not in excluidas)
    return bajas, aumentos, nuevas, comparadas


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
    antes_tsv = """table_name\tfilas
areas_verdes\t521
catalogos\t45
schema_migrations\t36
sesiones\t0
usuarios\t6
zonas_supervision\t534
"""
    excluidas = leer_exclusiones(Path("/dev/null"))  # archivo inexistente
    if excluidas != set():
        raise SystemExit(f"archivo inexistente debió devolver set vacío: {excluidas}")
    excluidas_test = {"schema_migrations", "sesiones"}
    antes = leer_conteos(antes_tsv)

    def caso(cuerpo: str, excl: set[str] | None = excluidas_test):
        return comparar(antes, leer_conteos("table_name\tfilas\n" + cuerpo), excl)

    base = "areas_verdes\t521\nschema_migrations\t38\nsesiones\t1\nusuarios\t6\nzonas_supervision\t534\n"

    # 1. Igual en negocio (las excluidas cambian): sin bajas ni aumentos
    bajas, aumentos, nuevas, n_comp = caso("catalogos\t45\n" + base)
    if bajas or aumentos or nuevas:
        raise SystemExit(f"Caso igual no debió reportar nada: {bajas} {aumentos} {nuevas}")
    if n_comp != 4:
        raise SystemExit(f"Debieron compararse 4 tablas de negocio, se compararon: {n_comp}")

    # 2. Sin exclusiones, las tablas técnicas que SUBEN tampoco fallan
    bajas, aumentos, _, _ = caso("catalogos\t45\n" + base, set())
    if bajas:
        raise SystemExit(f"Un aumento no debe fallar: {bajas}")
    if len(aumentos) != 2:
        raise SystemExit(f"Debieron reportarse 2 aumentos (schema_migrations y sesiones): {aumentos}")

    # 3. Sube (migraciones añaden catálogos): válido, se reporta
    bajas, aumentos, _, _ = caso("catalogos\t116\n" + base)
    if bajas:
        raise SystemExit(f"Subir catalogos 45->116 no debe fallar: {bajas}")
    if aumentos != ["catalogos: antes 45, después 116 (+71)"]:
        raise SystemExit(f"No se reportó el aumento de catalogos: {aumentos}")

    # 4. Baja en una tabla de negocio: debe fallar
    bajas, _, _, _ = caso("catalogos\t45\n" + base.replace("areas_verdes\t521", "areas_verdes\t520"))
    if len(bajas) != 1 or not bajas[0].startswith("areas_verdes: antes 521, después 520"):
        raise SystemExit(f"La baja de areas_verdes no fue detectada: {bajas}")

    # 5. Baja a cero también falla
    bajas, _, _, _ = caso("catalogos\t0\n" + base)
    if len(bajas) != 1 or "catalogos" not in bajas[0]:
        raise SystemExit(f"La baja de catalogos a 0 no fue detectada: {bajas}")

    # 6. Tabla que existía antes y desaparece: falla
    bajas, _, _, _ = caso("catalogos\t45\n" + base.replace("usuarios\t6\n", ""))
    if len(bajas) != 1 or "usuarios" not in bajas[0] or "ausente" not in bajas[0]:
        raise SystemExit(f"La tabla ausente después no fue detectada: {bajas}")

    # 7. Tabla nueva (no existía antes): se reporta y no falla
    bajas, aumentos, nuevas, _ = caso("catalogos\t45\nnueva_tabla_negocio\t10\n" + base)
    if bajas or aumentos:
        raise SystemExit(f"Una tabla nueva no debe fallar ni contarse como aumento: {bajas} {aumentos}")
    if nuevas != ["nueva_tabla_negocio"]:
        raise SystemExit(f"No se reportó la tabla nueva: {nuevas}")

    # 8. Mezcla: sube una, baja otra -> falla solo por la que baja
    bajas, aumentos, _, _ = caso(
        "catalogos\t116\n" + base.replace("zonas_supervision\t534", "zonas_supervision\t533")
    )
    if len(bajas) != 1 or "zonas_supervision" not in bajas[0] or len(aumentos) != 1:
        raise SystemExit(f"La mezcla debe fallar solo por la baja: {bajas} {aumentos}")

    # 9. Filtro de texto
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

    bajas, aumentos, nuevas, comparadas = comparar(antes, despues, excluidas)

    if aumentos:
        print("tablas que subieron de conteo (válido, p. ej. filas añadidas por migraciones):")
        for a in aumentos:
            print(f"  {a}")
    if nuevas:
        print("tablas nuevas: " + ", ".join(nuevas))
    if excluidas:
        excluidas_presentes = sorted(excluidas & (set(antes) | set(despues)))
        if excluidas_presentes:
            print(f"tablas excluidas de la comparación: {', '.join(excluidas_presentes)}")

    if bajas:
        sys.stderr.write("PÉRDIDA DE DATOS: tablas que bajaron de conteo o desaparecieron:\n")
        for d in bajas:
            sys.stderr.write(f"  {d}\n")
        sys.exit(1)

    print(f"ningún conteo bajó en {comparadas} tablas existentes ({len(aumentos)} subieron)")


if __name__ == "__main__":
    main()
