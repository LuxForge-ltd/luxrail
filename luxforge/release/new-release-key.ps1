# new-release-key.ps1 -- create LuxRail's release signing key, once. Run on joyeuse in PowerShell 7,
# from the LuxRail clone, on the branch that will carry the new public key:
#   pwsh -File luxforge/release/new-release-key.ps1
#
# 1. Runs upstream's e2e/scripts/release-key.sh in a golang container. It generates a minisign key
#    pair and writes the public key into internal/upgrade/release-keys.txt, install/install.sh and
#    install/install-cli.sh (replacing upstream's key), so LuxRail builds trust only LuxRail releases.
# 2. Stores the secret key in the KeePassXC vault as an attachment of `luxrail/release-signing-key`
#    (the public key goes in the entry's notes), after asking once for the vault password (hidden).
# 3. Checks the stored copy, then overwrites and deletes the secret key file.
# The secret key is never printed. Commit the three changed files with the phase that adds the key.
param(
    [string] $Vault = 'C:\Users\Ben\OneDrive\My Vault\LuxForge.kdbx',
    [string] $KeyFile = 'C:\Users\Ben\.keepass\luxForge-KeyFile.keyx',
    [string] $KeePassCli = 'C:\Program Files\KeePassXC\keepassxc-cli.exe',
    [string] $Entry = 'luxrail/release-signing-key',
    [string] $GoImage = 'golang:1.26-bookworm'
)
$ErrorActionPreference = 'Stop'
$repo = (git rev-parse --show-toplevel).Trim()
if (-not (Test-Path -LiteralPath (Join-Path $repo 'e2e/scripts/release-key.sh'))) { throw "run this from the LuxRail clone" }
foreach ($p in $KeePassCli, $Vault, $KeyFile) { if (-not (Test-Path -LiteralPath $p)) { throw "not found: $p" } }
$pw = [Net.NetworkCredential]::new('', (Read-Host "Vault password for $(Split-Path $Vault -Leaf)" -AsSecureString)).Password
$null = $pw | & $KeePassCli ls -q -k $KeyFile $Vault 'luxrail' 2>$null
if ($LASTEXITCODE -ne 0) { throw "the vault did not open, or it has no 'luxrail' group (create it first)" }
$existing = $pw | & $KeePassCli show -q -k $KeyFile $Vault $Entry 2>$null
if ($LASTEXITCODE -eq 0 -and $existing) { throw "$Entry already exists: a release key was made before. Remove that entry first if you really mean to replace the key." }

$out = Join-Path ([IO.Path]::GetTempPath()) ("luxrail-key-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $out | Out-Null
try {
    Write-Host "== Generating the key pair and embedding the public key (in a container)"
    docker run --rm -v "${repo}:/src" -v "${out}:/out" -w /src $GoImage sh -c 'git config --global --add safe.directory /src && sh e2e/scripts/release-key.sh /out >/dev/null && sed -n 2p /out/release.pub'
    if ($LASTEXITCODE -ne 0) { throw 'release-key.sh failed' }
    $pub = (Get-Content -LiteralPath (Join-Path $out 'release.pub'))[1].Trim()
    $sec = Join-Path $out 'release.key'
    Write-Host "   public key: $pub"

    Write-Host "== Storing the secret key in the vault as $Entry"
    $pw | & $KeePassCli add -q -k $KeyFile --notes "LuxRail minisign release signing key. Public key: $pub" $Vault $Entry | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'could not create the vault entry' }
    $pw | & $KeePassCli attachment-import -q -k $KeyFile $Vault $Entry release.key $sec | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'could not attach the secret key' }
    $check = Join-Path $out 'readback.key'
    $pw | & $KeePassCli attachment-export -q -k $KeyFile $Vault $Entry release.key $check | Out-Null
    if ($LASTEXITCODE -ne 0 -or (Get-FileHash $check).Hash -ne (Get-FileHash $sec).Hash) {
        throw 'the stored key does not match the generated one; nothing was deleted from the vault'
    }
    Write-Host '   stored and read back: matches'
} finally {
    $pw = $null
    Get-ChildItem -LiteralPath $out -File -ErrorAction SilentlyContinue | ForEach-Object {
        $len = [Math]::Max(1, $_.Length); [IO.File]::WriteAllBytes($_.FullName, (New-Object byte[] $len))
    }
    Remove-Item -LiteralPath $out -Recurse -Force -ErrorAction SilentlyContinue
}
Write-Host '== Done. Changed files (commit them with the phase):'
git -C $repo status --short -- internal/upgrade/release-keys.txt install/install.sh install/install-cli.sh
