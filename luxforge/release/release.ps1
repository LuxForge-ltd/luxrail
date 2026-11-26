# release.ps1 -- build, sign and publish a LuxRail release. Run on joyeuse in PowerShell 7, from a
# clean LuxRail clone at the commit to release (normally the tip of Dev):
#   pwsh -File luxforge/release/release.ps1 -DryRun   # build + verify, no key, no publish
#   pwsh -File luxforge/release/release.ps1           # build, sign, verify, publish
# The release is tagged v<VERSION>, for example v1.0.0, from the VERSION file at the repository root.
# -Version may be given, but must equal that tag.
#
# Steps, all in containers (Docker Desktop):
#   1. Copy the clean checkout into a Docker volume and tag it (locally) with -Version.
#   2. Build the console: `make web` in node:24.
#   3. goreleaser (upstream's .goreleaser.yaml, publishing skipped): 5050d, 5050-agent and the CLI for
#      linux amd64/arm64 (CLI also darwin/windows), tarballs and checksums.txt, then sign checksums.txt.
#      A real release signs with the secret key from the vault (`luxrail/release-signing-key`), handed
#      to the container through an environment variable only; a dry run signs with upstream's test key.
#   4. Verify checksums.txt.minisig against the key embedded in internal/upgrade/release-keys.txt.
#   5. Push the tag and publish a GitHub Release on LuxForge-ltd/luxrail with the tarballs,
#      checksums.txt(.minisig), latest.txt, install.sh and install-cli.sh. Installs then use
#      --release-base https://github.com/LuxForge-ltd/luxrail/releases/download --version <tag>.
param(
    [string] $Version,
    [switch] $DryRun,
    [string] $Repo = 'LuxForge-ltd/luxrail',
    [string] $Vault = 'C:\Users\Ben\OneDrive\My Vault\LuxForge.kdbx',
    [string] $KeyFile = 'C:\Users\Ben\.keepass\luxForge-KeyFile.keyx',
    [string] $KeePassCli = 'C:\Program Files\KeePassXC\keepassxc-cli.exe',
    [string] $Entry = 'luxrail/release-signing-key',
    [string] $NodeImage = 'node:24-bookworm',
    [string] $GoReleaserImage = 'goreleaser/goreleaser:v2.12.7',
    [string] $OutDir = (Join-Path ([IO.Path]::GetTempPath()) 'luxrail-dist')
)
$ErrorActionPreference = 'Stop'
function Step($m) { Write-Host "== $m" }
function Info($m) { Write-Host "   $m" }
function Docker-Run { docker @args; if ($LASTEXITCODE -ne 0) { throw "docker $($args[0..3] -join ' ') ... failed ($LASTEXITCODE)" } }

$root = (git rev-parse --show-toplevel).Trim()
Set-Location $root
if (git status --porcelain) { throw 'the working tree is not clean; commit or remove the changes first' }
$sha = (git rev-parse HEAD).Trim()
$projectVersion = (Get-Content -LiteralPath VERSION -Raw).Trim()
if ($projectVersion -notmatch '^[0-9]+\.[0-9]+\.[0-9]+$') { throw "VERSION holds '$projectVersion', not MAJOR.MINOR.PATCH" }
if (-not $Version) { $Version = "v$projectVersion" }
if ($Version -cne "v$projectVersion") { throw "-Version $Version does not match VERSION ($projectVersion); releases are tagged v$projectVersion" }
$key = (Get-Content internal/upgrade/release-keys.txt | Where-Object { $_ -and -not $_.StartsWith('#') } | Select-Object -First 1)
if (-not $key) { throw 'no release key is embedded; run luxforge/release/new-release-key.ps1 first' }
if (-not $DryRun -and (git ls-remote --tags origin "refs/tags/$Version")) { throw "tag $Version already exists on origin" }
Step "LuxRail $Version at $($sha.Substring(0,7)) ($((git rev-parse --abbrev-ref HEAD).Trim()))"
Info "embedded release key: $key"

$secret = $null
if (-not $DryRun) {
    foreach ($p in $KeePassCli, $Vault, $KeyFile) { if (-not (Test-Path -LiteralPath $p)) { throw "not found: $p" } }
    $pw = [Net.NetworkCredential]::new('', (Read-Host "Vault password for $(Split-Path $Vault -Leaf)" -AsSecureString)).Password
    $tmpKey = Join-Path ([IO.Path]::GetTempPath()) ("lr-" + [Guid]::NewGuid().ToString('N'))
    try {
        $pw | & $KeePassCli attachment-export -q -k $KeyFile $Vault $Entry release.key $tmpKey | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "could not read $Entry from the vault" }
        $secret = (Get-Content -LiteralPath $tmpKey -Raw)
    } finally {
        $pw = $null
        if (Test-Path -LiteralPath $tmpKey) { [IO.File]::WriteAllBytes($tmpKey, (New-Object byte[] 512)); Remove-Item -LiteralPath $tmpKey -Force }
    }
    Info 'release key read from the vault'
}

