param(
    [Parameter(Mandatory = $true)]
    [string]$InputCluster,

    [string]$ConfigPath = ".\pipeline.config.json",

    [ValidateSet("auto", "ejs", "jsx", "tsx")]
    [string]$Target = "auto",

    [switch]$SkipSplit,
    [switch]$ForceReact,
    [switch]$ContinueOnError,
    [switch]$DryRun
)

$scriptPath = Join-Path $PSScriptRoot "scripts\run-refactor-pipeline.ps1"

& $scriptPath `
    -InputCluster $InputCluster `
    -ConfigPath $ConfigPath `
    -Target $Target `
    -SkipSplit:$SkipSplit `
    -ForceReact:$ForceReact `
    -ContinueOnError:$ContinueOnError `
    -DryRun:$DryRun

