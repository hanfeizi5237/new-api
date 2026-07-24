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
import { useCallback, useRef, useState } from 'react'
import { toast } from 'sonner'

import {
  fetchGlobalTimeSeries,
  fetchGlobalTimeSeriesByModel,
  fetchUsageDetail,
  fetchUsageOverview,
} from '../api'
import { getDefaultDateRange, MAX_DATE_RANGE_DAYS } from '../constants'
import type {
  DateRange,
  DetailTab,
  Granularity,
  ModelTimeSeriesItem,
  TimeSeriesItem,
  UsageSummary,
  UserUsageDetail,
  UserUsageOverview,
} from '../types'

const pad = (n: number) => String(n).padStart(2, '0')
const formatDate = (date: Date) =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`

function getPresetDateRange(
  type: Granularity,
  endDateStr?: string,
): DateRange {
  const endDate = endDateStr ? new Date(`${endDateStr}T00:00:00`) : new Date()
  const end = new Date(endDate)
  const start = new Date(endDate)

  if (type === 'day') return { start: formatDate(end), end: formatDate(end) }
  if (type === 'week') {
    start.setDate(start.getDate() - 6)
    return { start: formatDate(start), end: formatDate(end) }
  }
  if (type === 'month') {
    start.setMonth(start.getMonth() - 1)
    return { start: formatDate(start), end: formatDate(end) }
  }
  return { start: formatDate(end), end: formatDate(end) }
}

function dateRangeToTimestamps(range: DateRange) {
  const startTs = Math.floor(new Date(`${range.start}T00:00:00`).getTime() / 1000)
  const endTs = Math.floor(new Date(`${range.end}T23:59:59`).getTime() / 1000)
  return { startTs, endTs }
}

function validateDateRange(range: DateRange): boolean {
  if (!range.start || !range.end) {
    toast.error('Please select a date range')
    return false
  }
  const start = new Date(`${range.start}T00:00:00`)
  const end = new Date(`${range.end}T00:00:00`)
  const diffDays = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)
  if (diffDays < 0) {
    toast.error('End date cannot be before start date')
    return false
  }
  if (diffDays > MAX_DATE_RANGE_DAYS) {
    toast.error(`Date range cannot exceed ${MAX_DATE_RANGE_DAYS} days`)
    return false
  }
  return true
}

export function useUserUsageData() {
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [overviewData, setOverviewData] = useState<UserUsageOverview[]>([])
  const [detailData, setDetailData] = useState<UserUsageDetail | null>(null)
  const [globalTimeSeries, setGlobalTimeSeries] = useState<TimeSeriesItem[]>([])
  const [globalTimeSeriesByModel, setGlobalTimeSeriesByModel] = useState<
    ModelTimeSeriesItem[]
  >([])
  const [dateRange, setDateRange] = useState<DateRange>(getDefaultDateRange())
  const [granularity, setGranularity] = useState<Granularity>('day')
  const [selectedUser, setSelectedUser] = useState<UserUsageOverview | null>(null)
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [activeDetailTab, setActiveDetailTab] = useState<DetailTab>('model')
  const abortControllerRef = useRef<AbortController | null>(null)

  const fetchOverview = useCallback(
    async (range: DateRange, granularityValue: Granularity) => {
      if (!validateDateRange(range)) return
      if (abortControllerRef.current) abortControllerRef.current.abort()
      const controller = new AbortController()
      abortControllerRef.current = controller

      setLoading(true)
      try {
        const { startTs, endTs } = dateRangeToTimestamps(range)
        const [overviewRes, seriesRes, seriesByModelRes] = await Promise.all([
          fetchUsageOverview(startTs, endTs, granularityValue),
          fetchGlobalTimeSeries(startTs, endTs, 'day'),
          fetchGlobalTimeSeriesByModel(startTs, endTs, 'day'),
        ])

        if (overviewRes.success) {
          setOverviewData(overviewRes.data || [])
        } else {
          toast.error(overviewRes.message || 'Failed to load data')
          setOverviewData([])
        }

        setGlobalTimeSeries(seriesRes.success ? seriesRes.data || [] : [])
        setGlobalTimeSeriesByModel(
          seriesByModelRes.success ? seriesByModelRes.data || [] : [],
        )
      } catch (err: unknown) {
        const e = err as { name?: string; message?: string }
        if (e?.name !== 'AbortError') {
          toast.error(`Failed to load overview data: ${e?.message || ''}`)
          setOverviewData([])
          setGlobalTimeSeries([])
          setGlobalTimeSeriesByModel([])
        }
      } finally {
        setLoading(false)
      }
    },
    [],
  )

  const loadOverview = useCallback(async () => {
    await fetchOverview(dateRange, granularity)
  }, [fetchOverview, dateRange, granularity])

  const loadDetail = useCallback(
    async (userId: number) => {
      if (!validateDateRange(dateRange)) return
      setDetailLoading(true)
      try {
        const { startTs, endTs } = dateRangeToTimestamps(dateRange)
        const res = await fetchUsageDetail(
          userId,
          startTs,
          endTs,
          granularity,
        )
        if (res.success) {
          setDetailData(res.data || null)
        } else {
          toast.error(res.message || 'Failed to load detail')
          setDetailData(null)
        }
      } catch (err: unknown) {
        const e = err as { message?: string }
        toast.error(`Failed to load detail data: ${e?.message || ''}`)
        setDetailData(null)
      } finally {
        setDetailLoading(false)
      }
    },
    [dateRange, granularity],
  )

  const openUserDetail = useCallback(
    async (user: UserUsageOverview) => {
      setSelectedUser(user)
      setDrawerVisible(true)
      setActiveDetailTab('model')
      await loadDetail(user.user_id)
    },
    [loadDetail],
  )

  const closeUserDetail = useCallback(() => {
    setDrawerVisible(false)
    setSelectedUser(null)
    setDetailData(null)
  }, [])

  const getSummary = useCallback((): UsageSummary => {
    const totalUsers = overviewData.length
    let totalCount = 0
    let totalQuota = 0
    let totalTokens = 0
    let totalErrors = 0
    overviewData.forEach((user) => {
      totalCount += user.total_count || 0
      totalQuota += user.total_quota || 0
      totalTokens += user.total_tokens || 0
      totalErrors += user.error_count || 0
    })
    return { totalUsers, totalCount, totalQuota, totalTokens, totalErrors }
  }, [overviewData])

  const exportCSV = useCallback(() => {
    if (overviewData.length === 0) {
      toast.error('No data to export')
      return
    }
    const headers = [
      'User',
      'Call Count',
      'Quota',
      'Token Usage',
      'Error Count',
    ]
    const rows = overviewData.map((user) => [
      user.username || `User ${user.user_id}`,
      user.total_count || 0,
      (user.total_quota / 1000000).toFixed(4),
      user.total_tokens || 0,
      user.error_count || 0,
    ])
    const summary = getSummary()
    rows.push([
      'Total',
      summary.totalCount,
      (summary.totalQuota / 1000000).toFixed(4),
      summary.totalTokens,
      summary.totalErrors,
    ])
    const csvContent = [
      headers.join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n')
    const bom = '\uFEFF'
    const blob = new Blob([bom + csvContent], {
      type: 'text/csv;charset=utf-8;',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `user_usage_${dateRange.start}_${dateRange.end}.csv`
    link.click()
    URL.revokeObjectURL(url)
    toast.success('Export successful')
  }, [overviewData, dateRange, getSummary])

  const handleDateRangeChange = useCallback((newRange: DateRange) => {
    setDateRange(newRange)
    setGranularity('day')
  }, [])

  const handleGranularityChange = useCallback(
    async (newGranularity: Granularity) => {
      const newRange = getPresetDateRange(
        newGranularity,
        dateRange?.end || getDefaultDateRange().end,
      )
      setGranularity(newGranularity)
      setDateRange(newRange)
      await fetchOverview(newRange, newGranularity)
    },
    [dateRange, fetchOverview],
  )

  return {
    loading,
    detailLoading,
    overviewData,
    detailData,
    globalTimeSeries,
    globalTimeSeriesByModel,
    dateRange,
    granularity,
    selectedUser,
    drawerVisible,
    activeDetailTab,
    loadOverview,
    loadDetail,
    openUserDetail,
    closeUserDetail,
    setActiveDetailTab,
    handleDateRangeChange,
    handleGranularityChange,
    exportCSV,
    getSummary,
  }
}