$vol = 'luxrail-release-' + [Guid]::NewGuid().ToString('N').Substring(0, 8)
docker volume create $vol | Out-Null
try {
    Step 'Copy and tag'
    Docker-Run run --rm -v "${root}:/src:ro" -v "${vol}:/w" --entrypoint sh alpine/git:2.47.2 -c `
        "cp -a /src/. /w/ && cd /w && git config --global --add safe.directory /w && git -c user.name=LuxForge -c user.email=ben@luxforge.dev tag -f -a $Version -m 'LuxRail $Version' >/dev/null && git describe --tags"
    Step 'Console (make web)'
    Docker-Run run --rm -v "${vol}:/w" -w /w $NodeImage bash -c 'git config --global --add safe.directory /w && make web 2>&1 | tail -n 2'
    Step 'Build and sign (goreleaser)'
    $env:MINISIGN_KEY = $secret
    try {
        $cmd = 'git config --global --add safe.directory /w && echo LuxRail > /tmp/notes.md && goreleaser release --clean --skip=publish,announce,validate --release-notes=/tmp/notes.md 2>&1 | grep -E "⨯|error|release succeeded|signing  " ; test -f dist/checksums.txt.minisig'
        if ($DryRun) {
            Docker-Run run --rm -e GOTOOLCHAIN=auto -v "${vol}:/w" -w /w --entrypoint sh $GoReleaserImage -c $cmd
        } else {
            Docker-Run run --rm -e GOTOOLCHAIN=auto -e MINISIGN_KEY -v "${vol}:/w" -w /w --entrypoint sh $GoReleaserImage -c $cmd
        }
    } finally { Remove-Item Env:MINISIGN_KEY -ErrorAction SilentlyContinue; $secret = $null }

    Step 'Verify the signature with the embedded key'
    if ($DryRun) {
        Info 'dry run: signed with upstream test key, so it is checked against that key instead'
        Docker-Run run --rm -e GOTOOLCHAIN=auto -v "${vol}:/w" -w /w --entrypoint sh $GoReleaserImage -c `
            'go run ./e2e/signtool verify -p e2e/keys/test.pub -m dist/checksums.txt -x dist/checksums.txt.minisig'
    } else {
        Docker-Run run --rm -e GOTOOLCHAIN=auto -v "${vol}:/w" -w /w --entrypoint sh $GoReleaserImage -c `
            "go run ./e2e/signtool verify -P '$key' -m dist/checksums.txt -x dist/checksums.txt.minisig && sh e2e/scripts/check-release-keys.sh"
    }

    Step "Collect to $OutDir"
    if (Test-Path -LiteralPath $OutDir) { Remove-Item -LiteralPath $OutDir -Recurse -Force }
    New-Item -ItemType Directory -Path $OutDir | Out-Null
    Docker-Run run --rm -v "${vol}:/w" -v "${OutDir}:/out" alpine:3.20 sh -c `
        'cp /w/dist/*.tar.gz /w/dist/*.zip /w/dist/checksums.txt /w/dist/checksums.txt.minisig /w/.release/latest.txt /w/install/install.sh /w/install/install-cli.sh /out/'
    Get-ChildItem -LiteralPath $OutDir | ForEach-Object { Info ("{0,-48} {1,10:N0} bytes" -f $_.Name, $_.Length) }
} finally {
    docker volume rm -f $vol | Out-Null
}

if ($DryRun) { Step 'Dry run: nothing was tagged on origin or published'; return }

Step "Publish $Version on $Repo"
git tag -a $Version -m "LuxRail $Version" $sha
git push origin "refs/tags/$Version"
if ($LASTEXITCODE -ne 0) { throw 'pushing the tag failed' }
$notes = @"
LuxRail $Version, built from ``$($sha.Substring(0,7))``: openrail with LuxForge's changes, signed with the LuxRail release key.

Install or upgrade with:

``````
--release-base https://github.com/$Repo/releases/download --release-pubkey $key --version $Version
``````
"@
gh release create $Version --repo $Repo --verify-tag --title "LuxRail $Version" --notes $notes (Get-ChildItem -LiteralPath $OutDir | ForEach-Object FullName)
if ($LASTEXITCODE -ne 0) { throw 'creating the GitHub release failed (the tag is pushed; re-run gh release create by hand)' }
Step "Done: https://github.com/$Repo/releases/tag/$Version"
