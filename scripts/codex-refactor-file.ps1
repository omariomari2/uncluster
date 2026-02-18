param(
    [Parameter(Mandatory = $true)]
    [string]$InputPath,

    [Parameter(Mandatory = $true)]
    [string]$OutputPath,

    [Parameter(Mandatory = $true)]
    [string]$PromptPath,

    [Parameter(Mandatory = $true)]
    [ValidateSet("ejs", "jsx", "tsx")]
    [string]$Target
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $InputPath -PathType Leaf)) {
    throw "Input file not found: $InputPath"
}

if (-not (Test-Path -LiteralPath $PromptPath -PathType Leaf)) {
    throw "Prompt file not found: $PromptPath"
}

$outputDir = Split-Path -Parent $OutputPath
if ($outputDir -and -not (Test-Path -LiteralPath $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
}

$promptBody = Get-Content -LiteralPath $PromptPath -Raw
$request = @"
$promptBody

Execution context:
- Input file path: $InputPath
- Output target type: $Target

Required behavior:
1. Read the file at the input path.
2. Transform it according to the prompt.
3. Return only the final file contents for the output file.
4. Do not include markdown fences, explanations, or any extra text.
5. Do not return a wrapper that just includes/references the source HTML file path.
6. Output the full transformed file content directly.
7. Do not compress or encode the markup into base64/gzip wrappers.
8. Keep output human-readable and editable.
"@

$request | codex exec - --dangerously-bypass-approvals-and-sandbox -o $OutputPath | Out-Null

if (-not (Test-Path -LiteralPath $OutputPath -PathType Leaf)) {
    throw "Codex did not produce output file: $OutputPath"
}

$outputText = Get-Content -LiteralPath $OutputPath -Raw
$inputText = Get-Content -LiteralPath $InputPath -Raw
$inputPathUnix = $InputPath.Replace('\', '/')

if ($outputText -match [Regex]::Escape($InputPath) -or $outputText -match [Regex]::Escape($inputPathUnix)) {
    throw "Rejected output: contains direct reference to source input path"
}

if ($outputText -match '(?is)<%-?\s*include\((["'']).*?\.html\1\)\s*%>') {
    throw "Rejected output: includes source .html via EJS include wrapper"
}

$suspiciousOutputPatterns = @(
    '(?is)\bzlib\.(gunzipSync|gzipSync|inflateSync|deflateSync)\b',
    '(?is)require\((["''])zlib\1\)',
    '(?is)\bBuffer\.from\((["''])[A-Za-z0-9+/=\r\n]{1024,}\1\s*,\s*(["''])base64\2\)',
    '(?is)\batob\((["''])[A-Za-z0-9+/=\r\n]{1024,}\1\)',
    '(?is)\bnew\s+TextDecoder\(\)\.decode\('
)

foreach ($pattern in $suspiciousOutputPatterns) {
    $outputHasPattern = [Regex]::IsMatch($outputText, $pattern)
    $inputHasPattern = [Regex]::IsMatch($inputText, $pattern)
    if ($outputHasPattern -and -not $inputHasPattern) {
        throw "Rejected output: detected encoded/compressed wrapper pattern ($pattern)"
    }
}

if ($Target -eq "ejs" -and $outputText -match '(?is)^<%[\s\S]{0,8000}(zlib|Buffer\.from|base64|atob)[\s\S]*%>\s*<%-?\s*[A-Za-z_$][A-Za-z0-9_$]*\s*%>\s*$') {
    throw "Rejected output: generated EJS wrapper template around encoded payload"
}

$minExpectedLength = [Math]::Max(32, [int]([Math]::Round($inputText.Length * 0.10)))
if ($outputText.Length -lt $minExpectedLength) {
    throw "Rejected output: generated file too small relative to input (input=$($inputText.Length), output=$($outputText.Length))"
}
