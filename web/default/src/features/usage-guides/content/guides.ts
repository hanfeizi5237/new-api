/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { ccSwitchGuide } from './guides/cc-switch'
import { cherryStudioGuide } from './guides/cherry-studio'
import { openclawGuide } from './guides/openclaw'
import { claudeCodeGuide } from './guides/claude-code'
import { codexCliGuide } from './guides/codex-cli'
import { pricingUsageGuide } from './guides/pricing-usage'
import { guideIds } from './types'
import type { GuideId, GuideStep, GuideTroubleshooting, UsageGuide } from './types'

export type { GuideId, GuideStep, GuideTroubleshooting, UsageGuide }

export { guideIds }

export const usageGuides: UsageGuide[] = [
  ccSwitchGuide,
  cherryStudioGuide,
  openclawGuide,
  claudeCodeGuide,
  codexCliGuide,
  pricingUsageGuide,
]

export function isGuideId(value: string | undefined): value is GuideId {
  if (!value) return false
  return guideIds.includes(value as GuideId)
}

export function getGuideById(guideId: string | undefined): UsageGuide {
  if (!isGuideId(guideId)) {
    return usageGuides[0]
  }
  return usageGuides.find((guide) => guide.id === guideId) ?? usageGuides[0]
}
