# M4 public-flip readiness checklist (kernel)

Operator / maintainer checklist for a **deliberate future** public visibility flip of **`github.com/iome-sh/memory`** (Palace kernel). This document is **readiness residual only** — it does **not** flip the repository public, does **not** invent Memory GA, and does **not** complete Option A M4.

**Serial stamp (program continuum):** free eng residual pin **s1467** · free eng concurrent **s1467+** after free-floor **s1465** · lag **s1466** · peers **s1468** (mcp) · **s1469** (TUI) · **s1470** (aion residual) mention only · free-floor peer **s1471** · free eng **s1473** final TUI-parity audit closeout · free eng after free-floor peer **s1473+**.

> **Hard non-claims (read first):** Repository is **still private** until a deliberate maintainer flip. **Residual PASS ≠ public flip**. This pin is **M4 readiness ≠ M4 complete** · **not Memory GA** · **dual_write OFF** (product path elsewhere) · **aion stays private** · **open boxes stay open** · hosted **Palace sunset** until scale · rates ~**$88** / ~**$119** honesty when pricing is mentioned (do not invent freemium Palace SLA here). Do **not** run `gh repo edit --visibility public` from this residual.

## Why this residual exists

Option A edge OSS phases: **M1** private kernel process bar (s1452) → **M2** private MCP host extract → **M3** edge dogfood/deprecations → **M4** deliberate public flip → **M5** post-OSS trust. M1 process bar can be green while visibility stays private. Without an explicit readiness SSOT, continuum work may over-read residual PASS as “we are public” or flip both modules out of order.

This pin freezes **kernel-first** flip order honesty and maintainer steps so M4 remains a deliberate act after re-audit.

## Flip order honesty (Option A)

| Order | Module | Status on this residual |
|-------|--------|-------------------------|
| **1. First** | **`github.com/iome-sh/memory`** (this kernel) | **Still private** · readiness residual only · never invent public here |
| **2. Then** | **`github.com/iome-sh/iomesh-memory-mcp`** (edge MCP host) | Peer continuum · **not flipped here** · never invent public here |
| Stay private | **aion** broker / CP / INSTALL_STORE / billing | **aion stays private** · not in public Option A surface |

Publish order is **kernel first**, then MCP host (+ GHCR `ghcr.io/iome-sh/iomesh-memory-mcp` when deliberate). Do **not** flip MCP public before the kernel. Do **not** invent either repo as public from this checklist PASS.

## Pre-flight re-audit (before any visibility flip)

Re-run and confirm **Pass** (or intentional residual) on:

1. **`docs/OPEN_SOURCE_AUDIT.md`** — full re-audit of process bar, security, honesty locks, residual risks.  
2. **LICENSE** (MIT) · **NOTICE** · **SECURITY.md** (security@iome.sh + advisory path) · **CONTRIBUTING** · **CODE_OF_CONDUCT** · **SUPPORT** · **RELEASING** · **CHANGELOG**.  
3. **CI** present and green on the tip that will be public: lint/gofmt · vet · test · govulncheck · `ci-success` (see `.github/workflows/ci.yml`).  
4. **Dependabot** / secret-scan / templates still present (s1452 bar).  
5. **Honesty locks** still true: kernel-only · **not Memory GA** · local-primary · **dual_write OFF** elsewhere · **Palace sunset** · mesh optional · future host named **`iomesh-memory-mcp`** (not `aion-memory-mcp`).  
6. **No accidental public-product claims** in README / RELEASE notes (no invent Memory GA, no freemium hosted Palace SLA, no invent dual_write ON).  
7. **History / serials** residual: program continuum serials may remain while private; decide forward CONTRIBUTING policy before flip (do **not** force-rewrite history lightly).  
8. **Sibling readiness**: MCP host process bar green **before** that repo’s flip (peer work; not completed by this residual). Kernel may go public first while MCP remains private.

Offline pin for this checklist:

```bash
make public-flip-readiness-gate
# or: bash scripts/public_flip_readiness_gate.sh
# soft skip (local only; do not use to invent flip readiness): SKIP_PUBLIC_FLIP_READINESS=1
```

Gate is **offline greps only** — it does **not** change GitHub visibility, does **not** require network beyond repo files, and **PASS ≠ public flip**. Soft skip via **`SKIP_PUBLIC_FLIP_READINESS=1`** is for local bypass only — soft skip ≠ invent public flip.

## Final pre-flight (s1473 TUI parity)

Closeout checklist for deliberate flip day. Completing this list is still **not** the flip — a human must change visibility.

