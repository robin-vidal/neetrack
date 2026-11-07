> PARKED 2026-11-07 (Samy): no new work unless he asks. The push and co-author rules below apply only if he revives it.

# Neetrack

Neetcode 150 tracker with spaced repetition. Go + React monorepo, self-hosted.

The full specification lives in `docs/cahier-des-charges.md` (written in French). **Read it
before planning anything.** This file only holds the rules that must stay true across every
session.

## Invariants, do not "simplify"

These four points are counter-intuitive and already settled. If you think one is wrong, say so.
Do not change it.

1. **Review state belongs to the pattern, not to the problem.** `pattern_cards` is keyed by
   `(user_id, pattern)`, never by problem. Putting an FSRS card on each problem breaks the
   product: it measures recall of one solution instead of mastery of a pattern.
2. **`attempts` is append-only.** No `UPDATE`, no `DELETE` on that table. FSRS state is derived
   from it and must stay reconstructible by replay.
3. **No scheduling logic on the client.** No FSRS formula, no retrievability computation, no
   session composition in TypeScript. The frontend renders what the API returns.
4. **The server computes the rating** from duration, hint usage and whether the solution was
   viewed. A rating sent by the client is only a fallback when the timer never ran.

## Language

- **Code is English.** Identifiers, types, functions, packages, database columns, commit
  messages, PR titles, godoc, JSDoc, log messages, error strings.
- **The user interface is French.** Every string a user reads stays in French: labels, buttons,
  headings, empty states, validation and error messages surfaced in the UI. The design files
  already carry the exact wording, keep it.
- The spec in `docs/` is French and stays French. Comments in code are English.
- An API error carries a machine code plus a French `message` field, since that message reaches
  the screen: `{"error":"invalid_credentials","message":"Identifiant ou mot de passe incorrect."}`

## Writing style

- **Godoc on exported symbols, nothing else.** One sentence starting with the symbol name, in
  idiomatic Go form: `// Compose builds a session from due patterns and unseen problems.` Do not
  restate the signature, no `Parameters` section, no decorative banner.
- **No comments inside function bodies.** Naming and structure carry the intent. If a block
  needs explaining, extract it into a well-named function. Only exception: a single line
  justifying a non-obvious choice the code cannot express (a workaround, an external
  constraint, a subtle invariant).
- Same rule in TypeScript: JSDoc on module exports, nothing inside bodies.
- **No em dashes** anywhere: code, commit messages, PR titles, documentation. Use a comma, a
  colon, parentheses, or one more sentence.

## Git

- **Always check the current branch before working** (`git branch --show-current`). Never commit
  directly on `main`.
- One branch per unit of work, not one per milestone. A milestone splits into several successive
  branches when it touches distinct parts of the codebase.
- Branch naming: `feat/`, `fix/`, `refactor/`, `test/`, `chore/`, `ci/`, `docs/` followed by a
  short slug.
- **Atomic commits**, conventional commit format: `feat(scheduler): compose session from due
  patterns`. One commit does one thing. If the message needs an "and", split it.
- **Never push on your own initiative.** When work on a branch is done, stop and tell me, with
  the list of commits. I review.
- If I answer "go": push the branch, open the PR with `gh pr create` passing **the title only**
  (empty body, no description, no summary, no checklist), merge it, then close it and delete the
  remote branch.
- **Never add yourself as co-author.** No `Co-authored-by` trailer, no mention of Claude, AI or
  assisted generation in commit messages, PR titles or descriptions.

## Technical conventions

- Go 1.25+, `net/http` and `http.ServeMux` only. **No web framework**, no ORM.
- `pgx/v5` + `sqlc`. SQL is hand-written in `internal/store/queries/`.
- No global state, no `init()`. Dependencies are built in `cmd/neetrack/main.go` and passed
  explicitly.
- `internal/scheduler` knows nothing about HTTP or SQL. Pure domain, testable without a database.
- Errors wrapped with `%w`, exported sentinels. The store layer knows nothing about HTTP status
  codes.
- `context.Context` as first parameter, propagated down to pgx.
- Frontend: React + TypeScript, TanStack Query for server state. No Redux.

## The design is given, not to be invented

`design/neetrack.jsx` and `design/neetrack-auth.jsx` are the visual reference. Port them to TSX
keeping the CSS, the structure and the French copy verbatim. While porting, strip the FSRS logic
they contain (it moves server-side) and replace `window.storage` with API calls.

Do not redesign. Do not swap the CSS for Tailwind. Do not translate the interface to English.

## How we work

- **One milestone at a time** (§12 of the spec). Do not start the next one until the previous
  one compiles, passes its tests and is merged.
- Write the `internal/scheduler` tests alongside the code, not after. That is where a bug is
  invisible to the naked eye.
- Read a file before modifying it. Do not rewrite a whole file for a one-line change.

## Do not touch

- `.env`, `compose.override.yaml`, anything holding credentials. Those live on the VPS, not in
  the repository.
- GitHub Actions secrets. The workflow references them, it does not create them.
- Migrations already applied. A schema fix is a new migration.

## Commands

```
make dev       # dev compose, Go + Vite hot reload
make test      # go test ./... plus frontend tests
make lint      # golangci-lint, eslint, tsc
make sqlc      # regenerate internal/store from .sql files
make migrate   # apply migrations locally
make adduser   # create an account (password read from stdin)
```
