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
import { formatNumber, formatQuota, stringToColor } from '@/lib/format'

import { ERROR_COLORS, USER_COLORS } from '../constants'
import type {
  ModelTimeSeriesItem,
  TimeSeriesItem,
  UserUsageDetail,
  UserUsageOverview,
  VChartSpec,
} from '../types'

const emptyPie = [{ type: 'No Data', value: 0 }]

const tsLabel = (ts: number) => new Date(ts * 1000).toLocaleDateString()

const tokensToMillions = (tokens: number) =>
  Number((Number(tokens || 0) / 1000000).toFixed(4))

export interface ChartSpecs {
  specUserRank: VChartSpec
  specUserTrend: VChartSpec
  specCountRank: VChartSpec
  specErrorUserRank: VChartSpec
  specDailyQuotaTrend: VChartSpec
  specDailyTokenTrend: VChartSpec
  specModelPie: VChartSpec
  specDetailModelQuotaPie: VChartSpec
  specModelRank: VChartSpec
  specDetailTimeTrend: VChartSpec
  specErrorRank: VChartSpec
  specTokenDistribution: VChartSpec
  specLatencyDistribution: VChartSpec
}

export function createEmptyChartSpecs(): ChartSpecs {
  return {
    specUserRank: {
      type: 'bar',
      data: [{ id: 'userRankData', values: [] }],
      xField: 'rawQuota',
      yField: 'User',
      seriesField: 'User',
      direction: 'horizontal',
      legends: { visible: false },
      title: { visible: true, text: 'User Quota Rank', subtext: '' },
      label: {
        visible: true,
        position: 'outside',
        formatMethod: (_v: number, d: any) =>
          formatQuota(d?.rawQuota || 0),
      },
      axes: [
        { orient: 'left', type: 'band', label: { visible: true } },
        { orient: 'bottom', type: 'linear', visible: false },
      ],
      color: { type: 'ordinal', range: USER_COLORS },
    },
    specUserTrend: {
      type: 'line',
      data: [{ id: 'userTrendData', values: [] }],
      xField: 'Time',
      yField: 'rawQuota',
      seriesField: 'User',
      legends: { visible: true, selectMode: 'single' },
      title: { visible: true, text: 'User Usage Trend', subtext: '' },
      axes: [
        { orient: 'left', label: { formatMethod: (v: number) => formatQuota(v) } },
      ],
      point: { visible: true, style: { size: 6 } },
      color: { type: 'ordinal', range: USER_COLORS },
    },
    specCountRank: {
      type: 'bar',
      data: [{ id: 'countRankData', values: [] }],
      xField: 'User',
      yField: 'Count',
      seriesField: 'User',
      legends: { visible: false },
      title: { visible: true, text: 'Call Count Rank', subtext: '' },
      color: { type: 'ordinal', range: USER_COLORS },
    },
    specErrorUserRank: {
      type: 'bar',
      data: [{ id: 'errorUserRankData', values: [] }],
      xField: 'User',
      yField: 'Errors',
      seriesField: 'User',
      legends: { visible: false },
      title: { visible: true, text: 'Error User Rank', subtext: '' },
      color: { type: 'ordinal', range: ERROR_COLORS },
    },
    specDailyQuotaTrend: {
      type: 'bar',
      data: [{ id: 'dailyQuotaTrendData', values: [] }],
      xField: 'Time',
      yField: 'value',
      seriesField: 'Model',
      stack: true,
      legends: { visible: true, selectMode: 'single' },
      title: { visible: true, text: 'Daily Quota Trend', subtext: '' },
      bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
      axes: [
        { orient: 'left', label: { formatMethod: (v: number) => formatQuota(v) } },
      ],
      tooltip: {
        mark: {
          content: [
            {
              key: (datum: any) => datum['Model'],
              value: (datum: any) => formatQuota(datum['rawQuota'] || 0),
            },
          ],
        },
        dimension: {
          content: [
            {
              key: (datum: any) => datum['Model'],
              value: (datum: any) => datum['rawQuota'] || 0,
            },
          ],
          updateContent: (array: any[]) => {
            array.sort((a, b) => b.value - a.value)
            let sum = 0
            for (let i = 0; i < array.length; i++) {
              if (array[i].key === 'Other') continue
              let value = Number.parseFloat(array[i].value)
              if (Number.isNaN(value)) value = 0
              if (array[i].datum?.TimeSum) sum = array[i].datum.TimeSum
              array[i].value = formatQuota(value)
            }
            array.unshift({ key: 'Total', value: formatQuota(sum) })
            return array
          },
        },
      },
      color: { type: 'ordinal', range: [] },
    },
    specDailyTokenTrend: {
      type: 'bar',
      data: [{ id: 'dailyTokenTrendData', values: [] }],
      xField: 'Time',
      yField: 'value',
      seriesField: 'Model',
      stack: true,
      legends: { visible: true, selectMode: 'single' },
      title: { visible: true, text: 'Daily Token Trend', subtext: '' },
      bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
      axes: [
        {
          orient: 'left',
          label: { formatMethod: (v: number) => `${Number(v || 0).toFixed(2)}M` },
        },
      ],
      tooltip: {
        mark: {
          content: [
            {
              key: (datum: any) => datum['Model'],
              value: (datum: any) => `${Number(datum['rawTokens'] || 0).toFixed(2)}M`,
            },
          ],
        },
        dimension: {
          content: [
            {
              key: (datum: any) => datum['Model'],
              value: (datum: any) =>
                `${Number(datum['rawTokens'] || 0).toFixed(2)}M`,
            },
          ],
          updateContent: (array: any[]) => {
            array.sort((a, b) => b.value - a.value)
            let sum = 0
            for (let i = 0; i < array.length; i++) {
              if (array[i].key === 'Other') continue
              let value = Number.parseFloat(array[i].value)
              if (Number.isNaN(value)) value = 0
              if (array[i].datum?.TimeSum) sum = array[i].datum.TimeSum
              array[i].value = `${Number(value).toFixed(2)}M`
            }
            array.unshift({ key: 'Total', value: `${Number(sum).toFixed(2)}M` })
            return array
          },
        },
      },
      color: { type: 'ordinal', range: [] },
    },
    specModelPie: {
      type: 'pie',
      data: [{ id: 'modelPieData', values: emptyPie }],
      outerRadius: 0.8,
      innerRadius: 0.5,
      padAngle: 0.6,
      valueField: 'value',
      categoryField: 'type',
      legends: { visible: true, orient: 'left' },
      title: { visible: true, text: 'Model Call Count', subtext: '' },
      color: { specified: {} },
    },
    specDetailModelQuotaPie: {
      type: 'pie',
      data: [{ id: 'detailQuotaPie', values: emptyPie }],
      outerRadius: 0.8,
      innerRadius: 0.5,
      padAngle: 0.6,
      valueField: 'value',
      categoryField: 'type',
      legends: { visible: true, orient: 'left' },
      title: { visible: true, text: 'Model Quota', subtext: '' },
      color: { specified: {} },
    },
    specModelRank: {
      type: 'bar',
      data: [{ id: 'modelRankData', values: [] }],
      xField: 'Model',
      yField: 'Count',
      seriesField: 'Model',
      legends: { visible: false },
      title: { visible: true, text: 'Model Call Rank', subtext: '' },
      color: { specified: {} },
    },
    specDetailTimeTrend: {
      type: 'line',
      data: [{ id: 'detailTimeTrend', values: [] }],
      xField: 'Time',
      yField: 'value',
      seriesField: 'Metric',
      legends: { visible: true, selectMode: 'single' },
      title: { visible: true, text: 'Time Trend', subtext: '' },
      axes: [{ orient: 'left' }],
      color: {
        specified: {
          Quota: '#3b82f6',
          Count: '#10b981',
          Tokens: '#8b5cf6',
        },
      },
    },
    specErrorRank: {
      type: 'bar',
      data: [{ id: 'errorRankData', values: [] }],
      xField: 'Error',
      yField: 'Count',
      seriesField: 'Error',
      legends: { visible: false },
      title: { visible: true, text: 'Error Reason Rank', subtext: '' },
      color: { type: 'ordinal', range: ERROR_COLORS },
    },
    specTokenDistribution: {
      type: 'bar',
      data: [{ id: 'tokenDistributionData', values: [] }],
      xField: 'Model',
      yField: 'Tokens',
      seriesField: 'Type',
      stack: true,
      legends: { visible: true },
      title: { visible: true, text: 'Token Distribution', subtext: '' },
      color: {
        specified: { 'Prompt Tokens': '#3b82f6', 'Completion Tokens': '#8b5cf6' },
      },
    },
    specLatencyDistribution: {
      type: 'bar',
      data: [{ id: 'latencyData', values: [] }],
      xField: 'Time',
      yField: 'Latency',
      seriesField: 'Metric',
      legends: { visible: false },
      title: { visible: true, text: 'Avg Latency', subtext: '' },
      color: { specified: { 'Avg Latency (ms)': '#f59e0b' } },
    },
  }
}

