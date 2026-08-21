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
import { Calendar } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'

import { MainDashboardView } from './components/main-dashboard-view'
import { UserDetailView } from './components/user-detail-view'
import { useUserUsageData } from './hooks/use-user-usage-data'
import {
  createEmptyChartSpecs,
  updateDetailCharts,
  updateOverviewCharts,
} from './lib/charts'

export function UserUsageDashboard() {
  const { t } = useTranslation()
  const usageData = useUserUsageData()

  const {
    loading,
    overviewData,
    globalTimeSeries,
    globalTimeSeriesByModel,
    granularity,
    dateRange,
    drawerVisible,
    selectedUser,
    detailLoading,
    detailData,
    activeDetailTab,
    getSummary,
    loadOverview,
    openUserDetail,
    closeUserDetail,
    setActiveDetailTab,
    handleDateRangeChange,
    handleGranularityChange,
    exportCSV,
  } = usageData

  const summary = getSummary()

  const baseSpecs = useMemo(() => createEmptyChartSpecs(), [])

  const charts = useMemo(() => {
    let specs = baseSpecs
    if (overviewData.length > 0) {
      specs = updateOverviewCharts(
        specs,
        overviewData,
        globalTimeSeries,
        globalTimeSeriesByModel,
      )
    }
    if (detailData) {
      specs = updateDetailCharts(specs, detailData)
    }
    return specs
  }, [baseSpecs, overviewData, globalTimeSeries, globalTimeSeriesByModel, detailData])

  useEffect(() => {
    loadOverview()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        <div className='flex items-center gap-2'>
          <Calendar className='size-5' />
          <span>{t('User Usage Dashboard')}</span>
        </div>
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          <MainDashboardView
            loading={loading}
            overviewData={overviewData}
            summary={summary}
            granularity={granularity}
            dateRange={dateRange}
            charts={charts}
            loadOverview={loadOverview}
            handleDateRangeChange={handleDateRangeChange}
            handleGranularityChange={handleGranularityChange}
            exportCSV={exportCSV}
            openUserDetail={openUserDetail}
          />

          <UserDetailView
            drawerVisible={drawerVisible}
            closeUserDetail={closeUserDetail}
            detailLoading={detailLoading}
            detailData={detailData}
            selectedUser={selectedUser}
            activeDetailTab={activeDetailTab}
            setActiveDetailTab={setActiveDetailTab}
            charts={charts}
            dateRange={dateRange}
          />
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
