#!/usr/bin/env bash
# Segunda corrida del layout sobre un contenedor Ubuntu. No llama a AWS.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
NOMBRE="campus-ansible-prueba"
docker rm -f "$NOMBRE" >/dev/null 2>&1 || true
docker run -d --name "$NOMBRE" ubuntu:24.04 sleep 600 >/dev/null
cleanup() { docker rm -f "$NOMBRE" >/dev/null 2>&1 || true; }
trap cleanup EXIT

docker exec -e DEBIAN_FRONTEND=noninteractive "$NOMBRE" apt-get update
docker exec -e DEBIAN_FRONTEND=noninteractive "$NOMBRE" apt-get install -y python3 python3-pip
docker exec "$NOMBRE" pip3 install --break-system-packages 'ansible-core==2.18.8'
docker exec "$NOMBRE" mkdir -p /src
docker cp "$ROOT/infra/ansible" "$NOMBRE:/src/ansible"
docker cp "$ROOT/deploy" "$NOMBRE:/src/deploy"
# El playbook busca deploy en ../../deploy respecto de infra/ansible.
docker exec "$NOMBRE" mkdir -p /src/infra
docker exec "$NOMBRE" mv /src/ansible /src/infra/ansible

run() {
  docker exec "$NOMBRE" ansible-playbook /src/infra/ansible/site.yml \
    -i /src/infra/ansible/inventory/prueba-local.ini \
    -e campus_en_contenedor_prueba=true \
    -e campus_habilitar_swap=false \
    -e campus_registrar_runner=false \
    --skip-tags systemd,packages,runner,swap
}

echo "Primera corrida"
run | tee /tmp/ansible-prueba-1.txt
echo "Segunda corrida"
run | tee /tmp/ansible-prueba-2.txt
if grep -E 'changed=[1-9]' /tmp/ansible-prueba-2.txt; then
  echo "La segunda corrida todavía cambió algo." >&2
  exit 1
fi
echo "Idempotencia OK."
