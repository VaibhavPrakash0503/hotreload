# build.ps1 — Windows build script for Hotreload
# Usage: .\build.ps1 [target]
# Targets: build (default), test, clean

param(
    [string]$Target = "build"
)

$BinaryName = "hotreload"
$CmdPkg     = "./cmd/hotreload"
$BinDir     = "bin"
$Version    = & git describe --tags --always --dirty 2>$null
if (-not $Version) { $Version = "dev" }
$Ldflags    = "-X main.version=$Version -s -w"

$null = New-Item -ItemType Directory -Force -Path $BinDir

switch ($Target) {
    "build" {
        Write-Host "Building hotreload for Windows..."
        & go build -ldflags $Ldflags -o "$BinDir\$BinaryName.exe" $CmdPkg
    }
    "test" {
        & go test ./internal/... -v -count=1
    }
    "clean" {
        Remove-Item -Recurse -Force $BinDir, "coverage.out" -ErrorAction SilentlyContinue
        Write-Host "Cleaned."
    }
    default {
        Write-Host "Unknown target '$Target'. Valid targets: build, test, clean"
    }
}
