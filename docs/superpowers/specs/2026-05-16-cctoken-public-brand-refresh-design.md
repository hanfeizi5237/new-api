# CCToken Public Brand Refresh Design

## Background

The current `web/default` public landing experience is functional but visually close to the upstream default presentation. CCToken needs a stronger public-facing identity that clearly communicates "AI model access hub" instead of a generic gateway or a crypto transfer station.

The user provided a reference image with a neon control-core composition. The final design must preserve that energy while replacing chain and coin semantics with mainstream model ecosystem semantics such as OpenAI, Claude, Gemini, DeepSeek, Grok, Qwen, Mistral, and Llama.

## Goals

- Rebuild the public homepage so it is visually distinct from the current default landing page.
- Establish a cohesive CCToken brand language across homepage, public header, auth layout, and footer.
- Keep existing user flows intact: navigation, sign in, sign up, pricing, dashboard entry.
- Preserve protected attribution behavior in the footer.
- Keep the work presentation-only with no backend or runtime contract changes.

## Non-Goals

- No changes to backend APIs, auth logic, pricing logic, or system configuration behavior.
- No attempt to remove protected project attribution or licensing references.
- No dependency on third-party vendor logo asset packs.

## Design Direction

### Visual Theme

The visual theme is "AI routing command core":

- dark navy and near-black base
- electric cyan and magenta light rails
- framed HUD panels and soft glass surfaces
- central routing-core illustration built with CSS and component layout
- model ecosystem badges replacing token-chain badges

### Differentiators

- The homepage should feel like a live routing surface, not a standard SaaS hero plus cards.
- Navigation and auth should feel like secure console access rather than plain app chrome.
- Footer should read as a branded operational base while retaining protected attribution.

## Information Architecture

### Homepage

The homepage remains section-based for compatibility, but each section becomes part of one continuous branded narrative:

1. Hero
   - headline focused on unified model access and intelligent routing
   - left/right ecosystem panels with model/provider labels
   - central routing-core visual
   - CTA cluster for sign up, pricing, and dashboard
2. Metrics strip
   - stronger operational numbers and reliability framing
3. Capability grid
   - routing, billing control, security, compatibility, observability
4. Workflow section
   - explain connect -> route -> govern -> optimize
5. Closing CTA
   - reinforce fast onboarding and operational confidence

### Public Header

- compact floating shell
- branded status chip
- stronger active nav treatment
- mobile overlay visually aligned with the homepage

### Auth Layout

- split-screen secure access console treatment on desktop
- atmospheric backdrop with trust and routing cues
- centered form card on mobile and desktop

### Footer

- brand summary, product/support/legal groupings
- mission copy aligned with model access
- protected attribution remains intact and visible

## Component Boundaries

- `web/default/src/features/home/components/sections/hero.tsx`
  - custom routing-core hero with ecosystem badges and CTA actions
- `web/default/src/features/home/components/sections/stats.tsx`
  - command-strip metrics presentation
- `web/default/src/features/home/components/sections/features.tsx`
  - branded capability panels
- `web/default/src/features/home/components/sections/how-it-works.tsx`
  - operation flow section
- `web/default/src/features/home/components/sections/cta.tsx`
  - closing recruitment panel
- `web/default/src/components/layout/components/public-header.tsx`
  - navigation shell refresh
- `web/default/src/features/auth/auth-layout.tsx`
  - secure access layout refresh
- `web/default/src/components/layout/components/footer.tsx`
  - footer brand refresh while preserving attribution behavior
- `web/default/src/styles/index.css`
  - shared visual utility classes for the new brand surface

## Data, State, and Runtime Boundaries

- No new persisted state.
- No change to `useSystemConfig`, `useAuthStore`, or route behavior.
- Existing `systemName`, `logo`, and footer attribution behavior continue to drive runtime branding.
- New visuals are deterministic and derived only from render-time component state.

## Accessibility and Responsiveness

- All important text remains real text, not baked into images.
- Decorative glows and grids must remain `aria-hidden`.
- Desktop composition may be dense, but mobile must collapse into readable stacked sections.
- Interactive controls retain current link destinations and button semantics.

## Risks and Mitigations

- Risk: aggressive visual styling could reduce readability.
  - Mitigation: keep primary copy high contrast and use glow only as support.
- Risk: homepage-only styles could leak into dashboard surfaces.
  - Mitigation: scope new classes to dedicated section wrappers and component-local markup.
- Risk: vendor logo assets are incomplete.
  - Mitigation: use branded text badges and abstract provider pills instead of hard asset dependency.

## Verification

### Automated

- `cd web/default && npm run typecheck`
- `cd web/default && npm run build`

### Manual

- Verify desktop and mobile homepage layouts.
- Verify `Sign in`, `Sign up`, `Pricing`, and `Dashboard` entry points still navigate correctly.
- Verify auth pages retain usability and visual consistency.
- Verify footer attribution still renders.
- Verify the visual language clearly signals AI model routing, not crypto transfer.
