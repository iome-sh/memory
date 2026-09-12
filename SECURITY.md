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

`github.com/iome-sh/memory` is a **local-primary hierarchical agent memory kernel** (Palace FS + optional Qdrant + optional ONNX embeddings). It is a **Go library**, not a multi-tenant cloud service.

| Trust boundary | Posture |
|----------------|---------|
| Local Palace filesystem (`BaseDir`) | **User data** — the process that opens the store can read/write all entries under that root. Treat the directory as confidential. |
| Multi-process shared root | **Single-writer contract** — one process per palace root. Concurrent writers to the same root are unsupported (last-write-wins on a given entry file). That is the product contract, not a hidden defect. Path layout is not a secure tenancy boundary; use separate roots (or OS isolation) per trust domain. Probe script exists (`scripts/two_process_writer_probe.sh`); flock is not shipped. |
| Optional embeddings | Loading ONNX / hugot models executes model graphs and may download assets; only load trusted model paths. |
| Optional Qdrant | Network client to operator-chosen endpoint; no cloud Memory SLA. Validate URLs and network exposure. |
| Compaction / host hooks | Host-supplied LLM or callbacks can see entry text; hosts must not log secrets. |

### Residual risks

- **Local FS palace is user data** — encryption at rest, backup, and access control are operator responsibilities. New palace dirs are created `0700` and kernel-written files `0600`; pre-existing trees are not retroactively chmod'd. Mode bits are not encryption at rest.  
- **Shared palace root ≠ multi-tenant security** — do not assume file layout isolates customers. Supported topology is **one host process per palace root**. In-process `writeMu` serializes `entity-graph.json` / `event-time.json` rewrites; multi-process writers to the same root remain unsupported (not flock, not cloud isolation). Probe script exists; flock is not shipped.  
- **Optional embeddings load models** — model supply chain and native ORT/CUDA stacks are out of band of the pure-Go default path. Stored vectors are optional (`PersistEmbeddings` default off); hash embeddings are never persisted as `QueryVec` / stored vectors.  
- **Library kernel** — hosted Palace remains sunset until deliberate scale; mesh is optional via TUI/ops packs.  
- **No mesh org header** — organization isolation for the I/O Mesh broker is a separate HTTP header (`X-IOMesh-Org`) on mesh clients; this library does not implement that. The public MCP host is **`iomesh-memory-mcp`**.  

### What this is *not*

- Not a multi-tenant hosted Palace  
- Not a substitute for OS-level sandboxing or disk encryption  
- Not a cloud Memory SLA or freemium hosted Memory product  
- Not automatic authority over remote mesh data without operator credentials  

## Hardening checklist for operators

1. Point `BaseDir` at a directory with appropriate OS permissions; do not share roots across untrusted tenants. Kernel creates new palace dirs `0700` and writes files `0600` — confirm with `stat` on a fresh palace.  
2. Prefer the pure-Go hugot backend for CI/dev; only enable ORT/CUDA with trusted native libraries  
3. Download models only from sources you trust; pin paths via `MEMORY_ONNX_MODEL_PATH`  
4. Do not commit palace contents, `.env`, or API keys  
5. Scope Qdrant endpoints to private networks when used  
6. Treat residual docs under `docs/operations/` as process truth for shipped-vs-residual claims  

## Dependency security

```bash
make vuln   # govulncheck ./...
make test   # unit tests (no mandatory ONNX/LongMemEval)
```

CI runs tests, `go vet`, `gofmt` check, and `govulncheck` on every PR. Heavy ONNX / LongMemEval gates remain optional (`workflow_dispatch`).

## Disclosure preference

Coordinated disclosure: we prefer to ship a fix (or mitigating docs) before public write-ups when the issue is exploitable in default configurations.
