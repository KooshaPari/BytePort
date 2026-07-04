// swift-tools-version:5.5
// swift-rs 0.1 expects a SwiftPM layout. The v8.2 scaffold pins
// `link_swift_package("byteport-share", "src")` from build.rs, which
// invokes `swift build -c <profile>` rooted here.
//
// IMPORTANT: swift-rs 0.1 hardcodes `-apple-macosx11` as the build
// target (see `swift_rs::build_utils::MACOS_TARGET_VERSION = "11"`).
// The Package.swift `.macOS(...)` declaration must match — `.v10_15`
// triggers a `Libraries require RPath!` panic inside swift-rs 0.1.
// Bumping to a later swift-rs line relaxes this constraint.

import PackageDescription

let package = Package(
    name: "byteport-share",
    platforms: [
        .macOS(.v11),
    ],
    products: [
        .library(
            name: "byteport-share",
            targets: ["byteport-share"]
        ),
    ],
    targets: [
        .target(
            name: "byteport-share",
            path: "Sources/byteport-share"
        ),
    ]
)
