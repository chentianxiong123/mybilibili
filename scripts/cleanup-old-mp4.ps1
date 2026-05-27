param(
    [string]$basePath = "d:/files/mybilibili/uploads"
)

Write-Host "[INFO] Scanning for old MP4 files in: $basePath" -ForegroundColor Yellow

$mp4Files = Get-ChildItem -Path "$basePath/manuscripts" -Recurse -Include *.mp4 -File -ErrorAction SilentlyContinue | Where-Object { $_.FullName -like "*/transcoded/*" }

if (-not $mp4Files) {
    Write-Host "[OK] No transcoded MP4 files found" -ForegroundColor Green
    exit 0
}

Write-Host "[INFO] Found $($mp4Files.Count) transcoded MP4 files" -ForegroundColor Cyan

$count = 0
foreach ($file in $mp4Files) {
    Write-Host "  Deleting: $($file.FullName)" -ForegroundColor Gray
    Remove-Item -Path $file.FullName -Force
    $count++
}

Write-Host ""
Write-Host "[OK] Deleted $count transcoded MP4 files" -ForegroundColor Green
