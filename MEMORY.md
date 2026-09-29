# Memory

Context carried between sessions. Fork: `rophy/zot`; this file lives on `develop` and never goes
upstream. Upstream: `project-zot/zot`. Branch rules and commit policy: see `CLAUDE.md`.

## Current tasks

| # | Task | Branch | Status |
| --- | --- | --- | --- |
| 4 | GC on demand API (issue #4472) | `feat/gc-trigger-api` (`484d3645`) | done and tested; upstream PR not opened |
| 5 | `TestEventRecorderReload` flaky test fix | `feat/fix-events-reload-test` (`eaf568f4`) | done, passed in staging CI; upstream PR not opened |
| 6 | `TestScheduler` flaky test fix | `feat/fix-scheduler-fairness-test` (`b8dbe75f`) | done, passed in staging CI; upstream PR not opened |
| 7 | Leaked controller in `TestDerivedImageListGqlAuthorization` | none | noticed only: its htpasswd watcher keeps polling a deleted file during later tests |

`staging` (`ab29194e`) = `develop` (older) + tasks 4, 5, 6. Its last CI run
(https://github.com/rophy/zot/actions/runs/36436324063): minimal suite passed; extended suite
passed everything except `pkg/extensions/search/cve`, which hit the bbolt deadlock since fixed
upstream (#4483). No staging
image has been pushed yet (`ghcr.io/rophy/zot:staging-<short-sha>`, private by default).

## Notes for upstream PRs

- Build `pr/<name>` per `CLAUDE.md`; the user signs commits off (DCO) and opens the PRs.
- Before creating any upstream issue or PR, check for existing ones:
  `gh issue list` / `gh pr list --repo project-zot/zot --author rophy --state all`, and whether the
  head branch already has a PR. Several sessions work in parallel; duplicates happened (#4484, #4485).
- Merged upstream: bcrypt test cost (#4481), bbolt deadlock in the CVE scan generator (#4482/#4483).
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
