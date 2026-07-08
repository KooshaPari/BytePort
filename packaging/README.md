# Linux packaging for BytePort

Native packaging recipes for the `byteport` CLI on every major Linux
distribution surface. The CLI binary is built from
`crates/byteport-cli/`; the desktop shell (Tauri) bundles it into
its own `.deb` / `.AppImage` / `.msi` / `.dmg` under
`frontend/web/src-tauri/`, so the recipes below cover only the CLI
surface.

> Mapping to PILLAR-TAXONOMY-v2.md v2.2:
> - **L41** (CLI UX) — shell completions, manpages, output format options
> - **L49** (Update Channels) — stable / beta / edge selection per channel
> - **L50** (Hardware/Edge) — `aarch64-linux-gnu` and `armv7-linux-gnueabihf` cross-builds
> - **L32** (Distribution) — 5+ distribution channels: snap, flatpak, AppImage, deb, rpm
> - **L33** (Install) — one-line-install recipes below

---

## At a glance

| Format      | Path                                       | Install one-liner                                              | When to use                                                |
| ----------- | ------------------------------------------ | -------------------------------------------------------------- | ---------------------------------------------------------- |
| **Snap**    | `snap/snapcraft.yaml`                      | `sudo snap install byteport`                                   | Snap Store + auto-updates (Ubuntu / derivatives)           |
| **Flatpak** | `flatpak/dev.kooshapari.byteport.yml`      | `flatpak install flathub dev.kooshapari.byteport`              | Sandboxed app from Flathub (any distro)                    |
| **AppImage** | `appimage/build.sh`                      | `chmod +x BytePort-*.AppImage && ./BytePort-*.AppImage`        | Portable single-file, no install required                  |
| **Debian**  | `debian/`                                  | `sudo dpkg -i byteport_0.1.0-1_amd64.deb`                      | Debian / Ubuntu / Pop!_OS / Mint / Kali                    |
| **RPM**     | `rpm/byteport.spec`                        | `sudo rpm -Uvh byteport-0.1.0-1.*.rpm`                         | Fedora / RHEL / Rocky / openSUSE                           |

For direct-binary installs that bypass package managers entirely,
see the upstream one-liner installer in the repo root
(`install.sh`):

```sh
curl -fsSL https://byteport.dev/install.sh | bash
```

Sibling directories in this repo (`brew/`, `aur/`, `scoop/`,
`winget/`, `chocolatey/`) handle macOS, Arch, and Windows
distribution surfaces — see [`packaging/README.md`](../README.md)
at the parent level for the cross-OS index. This file covers only
Linux-specific formats.

---

## Cross-build matrix (Pillar L50)

The recipes below all flow through `cargo build --target ...`. For
Pillar L50's Silver acceptance level (rpi-tested) the following
targets are exercised in CI:

| Target triple                       | Status               | Used by                  |
| ----------------------------------- | -------------------- | ------------------------ |
| `x86_64-unknown-linux-gnu`          | nightly + release    | snap, flatpak, deb, rpm  |
| `aarch64-unknown-linux-gnu`         | release              | snap, flatpak, deb, rpm  |
| `armv7-unknown-linux-gnueabihf`     | release (RPi 3/4)    | AppImage, deb            |
| `x86_64-unknown-linux-musl`         | musl static          | AppImage                 |
| `aarch64-unknown-linux-musl`        | musl static          | AppImage                 |

Snapcraft (`core22`) and Flatpak (`org.freedesktop.Platform 23.08`)
both target `x86_64` + `aarch64`; the AppImage build script detects
the host arch and selects the matching `linuxdeploy` AppImage.

---

## Update channels (Pillar L49)

Each distribution format maps to a release channel:

| Channel | Snap track           | Flatpak branch | Repo (deb / rpm)    |
| ------- | -------------------- | -------------- | ------------------- |
| stable  | `latest/stable`      | `stable`       | `release/v0.x`      |
| beta    | `latest/beta`        | `beta`         | `release-candidate` |
| nightly | `latest/edge` daily  | `nightly`      | `main` (post-merge) |

CI owns promotion: the upstream release pipeline (`.github/workflows/`)
is responsible for pushing each artifact to its respective channel.
Support matrix is tracked in `docs/operations/SUPPORT-MATRIX.md`.

---

## Channel comparison (when to use which)

### Snap — Store + auto-update
- **Pros:** sandboxed; automatic updates; store discovery; multi-arch
  handled transparently; integrates with Ubuntu's standard tooling.
