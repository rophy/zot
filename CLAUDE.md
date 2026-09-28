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
| `feat/<name>` | `develop` + one feature | normal commits |
| `pr/<name>` | `main` + one feature, for the upstream PR | rebuilt for each PR |
| `staging` | `develop` + all pending `feat/*` | rebuilt, force-pushed |

**Before working:**

```bash
git fetch origin main develop
git checkout develop
git merge origin/main            # merge, never rebase develop
git push origin develop
git checkout -b feat/<name>
```

**When completed, before creating a PR:** build a `pr/<name>` branch with only the feature
commits on top of `main`, excluding the develop-only commits:

```bash
git fetch origin main
git checkout -b pr/<name> feat/<name>
git rebase --onto origin/main develop pr/<name>
git diff --stat origin/main pr/<name>   # must not include CLAUDE.md or other develop-only files
git push -u origin pr/<name>
```

**Staging:** `staging` consolidates all pending features. Rules:

- Base it on `develop`, and merge only `feat/*` branches into it, never `pr/*`.
- Never branch from `staging`: new `feat/` branches start from `develop`.
- Rebuild it instead of accumulating history, whenever a feature is added, merged upstream,
  dropped, or its `feat/` branch is rebased.
- Resolve conflicts between features while merging into `staging`, not in the `feat/` branches.

```bash
git fetch origin
git checkout -B staging origin/develop
for b in origin/feat/<a> origin/feat/<b>; do   # every pending feat/ branch, one at a time
  git merge --no-ff "$b"                        # resolve conflicts here, then continue
done
git push --force-with-lease origin staging
```

Pushing `staging` runs `.github/workflows/staging.yaml`: the unit tests (`make test-minimal`,
`make test-extended`), then the full image for linux/amd64 and linux/arm64, pushed as
`ghcr.io/rophy/zot:staging-<short-sha>`. Upstream's own workflows are disabled in this fork.

## Commit Author and Sign-off

Author every commit as the user, and sign it off (DCO) with `-s`:

```bash
git config user.name "Rophy Tsai"
git config user.email "rophy@users.noreply.github.com"
git commit -s
```
