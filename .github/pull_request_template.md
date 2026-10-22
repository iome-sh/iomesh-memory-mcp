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
- [ ] Docs describe only what the host implements

## Security checklist (if touching FS roots, HTTP, transports)

- [ ] Residual risks still accurate in `SECURITY.md`
- [ ] No new network listeners or remote writes enabled by default
- [ ] Errors/logs do not dump user palace contents in tests
