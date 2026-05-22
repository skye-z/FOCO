param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsList
)

$ErrorActionPreference = "Stop"

$script:RootDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envFile = Join-Path (Split-Path -Parent $script:RootDir) ".env"
$script:Processes = @()

function Show-Help {
    Write-Host "Usage: .\run.ps1 --t <target>"
    Write-Host ""
    Write-Host "Targets:"
    Write-Host "  1  Start admin frontend    (port 3001)"
    Write-Host "  2  Start learner frontend  (port 3000)"
    Write-Host "  3  Start backend API       (port 8080)"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run.ps1 --t 1"
    Write-Host "  .\run.ps1 --t 1 --t 3"
    Write-Host "  .\run.ps1 --t 1 --t 2 --t 3"
}

function Import-DotEnv {
    param([string]$FilePath)

    if (-not (Test-Path -LiteralPath $FilePath)) {
        return
    }

    foreach ($line in Get-Content -LiteralPath $FilePath) {
        $trimmed = $line.Trim()
        if (-not $trimmed -or $trimmed.StartsWith("#")) {
            continue
        }

        $parts = $trimmed -split "=", 2
        if ($parts.Count -ne 2) {
            continue
        }

        $name = $parts[0].Trim()
        $value = $parts[1].Trim()

        if ($value.Length -ge 2) {
            if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
                $value = $value.Substring(1, $value.Length - 2)
            }
        }

        [System.Environment]::SetEnvironmentVariable($name, $value)
    }
}

function Ensure-NodeDeps {
    param([string]$Dir)

    $nodeModules = Join-Path $Dir "node_modules"
    if (-not (Test-Path -LiteralPath $nodeModules)) {
        Write-Host "Installing dependencies for $Dir..."
        Push-Location $Dir
        try {
            & npm install --no-fund --no-audit
            if (-not $?) {
                throw "npm install failed for $Dir"
            }
        } finally {
            Pop-Location
        }
    }
}

function Start-ManagedProcess {
    param(
        [string]$Name,
        [string]$WorkingDirectory,
        [string]$FilePath,
        [string[]]$ArgumentList = @()
    )

    Write-Host "==> Starting $Name..."
    $startProcessArgs = @{
        FilePath = $FilePath
        WorkingDirectory = $WorkingDirectory
        PassThru = $true
    }
    if ($ArgumentList.Count -gt 0) {
        $startProcessArgs.ArgumentList = $ArgumentList
    }
    $proc = Start-Process @startProcessArgs
    $script:Processes += $proc
}

function Stop-ManagedProcesses {
    if ($script:Processes.Count -eq 0) {
        return
    }

    Write-Host ""
    Write-Host "Stopping all processes..."
    foreach ($proc in $script:Processes) {
        try {
            if (-not $proc.HasExited) {
                Stop-Process -Id $proc.Id -Force -ErrorAction Stop
            }
        } catch {
        }
    }
    Write-Host "All processes stopped."
}

function Parse-Targets {
    param([string[]]$InputArgs)

    $targets = @()
    for ($i = 0; $i -lt $InputArgs.Count; $i++) {
        $arg = $InputArgs[$i]
        switch ($arg) {
            "--t" {
                if ($i + 1 -ge $InputArgs.Count) {
                    throw "Error: --t requires a value (1, 2, or 3)"
                }
                $targets += $InputArgs[$i + 1]
                $i++
            }
            "-h" { return @{ ShowHelp = $true; Targets = @() } }
            "--help" { return @{ ShowHelp = $true; Targets = @() } }
            default {
                throw "Error: unknown option $arg"
            }
        }
    }

    return @{ ShowHelp = $false; Targets = $targets }
}

try {
    Import-DotEnv -FilePath $envFile

    $parsed = Parse-Targets -InputArgs $ArgsList
    if ($parsed.ShowHelp) {
        Show-Help
        exit 0
    }

    if ($parsed.Targets.Count -eq 0) {
        Show-Help
        exit 1
    }

    foreach ($target in $parsed.Targets) {
        switch ($target) {
            "1" {
                $adminDir = Join-Path $script:RootDir "frontend\admin"
                Ensure-NodeDeps -Dir $adminDir
                Start-ManagedProcess -Name "Admin Frontend (port 3001)" -WorkingDirectory $adminDir -FilePath "npx.cmd" -ArgumentList @("next", "dev", "-p", "3001")
            }
            "2" {
                $learnerDir = Join-Path $script:RootDir "frontend\learner"
                Ensure-NodeDeps -Dir $learnerDir
                Start-ManagedProcess -Name "Learner Frontend (port 3000)" -WorkingDirectory $learnerDir -FilePath "npx.cmd" -ArgumentList @("next", "dev", "-p", "3000")
            }
            "3" {
                $backendDir = Join-Path $script:RootDir "backend"
                $backendExe = Join-Path $backendDir "cmd\api\main.exe"
                Write-Host "Building backend..."
                Push-Location $backendDir
                try {
                    & go build -o $backendExe ./cmd/api/
                    if (-not $?) {
                        throw "go build failed"
                    }
                } finally {
                    Pop-Location
                }
                Start-ManagedProcess -Name "Backend API (port 8080)" -WorkingDirectory $backendDir -FilePath $backendExe -ArgumentList @()
            }
            default {
                throw "Error: unknown target '$target' (must be 1, 2, or 3)"
            }
        }
    }

    Write-Host ""
    Write-Host "All requested services are starting. Press Ctrl+C to stop all."
    Write-Host ""

    while ($true) {
        Start-Sleep -Seconds 1
        $running = @($script:Processes | Where-Object { -not $_.HasExited })
        if ($running.Count -ne $script:Processes.Count) {
            throw "One or more processes exited unexpectedly."
        }
    }
} finally {
    Stop-ManagedProcesses
}
