## Summary

<!-- What and why (1–3 bullets). -->

-

## Type of change

- [ ] Feature
- [ ] Bug fix
- [ ] Security / hardening
- [ ] Docs / CI
- [ ] Refactor (no behavior change)

## Test plan

- [ ] `make check` or `make ci` (or CI green: lint, test, build, govulncheck)
- [ ] New/changed behavior covered by unit tests
- [ ] No secrets / palace data in tree
- [ ] Honesty locks intact (kernel-only · not Memory GA) if docs touch product narrative

## Security checklist (if touching FS roots, embeddings, network clients)

- [ ] Residual risks still accurate in `SECURITY.md`
- [ ] Model paths / URLs not hard-coded to private monorepo assets
- [ ] Errors/logs do not dump user palace contents in tests
