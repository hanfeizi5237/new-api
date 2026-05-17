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
export const guideIds = [
  'cc-switch',
  'cherry-studio',
  'openclaw',
  'claude-code',
  'codex-cli',
  'pricing-usage',
] as const

export type GuideId = (typeof guideIds)[number]

export type GuideStep = {
  title: string
  description: string
  code?: string
  note?: string
}

export type GuideTroubleshooting = {
  title: string
  content: string
}

export type GuidePricingRow = {
  model: string
  input: string
  output: string
  cacheRead: string
}

export type GuidePricing = {
  unit: string
  rows: GuidePricingRow[]
  notes: string[]
}

export type UsageGuide = {
  id: GuideId
  title: string
  shortTitle: string
  description: string
  summary: string
  officialUrl?: string
  tags: string[]
  recommendedFor: string[]
  prerequisites: string[]
  steps: GuideStep[]
  pricing?: GuidePricing
  verification: string[]
  troubleshooting: GuideTroubleshooting[]
}
