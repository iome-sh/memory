# Releasing

Ship from `main` via PR; cut annotated semver tags for Go module consumers.

Public module: `github.com/iome-sh/memory`  
Consumers: `go get github.com/iome-sh/memory@vX.Y.Z` (module proxy). This is a **library**, not product Memory GA.

## When to bump and tag

**Do not leave user-visible API waves only under `[Unreleased]`.** After merging a coherent capability set, cut a release in the same delivery loop (or immediately after):

| Trigger | Bump | Examples |
|---------|------|----------|
| New exported API surface (search/list/multihop/supersession/…) | **minor** or **patch** within `v1.5.x` as appropriate | `ListFactsAsOf`, `MultiHopRetrieve` options |
| Breaking exported API (document clearly) | **minor** (pre-stability policy) or **major** | Rename/remove public funcs |
| Docs-only / residual honesty pins | usually **no** tag | Ops residual gates, OSS process docs |
| Security fix on latest line | **patch** (`v1.5.y`) | CVE follow-up |

Checklist items that must move with the tag:

- [CHANGELOG.md](CHANGELOG.md) `[Unreleased]` → `## [X.Y.Z]`  
- [README.md](README.md) release notes / install line  
- [SECURITY.md](SECURITY.md) supported-versions table if the minor line changes  

## Checklist before a tag

1. [ ] `make ci` green locally  
2. [ ] GitHub Actions **ci-success** green on the release commit  
3. [ ] [CHANGELOG.md](CHANGELOG.md) updated  
4. [ ] No secrets or palace data in tree  
5. [ ] Honesty locks intact: kernel-only · not Memory GA · dual_write OFF product path  
6. [ ] Annotated tag `vX.Y.Z` pushed  

## Tag and publish (maintainers)

```bash
git checkout main
git pull origin main
# edit CHANGELOG.md + README if needed
git commit -am "chore: release vX.Y.Z"
git push origin main

git tag -a vX.Y.Z -m "vX.Y.Z — short release title"
git push origin vX.Y.Z

# Optional GitHub release notes from CHANGELOG
# gh release create vX.Y.Z --notes-file …
```

## Versioning policy

- **1.x** — SemVer for the Go module path `github.com/iome-sh/memory`  
- Prefer **patch/minor** within the current `v1.5.x` line for additive temporal/kernel APIs  
- Document breaking changes explicitly in CHANGELOG  

## Support / version policy

Clear support window for **`github.com/iome-sh/memory`** so integrators know what is maintained. Support docs ≠ product Memory GA and ≠ forever-green release machinery.

### What is supported

| Surface | Policy |
|---------|--------|
| **Latest tagged release** | Latest annotated `vX.Y.Z` on `main` is the primary supported module version |
| **Current minor line** | Security and critical fixes land on the **current** minor line (today: `v1.5.x`) when feasible |
| **`main` tip** | Development tip; CI-gated, but **not** a production pin — may move without a tag |
| **Older minor lines** | Best-effort only until an EOL note appears in [SECURITY.md](SECURITY.md) |

See [SECURITY.md](SECURITY.md) **Supported versions** for the security-fix matrix maintainers keep current when publishing a new minor.

### Production pin (consumers)

Prefer **semver tags** in production `go.mod` — do **not** rely on floating `@main` alone:

```bash
# Preferred: pin a release tag
go get github.com/iome-sh/memory@vX.Y.Z

# After tidy, go.mod / go.sum lock the module graph
go mod tidy
```

| Practice | Why |
|----------|-----|
| Pin `vX.Y.Z` (or a known commit SHA) | Reproducible builds; known support surface |
| Avoid production-only `@main` / `@latest` without a follow-up pin | Tip moves; not a support contract |
| Record `go.sum` hashes in VCS | Integrity via the public module proxy checksum DB |

### Verification (library path — not cosign binaries)

This package is a **Go library**. Consumers verify via the **standard Go toolchain**, not binary cosign:

```bash
go get github.com/iome-sh/memory@vX.Y.Z
go mod download github.com/iome-sh/memory@vX.Y.Z
# go.sum records h1:… hashes; sum.golang.org attests public module contents when using the default proxy
```

