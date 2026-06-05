# Verifies that the default runtime uses only the aggregated backend services.
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$violations = New-Object System.Collections.Generic.List[string]

function Add-Violation {
    param([string]$Message)
    [void]$violations.Add($Message)
}

function Read-Text {
    param([string]$RelativePath)
    $path = Join-Path $repoRoot $RelativePath
    return Get-Content -LiteralPath $path -Raw
}

$runtimeBackends = @(
    "mybilibili-gateway",
    "mybilibili-account-social",
    "mybilibili-video-media",
    "mybilibili-content-interaction",
    "mybilibili-search-recommend",
    "mybilibili-ai"
)

$gatewayTargets = @(
    "mybilibili-account-social",
    "mybilibili-video-media",
    "mybilibili-content-interaction",
    "mybilibili-search-recommend",
    "mybilibili-ai"
)

$legacyBackends = @(
    "mybilibili-user",
    "mybilibili-video",
    "mybilibili-danmaku",
    "mybilibili-search",
    "mybilibili-comment",
    "mybilibili-interaction",
    "mybilibili-message",
    "mybilibili-live",
    "mybilibili-analytics"
)

$gatewayConfigPath = "mybilibili-gateway/src/main/resources/application.yml"
$gatewayConfig = Read-Text $gatewayConfigPath
$routeMatches = [regex]::Matches($gatewayConfig, "lb://(mybilibili-[A-Za-z0-9-]+)")
$actualGatewayTargets = New-Object System.Collections.Generic.HashSet[string]

foreach ($match in $routeMatches) {
    $target = $match.Groups[1].Value
    [void]$actualGatewayTargets.Add($target)
    if ($gatewayTargets -notcontains $target) {
        Add-Violation "$gatewayConfigPath routes to non-runtime service $target."
    }
}

foreach ($expectedTarget in $gatewayTargets) {
    if (-not $actualGatewayTargets.Contains($expectedTarget)) {
        Add-Violation "$gatewayConfigPath does not route to expected runtime service $expectedTarget."
    }
}

$startupScripts = @(
    "scripts/start-all.ps1",
    "scripts/start-all.bat",
    "scripts/test-services.sh"
)

foreach ($scriptPath in $startupScripts) {
    $content = Read-Text $scriptPath
    foreach ($backend in $runtimeBackends) {
        if (-not $content.Contains($backend)) {
            Add-Violation "$scriptPath does not mention runtime backend $backend."
        }
    }

    foreach ($legacyBackend in $legacyBackends) {
        $startsLegacyBackend =
            $content -match ("Start-MavenService\s+`"[^`"]*`"\s+`"" + [regex]::Escape($legacyBackend) + "`"") -or
            $content -match ("cd\s+(/d\s+)?[^`r`n]*" + [regex]::Escape($legacyBackend) + "(?=($|[\\/\s`"']))[^`r`n]*(&&|;)?\s*mvn\s+spring-boot:run") -or
            $content -match ("cd\s+" + [regex]::Escape($legacyBackend) + "(?=($|[\\/\s`"']))[^`r`n]*") -and
            $content -match "mvn\s+spring-boot:run"

        if ($startsLegacyBackend) {
            Add-Violation "$scriptPath appears to start legacy backend $legacyBackend."
        }
    }
}

if ($violations.Count -gt 0) {
    Write-Host "[FAIL] Aggregation runtime check failed." -ForegroundColor Red
    foreach ($violation in ($violations | Sort-Object)) {
        Write-Host "  - $violation" -ForegroundColor Red
    }
    exit 1
}

Write-Host "[OK] Aggregation runtime check passed." -ForegroundColor Green
Write-Host "Gateway route targets: $($actualGatewayTargets.Count)"
Write-Host "Runtime backends: $($runtimeBackends.Count)"
