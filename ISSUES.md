# Issues

Open work in the fork `rophy/zot`. This file lives on `develop` and never goes upstream. Remove an
entry once it is merged upstream or dropped. Handover notes: `MEMORY.md`.

## Open issues

| # | Issue | Branch | Status |
| --- | --- | --- | --- |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`484d3645`) | done and tested; upstream PR not opened |
| 5 | `TestEventRecorderReload` flaky test fix | `feat/fix-events-reload-test` (`eaf568f4`) | done, passed in staging CI; upstream PR not opened |
| 6 | `TestScheduler` flaky test fix | `feat/fix-scheduler-fairness-test` (`b8dbe75f`) | done, passed in staging CI; upstream PR not opened |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |

`staging` (`ab29194e`) = `develop` (older) + issues 4, 5, 6. Its last CI run
(https://github.com/rophy/zot/actions/runs/36436324063): minimal suite passed; extended suite
passed everything except `pkg/extensions/search/cve`, which hit the bbolt deadlock since fixed
upstream (#4483). No staging image has been pushed yet (`ghcr.io/rophy/zot:staging-<short-sha>`, private by default).
