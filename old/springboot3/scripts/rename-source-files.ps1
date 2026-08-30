param(
    [string]$basePath = "d:/files/mybilibili/uploads"
)

Write-Host "[INFO] Renaming source files to 'video' in: $basePath" -ForegroundColor Yellow

$sourceFiles = Get-ChildItem -Path "$basePath/manuscripts" -Recurse -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -eq "source" }

if (-not $sourceFiles) {
    Write-Host "[OK] No source directories found" -ForegroundColor Green
    exit 0
}

Write-Host "[INFO] Found $($sourceFiles.Count) source directories" -ForegroundColor Cyan

$count = 0
foreach ($sourceDir in $sourceFiles) {
    $sourceFilesInDir = Get-ChildItem -Path $sourceDir.FullName -File -ErrorAction SilentlyContinue
    foreach ($file in $sourceFilesInDir) {
        $newName = Join-Path $sourceDir.FullName "video" + $file.Extension
        Write-Host "  Renaming: $($file.FullName) -> $newName" -ForegroundColor Gray
        Rename-Item -Path $file.FullName -NewName ("video" + $file.Extension) -Force
        $count++
    }
}

Write-Host ""
Write-Host "[OK] Renamed $count files" -ForegroundColor Green