- **Cons:** ~150 ms startup overhead from squashfs mount; needs store
  review to enable `latest/stable`. Strict confinement requires
  explicit interface declarations in `snapcraft.yaml` plugs.
- **Best for:** desktop Ubuntu users who want "just works" + auto-updates.

### Flatpak — Sandboxed, distro-agnostic
- **Pros:** runs on any distro; sandbox + portals give stronger
  isolation than deb / rpm; Flathub discovery.
- **Cons:** larger install size (downloads the runtime); not all
  distros ship `flatpak` by default.
- **Best for:** users on Arch / Fedora / RHEL who want the same
  app across multiple hosts.

### AppImage — Portable single-file
- **Pros:** no install required, runs from a USB stick, perfect for
  CI smoke-tests and air-gapped environments.
- **Cons:** no built-in auto-update (third-party `AppImageUpdate`
  optional); filesystem-based sandboxing is minimal.
- **Best for:** system administrators, ephemeral VMs, portable demos.

### Debian / Ubuntu (.deb)
- **Pros:** native integration with `apt` / `dpkg` / Launchpad PPA;
  minimal dependencies.
- **Cons:** tied to the Debian family; version drift if Ubuntu LTS
  is behind upstream glibc.
- **Best for:** Debian Stable / Ubuntu LTS production deployments.

### RPM (Fedora / RHEL / openSUSE)
- **Pros:** native for the Fedora / RHEL / SUSE family; COPR
  pre-release channel; integrates with `dnf` / `yum` / `zypper`.
- **Cons:** SELinux policy may require updates; vendor ABI drift.
- **Best for:** Fedora / RHEL / Rocky / CentOS Stream production.

---

## Per-recipe design notes

### `snap/snapcraft.yaml`
- `base: core22` (Ubuntu 22.04 LTS) today; a core24 migration TODO
  comment blocks the next bump until the `linux-dbus` FFI crate's
  glib requirement and rustc 1.78 MSRV are aligned.
- `adopt-info: byteport` so snapcraft reads version from the upstream
  `.cargo` config rather than a YAML literal.
- Plug list mirrors the Tauri shell's runtime needs (network,
  home, removable-media, wayland, x11, opengl, ssh-keys).

### `flatpak/dev.kooshapari.byteport.yml`
- `runtime: org.freedesktop.Platform` `23.08` — Fedora 39 / Debian 12
  baseline. finish-args map 1:1 onto the snap plugs above.
- Single `simple` buildsystem module wraps `cargo install --path`
  so we don't pull rustc into the runtime.
- Manpage + bash completion land under `/app/share/{man,bash-completion}/`.

### `appimage/`
- `build.sh` is an opinionated wrapper over `linuxdeploy`. The
  github release artifact (`linuxdeploy-x86_64.AppImage`) is pinned
  for reproducibility.
- `byteport.desktop` is a freedesktop Entry with the standard
  `Exec=`, `Icon=`, and `Categories=` fields.
- `byteport.svg` is a 512x512 brand icon (cube with portal glow)
  matching the upstream Tauri icon at
  `frontend/web/src-tauri/icons/`.

### `debian/`
- Built against `debhelper (>= 12)` and `Standards-Version: 4.6.2`.
- Binary package depends only on `libc6 (>= 2.31)` + `libssl3` +
  `ca-certificates` to keep PPA/Buster backports viable.
- `rules` is minimal: `dh` + cargo override; we override only
  `dh_auto_build` and `dh_auto_install`.

### `rpm/byteport.spec`
- Uses `%cargo_build` macro from `redhat-rpm-config`.
- `Requires:` declared explicitly (libc, libssl.so.3) so the
  resulting `.rpm` is self-describing for `dnf repoquery`.
- Changelog includes the maintainer tag with stable email + URL.

---

## Cross-references

- [`packaging/install.sh`](../install.sh) — direct-binary install (one-liner)
- [`frontend/web/src-tauri/tauri.conf.json`](../../frontend/web/src-tauri/tauri.conf.json) —
  desktop shell config (Tauri produces `msi`, `dmg`, `appimage`, `deb`)
- [`.goreleaser.yml`](../../.goreleaser.yml) — Go backend tarball output
- [`CHARTER.md`](../../CHARTER.md) § 6 (Governance Model)
- [`crates/byteport-cli/Cargo.toml`](../../crates/byteport-cli/Cargo.toml) —
  binary source (`name = "byteport"`, version `0.1.0`, Apache-2.0)

---

## License

Apache-2.0 — see [`LICENSE`](../../LICENSE).
