# Display Group Ratio Design

## Background

The current pricing system uses real billing configuration for both quota deduction and frontend pricing display. This creates a hard coupling between:

- actual billing behavior used during relay requests
- public pricing values shown to normal users in the pricing directory, model detail pages, and related UI surfaces

The new requirement is to decouple display pricing from actual billing without changing the real billing path. Administrators need to keep the existing real group ratio for quota deduction while configuring a separate display-only ratio for normal user-facing pricing pages.

## Goals

- Add a display-only group ratio field named `display_group_ratio`.
- Keep `group_ratio` as the real billing ratio used by the backend.
- Make all normal user-facing pricing displays use `display_group_ratio`.
- Hide the real `group_ratio` from ordinary frontend pricing responses.
- Preserve existing billing, quota pre-consume, settlement, and relay behavior.
- Keep the admin configuration workflow aligned with the existing group pricing mental model.

## Non-Goals

- No changes to actual quota deduction formulas.
- No changes to `model_ratio`, `completion_ratio`, `billing_expr`, or model-level billing semantics.
- No per-model display override in this phase.
- No per-user display pricing logic in this phase.
- No attempt to make display pricing and real billing simultaneously visible to ordinary users.

## Core Design

### Real vs display responsibilities

The pricing system will explicitly separate two concepts at the group layer:

- `group_ratio`: real billing ratio, only used for request-time billing and quota calculation
- `display_group_ratio`: display-only ratio, only used when building pricing data for ordinary user-facing pages

This keeps the existing billing path intact while allowing public pricing to present a different multiplier.

### Why group-level instead of model-level

The existing admin workflow already treats group pricing as the main place where operators control user-visible pricing differences across groups. Adding a display-only field at the same layer preserves operator expectations and avoids introducing a second pricing control surface with overlapping intent.

### Fallback behavior

If a group does not have an explicit `display_group_ratio` configured, the system should fall back to the current real `group_ratio` for display purposes. This ensures backward compatibility and avoids blank or broken pricing pages during rollout.

## Data Model

### New setting

Introduce a new group pricing setting map:

- key: group name
- value: display ratio as `float64`

Recommended option/config name:

- `DisplayGroupRatio`

Recommended runtime accessor names:

- `GetDisplayGroupRatio(group string) float64`
- `GetDisplayGroupRatioCopy() map[string]float64`
- update/reset helpers mirroring the existing `group_ratio` setting patterns

### Compatibility

- Existing deployments without `DisplayGroupRatio` continue to work.
- When missing, display code uses `group_ratio` as fallback.
- No migration of request billing data is required.

## Backend Changes

### Actual billing path

No behavior changes are allowed in the real billing path.

The following logic must keep using real `group_ratio` only:

- relay pre-consume quota calculation
- relay final quota settlement
- group auto-selection billing impact
- any logic in the request-time pricing helpers

### Public pricing API

The pricing API response used by ordinary user-facing pages should be adjusted so that:

- the response exposes display-oriented group ratio data instead of real ratio data
- ordinary pricing consumers do not receive the real `group_ratio` map

Recommended response shape change:

- keep the response field name `group_ratio` for compatibility, but fill it with display ratios only
- do not expose a separate real-ratio field in `/api/pricing`

This keeps frontend refactoring smaller while meeting the requirement that real group ratios stay hidden from normal displays.

### Admin APIs

The admin settings APIs should expose and persist `DisplayGroupRatio` so operators can configure it alongside existing group ratio settings.

The admin console may still read and edit both:

- real `group_ratio`
- display `display_group_ratio`

Because admins are trusted operators, hiding real ratios from admin-only configuration flows is not required.

## Frontend Changes

### Admin console

In the group pricing settings UI, add a new field next to the existing real ratio field:

- real ratio: actual billing multiplier
- display ratio: frontend-only multiplier

The labels must clearly distinguish these roles to avoid operator confusion.

Suggested Chinese labels:

- `真实倍率`
- `展示倍率`

### User-facing pricing surfaces

All normal user-facing places that display group-influenced model prices should use display ratios from `/api/pricing`.

This includes at minimum:

- pricing directory page
- model detail page
- group pricing breakdown tables
- any other pricing page logic that currently consumes `group_ratio` from `/api/pricing`

Because `/api/pricing` will already return display ratios in the `group_ratio` field, most frontend display code can remain structurally unchanged.

## API Contract Strategy

### Public contract

For `/api/pricing`:

- `group_ratio` means display ratio data
- it must never expose the real billing ratios to ordinary consumers

### Internal/admin contract

For admin configuration endpoints and admin settings payloads:

- preserve access to real `group_ratio`
- add access to `display_group_ratio`

This creates a deliberate distinction between:

- public display contract
- admin configuration contract

## Error Handling

- Invalid `display_group_ratio` payloads should fail validation the same way existing ratio-setting payloads fail.
- Missing `display_group_ratio` should not fail requests; fallback to real `group_ratio`.
- If a model/group combination has no usable display ratio after fallback, existing pricing-page empty-state behavior may remain unchanged.

## Security and Disclosure

The key product requirement is that ordinary users should not infer real internal billing ratios from pricing page API responses.

Therefore:

- public pricing endpoints must not expose the real group ratio map
- admin-only configuration flows may still expose real ratio values
- actual billing logs and settlement records remain unchanged and are outside this display-only requirement

## Performance

- No new request-time billing work is added to the relay path.
- Display ratio lookup is map-based and only affects pricing-page aggregation.
- No new slow path should be introduced for chat relay or quota deduction.

## Testing Strategy

### Backend

Add or update tests covering:

- display ratio lookup with fallback to real group ratio
- `/api/pricing` returning display ratios instead of real ratios
- no change in request-time billing helpers when display ratios are configured

### Frontend

Add or update tests covering:

- pricing display renders from returned display group ratios
- fallback display remains correct when display ratio is not configured
- admin form correctly loads and saves real ratio plus display ratio

### Manual acceptance

Verify all of the following:

1. Admin sets `group_ratio=1.5`, `display_group_ratio=1.0` for a group.
2. Ordinary user visits pricing pages and sees prices derived from `1.0`.
3. Actual API invocation still deducts quota as if group ratio were `1.5`.
4. Removing `display_group_ratio` makes frontend display fall back to the real ratio.
5. Admin configuration page clearly distinguishes real vs display fields.

## Implementation Notes

### Preferred rollout shape

1. add backend setting storage and accessors for `DisplayGroupRatio`
2. expose the new field in admin settings flows
3. switch `/api/pricing` to use display ratios with fallback
4. update admin UI labels and editing controls
5. verify ordinary frontend pricing displays without altering real billing logic

### Intentional compatibility choice

Reusing the public `/api/pricing` field name `group_ratio` for display ratios is intentional. It minimizes frontend churn and reduces the risk of missing a display surface that still reads the old field.

## Final Recommendation

Implement `display_group_ratio` as a group-level display-only multiplier.

This is the simplest design that:

- matches the current admin pricing mental model
- keeps actual billing unchanged
- gives operators direct control over user-visible pricing
- prevents ordinary pricing APIs from leaking real billing multipliers
