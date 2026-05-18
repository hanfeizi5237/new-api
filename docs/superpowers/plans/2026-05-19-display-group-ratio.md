# Display Group Ratio Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a group-level `display_group_ratio` setting so ordinary user-facing pricing pages use display-only ratios while real quota deduction continues to use the existing `group_ratio`.

**Architecture:** Keep the billing path unchanged and add a second group-ratio map dedicated to display. Persist the new setting through the existing option/config pipeline, expose it to admin settings, and swap the public `/api/pricing` response to return display ratios with fallback to real ratios.

**Tech Stack:** Go, Gin, project config/option system, React 19, TypeScript, React Hook Form, Zod, Bun, Rsbuild.

---

## File Structure Map

**Backend runtime and config**
- Modify: `setting/ratio_setting/group_ratio.go`
  - add `DisplayGroupRatio` storage, default handling, validation, and accessors
- Modify: `model/option.go`
  - register the new option in `OptionMap` and wire update handling
- Modify: `controller/option.go`
  - validate `DisplayGroupRatio` payloads during option updates
- Modify: `controller/pricing.go`
  - replace public pricing response `group_ratio` payload with display ratios plus fallback
- Create: `setting/ratio_setting/group_ratio_test.go`
  - cover display-ratio fallback and validation
- Create: `controller/pricing_test.go`
  - verify `/api/pricing` returns display ratios instead of real ratios

**Admin frontend**
- Modify: `web/default/src/features/system-settings/models/index.tsx`
  - include `DisplayGroupRatio` in group defaults
- Modify: `web/default/src/features/system-settings/models/ratio-settings-card.tsx`
  - add schema, default normalization, change tracking, and save payload
- Modify: `web/default/src/features/system-settings/models/group-ratio-form.tsx`
  - add text field wiring for `DisplayGroupRatio`
- Modify: `web/default/src/features/system-settings/models/group-ratio-visual-editor.tsx`
  - add editable display-ratio column in the visual group pricing table

**User-facing frontend**
- Keep behavior by contract compatibility through `/api/pricing`
- Verify affected readers still consume `group_ratio` from public payload:
  - `web/default/src/features/pricing/lib/price.ts`
  - `web/default/src/features/pricing/lib/dynamic-price.ts`
  - `web/default/src/features/pricing/components/model-details.tsx`
  - `web/default/src/features/pricing/components/pricing-sidebar.tsx`
  - `web/default/src/features/pricing/components/pricing-columns.tsx`
  - `web/default/src/features/pricing/components/model-card.tsx`

**Verification commands**
- Backend targeted tests: `go test ./setting/ratio_setting ./controller`
- Frontend typecheck: `cd web/default && bun run typecheck`
- Frontend build check: `cd web/default && bun run build`

## Task 1: Add Backend Display Ratio Storage

**Files:**
- Modify: `setting/ratio_setting/group_ratio.go`
- Modify: `model/option.go`
- Modify: `controller/option.go`
- Test: `setting/ratio_setting/group_ratio_test.go`

- [ ] **Step 1: Write the failing backend tests for display ratio fallback and validation**

```go
package ratio_setting

import "testing"

func TestGetDisplayGroupRatioFallsBackToRealRatio(t *testing.T) {
	_ = UpdateGroupRatioByJSONString(`{"default":1.5}`)
	_ = UpdateDisplayGroupRatioByJSONString(`{}`)

	if got := GetDisplayGroupRatio("default"); got != 1.5 {
		t.Fatalf("expected fallback display ratio 1.5, got %v", got)
	}
}

func TestGetDisplayGroupRatioUsesExplicitValue(t *testing.T) {
	_ = UpdateGroupRatioByJSONString(`{"default":1.5}`)
	_ = UpdateDisplayGroupRatioByJSONString(`{"default":1.0}`)

	if got := GetDisplayGroupRatio("default"); got != 1.0 {
		t.Fatalf("expected explicit display ratio 1.0, got %v", got)
	}
}

func TestCheckDisplayGroupRatioRejectsNegativeValue(t *testing.T) {
	if err := CheckDisplayGroupRatio(`{"default":-1}`); err == nil {
		t.Fatal("expected negative display ratio validation error")
	}
}
```

- [ ] **Step 2: Run the targeted backend tests to confirm they fail before implementation**

Run: `go test ./setting/ratio_setting -run 'Test(GetDisplayGroupRatio|CheckDisplayGroupRatio)'`
Expected: FAIL with undefined `UpdateDisplayGroupRatioByJSONString`, `GetDisplayGroupRatio`, or `CheckDisplayGroupRatio`.

