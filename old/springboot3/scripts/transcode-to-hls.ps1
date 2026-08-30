param(
    [string]$basePath = "d:/files/mybilibili/uploads"
)

$cmd = Get-Command ffmpeg -ErrorAction SilentlyContinue
if (-not $cmd) {
    Write-Host "[ERROR] ffmpeg not found"
    exit 1
}
Write-Host "[OK] ffmpeg ready"

$manuscriptsPath = Join-Path $basePath "manuscripts"
Write-Host "[INFO] Scanning: $manuscriptsPath"

$videoDirs = Get-ChildItem -Path $manuscriptsPath -Directory -ErrorAction SilentlyContinue
Write-Host "[INFO] Found manuscripts: $($videoDirs.Count)"

$qualities = @(
    @{name="480p"; scale="854:480"},
    @{name="720p"; scale="1280:720"},
    @{name="1080p"; scale="1920:1080"}
)

$count = 0
$success = 0
$failed = 0

foreach ($manuscriptDir in $videoDirs) {
    $videosDir = Join-Path $manuscriptDir.FullName "videos"
    if (-not (Test-Path $videosDir)) { continue }

    $videoDirs2 = Get-ChildItem -Path $videosDir -Directory -ErrorAction SilentlyContinue
    foreach ($videoDir in $videoDirs2) {
        $sourceDir = Join-Path $videoDir.FullName "source"
        $transcodedDir = Join-Path $videoDir.FullName "transcoded"

        # 查找源视频
        if (-not (Test-Path $sourceDir)) { continue }
        $sourceFiles = Get-ChildItem -Path $sourceDir -File -ErrorAction SilentlyContinue | Where-Object { $_.Extension -in @('.mp4', '.avi', '.mov', '.mkv') }
        if (-not $sourceFiles) { continue }
        $sourceVideo = $sourceFiles[0]

        $videoId = $videoDir.Name
        $manuscriptId = $manuscriptDir.Name

        # 删除旧的码率文件夹
        foreach ($q in $qualities) {
            $qDir = Join-Path $transcodedDir $q.name
            if (Test-Path $qDir) {
                Write-Host "[Cleanup] Removing: $qDir" -ForegroundColor DarkYellow
                Remove-Item -Path $qDir -Recurse -Force
            }
        }

        Write-Host ""
        Write-Host "[Processing] Video: $videoId, Manuscript: $manuscriptId" -ForegroundColor Cyan
        Write-Host "  Source: $($sourceVideo.FullName)" -ForegroundColor Gray

        # 转码三个码率
        foreach ($q in $qualities) {
            $qDir = Join-Path $transcodedDir $q.name
            New-Item -ItemType Directory -Path $qDir -Force | Out-Null

            $m3u8File = Join-Path $qDir "playlist.m3u8"

            Write-Host "  [$($q.name)] Transcoding..." -ForegroundColor Yellow

            $argString = "-i `"$($sourceVideo.FullName)`" -vf scale=w=$($q.scale):force_original_aspect_ratio=decrease -c:v libx264 -preset fast -b:v 2000k -c:a aac -b:a 128k -hls_time 10 -hls_list_size 0 -hls_segment_filename `"$qDir\%03d.ts`" -y `"$m3u8File`""

            $process = Start-Process -FilePath "cmd" -ArgumentList "/c ffmpeg $argString" -NoNewWindow -Wait -PassThru

            if ($process.ExitCode -eq 0 -and (Test-Path $m3u8File)) {
                Write-Host "  [$($q.name)] Success" -ForegroundColor Green
                $success++
            } else {
                Write-Host "  [$($q.name)] Failed" -ForegroundColor Red
                $failed++
            }
            $count++
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor White
Write-Host "Done! Total: $count, Success: $success, Failed: $failed" -ForegroundColor White
Write-Host "========================================" -ForegroundColor White
