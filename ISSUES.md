# Issues

Open work in the fork `rophy/zot`. This file lives on `develop` and never goes upstream. Remove an
entry once it is merged upstream or dropped.

## Open issues

| # | Issue | Branch | Status |
| --- | --- | --- | --- |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`e03da4d8`) | done and tested, rebased on upstream `89465fe1`; reviewed 2026-09-30, status-lag fix and access docs added; waiting for maintainers on #4472 (labeled, no comments), then open the PR |
| 5 | `TestEventRecorderReload` flaky test fix | `feat/fix-events-reload-test` (`eaf568f4`) | done, passed in staging CI; upstream PR not opened |
| 6 | `TestScheduler` flaky test fix | `feat/fix-scheduler-fairness-test` (`b8dbe75f`) | **on hold until reproduced**: fix done, but the flake is too rare to justify a PR yet (see below) |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |

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
