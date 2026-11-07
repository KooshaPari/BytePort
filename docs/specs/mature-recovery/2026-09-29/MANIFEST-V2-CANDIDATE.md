# BytePort manifest v2 candidate — generalized single-YAML desired state

Date: 2026-09-30.
Status: **CANDIDATE DESIGN / NOT FROZEN**.

## Design objective

Preserve the direct-user interaction thesis:

`repository + one compact YAML → infrastructure/deployment lifecycle`

while supporting:
- application services/jobs;
- managed resources;
- bare-metal/VM/cloud/local targets;
- networks/storage;
- lifecycle policy;
- dependencies;
- build/configuration adapters;
- future resource/provider kinds.

The historical uppercase `NAME/SERVICES` manifest remains a compatibility input, not the mature ontology.

## Prior-art conclusions

### Compose-style fixed sections
Strength: excellent application ergonomics.
Weakness: fixed service/network/volume ontology does not generalize naturally to broad infrastructure.

### Kubernetes/Crossplane-style resource objects
Strength: strong kind/version/provider extensibility and desired/external separation.
Weakness: verbose and control-plane-shaped for the one-file repository UX.

### CloudFormation/Pulumi generic resource map
Strength: logical IDs + typed resources + properties map generalize broadly.
Weakness: provider type systems can leak directly into portable product identity.

## Candidate shape

~~~yaml
version: byteport/v2alpha1
name: example

targets:
  prod:
    adapter: aws
    spec:
      region: us-west-2

  lab:
    adapter: ironic
    lifecycle:
      default: manage
    spec:
      endpoint: secret://infra/ironic-endpoint
      credentials: secret://infra/ironic

resources:
  network:
    kind: network
    target: prod
    lifecycle: manage
    spec:
      cidr: 10.20.0.0/16

  db:
    kind: managed.postgres
    target: prod
    depends_on: [network]
    lifecycle: orphan_on_remove
    spec:
      size: small

  web:
    kind: service
    target: prod
    depends_on: [db]
    source:
      path: ./web
    build:
      adapter: auto
    spec:
      port: 8080

  lab-host:
    kind: host
    target: lab
    lifecycle: destroy_on_explicit_intent
    spec:
      selector:
        asset_tag: lab-01
      image: ubuntu-24.04
      configuration:
        adapter: colmena
        ref: ./infra/lab-host.nix

outputs:
  app_url:
    from:
      resource: web
      field: endpoint
~~~

This is illustrative, not final syntax.

## Core-owned fields

BytePort core should understand and validate:
- version;
- name;
- target logical IDs;
- target adapter;
- resource logical ID;
- resource kind;
- target reference;
- depends_on;
- lifecycle;
- source/build/configuration adapter references;
- spec digest;
- output/reference structure.

Adapter-specific `spec` is validated against the selected adapter/resource-kind schema.

Core MUST NOT silently accept arbitrary unknown core keys.

## Logical identity

Resource logical ID is stable manifest identity within a stack/project.
Actual identity remains:
`ManifestRevision + logical resource ID → DesiredResource`.

Changing the manifest file/branch without resolving a new ManifestRevision cannot mutate an existing desired identity in place.

## Lifecycle

Candidate lifecycle values:
- observe_only;
- create_observe;
- manage;
- orphan_on_remove;
- destroy_on_explicit_intent.

Resource deletion from YAML does not itself prove provider destruction authorization.

## Targets

Target is first-class and provider-neutral.
Adapter-specific target settings remain under `spec`.

Examples:
- cloud account/region;
- bare-metal engine/site;
- local host;
- SSH/Nix host set;
- VM/MicroVM runtime.

A resource kind may require target capabilities that another target does not provide; validation fails/degrades explicitly.

## Build vs provision vs configure

These remain separate:
- `build`: produce/resolve runnable artifact;
- target adapter: provision/realize infrastructure;
- `configuration`: post-provision host/application configuration where applicable.

One resource may use none, one or several.

## Historical v1 compatibility

Recovered historical manifest:
`NAME + SERVICES[] + INFRASTRUCTURE? + PORTFOLIO?`

should be imported through a versioned compatibility adapter:
- each SERVICE → DesiredResource(kind=service);
- historical infrastructure target fields → Target candidate;
- PORTFOLIO → downstream projection/publication configuration;
- missing mature identities/lifecycle use explicit compatibility defaults.

Do not mutate the historical parser into a forever-growing implicit schema.

## Schema/extension strategy

Preferred:
- BytePort core JSON Schema or equivalent for core structure;
- adapter/resource-kind schemas registered by adapters;
- exact schema versions recorded in ManifestRevision;
- deterministic normalized graph plus digest;
- raw manifest bytes retained/digested separately.

Avoid:
- unversioned permissive YAML;
- arbitrary executable expressions in core v2;
- raw provider-specific types as the only resource identity;
- inline secret values.

## Open design questions

- whether `kind` should be namespaced/versioned, e.g. `aws.ec2.instance/v1`;
- portable kind vs provider-native kind layering;
- imports/modules/multi-file composition;
- expression/reference syntax;
- parameter/environment overlays;
- whether targets live in the same file or can be policy-bound externally;
- how portfolio/productization config appears;
- generated schema/IDE support;
- migration from v1.

## Acceptance before freeze

Adversarial fixtures must include:
- unknown core key;
- unknown adapter;
- resource points to missing target;
- dependency cycle;
- target capability mismatch;
- inline secret attempt;
- removed managed resource without destruction intent;
- same logical ID with changed kind;
- v1 import;
- mixed service/database/network/host graph;
- bare-metal target;
- identical semantic graph from stable raw revision.

No live parser integration until these design questions and fixtures are resolved.
