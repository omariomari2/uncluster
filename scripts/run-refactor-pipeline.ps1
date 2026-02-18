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

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-ConfigValue {
    param(
        [Parameter(Mandatory = $true)][object]$Object,
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter()][object]$Default = $null,
        [switch]$Required
    )

    $prop = $Object.PSObject.Properties[$Name]
    if ($null -ne $prop -and $null -ne $prop.Value -and "$($prop.Value)" -ne "") {
        return $prop.Value
    }

    if ($Required) {
        throw "Missing required config key: $Name"
    }

    return $Default
}

function Resolve-PathFromBase {
    param(
        [Parameter(Mandatory = $true)][string]$PathValue,
        [Parameter(Mandatory = $true)][string]$BasePath
    )

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }

    return [System.IO.Path]::GetFullPath((Join-Path $BasePath $PathValue))
}

function Expand-Template {
    param(
        [Parameter(Mandatory = $true)][string]$Template,
        [Parameter(Mandatory = $true)][hashtable]$Values
    )

    $result = $Template
    foreach ($key in $Values.Keys) {
        $token = "{${key}}"
        $replacement = [string]$Values[$key]
        $result = $result.Replace($token, $replacement)
    }
    return $result
}

function Invoke-ExternalCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Label,
        [Parameter(Mandatory = $true)][string]$CommandText,
        [switch]$DryRun,
        [switch]$AllowFailure
    )

    Write-Host "[$Label] $CommandText"

    if ($DryRun) {
        return 0
    }

    Invoke-Expression $CommandText
    $succeeded = $?
    $exitCode = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { if ($succeeded) { 0 } else { 1 } }

    if ($exitCode -ne 0 -and -not $AllowFailure) {
        throw "$Label failed with exit code $exitCode"
    }

    return $exitCode
}