- [ ] **Step 3: Add `DisplayGroupRatio` storage and helpers in `group_ratio.go`**

```go
type GroupRatioSetting struct {
	GroupRatio              *types.RWMap[string, float64]            `json:"group_ratio"`
	DisplayGroupRatio       *types.RWMap[string, float64]            `json:"display_group_ratio"`
	GroupGroupRatio         *types.RWMap[string, map[string]float64] `json:"group_group_ratio"`
	GroupSpecialUsableGroup *types.RWMap[string, map[string]string]  `json:"group_special_usable_group"`
}

var displayGroupRatioMap = types.NewRWMap[string, float64]()

func GetDisplayGroupRatioCopy() map[string]float64 {
	return displayGroupRatioMap.ReadAll()
}

func DisplayGroupRatio2JSONString() string {
	return displayGroupRatioMap.MarshalJSONString()
}

func UpdateDisplayGroupRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(displayGroupRatioMap, jsonStr)
}

func GetDisplayGroupRatio(name string) float64 {
	if ratio, ok := displayGroupRatioMap.Get(name); ok {
		return ratio
	}
	return GetGroupRatio(name)
}

func CheckDisplayGroupRatio(jsonStr string) error {
	checkGroupRatio := make(map[string]float64)
	if err := json.Unmarshal([]byte(jsonStr), &checkGroupRatio); err != nil {
		return err
	}
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("display group ratio must be not less than 0: " + name)
		}
	}
	return nil
}
```

- [ ] **Step 4: Register the new option in `model/option.go`**

```go
common.OptionMap["DisplayGroupRatio"] = ratio_setting.DisplayGroupRatio2JSONString()
```

```go
case "DisplayGroupRatio":
	err = ratio_setting.UpdateDisplayGroupRatioByJSONString(value)
```

- [ ] **Step 5: Add option validation in `controller/option.go`**

```go
case "DisplayGroupRatio":
	err = ratio_setting.CheckDisplayGroupRatio(option.Value.(string))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
```

- [ ] **Step 6: Re-run the targeted backend tests and package tests**

Run: `go test ./setting/ratio_setting`
Expected: PASS for the new display-ratio tests.

- [ ] **Step 7: Commit the backend option plumbing**

```bash
git add setting/ratio_setting/group_ratio.go setting/ratio_setting/group_ratio_test.go model/option.go controller/option.go
git commit -m "feat: add display group ratio setting"
```

## Task 2: Switch Public Pricing API to Display Ratios

**Files:**
- Modify: `controller/pricing.go`
- Test: `controller/pricing_test.go`

- [ ] **Step 1: Write the failing pricing API test**

