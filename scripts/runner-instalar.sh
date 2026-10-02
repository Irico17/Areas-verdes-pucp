#!/usr/bin/env bash
# Instala y configura un runner self-hosted de GitHub Actions en la EC2.
# Se ejecuta como root en la instancia vía AWS SSM Session Manager o Run Shell Script.
# Uso: runner-instalar.sh <develop|qa|produccion>
set -euo pipefail
set +x

AMBIENTE="${1:-}"
case "$AMBIENTE" in
  develop)
    LABEL="campus-develop"
    ;;
  qa)
    LABEL="campus-qa"
    ;;
  produccion)
    LABEL="campus-prod"
    ;;
  *)
    echo "Uso: $0 <develop|qa|produccion>" >&2
    exit 1
    ;;
esac

if [ "$(id -u)" -ne 0 ]; then
  echo "ERROR: Este script debe ejecutarse como root en la instancia." >&2
  exit 1
fi

RUNNER_DIR="/opt/actions-runner-${AMBIENTE}"
HOSTNAME_SHORT="$(hostname -s 2>/dev/null || hostname)"
RUNNER_NAME="campus-${AMBIENTE}-${HOSTNAME_SHORT}"

# Asegurar usuario de sistema runner en grupo docker
id -u runner >/dev/null 2>&1 || useradd -r -m -d /home/runner -s /bin/bash -G docker runner
usermod -aG docker runner || true

# Idempotencia: si ya está configurado, solo asegurar el servicio systemd
if [ -f "$RUNNER_DIR/.runner" ]; then
  echo "El runner para '$AMBIENTE' ya se encuentra configurado en $RUNNER_DIR."
  echo "Asegurando servicio systemd..."
  cd "$RUNNER_DIR"
  ./svc.sh install runner 2>/dev/null || true
  ./svc.sh start 2>/dev/null || true
  ./svc.sh status || true
  echo "Runner para '$AMBIENTE' asegurado exitosamente."
  exit 0
fi

# Obtener token de registro sin imprimirlo
TOKEN="${RUNNER_TOKEN:-}"
if [ -z "$TOKEN" ]; then
  echo "Obteniendo token de registro desde AWS SSM (/campus/runner-token/${AMBIENTE})..."
  TOKEN="$(aws ssm get-parameter --with-decryption --name "/campus/runner-token/${AMBIENTE}" --query 'Parameter.Value' --output text 2>/dev/null || true)"
fi

if [ -z "$TOKEN" ] || [ "$TOKEN" = "None" ]; then
  echo "ERROR: No se encontró el token de registro en RUNNER_TOKEN ni en SSM (/campus/runner-token/${AMBIENTE})." >&2
  echo "El operador debe exportar RUNNER_TOKEN o guardar un SecureString en SSM antes de ejecutar este script." >&2
  exit 1
fi

# Versión conocida de respaldo en caso de rate-limit de la API pública de GitHub
FALLBACK_VER="2.337.0"
FALLBACK_SHA="70920811a4f8ad4328818682bca5c6469c1c942fab52448868071d0063816613"

API_URL="https://api.github.com/repos/actions/runner/releases/latest"
RELEASE_JSON="$(curl -fsSL -H "User-Agent: campus-runner-installer" "$API_URL" 2>/dev/null || true)"

RUNNER_VER=""
EXPECTED_SHA=""
RUNNER_URL=""

if [ -n "$RELEASE_JSON" ] && command -v jq >/dev/null 2>&1 && echo "$RELEASE_JSON" | jq -e .tag_name >/dev/null 2>&1; then
  RUNNER_VER="$(echo "$RELEASE_JSON" | jq -r .tag_name | sed 's/^v//')"
  RUNNER_URL="$(echo "$RELEASE_JSON" | jq -r '.assets[] | select(.name | test("actions-runner-linux-x64-.*\\.tar\\.gz$")) | .browser_download_url' | head -n 1)"
  BODY="$(echo "$RELEASE_JSON" | jq -r .body)"
  EXPECTED_SHA="$(echo "$BODY" | grep -oE 'BEGIN SHA linux-x64 -->[a-f0-9]{64}' | sed 's/BEGIN SHA linux-x64 -->//' | head -n 1 || true)"
  if [ -z "$EXPECTED_SHA" ]; then
    EXPECTED_SHA="$(echo "$BODY" | grep "actions-runner-linux-x64" | grep -oE '[a-f0-9]{64}' | head -n 1 || true)"
  fi
fi

if [ -z "$RUNNER_VER" ] || [ -z "$RUNNER_URL" ] || [ -z "$EXPECTED_SHA" ]; then
  echo "Aviso: no fue posible resolver la última release dinámicamente (API rate-limit o sin conexión). Usando versión conocida v${FALLBACK_VER}."
  RUNNER_VER="$FALLBACK_VER"
  RUNNER_URL="https://github.com/actions/runner/releases/download/v${FALLBACK_VER}/actions-runner-linux-x64-${FALLBACK_VER}.tar.gz"
  EXPECTED_SHA="$FALLBACK_SHA"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
  unset TOKEN RUNNER_TOKEN
}
trap cleanup EXIT

TARBALL="$TMP_DIR/actions-runner-linux-x64-${RUNNER_VER}.tar.gz"

echo "Descargando actions-runner linux-x64 v${RUNNER_VER}..."
curl -fsSL -o "$TARBALL" "$RUNNER_URL"

echo "Verificando SHA-256..."
echo "${EXPECTED_SHA}  ${TARBALL}" | sha256sum -c -

echo "Descomprimiendo en $RUNNER_DIR..."
mkdir -p "$RUNNER_DIR"
tar -xzf "$TARBALL" -C "$RUNNER_DIR"
chown -R runner:runner "$RUNNER_DIR"

echo "Configurando runner '$RUNNER_NAME' con etiqueta '$LABEL'..."
if command -v runuser >/dev/null 2>&1; then
  RUNNER_TOKEN="$TOKEN" runuser -u runner -- bash -c 'cd "$1" && ./config.sh --unattended --replace --url "https://github.com/Irico17/Areas-verdes-pucp" --token "$RUNNER_TOKEN" --name "$2" --labels "$3" --work _work' _ "$RUNNER_DIR" "$RUNNER_NAME" "$LABEL"
elif command -v sudo >/dev/null 2>&1; then
  RUNNER_TOKEN="$TOKEN" sudo -E -u runner bash -c 'cd "$1" && ./config.sh --unattended --replace --url "https://github.com/Irico17/Areas-verdes-pucp" --token "$RUNNER_TOKEN" --name "$2" --labels "$3" --work _work' _ "$RUNNER_DIR" "$RUNNER_NAME" "$LABEL"
else
  RUNNER_TOKEN="$TOKEN" su -s /bin/bash runner -c 'cd "$1" && ./config.sh --unattended --replace --url "https://github.com/Irico17/Areas-verdes-pucp" --token "$RUNNER_TOKEN" --name "$2" --labels "$3" --work _work' _ "$RUNNER_DIR" "$RUNNER_NAME" "$LABEL"
fi

echo "Instalando e iniciando servicio systemd..."
cd "$RUNNER_DIR"
./svc.sh install runner
./svc.sh start
./svc.sh status || true

echo "Instalación de runner para '$AMBIENTE' completada exitosamente."
