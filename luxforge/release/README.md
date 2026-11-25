# LuxRail releases

LuxRail is released the way upstream openrail is, with LuxRail's own signing key: goreleaser builds
`5050d`, `5050-agent` and the `openrail` CLI, `checksums.txt` lists every file and is signed with
minisign, and every LuxRail binary trusts only the public key embedded in
`internal/upgrade/release-keys.txt` (also in `install/install.sh` and `install/install-cli.sh`).
Releases are published as GitHub Releases on `LuxForge-ltd/luxrail`. No GitHub Actions are used:
everything runs in Docker on joyeuse.

## Once: the release key

```
pwsh -File luxforge/release/new-release-key.ps1
```

Run it from the LuxRail clone, on the branch that will carry the key. It asks once for the vault
password, generates the key pair with upstream's `e2e/scripts/release-key.sh` in a container,
writes the public key into the three files above, stores the secret key as the attachment
`release.key` of the vault entry `luxrail/release-signing-key` (with the public key in its notes),
reads it back to check it, and wipes the key file. It refuses to run if that entry already exists.
Commit the three changed files.

## Each release

From a clean clone at the commit to release (normally the tip of `Dev`):

```
pwsh -File luxforge/release/release.ps1 -Version v0.1.9-lf.1 -DryRun
pwsh -File luxforge/release/release.ps1 -Version v0.1.9-lf.1
```

1. The checkout is copied into a Docker volume and tagged with the version.
2. `make web` builds the console (node:24).
3. goreleaser (v2.12.7, upstream's `.goreleaser.yaml`, publishing skipped, Go 1.26 through
   `GOTOOLCHAIN=auto`) builds the tarballs and `checksums.txt` and signs it. A real release reads the
   secret key from the vault (hidden password prompt) and hands it to the container through an
   environment variable only; a dry run signs with upstream's test key.
4. The signature is verified against the embedded key, and `check-release-keys.sh` confirms the
   three files agree.
5. The tag is pushed and a GitHub Release is created with the tarballs, `checksums.txt`,
   `checksums.txt.minisig`, `latest.txt`, `install.sh` and `install-cli.sh`.

A build takes about six minutes on joyeuse.

## Versions

Release tags are `v<upstream version>-lf.<n>`, for example `v0.1.9-lf.1`: the upstream release
LuxRail is based on, then LuxRail's own count. The upgrader accepts any tag of that form when it is
named with `--version`; its "update available" notice compares only the `X.Y.Z` part, so pass
`--version` explicitly. LuxRail's project version in `VERSION` is separate (one step per phase).

## Installing and upgrading

New install:

```
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/download/<tag>/install.sh | sudo sh -s -- \
  --release-base https://github.com/LuxForge-ltd/luxrail/releases/download \
  --release-pubkey <public key> --version <tag>
```

Upgrade an existing openrail or LuxRail install in place (a state backup and a VM snapshot first).
`5050d upgrade` reads where releases come from and which keys it trusts from
`/etc/5050/release.json`, which the installer writes when it is given `--release-base`. An install
made from upstream has no such file, so write it once:

```
{
  "releaseBase": "https://github.com/LuxForge-ltd/luxrail/releases/download",
  "publicKeys": ["<public key>"]
}
```

then:

```
sudo 5050d upgrade --version <tag>
```

From then on the install runs LuxRail binaries, which trust only the LuxRail key.
