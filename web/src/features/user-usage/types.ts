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

// ============================================================================
// API Response Types (matching dto/user_usage.go)
// ============================================================================

export interface TimeSeriesItem {
  timestamp: number
  count: number
  quota: number
  tokens: number
  avg_use_ms: number
}

export interface ModelTimeSeriesItem {
  timestamp: number
  model_name: string
  quota: number
  tokens: number
}

export interface UserUsageOverview {
  user_id: number
  username: string
  display_name: string
  total_count: number
  total_quota: number
  total_tokens: number
  error_count: number
  time_series: TimeSeriesItem[]
}

export interface UserUsageSummary {
  user_id: number
  username: string
  display_name: string
  total_count: number
  total_quota: number
  total_tokens: number
  error_count: number
  avg_use_time_ms: number
}

export interface ModelDistribution {
  model_name: string
  count: number
  quota: number
  prompt_tokens: number
  completion_tokens: number
  error_count: number
}

export interface ErrorDistribution {
  model_name: string
  error_content: string
  count: number
  latest_at: number
}

export interface UserUsageDetail {
  summary: UserUsageSummary
  model_distribution: ModelDistribution[]
  time_distribution: TimeSeriesItem[]
  error_distribution: ErrorDistribution[]
}

// ============================================================================
// API Response Wrappers
// ============================================================================

export interface ApiResponse<T> {
  success: boolean
  message?: string
  data?: T
}

// ============================================================================
// Domain Types
// ============================================================================

export type Granularity = 'day' | 'week' | 'month'

export interface DateRange {
  start: string
  end: string
}

export interface UsageSummary {
  totalUsers: number
  totalCount: number
  totalQuota: number
  totalTokens: number
  totalErrors: number
}

export type DetailTab = 'model' | 'time' | 'errors'

// VChart spec is a loosely-typed object
export type VChartSpec = Record<string, any>