function userLabel(u: UserUsageOverview): string {
  return u.display_name || u.username || `User ${u.user_id}`
}

export function updateOverviewCharts(
  specs: ChartSpecs,
  overviewData: UserUsageOverview[],
  globalTimeSeries: TimeSeriesItem[] = [],
  globalTimeSeriesByModel: ModelTimeSeriesItem[] = [],
): ChartSpecs {
  if (!overviewData || overviewData.length === 0) return specs

  const sortedByQuota = [...overviewData].sort(
    (a, b) => (b.total_quota || 0) - (a.total_quota || 0),
  )
  const topUsersByQuota = sortedByQuota.slice(0, 10)
  const totalQuota = sortedByQuota.reduce((s, i) => s + (i.total_quota || 0), 0)

  const specUserRank: VChartSpec = {
    ...specs.specUserRank,
    data: [
      {
        id: 'userRankData',
        values: topUsersByQuota.map((u) => ({
          User: userLabel(u),
          rawQuota: u.total_quota || 0,
        })),
      },
    ],
    title: { ...specs.specUserRank.title, subtext: `Total: ${formatQuota(totalQuota)}` },
  }

  const sortedByCount = [...overviewData]
    .sort((a, b) => (b.total_count || 0) - (a.total_count || 0))
    .slice(0, 10)

  const specCountRank: VChartSpec = {
    ...specs.specCountRank,
    data: [
      {
        id: 'countRankData',
        values: sortedByCount.map((u) => ({
          User: userLabel(u),
          Count: u.total_count || 0,
        })),
      },
    ],
    title: {
      ...specs.specCountRank.title,
      subtext: `Total: ${formatNumber(overviewData.reduce((s, i) => s + (i.total_count || 0), 0))}`,
    },
  }

  const sortedByErrors = [...overviewData]
    .filter((u) => (u.error_count || 0) > 0)
    .sort((a, b) => (b.error_count || 0) - (a.error_count || 0))
    .slice(0, 10)

  const specErrorUserRank: VChartSpec = {
    ...specs.specErrorUserRank,
    data: [
      {
        id: 'errorUserRankData',
        values: sortedByErrors.map((u) => ({
          User: userLabel(u),
          Errors: u.error_count || 0,
        })),
      },
    ],
    title: {
      ...specs.specErrorUserRank.title,
      subtext: `Total: ${formatNumber(overviewData.reduce((s, i) => s + (i.error_count || 0), 0))}`,
    },
  }

  const trendMap = new Map<string, any>()
  topUsersByQuota.forEach((u) => {
    ;(u.time_series || []).forEach((point) => {
      const key = `${point.timestamp}-${u.username}`
      trendMap.set(key, {
        Time: tsLabel(point.timestamp),
        User: userLabel(u),
        rawQuota: point.quota || 0,
      })
    })
  })

  const specUserTrend: VChartSpec = {
    ...specs.specUserTrend,
    data: [{ id: 'userTrendData', values: [...trendMap.values()] }],
    title: { ...specs.specUserTrend.title, subtext: `Top ${topUsersByQuota.length} Users` },
  }

  // Build model-level stacked bar data for daily trends
  const quotaByModelValues: any[] = []
  const tokenByModelValues: any[] = []
  const modelColors: Record<string, string> = {}

  if (globalTimeSeriesByModel && globalTimeSeriesByModel.length > 0) {
    const modelSet = new Set<string>()
    globalTimeSeriesByModel.forEach((p) => {
      const time = tsLabel(p.timestamp)
      const model = p.model_name || 'Unknown'
      modelSet.add(model)
      quotaByModelValues.push({
        Time: time,
        Model: model,
        rawQuota: p.quota || 0,
        value: p.quota || 0,
      })
      tokenByModelValues.push({
        Time: time,
        Model: model,
        rawTokens: tokensToMillions(p.tokens || 0),
        value: tokensToMillions(p.tokens || 0),
      })
    })
    modelSet.forEach((m) => {
      modelColors[m] = stringToColor(m)
    })
  } else {
    ;(globalTimeSeries || []).forEach((p) => {
      const time = tsLabel(p.timestamp)
      quotaByModelValues.push({
        Time: time,
        Model: 'Total',
        rawQuota: p.quota || 0,
        value: p.quota || 0,
      })
      tokenByModelValues.push({
        Time: time,
        Model: 'Total',
        rawTokens: tokensToMillions(p.tokens || 0),
        value: tokensToMillions(p.tokens || 0),
      })
    })
    modelColors['Total'] = '#3b82f6'
  }

  // Add TimeSum for dimension tooltip
  const quotaTimeMap = new Map<string, number>()
  quotaByModelValues.forEach((d) => {
    quotaTimeMap.set(d.Time, (quotaTimeMap.get(d.Time) || 0) + (d.rawQuota || 0))
  })
  quotaByModelValues.forEach((d) => {
    d.TimeSum = quotaTimeMap.get(d.Time) || 0
  })

  const tokenTimeMap = new Map<string, number>()
  tokenByModelValues.forEach((d) => {
    tokenTimeMap.set(d.Time, (tokenTimeMap.get(d.Time) || 0) + (d.rawTokens || 0))
  })
  tokenByModelValues.forEach((d) => {
    d.TimeSum = tokenTimeMap.get(d.Time) || 0
  })

  const specDailyQuotaTrend: VChartSpec = {
    ...specs.specDailyQuotaTrend,
    data: [{ id: 'dailyQuotaTrendData', values: quotaByModelValues }],
    color: { type: 'ordinal', range: Object.values(modelColors) },
  }

  const specDailyTokenTrend: VChartSpec = {
    ...specs.specDailyTokenTrend,
    data: [{ id: 'dailyTokenTrendData', values: tokenByModelValues }],
    color: { type: 'ordinal', range: Object.values(modelColors) },
  }

  return {
    ...specs,
    specUserRank,
    specUserTrend,
    specCountRank,
    specErrorUserRank,
    specDailyQuotaTrend,
    specDailyTokenTrend,
  }
}

