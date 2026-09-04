# Open-source readiness audit

Checklist for the OSS **process bar** of **github.com/iome-sh/memory** (Palace kernel) vs public **iomesh-tui**. Visibility flip is **complete** (public MIT). Re-run before each major release.

> Operator residual + offline gate live in [`docs/PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md) (`make public-flip-readiness-gate`). **Public MIT** (flip complete). **Public MIT ≠ Memory GA.** **Residual PASS ≠ public flip.** Gate PASS ≠ product GA. Readiness ≠ invent Memory GA / dual_write ON.

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Public** (MIT) — flipped deliberately. Flip complete is current fact. See [`PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md). |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |
| Residual PASS ≠ public flip | Pass (gate PASS ≠ product GA) |

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
| govulncheck in CI | Pass |
| Residual: git history may contain internal ledger serials | **Partial** — do not rewrite history; forward policy in CONTRIBUTING |

## Honesty locks (product narrative)

| Claim | Status |
|-------|--------|
| Kernel-only · **not product Memory GA** · **not Memory GA** · public MIT ≠ Memory GA | Pass |
| Local filesystem library · local-primary Memory path | Pass |
| dual_write OFF (host policy default elsewhere, **not** a kernel product flag) | Pass (docs/CONTRIBUTING/SECURITY) |
| Hosted Palace sunset until scale | Pass |
| Mesh optional (TUI / ops packs) | Pass |
| Public MCP host named **iomesh-memory-mcp** | Pass |
| This library does **not** implement mesh `X-IOMesh-Org` or cloud multi-tenant security | Pass |
| Residual PASS ≠ live dogfood / invent GA | Pass |
| Residual PASS ≠ public flip | Pass (gate PASS ≠ product GA) |
| Open **kernel first**, then MCP; sibling flip was a separate deliberate act | Pass |
| Open boxes stay open | Pass (Partial rows remain honest) |

## Open-source process artifacts

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present |
| NOTICE (third-party acknowledgements) | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING | Present (+ public repository policy) |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING | Present (library module · no GoReleaser required · signing/matrix for peers) |
| PR template | Present |
| Issue templates (bug/feature) + security contact link | Present |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| Optional LongMemEval/ONNX (dispatch only — not mandatory CI) | Present |
| Dependabot (gomod + actions) | Present |
| Makefile `ci` / `check` / `vuln` / `fmt-check` | Present |
| Makefile `public-flip-readiness-gate` | Present |
| `docs/PUBLIC_FLIP_READINESS.md` | Present |
| README badges + links + honesty locks | Present |

## Process-bar audit vs iomesh-tui

Closeout matrix vs public **iomesh-tui** process bar for a Go **library** module. **Do not** treat this section as a visibility flip. Flip is already complete.

| Check | Status |
|-------|--------|
| LICENSE MIT IOMesh | Pass |
| NOTICE | Pass |
| COC/CONTRIBUTING/SECURITY/SUPPORT/CHANGELOG/RELEASING | Pass |
| CI lint/test/build/govulncheck/ci-success | Pass |
| Dependabot gomod+actions | Pass |
| Issue templates + security contact | Pass |
| Public repository policy in CONTRIBUTING | Pass |
| Branch protection recommended documented | Pass |
| No committed secrets | Pass |
| Repo description honesty | Pass after settings (document) |
| CodeQL | Partial / recommended; not claimed green |
| Visibility public MIT | Pass (flip complete) |
| residual PASS ≠ public flip | Pass |
| History may contain ledger serials | Partial — forward policy |
| GoReleaser binaries | N/A (library module; tags + go get) |

### Notes on Partial / settings rows

- **CodeQL:** not claimed green. Enable default setup (recommended).  
- **History serials:** residual docs may still stamp continuum serials; CONTRIBUTING public repository policy governs the **forward** PR surface. Do not rewrite published history lightly.  
- **Repo description / topics / homepage / delete-branch-on-merge:** maintainer GitHub settings — see [`PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md) post-flip steps.  
- **GoReleaser:** N/A for this library; primary artifact is annotated git tag + `go get` (see [RELEASING.md](../RELEASING.md)). Binary products (iomesh-tui / iomesh-memory-mcp) differ.  
- **Signing / matrix:** library consumers verify via `go.sum` / module proxy checksum DB — **not** cosign on this package. Peer binary cosign/SBOM is mention-only; docs ≠ invent cosign-on-library · ≠ invent Memory GA (see [RELEASING.md](../RELEASING.md#signing--release-matrix)).

## Residual risks (Fail / Partial notes)

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility public | **Pass** (deliberate flip complete) | Process bar / readiness docs were not the flip act |
| Multi-tenant isolation | N/A / residual | Not a design goal of local FS palace |
| Model supply chain | Residual | Operators download models; not vendored by default |
| Concurrent writers on shared root | Residual | Documented; not multi-tenant safe |
| History may contain private program serials | Partial | No history rewrite; CONTRIBUTING forward policy |
| Product install surface | Outside this repo | TUI + **iomesh-memory-mcp** |
| CodeQL not enabled | Partial | Recommended; not claimed green here |
| Open boxes stay open | Intentional | Dependabot PRs, Partial rows — not closed by this residual |

## Maintainer actions after going public (visibility already public)

See expanded residual-honest checklist in [`docs/PUBLIC_FLIP_READINESS.md`](PUBLIC_FLIP_READINESS.md). Summary:

1. GitHub visibility → Public (**done** — flip complete)  
2. Enable **Private vulnerability reporting**  
3. Branch protection on `main`: require PR + status check **`ci-success`**  
4. (Optional) topics: `golang`, `memory`, `rag`, `embeddings`, `qdrant`  
5. CONTRIBUTING **public repository policy** already present — enforce on forward PRs  
6. (Optional recommended) Enable CodeQL default setup  
7. Do **not** publish private monorepo install paths or invent Memory GA  
8. Flip order: **this kernel first** (done), then **`iomesh-memory-mcp`** public (done as a separate act; never invent Memory GA from residual PASS)

## Out of scope for this kernel package

- Multitenant hosted Palace / cloud Memory SLA  
- Mesh `X-IOMesh-Org` HTTP headers (mesh clients, not this library)  
- Product Memory GA edge binary (→ **iomesh-memory-mcp**)  
- Guarantees about third-party Qdrant / model hubs  
- Re-doing the visibility flip (deliberate maintainer act already done · residual PASS ≠ public flip)  
- GoReleaser binary shipping (library module; tags + go get)

## Audit verdict

| Dimension | Verdict |
|-----------|---------|
| Process bar vs iomesh-tui | **Pass** (artifacts + CI spirit + public policy) |
| Public-flip **readiness** docs + offline gate | **Pass** (public MIT · gate PASS ≠ product GA) |
| Visibility public flip | **Done** — **public MIT** |
| Product honesty | **Pass** — public MIT ≠ Memory GA · local filesystem library |
| Security docs | **Pass** |
| Live dogfood / rates | N/A — residual PASS ≠ dogfood; no SKU or dollar figures here; mesh optional; no freemium Palace invent |

**Overall:** Flip complete (**public MIT**). **Public MIT ≠ Memory GA.** **Residual PASS ≠ public flip.** Gate PASS ≠ product GA. Kernel first, then MCP. **not Memory GA** · dual_write OFF (host policy, not kernel flag) · open boxes stay open · Palace sunset.


## Public import

```bash
# No PAT / GOPRIVATE:
go get github.com/iome-sh/memory@main
```
