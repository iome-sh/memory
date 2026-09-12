**Status: FLIP COMPLETE (public)** — no `GOPRIVATE` required for this module.

# M4 public-flip readiness (kernel)

**Not current operator documentation.** Visibility is already **public MIT**. This file is a **maintainer residual** kept so `make public-flip-readiness-gate` still has an offline SSOT. It is **not** a how-to-go-public guide and **not** a product spec. New contributors should read [README.md](../README.md), [SECURITY.md](../SECURITY.md), and [CONTRIBUTING.md](../CONTRIBUTING.md).

Operator / maintainer residual for the **completed** public visibility flip of **`github.com/iome-sh/memory`** (Palace kernel). **Flip complete is current fact.** Pre-flip checklist language below is **historical**. This document is **readiness residual only** — it does **not** flip visibility again, and **M4 readiness ≠ M4 complete** (readiness docs are not a product release).

This file is a maintainer residual for the completed public MIT visibility of this module. Sibling host and TUI live in their own public repositories.

> **Hard non-claims (read first):** Repository is **public MIT** (flip complete). **Residual PASS ≠ public flip** (gate PASS is not a product release · gate does not invent or re-do the flip). This pin is **M4 readiness ≠ M4 complete** · **private control-plane / broker stays private** · **open boxes stay open** · hosted **Palace sunset** until scale · mesh optional · do not invent freemium Palace SLA or a priced SKU here. Do **not** run `gh repo edit --visibility public` from this residual.

## Why this residual exists

Option A edge OSS phases (**historical**): **M1** private kernel process bar → **M2** private MCP host extract → **M3** edge dogfood/deprecations → **M4** deliberate public flip → **M5** post-OSS trust. This kernel completed M4 visibility. M1 process bar could be green while visibility was still private. Without an explicit readiness SSOT, continuum work may over-read residual PASS as a product release or as flipping sibling modules out of order.

This pin freezes **kernel-first** flip order. Gate PASS is not a product release.

## Flip order (Option A)

| Order | Module | Status on this residual |
|-------|--------|-------------------------|
| **1. First** | **`github.com/iome-sh/memory`** (this kernel) | **Public MIT** · flip complete |
| **2. Then** | **`github.com/iome-sh/iomesh-memory-mcp`** (edge MCP host) | Peer continuum · **not flipped here** |
| Stay private | private control-plane / broker / INSTALL_STORE / billing | **private control-plane / broker stays private** · not in public Option A surface |

Publish order is **kernel first**, then MCP host (+ GHCR `ghcr.io/iome-sh/iomesh-memory-mcp` when deliberate). Do **not** flip MCP public before the kernel. Do **not** invent a product release from this checklist PASS.

## Historical pre-flight re-audit (used before the visibility flip)

Re-run and confirm **Pass** (or intentional residual) on:

1. **`docs/OPEN_SOURCE_AUDIT.md`** — full re-audit of process bar, security, scope locks, residual risks. 
2. **LICENSE** (MIT) · **NOTICE** · **SECURITY.md** (security@iome.sh + advisory path) · **CONTRIBUTING** · **CODE_OF_CONDUCT** · **SUPPORT** · **RELEASING** · **CHANGELOG**. 
3. **CI** present and green on the public tip: lint/gofmt · vet · test · govulncheck · `ci-success` (see `.github/workflows/ci.yml`). 
4. **Dependabot** / secret-scan / templates still present ( bar). 
5. **Scope locks** still true: library kernel · local-primary · **Palace sunset** · mesh optional · host named **`iomesh-memory-mcp`**. 
6. **No accidental public-product claims** in README / RELEASE notes (no invent freemium hosted Palace SLA). 
7. **History / serials** residual: program continuum serials may remain in historical residual docs; CONTRIBUTING public-repository policy governs the **forward** PR surface (do **not** force-rewrite history lightly). 
8. **Sibling readiness**: MCP host process bar green **before** that repo’s flip (peer work; not completed by this residual). Kernel is public; MCP host flip is a separate deliberate act.

Offline pin for this checklist:

```bash
make public-flip-readiness-gate
# or: bash scripts/public_flip_readiness_gate.sh
# soft skip (local only; do not use to invent flip readiness): SKIP_PUBLIC_FLIP_READINESS=1
```

Gate is **offline greps only** — it does **not** change GitHub visibility, does **not** require network beyond repo files, and **residual PASS ≠ public flip** (gate PASS is not a product release). Soft skip via **`SKIP_PUBLIC_FLIP_READINESS=1`** is for local bypass only — soft skip ≠ invent public flip.

## Historical pre-flight ( TUI parity)

Closeout checklist used on flip day (**historical**). Completing this list was **not** the flip — a human changed visibility. **Flip complete (public MIT).** Remaining open boxes (CodeQL, etc.) stay open.

