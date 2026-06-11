# Security Policy — BytePort

BytePort is a transport substrate used by production services in the
Phenotype ecosystem. Security is a primary feature, not an afterthought.
This document explains how to report a vulnerability, what to expect from
us, and how we handle disclosure.

---

## Supported Versions

BytePort follows [semantic versioning](https://semver.org/). The
following table lists the release lines that currently receive security
updates:

| Version line | Status            | Security fixes | End-of-life  |
|--------------|-------------------|----------------|--------------|
| `1.4.x`      | **Active**        | ✅ Backported  | TBA          |
| `1.3.x`      | Maintenance       | ✅ Until 2026-09-30 | 2026-09-30 |
| `1.2.x`      | End of life       | ❌ No longer receiving updates | 2025-12-31 |
| `1.1.x`      | End of life       | ❌ No longer receiving updates | 2025-06-30 |
| `1.0.x`      | End of life       | ❌ No longer receiving updates | 2024-12-31 |
| `< 1.0`      | Not supported     | ❌             | n/a          |

> **Recommendation:** always run the latest patch release of the latest
> two minor lines. We will publish a CVE and a GHSA for every security
> fix landed on `main`.

## Reporting a Vulnerability

**Please do not file a public GitHub issue for security bugs.**

The fastest, most private way to report a vulnerability is via one of
the channels below. Choose the one you are most comfortable with:

1. **GitHub private vulnerability reporting** —
   *Repository → Security → Advisories → "New draft security advisory"*.
   This is the preferred channel; it gives you a private thread with the
   maintainers, a CVE assignment, and a coordinated disclosure workflow.
2. **Email** — `security@phenotype.internal` (PGP key fingerprint:
   `B5C7 1F2E 9D44 8A6B 7E3C  4F2A 19AB 6C3D 8E1F 0A2B`). The mailbox
   is monitored 24/7 and triaged within 24 hours.
3. **Signal** — `@koosha.42` on Signal. Ask for our Signal safety
   number out-of-band before sharing details.

When you write in, please include (to the extent you can):

- A clear description of the issue and its impact.
- A reproducer — minimal Rust/Go/TypeScript snippet, packet capture,
  or a curl-driven repro against the sidecar.
- The affected commit SHA, tag, or release version.
- Any known workarounds or mitigations.
- Your name / handle for credit (optional; we will not credit by
  default if you request anonymity).

### What *not* to send

- Do not include real customer data, tokens, or PII in a report.
- Do not exploit the issue beyond what is necessary to demonstrate it.
- Do not publish details, screenshots, or PoCs publicly until we have
  agreed a disclosure date (see §4).

## Response Timeline

We commit to the following SLOs. "Business hours" = 09:00–18:00 UTC,
Mon–Fri excluding Phenotype holidays.

| Stage                            | SLO                             |
|----------------------------------|---------------------------------|
| **Acknowledgement**              | ≤ 24 hours, every report        |
| **Triage & severity assignment** | ≤ 3 business days               |
| **Patch for Critical / High**    | ≤ 7 days                        |
| **Patch for Medium**             | ≤ 30 days                       |
| **Patch for Low / Informational**| ≤ 90 days (or accepted-risk)    |
| **CVE / GHSA assignment**        | ≤ 24 hours after triage         |
| **Disclosure coordination**      | Per §4                          |

We will keep you informed at every step. If we cannot meet an SLO we
will tell you why, and we will agree a new date with you.

## Coordinated Disclosure

We follow a 90-day coordinated disclosure window from the date of
acknowledgement, modelled on [Google's project-zero timeline](https://googleprojectzero.blogspot.com/p/vulnerability-disclosure-faq.html).
Concretely:

- **Day 0** — you report the issue.
- **Day 0–7** — we triage, agree severity, and start a fix branch.
- **Day 7–60** — we develop, test, and backport the fix on a private
  advisory branch.
- **Day 60–75** — we prepare the advisory, CVE, and release notes.
- **Day 75–90** — embargo; we agree a release date with you and
  downstream consumers.
- **Day 90** — public disclosure: advisory + CVE + release tags +
  blog post. We credit you in the advisory unless you opted out.

We are happy to negotiate the disclosure date, especially for
issues that require substantial coordinated rollout across the
Phenotype mesh. Just tell us your constraints.

## Severity Rating

We use CVSS v3.1 base scores as a starting point:

| Severity     | CVSS range  | Examples                                       |
|--------------|-------------|------------------------------------------------|
| **Critical** | 9.0 – 10.0  | RCE, auth bypass on admin RPC, unauthenticated  |
|              |             | remote code execution via codec                |
| **High**     | 7.0 – 8.9   | Privilege escalation, memory-safety RCE with   |
|              |             | auth, sidecar-wide DoS                         |
| **Medium**   | 4.0 – 6.9   | Information disclosure, targeted DoS, partial  |
|              |             | auth bypass                                    |
| **Low**      | 0.1 – 3.9   | Local-only info leaks, hardening recommendations |
| **Info**     | 0.0         | Best-practice deviations, no direct impact     |

## Security Tooling

BytePort is scanned continuously by:

- `cargo audit` + `cargo deny` (RustSec + license).
- `govulncheck` (Go vuln DB, on every PR).
- `golangci-lint` security ruleset (`gas`, `gosec`).
- `osv-scanner` across lockfiles (Rust, Go, npm).
- GitHub CodeQL (Rust, Go, JavaScript/TypeScript).
- OpenSSF Scorecard (weekly).
- `pnpm audit --prod` in CI.
- `trivy` filesystem scan in the release pipeline.

Reproduce locally with:

```bash
task security:scan
```

## Out of Scope

The following are **not** considered security vulnerabilities in
BytePort and should be filed as regular bugs:

- Denial of service via resource exhaustion from a peer that the
  operator has already authorised at the admin layer.
- Reports against unsupported (EOL) release lines.
- Findings that require physical access to the host running the
  sidecar.
- "Theoretical" issues without a concrete attack path.
- Reports against third-party dependencies that are not exploitable
  via BytePort's own code paths (please report those upstream too).

## Bug Bounty

BytePort is not currently running a paid bug bounty programme. We do
publicly credit researchers in the GitHub Security Advisory and in the
release notes, and we are happy to coordinate a joint blog post with
you after disclosure.

## Recognition

We are grateful to the following researchers for responsible
disclosures (most recent first):

- _Awaiting first advisory._

Thank you for helping keep BytePort — and everyone who depends on it —
safe.
