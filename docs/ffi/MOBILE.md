# BytePort Mobile FFI — macOS Share Extension + Android Companion

> Audit ref: `BytePort-audit.json` lines L121-L130 (PILLAR-TAXONOMY-v2.md v2.1
> cross-platform FFI). Prior scores: `macos_ffi = 0/100`, `ios_ffi = 0/100`,
> `android_ffi = 0/100`, `ffi_toolchain = 0/100`, `cross_compile = 0/100`.
> This scaffold is the minimum viable bootstrap to begin raising those pillars.

This document accompanies the PR that adds two new workspace members:

```
ffi/
├── macos-share/         # swift-rs 0.1 scaffold
└── android-companion/   # ndk 0.8 scaffold
```

Both crates are **stubs only** — they compile, expose the correct public API
shape, and verify entry-point contracts via `cargo test`. The actual byte
hand-off to the BytePort host process is deferred to the v8.2 rollout.

## Why two scaffolds in one PR

Both crates share the same audit pillar cluster (mobile FFI) and the same
cross-compile infrastructure. Bundling them keeps the PR review surface tight
and avoids two identical "add a Cargo workspace member + scaffold" reviews.

## Design — macOS share extension

The Swift code lives in `ffi/macos-share/src/main.swift` and is compiled by
`swift-rs` (via `ffi/macos-share/build.rs`) into a static library that the
Rust `byteport-macos-share` crate links against.

```
User picks "Share to BytePort" in any macOS app
   ↓
NSItemProvider populates extensionContext.attachments
   ↓
BytePortShareExtension.bootstrapSession(itemCount:)
   ↓
Rust: open_share_session(item_count) → ShareHandle { session_id }
   ↓
[future] stream bytes into host process via XPC / NSFileCoordinator
```

The Rust shim is platform-gated: on `target_os = "macos"` it returns a real
`ShareHandle`; on any other OS it returns `ShareError::NotWired`. This lets
the workspace compile on Linux CI hosts without forcing macOS-only
toolchain requirements.

### Activation (deferred)

The `.appex` bundle is built by Xcode (or `xcodebuild`). The Info.plist
shape is documented in `ffi/macos-share/Info.plist`; the real bundle needs
a `MainInterface.storyboard` and an Xcode project. Tracking under issue
**[TBD — see BytePort-audit.json#L121-L130]**.

## Design — Android companion

The Android counterpart uses `ndk = "0.8"` and `jni = "0.21"` to expose a
JNI entry point that the Kotlin share-receiver calls.

```
User picks "Share to BytePort" in any Android app
   ↓
BytePortCompanionActivity receives Intent.ACTION_SEND
   ↓
Kotlin: BytePortCompanion.openSession(itemCount) (JNI)
   ↓
Rust: open_companion_session(item_count) → CompanionHandle { session_id }
   ↓
[future] ContentResolver.openInputStream(uri) → BytePort host via Unix socket
```

The `Java_io_byteport_android_BytePortCompanion_openSession` symbol is
exposed via `#[no_mangle] extern "system"` and currently lives in
`src/lib.rs::jni_bridge`. Platform-gated to `target_os = "android"`; on
other targets it returns `CompanionError::NoJvm`.

### Activation (deferred)

The `.apk` is built by Gradle using the Android NDK toolchain. The manifest
shape is in `ffi/android-companion/AndroidManifest.xml`; the real APK needs
a `BytePortCompanionActivity.kt` and a Gradle module. Tracking under the
same issue as macOS.

## Cross-compile matrix (documented, not CI-gated yet)

| Target                        | Crate                  | Required toolchain         |
| ----------------------------- | ---------------------- | -------------------------- |
| `aarch64-apple-darwin`        | `byteport-macos-share` | Xcode + Swift 5.9          |
| `x86_64-apple-darwin`         | `byteport-macos-share` | Xcode + Swift 5.9          |
| `aarch64-linux-android`       | `byteport-android-companion` | Android NDK r26 + API 34 |
| `x86_64-linux-android`        | `byteport-android-companion` | Android NDK r26 + API 34 |
| `linux-x86_64` (host CI)      | both (stub mode)       | stock Rust 1.82            |

CI gating is **deferred** — the acceptance criterion for this PR is only
that `cargo check --workspace --all-targets` passes on the host (linux-x86_64).
Once the cross-compile runners are provisioned (B6 of v8.1), the matrix above
becomes a CI matrix.

## Verification

```bash
# From repo root:
cargo check --workspace --all-targets
cargo test -p byteport-macos-share --lib
cargo test -p byteport-android-companion --lib
```

The macOS-target tests will be `cfg`-skipped on Linux CI (correct behaviour).
The Android-target tests will be `cfg`-skipped on Linux CI (correct behaviour).
The zero-item-count test runs on every platform and validates the entry-point
contract.

## Next steps (post-merge)

1. Wire the `.appex` build via `xcodebuild` (issue TBD).
2. Wire the `.apk` build via Gradle (issue TBD).
3. Provision cross-compile CI runners (B6 of v8.1).
4. Implement the actual byte hand-off in both `open_share_session` /
   `open_companion_session` (currently stubs).
5. Re-audit and confirm the FFI pillars moved from 0/100 → ≥40/100.

Refs: BytePort-audit.json#L121-L130, PILLAR-TAXONOMY-v2.md L121-L130.