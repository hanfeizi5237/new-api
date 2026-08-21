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
import {
  Activity,
  AlertCircle,
  ArrowUpRight,
  BarChart3,
  CreditCard,
  Download,
  FileText,
  RefreshCw,
  ShieldAlert,
  Timer,
  TrendingUp,
  Users,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatNumber, formatQuota, formatTokens } from '@/lib/format'
import { cn } from '@/lib/utils'

import { GRANULARITY_OPTIONS } from '../constants'
import type {
  ChartSpecs,
} from '../lib/charts'
import type { Granularity, UsageSummary, UserUsageOverview } from '../types'
import { VChartCard } from './vchart-card'

interface MainDashboardViewProps {
  loading: boolean
  overviewData: UserUsageOverview[]
  summary: UsageSummary
  granularity: Granularity
  dateRange: { start: string; end: string }
  charts: ChartSpecs
  loadOverview: () => void
  handleDateRangeChange: (range: { start: string; end: string }) => void
  handleGranularityChange: (g: Granularity) => void
  exportCSV: () => void
  openUserDetail: (user: UserUsageOverview) => void
}

const STAT_CARDS = [
  { key: 'totalUsers', label: 'Total Users', icon: Users, color: 'text-blue-600 bg-blue-50 dark:bg-blue-950' },
  { key: 'totalCount', label: 'Total Calls', icon: Activity, color: 'text-green-600 bg-green-50 dark:bg-green-950' },
  { key: 'totalQuota', label: 'Total Quota', icon: CreditCard, color: 'text-orange-600 bg-orange-50 dark:bg-orange-950' },
  { key: 'totalTokens', label: 'Total Tokens', icon: FileText, color: 'text-purple-600 bg-purple-50 dark:bg-purple-950' },
  { key: 'totalErrors', label: 'Total Errors', icon: AlertCircle, color: 'text-red-600 bg-red-50 dark:bg-red-950' },
] as const

