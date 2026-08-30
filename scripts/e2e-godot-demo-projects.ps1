<#
.SYNOPSIS
Runs deadweight.gdt against every declared main scene in a Godot demo-project checkout.

.DESCRIPTION
This opt-in contributor check keeps the external demo corpus outside this repository.
It analyzes supported format-3 and format-4 text scenes, classifies UID-root inputs separately, and
returns a nonzero exit code only for setup failures or unexpected fatal analysis.

.PARAMETER DeadweightPath
Path to a built deadweight.gdt executable.

.PARAMETER DemoRoot
Path to a godot-demo-projects checkout.

.PARAMETER ExpectedCommit
Optional exact godot-demo-projects commit required before analysis starts.

.EXAMPLE
pwsh ./scripts/e2e-godot-demo-projects.ps1 `
  -DeadweightPath ./bin/deadweight.gdt.exe `
  -DemoRoot ../godot-demo-projects `
  -ExpectedCommit 0db80ca5fd22b9a40e05b9bc1e00af867fb7c712
#>

[CmdletBinding()]
param(
    [string] $DeadweightPath,
    [string] $DemoRoot,
    [string] $ExpectedCommit
)

$ErrorActionPreference = 'Stop'

function Stop-WithUsageError {
    param([string] $Message)

    [Console]::Error.WriteLine("ERROR: $Message")
    [Console]::Error.WriteLine('Usage: e2e-godot-demo-projects.ps1 -DeadweightPath PATH -DemoRoot PATH [-ExpectedCommit SHA]')
    exit 2
}

if ([string]::IsNullOrWhiteSpace($DeadweightPath)) {
    Stop-WithUsageError 'DeadweightPath is required.'
}
if ([string]::IsNullOrWhiteSpace($DemoRoot)) {
    Stop-WithUsageError 'DemoRoot is required.'
}

try {
    $resolvedBinary = (Resolve-Path -LiteralPath $DeadweightPath).Path
    $resolvedDemoRoot = (Resolve-Path -LiteralPath $DemoRoot).Path
}
catch {
    Stop-WithUsageError $_.Exception.Message
}

if (-not (Test-Path -LiteralPath $resolvedBinary -PathType Leaf)) {
    Stop-WithUsageError "DeadweightPath is not a file: $resolvedBinary"
}
if (-not (Test-Path -LiteralPath $resolvedDemoRoot -PathType Container)) {
    Stop-WithUsageError "DemoRoot is not a directory: $resolvedDemoRoot"
}

$corpusCommit = (& git -C $resolvedDemoRoot rev-parse HEAD 2>&1 | Out-String).Trim()
if ($LASTEXITCODE -ne 0) {
    Stop-WithUsageError "DemoRoot is not a readable Git checkout: $resolvedDemoRoot"
}
if (-not [string]::IsNullOrWhiteSpace($ExpectedCommit) -and $corpusCommit -ne $ExpectedCommit) {
    Stop-WithUsageError "DemoRoot commit is $corpusCommit; expected $ExpectedCommit."
}

$results = foreach ($projectFile in Get-ChildItem -LiteralPath $resolvedDemoRoot -Recurse -Filter project.godot -File | Sort-Object FullName) {
    $mainSceneMatch = Select-String -LiteralPath $projectFile.FullName -Pattern '^run/main_scene="([^"]+)"' | Select-Object -First 1
    if ($null -eq $mainSceneMatch) {
        continue
    }

    $mainScene = $mainSceneMatch.Matches[0].Groups[1].Value
    $project = $projectFile.Directory.FullName.Substring($resolvedDemoRoot.Length).TrimStart([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar)

    if ($mainScene.StartsWith('uid://', [StringComparison]::Ordinal)) {
        [PSCustomObject]@{
            Project = $project
            Scene = $mainScene
            Category = 'UNSUPPORTED_UID_ROOT'
            Detail = 'run/main_scene uses uid://'
        }
        continue
    }

    $commandOutput = (& $resolvedBinary --no-color --project $projectFile.Directory.FullName inspect $mainScene 2>&1 | Out-String).TrimEnd()
    $commandExit = $LASTEXITCODE

    $category = 'UNEXPECTED_FATAL'
    if ($commandExit -eq 0 -and $commandOutput -match '(?m)^Analysis:\s+(COMPLETE|PARTIAL)\s*$') {
        $category = $Matches[1]
    }
    $firstLine = ($commandOutput -split '\r?\n' | Select-Object -First 1)
    [PSCustomObject]@{
        Project = $project
        Scene = $mainScene
        Category = $category
        Detail = $firstLine
    }
}

$categoryOrder = @(
    'COMPLETE',
    'PARTIAL',
    'UNSUPPORTED_UID_ROOT',
    'UNEXPECTED_FATAL'
)

Write-Output "CORPUS_COMMIT $corpusCommit"
Write-Output "MAIN_SCENES $($results.Count)"
foreach ($category in $categoryOrder) {
    $count = @($results | Where-Object Category -EQ $category).Count
    Write-Output "$category $count"
}

$unexpected = @($results | Where-Object Category -EQ 'UNEXPECTED_FATAL')
if ($unexpected.Count -gt 0) {
    Write-Output 'UNEXPECTED_FATAL_DETAILS'
    foreach ($result in $unexpected | Sort-Object Project, Scene) {
        Write-Output "$($result.Project) | $($result.Scene) | $($result.Detail)"
    }
    exit 1
}

exit 0
