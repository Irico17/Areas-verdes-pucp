# Sube el bloque «AWS CLI» del Learner Lab al environment de GitHub aws-lab.
# Lee stdin si hay datos canalizados, y si no, el portapapeles.
# No imprime los valores y no los escribe en el repositorio.
#
#   Get-Clipboard -Raw | .\scripts\actualizar-credenciales-lab.ps1
[CmdletBinding()]
param(
    [string]$EnvironmentName = "aws-lab"
)

$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Error "Falta el comando gh. Instálelo y entre con: gh auth login"
}

function Valor-Asignacion([string]$Linea) {
    $raw = ($Linea -replace '^[^=]*=', '').Trim()
    if ($raw.Length -ge 2) {
        $q = $raw.Substring(0, 1)
        if (($q -eq '"' -or $q -eq "'") -and $raw.Substring($raw.Length - 1, 1) -eq $q) {
            $raw = $raw.Substring(1, $raw.Length - 2)
        }
    }
    if ($raw -match '\s') {
        throw "Un valor del bloque contiene espacios."
    }
    return $raw
}

function Parsear-Bloque([string]$Texto) {
    $key = ""
    $secret = ""
    $token = ""
    $region = ""
    foreach ($cruda in ($Texto -split "`r?`n")) {
        $linea = $cruda.TrimEnd("`r").Trim()
        if ([string]::IsNullOrWhiteSpace($linea) -or $linea.StartsWith("#")) { continue }
        $limpia = $linea -replace '^(?i)export\s+', ''
        $nombre = ($limpia -replace '=.*$', '').Trim().ToLowerInvariant()
        $valor = Valor-Asignacion $limpia
        switch ($nombre) {
            "aws_access_key_id" { $key = $valor }
            "aws_secret_access_key" { $secret = $valor }
            "aws_session_token" { $token = $valor }
            { $_ -in @("aws_default_region", "aws_region", "region") } { $region = $valor }
        }
    }
    return @{ Key = $key; Secret = $secret; Token = $token; Region = $region }
}

$texto = ""
if ([Console]::IsInputRedirected) {
    $texto = [Console]::In.ReadToEnd()
} else {
    try { $texto = Get-Clipboard -Raw -ErrorAction SilentlyContinue } catch { $texto = "" }
    if ([string]::IsNullOrWhiteSpace($texto)) {
        Write-Host "Pegue el bloque de AWS Details, AWS CLI, Show, y pulse Enter. No se muestra."
        $texto = Get-Clipboard -Raw
    }
}

$creds = Parsear-Bloque $texto
$texto = $null

if (-not $creds.Key -or -not $creds.Secret -or -not $creds.Token) {
    Write-Error "El bloque no trae aws_access_key_id, aws_secret_access_key y aws_session_token."
}
if (-not $creds.Region) { $creds.Region = "us-east-1" }
if ($creds.Region -notin @("us-east-1", "us-west-2")) {
    Write-Error "La región debe ser us-east-1 o us-west-2 (Learner Lab)."
}
if ($creds.Key.Length -lt 16 -or $creds.Secret.Length -lt 16 -or $creds.Token.Length -lt 16) {
    Write-Error "Alguna credencial es demasiado corta."
}

function Subir([string]$Nombre, [string]$Valor) {
    $Valor | & gh secret set $Nombre --env $EnvironmentName | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "gh secret set $Nombre falló" }
}

Subir "AWS_ACCESS_KEY_ID" $creds.Key
Subir "AWS_SECRET_ACCESS_KEY" $creds.Secret
Subir "AWS_SESSION_TOKEN" $creds.Token
& gh variable set AWS_REGION --env $EnvironmentName --body $creds.Region | Out-Null
if ($LASTEXITCODE -ne 0) { throw "gh variable set AWS_REGION falló" }

$region = $creds.Region
$creds = $null
Write-Host "Credenciales cargadas en el environment $EnvironmentName (región $region). Los valores no se muestran."
Write-Host "Caducan con la sesión del Learner Lab. Cuando Start Lab entregue otras, vuelva a ejecutar este script."
