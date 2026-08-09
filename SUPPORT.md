# Support

How to get help for **`github.com/iome-sh/memory`** — the public **library kernel** only.

## How to file issues (GitHub)

| Need | Where |
|------|--------|
| Usage questions / bugs | [GitHub Issues](https://github.com/iome-sh/memory/issues) — use issue templates when available |
| Feature / API requests | Same Issues tracker; link relevant docs or a minimal repro |
| Security vulnerability | Private [Security Advisory](https://github.com/iome-sh/memory/security/advisories/new) or **security@iome.sh** — see [SECURITY.md](SECURITY.md) |
| Kernel API / roadmap | [README.md](README.md), [docs/temporal-memory-kernel-roadmap.md](docs/temporal-memory-kernel-roadmap.md), [docs/operations/](docs/operations/) residual honesty pins |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

When opening an issue, prefer a **title that names the API surface** (e.g. search, as-of facts, embeddings) and a short, redacted repro.

## Scope — library kernel only

This repository supports the **Go module** surface:

- Package API: store, search, temporal helpers, multi-hop, optional embeddings/vectors  
- Install via `go get github.com/iome-sh/memory@vX.Y.Z`  
- Docs and residual honesty pins under `docs/`

It does **not** support hosted multitenant Memory, private control-plane brokers, or product GA declarations from this tracker alone.

## Related host — iomesh-memory-mcp

| Component | Role | Support home |
|-----------|------|----------------|
| **This repo (`memory`)** | Kernel library | Issues on **iome-sh/memory** |
| **[iomesh-memory-mcp](https://github.com/iome-sh/iomesh-memory-mcp)** | Lean MCP host binary that *consumes* this kernel | Issues on **iome-sh/iomesh-memory-mcp** |
| **[iomesh-tui](https://github.com/iome-sh/iomesh-tui)** | Optional agent TUI/CLI (mesh hooks) | Issues on **iome-sh/iomesh-tui** |

File host binary, install packaging, and MCP protocol issues on the **host** repo. File pure kernel/API bugs here. Cross-link when both are involved.

## What we maintain

- **Latest tagged `vX.Y.Z`** on `main` + the **current minor line** (see [RELEASING.md](RELEASING.md) **Support / version policy (E5)**)  
- **Best effort** on `main` tip (development; not a production pin)  
- Security fixes on the default / current minor when feasible — table in [SECURITY.md](SECURITY.md)  
- **No cloud Memory SLA** — library / local-primary kernel, not a hosted Palace product  

Production consumers should **pin semver tags** in `go.mod` (not only `@main`). Verification is via `go get` / `go.sum` / the public module proxy — not cosign on this library.

## What we do not provide here

- Hosted Palace / multitenant cloud Memory onboarding or uptime guarantees  
- Product **Memory GA** or **Edge Memory GA** install support as if declared from this package alone  
- Forever-green signed binary release guarantees (this is a library; binary peers own their packaging)  
- Guarantees about third-party Qdrant, ONNX Runtime, or model-hub availability  
- Private monorepo broker / control-plane support via this package  
- dual_write product defaults (product path is dual_write **OFF** elsewhere)

## Before filing an issue

1. Run `make check` or note CI failures  
2. Redact API keys, palace contents, and private paths from logs  
3. Include module version (`go list -m github.com/iome-sh/memory`) or commit SHA and OS  
4. Confirm the report is about the **kernel API** — not a request to invent hosted or bare Memory GA  
5. Prefer a **pinned tag** in the report when the bug is version-specific  

## Honesty

Support policy docs (**s1499** tip) ≠ invent Edge Memory GA · ≠ invent bare Memory GA · ≠ invent forever-green signed releases · dual_write **OFF** product path elsewhere · no invent live dogfood. Full version window: [RELEASING.md](RELEASING.md) · security matrix: [SECURITY.md](SECURITY.md).  
