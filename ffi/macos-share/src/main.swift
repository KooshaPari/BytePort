// BytePort macOS Share Extension — Swift stub.
//
// This file is compiled by swift-rs (see ../build.rs) and produces a
// static library + C header that the Rust side links against.
//
// SCAFFOLD ONLY. The full ShareViewController lifecycle (NSItemProvider
// enumeration, BytePort IPC hand-off) is not implemented. See
// docs/ffi/MOBILE.md for the design.

import Foundation

// Mirror of Rust `ShareHandle` so Swift callers can construct one
// without round-tripping through Rust for trivial cases.
@objc public final class BytePortShareHandle: NSObject {
    @objc public let sessionId: UInt64

    @objc public init(sessionId: UInt64) {
        self.sessionId = sessionId
        super.init()
    }
}

// Stub of the future ShareViewController class. In production this
// would subclass SLComposeServiceViewController (or
// UIViewController under the NSExtensionPrincipalClass key) and
// enumerate attachments via NSItemProvider.
@objc public final class BytePortShareExtension: NSObject {
    @objc public static func bootstrapSession(itemCount: Int) -> BytePortShareHandle {
        // No-op stub. Real impl will:
        //   1. Read NSItemProvider attachments from extensionContext
        //   2. Stream bytes into BytePort via the host-app IPC channel
        //   3. Return a handle the host process polls for completion
        let sessionId = UInt64.random(in: 1...UInt64.max)
        return BytePortShareHandle(sessionId: sessionId)
    }
}