function Get-RelativePath {
    param(
        [Parameter(Mandatory = $true)][string]$BasePath,
        [Parameter(Mandatory = $true)][string]$TargetPath
    )

    $baseFull = [System.IO.Path]::GetFullPath($BasePath).TrimEnd('\') + '\'
    $targetFull = [System.IO.Path]::GetFullPath($TargetPath)

    $baseUri = New-Object System.Uri($baseFull)
    $targetUri = New-Object System.Uri($targetFull)
    $relative = [System.Uri]::UnescapeDataString($baseUri.MakeRelativeUri($targetUri).ToString())
    return $relative.Replace('/', '\')
}

function Get-LocalDependencyPaths {
    param(
        [Parameter(Mandatory = $true)][string]$PagePath,
        [Parameter(Mandatory = $true)][string]$SplitRoot
    )

    $dependencies = New-Object System.Collections.Generic.HashSet[string]
    $pageDir = Split-Path -Parent $PagePath
    $content = Get-Content -LiteralPath $PagePath -Raw
    $matches = [System.Text.RegularExpressions.Regex]::Matches($content, '(?i)(?:src|href)\s*=\s*["'']([^"''#?]+)')

    foreach ($match in $matches) {
        $ref = $match.Groups[1].Value.Trim()
        if ($ref -eq "") { continue }
        if ($ref.StartsWith("http://") -or $ref.StartsWith("https://") -or $ref.StartsWith("//") -or $ref.StartsWith("data:") -or $ref.StartsWith("mailto:")) { continue }

        $candidate = if ([System.IO.Path]::IsPathRooted($ref)) {
            Join-Path $SplitRoot $ref.TrimStart('\', '/')
        } else {
            Join-Path $pageDir $ref
        }

        if (Test-Path -LiteralPath $candidate -PathType Leaf) {
            $full = [System.IO.Path]::GetFullPath($candidate)
            [void]$dependencies.Add($full)
        }
    }

    return @($dependencies)
}

function Get-RiskHits {
    param(
        [Parameter(Mandatory = $true)][string]$PagePath,
        [Parameter(Mandatory = $true)][string]$SplitRoot,
        [Parameter(Mandatory = $true)][string[]]$Patterns
    )

    $filesToScan = New-Object System.Collections.Generic.List[string]
    $filesToScan.Add([System.IO.Path]::GetFullPath($PagePath))
    foreach ($dep in (Get-LocalDependencyPaths -PagePath $PagePath -SplitRoot $SplitRoot)) {
        $filesToScan.Add($dep)
    }

    $hitSet = New-Object System.Collections.Generic.HashSet[string]
    foreach ($filePath in $filesToScan) {
        $text = Get-Content -LiteralPath $filePath -Raw
        foreach ($pattern in $Patterns) {
            if ([System.Text.RegularExpressions.Regex]::IsMatch($text, $pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)) {
                [void]$hitSet.Add($pattern)
            }
        }
    }

    return @($hitSet)
}

$workspace = (Get-Location).Path
$inputPath = Resolve-PathFromBase -PathValue $InputCluster -BasePath $workspace
$configFullPath = Resolve-PathFromBase -PathValue $ConfigPath -BasePath $workspace

if (-not (Test-Path -LiteralPath $configFullPath -PathType Leaf)) {
    throw "Config file not found: $configFullPath"
}

$config = Get-Content -LiteralPath $configFullPath -Raw | ConvertFrom-Json
$pathsConfig = Get-ConfigValue -Object $config -Name "paths" -Required
$promptsConfig = Get-ConfigValue -Object $config -Name "prompts" -Required

$runRoot = Resolve-PathFromBase -PathValue (Get-ConfigValue -Object $pathsConfig -Name "runRoot" -Default ".runs") -BasePath $workspace
$splitFolderName = Get-ConfigValue -Object $pathsConfig -Name "splitOutput" -Default "split"
$ejsFolderName = Get-ConfigValue -Object $pathsConfig -Name "ejsOutput" -Default "ejs"
$reactFolderName = Get-ConfigValue -Object $pathsConfig -Name "reactOutput" -Default "react"

$ejsPromptPath = Resolve-PathFromBase -PathValue (Get-ConfigValue -Object $promptsConfig -Name "ejs" -Required) -BasePath $workspace
$reactPromptPath = Resolve-PathFromBase -PathValue (Get-ConfigValue -Object $promptsConfig -Name "react" -Required) -BasePath $workspace

if (-not (Test-Path -LiteralPath $ejsPromptPath -PathType Leaf)) {
    throw "EJS prompt file not found: $ejsPromptPath"
}
if (-not (Test-Path -LiteralPath $reactPromptPath -PathType Leaf)) {
    throw "React prompt file not found: $reactPromptPath"
}

$splitterCommandTemplate = Get-ConfigValue -Object $config -Name "splitterCommand" -Required
$ejsCommandTemplate = Get-ConfigValue -Object $config -Name "ejsCommand" -Required
$reactCommandTemplate = Get-ConfigValue -Object $config -Name "reactCommand" -Required

$defaultReactTarget = Get-ConfigValue -Object $config -Name "defaultReactTarget" -Default "tsx"
$riskPatterns = Get-ConfigValue -Object $config -Name "riskPatterns" -Default @(
    "webflow",
    "__wf",
    "framer",
    "gsap",
    "barba",
    "locomotive",
    "lenis",
    "swiper",
    "anime(\.min)?\.js",
    "three(\.min)?\.js",
    "jquery\.[a-z0-9\-_]+(\.min)?\.js"
)

if (-not (Test-Path -LiteralPath $runRoot)) {
    New-Item -ItemType Directory -Path $runRoot -Force | Out-Null
}

$runId = Get-Date -Format "yyyyMMdd-HHmmss"
$runDir = Join-Path $runRoot $runId
$splitRoot = Join-Path $runDir $splitFolderName
$ejsRoot = Join-Path $runDir $ejsFolderName
$reactRoot = Join-Path $runDir $reactFolderName

New-Item -ItemType Directory -Path $runDir -Force | Out-Null
New-Item -ItemType Directory -Path $ejsRoot -Force | Out-Null
New-Item -ItemType Directory -Path $reactRoot -Force | Out-Null

if (-not $SkipSplit) {
    New-Item -ItemType Directory -Path $splitRoot -Force | Out-Null

    $splitCommand = Expand-Template -Template $splitterCommandTemplate -Values @{
        input       = $inputPath
        split_output = $splitRoot
        run_dir     = $runDir
    }

    [void](Invoke-ExternalCommand -Label "splitter" -CommandText $splitCommand -DryRun:$DryRun -AllowFailure:$ContinueOnError)
} else {
    $splitRoot = $inputPath
}

if ($DryRun -and -not $SkipSplit -and (Test-Path -LiteralPath $inputPath -PathType Container)) {
    Write-Host "[dry-run] Using input directory as split source preview: $inputPath"
    $splitRoot = $inputPath
}

if (-not (Test-Path -LiteralPath $splitRoot -PathType Container)) {
    throw "Split output directory not found: $splitRoot"
}

$pages = @(Get-ChildItem -Path $splitRoot -Recurse -File | Where-Object { $_.Extension -in @(".html", ".htm") } | Sort-Object FullName)
if ($pages.Count -eq 0) {
    throw "No HTML pages found in split directory: $splitRoot"
}

$effectiveTarget = if ($Target -eq "auto") { $defaultReactTarget } else { $Target }
$manifestRows = New-Object System.Collections.Generic.List[object]
$ejsCount = 0
$reactCount = 0
$fallbackCount = 0
$failedCount = 0

foreach ($page in $pages) {
    $relativeHtml = Get-RelativePath -BasePath $splitRoot -TargetPath $page.FullName
    $relativeStem = [System.IO.Path]::ChangeExtension($relativeHtml, $null).TrimEnd(".")
    $relativeDir = Split-Path -Parent $relativeHtml
    $fileBase = [System.IO.Path]::GetFileNameWithoutExtension($page.Name)

    $ejsRelative = if ($relativeDir) { Join-Path $relativeDir "$fileBase.ejs" } else { "$fileBase.ejs" }
    $ejsPath = Join-Path $ejsRoot $ejsRelative
    $ejsDir = Split-Path -Parent $ejsPath
    New-Item -ItemType Directory -Path $ejsDir -Force | Out-Null

    $ejsCommand = Expand-Template -Template $ejsCommandTemplate -Values @{
        input        = $page.FullName
        output       = $ejsPath
        prompt       = $ejsPromptPath
        page_name    = $fileBase
        page_relative = $relativeHtml
        split_root   = $splitRoot
        run_dir      = $runDir
    }

    try {
        [void](Invoke-ExternalCommand -Label "ejs:$relativeHtml" -CommandText $ejsCommand -DryRun:$DryRun -AllowFailure:$ContinueOnError)
        $ejsCount++
    } catch {
        $failedCount++
        $manifestRows.Add([PSCustomObject]@{
            page        = $relativeHtml
            status      = "failed-ejs"
            risk_hits   = @()
            ejs_output  = $ejsPath
            react_output = $null
            note        = $_.Exception.Message
        })
        if (-not $ContinueOnError) { throw }
        continue
    }

    if ($Target -eq "ejs") {
        $manifestRows.Add([PSCustomObject]@{
            page         = $relativeHtml
            status       = "ejs-only"
            risk_hits    = @()
            ejs_output   = $ejsPath
            react_output = $null
            note         = "React conversion skipped because target=ejs"
        })
        continue
    }

    $riskHits = @(Get-RiskHits -PagePath $page.FullName -SplitRoot $splitRoot -Patterns $riskPatterns)
    $shouldSkipReact = ($Target -eq "auto" -and $riskHits.Count -gt 0 -and -not $ForceReact)

    if ($shouldSkipReact) {
        $fallbackCount++
        $manifestRows.Add([PSCustomObject]@{
            page         = $relativeHtml
            status       = "fallback-ejs"
            risk_hits    = $riskHits
            ejs_output   = $ejsPath
            react_output = $null
            note         = "React conversion skipped due to risky dependencies"
        })
        Write-Host "[react:$relativeHtml] skipped (risk patterns: $($riskHits -join ', '))"
        continue
    }

    $reactExt = if ($effectiveTarget -eq "jsx") { ".jsx" } else { ".tsx" }
    $reactRelative = if ($relativeDir) { Join-Path $relativeDir "$fileBase$reactExt" } else { "$fileBase$reactExt" }
    $reactPath = Join-Path $reactRoot $reactRelative
    $reactDir = Split-Path -Parent $reactPath
    New-Item -ItemType Directory -Path $reactDir -Force | Out-Null

    $reactCommand = Expand-Template -Template $reactCommandTemplate -Values @{
        input         = $ejsPath
        output        = $reactPath
        prompt        = $reactPromptPath
        target        = $effectiveTarget
        page_name     = $fileBase
        page_relative = $relativeHtml
        run_dir       = $runDir
    }

    try {
        [void](Invoke-ExternalCommand -Label "react:$relativeHtml" -CommandText $reactCommand -DryRun:$DryRun -AllowFailure:$ContinueOnError)
        $reactCount++
        $manifestRows.Add([PSCustomObject]@{
            page         = $relativeHtml
            status       = "react"
            risk_hits    = $riskHits
            ejs_output   = $ejsPath
            react_output = $reactPath
            note         = "Converted to $effectiveTarget"
        })
    } catch {
        $failedCount++
        $manifestRows.Add([PSCustomObject]@{
            page         = $relativeHtml
            status       = "failed-react"
            risk_hits    = $riskHits
            ejs_output   = $ejsPath
            react_output = $reactPath
            note         = $_.Exception.Message
        })
        if (-not $ContinueOnError) { throw }
    }
}

$manifest = [PSCustomObject]@{
    run_id = $runId
    created_at = (Get-Date).ToString("o")
    input = $inputPath
    target = $Target
    effective_target = $effectiveTarget
    split_root = $splitRoot
    ejs_root = $ejsRoot
    react_root = $reactRoot
    counts = [PSCustomObject]@{
        pages = $pages.Count
        ejs_converted = $ejsCount
        react_converted = $reactCount
        fallback_ejs = $fallbackCount
        failed = $failedCount
    }
    pages = $manifestRows
}

$manifestPath = Join-Path $runDir "manifest.json"
$manifest | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $manifestPath -Encoding UTF8

Write-Host ""
Write-Host "Run completed:"
Write-Host "  run directory  : $runDir"
Write-Host "  manifest       : $manifestPath"
Write-Host "  html pages     : $($pages.Count)"
Write-Host "  ejs converted  : $ejsCount"
Write-Host "  react converted: $reactCount"
Write-Host "  ejs fallbacks  : $fallbackCount"
Write-Host "  failures       : $failedCount"
