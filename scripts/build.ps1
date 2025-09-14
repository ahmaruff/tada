# scripts/build.ps1
Write-Host "Building Tada binaries..."

$platforms = @(
    @{os="windows"; arch="amd64"; ext=".exe"},
    @{os="linux"; arch="amd64"; ext="" },
    @{os="darwin"; arch="amd64"; ext="" },
    @{os="darwin"; arch="arm64"; ext="" }  # M1 Macs
)

# Create dist directory
New-Item -ItemType Directory -Force -Path "dist" | Out-Null

foreach ($platform in $platforms) {
    $output = "dist/tada-$($platform.os)-$($platform.arch)$($platform.ext)"
    Write-Host "Building $output..."

    # Set env vars just for this build
    $env:GOOS = $platform.os
    $env:GOARCH = $platform.arch

    go build -ldflags "-s -w" -o $output
}

# Reset back to Windows defaults
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "Build complete! Check dist/ directory"