```go
func TestGetPricingUsesDisplayGroupRatio(t *testing.T) {
	ratio_setting.UpdateGroupRatioByJSONString(`{"default":1.5}`)
	ratio_setting.UpdateDisplayGroupRatioByJSONString(`{"default":1.0}`)

	r := gin.New()
	r.GET("/pricing", GetPricing)

	req := httptest.NewRequest(http.MethodGet, "/pricing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), `"group_ratio":{"default":1`) {
		t.Fatalf("expected display ratio in response, got %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), `1.5`) {
		t.Fatalf("expected real group ratio to stay hidden, got %s", w.Body.String())
	}
}
```

- [ ] **Step 2: Run the controller test to verify it fails**

Run: `go test ./controller -run TestGetPricingUsesDisplayGroupRatio`
Expected: FAIL because `/api/pricing` still emits the real group ratio map.

- [ ] **Step 3: Add a helper in `controller/pricing.go` to build display ratios with fallback**

```go
func buildDisplayGroupRatioMap(userGroup string) map[string]float64 {
	display := map[string]float64{}
	for group, ratio := range ratio_setting.GetGroupRatioCopy() {
		display[group] = ratio_setting.GetDisplayGroupRatio(group)
		if userGroup != "" {
			if override, ok := ratio_setting.GetGroupGroupRatio(userGroup, group); ok {
				display[group] = override
			}
		}
		if display[group] == 0 && ratio != 0 {
			display[group] = ratio
		}
	}
	return display
}
```

- [ ] **Step 4: Replace public pricing response population to use display ratios**

```go
groupRatio := buildDisplayGroupRatioMap(group)
usableGroup = service.GetUserUsableGroups(group)
for groupName := range groupRatio {
	if _, ok := usableGroup[groupName]; !ok {
		delete(groupRatio, groupName)
	}
}
```

Note: keep the JSON field name `group_ratio` unchanged for compatibility, but ensure its contents are display-only values.

- [ ] **Step 5: Re-run controller tests and the combined backend verification set**

Run: `go test ./controller ./setting/ratio_setting`
Expected: PASS, including the new `/pricing` behavior.

- [ ] **Step 6: Commit the public pricing API change**

```bash
git add controller/pricing.go controller/pricing_test.go
git commit -m "feat: expose display ratios in pricing api"
```

## Task 3: Add Admin Console Support for `DisplayGroupRatio`

**Files:**
- Modify: `web/default/src/features/system-settings/models/index.tsx`
- Modify: `web/default/src/features/system-settings/models/ratio-settings-card.tsx`
- Modify: `web/default/src/features/system-settings/models/group-ratio-form.tsx`
- Modify: `web/default/src/features/system-settings/models/group-ratio-visual-editor.tsx`

- [ ] **Step 1: Add `DisplayGroupRatio` to the frontend group defaults and schema**

```ts
const groupSchema = z.object({
  GroupRatio: z.string().superRefine(/* existing JSON validation */),
  DisplayGroupRatio: z.string().superRefine((value, ctx) => {
    const result = validateJsonString(value)
    if (!result.valid) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: result.message || 'Invalid JSON',
      })
    }
  }),
  TopupGroupRatio: z.string().superRefine(/* existing JSON validation */),
})
```

```ts
export const defaultModelSettings = {
  TopupGroupRatio: '',
  GroupRatio: '',
  DisplayGroupRatio: '',
  UserUsableGroups: '',
  GroupGroupRatio: '',
}
```

- [ ] **Step 2: Run TypeScript typecheck to confirm the new field is not yet fully wired**

Run: `cd web/default && bun run typecheck`
Expected: FAIL with missing `DisplayGroupRatio` references in form values or props until the rest of the UI is updated.

- [ ] **Step 3: Wire `DisplayGroupRatio` through `ratio-settings-card.tsx` save/load flow**

```ts
const groupNormalizedDefaults = useRef({
  GroupRatio: normalizeJsonString(groupDefaults.GroupRatio),
  DisplayGroupRatio: normalizeJsonString(groupDefaults.DisplayGroupRatio),
  TopupGroupRatio: normalizeJsonString(groupDefaults.TopupGroupRatio),
  UserUsableGroups: normalizeJsonString(groupDefaults.UserUsableGroups),
  GroupGroupRatio: normalizeJsonString(groupDefaults.GroupGroupRatio),
})
```

```ts
await Promise.all([
  updateOption({ key: 'GroupRatio', value: normalizeJsonString(values.GroupRatio) }),
  updateOption({
    key: 'DisplayGroupRatio',
    value: normalizeJsonString(values.DisplayGroupRatio),
  }),
  updateOption({ key: 'TopupGroupRatio', value: normalizeJsonString(values.TopupGroupRatio) }),
])
```

- [ ] **Step 4: Add the new textarea field and visual-editor prop in `group-ratio-form.tsx`**

```tsx
type GroupRatioFormValues = {
  GroupRatio: string
  DisplayGroupRatio: string
  TopupGroupRatio: string
  UserUsableGroups: string
  GroupGroupRatio: string
}
```

```tsx
<GroupRatioVisualEditor
  groupRatio={form.watch('GroupRatio')}
  displayGroupRatio={form.watch('DisplayGroupRatio')}
  topupGroupRatio={form.watch('TopupGroupRatio')}
  userUsableGroups={form.watch('UserUsableGroups')}
  groupGroupRatio={form.watch('GroupGroupRatio')}
  autoGroups={form.watch('AutoGroups')}
  onChange={form.setValue}
/>
```

- [ ] **Step 5: Add a display-ratio column to `group-ratio-visual-editor.tsx`**

```ts
type GroupPricingRow = {
  _id: string
  name: string
  ratio: number
  displayRatio: number
  selectable: boolean
  description: string
}
```

```ts
function serializeGroupPricingRows(rows: GroupPricingRow[]) {
  const groupRatio: Record<string, number> = {}
  const displayGroupRatio: Record<string, number> = {}
  const userUsableGroups: Record<string, string> = {}

  for (const row of rows) {
    const name = row.name.trim()
    if (!name) continue
    groupRatio[name] = normalizeRatio(row.ratio)
    displayGroupRatio[name] = normalizeRatio(row.displayRatio)
    if (row.selectable) {
      userUsableGroups[name] = row.description
    }
  }

  return {
    GroupRatio: JSON.stringify(groupRatio, null, 2),
    DisplayGroupRatio: JSON.stringify(displayGroupRatio, null, 2),
    UserUsableGroups: JSON.stringify(userUsableGroups, null, 2),
  }
}
```

UI label guidance:

```tsx
<TableHead>{t('Real Ratio')}</TableHead>
<TableHead>{t('Display Ratio')}</TableHead>
```

- [ ] **Step 6: Re-run frontend typecheck after full admin wiring**

Run: `cd web/default && bun run typecheck`
Expected: PASS.

- [ ] **Step 7: Commit the admin console support**

```bash
git add web/default/src/features/system-settings/models/index.tsx web/default/src/features/system-settings/models/ratio-settings-card.tsx web/default/src/features/system-settings/models/group-ratio-form.tsx web/default/src/features/system-settings/models/group-ratio-visual-editor.tsx
git commit -m "feat: add display group ratio admin controls"
```

## Task 4: Verify User-Facing Pricing Surfaces and i18n

**Files:**
- Modify: `web/default/src/i18n/locales/zh.json`
- Modify: `web/default/src/i18n/locales/en.json`
- Verify readers:
  - `web/default/src/features/pricing/lib/price.ts`
  - `web/default/src/features/pricing/lib/dynamic-price.ts`
  - `web/default/src/features/pricing/components/model-details.tsx`
  - `web/default/src/features/pricing/components/pricing-sidebar.tsx`

- [ ] **Step 1: Add any missing i18n labels for the new admin fields**

```json
{
  "Real Ratio": "真实倍率",
  "Display Ratio": "展示倍率",
  "Display-only group ratio used for public pricing pages.": "仅用于前台价格展示的分组倍率。"
}
```

- [ ] **Step 2: Verify the public pricing readers still work through the existing `group_ratio` contract**

Check these call sites and confirm no code changes are required beyond the API contract swap:

```ts
const groupRatio = model.group_ratio || {}
const minRatio = getMinGroupRatio(enableGroups, groupRatio)
```

Expected locations:
- `web/default/src/features/pricing/lib/price.ts`
- `web/default/src/features/pricing/lib/dynamic-price.ts`
- `web/default/src/features/pricing/components/pricing-sidebar.tsx`
- `web/default/src/features/pricing/components/model-details.tsx`

If any surface reads a real ratio from a different payload, update it to stay on the public display contract.

- [ ] **Step 3: Run frontend build verification**

Run: `cd web/default && bun run build`
Expected: PASS.

- [ ] **Step 4: Run the final combined verification set**

Run: `go test ./setting/ratio_setting ./controller && cd web/default && bun run typecheck && bun run build`
Expected: all commands PASS.

- [ ] **Step 5: Perform manual acceptance in the app**

Manual checklist:
1. In admin settings, set `GroupRatio` to `{"default":1.5}`.
2. Set `DisplayGroupRatio` to `{"default":1.0}`.
3. Save settings and refresh the pricing page.
4. Confirm the sidebar group chip shows `x1` instead of `x1.5`.
5. Confirm model list and model detail prices use the `1.0` display ratio.
6. Make a real API request under the same group and confirm backend usage/quota still reflects `1.5`.

- [ ] **Step 6: Commit the user-facing and localization changes**

```bash
git add web/default/src/i18n/locales/en.json web/default/src/i18n/locales/zh.json
git commit -m "feat: use display group ratios in pricing views"
```

## Spec Coverage Self-Review

- `display_group_ratio` setting added: covered by Task 1 and Task 3.
- actual billing path unchanged: protected by Task 1 scope and Task 2 verification.
- `/api/pricing` hides real group ratios: covered by Task 2.
- admin can configure both real and display ratios: covered by Task 3.
- ordinary pricing pages display only the public display ratio: covered by Task 4.
- fallback to real ratio when missing: covered by Task 1 tests and Task 2 runtime behavior.

## Placeholder Scan Self-Review

- No `TODO` or `TBD` placeholders remain.
- Every code-touching task names exact files.
- Every verification step lists concrete commands and expected outcomes.

## Type Consistency Self-Review

- backend naming uses `DisplayGroupRatio`, `GetDisplayGroupRatio`, `GetDisplayGroupRatioCopy`, `UpdateDisplayGroupRatioByJSONString`, and `CheckDisplayGroupRatio`
- frontend naming uses `DisplayGroupRatio` in settings payloads and `displayGroupRatio` in component props
- public API compatibility intentionally preserves the JSON field name `group_ratio`
