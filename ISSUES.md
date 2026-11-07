# Issues

Open work in the fork `rophy/zot`. This file lives on `develop` and never goes upstream. Remove an
entry once it is merged upstream or dropped.

## Open issues

| # | Issue | Branch | Status |
| --- | --- | --- | --- |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`406d827c`), upstream-ready `pr/gc-trigger-api` (`50887144`, on upstream `main` `e167e126`), body in `../zot-gc/pr.md` | squashed; 3 CodeRabbit reviews on fork PR rophy/zot#1 (closed) all resolved; deleted counts split off to `feat/gc-status-deleted-counts` as a possible follow-up; rebased 2026-10-07 (fixed a nil `sweep` panic in upstream test from #4514), upstream PR #4522 opened 2026-10-07; coverage tests added 2026-10-11; rebased on main 2026-10-13 for the new codecov config (#4532); Copilot review fixes pushed 2026-10-16 (adminPolicy via AuthzOnlyAdminsMiddleware, CORS, StatIndex, test ports; shutdown race answered, not changed), replies drafted, not posted; rebased on v2.1.23 2026-11-02 (import conflict in `gc.go`); maintainers asked for authn/z review of mgmt endpoint and a rebase 2026-11-06, rebased on main `e167e126` 2026-11-07; staging CI found a test race in the SweepNow shutdown case, fixed 2026-11-07 (`test(gc)` commit) |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |
| 8 | Per-store storage usage metric, dedupe-aware (issue #4523) | none | proposal posted 2026-10-07 (local first, initial value from `PopulateStorageMetrics` walk, hook updates; S3 later); waiting for maintainers, no code yet |
| 11 | `manifestCheckInterval: -1` (never check stored tags) for on-demand sync (issue #4128) | `feat/sync-skip-upstream-if-local` (`7d0bbcf8`), upstream `pr/sync-skip-upstream-if-local` (`62032bfb`, on upstream `main` `e167e126`) | upstream PR #4646 opened 2026-11-01; Copilot finding (overlapping rules picked the wrong tag filters) fixed in 6b134968, thread resolved; overlaps #4612 (`ShouldCheckUpstream`, same files), whichever merges second needs a small rebase; #4128 marked stale 2026-10-31; 2026-11-07 reworked per rchincha on #4128: dropped `skipUpstreamIfLocal` and tag scope for a negative `manifestCheckInterval`, PR title/body updated |

`staging` = `develop` + issues 4, 11. An older CI run
(https://github.com/rophy/zot/actions/runs/36436324063): minimal suite passed; extended suite
passed everything except `pkg/extensions/search/cve`, which hit the bbolt deadlock since fixed
upstream (#4483). No staging image has been pushed yet (`ghcr.io/rophy/zot:staging-<short-sha>`, private by default).


