# Usage Guides Static Module Implementation Plan

Design reference: `docs/superpowers/specs/2026-05-16-usage-guides-design.md`

## Design-Implementation Mapping

- Design document: `docs/superpowers/specs/2026-05-16-usage-guides-design.md`
- Route entry: `web/default/src/routes/docs/index.tsx`
- Feature entry: `web/default/src/features/usage-guides/index.tsx`
- Static content: `web/default/src/features/usage-guides/content/guides.ts`
- Display components:
  - `web/default/src/features/usage-guides/components/guide-sidebar.tsx`
  - `web/default/src/features/usage-guides/components/guide-article.tsx`
- State: public route search param `guide`
- Persistent storage: none
- Backend/API impact: none
- Verification commands:
  - `cd web/default && bun run typecheck`
  - `cd web/default && bun run build`
- Manual acceptance:
  - `/docs` renders independently
  - all five predefined guides render and switch correctly
  - invalid guide values fall back cleanly
  - desktop/mobile layouts are readable
  - no historical public/authenticated route behavior is changed

## Execution Checklist

### Task 1: Source and route reconnaissance

- Status: `verified`
- Files:
  - `web/default/src/routes`
  - `web/default/src/features`
  - `web/default/package.json`
- Work:
  - confirm current TanStack Router generation workflow
  - confirm import aliases and UI primitives
  - confirm the final page does not expose upstream docs links
- Verification:
  - record chosen route path and content-source policy in this plan
- Notes:
  - chosen public route: `/docs`
  - content-source policy: no rendered or stored links to the upstream docs domain
  - API site examples use `https://www.cctoken.fun/`
  - confirmed current frontend uses TanStack Router file routes under `web/default/src/routes`
  - confirmed project scripts prefer Bun, but this environment does not provide `bun`; verification used `npm run` against the same package scripts

### Task 2: Static guide content model

- Status: `verified`
- Files:
  - `web/default/src/features/usage-guides/content/guides.ts`
- Work:
  - define typed static guide records for CC Switch, Cherry Studio, OpenClaw, Claude Code, and Codex CLI
  - paraphrase setup content from the referenced pages
  - remove upstream documentation links and keep all rendered guidance CCToken-owned
  - make API site examples point to `https://www.cctoken.fun/`
- Verification:
  - TypeScript typecheck covers the data structure
- Result:
  - implemented typed static guide data in `web/default/src/features/usage-guides/content/guides.ts`
  - removed upstream docs source links and paraphrased setup guidance for all five initial guides
  - updated API examples and guide copy so API site references use `https://www.cctoken.fun/`

### Task 3: Usage guide route and page shell

- Status: `verified`
- Files:
  - `web/default/src/routes/docs/index.tsx`
  - `web/default/src/features/usage-guides/index.tsx`
- Work:
  - add public `/docs` route
  - validate optional `guide` search param
  - render default guide when no valid guide is selected
- Verification:
  - route tree generation succeeds during build/typecheck
- Result:
  - added `web/default/src/routes/docs/index.tsx`
  - added `web/default/src/features/usage-guides/index.tsx`
  - default and invalid guide values fall back to `cc-switch`

### Task 4: Navigation and article components

- Status: `verified`
- Files:
  - `web/default/src/features/usage-guides/components/guide-sidebar.tsx`
  - `web/default/src/features/usage-guides/components/guide-article.tsx`
- Work:
  - implement desktop left menu and mobile selector/sheet
  - implement semantic article rendering for prerequisites, setup steps, verification, and troubleshooting
  - use existing design tokens and local UI primitives
- Verification:
  - manual desktop/mobile inspection
- Result:
  - implemented desktop sidebar and mobile select navigation
  - implemented article rendering for summary, prerequisites, setup steps, verification, and troubleshooting
  - mobile select trigger now shows the human-readable guide title rather than the slug

### Task 5: Integration hygiene

- Status: `verified`
- Files:
  - `web/default/src/routeTree.gen.ts`
  - possible generated route type files if produced by the project tooling
- Work:
  - allow router generation/build tooling to update generated route artifacts if required
  - avoid unrelated route, layout, auth, backend, or i18n churn
- Verification:
  - inspect git diff to ensure scope is limited to docs route/module/generated router artifacts
- Result:
  - generated router artifact updated: `web/default/src/routeTree.gen.ts`
  - implementation scope stayed within the new `/docs` module plus router generation
  - follow-up CCToken URL cleanup intentionally touched public docs links in README files, default/classic footers, default/classic docs-link settings, and navigation fallback logic so stale legacy upstream docs runtime config no longer renders externally

### Task 6: Automated verification

- Status: `verified`
- Commands:
  - `cd web/default && bun run typecheck`
  - `cd web/default && bun run build`
- Work:
  - run fresh verification
  - record command results in this plan
- Verification:
  - both commands pass or failures are documented with whether they are related to this module
- Verification notes:
  - `cd web/default && bun run typecheck` could not run because `bun` is not installed in this environment
  - fallback `cd web/default && npm run typecheck` passed on 2026-05-16
  - fallback `cd web/default && npm run build` passed on 2026-05-16
  - build regenerated `web/default/src/routeTree.gen.ts`
  - 2026-05-16 follow-up after CCToken URL cleanup:
    - `cd web/default && npm run typecheck` passed
    - `cd web/default && npm run build` passed
    - `go test ./setting/...` passed
    - classic locale JSON parse check passed
    - full repository search excluding generated dist and dependencies found no legacy upstream docs URL or placeholder API domain

### Task 7: Manual acceptance and delivery

- Status: `verified`
- Work:
  - run or use an existing frontend server
  - open `/docs` in the browser
  - verify all five guide selections and responsive behavior
  - report final route address to the user
- Completion condition:
  - no checklist item remains `todo` or `doing`
  - verification results are recorded
  - final route address is delivered
- Verification notes:
  - local preview server started successfully at `http://127.0.0.1:4012/`
  - local dev server started successfully at `http://127.0.0.1:4013/`
  - Playwright with installed system Chrome verified:
    - default route `http://127.0.0.1:4013/docs` renders `CC Switch`
    - `http://127.0.0.1:4013/docs?guide=codex-cli` renders `Codex CLI`
    - invalid guide `http://127.0.0.1:4013/docs?guide=unknown` falls back to `CC Switch`
    - mobile viewport `390x844` renders the compact selector and article layout correctly
    - `http://127.0.0.1:4013/sign-in` still renders, confirming public auth entry remains healthy
  - desktop screenshots also confirmed the left navigation and article shell render correctly
  - 2026-05-16 follow-up Chrome headless DOM check verified:
    - `http://127.0.0.1:4013/docs?guide=cherry-studio` renders Cherry Studio with `https://www.cctoken.fun/` and `https://www.cctoken.fun/v1`
    - top navigation no longer renders stale legacy upstream docs domains when runtime `docs_link` contains one
    - `/docs`, `/docs?guide=codex-cli`, and `/docs?guide=unknown` do not render legacy upstream docs domains or placeholder API domains

## Blocking Conditions

- If the referenced documentation cannot be reached, stop content work and record which source failed.
- If generated router files change beyond the new `/docs` route, inspect before continuing.
- If typecheck/build fails from unrelated pre-existing issues, record the exact failure and keep this task status incomplete until the user decides whether to handle unrelated failures.
- Do not claim the module is complete until automated verification and manual route checks have fresh evidence.