- [ ] Re-run `make public-flip-readiness-gate` + `make ci` (both green)  
- [ ] [`OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) **Final TUI-parity audit (s1473)** section green (Partial rows honest: CodeQL, history)  
- [ ] CONTRIBUTING **Public repository policy** present (forward PR surface after flip)  
- [ ] Fix GitHub description if it still mentions **sudo-jin** (maintainer: `gh repo edit`)  
- [ ] Homepage **https://iome.sh**  
- [ ] Topics: `golang`, `memory`, `rag`, `embeddings`, `qdrant`  
- [ ] Delete-branch-on-merge enabled  
- [ ] Enable **Private vulnerability reporting**  
- [ ] Branch protection: require PR + status check **`ci-success`** (+ branches up to date)  
- [ ] Enable CodeQL default setup (optional but recommended; not claimed green until enabled)  
- [ ] Confirm no secrets: `git grep` / history review residual  
- [ ] Flip visibility **Public** (**deliberate** human act only)  
- [ ] Then allow **`iomesh-memory-mcp`** flip second (separate deliberate act after its own re-audit)  

**Still private until the deliberate visibility step.** Residual PASS ≠ public flip · not Memory GA · dual_write OFF · aion stays private · kernel first, then MCP.

## Hard locks (must remain true through flip day)

| Lock | Honesty |
|------|---------|
| Visibility | **Still private** until deliberate maintainer act |
| Residual gate | **Residual PASS ≠ public flip** · readiness PASS ≠ invent public |
| Product GA | **not Memory GA** · kernel ≠ product Memory GA |
| dual_write | **dual_write OFF** on product path elsewhere (not a kernel product flag ON) |
| aion | **aion stays private** (broker · CP · billing · INSTALL_STORE) |
| Pricing | rates ~**$88** / ~**$119** when rates mentioned · no invent freemium Palace SLA here |
| Open work | **open boxes stay open** (Dependabot, residual risks, Partial audit rows) |
| Hosted Palace | **Palace sunset** until deliberate scale · local-primary FS palace |
| Flip order | **kernel first**, then **`iomesh-memory-mcp`** · never invent either public from this pin |
| M4 class | M4 readiness ≠ M4 complete · ≠ invent public flip done |

## Post-flip maintainer steps (future — not this residual)

Expand residual-honest from [`docs/OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) “after going public” section. Execute **only** after deliberate decision and pre-flight re-audit:

1. **Visibility (deliberate only):** GitHub → Settings → General → Danger Zone → Change visibility → Public.  
   - Alternative: `gh repo edit iome-sh/memory --visibility public` **only** with explicit human approval — **not** automated by this residual or CI.  
2. Enable **Private vulnerability reporting** (if not already).  
3. **Branch protection** on `main`: require PR + required status check **`ci-success`**.  
4. Confirm **SECURITY.md** contact path still works (security@iome.sh · advisory).  
5. (Optional) repository topics: `golang`, `memory`, `rag`, `qdrant`, `onnx`.  
6. Confirm **CONTRIBUTING** public-policy section if continuum ledger serials must be stripped from *forward* PRs (no reckless history rewrite).  
7. Do **not** publish private monorepo install paths, invent Memory GA, turn dual_write ON by default, or claim aion is public.  
8. Announce / tag only with **RELEASING.md** process — version tags ≠ visibility flip.  
9. **Then** schedule MCP host flip (`iomesh-memory-mcp`) as a **separate** deliberate act after its own re-audit (peer serials mention only).  
10. GHCR publish of `ghcr.io/iome-sh/iomesh-memory-mcp` remains MCP-host work — not this kernel residual.

## Peers (mention only)

| Peer | Role (mention only) |
|------|---------------------|
| **s1468** mcp | MCP host continuum / readiness peer |
| **s1469** TUI | Edge install honesty / product tip peer |
| **s1470** aion residual | Option A residual pin peer · aion stays private |
| **s1471** free-floor peer | Free-floor peer serial (does not rewrite free-floor **s1465**) |
| **s1473+** free eng | Free eng after free-floor peer · next continuum floor |

This residual closes **kernel readiness documentation + offline gate** only. It does **not** flip visibility, does **not** land MCP public, does **not** rewrite free-floor **s1465**, and does **not** invent live dogfood green.

## Related artifacts

| Artifact | Role |
|----------|------|
| [`docs/OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) | Process-bar audit SSOT · re-audit before flip |
| [`scripts/public_flip_readiness_gate.sh`](../scripts/public_flip_readiness_gate.sh) | Offline readiness greps · `make public-flip-readiness-gate` |
| s1452 OSS process bar | LICENSE / SECURITY / CI / community docs (still private intentional) |
| s1473 final TUI-parity audit | CONTRIBUTING public policy · OPEN_SOURCE_AUDIT final matrix · this pre-flight |
| Future MCP host | `github.com/iome-sh/iomesh-memory-mcp` · flip **after** kernel |

## Audit verdict class (s1467 readiness · s1473 final TUI parity)

| Dimension | Verdict |
|-----------|---------|
| M4 public-flip **readiness** docs + offline gate | **Pass** (s1467 residual) |
| Final TUI-parity audit + pre-flight checklist | **Pass** (s1473 · still private) |
| Visibility public flip | **Not done** — **still private** by design |
| Residual PASS = public flip? | **No** — **residual PASS ≠ public flip** |
| Memory GA | **No** — **not Memory GA** |
| dual_write | **OFF** elsewhere |
| aion | **private** |
| M4 complete / invent public | **No** — readiness + final audit only |

**Overall:** **Ready for deliberate public flip** after final pre-flight (s1473) + human checklist. **Still private.** Kernel first, then MCP. Residual PASS ≠ public flip · not Memory GA · dual_write OFF · aion stays private · open boxes stay open · Palace sunset · rates ~$88/$119 when rates apply.
