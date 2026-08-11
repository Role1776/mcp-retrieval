<h1 align="center">Contributing to mcp-retrieval</h1>

<p align="center">
  <b>How to get a change reviewed and merged without either of us losing an afternoon to it.</b>
</p>

---

mcp-retrieval is a small, layered Go project — an MCP server wrapping web retrieval into a handful of tools. That size is the reason for most of the rules below: a change that would be routine in a large codebase can rewrite a meaningful fraction of this one, and a pull request that touches every layer at once is effectively unreviewable here.

Contributions are welcome. The guidelines below are about **the shape of the pull request**, not about whether an idea is worth having.

> [!NOTE]
> These rules apply to code contributions. Typo fixes, documentation corrections, and small README improvements can ignore the size and scope limits entirely — just send them.

---

## The Rules in Short

| Rule | Limit |
| :--- | :--- |
| **One logical change per PR** | If the title needs the word "and", split it |
| **Size** | ≤ 300 new lines of code, excluding tests and documentation |
| **Independence** | Each PR branches from current `main` and contains only its own commits |
| **Open at once** | 2–3 pull requests maximum |
| **Compatibility** | Existing configs and MCP clients keep working after `git pull` with no manual steps |
| **Refactoring** | Its own PR, with no functional changes mixed in |
| **Verification** | `go build`, `go vet`, `go test` all green, plus a note that you ran the server |

Each of these is expanded below.

---

## 1. One PR, One Logical Change

A pull request should do exactly one thing, and its title should say what that thing is without needing a conjunction.

The most common violation is bundling a **new tool or capability** together with the **plumbing that reshapes a shared layer**. Those are two separate concerns with different risk profiles — changing a `usecase` or `domain` type ripples through every tool that touches it, while a new handler is ordinary logic that can be reverted freely. Reviewing them together means reviewing the risky part while distracted by the boring part.

**Not acceptable — one PR:**

```text
Add a web_summarize tool and change how the usecase resolves timeouts
```

**Acceptable — two PRs:**

```text
PR 1: Rework timeout resolution in usecase/web
PR 2: Add the web_summarize tool on top of the new timeout handling
```

Other pairs that belong in separate pull requests:

