<#
.SYNOPSIS
  Carga en GitHub los secretos y variables definidos en .env-github.

.DESCRIPTION
  Lee .env-github (secciones [ambito:tipo], lineas NOMBRE=valor) y ejecuta
  gh secret set / gh variable set. Ambitos: repo, develop, qa, produccion, aws-lab.
  Tipos: secrets o variables.
  No imprime ningun valor y no escribe archivos. Omite lineas vacias, comentarios,
  valores vacios, entre < > o PENDIENTE.
  Este script no contiene valores. El archivo .env-github se obtiene de quien administra el repo
  y NO se versiona.

.EXAMPLE
  .\scripts\cargar-secretos-github.ps1 -DryRun
  .\scripts\cargar-secretos-github.ps1
  .\scripts\cargar-secretos-github.ps1 -Ambito aws-lab
#>
param(
    [string]$Archivo = "",
    [string]$Repo = "GRUPO-12-DP2/-areas-verdes-pucp",
    [string]$Ambito = "",
    [switch]$DryRun
)
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
if (-not $Archivo) { $Archivo = Join-Path $Root ".env-github" }
if (-not (Test-Path $Archivo)) { Write-Error "No existe $Archivo. Pidelo a quien administra el repositorio." }
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) { Write-Error "Falta el comando gh. Instalalo y entra con: gh auth login" }

$ambientes = @("develop", "qa", "produccion", "aws-lab")
$entornosListos = @{}
$ok = 0; $omitidos = 0; $fallos = 0
$ambito = ""; $tipo = ""

foreach ($cruda in (Get-Content -Path $Archivo -Encoding UTF8)) {
    $linea = $cruda.TrimEnd("`r")
    if ($linea.Trim() -eq "" -or $linea.TrimStart().StartsWith("#")) { continue }

    if ($linea -match '^\[([a-z0-9-]+):(secrets|variables)\]\s*$') {
        $ambito = $Matches[1]; $tipo = $Matches[2]
        if ($ambito -ne "repo" -and $ambientes -notcontains $ambito) { Write-Error "Ambito desconocido: $ambito" }
        continue
    }
    if (-not $ambito) { continue }
    if ($Ambito -and $Ambito -ne $ambito) { continue }

    $i = $linea.IndexOf("=")
    if ($i -le 0) { continue }
    $nombre = $linea.Substring(0, $i).Trim()
    $valor = $linea.Substring($i + 1).Trim()
    if ($nombre -notmatch '^[A-Za-z_][A-Za-z0-9_]*$') { Write-Warning "Nombre no valido omitido"; continue }
    if ($valor -eq "" -or $valor -eq "PENDIENTE" -or $valor -match '^<.*>$') {
        Write-Host "OMITIDO   [$ambito`:$tipo] $nombre (sin valor)"
        $omitidos++
        continue
    }

    if ($DryRun) {
        Write-Host "CARGARIA  [$ambito`:$tipo] $nombre"
        continue
    }

    try {
        if ($ambito -ne "repo" -and -not $entornosListos.ContainsKey($ambito)) {
            gh api -X PUT "repos/$Repo/environments/$ambito" --silent
            if ($LASTEXITCODE -ne 0) { throw "no se pudo crear el environment $ambito" }
            $entornosListos[$ambito] = $true
        }
        $extra = @()
        if ($ambito -ne "repo") { $extra = @("--env", $ambito) }
        if ($tipo -eq "secrets") {
            $valor | gh secret set $nombre -R $Repo @extra
        } else {
            gh variable set $nombre --body $valor -R $Repo @extra
        }
        if ($LASTEXITCODE -ne 0) { throw "gh devolvio $LASTEXITCODE" }
        Write-Host "CARGADO   [$ambito`:$tipo] $nombre"
        $ok++
    } catch {
        Write-Host "FALLO     [$ambito`:$tipo] $nombre : $($_.Exception.Message)"
        $fallos++
    }
}
Write-Host ""
Write-Host "Cargados: $ok  Omitidos: $omitidos  Fallos: $fallos"
if ($fallos -gt 0) { exit 1 }
