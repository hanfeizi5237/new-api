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

import React from 'react';
import { Card, Button, Table, Tag, Typography } from '@douyinfe/semi-ui';
import { VChart } from '@visactor/react-vchart';
import {
  Users,
  Activity,
  CreditCard,
  FileText,
  AlertCircle,
  Download,
  RefreshCw,
  ArrowUpRight,
  BarChart3,
  TrendingUp,
  Timer,
  ShieldAlert,
} from 'lucide-react';
import { renderQuota, renderNumber } from '../../helpers';

const { Text } = Typography;
const CHART_CONFIG = { mode: 'desktop-browser' };
const dateInputStyle = {
  color: '#111827',
  backgroundColor: '#ffffff',
  border: '1px solid #d1d5db',
};

const MainDashboardView = ({
  loading,
  overviewData,
  summary,
  loadOverview,
  handleDateRangeChange,
  handleGranularityChange,
  granularity,
  dateRange,
  exportCSV,
  openUserDetail,
  charts,
}) => {
  const statsCards = [
    { label: '总用户数', value: summary.totalUsers, icon: Users, color: 'blue' },
    { label: '总调用次数', value: renderNumber(summary.totalCount), icon: Activity, color: 'green' },
    { label: '总消耗额度', value: renderQuota(summary.totalQuota, 2), icon: CreditCard, color: 'orange' },
    { label: '总 Token 用量', value: renderNumber(summary.totalTokens), icon: FileText, color: 'purple' },
    { label: '总错误数', value: renderNumber(summary.totalErrors), icon: AlertCircle, color: 'red' },
  ];

  const statColorMap = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    orange: 'bg-orange-50 text-orange-600',
    purple: 'bg-purple-50 text-purple-600',
    red: 'bg-red-50 text-red-600',
  };

  const columns = [
    {
      title: '用户',
      dataIndex: 'username',
      render: (text, record) => (
        <div className='flex items-center gap-2'>
          <Text strong>{text || `用户${record.user_id}`}</Text>
          {record.display_name && <Text type='secondary' size='small'>({record.display_name})</Text>}
        </div>
      ),
    },
    { title: '调用次数', dataIndex: 'total_count', render: (v) => <Text>{renderNumber(v || 0)}</Text>, sorter: (a, b) => (a.total_count || 0) - (b.total_count || 0) },
    { title: '消耗额度', dataIndex: 'total_quota', render: (v) => <Text>{renderQuota(v || 0, 2)}</Text>, sorter: (a, b) => (a.total_quota || 0) - (b.total_quota || 0) },
    { title: 'Token 用量', dataIndex: 'total_tokens', render: (v) => <Text>{renderNumber(v || 0)}</Text>, sorter: (a, b) => (a.total_tokens || 0) - (b.total_tokens || 0) },
    { title: '失败数', dataIndex: 'error_count', render: (v) => <Tag color={v > 0 ? 'red' : 'green'} shape='circle'>{v || 0}</Tag>, sorter: (a, b) => (a.error_count || 0) - (b.error_count || 0) },
    {
      title: '操作',
      render: (_, record) => (
        <Button type='primary' theme='light' size='small' icon={<ArrowUpRight size={14} />} onClick={() => openUserDetail(record)}>
          查看详情
        </Button>
      ),
    },
  ];

  return (
    <div className='space-y-4'>
      <Card bordered headerLine className='!rounded-2xl'>
        <div className='flex flex-col gap-3'>
          <div>
            <div className='text-lg font-semibold'>用户用量总览</div>
            <div className='text-sm text-gray-500'>按用户查看调用次数、消耗金额、Token 与失败情况</div>
          </div>

          <div className='flex flex-wrap items-end gap-3'>
            <div className='flex flex-col gap-1'>
              <Text type='tertiary' size='small'>开始日</Text>
              <input
                type='date'
                value={dateRange.start}
                onChange={(e) => handleDateRangeChange({ start: e.target.value, end: dateRange.end })}
                className='rounded px-3 py-2 text-sm h-[36px]'
                style={dateInputStyle}
              />
            </div>
            <div className='pb-2 text-gray-400'>~</div>
            <div className='flex flex-col gap-1'>
              <Text type='tertiary' size='small'>结束日</Text>
              <input
                type='date'
                value={dateRange.end}
                onChange={(e) => handleDateRangeChange({ start: dateRange.start, end: e.target.value })}
                className='rounded px-3 py-2 text-sm h-[36px]'
                style={dateInputStyle}
              />
            </div>
            <div className='flex items-center gap-2 pb-[2px]'>
              <Button theme={granularity === 'day' ? 'solid' : 'light'} type='primary' onClick={() => handleGranularityChange('day')}>日</Button>
              <Button theme={granularity === 'week' ? 'solid' : 'light'} type='primary' onClick={() => handleGranularityChange('week')}>近7日</Button>
              <Button theme={granularity === 'month' ? 'solid' : 'light'} type='primary' onClick={() => handleGranularityChange('month')}>月</Button>
            </div>
            <Button type='primary' icon={<RefreshCw size={14} />} onClick={loadOverview} loading={loading}>查询</Button>
            <Button icon={<Download size={14} />} onClick={exportCSV} disabled={overviewData.length === 0}>导出 CSV</Button>
          </div>

          <div className='text-sm text-gray-500'>
            当前查询区间：{dateRange.start} ~ {dateRange.end}
            {granularity === 'day' && '（按日维度）'}
            {granularity === 'week' && '（近7日维度）'}
            {granularity === 'month' && '（月维度）'}
          </div>
        </div>
      </Card>

      <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4'>
        {statsCards.map((card) => (
          <Card key={card.label} className='hover:shadow-lg transition-shadow !rounded-2xl'>
            <div className='flex items-center gap-3'>
              <div className={`p-3 rounded-lg ${statColorMap[card.color]}`}>
                <card.icon size={20} />
              </div>
              <div>
                <Text type='tertiary' size='small'>{card.label}</Text>
                <div className='text-xl font-semibold mt-1'>{card.value}</div>
              </div>
            </div>
          </Card>
        ))}
      </div>

      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
        <Card title={<div className='flex items-center gap-2'><BarChart3 size={16} />用户消耗排行</div>} bordered headerLine className='!rounded-2xl' bodyStyle={{ padding: 0 }}>
          <div className='h-96 p-2'><VChart spec={charts.specUserRank} option={CHART_CONFIG} /></div>
        </Card>
        <Card title={<div className='flex items-center gap-2'><TrendingUp size={16} />用户用量趋势</div>} bordered headerLine className='!rounded-2xl' bodyStyle={{ padding: 0 }}>
          <div className='h-96 p-2'><VChart spec={charts.specUserTrend} option={CHART_CONFIG} /></div>
        </Card>
      </div>

      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
        <Card title={<div className='flex items-center gap-2'><Timer size={16} />调用次数排行</div>} bordered headerLine className='!rounded-2xl' bodyStyle={{ padding: 0 }}>
          <div className='h-80 p-2'><VChart spec={charts.specCountRank} option={CHART_CONFIG} /></div>
        </Card>
        <Card title={<div className='flex items-center gap-2'><ShieldAlert size={16} />失败用户排行</div>} bordered headerLine className='!rounded-2xl' bodyStyle={{ padding: 0 }}>
          <div className='h-80 p-2'><VChart spec={charts.specErrorUserRank} option={CHART_CONFIG} /></div>
        </Card>
      </div>

      <Card title='用户列表' bordered headerLine className='!rounded-2xl'>
        <Table columns={columns} dataSource={overviewData} loading={loading} pagination={{ pageSize: 20, showTotal: true, showSizeChanger: true }} rowKey='user_id' empty='暂无数据' />
      </Card>
    </div>
  );
};

export default MainDashboardView;