export function MainDashboardView({
  loading,
  overviewData,
  summary,
  granularity,
  dateRange,
  charts,
  loadOverview,
  handleDateRangeChange,
  handleGranularityChange,
  exportCSV,
  openUserDetail,
}: MainDashboardViewProps) {
  const { t } = useTranslation()

  const diffDays = Math.ceil(
    (new Date(`${dateRange.end}T00:00:00`).getTime() -
      new Date(`${dateRange.start}T00:00:00`).getTime()) /
      (1000 * 60 * 60 * 24),
  )
  const showDailySummary = diffDays >= 1

  const summaryValues: Record<string, string | number> = {
    totalUsers: summary.totalUsers,
    totalCount: formatNumber(summary.totalCount),
    totalQuota: formatQuota(summary.totalQuota),
    totalTokens: formatTokens(summary.totalTokens),
    totalErrors: formatNumber(summary.totalErrors),
  }

  return (
    <div className='space-y-4'>
      {/* Controls */}
      <Card className='p-4'>
        <div className='flex flex-col gap-3'>
          <div className='text-lg font-semibold'>{t('User Usage Overview')}</div>
          <div className='flex flex-wrap items-center gap-3'>
            <div className='flex items-center gap-2'>
              <Input
                type='date'
                value={dateRange.start}
                onChange={(e) =>
                  handleDateRangeChange({
                    ...dateRange,
                    start: e.target.value,
                  })
                }
                className='w-[150px]'
              />
              <span className='text-muted-foreground'>~</span>
              <Input
                type='date'
                value={dateRange.end}
                onChange={(e) =>
                  handleDateRangeChange({
                    ...dateRange,
                    end: e.target.value,
                  })
                }
                className='w-[150px]'
              />
            </div>

            <div className='flex items-center gap-1.5'>
              {GRANULARITY_OPTIONS.map((opt) => (
                <Button
                  key={opt.value}
                  size='sm'
                  variant={granularity === opt.value ? 'default' : 'outline'}
                  onClick={() => handleGranularityChange(opt.value)}
                >
                  {t(opt.label)}
                </Button>
              ))}
            </div>

            <Button
              size='sm'
              onClick={loadOverview}
              disabled={loading}
            >
              <RefreshCw className={cn('size-4', loading && 'animate-spin')} />
              {t('Query')}
            </Button>
            <Button
              size='sm'
              variant='outline'
              onClick={exportCSV}
              disabled={overviewData.length === 0}
            >
              <Download className='size-4' />
              {t('Export CSV')}
            </Button>
          </div>
        </div>
      </Card>

      {/* Stat Cards */}
      <div className='grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5'>
        {STAT_CARDS.map((card) => {
          const Icon = card.icon
          return (
            <Card
              key={card.key}
              className='flex items-center gap-2 px-3 py-2'
            >
              <div className={cn('rounded-md p-1.5', card.color)}>
                <Icon className='size-4' />
              </div>
              <div className='min-w-0'>
                <div className='text-muted-foreground text-[11px]'>
                  {t(card.label)}
                </div>
                <div className='truncate text-sm font-semibold sm:text-base'>
                  {summaryValues[card.key]}
                </div>
              </div>
            </Card>
          )
        })}
      </div>

      {/* Daily Trend Charts */}
      {showDailySummary && (
        <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
          <VChartCard
            spec={charts.specDailyQuotaTrend}
            title={t('Daily Quota Trend')}
            icon={<TrendingUp className='size-4' />}
          />
          <VChartCard
            spec={charts.specDailyTokenTrend}
            title={t('Daily Token Trend')}
            icon={<TrendingUp className='size-4' />}
          />
        </div>
      )}

      {/* Ranking Charts */}
      <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
        <VChartCard
          spec={charts.specUserRank}
          title={t('User Quota Rank')}
          icon={<BarChart3 className='size-4' />}
        />
        <VChartCard
          spec={charts.specModelTrend}
          title={t('Model Usage Trend')}
          icon={<TrendingUp className='size-4' />}
        />
      </div>

      <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
        <VChartCard
          spec={charts.specCountRank}
          title={t('Call Count Rank')}
          icon={<Timer className='size-4' />}
          chartClassName='h-72 sm:h-80'
        />
        <VChartCard
          spec={charts.specErrorUserRank}
          title={t('Error User Rank')}
          icon={<ShieldAlert className='size-4' />}
          chartClassName='h-72 sm:h-80'
        />
      </div>

      {/* User List Table */}
      <Card className='p-4'>
        <div className='mb-3 text-sm font-semibold'>{t('User List')}</div>
        <div className='overflow-x-auto'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('User')}</TableHead>
                <TableHead className='text-right'>{t('Call Count')}</TableHead>
                <TableHead className='text-right'>{t('Quota')}</TableHead>
                <TableHead className='text-right'>{t('Tokens')}</TableHead>
                <TableHead className='text-right'>{t('Errors')}</TableHead>
                <TableHead className='text-right'>{t('Action')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {overviewData.length === 0 && !loading ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className='text-muted-foreground py-8 text-center'
                  >
                    {t('No data')}
                  </TableCell>
                </TableRow>
              ) : (
                overviewData.map((user) => (
                  <TableRow key={user.user_id}>
                    <TableCell>
                      <div className='flex items-center gap-2'>
                        <span className='font-medium'>
                          {user.username || `User ${user.user_id}`}
                        </span>
                        {user.display_name && (
                          <span className='text-muted-foreground text-xs'>
                            ({user.display_name})
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatNumber(user.total_count || 0)}
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatQuota(user.total_quota || 0)}
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatTokens(user.total_tokens || 0)}
                    </TableCell>
                    <TableCell className='text-right'>
                      <Badge
                        variant={
                          (user.error_count || 0) > 0 ? 'destructive' : 'secondary'
                        }
                      >
                        {user.error_count || 0}
                      </Badge>
                    </TableCell>
                    <TableCell className='text-right'>
                      <Button
                        size='sm'
                        variant='ghost'
                        onClick={() => openUserDetail(user)}
                      >
                        <ArrowUpRight className='size-3.5' />
                        {t('Detail')}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </Card>
    </div>
  )
}
