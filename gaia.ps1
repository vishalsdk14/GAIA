# GAIA Unified CLI for PowerShell
# This script routes commands to the GAIA Orchestration Kernel.

$KERNEL_URL = if ($env:GAIA_KERNEL_URL) { $env:GAIA_KERNEL_URL } else { "http://127.0.0.1:8080" }

function Show-Usage {
    Write-Host "GAIA CLI - The OS for AI Agents (PowerShell Version)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\gaia.ps1 [command] [args]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  submit <goal>    Submit a new high-level goal to the Kernel"
    Write-Host "  status <task_id> Get the current status and plan of a task"
    Write-Host "  list-agents      List all currently registered agents"
    Write-Host "  list-caps        List all available agent capabilities"
    Write-Host ""
    Write-Host "Environment Variables:"
    Write-Host "  GAIA_KERNEL_URL  Default: http://127.0.0.1:8080"
}

$command = $args[0]
$subArg = $args[1]

switch ($command) {
    "submit" {
        if (-not $subArg) {
            Write-Error "Error: Goal is required."
            exit 1
        }
        $payload = @{ goal = $subArg } | ConvertTo-Json
        try {
            $response = Invoke-RestMethod -Uri "$KERNEL_URL/api/v1/tasks" -Method Post -Body $payload -ContentType "application/json"
            $response | ConvertTo-Json -Depth 10
        } catch {
            Write-Error "Failed to submit goal: $_"
        }
    }
    "status" {
        if (-not $subArg) {
            Write-Error "Error: Task ID is required."
            exit 1
        }
        try {
            $response = Invoke-RestMethod -Uri "$KERNEL_URL/api/v1/tasks/$subArg" -Method Get
            $response | ConvertTo-Json -Depth 10
        } catch {
            Write-Error "Failed to get status: $_"
        }
    }
    "list-agents" {
        try {
            $response = Invoke-RestMethod -Uri "$KERNEL_URL/api/v1/registry/agents" -Method Get
            $response | ConvertTo-Json -Depth 10
        } catch {
            Write-Error "Failed to list agents: $_"
        }
    }
    "list-caps" {
        try {
            $response = Invoke-RestMethod -Uri "$KERNEL_URL/api/v1/registry/capabilities" -Method Get
            $response | ConvertTo-Json -Depth 10
        } catch {
            Write-Error "Failed to list capabilities: $_"
        }
    }
    Default {
        Show-Usage
    }
}
