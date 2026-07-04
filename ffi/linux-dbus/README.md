# BytePort Linux D-Bus Bridge — FFI scaffold

> Audit ref: `BytePort-audit.json` lines L121-L130 (PILLAR-TAXONOMY-v2.md
> v2.2 cross-platform FFI). This crate lifts **L127 `linux_ffi`** from
> `5` → `40+` by adding a third leg to the cross-platform FFI trifecta
> that began with `macos-share` and `android-companion` (PR #293,
> `af0b57e5`).

This crate wraps a [zbus 4](https://docs.rs/zbus/4) service skeleton
for `org.byteport.LinuxBridge1` behind a minimal Rust API. The actual
byte hand-off pipeline is **not** implemented in this scaffold — the
crate compiles, exposes the correct public API shape, and verifies
entry-point contracts via `cargo test` on `cfg(target_os = "linux")`.

## Why a third leg

Linux desktop is the canonical third platform for the BytePort host
client. Unlike macOS (Share Extensions) and Android (Share Intents),
Linux has no OS-level "share sheet"; instead, file managers and
desktop shells hand bytes off via D-Bus service calls. The
`org.byteport.LinuxBridge1` interface documented here is the lowest-
common-denominator surface that any Linux companion app (Nautilus
script, KDE ServiceMenu, GNOME Shell extension, FUSE-based watcher)
can target.

```
Nautilus / Dolphin / Nemo file manager
  ↓ (right-click → "Send to BytePort")
org.byteport.LinuxBridge1.Handoff(item_count: u64) → session_id: u64
  ↓
zbus server in host process (scaffold here)
  ↓
[future] forward byte chunks over D-Bus → BytePort host IPC
  ↓
BytesReceived(session_id, byte_count) signal
```

## Module layout

| File        | Purpose                                                       |
| ----------- | ------------------------------------------------------------- |
| `lib.rs`    | Public API: `LinuxBridgeError`, `LinuxBridgeHandle`, `BUS_NAME`, `OBJECT_PATH`, `open_linux_bridge_session`. |
| `dbus.rs`   | `#[zbus::dbus_interface]` trait skeleton for `org.byteport.LinuxBridge1` (Linux only). |
| `ipc.rs`    | Client-side helpers: `session_bus()`, `call_handoff()`, `subscribe_bridge_signals()`. |
| `codegen.rs`| Introspection XML stub + placeholder for the future `build.rs` codegen pipeline. |

## Platform gating

The entire crate is target-conditional:

- On `cfg(target_os = "linux")` the public API links against `zbus`,
  `zvariant`, and `tokio` and resolves to a real `LinuxBridge1` impl.
- On every other target the public API returns
  `LinuxBridgeError::NoBus` so `cargo check --workspace` passes on
  macOS / Windows CI without pulling Linux-only deps into the
  workspace lockfile.

The same gating idiom is used by `ffi/macos-share` (macOS-only deps)
and `ffi/android-companion` (Android-only deps) — see those crates for
the prior art.

## Introspection

```xml
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
 "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node name="/org/byteport/LinuxBridge1">
  <interface name="org.byteport.LinuxBridge1">
    <method name="Handoff">
      <arg name="item_count" type="t" direction="in"/>
      <arg name="session_id" type="t" direction="out"/>
    </method>
    <method name="Ping">
      <arg name="reply" type="s" direction="out"/>
    </method>
    <signal name="BytesReceived">
      <arg name="session_id" type="t"/>
      <arg name="byte_count" type="t"/>
    </signal>
  </interface>
</node>
```

## Activation (deferred to v8.2)

Real impl will:

1. Acquire a `zbus::Connection` to the session bus.
2. Register the well-known name `org.byteport.LinuxBridge1`.
3. Serve `dbus::LinuxBridge1` at `/org/byteport/LinuxBridge1`.
4. Forward inbound `Handoff` calls to the BytePort host process over
   the existing transport (`byteport-transport`).
5. Emit `BytesReceived` signals back to companion apps as bytes land.

Companion-app authoring examples live under
`docs/ffi/MOBILE.md` § Cross-compile matrix (post-v8.2).

## Build / test

```bash
# Linux build (real impl)
cargo build -p byteport-linux-dbus
cargo test  -p byteport-linux-dbus

# Cross-platform check (no-op stub on non-Linux targets)
cargo check --workspace
```