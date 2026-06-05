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

function Read-Text {
    param([string]$RelativePath)
    $path = Join-Path $repoRoot $RelativePath
    return Get-Content -LiteralPath $path -Raw
}

$moduleAggregate = @{
    "mybilibili-account-social" = "mybilibili-account-social"
    "mybilibili-user" = "mybilibili-account-social"
    "mybilibili-message" = "mybilibili-account-social"

    "mybilibili-video-media" = "mybilibili-video-media"
    "mybilibili-video" = "mybilibili-video-media"
    "mybilibili-live" = "mybilibili-video-media"

    "mybilibili-content-interaction" = "mybilibili-content-interaction"
    "mybilibili-comment" = "mybilibili-content-interaction"
    "mybilibili-interaction" = "mybilibili-content-interaction"
    "mybilibili-danmaku" = "mybilibili-content-interaction"

    "mybilibili-search-recommend" = "mybilibili-search-recommend"
    "mybilibili-search" = "mybilibili-search-recommend"
    "mybilibili-analytics" = "mybilibili-search-recommend"

    "mybilibili-ai" = "mybilibili-ai"
}

$knownInternalFeign = @{
    "mybilibili-message/src/main/java/com/mybilibili/message/feign/UserClient.java" = @{
        Adapter = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/local/MessageUserClientLocalAdapter.java"
        Application = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/AccountSocialApplication.java"
        ForbiddenClientRef = "com.mybilibili.message.feign.UserClient.class"
        ForbiddenBasePackage = "com.mybilibili.message.feign"
        InterfaceName = "UserClient"
    }
    "mybilibili-user/src/main/java/com/mybilibili/user/feign/UserClient.java" = @{
        Adapter = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/local/UserSelfClientLocalAdapter.java"
        Application = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/AccountSocialApplication.java"
        ForbiddenClientRef = "com.mybilibili.user.feign.UserClient.class"
        ForbiddenBasePackage = "com.mybilibili.user.feign"
        InterfaceName = "UserClient"
    }
    "mybilibili-comment/src/main/java/com/mybilibili/comment/feign/LikeClient.java" = @{
        Adapter = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/local/CommentLikeClientLocalAdapter.java"
        Application = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/ContentInteractionApplication.java"
        ForbiddenClientRef = "com.mybilibili.comment.feign.LikeClient.class"
        ForbiddenBasePackage = "com.mybilibili.comment.feign"
        InterfaceName = "LikeClient"
    }
    "mybilibili-comment/src/main/java/com/mybilibili/comment/feign/DynamicClient.java" = @{
        Adapter = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/local/CommentDynamicClientLocalAdapter.java"
        Application = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/ContentInteractionApplication.java"
        ForbiddenClientRef = "com.mybilibili.comment.feign.DynamicClient.class"
        ForbiddenBasePackage = "com.mybilibili.comment.feign"
        InterfaceName = "DynamicClient"
    }
}

$scannedFeignCount = 0
$internalFeignCount = 0

foreach ($moduleName in ($moduleAggregate.Keys | Sort-Object)) {
    $sourceRoot = Join-Path (Join-Path $repoRoot $moduleName) "src/main/java"
    if (-not (Test-Path -LiteralPath $sourceRoot)) {
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
        $currentAggregate = $moduleAggregate[$moduleName]
        if ($targetService -eq $currentAggregate) {
            $internalFeignCount++
            if (-not $knownInternalFeign.ContainsKey($relativePath)) {
                Add-Violation "$relativePath introduces an internal Feign call to $targetService. Use a local service/facade instead."
                continue
            }

            if ($content -notmatch "@Deprecated\s*\(") {
                Add-Violation "$relativePath is a known legacy internal Feign contract but is not marked @Deprecated."
            }
        }
    }
}

foreach ($entry in $knownInternalFeign.GetEnumerator()) {
    $interfacePath = $entry.Key
    $metadata = $entry.Value
    $fullInterfacePath = Join-Path $repoRoot $interfacePath

    if (-not (Test-Path -LiteralPath $fullInterfacePath)) {
        Add-Violation "$interfacePath is missing."
        continue
    }

    $interfaceContent = Read-Text $interfacePath
    if ($interfaceContent -notmatch "@Deprecated\s*\(") {
        Add-Violation "$interfacePath must stay @Deprecated until the old standalone module is retired."
    }

    $adapterPath = $metadata.Adapter
    $fullAdapterPath = Join-Path $repoRoot $adapterPath
    if (-not (Test-Path -LiteralPath $fullAdapterPath)) {
        Add-Violation "$adapterPath is missing for $interfacePath."
    } else {
        $adapterContent = Read-Text $adapterPath
        if ($adapterContent -notmatch "@Component") {
            Add-Violation "$adapterPath must be a Spring component."
        }
        if ($adapterContent -notmatch ("implements\s+" + [regex]::Escape($metadata.InterfaceName))) {
            Add-Violation "$adapterPath must implement $($metadata.InterfaceName)."
        }
    }

    $applicationPath = $metadata.Application
    if (Test-Path -LiteralPath (Join-Path $repoRoot $applicationPath)) {
        $applicationContent = Read-Text $applicationPath
        if ($applicationContent.Contains($metadata.ForbiddenClientRef)) {
            Add-Violation "$applicationPath must not register $($metadata.ForbiddenClientRef) as a Feign client."
        }

        $quotedBasePackage = '"' + $metadata.ForbiddenBasePackage + '"'
        if ($applicationContent.Contains($quotedBasePackage)) {
            Add-Violation "$applicationPath must not scan $($metadata.ForbiddenBasePackage) as a Feign base package."
        }
    } else {
        Add-Violation "$applicationPath is missing."
    }
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
Write-Host "Legacy internal Feign contracts: $internalFeignCount"
