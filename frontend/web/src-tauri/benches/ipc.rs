// TODO: rewrite after IPC refactor
//
// The original `IpcEnvelope` + `IpcEnvelope::sample_project_lookup()` fixture
// referenced here no longer exists in `app_lib::ipc` (the current IPC surface
// is `CreateUploadArgs` / `CreateUploadResponse`). Once the envelope contract
// is reintroduced (or replaced with a more meaningful IPC serialization bench
// against the real types), re-enable this benchmark.
//
// Keeping the file present (instead of deleting) so the `[[bench]] name = "ipc"`
// entry in Cargo.toml stays valid and the bench harness target does not need
// to be re-added later.

fn main() {
    // No-op bench placeholder. criterion-style benches will live here once the
    // IPC refactor lands.
}
