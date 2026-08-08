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
- [ ] Honesty locks intact (edge host · dual_write OFF · not Memory GA) if docs touch product narrative

## Security checklist (if touching FS roots, HTTP, transports)

- [ ] Residual risks still accurate in `SECURITY.md`
- [ ] No aion broker / dual_write default ON
- [ ] Errors/logs do not dump user palace contents in tests
