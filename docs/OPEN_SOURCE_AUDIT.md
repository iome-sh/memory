# Open-source readiness audit

Checklist for bringing **github.com/iome-sh/memory** (Palace kernel) to the same OSS **process bar** as public **iomesh-tui**. Re-run before any deliberate visibility flip and before each major release.

**Serial stamp (program continuum):** free eng residual pin **s1467** (M4 public-flip **readiness**) · free eng concurrent **s1467+** after free-floor **s1465** · lag **s1466** · peers s1468 (mcp) · s1469 (TUI) · s1470 (aion residual) mention only · free-floor peer **s1471** · free eng **s1473** final TUI-parity audit closeout · free eng **s1473+**. Prior process bar: **s1452** (still private intentional Pass). Does **not** rewrite free-floor **s1465**.

> **M4 readiness pointer:** operator checklist + offline gate live in [`docs/PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md) (`make public-flip-readiness-gate`). **Still private.** **Residual PASS ≠ public flip.** Readiness ≠ invent public / Memory GA / dual_write ON.

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Public** (MIT) — flipped deliberately. Deliberate future flip only after re-audit + [`PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md). |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |
| Residual PASS ≠ public flip | Pass (s1467 readiness pin · s1473 final audit) |

## Security

| Check | Status |
|-------|--------|
| No committed API keys / private keys / `.env` secrets | Pass (tests use fixtures / temp dirs) |
| Local Palace FS treated as user data in SECURITY.md | Pass |
| Multi-process shared root **not** claimed as multi-tenant isolation | Pass |
| Optional embeddings / model load residual risks documented | Pass |
| Kernel-only · not Memory GA residual risks documented | Pass |
| Vulnerability reporting path (advisory + security@iome.sh) | Pass |
| `.gitignore` covers local data / binaries / env | Pass (existing) |
| govulncheck in CI | Pass (s1452) |
| Residual: git history may contain internal ledger serials | **Partial** — do not rewrite history; forward policy in CONTRIBUTING (s1473) |

## Honesty locks (product narrative)

| Claim | Status |
|-------|--------|
| Kernel-only · **not product Memory GA** · **not Memory GA** | Pass |
| Local-primary Memory path | Pass |
| dual_write OFF product path (elsewhere) | Pass (docs/CONTRIBUTING/SECURITY) |
| Hosted Palace sunset until scale | Pass |
| Mesh optional (TUI / ops packs) · aion broker stays private | Pass |
| Future public MCP host named **iomesh-memory-mcp** (not aion-memory-mcp) | Pass |
| Residual PASS ≠ live dogfood / invent GA | Pass |
| Residual PASS ≠ public flip | Pass (s1467 · s1473) |
| Option A edge OSS: open **kernel first**, then MCP; private flip separate deliberate act | Pass |
| Open boxes stay open | Pass (Partial rows remain honest) |

## Open-source process artifacts

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present (s1452) |
| NOTICE (third-party acknowledgements) | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING | Present (+ public repository policy s1473) |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING | Present (library module · no GoReleaser required s1473 · M5 signing/matrix residual s1491) |
| PR template | Present |
| Issue templates (bug/feature) + security contact link | Present |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| Optional LongMemEval/ONNX (dispatch only — not mandatory CI) | Present |
| Dependabot (gomod + actions) | Present |
| Makefile `ci` / `check` / `vuln` / `fmt-check` | Present |
| Makefile `public-flip-readiness-gate` (s1467) | Present |
| `docs/PUBLIC_FLIP_READINESS.md` (s1467 M4 readiness · s1473 final pre-flight) | Present |
| README badges + links + honesty locks | Present |

## Final TUI-parity audit (s1473)

Closeout matrix vs public **iomesh-tui** process bar for a Go **library** module. **Do not** treat this section as a visibility flip.

| Check | Status |
|-------|--------|
| LICENSE MIT IOMesh | Pass |
| NOTICE | Pass |
| COC/CONTRIBUTING/SECURITY/SUPPORT/CHANGELOG/RELEASING | Pass |
| CI lint/test/build/govulncheck/ci-success | Pass |
| Dependabot gomod+actions | Pass |
| Issue templates + security contact | Pass |
| Public repository policy in CONTRIBUTING | Pass (s1473) |
| Branch protection recommended documented | Pass |
| No committed secrets | Pass |
| No aion import | Pass |
| Repo description honesty (not sudo-jin) | Pass after settings (document) |
| CodeQL | Partial / enable on flip (document) |
| Visibility still private | Pass intentional |
| residual PASS ≠ public flip | Pass |
| History may contain ledger serials | Partial — forward policy |
| GoReleaser binaries | N/A (library module; tags + go get) |

