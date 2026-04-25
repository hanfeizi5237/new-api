/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect } from 'react';
import { Typography } from '@douyinfe/semi-ui';
import { Calendar } from 'lucide-react';
import MainDashboardView from '../../components/user-usage/MainDashboardView';
import UserDetailView from '../../components/user-usage/UserDetailView';
import { useUserUsageData } from '../../hooks/user-usage/useUserUsageData';
import { useUserUsageCharts } from '../../hooks/user-usage/useUserUsageCharts';

const { Text } = Typography;

const UserUsageDashboard = () => {
  const usageData = useUserUsageData();
  const charts = useUserUsageCharts();

  const {
    loading,
    overviewData,
    globalTimeSeries,
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
  } = usageData;

  const summary = getSummary();

  useEffect(() => {
    loadOverview();
  }, []);

  useEffect(() => {
    if (overviewData.length > 0) {
      charts.updateOverviewCharts(overviewData, globalTimeSeries);
    }
  }, [overviewData, globalTimeSeries]);

  useEffect(() => {
    if (detailData) {
      charts.updateDetailCharts(detailData);
    }
  }, [detailData]);

  return (
    <div className='mt-[60px] px-2 pb-8'>
      <div className='flex items-center justify-between mb-4'>
        <div className='flex items-center gap-2'>
          <Calendar size={20} />
          <Text strong size='extra-large'>用户用量看板</Text>
        </div>
      </div>

      <MainDashboardView
        loading={loading}
        overviewData={overviewData}
        granularity={granularity}
        dateRange={dateRange}
        summary={summary}
        loadOverview={loadOverview}
        handleDateRangeChange={handleDateRangeChange}
        handleGranularityChange={handleGranularityChange}
        exportCSV={exportCSV}
        openUserDetail={openUserDetail}
        charts={charts}
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
  );
};

export default UserUsageDashboard;
