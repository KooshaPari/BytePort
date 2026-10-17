# BytePort mature-recovery vertical fixture

This is a deliberately tiny application used only to test BytePort's source → artifact → runtime evidence chain.

It is **not** a BytePort implementation and does not define the canonical `odin.nvms` schema.

Build arguments:
- `SOURCE_COMMIT`
- `MANIFEST_DIGEST`
- `FIXTURE_NONCE`

The image writes these values into `/__byteport_probe`. A deployment oracle must independently fetch that endpoint and compare it with the requested SourceSnapshot/ManifestRevision/build run.

The fixture intentionally uses a standard OCI container build so candidate build engines can be compared without giving BytePort a bespoke builder.

Required experiment:
1. resolve this fixture to an immutable source commit;
2. bind a separately hashed experiment manifest/configuration;
3. build through candidate engine;
4. record immutable image digest;
5. deploy the digest through a runtime adapter;
6. fetch `/__byteport_probe`;
7. verify source/manifest/nonce and artifact digest chain;
8. use a ProviderResourceID deliberately different from ProjectID;
9. restart/reconcile;
10. stop only the intended provider resource.

Negative controls include wrong digest, mutable-tag replacement, branch movement, wrong nonce, provider success with lost response, local persistence failure, wrong principal/target and stale portfolio projection.
