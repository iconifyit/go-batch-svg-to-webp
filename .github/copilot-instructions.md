## Code review instructions

> ### ⛔ MANDATORY OUTPUT FORMAT — READ THIS FIRST ⛔
>
> **Every single review comment you post MUST begin with this exact prefix, with no exceptions:**
>
> ```
> [SEV: security|core|edge|cosmetic] [fix-now|defer-ok] <one-sentence summary>
> ```
>
> - This is a hard requirement, not a suggestion. A finding without this prefix is **invalid and must not be posted**.
> - Apply it to **every** inline comment on **every** review, including re-reviews.
> - Choose the severity from the actual impact (see "Finding severity" below): `security` and `core` are always `fix-now`; `edge`/`cosmetic` are usually `defer-ok`.
> - If you cannot classify a finding into one of the four severities, do not report it.
> - Before submitting the review, verify that **100% of your comments** carry the prefix. Do not submit an untagged comment.

When performing a pull request review, follow these instructions for selecting,
classifying, and reporting findings.

### Review objective

Review the proposed changes for defects that could make the pull request unsafe,
incorrect, insufficiently tested, or unnecessarily difficult to maintain.

Focus on behavior introduced or affected by this pull request. Inspect surrounding
code when necessary to understand the change, but do not report unrelated
pre-existing problems unless the pull request makes them materially worse.

Report only findings that are specific, reproducible, and actionable. Do not
report speculative concerns without identifying a concrete failure mode.

### Review priorities

Review in this order and allocate attention accordingly. The examples are
representative, not exhaustive; report other findings that satisfy the category
definitions.

1. **Security** — exploitable vulnerabilities or sensitive-data exposure,
   including but not limited to injection, authentication or authorization gaps,
   embedded secrets, unsafe deserialization, path traversal, SSRF, and sensitive
   data exposed through logs, errors, or responses.
2. **Correctness** — behavior that produces an incorrect result or state,
   including but not limited to logic errors, broken primary flows, incorrect
   data mutations, race conditions, unhandled asynchronous failures, and
   resource leaks.
3. **Edge cases** — incorrect behavior under bounded or unusual conditions,
   including but not limited to empty, null, zero, maximum, or malformed inputs;
   off-by-one errors; timezone and encoding issues; pagination boundaries;
   retries; and concurrent access.
4. **Tests** — inadequate verification of changed behavior, including but not
   limited to missing coverage, tests that would pass if the implementation were
   broken, assertions that do not verify the intended outcome, and important
   failure paths that are not exercised.
5. **Maintainability** — code qualities that create a concrete risk of future
   defects, including but not limited to dead code, unnecessary duplication,
   misleading names, incorrect documentation, and avoidable complexity.
6. **Style** — review last and only when the code violates this repository's
   documented conventions. Do not report personal preferences or formatting
   that an automated formatter should handle.

### Finding severity

Assign exactly one of the following four severity labels to every finding. These
four labels are exhaustive, but the examples listed under them are not. Classify
unlisted problems according to their actual impact.

- `[SEV: security]` — an exploitable vulnerability, authorization failure,
  secret exposure, or sensitive-data disclosure. Always `fix-now`.
- `[SEV: core]` — breaks, corrupts, or materially compromises a primary feature,
  data path, or expected workflow. Always `fix-now`.
- `[SEV: edge]` — produces incorrect behavior in a bounded, non-primary case.
  Use `fix-now` when the case can cause meaningful harm; otherwise use
  `defer-ok`.
- `[SEV: cosmetic]` — affects naming, documentation, readability, or documented
  style without changing runtime behavior. Normally `defer-ok`.

Use the impact of the actual failure—not the size of the proposed fix—to assign
severity.

### Reporting findings

- Report one distinct finding per comment.
- Anchor the comment to the smallest relevant changed line or range.
- Use this exact opening format:

  `[SEV: <security|core|edge|cosmetic>] [<fix-now|defer-ok>] <one-sentence summary>`

- After the opening line, explain:

  1. the conditions that trigger the problem;
  2. the incorrect or unsafe result; and
  3. why the pull request causes or exposes that result.

- Propose a concrete correction. Include a code suggestion when the change is
  local and unambiguous.
- Do not report a theoretical concern unless you can describe a plausible
  execution path or input that triggers it.
- Do not duplicate a finding across lines or files. Report it once and mention
  other known occurrences in the same comment.
- Do not request defensive checks on every internal call unless required by a
  documented project rule or a demonstrated failure mode.
- Do not report issues already enforced reliably by the repository's formatter,
  linter, compiler, or type checker unless the pull request demonstrates that
  the enforcement is absent or bypassed.

Example of a correctly formatted finding:

> `[SEV: edge] [fix-now]` Pagination offset is not clamped, so paging past the
> search backend's limit returns a 500 instead of an empty page.
>
> When `start` exceeds the backend's maximum offset, the query throws and the
> request fails. Any client that pages deep enough triggers it. This pull
> request introduces the unclamped pass-through in `searchIcons()`.
>
> Suggested fix: clamp `start` to the configured maximum offset before
> building the query.

### Review summary

Conclude with a short summary containing:

- the number of findings at each severity;
- whether any `fix-now` findings remain;
- the highest-risk area reviewed; and
- any important behavior that could not be verified from the pull request.

If there are no actionable findings, state explicitly:

`No actionable security, correctness, edge-case, testing, maintainability, or documented-style issues found.`

Do not invent findings merely to populate the review.

### Project rules that constrain the review

Treat the following repository rules as authoritative. Do not flag code that
follows them. When a finding depends on a project rule, cite the relevant rule
or file in the comment.

- **Rule sources:** `CLAUDE.md`, `AGENTS.md`, `docs/CHANGE-WORKFLOW.md`.
- **Architecture and validation:** Validate at system boundaries (user input,
  external API responses). Do not request redundant defensive validation on
  trusted internal calls or framework/library guarantees. This is a Next.js
  front end that talks to the API via a signed `http`/`axios` client — the API,
  not the FE, is the trust boundary for data it returns.
- **Coding conventions:** `camelCase` for variables/functions; `PascalCase` for
  classes, enums, and React components; `UPPER_SNAKE_CASE` for global constants
  (prefix env-derived constants with `k`); `lower_snake_case` only for database
  column names. Align object properties on the colon. Put `else`/`else if` on
  its own line; prefer an early `return`/`throw` over a redundant `else`. Use
  ternaries only to choose between two values, never to choose between actions,
  and no more than two per expression. Document functions with JSDoc.
- **Testing conventions:** Every test opens with a comment stating its scenario.
  Seed realistic domain data — no `foo`/`bar`/arbitrary IDs. Freeze time to a
  fixed ISO timestamp and derive relative dates from it; no future dates. No
  empty-table tests unless explicitly requested. A test must fail if the
  implementation it covers is removed. E2E lives in `e2e/` (Playwright);
  integration/unit under the relevant module.
- **Generated or vendored code:** Do not give ordinary source-level review to
  auto-generated sitemap/RSS files (`public/sitemaps/**`, RSS feed XML),
  lockfiles (`yarn.lock`, `package-lock.json`), or Jest `__snapshots__`.