* A bug fix and a new feature — even if you found the bug while building the feature.
* A new tool and unrelated config plumbing for a different setting.
* Changing tool *output* (the response shape) and changing the *retrieval* behavior behind it.
* Anything plus a drive-by cleanup of surrounding code (see [Refactoring](#6-refactoring-goes-in-its-own-pr)).

The test is simple: **if a reviewer might want to accept one half and reject the other, they are two pull requests.**

---

## 2. Size Limit: 300 Lines

**A pull request should add no more than ~300 lines of code.** Tests and documentation do not count toward the limit — write as many as the change deserves.

For a project this size, 300 lines is already a substantial change. If your diff is heading past it, that is almost always a signal that rule 1 has been broken and there are two or three changes in there waiting to be separated.

This is a guideline with a hard ceiling, not a precise budget. 320 lines for a genuinely single, cohesive feature is fine — just say so in the description. 700 lines is not, regardless of how cohesive it is, and it will be sent back to be split.

> [!TIP]
> Check before you open the PR: `git diff --stat main...HEAD`.

---

## 3. Pull Requests Must Be Independent

**Branch every pull request from the current `main`.** The diff of your PR must contain only your own commits — never the commits of another open pull request.

This is the rule most often broken by accident, usually by branching feature B off feature A's branch because A "isn't merged yet". The result is a pull request whose diff is mostly somebody else's unreviewed work, where there is no way to see what is actually new without reconstructing the branch history by hand. If seven pull requests each contain the previous six, there are not seven reviewable changes — there is one large one wearing seven hats.

**If a change genuinely builds on another one that is still under review, you have two options.**

### Option A — wait (preferred)

1. Open the PR that goes first (say, the shared `domain` change). Nothing else.
2. Wait for it to be reviewed and merged.
3. `git fetch && git rebase origin/main` your follow-up work onto the new `main`.
4. Open the next PR.

Simplest for everyone, and the right default when the follow-up is not urgent.

### Option B — stack explicitly

If waiting is impractical, open the dependent PR with its **base branch set to the parent PR's branch**, not `main`, and note the dependency in the description (`Depends on #NN`).

GitHub then shows only your own changes in the diff — which is the entire point of this rule — and retargets the PR to `main` automatically once the parent is merged.

> [!IMPORTANT]
> If pull requests here are **squash-merged**, the parent's commits are replaced by a single new commit when it lands. Your stacked branch still carries the originals, and will show duplicated changes and conflicts until you `git fetch && git rebase origin/main` it onto the squashed result. Expect one rebase per merged parent.

What is **not** acceptable under either option is a PR based on `main` whose diff contains another open PR's commits. That is the case this rule exists to prevent.

**Keep no more than 2–3 pull requests open at a time.** A queue of eleven open PRs that all touch the same usecase layer will conflict with each other no matter what order they merge in, and resolving that is work the maintainer did not sign up for.

> [!WARNING]
> **Don't force-push to a branch once review has started.** Rewriting history mid-review discards the review comments' context and makes it impossible to see what changed since the last look. Add new commits instead; they get squashed on merge anyway.

---

## 4. Don't Break Existing Setups

mcp-retrieval is self-hosted and embedded into other people's MCP clients. People wire the binary into Claude Desktop or an IDE agent and pull updates occasionally. **After `git pull` and a rebuild, an existing setup must keep working with no manual intervention.**

Concretely:

* **New config fields and environment variables must be optional**, with a sensible default (an `envDefault` tag, and a matching line in `.env.example`) so a config that worked yesterday still works today. A field that must be set for the server to start is a breaking change.
* **Never rename or repurpose an existing config key or env var.** If `MAX_RESULTS` means snippets-per-query today, it means that forever.
* **Don't rename or remove an existing tool, or change the meaning of its parameters**, without a reason stated in the PR description. Clients have wired those tool names and schemas into prompts; a rename silently breaks them. Adding a new optional parameter is fine; removing or repurposing one is not.
* **Keep the two transports (`stdio` and `http`) at parity.** A tool must behave identically regardless of transport.
* **Don't add new system-level dependencies** without discussing it in an issue first.

New behavior should be additive: off by default, or on by default only when it cannot surprise anyone.

---

## 5. Commits

Split code from documentation:

```text
Add freshness filter to web_search
Document the freshness filter on web_search
```

Beyond that, keep commit messages in the imperative mood (`Add ...`, `Fix ...`, not `Added ...`), and describe what the commit does rather than which files it touched. If commits are squashed on merge, a handful of clean ones is plenty — no need to rewrite history to reach exactly one.

---

## 6. Refactoring Goes in Its Own PR

If, while implementing something, you find code that wants restructuring — a function that has grown too long, a pattern duplicated across the three tool handlers, a helper worth extracting into `pkg/` — **do not fix it in the same pull request**.

Open a separate PR that does the refactoring and **nothing else**: no new features, no behavior changes, no new configuration. A pure refactoring PR is easy to review, because the reviewer's only question is "does this do the same thing as before?". Mixed in with a feature, the same change is nearly impossible to review, because there is no way to tell which lines moved and which lines are new logic.

This also keeps the 300-line limit honest. "Half of this diff is just cleanup" is not an exemption — it is two pull requests.

If the refactoring is a prerequisite for your feature, submit it first and sequentially, per [rule 3](#3-pull-requests-must-be-independent).

---

## 7. Verify That It Actually Runs

### Automated checks

The following must all be clean before review, and each runs locally with the same command:

| Check | Command | What it catches |
| :--- | :--- | :--- |
| **Build** | `go -C app build ./...` | Anything that doesn't compile |
| **Vet** | `go vet ./...` | Suspicious constructs the compiler allows |
| **Tests** | `go -C app test ./...` | Regressions in covered behavior |

If a linter such as `golangci-lint` is configured in the repo, run it too and keep it green.

A green run does not mean a change is in scope. All the rules above still apply.

### Running it for real

Building is not the same as working. The minimum bar is that **you ran the server and exercised the path you touched**, and that you say so in the pull request description.

```bash
make build                      # the Go module lives in app/
./bin/mcp-retrieval -env app/.env    # or plain ./bin/mcp-retrieval for stdio defaults
```

If your change touches a tool, call it from a real MCP client (or over HTTP against `http://localhost:8080/mcp`) and confirm the response — the structured output and, on failures, the `isError` message — looks right. If it touches config loading, test it against both a present and an absent `.env` file, since the loader treats a missing file as "use defaults".

Tests are very welcome, and they do not count toward the size limit. The existing `config_test.go` is the pattern to follow. If you add the first tests for a package, that is a contribution in its own right — send it as a PR that adds tests and nothing else.

---

## 8. Pull Request Descriptions

Explain **why**, not **what** — the diff already covers what. A useful description answers:

* What problem does this solve, and how does it show up for someone running the server?
* Why this approach rather than an obvious alternative?
* What did you run to verify it? (see [rule 7](#7-verify-that-it-actually-runs))
* Any new config fields or environment variables, with their defaults.
* Anything you deliberately left out of scope.

Two or three honest sentences beat a bulleted restatement of the diff.

---

## Before You Open a Pull Request

```text
[ ] One logical change — the title needs no "and"
[ ] ≤ ~300 new lines of code (git diff --stat main...HEAD)
[ ] Branched from current main; diff contains only my commits
[ ] No more than 2–3 of my PRs open at once
[ ] New config fields / env vars are optional and defaulted
[ ] Existing configs and tool schemas still work untouched
[ ] No unrelated refactoring mixed in
[ ] go -C app build ./... && go -C app vet ./... && go -C app test ./... — all clean
[ ] Ran the server and exercised the path; said so in the description
[ ] Description explains why, not what
```

---

## Larger Ideas

For anything substantial — a new subsystem, a new transport, a change to the layer boundaries, a new dependency — **open an issue first** and describe the idea before writing the code. It is a much cheaper conversation than a 700-line pull request that gets sent back to be split into five.

This is not a gate on ambition. Big ideas are fine; they just need to arrive as a sequence of small pull requests, and agreeing on the shape of that sequence up front saves everyone the rework.

---

## License

By contributing, you agree that your contributions are licensed under the **MIT License**, the same as the rest of the project. See [`LICENSE`](LICENSE).
