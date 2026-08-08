# Releasing

Ship from `main` via PR; cut annotated semver tags for Go module consumers.

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

## Artifacts

| Path | Notes |
|------|--------|
| Go module tag | Primary consumer path |
| `go get` | `go get github.com/iome-sh/memory@vX.Y.Z` |
| CI `build` job | packages + longmemeval cmd binaries smoke-build |

```bash
make build
go get github.com/iome-sh/memory@v1.5.7
```

## Honesty

Tags describe the **kernel library**, not product Memory GA. Hosted Palace sunset and dual_write product defaults live outside this module. Residual honesty pins may land after a tag without inventing GA.
