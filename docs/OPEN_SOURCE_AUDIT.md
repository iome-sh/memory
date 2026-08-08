# Open-source readiness audit

Checklist for bringing **github.com/iome-sh/memory** (Palace kernel) to the same OSS **process bar** as public **iomesh-tui**. Re-run before any deliberate visibility flip and before each major release.

**Serial stamp (program continuum):** free eng concurrent **s1452+** after free-floor **s1450** · peer residual aion s1454 mention only · does **not** rewrite free-floor **s1455**. Peers: TUI s1453 edge install honesty · aion residual s1454 Option A charter (mention only).

## Visibility

| Check | Status |
|-------|--------|
| Repository visibility | **Still private** — do **not** flip public on this serial. Deliberate future flip only after re-audit. |
| Private vulnerability reporting path documented | Pass (SECURITY.md · security@iome.sh · advisory) |
| No accidental “we are public Memory GA” claims | Pass (honesty locks below) |

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
| Residual: git history may contain internal ledger serials | **Partial** — do not rewrite history; forward docs may carry continuum serials while private |

## Honesty locks (product narrative)

| Claim | Status |
|-------|--------|
| Kernel-only · **not product Memory GA** | Pass |
| Local-primary Memory path | Pass |
| dual_write OFF product path (elsewhere) | Pass (docs/CONTRIBUTING/SECURITY) |
| Hosted Palace sunset until scale | Pass |
| Mesh optional (TUI / ops packs) · aion broker stays private | Pass |
| Future public MCP host named **iomesh-memory-mcp** (not aion-memory-mcp) | Pass |
| Residual PASS ≠ live dogfood / invent GA | Pass |
| Option A edge OSS: open kernel later; private flip separate deliberate act | Pass |

## Open-source process artifacts

| Artifact | Status |
|----------|--------|
| LICENSE (MIT · IOMesh Technology Ltd.) | Present (s1452) |
| NOTICE (third-party acknowledgements) | Present |
| CODE_OF_CONDUCT | Present |
| CONTRIBUTING | Present |
| SECURITY | Present |
| SUPPORT | Present |
| CHANGELOG | Present |
| RELEASING | Present |
| PR template | Present |
| Issue templates (bug/feature) + security contact link | Present |
| CI (lint/gofmt, test, build, govulncheck, ci-success) | Present |
| Optional LongMemEval/ONNX (dispatch only — not mandatory CI) | Present |
| Dependabot (gomod + actions) | Present |
| Makefile `ci` / `check` / `vuln` / `fmt-check` | Present |
| README badges + links + honesty locks | Present |

## Residual risks (Fail / Partial notes)

| Risk | Rating | Notes |
|------|--------|-------|
| Visibility still private | **Pass (intentional)** | s1452 does not flip public |
| Multi-tenant isolation | N/A / residual | Not a design goal of local FS palace |
| Model supply chain | Residual | Operators download models; not vendored by default |
| Concurrent writers on shared root | Residual | Documented; not multi-tenant safe |
| History may contain private program serials | Partial | No history rewrite; strip or keep serials only as continuum stamps while private |
| Product install surface | Outside this repo | TUI + future **iomesh-memory-mcp**; aion stays private |

## Maintainer actions **after** going public (future — not this PR)

1. GitHub → Settings → General → Danger Zone → Change visibility → Public (**deliberate**)  
2. Enable **Private vulnerability reporting**  
3. Branch protection on `main`: require PR + status check **`ci-success`**  
4. (Optional) topics: `golang`, `memory`, `rag`, `qdrant`, `onnx`  
5. Confirm CONTRIBUTING public-policy section if ledger serials must be stripped from forward PRs  
6. Do **not** publish private monorepo install paths or invent Memory GA  

## Out of scope for this kernel package

- Multitenant hosted Palace / cloud Memory SLA  
- Product Memory GA edge binary (→ **iomesh-memory-mcp** when extracted)  
- Private aion broker / control plane  
- Guarantees about third-party Qdrant / model hubs  

## Audit verdict (s1452)

| Dimension | Verdict |
|-----------|---------|
| Process bar vs iomesh-tui | **Pass** (artifacts + CI spirit aligned) |
| Visibility public-ready flip | **Not done** — still private by design |
| Product honesty | **Pass** |
| Security docs | **Pass** |
| Live dogfood / rates | N/A — residual PASS ≠ dogfood; no freemium Palace invent |

**Overall:** Ready for **private** OSS process bar. Public flip remains a separate, deliberate decision after re-audit.
