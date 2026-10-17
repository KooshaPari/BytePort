# BytePort manifest schema recovery — pass 1

Date: 2026-09-30.

## Recovered shape

Current-tree independent sources agree on a YAML-like `odin.nvms` shape.

`SPEC.md` provides:
- NAME;
- DESCRIPTION;
- SERVICES[] with NAME, PATH, PORT and optional RUNTIME/BUILD/ENV;
- INFRASTRUCTURE with compute/region/instance_type;
- PORTFOLIO with generate_page/description_source.

`setup-windows.ps1` independently emits an `odin.nvms.template` containing NAME, DESCRIPTION and SERVICES with NAME/PATH/PORT.

This is enough to recover a **candidate historical manifest schema family**, not enough to declare every SPEC field mandatory/current.

## Authority caveat

The same SPEC says:
- it reflects actual shipping implementation;
- prior NanoVMS references are retired.

Those claims conflict with the frozen current implementation, which still performs NanoVMS sandbox calls. Therefore the document is internally useful for manifest design but not reliable implementation-status authority.

## Minimal B03 parser contract

For the first selected-app slice, require only:
- NAME non-empty;
- at least one SERVICES item;
- service NAME non-empty and unique;
- PATH non-empty, repository-relative and non-escaping;
- PORT valid 1..65535;
- unknown top-level/service keys rejected or explicitly version-gated;
- exact manifest bytes content-digested before parsing;
- schema version recorded explicitly by BytePort even if historical files lacked a VERSION field.

Optional historical fields may be retained as extension data until authority is stronger:
- RUNTIME;
- BUILD;
- ENV;
- INFRASTRUCTURE;
- PORTFOLIO.

## Security boundary

BUILD commands and ENV values are untrusted manifest content. Parsing a manifest does not authorize executing BUILD commands or exposing ENV. BuildEngine/secret policy independently decides what is allowed.

## Next experiment

Use a checked-in recovery fixture derived from the overlapping SPEC/setup shape, clearly labeled as a recovery fixture rather than pretending the missing historical `odin.nvms` file exists.

Negative fixtures:
- missing NAME;
- no services;
- duplicate service names;
- path escape;
- invalid port;
- unknown key;
- malformed YAML.

B03 remains prototype-only until exact source-ref resolution and parser fixtures pass.
