# Support

## How to get help

| Need | Where |
|------|--------|
| Usage questions / bugs | [GitHub Issues](https://github.com/iome-sh/memory/issues) (use the templates when the repo is open to contributors) |
| Security vulnerability | Private [Security Advisory](https://github.com/iome-sh/memory/security/advisories/new) or **security@iome.sh** — see [SECURITY.md](SECURITY.md) |
| Kernel API / roadmap | [README.md](README.md), [docs/temporal-memory-kernel-roadmap.md](docs/temporal-memory-kernel-roadmap.md), [docs/operations/](docs/operations/) residual honesty pins |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## What we maintain

- **Best effort** on `main` and the latest `v1.5.x` tag line  
- Security fixes on the default branch when feasible  
- **No cloud Memory SLA** — this is a **library / local-primary kernel**, not a hosted Palace product  

## What we do not provide here

- Hosted Palace / multitenant cloud Memory onboarding or uptime guarantees  
- Product Memory GA install support (edge install honesty lives in the public TUI and future `iomesh-memory-mcp` host)  
- Guarantees about third-party Qdrant, ONNX Runtime, or model-hub availability  
- Private monorepo (`aion`) broker / control-plane support via this package  

## Before filing an issue

1. Run `make check` or note CI failures  
2. Redact API keys, palace contents, and private paths from logs  
3. Include module version (`go list -m github.com/iome-sh/memory`) or commit SHA and OS  
4. Confirm the report is about the **kernel API** — not a request to invent hosted Memory GA  
