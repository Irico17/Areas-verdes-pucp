#!/usr/bin/env bash
# Falla si el texto del plan contiene un secreto que esté en el entorno.
# No imprime el valor.
set -euo pipefail
archivo="${1:?falta el archivo del plan}"
if [ ! -f "$archivo" ]; then
  echo "No existe el plan de texto: $archivo" >&2
  exit 1
fi
python3 - "$archivo" <<'PY'
import os
import sys

path = sys.argv[1]
blob = open(path, "r", encoding="utf-8", errors="replace").read()
nombres = [
    "AWS_ACCESS_KEY_ID",
    "AWS_SECRET_ACCESS_KEY",
    "AWS_SESSION_TOKEN",
    "TF_VAR_db_password",
    "TF_VAR_dev_password",
    "RUNNER_REG_TOKEN_PAT",
]
encontrados = []
for nombre in nombres:
    valor = os.environ.get(nombre, "")
    if len(valor) >= 8 and valor in blob:
        encontrados.append(nombre)
bloque = os.environ.get("BLOQUE_CREDENCIALES_AWS", "")
if bloque:
    for linea in bloque.replace("\r", "\n").replace(";", "\n").splitlines():
        linea = linea.strip()
        if len(linea) >= 8 and linea in blob:
            encontrados.append("bloque_credenciales_aws")
            break
if "-----BEGIN " in blob:
    encontrados.append("PEM")
if encontrados:
    sys.stderr.write(
        "El plan contiene material sensible (%s). No se sube ni se aplica.\n"
        % ", ".join(encontrados)
    )
    sys.exit(1)
PY
