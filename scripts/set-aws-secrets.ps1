# Sube las credenciales temporales del Learner Lab a los secrets del repo.
# Lee el bloque del portapapeles (AWS Details > AWS CLI > Show) o, si no trae
# las tres claves, el perfil [default] de ~/.aws/credentials.
# No imprime los valores ni los escribe en el repositorio.
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Error "Falta el comando gh. Instálelo y entre con: gh auth login"
}

function Extraer([string]$Texto) {
    $key = ""
    $secret = ""
    $token = ""
    $enDefault = $false
    $vioSeccion = $false
    foreach ($cruda in ($Texto -split "`r?`n")) {
        $linea = $cruda.TrimEnd("`r")
        if ($linea -match '^\[(.+)\]$') {
            $vioSeccion = $true
            $enDefault = ($Matches[1] -eq "default")
            continue
        }
        if ($vioSeccion -and -not $enDefault) { continue }
        $limpia = $linea -replace '^export\s+', ''
        if ($limpia -match '^(?i:aws_access_key_id)\s*=\s*(.*)$') {
            $key = $Matches[1].Trim().Trim('"').Trim("'")
        } elseif ($limpia -match '^(?i:aws_secret_access_key)\s*=\s*(.*)$') {
            $secret = $Matches[1].Trim().Trim('"').Trim("'")
        } elseif ($limpia -match '^(?i:aws_session_token)\s*=\s*(.*)$') {
            $token = $Matches[1].Trim().Trim('"').Trim("'")
        }
    }
    return @{ Key = $key; Secret = $secret; Token = $token }
}

function Completas($c) {
    return $c.Key -and $c.Secret -and $c.Token
}

$fuente = ""
$creds = $null

try {
    $clip = Get-Clipboard -Raw -ErrorAction SilentlyContinue
} catch {
    $clip = $null
}
if ($clip) {
    $prova = Extraer $clip
    if (Completas $prova) {
        $creds = $prova
        $fuente = "portapapeles"
    }
}

if (-not $creds) {
    $ruta = if ($env:AWS_SHARED_CREDENTIALS_FILE) { $env:AWS_SHARED_CREDENTIALS_FILE } else { Join-Path $env:USERPROFILE ".aws\credentials" }
    if (Test-Path -LiteralPath $ruta) {
        $prova = Extraer (Get-Content -LiteralPath $ruta -Raw)
        if (Completas $prova) {
            $creds = $prova
            $fuente = "archivo"
        }
    }
}

if (-not $creds) {
    Write-Host "No hay tres claves en el portapapeles ni en ~/.aws/credentials."
    Write-Host "Pegue el bloque de AWS CLI > Show y termine con una línea vacía."
    $lineas = New-Object System.Collections.Generic.List[string]
    while ($true) {
        $linea = [Console]::ReadLine()
        if ([string]::IsNullOrEmpty($linea)) {
            if ($lineas.Count -gt 0) { break }
            continue
        }
        $lineas.Add($linea)
    }
    $creds = Extraer ($lineas -join "`n")
    $fuente = "pegado"
}

if (-not (Completas $creds)) {
    Write-Error "El bloque no trae aws_access_key_id, aws_secret_access_key y aws_session_token."
}

function Subir([string]$Nombre, [string]$Valor) {
    & gh secret set $Nombre --body $Valor
    if ($LASTEXITCODE -ne 0) {
        throw "gh secret set $Nombre falló"
    }
}

Subir "AWS_ACCESS_KEY_ID" $creds.Key
Subir "AWS_SECRET_ACCESS_KEY" $creds.Secret
Subir "AWS_SESSION_TOKEN" $creds.Token
$creds = $null
$clip = $null

Write-Host "Secrets actualizados desde $fuente (los valores no se muestran)."

if ([Environment]::UserInteractive -and -not [Console]::IsInputRedirected) {
    $respuesta = Read-Host "Lanzar el despliegue (gh workflow run ci --ref main)? [s/N]"
    if ($respuesta -match '^[sSyY]$') {
        & gh workflow run ci --ref main
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        Write-Host "Despliegue pedido en main. El job imprime la URL al terminar."
    } else {
        Write-Host "Para lanzarlo después: gh workflow run ci --ref main"
    }
} else {
    Write-Host "Para lanzarlo: gh workflow run ci --ref main"
}
