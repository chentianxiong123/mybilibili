# Verifies Feign boundaries after backend service aggregation.
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$violations = New-Object System.Collections.Generic.List[string]

function Normalize-Path {
    param([string]$Path)
    return $Path.Replace("\", "/")
}

function Add-Violation {
    param([string]$Message)
    [void]$violations.Add($Message)
}

$runtimeServices = @{
    "mybilibili-account-social" = "mybilibili-account-social"
    "mybilibili-video-media" = "mybilibili-video-media"
    "mybilibili-content-interaction" = "mybilibili-content-interaction"
    "mybilibili-search-recommend" = "mybilibili-search-recommend"
    "mybilibili-ai" = "mybilibili-ai"
}

$legacyPackageAggregate = @{
    "com.mybilibili.user" = "mybilibili-account-social"
    "com.mybilibili.message" = "mybilibili-account-social"
    "com.mybilibili.video" = "mybilibili-video-media"
    "com.mybilibili.live" = "mybilibili-video-media"
    "com.mybilibili.comment" = "mybilibili-content-interaction"
    "com.mybilibili.interaction" = "mybilibili-content-interaction"
    "com.mybilibili.danmaku" = "mybilibili-content-interaction"
    "com.mybilibili.search" = "mybilibili-search-recommend"
    "com.mybilibili.analytics" = "mybilibili-search-recommend"
    "com.mybilibili.ai" = "mybilibili-ai"
}

$moduleRoots = @(
    "mybilibili-account-social",
    "mybilibili-video-media",
    "mybilibili-content-interaction",
    "mybilibili-search-recommend",
    "mybilibili-ai"
)

$scannedFeignCount = 0
$internalFeignCount = 0

foreach ($moduleRoot in $moduleRoots) {
    $sourceRoot = Join-Path (Join-Path $repoRoot $moduleRoot) "src/main/java"
    if (-not (Test-Path -LiteralPath $sourceRoot)) {
        Add-Violation "$moduleRoot has no src/main/java directory."
        continue
    }

    $javaFiles = Get-ChildItem -LiteralPath $sourceRoot -Recurse -Filter *.java -File
    foreach ($file in $javaFiles) {
        $content = Get-Content -LiteralPath $file.FullName -Raw
        if ($content -notmatch "@FeignClient") {
            continue
        }

        $scannedFeignCount++
        $relativePath = Normalize-Path $file.FullName.Substring($repoRoot.Length + 1)
        $nameMatch = [regex]::Match($content, '(?s)@FeignClient\s*\(.*?name\s*=\s*"([^"]+)"')
        if (-not $nameMatch.Success) {
            Add-Violation "$relativePath declares @FeignClient without an explicit name."
            continue
        }

        $targetService = $nameMatch.Groups[1].Value
        if (-not $runtimeServices.ContainsKey($targetService)) {
            Add-Violation "$relativePath targets non-runtime service $targetService."
            continue
        }

        $packageMatch = [regex]::Match($content, '(?m)^\s*package\s+([^;]+);')
        if (-not $packageMatch.Success) {
            Add-Violation "$relativePath has no package declaration."
            continue
        }

        $packageName = $packageMatch.Groups[1].Value
        $currentAggregate = $moduleRoot
        foreach ($entry in $legacyPackageAggregate.GetEnumerator()) {
            if ($packageName -eq $entry.Key -or $packageName.StartsWith($entry.Key + ".")) {
                $currentAggregate = $entry.Value
                break
            }
        }

        if ($targetService -eq $currentAggregate) {
            $internalFeignCount++
            Add-Violation "$relativePath introduces an internal Feign call to $targetService. Use a local service/facade/port instead."
        }
    }
}

if (Test-Path -LiteralPath (Join-Path $repoRoot "legacy-services")) {
    Add-Violation "legacy-services must not be scanned for Feign boundaries; retired services belong in .trash/."
}

if ($violations.Count -gt 0) {
    Write-Host "[FAIL] Feign boundary check failed." -ForegroundColor Red
    foreach ($violation in ($violations | Sort-Object)) {
        Write-Host "  - $violation" -ForegroundColor Red
    }
    exit 1
}

Write-Host "[OK] Feign boundary check passed." -ForegroundColor Green
Write-Host "Scanned Feign clients: $scannedFeignCount"
Write-Host "Internal Feign violations: $internalFeignCount"
