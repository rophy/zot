# Memory

Context carried between sessions. Fork: `rophy/zot`; this file lives on `develop` and never goes
upstream. Upstream: `project-zot/zot`. Branch rules and commit policy: see `CLAUDE.md`.

## Current tasks

| # | Task | Branch | Status |
| --- | --- | --- | --- |
| 1 | Upstream issue: bcrypt cost makes tests slow | `fix/test-bcrypt-cost` (`f77e51aa`) | fix done; upstream CI proof collected; measure `pkg/api` before/after (next), draft issue |
| 2 | Confirm the bcrypt fix in staging CI | `staging` | not started: rebuild staging with it |
| 3 | bbolt deadlock in the CVE scan task generator | `fix/boltdb-nested-tx-deadlock` (`02313f7f`) | documented in its `ISSUES.md`; fix and regression test not written; not reported upstream |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`484d3645`) | done and tested; upstream PR not opened |
| 5 | `TestEventRecorderReload` flaky test fix | `feat/fix-events-reload-test` (`eaf568f4`) | done, passed in staging CI; upstream PR not opened |
| 6 | `TestScheduler` flaky test fix | `feat/fix-scheduler-fairness-test` (`b8dbe75f`) | done, passed in staging CI; upstream PR not opened |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |

`staging` (`ab29194e`) = `develop` (older) + tasks 4, 5, 6. Its last CI run
(https://github.com/rophy/zot/actions/runs/36436324063): minimal suite passed; extended suite
passed everything except `pkg/extensions/search/cve`, which hit the task 3 deadlock. No staging
image has been pushed yet (`ghcr.io/rophy/zot:staging-<short-sha>`, private by default).

## Task 1: upstream issue about slow tests from bcrypt cost

**Goal:** file an issue on `project-zot/zot` showing that the test htpasswd credentials' bcrypt
cost makes the auth-heavy tests very slow, with proof from upstream's own CI, then offer the fix
(one line in `pkg/test/common/fs.go`: cost 10 → `bcrypt.MinCost`).

**Upstream CI logs:** `gh` (logged in as `rophy`, on the user's machine) reads them with
`gh api repos/project-zot/zot/actions/jobs/<job-id>/logs`. Cloud sessions without
`project-zot/zot` attached can't; they read this file from
https://raw.githubusercontent.com/rophy/zot/develop/MEMORY.md.

**Cause (verified):**

- `test.GetBcryptCredString` (`pkg/test/common/fs.go`) hashes test passwords with bcrypt cost 10.
- htpasswd verifies the hash on every authenticated request, with no cache
  (`HTPasswd.Authenticate`, `pkg/api/htpasswd.go`).
- The tests run with `-race`, which makes bcrypt about 14x slower.

**Evidence already collected (upstream code, `main` @ `4f79e998`):**

| Measurement | cost 10 | `bcrypt.MinCost` |
| --- | --- | --- |
| one bcrypt cost-10 compare, normal build | 69 ms | |
| one bcrypt cost-10 compare, `-race` | 998 ms | |
| `TestUserData` (`pkg/extensions/search`) | 184.7 s | 6.2 s |
| 7 slowest `pkg/extensions/search` tests together | 392 s | 12 s |
| whole `pkg/extensions/search` package | 505 s | 87.5 s |

Measured locally with the `make test-extended` build tags and `-race`. In the fork's CI
(4-core `ubuntu-latest`) `pkg/extensions/search` took 569 s and `pkg/api` 1341 s. 25 test files
use the helper (`pkg/api`, `pkg/extensions/...`, `pkg/cli/...`, `pkg/log`, `pkg/debug/pprof`,
`pkg/test/...`), so the saving goes beyond `search`. Failures seen when running these packages
locally also happen without the change (root sandbox, see below).

Reproduce: run `go test -race -json -run '^TestUserData$'` with the `make test-extended` tags on
`./pkg/extensions/search/`, before and after changing the cost in `GetBcryptCredString`.

**Upstream CI evidence (collected 2026-09-29):** the 8 latest successful `test.yaml` push runs
on `main`, job "Run zot with extensions tests". It runs
`go test -failfast -tags <extensions> -trimpath -race -timeout 20m -cover ... ./...` (Makefile
`test-extended`), without `-v`, so the logs only give per-package times (seconds):

| Run / job | Date | Commit | Job | `pkg/api` | `pkg/extensions/search` | `pkg/cli/server` | `.../search/cve` | `pkg/extensions/sync` |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| [36352277786](https://github.com/project-zot/zot/actions/runs/36352277786/job/108717836448) | 2026-09-27 | `4f79e998` | 24m21s | 1084 | 530 | 459 | 393 | 379 |
| [36349169550](https://github.com/project-zot/zot/actions/runs/36349169550/job/108704328777) | 2026-09-27 | `daf3f4dc` | 24m3s | 1073 | 532 | 441 | 384 | 372 |
| [36332986476](https://github.com/project-zot/zot/actions/runs/36332986476/job/108658375486) | 2026-09-27 | `7daa9281` | 24m12s | 1095 | 525 | 480 | 397 | 404 |
| [36319801875](https://github.com/project-zot/zot/actions/runs/36319801875/job/108621321169) | 2026-09-27 | `dc319658` | 24m5s | 1077 | 538 | 444 | 379 | 389 |
| [36304865840](https://github.com/project-zot/zot/actions/runs/36304865840/job/108579384090) | 2026-09-27 | `325356d8` | 24m51s | 1100 | 543 | 461 | 401 | 400 |
| [36221655026](https://github.com/project-zot/zot/actions/runs/36221655026/job/108347930269) | 2026-09-26 | `177fc0a7` | 24m3s | 1086 | 534 | 444 | 392 | 363 |
| [36195188960](https://github.com/project-zot/zot/actions/runs/36195188960/job/108269305210) | 2026-09-25 | `f4bf01c0` | 24m58s | 1087 | 535 | 467 | 402 | 405 |
| [36181271579](https://github.com/project-zot/zot/actions/runs/36181271579/job/108223960862) | 2026-09-25 | `2d721b46` | 24m15s | 1099 | 539 | 454 | 398 | 403 |

- `pkg/api` takes ~1086 s against the 20m (1200 s) per-package timeout: ~90% of the budget. The
  timeout was last raised 15m → 20m in `7642e5af` (Dec 2023). In the job without extensions
  (`-timeout 12m`), `pkg/api` takes only ~130 s.
- The slowest package, `pkg/extensions/search/cve/trivy` (~1125 s, once timed out at 20m in run
  35320388162), does not use the helper: not a bcrypt case, don't claim it in the issue.
- Other slow packages without the helper: `pkg/storage/gc`, `pkg/storage`, `pkg/storage/s3`,
  `pkg/extensions/imagetrust`, `pkg/exporter/api`.
- No "test timed out" failure in a helper-using package among ~45 failed extension/minimal jobs
  checked (failed `test.yaml` runs 2026-09-17 to 09-27, mostly PRs).
- No existing upstream issue or PR about bcrypt cost or slow tests (searched 2026-09-29).

**`pkg/api` code involved (upstream `main` @ `4f79e998`):**

- Hash: `pkg/test/common/fs.go:243` `bcrypt.GenerateFromPassword(pw, 10)`, once per helper call
  (~1 s under `-race`). **This is the line to tune** (→ `bcrypt.MinCost`).
- Compare: `pkg/api/authn.go:202` `basicAuthn` → `HTPasswd.Authenticate` →
  `pkg/api/htpasswd.go:111` `bcrypt.CompareHashAndPassword`, on every basic-auth request, no cache
  (~1 s under `-race`). This dominates: tests send dozens to hundreds of authenticated requests.
- Helper call sites in `pkg/api`: `controller_test.go` ~60 (build tags `sync && scrub && metrics
  && search && lint && userprefs && mgmt && imagetrust && ui`, so it only runs in the extensions
  job; this explains 1086 s vs 130 s), `htpasswd_test.go` ~30 (no tags), `authn_test.go` 4
  (`mgmt`), `mtls_test.go` 3, `routes_test.go` 1. Loops that hash several credential strings:
  `controller_test.go:922-929`, `:979`, `:7856`.

**Still to do:**

1. Measure the whole `pkg/api` package before/after `MinCost` (cloud session, ~20+ min per run):
   `env GOEXPERIMENT=jsonv2 go test -tags events,imagetrust,lint,metrics,mgmt,profile,scrub,search,sync,ui,userprefs -trimpath -race -timeout 60m ./pkg/api/`
   with `-json` for per-test times, on `main` and on `fix/test-bcrypt-cost`. Record here.
2. Draft the issue: problem, cause, upstream CI numbers above, local before/after, one-line fix.

## Task 3: deadlock notes

`BoltDB.FilterTags` calls `filterFunc` inside its `DB.View`; the CVE scan task generator's filter
calls `GetImageMeta` (nested `DB.View`) through `IsResultCached` (added in #4410) and
`IsImageFormatScannable` (older). A concurrent write that grows the DB file needs bbolt's mmap lock
exclusively, which waits for the outer read transaction, while the nested one waits for the mmap
lock. Full write-up, stacks and fix options in `ISSUES.md` on the branch. Drop `ISSUES.md` from
the branch's `pr/`.

## Notes for upstream PRs

- Build `pr/<name>` per `CLAUDE.md`; the user signs commits off (DCO) and opens the PRs.
- The GC and scheduler-test branches both add an import to `pkg/scheduler/scheduler_test.go`;
  whichever merges upstream second needs a trivial rebase.

## Environment notes (cloud sessions)

- The sandbox runs as root: permission-based tests fail locally (`pkg/storage/gc` sync staging,
  `TestCopyFiles`, `TestLogErrors`, `TestAPIKeysOpenDBError`, `TestCookiestoreCleanup` panic, ...)
  but pass in CI. Compare against `develop` before blaming a change.
- The Makefile exports `GOEXPERIMENT=jsonv2`; install tools with
  `env -u GOEXPERIMENT GOTOOLCHAIN=go1.27.0 go install ...` (e.g. golangci-lint `v2.13.2`,
  swag `v1.16.6`, actionlint). The preinstalled golangci-lint is too old for Go 1.27.
- `make testdata-images` needs skopeo (`apt-get install -y skopeo` works).
- Docker builds need the proxy CA inside the build (see `/root/.ccr/README.md`).
- This session's GitHub access can't re-run or dispatch workflows, or delete branches (403); the
  user does those.
- Upstream workflows are disabled in the fork; `staging.yaml` is the only one that runs.
