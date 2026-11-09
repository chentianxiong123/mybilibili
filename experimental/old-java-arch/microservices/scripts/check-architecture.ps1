# Runs local architecture checks that keep the aggregated backend shape stable.
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$checks = @(
    "check-feign-boundaries.ps1",
    "check-aggregation-runtime.ps1"
)

foreach ($check in $checks) {
    $checkPath = Join-Path $scriptDir $check
    Write-Host "Running $check ..."
    & powershell -NoProfile -ExecutionPolicy Bypass -File $checkPath
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
    Write-Host ""
}

Write-Host "[OK] Architecture checks passed." -ForegroundColor Green