- **No `GOPRIVATE` / PAT** required for this public module.  
- Cosign / SBOM / GoReleaser apply to **binary peers** (`iomesh-memory-mcp`, `iomesh-tui`) — not to this library surface. See [Signing / release matrix](#signing--release-matrix) below.  
- Host release matrix for the MCP binary lives in the peer’s `RELEASING.md` (mention only; do not invent peer forever-green from this tip).

### Security-supported versions

When cutting a **new minor** (or deliberately EOL-ing an old line):

1. Update [SECURITY.md](SECURITY.md) **Supported versions** table  
2. Prefer security patches on the **current** minor line (`vX.Y.z`)  
3. Document any EOL of prior minors in CHANGELOG + SECURITY  

Reporting: private Security Advisory or **security@iome.sh** — see [SECURITY.md](SECURITY.md). Day-to-day usage questions: [SUPPORT.md](SUPPORT.md) / GitHub Issues.

### What this policy does **not** mean

| Non-claim | Honesty |
|-----------|---------|
| docs ≠ invent **Memory GA** | Kernel library support ≠ sold “Memory GA” |
| docs ≠ invent **forever-green signed releases** | Tags + `go get` / `go.sum` / proxy; no invent cosign-on-library or always-green peer GoReleaser |
| docs ≠ invent **dual_write ON** | dual_write **OFF** on the product path elsewhere |
| docs ≠ invent **live dogfood** | No invent operator dogfood green from docs |
| Docs-only change | **No release tag invented** by documentation |

## Artifacts

| Path | Notes |
|------|--------|
| Go module tag | **Primary consumer path** |
| `go get` | `go get github.com/iome-sh/memory@vX.Y.Z` |
| CI `build` job | packages + longmemeval cmd binaries smoke-build |

```bash
make build
go get github.com/iome-sh/memory@v1.5.7
```

### Library module (no GoReleaser required)

This package is a **Go library module**. The primary release artifact is an **annotated git tag** plus consumers running `go get github.com/iome-sh/memory@vX.Y.Z`.

- **GoReleaser is not required** here (unlike binary products **iomesh-tui** / **iomesh-memory-mcp**, which ship cross-built archives).  
- Optional: `gh release create vX.Y.Z` (or `gh release edit`) for human-readable notes from CHANGELOG — notes only, not binary assets.  
- Optional longmemeval / helper cmd binaries may be built in CI for smoke; they are not the module’s public product surface.

## Signing / release matrix

How library consumers verify releases, and how this module peers with binary hosts. **Docs ≠ invent cosign on this library** · **≠ invent Memory GA**.

### Matrix (this module vs binary peers)

| Concern | **memory** (this repo · Go library) | Binary peers (**iomesh-tui**, **iomesh-memory-mcp**) — *mention only* |
|---------|-------------------------------------|------------------------------------------------------------------------|
| Primary artifact | Annotated semver **git tag** (`vX.Y.Z`) on `main` | GitHub Release archives from **GoReleaser** on `v*` tags |
| Consumer install | `go get github.com/iome-sh/memory@vX.Y.Z` (or `@main`) | `go install …@vX.Y.Z` and/or downloaded release binaries |
| Integrity / “signing” | Go module graph: **`go.sum`** + public module **proxy checksum database** (sum.golang.org) when fetching via the default proxy | **checksums.txt** + per-archive **SPDX SBOM** + **keyless cosign** (`cosign sign-blob` via GitHub OIDC / Fulcio) on tag releases |
| Cosign on this package | **N/A** — no library binary to sign; do **not** invent cosign green here | Documented in each peer’s `RELEASING.md` (do **not** invent those releases forever-green from this residual) |
| SBOM / GoReleaser | **N/A** for the module surface | Peer GoReleaser + syft SBOM (peers own greenness) |
| Optional notes-only release | `gh release create` from CHANGELOG (no binary assets required) | GoReleaser publishes assets; notes optional |
| Visibility / license | **Public** MIT | Peers public MIT |

### Consumer verification (library path)

Operators and integrators trust module versions through the **standard Go toolchain**, not cosign:

```bash
# Pin a release tag (preferred for production go.mod)
go get github.com/iome-sh/memory@vX.Y.Z

# Inspect the module graph / locked hashes after tidy
go mod download github.com/iome-sh/memory@vX.Y.Z
# go.sum records h1:… hashes; proxy + checksum DB attest public module contents
```

- **No `GOPRIVATE` / PAT** required for this public module.  
- Prefer **semver tags** over floating `@main` in production `go.mod`.  
- CI on this repo (`make ci` / GitHub Actions **ci-success**) gates what lands on `main` before maintainers cut tags — it is **not** a substitute for product Memory GA or signed forever-green binary releases.

### Peer packaging (mention only — not invent green)

Binary edge hosts that *consume* this library ship their **own** release packaging:

- **iomesh-tui** / **iomesh-memory-mcp**: GoReleaser cross-builds, `checksums.txt`, SPDX SBOM, keyless cosign on tag releases (see each peer’s `RELEASING.md`).  
- This page **mentions** that path only. It does **not** claim peer release workflows are forever green, that cosign verify always succeeds in every environment, or that GHCR images are published.

### What this matrix does **not** mean

| Non-claim | Honesty |
|-----------|---------|
| docs ≠ invent **cosign on the library** | Library trust = tag + `go get` / `go.sum` / proxy checksum DB |
| docs ≠ invent **signed forever-green releases** | Neither this module’s tags nor peer GoReleaser runs are invented green |
| docs ≠ invent **Memory GA** | Kernel library only |
| docs ≠ invent **dual_write ON** | dual_write **OFF** on the product path elsewhere |
| Public MIT | Visibility already **public**; docs ≠ re-flip |

## Honesty

Tags describe the **kernel library**, not product Memory GA. This is a local filesystem library. Hosted Palace sunset and dual_write product defaults live outside this module. **Public MIT** · dual_write **OFF** product path · open boxes stay open · docs invent **no** release tag.
