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
  AlertTriangle,
  BarChart3,
  CalendarRange,
  FileBarChart,
  PieChart,
  Timer,
  TrendingUp,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { formatNumber, formatQuota, formatTokens } from '@/lib/format'

import { DETAIL_TABS } from '../constants'
import type { ChartSpecs } from '../lib/charts'
import type { DetailTab, UserUsageDetail, UserUsageOverview } from '../types'
import { VChartCard } from './vchart-card'

function renderTabContent(
  loading: boolean,
  data: unknown,
  loadingText: string,
  noDataText: string,
  topContent: React.ReactNode,
  bottomContent: React.ReactNode,
): React.ReactNode {
  if (loading) {
    return (
      <div className='text-muted-foreground py-20 text-center'>
        {loadingText}
      </div>
    )
  }
  if (!data) {
    return (
      <div className='text-muted-foreground py-20 text-center'>{noDataText}</div>
    )
  }
  return (
    <>
      {topContent}
      {bottomContent}
    </>
  )
}

interface UserDetailViewProps {
  drawerVisible: boolean
  closeUserDetail: () => void
  detailLoading: boolean
  detailData: UserUsageDetail | null
  selectedUser: UserUsageOverview | null
  activeDetailTab: DetailTab
  setActiveDetailTab: (tab: DetailTab) => void
  charts: ChartSpecs
  dateRange: { start: string; end: string }
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <Card className='p-3'>
      <div className='text-center'>
        <div className='text-muted-foreground text-xs'>{label}</div>
        <div className='mt-1 text-2xl font-semibold'>{value}</div>
      </div>
    </Card>
  )
}

export function UserDetailView({
  drawerVisible,
  closeUserDetail,
  detailLoading,
  detailData,
  selectedUser,
  activeDetailTab,
  setActiveDetailTab,
  charts,
  dateRange,
}: UserDetailViewProps) {
  const { t } = useTranslation()

  if (!selectedUser) return null

  const { summary, model_distribution, time_distribution, error_distribution } =
    detailData || {}

  const userName =
    selectedUser.username || `User ${selectedUser.user_id}`

  return (
    <Sheet open={drawerVisible} onOpenChange={(open) => !open && closeUserDetail()}>
      <SheetContent
        side='right'
        className='flex w-full flex-col gap-0 overflow-y-auto sm:max-w-[1100px]'
      >
        <SheetHeader className='border-b'>
          <SheetTitle className='flex items-center gap-2'>
            <BarChart3 className='size-4' />
            <span>
              {userName}
              {selectedUser.display_name && (
                <span className='text-muted-foreground ml-2 text-sm font-normal'>
                  ({selectedUser.display_name})
                </span>
              )}
              {' - '}
              {t('Usage Detail')}
            </span>
          </SheetTitle>
        </SheetHeader>

        <div className='flex-1 space-y-4 p-4'>
          {summary && (
            <>
              <Card className='mb-2 p-3'>
                <div className='text-muted-foreground flex items-center gap-2 text-sm'>
                  <CalendarRange className='size-4' />
                  <span>
                    {t('Date Range')}: {dateRange?.start} ~ {dateRange?.end}
                  </span>
                </div>
              </Card>
              <div className='mb-4 grid grid-cols-2 gap-3 lg:grid-cols-5'>
                <MetricCard
                  label={t('Call Count')}
                  value={formatNumber(summary.total_count || 0)}
                />
                <MetricCard
                  label={t('Quota')}
                  value={formatQuota(summary.total_quota || 0)}
                />
                <MetricCard
                  label={t('Tokens')}
                  value={formatTokens(summary.total_tokens || 0)}
                />
                <MetricCard
                  label={t('Avg Latency')}
                  value={`${((summary.avg_use_time_ms || 0) / 1000).toFixed(2)}s`}
                />
                <MetricCard
                  label={t('Errors')}
                  value={formatNumber(summary.error_count || 0)}
                />
              </div>
            </>
          )}

          <Tabs
            value={activeDetailTab}
            onValueChange={(v) => setActiveDetailTab(v as DetailTab)}
          >
            <TabsList>
              {DETAIL_TABS.map((tab) => (
                <TabsTrigger key={tab.key} value={tab.key}>
                  {t(tab.label)}
                </TabsTrigger>
              ))}
            </TabsList>

            <TabsContent value='model' className='mt-4 space-y-4'>
              {renderTabContent(
                detailLoading,
                detailData,
                t('Loading...'),
                t('No data'),
                <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                  <VChartCard
                    spec={charts.specModelPie}
                    title={t('Model Call Count')}
                    icon={<PieChart className='size-4' />}
                  />
                  <VChartCard
                    spec={charts.specDetailModelQuotaPie}
                    title={t('Model Quota')}
                    icon={<PieChart className='size-4' />}
                  />
                </div>
                ,
                <>
                  <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                    <VChartCard
                      spec={charts.specModelRank}
                      title={t('Model Call Rank')}
                    />
                    <VChartCard
                      spec={charts.specTokenDistribution}
                      title={t('Token Distribution')}
                      icon={<FileBarChart className='size-4' />}
                    />
                  </div>
                  <ModelDistributionTable
                    modelDistribution={model_distribution || []}
                  />
                </>,
              )}
            </TabsContent>

            <TabsContent value='time' className='mt-4 space-y-4'>
              {renderTabContent(
                detailLoading,
                detailData,
                t('Loading...'),
                t('No data'),
                <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                  <VChartCard
                    spec={charts.specDetailTimeTrend}
                    title={t('Time Trend')}
                    icon={<TrendingUp className='size-4' />}
                  />
                  <VChartCard
                    spec={charts.specLatencyDistribution}
                    title={t('Avg Latency')}
                    icon={<Timer className='size-4' />}
                  />
                </div>,
                <TimeTrendTable timeDistribution={time_distribution || []} />,
              )}
            </TabsContent>

            <TabsContent value='errors' className='mt-4 space-y-4'>
              {renderTabContent(
                detailLoading,
                detailData,
                t('Loading...'),
                t('No data'),
                <VChartCard
                  spec={charts.specErrorRank}
                  title={t('Error Reason Rank')}
                  icon={<AlertTriangle className='size-4' />}
                  chartClassName='h-72 sm:h-80'
                />,
                <ErrorsTable
                  errorDistribution={error_distribution || []}
                />,
              )}
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  )
}

