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
    "mybilibili-account-social" = @{
        Aggregate = "mybilibili-account-social"
        SourceRoot = "mybilibili-account-social"
    }
    "mybilibili-user" = @{
        Aggregate = "mybilibili-account-social"
        SourceRoot = "legacy-services/mybilibili-user"
    }
    "mybilibili-message" = @{
        Aggregate = "mybilibili-account-social"
        SourceRoot = "legacy-services/mybilibili-message"
    }

    "mybilibili-video-media" = @{
        Aggregate = "mybilibili-video-media"
        SourceRoot = "mybilibili-video-media"
    }
    "mybilibili-video" = @{
        Aggregate = "mybilibili-video-media"
        SourceRoot = "legacy-services/mybilibili-video"
    }
    "mybilibili-live" = @{
        Aggregate = "mybilibili-video-media"
        SourceRoot = "legacy-services/mybilibili-live"
    }

    "mybilibili-content-interaction" = @{
        Aggregate = "mybilibili-content-interaction"
        SourceRoot = "mybilibili-content-interaction"
    }
    "mybilibili-comment" = @{
        Aggregate = "mybilibili-content-interaction"
        SourceRoot = "legacy-services/mybilibili-comment"
    }
    "mybilibili-interaction" = @{
        Aggregate = "mybilibili-content-interaction"
        SourceRoot = "legacy-services/mybilibili-interaction"
    }
    "mybilibili-danmaku" = @{
        Aggregate = "mybilibili-content-interaction"
        SourceRoot = "legacy-services/mybilibili-danmaku"
    }

    "mybilibili-search-recommend" = @{
        Aggregate = "mybilibili-search-recommend"
        SourceRoot = "mybilibili-search-recommend"
    }
    "mybilibili-search" = @{
        Aggregate = "mybilibili-search-recommend"
        SourceRoot = "legacy-services/mybilibili-search"
    }
    "mybilibili-analytics" = @{
        Aggregate = "mybilibili-search-recommend"
        SourceRoot = "legacy-services/mybilibili-analytics"
    }

    "mybilibili-ai" = @{
        Aggregate = "mybilibili-ai"
        SourceRoot = "mybilibili-ai"
    }
}

$knownInternalFeign = @{
    "legacy-services/mybilibili-message/src/main/java/com/mybilibili/message/feign/UserClient.java" = @{
        Adapter = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/local/MessageUserClientLocalAdapter.java"
        Application = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/AccountSocialApplication.java"
        ForbiddenClientRef = "com.mybilibili.message.feign.UserClient.class"
        ForbiddenBasePackage = "com.mybilibili.message.feign"
        ForbiddenImport = "import com.mybilibili.message.feign.UserClient;"
        AdapterInterfaceName = "UserLookupPort"
    }
    "legacy-services/mybilibili-user/src/main/java/com/mybilibili/user/feign/UserClient.java" = @{
        Adapter = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/local/UserSelfClientLocalAdapter.java"
        Application = "mybilibili-account-social/src/main/java/com/mybilibili/accountsocial/AccountSocialApplication.java"
        ForbiddenClientRef = "com.mybilibili.user.feign.UserClient.class"
        ForbiddenBasePackage = "com.mybilibili.user.feign"
        ForbiddenImport = "import com.mybilibili.user.feign.UserClient;"
        AdapterInterfaceName = "UserLookupPort"
    }
    "legacy-services/mybilibili-comment/src/main/java/com/mybilibili/comment/feign/LikeClient.java" = @{
        Adapter = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/local/CommentLikeClientLocalAdapter.java"
        Application = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/ContentInteractionApplication.java"
        ForbiddenClientRef = "com.mybilibili.comment.feign.LikeClient.class"
        ForbiddenBasePackage = "com.mybilibili.comment.feign"
        ForbiddenImport = "import com.mybilibili.comment.feign.LikeClient;"
        AdapterInterfaceName = "LikeInteractionPort"
    }
    "legacy-services/mybilibili-comment/src/main/java/com/mybilibili/comment/feign/DynamicClient.java" = @{
        Adapter = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/local/CommentDynamicClientLocalAdapter.java"
        Application = "mybilibili-content-interaction/src/main/java/com/mybilibili/contentinteraction/ContentInteractionApplication.java"
        ForbiddenClientRef = "com.mybilibili.comment.feign.DynamicClient.class"
        ForbiddenBasePackage = "com.mybilibili.comment.feign"
        ForbiddenImport = "import com.mybilibili.comment.feign.DynamicClient;"
        AdapterInterfaceName = "DynamicInteractionPort"
    }
}

$scannedFeignCount = 0
$internalFeignCount = 0

foreach ($moduleName in ($moduleAggregate.Keys | Sort-Object)) {
    $module = $moduleAggregate[$moduleName]
    $sourceRoot = Join-Path (Join-Path $repoRoot $module.SourceRoot) "src/main/java"
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
        $currentAggregate = $module.Aggregate
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

foreach ($moduleName in ($moduleAggregate.Keys | Sort-Object)) {
    $module = $moduleAggregate[$moduleName]
    $sourceRoot = Join-Path (Join-Path $repoRoot $module.SourceRoot) "src/main/java"
    if (-not (Test-Path -LiteralPath $sourceRoot)) {
        continue
    }

    $javaFiles = Get-ChildItem -LiteralPath $sourceRoot -Recurse -Filter *.java -File
    foreach ($file in $javaFiles) {
        $content = Get-Content -LiteralPath $file.FullName -Raw
        $relativePath = Normalize-Path $file.FullName.Substring($repoRoot.Length + 1)
        foreach ($entry in $knownInternalFeign.GetEnumerator()) {
            $metadata = $entry.Value
            if ($relativePath -ne $entry.Key -and $content.Contains($metadata.ForbiddenImport)) {
                Add-Violation "$relativePath imports legacy internal Feign $($metadata.ForbiddenImport). Use the local port instead."
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
        if ($adapterContent -notmatch ("implements\s+" + [regex]::Escape($metadata.AdapterInterfaceName))) {
            Add-Violation "$adapterPath must implement $($metadata.AdapterInterfaceName)."
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
