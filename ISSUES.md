# Issues

Open work in the fork `rophy/zot`. This file lives on `develop` and never goes upstream. Remove an
entry once it is merged upstream or dropped.

## Open issues

| # | Issue | Branch | Status |
| --- | --- | --- | --- |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`547797e7`), upstream-ready `pr/gc-trigger-api` (`58b55512`, on upstream `main` `ae4958cd`), body in `../zot-gc/pr.md` | squashed; 3 CodeRabbit reviews on fork PR rophy/zot#1 (closed) all resolved; deleted counts split off to `feat/gc-status-deleted-counts` as a possible follow-up; rebased 2026-10-07 (fixed a nil `sweep` panic in upstream test from #4514), upstream PR #4522 opened 2026-10-07; coverage tests added 2026-10-11; rebased on main 2026-10-13 for the new codecov config (#4532); Copilot review fixes pushed 2026-10-16 (adminPolicy via AuthzOnlyAdminsMiddleware, CORS, StatIndex, test ports; shutdown race answered, not changed), replies drafted, not posted |
| 5 | `TestEventRecorderReload` flaky test fix | `feat/fix-events-reload-test` (`eaf568f4`) | done, passed in staging CI; upstream PR not opened |
| 6 | `TestScheduler` flaky test fix | `feat/fix-scheduler-fairness-test` (`b8dbe75f`) | **on hold until reproduced**: fix done, but the flake is too rare to justify a PR yet (see below) |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |
| 8 | Per-store storage usage metric, dedupe-aware (issue #4523) | none | proposal posted 2026-10-07 (local first, initial value from `PopulateStorageMetrics` walk, hook updates; S3 later); waiting for maintainers, no code yet |
| 9 | zot exits 0 when it can't start (#4636), and exits on SIGTERM without draining in-flight requests (#4637) | `fix/sigterm-drain` (`9549f362`), `pr/sigterm-drain` (`5c46a60c`, on upstream main fb6694ea) | issues posted 2026-10-31 (#4636 is a regression of #1392 via #1723); fix also stops printing usage on start errors; verified: cli/server tests pass with -race (minimal + extended, with the localstack S3 mock), lint clean, both new tests fail on upstream main, end to end (busy port exit 1, download and upload finish during SIGTERM); blackbox suites pass (2 pushpull tests fail on main too: host setup); PR #4638 opened 2026-10-31 |
| 10 | On-demand sync does not serve a cached tag until a slow upstream check fails | `fix/cached-tag-slow-upstream` (`6015d106`) | found 2026-10-17, reproduced on upstream main 7d668e24; fixed locally with tests (check bounded by responseHeaderTimeout when the tag is stored locally; end to end: 429 / hang / proxy hang served with 200 after 30 s); issue drafted in scratchpad `issue-cached-tag-slow-upstream.md`; nothing pushed or posted |
| 11 | GC prunes a new sparse index before its platform manifests are synced | `fix/gc-sparse-index-delay` (`4edf1683`) | found 2026-10-17 by exploratory tests, reproduced with main code (staging build prunes the tag after an index-only read; fix keeps it); issue drafted in scratchpad `issue-gc-sparse-index.md`; nothing pushed or posted |

`staging` (`ab29194e`) = `develop` (older) + issues 4, 5, 6. Its last CI run
(https://github.com/rophy/zot/actions/runs/36436324063): minimal suite passed; extended suite
passed everything except `pkg/extensions/search/cve`, which hit the bbolt deadlock since fixed
upstream (#4483). No staging image has been pushed yet (`ghcr.io/rophy/zot:staging-<short-sha>`, private by default).

Issues 4 and 6 both add an import to `pkg/scheduler/scheduler_test.go`; whichever merges upstream
second needs a trivial rebase.

## Issue 6: `TestScheduler` flake, on hold

The fairness check (`scheduler_test.go:258` on upstream `main`) reads the task log order, which
depends on worker timing; the fix checks the order the scheduler picked generators in instead.
Evidence collected 2026-09-29 (not enough for an upstream PR):

- Upstream CI (runners with 8/16 cores): 1 failure in ~1000 `test.yaml` runs since 2026-07-16,
  run 32722842702 (PR #4351, 2026-08-24). Searched all failed runs and failed first attempts of
  re-run runs; the other `pkg/scheduler` failures were `[build failed]`.
- Fork staging CI (4-core `ubuntu-latest`): 1 failure in 3 minimal runs (run 36420507713 attempt 2).
- Experiment branch `ci/scheduler-flake` (workflow `scheduler-flake.yaml`: minimal suite, 5 runs on
  `develop` + 5 with the fix): run 36581242896, 0/5 unfixed failures, 0/5 fixed.

Resume when it reproduces: re-run the experiment (`gh run rerun <id> --repo rophy/zot`) or push to
`ci/scheduler-flake` for more samples; each job reports `TestScheduler <variant> <n>: PASS|FAIL` in
the run summary. Then open a PR with the failure rate before and after the fix.