function ModelDistributionTable({
  modelDistribution,
}: {
  modelDistribution: {
    model_name: string
    count: number
    quota: number
    prompt_tokens: number
    completion_tokens: number
    error_count: number
  }[]
}) {
  const { t } = useTranslation()
  const totalCount = modelDistribution.reduce(
    (s, m) => s + (m.count || 0),
    0,
  )

  return (
    <Card className='p-4'>
      <div className='overflow-x-auto'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Model')}</TableHead>
              <TableHead className='text-right'>{t('Count')}</TableHead>
              <TableHead className='text-right'>{t('Percent')}</TableHead>
              <TableHead className='text-right'>{t('Quota')}</TableHead>
              <TableHead className='text-right'>{t('Prompt Tokens')}</TableHead>
              <TableHead className='text-right'>{t('Completion Tokens')}</TableHead>
              <TableHead className='text-right'>{t('Errors')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {modelDistribution.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className='text-muted-foreground py-8 text-center'
                >
                  {t('No model data')}
                </TableCell>
              </TableRow>
            ) : (
              modelDistribution.map((m) => {
                const pct =
                  totalCount > 0
                    ? ((m.count / totalCount) * 100).toFixed(1)
                    : '0'
                return (
                  <TableRow key={m.model_name}>
                    <TableCell>
                      <Badge variant='secondary'>{m.model_name}</Badge>
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatNumber(m.count || 0)}
                    </TableCell>
                    <TableCell>
                      <div className='flex items-center gap-2'>
                        <Progress value={parseFloat(pct)} className='w-20' />
                        <span className='text-xs'>{pct}%</span>
                      </div>
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatQuota(m.quota || 0)}
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatTokens(m.prompt_tokens || 0)}
                    </TableCell>
                    <TableCell className='text-right tabular-nums'>
                      {formatTokens(m.completion_tokens || 0)}
                    </TableCell>
                    <TableCell className='text-right'>
                      <Badge
                        variant={
                          (m.error_count || 0) > 0
                            ? 'destructive'
                            : 'secondary'
                        }
                      >
                        {m.error_count || 0}
                      </Badge>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  )
}

function TimeTrendTable({
  timeDistribution,
}: {
  timeDistribution: {
    timestamp: number
    count: number
    quota: number
    tokens: number
    avg_use_ms: number
  }[]
}) {
  const { t } = useTranslation()

  return (
    <Card className='p-4'>
      <div className='overflow-x-auto'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Time')}</TableHead>
              <TableHead className='text-right'>{t('Count')}</TableHead>
              <TableHead className='text-right'>{t('Quota')}</TableHead>
              <TableHead className='text-right'>{t('Tokens')}</TableHead>
              <TableHead className='text-right'>{t('Avg Latency')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {timeDistribution.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className='text-muted-foreground py-8 text-center'
                >
                  {t('No time data')}
                </TableCell>
              </TableRow>
            ) : (
              timeDistribution.map((item) => (
                <TableRow key={item.timestamp}>
                  <TableCell>
                    {new Date(item.timestamp * 1000).toLocaleDateString('zh-CN')}
                  </TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {formatNumber(item.count || 0)}
                  </TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {formatQuota(item.quota || 0)}
                  </TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {formatTokens(item.tokens || 0)}
                  </TableCell>
                  <TableCell className='text-right tabular-nums'>
                    {((item.avg_use_ms || 0) / 1000).toFixed(2)}s
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  )
}

function ErrorsTable({
  errorDistribution,
}: {
  errorDistribution: {
    model_name: string
    error_content: string
    count: number
    latest_at: number
  }[]
}) {
  const { t } = useTranslation()

  return (
    <Card className='p-4'>
      <div className='overflow-x-auto'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Model')}</TableHead>
              <TableHead>{t('Error')}</TableHead>
              <TableHead className='text-right'>{t('Count')}</TableHead>
              <TableHead>{t('Latest')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {errorDistribution.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className='text-muted-foreground py-8 text-center'
                >
                  {t('No error data')}
                </TableCell>
              </TableRow>
            ) : (
              errorDistribution.map((e) => (
                <TableRow key={`${e.model_name}-${e.error_content}`}>
                  <TableCell>
                    <Badge variant='outline'>{e.model_name || t('Unknown')}</Badge>
                  </TableCell>
                  <TableCell className='max-w-[420px] truncate'>
                    {e.error_content || t('No content')}
                  </TableCell>
                  <TableCell className='text-right'>
                    <Badge variant='destructive'>{e.count || 0}</Badge>
                  </TableCell>
                  <TableCell className='text-muted-foreground text-sm'>
                    {e.latest_at
                      ? new Date(e.latest_at * 1000).toLocaleString('zh-CN')
                      : '-'}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  )
}
