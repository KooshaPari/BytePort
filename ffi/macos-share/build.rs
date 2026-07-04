// swift-rs build script — compiles src/main.swift as a static library
// and surfaces the generated C header for the Rust FFI bridge.
//
// SCAFFOLD MODE: The Swift compilation step is intentionally skipped in
// this scaffold release. Real swift-rs invocation is deferred to v8.2
// once we resolve the swift-rs API surface (the SwiftLinker symbol
// moved in a recent swift-rs release and is being tracked upstream).
//
// On a real macOS host the Swift code in src/main.swift is compiled by
// the proper swift-rs pipeline via the parent crate's xtask:
//
//     cargo run -p byteport-xtask -- share-ext build --target aarch64-apple-darwin
//
// For this scaffold, this build script is a no-op on every target so
// `cargo check --workspace` passes cleanly. The cdylib + staticlib
// crate-types declared in Cargo.toml still resolve; they are linked by
// the host process at runtime once the share-extension is wired up.
//
// See docs/ffi/MOBILE.md § Cross-compile matrix for the full plan.

fn main() {
    // Scaffold-only: intentionally empty.
    //
    // When the swift-rs API surface is re-introduced, replace this with:
    //
    //   swift_rs::SwiftLinker::new("10.15")
    //       .with_package("byteport-share", "src/main.swift")
    //       .link();
}
