// swift-rs build script — compiles src/ (a Swift Package Manager
// package) into a static library that the Rust FFI bridge links
// against.
//
// v8.2: re-enables the real swift-rs 0.1 pipeline. The previous scaffold
// release (v0.x) left this as a no-op because the API surface
// (`SwiftLinker::new(...).with_package(...).link()` chain shown in many
// blog posts) is from a NEWER swift-rs line. swift-rs 0.1 — the line
// available on crates.io when this scaffold was first cut — actually
// exposes:
//
//     swift_rs::build_utils::link_swift;
//     swift_rs::build_utils::link_swift_package("byteport-share", "src");
//
// where `link_swift_package(name, root)` invokes `swift build -c <profile>`
// in `root` (which must contain a `Package.swift`). The first call wires
// the Swift runtime libraries; the second compiles the package.
//
// To keep `cargo check --workspace` (run on every PR + CI) fast and
// free of a Swift toolchain round-trip, the real swift invocation is
// gated behind the `link-swift` *feature* AND `cfg(target_os = "macos")`.
// Default builds skip the Swift step:
//
//     cargo check --workspace                              ← fast, no swift
//     cargo build -p byteport-macos-share --features link-swift
//                                                        ← runs swift build
//
// Non-macOS targets never invoke swift.
//
// See docs/ffi/SWIFT-RS.md for the API mismatch write-up.

fn main() {
    // Re-emit the build script when the Swift sources + manifest move so
    // cargo rebuilds the static library.
    println!("cargo:rerun-if-changed=src/Package.swift");
    println!("cargo:rerun-if-changed=src/Sources/byteport-share/main.swift");
    println!("cargo:rerun-if-changed=src/lib.rs");
    println!("cargo:rerun-if-changed=Cargo.toml");

    // Only run `swift build` on macOS + when the `link-swift` feature
    // is enabled. `cargo check --workspace` (no features) skips this
    // block entirely so the PR / CI feedback loop stays fast.
    #[cfg(all(target_os = "macos", feature = "link-swift"))]
    {
        use swift_rs::build_utils::{link_swift, link_swift_package};

        // Wire the Swift runtime libraries (cargo:rustc-link-search=...).
        link_swift();

        // Compile the Swift package rooted at `src/` (which must contain
        // a `Package.swift` with target `byteport-share`). Emits a static
        // library named `libbyteport-share.a` for downstream `extern "C"`
        // linking.
        link_swift_package("byteport-share", "src");
    }

    // Default + non-macOS path: nothing to do. The crate compiles
    // because the FFI surface is gated by `#[cfg(target_os = "macos")]`
    // in `src/lib.rs`.
    #[cfg(not(all(target_os = "macos", feature = "link-swift")))]
    {
        // Explicit no-op. The scaffolding is intentionally Swift-less
        // for non-macOS CI + default `cargo check` runs.
    }
}
