# Issues

Open work in the fork `rophy/zot`. This file lives on `develop` and never goes upstream. Remove an
entry once it is merged upstream or dropped.

## Open issues

| # | Issue | Branch | Status |
| --- | --- | --- | --- |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |
| 8 | Per-store storage usage metric, dedupe-aware (issue #4523) | none | proposal posted 2026-10-07 (local first, initial value from `PopulateStorageMetrics` walk, hook updates; S3 later); waiting for maintainers, no code yet |
| 12 | GC status reports deleted counts (follow-up to merged #4522) | `feat/gc-status-deleted-counts` (`501aec64`) | possible follow-up, split off from #4522; based on the old `feat/gc-trigger-api`, needs a rebuild on `develop` before any PR |

`staging` = `develop` only: no pending `feat/` or `fix/` branch.
