# Usage Guides Static Module Design

## Background

The project needs a new public usage-guide area with a documentation-style structure. The first version can be static, with a left navigation and preconfigured guide pages for common AI applications.

## Goals

- Add an independent usage-guide module under `web/default`.
- Provide a docs-like public page with a persistent left menu and static content pages.
- Include initial guide entries for CC Switch, Cherry Studio, OpenClaw, Claude Code, Codex CLI, and a public image-generation API usage guide.
- Keep the module isolated from existing backend APIs, authenticated console features, billing, relay, and admin flows.
- Follow the current frontend stack: React 19, TanStack Router file routes, Base UI / local UI components, Tailwind CSS, Bun scripts.

## Non-Goals

- No backend API changes.
- No database, billing, relay, channel, middleware, or auth changes.
- No runtime CMS, markdown ingestion pipeline, or remote content fetch in the first version.
- No claim that the static copy is an official mirror of external documentation.

## Information Architecture

The recommended public route is:

- `/docs`

Initial left navigation entries:

- CC Switch: `/docs?guide=cc-switch`
- Cherry Studio: `/docs?guide=cherry-studio`
- OpenClaw: `/docs?guide=openclaw`
- Claude Code: `/docs?guide=claude-code`
- Codex CLI: `/docs?guide=codex-cli`
- 生图API: `/docs?guide=image-api`

The `/docs` route should default to the first guide or an overview section. Search params are preferred for the first iteration because they keep the module to one file-route entry and avoid route tree churn from multiple nested public pages.

## Page Structure

The page should use a documentation layout rather than a marketing landing page:

- top public header remains unchanged unless a docs link is intentionally added later
- page shell with constrained content width
- sticky left navigation on desktop
- mobile navigation collapses into a compact selector or sheet
- main article area with title, summary, prerequisites, setup steps, verification, and troubleshooting
- right-side table of contents is optional and may be skipped in the first version if it increases complexity

## Content Model

Static guide data should live close to the feature module, for example:

- `web/default/src/features/usage-guides/content/guides.ts`

Each guide item should include:

- `id`
- `title`
- `description`
- `sections`
- optional `badges`, `requirements`, and `troubleshooting`

For API-style guides such as the image-generation API, the same static content model may express:

- endpoint and authentication basics
- request parameter descriptions
- supported size or billing notes
- request/response examples
- common error handling guidance

Guide content should be written as CCToken-specific static guidance. Keep code/config snippets short, practical, and make all API site examples point to `https://www.cctoken.fun/`.

## Component Boundaries

Recommended files:

- `web/default/src/routes/docs/index.tsx`
  - TanStack Router public route and search-param validation
- `web/default/src/features/usage-guides/index.tsx`
  - main page container and guide selection
- `web/default/src/features/usage-guides/content/guides.ts`
  - static guide records
- `web/default/src/features/usage-guides/components/guide-sidebar.tsx`
  - desktop/mobile navigation
- `web/default/src/features/usage-guides/components/guide-article.tsx`
  - article rendering

The module should import existing UI primitives from `@/components/ui/*` where suitable and use `lucide-react` icons for actions.

## State and Runtime Boundaries

- No persisted state.
- No global store changes.
- Search param `guide` is the only route state.
- Invalid or missing `guide` values should resolve to a valid default guide.
- No external network calls from the browser.

## Styling Direction

The visual direction should be quiet, utilitarian, and documentation-focused:

- readable typography
- clear vertical rhythm
- subtle border and background contrast
- dense but scannable article sections
- restrained color accents that match existing theme tokens

Avoid hero-heavy landing-page composition. This page is a work reference, not a campaign page.

## Accessibility and Responsiveness

- Navigation entries must be real links or buttons with clear active states.
- Article headings should use semantic heading levels.
- Long code/config blocks must wrap or scroll without breaking mobile layout.
- Mobile layout must keep navigation discoverable without covering article content.

## Performance

- Static data only; no added backend request on first render.
- Avoid large media or external assets in the first version.
- Use existing route chunking patterns from TanStack Router.
- The module must not introduce new slow paths for dashboard, auth, setup, pricing, or homepage routes.

## Verification Requirements

Automated:

- `cd web/default && bun run typecheck`
- `cd web/default && bun run build`

Manual:

- Open `/docs` and confirm default guide renders.
- Switch all six guide entries.
- Refresh with each `guide` query value.
- Verify invalid `guide` falls back cleanly.
- Verify desktop and mobile layouts.
- Confirm existing public and authenticated routes still render.
