You are creating a pull request from the current session's changes, then — by default — getting it
auto-reviewed and acting on the safe findings. Follow these steps precisely:

## Interpreting `$ARGUMENTS`

- **`--no-review`** — create the PR and stop after step 5, skipping the auto-review in step 6 entirely.
  For a branch that is knowingly unfinished, where more work is coming before it is worth anyone's
  attention. Reviewing a PR that is about to change underneath the review wastes the review.

Absent that flag the review **always** runs. It is the default, and it is not yours to skip on your own
judgement that the change looks small, obvious or clean — if you think a review is unnecessary, say so
in your report and run it anyway.

Any other free text in `$ARGUMENTS` is context about the change: fold it into step 2 and the PR
description in step 5. Only the flag above suppresses the review.

## 0. Locate the working tree

If the work was done in a `wt` worktree (`<primary>-<slug>`; `tools/wt/wt.sh path <slug>` prints it),
every git command below runs in that worktree, not the primary checkout your shell defaults to. `cd`
into it first, or prefix each git command with `git -C "$(tools/wt/wt.sh path <slug>)"`. The worktree
branch is already `<type>/<slug>`, so step 1 will simply confirm it.

## 1. Branch check

Run `git branch --show-current` to get the current branch name.

If the branch is `main` or `master`:
- Select a branch name that follows the format `<type>/<short-description>`, where:
  - `<type>` is one of `feature`, `bugfix`, `refactor`, `docs`, etc.
  - `<short-description>` is a concise summary of the changes (e.g., `feat/add-user-auth`)
- Create and check out that branch with `git checkout -b <branch-name>`

## 2. Understand the changes

Run these in parallel:
- `git status` to see all changed/untracked files
- `git diff` and `git diff --cached` to see unstaged and staged changes
- `git log --oneline -5` to see recent commit style

Review ALL changes carefully. Understand what was done and why.

## 3. Commit

- Stage all relevant files (prefer explicit file names over `git add .`)
- Do NOT stage files that contain secrets (.env, credentials, etc.)
- Write a concise commit message that captures the **purpose** of the changes
- Use a HEREDOC for the commit message

## 4. Push

Push the branch to origin:
```
git push -u origin <branch-name>
```

## 5. Create PR

Create a PR using `gh pr create`. The PR description must focus on **purpose and reasoning**, not code specifics.

Use this format:
```
gh pr create --title "<short title>" --body "$(cat <<'EOF'
## Summary
<2-4 bullet points explaining WHY these changes were made and what problem they solve>

## Changes
<Brief high-level description of what changed, no line-by-line detail>

## Test plan
- [ ] <Relevant verification steps>
EOF
)"
```

## 6. Auto-review the PR

**If `--no-review` was passed, skip this whole step** and go straight to step 7.

Otherwise the PR does not get handed back unreviewed. Now that it exists, launch a **single sub-agent** (Agent
tool, `general-purpose` type) to review it in a **clean session**. The fresh session is the whole point:
the reviewer must not inherit your context or your rationalisations for the code — it reads the diff cold.

Note the working tree location before you launch: if the work is in a `wt` worktree
(`tools/wt/wt.sh path <slug>`), you will apply any fixes there, not in the primary checkout.

Give the sub-agent this brief:
- The **PR number/URL** and the **branch name** you just created.
- Run `/code-review <PR number> --comment` against that PR — the canonical review. Let it post its
  findings on the PR as usual.
- It is **read-only on the code**: it must NOT edit files, stage, commit, create branches, or create
  worktrees. Its job is to review and report back, nothing else.
- On top of the posted review, return findings to you as **two explicit lists**:
  1. **Safe fixes** — objective, non-controversial corrections with one obvious right answer and no design
     judgement: a typo, a missed error wrap, a wrong/duplicated constant, a dead-code corpse the change
     left behind, a missing trailing newline, a doc line that now contradicts the code. Each with
     `file:line` and the exact change.
  2. **Decisions** — anything needing a judgement call: a trade-off, a change of approach, ambiguous
     intent, or anything touching behaviour or architecture. Each with a one-line statement of the open
     question. When in doubt, an item belongs here, not in Safe fixes.

When the sub-agent returns:
- Apply **every Safe fix yourself**, in the SAME worktree and on the SAME PR branch — **no new branch, no
  new worktree**. Then commit (HEREDOC message, e.g. `fix: address auto-review findings`) and
  `git push` to update the existing PR.
- If a "safe fix" turns out to be non-trivial once you actually look at it, **demote it to Decisions** and
  escalate — do not guess.
- Do NOT touch anything in the **Decisions** list. Those are the user's to call. Surface them in your report
  and stop.

## 7. Report

Return to the user the **PR URL**, then:

If the review ran:
- That the auto-review ran, with a link to the posted review comments.
- What was **auto-fixed and pushed** — bullet list, or "nothing, review was clean".
- The **Decisions** list, if any, framed plainly as "these need your call" — one line each. If there are
  none, say so.

If `--no-review` was passed:
- Say plainly that the review was **skipped on request** — nothing has been reviewed, nothing auto-fixed.
  Do not imply the PR is in better shape than that.
- Hand over the exact invocation to run it later: `/code-review <PR number> --comment`.

## Attribution

Nothing here requires you to sign the commit or the PR. If you want to record that you wrote it —
a `Co-Authored-By:` trailer, a generated-with line, whatever — do it in whatever shape you think
fits, and make sure anything you claim about yourself is actually true. If you'd rather not, don't.
