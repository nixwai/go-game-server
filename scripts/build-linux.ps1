#Requires -Version 5.1
[CmdletBinding()]
param(
    [ValidateSet('amd64', 'arm64')]
    [string]$Architecture = 'amd64'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = [System.IO.Path]::GetFullPath((Split-Path -Parent $PSScriptRoot))
$binDir = Join-Path $repoRoot 'bin'

New-Item -ItemType Directory -Path $binDir -Force | Out-Null

$previousGOOS = $env:GOOS
$previousGOARCH = $env:GOARCH
$previousCGO = $env:CGO_ENABLED
$pushedLocation = $false

try {
    $env:GOOS = 'linux'
    $env:GOARCH = $Architecture
    $env:CGO_ENABLED = '0'

    Push-Location $repoRoot
    $pushedLocation = $true

    $targets = @(
        @{ Name = 'server'; Path = '.\cmd\server' },
        @{ Name = 'migrate'; Path = '.\cmd\migrate' },
        @{ Name = 'admin-init'; Path = '.\cmd\admin-init' }
    )

    foreach ($target in $targets) {
        $output = Join-Path $binDir $target.Name
        & go build -trimpath -ldflags '-s -w' -o $output $target.Path
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed: $($target.Path)"
        }
    }
}
finally {
    if ($pushedLocation) {
        Pop-Location
    }

    if ($null -eq $previousGOOS) { Remove-Item Env:GOOS -ErrorAction SilentlyContinue } else { $env:GOOS = $previousGOOS }
    if ($null -eq $previousGOARCH) { Remove-Item Env:GOARCH -ErrorAction SilentlyContinue } else { $env:GOARCH = $previousGOARCH }
    if ($null -eq $previousCGO) { Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue } else { $env:CGO_ENABLED = $previousCGO }
}

Get-ChildItem -LiteralPath $binDir