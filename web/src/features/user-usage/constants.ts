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
import type { DetailTab, Granularity } from './types'

export const MAX_DATE_RANGE_DAYS = 31
export const USER_USAGE_PAGE_SIZE = 20

export const GRANULARITY_OPTIONS: { label: string; value: Granularity }[] = [
  { label: 'Day', value: 'day' },
  { label: 'Last 7 Days', value: 'week' },
  { label: 'Month', value: 'month' },
]

export const DETAIL_TABS: { key: DetailTab; label: string }[] = [
  { key: 'model', label: 'Model Distribution' },
  { key: 'time', label: 'Time Trend' },
  { key: 'errors', label: 'Error Stats' },
]

export const USER_COLORS = [
  '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
  '#ec4899', '#06b6d4', '#f97316', '#6366f1', '#14b8a6',
]

export const ERROR_COLORS = ['#ef4444', '#f97316', '#f59e0b', '#8b5cf6']

export function getDefaultDateRange(): { start: string; end: string } {
  const now = new Date()
  const today = now.toISOString().split('T')[0]
  return { start: today, end: today }
}
