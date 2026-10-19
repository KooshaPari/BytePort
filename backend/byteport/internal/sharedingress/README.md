# Shared ingress read model

This package is the first implementation slice for BytePort issue #438.

It deliberately supports **observation and planning only**. There is no Cloudflare
API client and no apply/delete/update method in this package.

## Invariants

- snapshots contain provider identifiers, topology metadata, ownership, and
  secret **references** only; never secret values;
- tunnel identity is immutable during import/adoption;
- every mutable route, DNS record, and Access application has an explicit owner;
- Cloudflare ingress route order is semantic: position is part of drift and the
  single catch-all route must be terminal;
- Cloudflare DNS records carry their provider record IDs so multiple records may
  legitimately share the same name and type;
- Access policy references preserve their ordered sequence and explicit unique,
  positive precedence; policy reorder/precedence changes are authorization drift;
- a plan is unsafe if it would add, update, or remove a resource whose owner is
  outside the caller's managed-owner scope;
- desired-only conflicts do not invent a live owner;
- ownership transfer is never inferred from desired state;
- connectors are inventory evidence, not desired mutable resources;
- changes/conflicts are sorted deterministically for review and audit;
- import/plan acceptance precedes any future provider apply capability.

A future Cloudflare reader should translate API responses into `Snapshot`, then
persist or emit that snapshot as evidence. The reader must not populate email
addresses, bearer tokens, client secrets, tunnel credentials, or other secret
material into this model.
