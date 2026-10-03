# Protecciones del repositorio

Este documento fija las protecciones que se aplican al repositorio personal `Irico17/Areas-verdes-pucp` y los comandos `gh api` exactos para repetirlas en otro repositorio (por ejemplo, el de la organización `GRUPO-12-DP2`). **Los comandos de la organización no se han ejecutado**: se dejan aquí para que quien administre esa organización los aplique con su propia sesión (`gh auth login`) y revise el resultado.

Requisitos: `gh` autenticado con un usuario con permiso de administración sobre el repositorio, y un repositorio público o un plan que admita protección de ramas y environments con revisores (en un repositorio privado de un plan gratuito la API responde 403 con «Upgrade to GitHub Pro or make this repository public»; en ese caso las reglas no se pueden activar y hay que decidirlo con el plan).

Las variables de abajo son las únicas que hay que cambiar.

```bash
export REPO="Irico17/Areas-verdes-pucp"     # en la organización: GRUPO-12-DP2/<nombre-del-repo>
export REVISOR="Irico17"                    # usuario de GitHub que aprueba producción
export REVISOR_ID="$(gh api "users/$REVISOR" -q .id)"
```

## 1. Environment `produccion` con revisor obligatorio

El despliegue a producción (`.github/workflows/deploy.yml`, job `deploy` con `environment: produccion`) se detiene hasta que un revisor lo aprueba. Se limita además a la rama `develop` (el workflow se lanza con `--ref develop`).

```bash
# Crea o actualiza el environment con un revisor.
# prevent_self_review=false: si el revisor es la misma persona que lanza el despliegue, puede aprobarlo.
# Con más de una persona en el equipo conviene poner true.
gh api -X PUT "repos/$REPO/environments/produccion" --input - <<EOF2
{
  "reviewers": [ { "type": "User", "id": $REVISOR_ID } ],
  "prevent_self_review": false,
  "deployment_branch_policy": { "protected_branches": false, "custom_branch_policies": true }
}
EOF2

# Solo se puede desplegar a producción desde develop.
gh api -X POST "repos/$REPO/environments/produccion/deployment-branch-policies" \
  -f name=develop -f type=branch

# Comprobación
gh api "repos/$REPO/environments/produccion" -q '.protection_rules'
gh api "repos/$REPO/environments/produccion/deployment-branch-policies" -q '.branch_policies[].name'
```

Los environments `develop` y `qa` no llevan revisores. `qa` admite `develop` y los tags `rc-*`:

```bash
gh api -X PUT "repos/$REPO/environments/qa" --input - <<'EOF2'
{ "deployment_branch_policy": { "protected_branches": false, "custom_branch_policies": true } }
EOF2
gh api -X POST "repos/$REPO/environments/qa/deployment-branch-policies" -f name=develop -f type=branch
gh api -X POST "repos/$REPO/environments/qa/deployment-branch-policies" -f name='rc-*' -f type=tag
```

## 2. Protección de `main`

Sin force push, sin borrado de la rama y con los checks `test`, `backend`, `escaneo` y `e2e` (jobs de `.github/workflows/ci.yml`) obligatorios. `enforce_admins` queda en `false` a propósito: quien administra el repositorio puede seguir avanzando `main` con un fast-forward directo (por ejemplo, desde la API) sin abrir un PR. No se exige PR ni revisión, porque el flujo del equipo avanza `main` por fast-forward desde `develop`.

```bash
gh api -X PUT "repos/$REPO/branches/main/protection" --input - <<'EOF2'
{
  "required_status_checks": { "strict": false, "contexts": ["test", "backend", "escaneo", "e2e"] },
  "enforce_admins": false,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": false,
  "required_conversation_resolution": false
}
EOF2

# Comprobación
gh api "repos/$REPO/branches/main/protection" \
  -q '{checks: .required_status_checks.contexts, enforce_admins: .enforce_admins.enabled, force_push: .allow_force_pushes.enabled, deletions: .allow_deletions.enabled}'
```

Notas:

- Los nombres de `contexts` son los nombres de los jobs (`test`, `backend`, `escaneo`, `e2e`). `imagenes` no se exige: solo corre en `develop`, `backend/arquitectura-equipo` y tags `rc-*`, y en `main` queda `skipped`.
- Si un commit solo cambia `*.md` o `docs/**`, el CI no corre (`paths-ignore`). Para un administrador con `enforce_admins=false` eso no bloquea el push; para quien no sea administrador, un push directo sin checks se rechaza.
- Para endurecerlo más tarde: poner `enforce_admins` en `true`, añadir `required_pull_request_reviews` y usar PR.

## 3. Quitar las protecciones (marcha atrás)

```bash
gh api -X DELETE "repos/$REPO/branches/main/protection"
gh api -X PUT "repos/$REPO/environments/produccion" --input - <<'EOF2'
{ "reviewers": [], "prevent_self_review": false }
EOF2
```

## 4. Estado aplicado en `Irico17/Areas-verdes-pucp`

| Elemento | Valor |
| --- | --- |
| Environment `produccion` | revisor obligatorio `Irico17`, `prevent_self_review=false`, solo rama `develop` |
| Environment `qa` | ramas `develop` y tags `rc-*` |
| Rama `main` | checks `test`, `backend`, `escaneo`, `e2e`; sin force push; sin borrado; `enforce_admins=false` |

## 5. Lo que estas reglas no cubren

- No sustituyen la validación institucional con la PUCP.
- Un revisor que también lanza el despliegue puede aprobarse a sí mismo mientras `prevent_self_review` sea `false`.
- Los secretos del repositorio (`gh secret set`) y las variables de Actions se gestionan aparte; nunca se escriben en el repositorio.
