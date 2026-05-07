param(
    [string]$OutputDir = "bin"
)

$ErrorActionPreference = "Stop"

function Get-GoOS {
    $os = $env:OS
    if ($os -eq "Windows_NT") {
        return "windows"
    }
    if (Get-Command uname -ErrorAction SilentlyContinue) {
        $uname = (uname -s)
        if ($uname -match "Linux") {
            return "linux"
        }
        if ($uname -match "Darwin") {
            return "darwin"
        }
    }
    return "unknown"
}

function Get-GoArch {
    if ($env:OS -eq "Windows_NT") {
        if ([Environment]::Is64BitOperatingSystem) {
            $arch = $env:PROCESSOR_ARCHITECTURE
            if ($arch -eq "ARM64") {
                return "arm64"
            }
            return "amd64"
        }
        return "386"
    }
    if (Get-Command uname -ErrorAction SilentlyContinue) {
        $cpu = (uname -m)
        if ($cpu -eq "aarch64" -or $cpu -eq "arm64") {
            return "arm64"
        }
        if ($cpu -eq "x86_64") {
            return "amd64"
        }
        if ($cpu -match "i386" -or $cpu -match "i686") {
            return "386"
        }
    }
    return "amd64"
}

$goos = Get-GoOS
$goarch = Get-GoArch

Write-Host "Detected platform: $goos/$goarch"

$ext = if ($goos -eq "windows") { ".exe" } else { "" }
$outputName = "acr-uploader-${goos}-${goarch}${ext}"
$outputPath = Join-Path $OutputDir $outputName

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$env:GOOS = $goos
$env:GOARCH = $goarch

Write-Host "Building $outputPath..."
go build -ldflags="-s -w" -gcflags="all=-l" -o $outputPath main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful! Output: $outputPath"
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}