### Notes on Partial / settings rows

- **CodeQL:** not claimed green. Enable default setup on or after deliberate public flip (recommended).  
- **History serials:** residual docs may still stamp continuum serials while private; CONTRIBUTING public repository policy governs the **forward** PR surface after flip.  
- **Repo description / topics / homepage / delete-branch-on-merge:** maintainer GitHub settings — see [`PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md) final pre-flight. Description must not claim `sudo-jin` module path.  
- **GoReleaser:** N/A for this library; primary artifact is annotated git tag + `go get` (see [RELEASING.md](../RELEASING.md)). Binary products (iomesh-tui / iomesh-memory-mcp) differ.  
- **M5 signing / matrix (s1491):** library consumers verify via `go.sum` / module proxy checksum DB — **not** cosign on this package. Peer binary cosign/SBOM is mention-only; residual PASS ≠ invent M5 complete · ≠ invent cosign-on-library · ≠ invent Memory GA (see [RELEASING.md](../RELEASING.md#m5-signing--release-matrix-s1491-residual-tip)).

## Residual risks (Fail / Partial notes)

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility public | **Pass** (deliberate flip) | s1452 process bar · s1467 readiness · s1473 final audit — none flips public |
| Multi-tenant isolation | N/A / residual | Not a design goal of local FS palace |
| Model supply chain | Residual | Operators download models; not vendored by default |
| Concurrent writers on shared root | Residual | Documented; not multi-tenant safe |
| History may contain private program serials | Partial | No history rewrite; CONTRIBUTING forward policy after flip |
| Product install surface | Outside this repo | TUI + future **iomesh-memory-mcp**; aion stays private |
| CodeQL not enabled | Partial | Enable on flip (recommended); not claimed green here |
| Open boxes stay open | Intentional | Dependabot PRs, Partial rows, MCP host flip — not closed by this residual |

## Maintainer actions **after** going public (future — not this PR)

See expanded residual-honest checklist in [`docs/PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md). Summary:

1. GitHub → Settings → General → Danger Zone → Change visibility → Public (**deliberate**)  
2. Enable **Private vulnerability reporting**  
3. Branch protection on `main`: require PR + status check **`ci-success`**  
4. (Optional) topics: `golang`, `memory`, `rag`, `embeddings`, `qdrant`  
5. CONTRIBUTING **public repository policy** already present (s1473) — enforce on forward PRs after flip  
6. (Optional recommended) Enable CodeQL default setup  
7. Do **not** publish private monorepo install paths or invent Memory GA  
8. Flip order: **this kernel first**, then deliberate **`iomesh-memory-mcp`** public (never invent either public from residual PASS)

## Out of scope for this kernel package

- Multitenant hosted Palace / cloud Memory SLA  
- Product Memory GA edge binary (→ **iomesh-memory-mcp** when extracted)  
- Private aion broker / control plane  
- Guarantees about third-party Qdrant / model hubs  
- Visibility flip itself (deliberate maintainer act only · residual PASS ≠ public flip)  
- GoReleaser binary shipping (library module; tags + go get)

## Audit verdict (s1473 final TUI parity · s1467 readiness · process bar s1452)

| Dimension | Verdict |
|-----------|---------|
| Process bar vs iomesh-tui | **Pass** (artifacts + CI spirit + public policy · s1473) |
| M4 public-flip **readiness** docs + offline gate | **Pass** (s1467 · still private) |
| Final TUI-parity audit matrix | **Pass** (s1473 · Partial rows honest: CodeQL, history serials) |
| Visibility public flip | **Done** |
| Product honesty | **Pass** |
| Security docs | **Pass** |
| Live dogfood / rates | N/A — residual PASS ≠ dogfood; rates ~$88/$119 when pricing mentioned elsewhere; no freemium Palace invent |

**Overall:** **Ready for deliberate public flip** after maintainer checklist in [`PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md). Repository is **still private** until a human flips visibility. **Residual PASS ≠ public flip.** Kernel first, then MCP. **not Memory GA** · dual_write OFF · aion stays private · open boxes stay open · Palace sunset.


## Public import

```bash
# No PAT / GOPRIVATE:
go get github.com/iome-sh/memory@main
```
