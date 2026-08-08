# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v1.5.x` (latest minor on `main`) | ✅ security fixes |
| `main` (unreleased) | ✅ development tip |
| `v1.4.x` | best-effort |
| `v1.0.x` – `v1.3.x` | best-effort |
| older tags | best-effort until EOL notice |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Preferred channels (in order):

1. **GitHub Security Advisory** (private) — Security → Advisories → Report a vulnerability on this repository  
2. Email **security@iome.sh**

Include:

- Description of the issue and impact  
- Reproduction steps or proof-of-concept  
- Affected commit / tag if known  

We aim to acknowledge reports within **72 hours** and provide a remediation timeline after triage.

## Threat model (memory kernel)

`github.com/iome-sh/memory` is a **local-primary hierarchical agent memory kernel** (Palace FS + optional Qdrant + optional ONNX embeddings). It is a **Go library**, not a multi-tenant cloud service and **not product Memory GA**.

| Trust boundary | Posture |
|----------------|---------|
| Local Palace filesystem (`BaseDir`) | **User data** — the process that opens the store can read/write all entries under that root. Treat the directory as confidential. |
| Multi-process shared root | **Not multi-tenant isolation** — concurrent writers to the same palace root are not a secure tenancy boundary; use separate roots (or OS isolation) per trust domain. |
| Optional embeddings | Loading ONNX / hugot models executes model graphs and may download assets; only load trusted model paths. |
| Optional Qdrant | Network client to operator-chosen endpoint; no cloud Memory SLA. Validate URLs and network exposure. |
| Compaction / host hooks | Host-supplied LLM or callbacks can see entry text; hosts must not log secrets. |

### Residual risks (honest)

- **Local FS palace is user data** — encryption at rest, backup, and access control are operator responsibilities.  
- **Shared palace root ≠ multi-tenant security** — do not assume file layout isolates customers.  
- **Optional embeddings load models** — model supply chain and native ORT/CUDA stacks are out of band of the pure-Go default path.  
- **Kernel-only** — this package is not Memory GA; product dual_write defaults OFF elsewhere; hosted Palace remains sunset until deliberate scale; mesh is optional via TUI/ops packs.  
- **Future public MCP host** (planned naming honesty): **`iomesh-memory-mcp`** — not `aion-memory-mcp`. The private aion broker / control plane stays private.  

### What this is *not*

- Not a multi-tenant hosted Palace  
- Not a substitute for OS-level sandboxing or disk encryption  
- Not a cloud Memory SLA or freemium hosted Memory product  
- Not automatic authority over remote mesh data without operator credentials  

## Hardening checklist for operators

1. Point `BaseDir` at a directory with appropriate OS permissions; do not share roots across untrusted tenants  
2. Prefer the pure-Go hugot backend for CI/dev; only enable ORT/CUDA with trusted native libraries  
3. Download models only from sources you trust; pin paths via `MEMORY_ONNX_MODEL_PATH`  
4. Do not commit palace contents, `.env`, or API keys  
5. Scope Qdrant endpoints to private networks when used  
6. Treat residual honesty docs under `docs/operations/` as process truth for GA claims  

## Dependency security

```bash
make vuln   # govulncheck ./...
make test   # unit tests (no mandatory ONNX/LongMemEval)
```

CI runs tests, `go vet`, `gofmt` check, and `govulncheck` on every PR. Heavy ONNX / LongMemEval gates remain optional (`workflow_dispatch`).

## Disclosure preference

Coordinated disclosure: we prefer to ship a fix (or mitigating docs) before public write-ups when the issue is exploitable in default configurations.
