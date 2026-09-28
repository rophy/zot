# Memory

Context carried between sessions. Fork: `rophy/zot`; this file lives on `develop` and never goes
upstream. Upstream: `project-zot/zot`. Branch rules and commit policy: see `CLAUDE.md`.

## Current tasks

| # | Task | Branch | Status |
| --- | --- | --- | --- |
| 1 | Upstream issue: bcrypt cost makes tests slow | `fix/test-bcrypt-cost` (`f77e51aa`) | fix done; collect upstream CI proof, draft issue (next) |
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

**Upstream CI logs need the GitHub API for `project-zot/zot`**, which a session only gets with that
repository attached. A session can't hold both `rophy/zot` and `project-zot/zot` (both check out
to `<cwd>/zot`), so this part runs in a session started with `project-zot/zot`; it reads this file
from https://raw.githubusercontent.com/rophy/zot/develop/MEMORY.md.

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

**Still to do:**

1. From several recent successful `Running tests` runs on `main` (workflow `test.yaml`), get the
   log of the job "Run zot with extensions tests" (and "Running zot without extensions tests") and
   extract the per-package `ok zotregistry.dev/zot/v2/<pkg> <seconds>s` lines, mainly
   `pkg/extensions/search`, `pkg/api`, `pkg/cli/server`, `pkg/extensions`. `make test-extended`
   does not use `-v`, so upstream logs only have per-package times; per-test proof comes from the
   measurements above. Known from the public run pages: `Running tests` takes 26-33 min per run on
   `main`; run 34047078040: "Run zot with extensions tests" 24m58s, "Running zot without
   extensions tests" 15m24s.
2. Check upstream for an existing issue, then draft the issue: problem, cause, numbers, one-line fix.

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