export function updateDetailCharts(
  specs: ChartSpecs,
  detailData: UserUsageDetail,
): ChartSpecs {
  if (!detailData) return specs

  const modelDistribution = detailData.model_distribution || []
  const timeDistribution = detailData.time_distribution || []
  const errorDistribution = detailData.error_distribution || []
  const colors: Record<string, string> = {}
  modelDistribution.forEach((m) => {
    colors[m.model_name] = stringToColor(m.model_name || 'Unknown')
  })

  const specModelPie: VChartSpec = {
    ...specs.specModelPie,
    data: [
      {
        id: 'modelPieData',
        values:
          modelDistribution.length > 0
            ? modelDistribution.map((m) => ({
                type: m.model_name,
                value: m.count || 0,
              }))
            : emptyPie,
      },
    ],
    title: {
      ...specs.specModelPie.title,
      subtext: `Total: ${formatNumber(modelDistribution.reduce((s, m) => s + (m.count || 0), 0))}`,
    },
    color: { specified: colors },
  }

  const specDetailModelQuotaPie: VChartSpec = {
    ...specs.specDetailModelQuotaPie,
    data: [
      {
        id: 'detailQuotaPie',
        values:
          modelDistribution.length > 0
            ? modelDistribution.map((m) => ({
                type: m.model_name,
                value: m.quota || 0,
              }))
            : emptyPie,
      },
    ],
    title: {
      ...specs.specDetailModelQuotaPie.title,
      subtext: `Total: ${formatQuota(modelDistribution.reduce((s, m) => s + (m.quota || 0), 0))}`,
    },
    color: { specified: colors },
  }

  const specModelRank: VChartSpec = {
    ...specs.specModelRank,
    data: [
      {
        id: 'modelRankData',
        values: modelDistribution.map((m) => ({
          Model: m.model_name,
          Count: m.count || 0,
        })),
      },
    ],
    color: { specified: colors },
  }

  const timeValues: any[] = []
  const latencyValues: any[] = []
  timeDistribution.forEach((t) => {
    const time = tsLabel(t.timestamp)
    timeValues.push({ Time: time, Metric: 'Quota', value: t.quota || 0 })
    timeValues.push({ Time: time, Metric: 'Count', value: t.count || 0 })
    timeValues.push({ Time: time, Metric: 'Tokens', value: t.tokens || 0 })
    latencyValues.push({ Time: time, Metric: 'Avg Latency (ms)', Latency: t.avg_use_ms || 0 })
  })

  const specDetailTimeTrend: VChartSpec = {
    ...specs.specDetailTimeTrend,
    data: [{ id: 'detailTimeTrend', values: timeValues }],
  }

  const specLatencyDistribution: VChartSpec = {
    ...specs.specLatencyDistribution,
    data: [{ id: 'latencyData', values: latencyValues }],
  }

  const tokenValues: any[] = []
  modelDistribution.forEach((m) => {
    tokenValues.push({
      Model: m.model_name,
      Type: 'Prompt Tokens',
      Tokens: m.prompt_tokens || 0,
    })
    tokenValues.push({
      Model: m.model_name,
      Type: 'Completion Tokens',
      Tokens: m.completion_tokens || 0,
    })
  })

  const specTokenDistribution: VChartSpec = {
    ...specs.specTokenDistribution,
    data: [{ id: 'tokenDistributionData', values: tokenValues }],
  }

  const specErrorRank: VChartSpec = {
    ...specs.specErrorRank,
    data: [
      {
        id: 'errorRankData',
        values: errorDistribution.map((e) => ({
          Error: e.error_content || 'Unknown Error',
          Count: e.count || 0,
        })),
      },
    ],
    title: {
      ...specs.specErrorRank.title,
      subtext: `Total: ${formatNumber(errorDistribution.reduce((s, e) => s + (e.count || 0), 0))}`,
    },
  }

  return {
    ...specs,
    specModelPie,
    specDetailModelQuotaPie,
    specModelRank,
    specDetailTimeTrend,
    specLatencyDistribution,
    specTokenDistribution,
    specErrorRank,
  }
}
