# Contributing

Thanks for helping improve **memory** (Palace kernel). Please treat quality, honesty locks, and tests as first-class.

## What this repo is

- **Local-primary memory kernel** — hierarchical Palace FS, hybrid search, temporal lite, optional Qdrant, optional ONNX embeddings  
- **Kernel-only** — **not product Memory GA**  
- Product path elsewhere: **dual_write OFF** by default; hosted Palace **sunset** until scale; mesh optional via public TUI / ops packs  
- Future public MCP host naming honesty: **`iomesh-memory-mcp`** (not `aion-memory-mcp`); private aion broker stays private  

This repository may remain **private** until a deliberate visibility flip. Do not assume public contribution workflows until that flip lands.

## Development setup

```bash
# Go version: see go.mod (CI uses that exact toolchain via GOTOOLCHAIN=auto)
git clone https://github.com/iome-sh/memory.git
cd memory
make test
make vet
make build
```

Optional:

```bash
make test-race
make cover
make vuln
make ci          # fmt-check + vet + test + vuln + build (local gate)
make residual-gate   # offline residual honesty pins (s1297 / s1303 / s1313)
```

Heavy optional gates (models / network; not required for PR CI):

```bash
make test-onnx
make longmemeval-recall-gate
make longmemeval-bench
```

## Coding standards

- Prefer **pure Go** for default CI paths; keep ORT/CGO behind `-tags ORT`  
- Do **not** invent product Memory GA, freemium hosted Palace, or dual_write-on product narrative in docs  
- Residual docs under `docs/operations/` are honesty pins — keep gates green when you touch related claims  
- Prefer small, focused PRs with tests for new behavior  
- Run `gofmt` (or `make fmt`) before commit  

## Tests

| Area | Focus |
|------|--------|
| PalaceStore / write / versioning | core FS kernel |
| SearchMemory / ListMemory / meta index | hybrid search + timeline |
| MultiHop / supersession / facts-as-of | temporal lite / graph lite |
| Embedding backends | hash default + hugot (optional ONNX) |
| VectorStore | Qdrant client; Podman optional |
| cmd/longmemeval-* | benchmark/server tools (optional heavy gates) |

New features should include unit tests. Network-heavy tests must soft-skip without keys/models. Do not require live Qdrant or ONNX for the default `make test` path.

## Security-sensitive changes

If you touch filesystem roots, embedding model loading, or network clients:

1. Add/adjust unit tests  
2. Update [SECURITY.md](SECURITY.md) if the threat model changes  
3. Prefer explicit path/URL validation and clear residual-risk docs  

Report vulnerabilities privately — see [SECURITY.md](SECURITY.md). **Do not open public issues for exploits.**

## Honesty locks (do not regress)

- Kernel-only · **not Memory GA**  
- Local-primary · dual_write OFF product path (elsewhere)  
- Hosted Palace sunset · mesh optional  
- Residual PASS ≠ live dogfood / invent GA  
- Future MCP host name: **iomesh-memory-mcp**  

## Pull requests

- Clear description of *what* and *why*  
- Link related issues when applicable  
- Ensure CI is green (`ci-success` when branch protection requires it)  
- Do not commit palace data, API keys, `.env`, or real user memory  
- Update [CHANGELOG.md](CHANGELOG.md) **Unreleased** for user-visible API changes  

### CI on PR and merge

GitHub Actions workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on:

| Event | When |
|-------|------|
| `pull_request` | opened / synchronize / reopened / ready_for_review → `main` |
| `push` | commits to `main` (after merge) |
| `merge_group` | GitHub merge queue (if enabled) |
| `workflow_dispatch` | manual re-run; optional LongMemEval/ONNX |

Jobs: **lint** · **test** · **build** · **govulncheck** · **ci-success** (aggregate). Optional **longmemeval** only on dispatch.

Local parity:

```bash
make ci
```

## License

By contributing, you agree that your contributions are licensed under the MIT License (see [LICENSE](LICENSE)).
