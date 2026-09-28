## Git Commit Policy

**Commit message format:**

```
<type>: <short description>

[optional body explaining why/what changed]
```

**RULES:**

- NO "Generated with Claude Code" footer
- NO "Co-Authored-By: Claude" line
- NO mention of "Claude" or "Happy" anywhere
- Keep messages short (1-5 lines preferred)
- Types: feat, fix, refactor, chore, docs, build, test

## Branch Workflow

| Branch | Contents | History |
| --- | --- | --- |
| `main` | matches upstream | fast-forward only |
| `develop` | `main` + personal baseline (this file, etc.), never goes upstream | merge `main` in |
| `feat/<name>` | `develop` + one feature, for upstream | normal commits |
| `fix/<name>` | `develop` + one bug fix, for upstream | normal commits |
| `base/<name>` | `develop` + a change to the personal baseline, never goes upstream | merged back into `develop` |
| `pr/<name>` | `main` + one `feat/` or `fix/` branch, for the upstream PR | rebuilt for each PR |
| `staging` | `develop` + all pending `feat/*` and `fix/*` | rebuilt, force-pushed |

`feat/` and `fix/` branches are handled the same way; below, `<type>` is `feat` or `fix`.

**Before working:**

```bash
git fetch origin main develop
git checkout develop
git merge origin/main            # merge, never rebase develop
git push origin develop
git checkout -b <type>/<name>    # base/<name> for a baseline change
```

**When completed, before creating a PR:** build a `pr/<name>` branch with only the `<type>/<name>`
commits on top of `main`, excluding the develop-only commits:

```bash
git fetch origin main
git checkout -b pr/<name> <type>/<name>
git rebase --onto origin/main develop pr/<name>
git diff --stat origin/main pr/<name>   # must not include CLAUDE.md or other develop-only files
git push -u origin pr/<name>
```

**Staging:** `staging` consolidates all pending `feat/` and `fix/` branches. Rules:

- Base it on `develop`, and merge only `feat/*` and `fix/*` branches into it, never `pr/*` or `base/*`.
- Never branch from `staging`: new branches start from `develop`.
- Rebuild it instead of accumulating history, whenever a branch is added, merged upstream,
  dropped, or rebased.
- Resolve conflicts between branches while merging into `staging`, not in the branches themselves.

```bash
git fetch origin
git checkout -B staging origin/develop
for b in origin/feat/<a> origin/fix/<b>; do    # every pending feat/ and fix/ branch, one at a time
  git merge --no-ff "$b"                        # resolve conflicts here, then continue
done
git push --force-with-lease origin staging
```

Pushing `staging` runs `.github/workflows/staging.yaml`: the unit tests of `make test-minimal`
and `make test-extended` (without coverage, with a 60m timeout), then the full image for linux/amd64 and linux/arm64, pushed as
`ghcr.io/rophy/zot:staging-<short-sha>`. Upstream's own workflows are disabled in this fork.

## Commit Author and Sign-off

Author every commit as the user, and sign it off (DCO) with `-s`:

```bash
git config user.name "Rophy Tsai"
git config user.email "rophy@users.noreply.github.com"
git commit -s
```
