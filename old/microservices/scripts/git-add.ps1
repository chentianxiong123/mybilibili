# Shows changed files after service aggregation cleanup.
$ErrorActionPreference = "Stop"

Write-Host "Current git changes:"
git status --short
