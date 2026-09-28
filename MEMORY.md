# Memory

Context carried between sessions. Fork: `rophy/zot` (this file lives on `develop`, never goes
upstream). Upstream: `project-zot/zot`. Branch rules: see `CLAUDE.md`.

## Next: upstream issue about slow tests from bcrypt cost

**Goal:** file an issue on `project-zot/zot` showing that the test htpasswd credentials' bcrypt
cost makes the auth-heavy tests very slow, with proof from upstream's own CI, then offer the fix
(branch `fix/test-bcrypt-cost`, one line in `pkg/test/common/fs.go`).

**Why a separate session:** reading upstream CI job logs needs the GitHub API for
`project-zot/zot`, which a session only gets with that repository attached. A session can't hold
both `rophy/zot` and `project-zot/zot` (both check out to `<cwd>/zot`), so the upstream work
runs in a session started with `project-zot/zot`. Read this file there from
https://raw.githubusercontent.com/rophy/zot/develop/MEMORY.md.

**Cause (verified):**

- `test.GetBcryptCredString` (`pkg/test/common/fs.go`) hashes test passwords with bcrypt cost 10.
- htpasswd verifies the hash on every authenticated request, with no cache
  (`pkg/api/htpasswd.go`, `bcrypt.CompareHashAndPassword` in `HTPasswd.Authenticate`).
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
(4-core `ubuntu-latest`) `pkg/extensions/search` took 569 s and `pkg/api` 1341 s.
25 test files use the helper, in `pkg/api`, `pkg/extensions/...`, `pkg/cli/...`, `pkg/log`,
`pkg/debug/pprof`, `pkg/test/...`, so the saving goes beyond `search`.

Reproduce: run `go test -race -json -run '^TestUserData$'` with the `make test-extended` tags on
`./pkg/extensions/search/`, before and after changing the cost in `GetBcryptCredString`.

**What to collect in the upstream session:**

1. From several recent successful `Running tests` runs on `main` (workflow `test.yaml`), the log
   of the job "Run zot with extensions tests" (and "Running zot without extensions tests"):
   the per-package `ok zotregistry.dev/zot/v2/<pkg> <seconds>s` lines, mainly
   `pkg/extensions/search`, `pkg/api`, `pkg/cli/server`, `pkg/extensions`. `make test-extended`
   does not use `-v`, so upstream logs only have per-package times; per-test proof comes from the
   measurements above.
2. Known so far from the public run pages: `Running tests` takes 26-33 min per run on `main`;
   run 34047078040: "Run zot with extensions tests" 24m58s, "Running zot without extensions
   tests" 15m24s.
3. Draft the issue: problem, cause, the numbers above plus upstream's package times, the
   one-line fix. Check for an existing issue first.

## Other open items

- **Deadlock** (`fix/boltdb-nested-tx-deadlock`, details in its `ISSUES.md`): nested bbolt read
  transaction in the CVE scan task generator deadlocks with a write that remaps the DB. Real
  production bug, not reported upstream yet, fix not written. Drop `ISSUES.md` from its `pr/`.
- **Staging** needs a rebuild with `fix/test-bcrypt-cost` to confirm the fix in CI (the root-owned
  sandbox can't run the permission-based tests). The deadlock above can still fail a run.
- **Upstream PRs not opened yet:** `feat/gc-trigger-api` (issue #4472),
  `feat/fix-events-reload-test`, `feat/fix-scheduler-fairness-test`, `fix/test-bcrypt-cost`.
  The GC and scheduler-test branches both add an import to `pkg/scheduler/scheduler_test.go`.
- Minor: `TestDerivedImageListGqlAuthorization` leaves its controller running (its htpasswd
  watcher keeps polling a deleted file during later tests).