- [ ] Re-run `make public-flip-readiness-gate` + `make ci` (both green) 
- [ ] [`OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) **Final TUI-parity audit** section green (Partial rows honest: CodeQL, history) 
- [ ] CONTRIBUTING **Public repository policy** present (forward PR surface) 
- [ ] Fix GitHub description if it still names a personal maintainer handle (maintainer: `gh repo edit`) 
- [ ] Homepage **https://iome.sh** 
- [ ] Topics: `golang`, `memory`, `rag`, `embeddings`, `qdrant` 
- [ ] Delete-branch-on-merge enabled 
- [ ] Enable **Private vulnerability reporting** 
- [ ] Branch protection: require PR + status check **`ci-success`** (+ branches up to date) 
- [ ] Enable CodeQL default setup (optional but recommended; not claimed green until enabled) 
- [ ] Confirm no secrets: `git grep` / history review residual 
- [x] Flip visibility **Public** (**deliberate** human act — **done**) 
- [ ] Then allow **`iomesh-memory-mcp`** flip second (separate deliberate act after its own re-audit) 

**Public MIT** (flip complete). Residual PASS ≠ public flip · private control-plane / broker stays private · kernel first, then MCP.

## Hard locks (must remain true)

| Lock | Truth |
|------|---------|
| Visibility | **Public MIT** — flip complete (current fact) |
| Residual gate | **Residual PASS ≠ public flip** · gate PASS is not a product release |
| Library | this package is the **library kernel**, not a hosted Memory product |
| private plane | **private control-plane / broker stays private** (broker · CP · billing · INSTALL_STORE) |
| Pricing | no SKU or dollar figures in this kernel · mesh optional · no invent freemium Palace SLA |
| Open work | **open boxes stay open** (Dependabot, residual risks, Partial audit rows) |
| Hosted Palace | **Palace sunset** until deliberate scale · local-primary FS palace |
| Flip order | **kernel first**, then **`iomesh-memory-mcp`** |
| M4 class | M4 readiness ≠ M4 complete |

## Post-flip maintainer steps (ongoing; visibility already public)

Expand residual-honest from [`docs/OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) post-flip section. Visibility change is **done**. Remaining steps are settings / sibling work:

1. **Visibility (historical / done):** GitHub → Settings → General → Danger Zone → Change visibility → Public. 
 - Alternative was: `gh repo edit iome-sh/memory --visibility public` **only** with explicit human approval — **not** automated by this residual or CI. 
2. Enable **Private vulnerability reporting** (if not already). 
3. **Branch protection** on `main`: require PR + required status check **`ci-success`**. 
4. Confirm **SECURITY.md** contact path still works (security@iome.sh · advisory). 
5. (Optional) repository topics: `golang`, `memory`, `rag`, `qdrant`, `onnx`. 
6. Confirm **CONTRIBUTING** public-policy section: continuum ledger serials stay off the *forward* PR surface (no reckless history rewrite). 
7. Do **not** publish private monorepo install paths or claim the private control-plane / broker is public. 
8. Announce / tag only with **RELEASING.md** process — version tags ≠ visibility flip. 
9. **Then** schedule MCP host flip (`iomesh-memory-mcp`) as a **separate** deliberate act after its own re-audit (peer serials mention only). 
10. GHCR publish of `ghcr.io/iome-sh/iomesh-memory-mcp` remains MCP-host work — not this kernel residual.

## Peers (mention only)

| Peer | Role (mention only) |
|------|---------------------|
| mcp | MCP host continuum / readiness peer |
| TUI | Edge install / product tip peer |
| private-plane residual | Option A residual pin peer · private control-plane / broker stays private |
| free-floor peer | Free-floor peer serial (does not rewrite free-floor ) |
| free eng | Free eng after free-floor peer · next continuum floor |

This residual closes **kernel readiness documentation + offline gate** only. It does **not** flip visibility, does **not** land MCP public, does **not** rewrite free-floor , and does **not** invent live dogfood green.

## Related artifacts

| Artifact | Role |
|----------|------|
| [`docs/OPEN_SOURCE_AUDIT.md`](OPEN_SOURCE_AUDIT.md) | Process-bar audit SSOT · re-audit before major releases |
| [`scripts/public_flip_readiness_gate.sh`](../scripts/public_flip_readiness_gate.sh) | Offline readiness greps · `make public-flip-readiness-gate` |
| OSS process bar | LICENSE / SECURITY / CI / community docs (historical: process bar landed while private) |
| final TUI-parity audit | CONTRIBUTING public policy · OPEN_SOURCE_AUDIT final matrix · this pre-flight |
| Future MCP host | `github.com/iome-sh/iomesh-memory-mcp` · flip **after** kernel |

## Audit verdict class ( readiness · final TUI parity)

| Dimension | Verdict |
|-----------|---------|
| M4 public-flip **readiness** docs + offline gate | **Pass** ( residual) |
| Final TUI-parity audit + pre-flight checklist | **Pass** ( · historical; flip complete) |
| Visibility public flip | **Done** — **public MIT** |
| Residual PASS = public flip? | **No** — **residual PASS ≠ public flip** |
| Library | library kernel · local filesystem palace |
| private plane | **private** |
| M4 complete | **No** — readiness + final audit only · M4 readiness ≠ M4 complete |

**Overall:** Flip complete (**public MIT**). Kernel first, then MCP. Residual PASS ≠ public flip · private control-plane / broker stays private · open boxes stay open · Palace sunset · mesh optional · no invent freemium Palace SLA.
