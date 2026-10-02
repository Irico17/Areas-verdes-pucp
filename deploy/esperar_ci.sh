#!/usr/bin/env bash
# Espera a que el workflow ci de este SHA termine en success.
# Uso: deploy/esperar_ci.sh <sha>
set -euo pipefail

SHA="${1:?falta el SHA}"
if ! [[ "$SHA" =~ ^[0-9a-f]{40}$ ]]; then
  echo "SHA inválido: $SHA" >&2
  exit 2
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh no está instalado" >&2
  exit 2
fi

intentos="${ESPERAR_CI_INTENTOS:-60}"
pausa="${ESPERAR_CI_PAUSA:-15}"
i=1
while [ "$i" -le "$intentos" ]; do
  json="$(gh run list --workflow ci.yml --commit "$SHA" --limit 5 --json status,conclusion,databaseId,createdAt)"
  decision="$(python3 -c '
import json, sys
runs = json.loads(sys.argv[1] or "[]")
if not runs:
    print("ausente")
    raise SystemExit(0)
en_curso = [r for r in runs if r.get("status") != "completed"]
if en_curso:
    print("curso")
    raise SystemExit(0)
ok = [r for r in runs if r.get("conclusion") == "success"]
if ok:
    print("ok")
else:
    print("fallo")
' "$json")"
  case "$decision" in
    ok)
      echo "CI en verde para $SHA"
      exit 0
      ;;
    fallo)
      echo "CI falló para $SHA. No se despliega." >&2
      exit 1
      ;;
    curso|ausente)
      echo "CI $decision ($i/$intentos). Esperando ${pausa}s."
      ;;
  esac
  i=$((i + 1))
  sleep "$pausa"
done

echo "No hubo CI en verde para $SHA a tiempo." >&2
exit 1
