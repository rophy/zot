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

`main` matches upstream. `develop` is `main` plus personal commits (this file, etc.)
that must never go upstream.

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

## Commit Author and Sign-off

Author every commit as the user, and sign it off (DCO) with `-s`:

```bash
git config user.name "Rophy Tsai"
git config user.email "rophy@users.noreply.github.com"
git commit -s
```
