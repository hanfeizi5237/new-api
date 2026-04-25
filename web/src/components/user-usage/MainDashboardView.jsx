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
import { Card, Button, Table, Tag, Typography, Form } from '@douyinfe/semi-ui';
import {
  Users,
  Activity,
  CreditCard,
  FileText,
  AlertCircle,
  Download,
  RefreshCw,
  ArrowUpRight,
} from 'lucide-react';
import { renderQuota, renderNumber } from '../../helpers';

const { Text } = Typography;

const MainDashboardView = ({
  loading,
  overviewData,
  dateRange,
  granularity,
  summary,
  loadOverview,
  handleDateRangeChange,
  handleGranularityChange,
  exportCSV,
  openUserDetail,
  charts,
  t,
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
          {record.display_name && (
            <Text type='secondary' size='small'>({record.display_name})</Text>
          )}
        </div>
      ),
    },
    {
      title: '调用次数',
      dataIndex: 'total_count',
      render: (v) => <Text>{renderNumber(v || 0)}</Text>,
      sorter: (a, b) => (a.total_count || 0) - (b.total_count || 0),
    },
    {
      title: '消耗额度',
      dataIndex: 'total_quota',
      render: (v) => <Text>{renderQuota(v || 0, 2)}</Text>,
      sorter: (a, b) => (a.total_quota || 0) - (b.total_quota || 0),
    },
    {
      title: 'Token 用量',
      dataIndex: 'total_tokens',
      render: (v) => <Text>{renderNumber(v || 0)}</Text>,
      sorter: (a, b) => (a.total_tokens || 0) - (b.total_tokens || 0),
    },
    {
      title: '错误数',
      dataIndex: 'error_count',
      render: (v) => (
        <Tag color={v > 0 ? 'red' : 'green'} shape='circle'>
          {v || 0}
        </Tag>
      ),
      sorter: (a, b) => (a.error_count || 0) - (b.error_count || 0),
    },
    {
      title: '操作',
      render: (_, record) => (
        <Button
          type='primary'
          theme='light'
          size='small'
          icon={<ArrowUpRight size={14} />}
          onClick={() => openUserDetail(record)}
        >
          查看详情
        </Button>
      ),
    },
  ];

  const rowSelection = {
    onChange: (selectedRowKeys, selectedRows) => {},
  };

  return (
    <div className='space-y-4'>
      {/* 顶部工具栏 */}
      <div className='flex flex-wrap items-center gap-3'>
        <Form.DatePicker
          field='dateRange'
          type='dateRange'
          placeholder={['开始日期', '结束日期']}
          showClear
          style={{ width: 240 }}
          onChange={(dates) => {
            if (dates && dates.length === 2) {
              handleDateRangeChange({ start: dates[0], end: dates[1] });
            }
          }}
        />

        <select
          value={granularity}
          onChange={(e) => handleGranularityChange(e.target.value)}
          className='border rounded px-3 py-2 text-sm'
          style={{ width: 100 }}
        >
          <option value='day'>天</option>
          <option value='week'>周</option>
          <option value='month'>月</option>
        </select>

        <Button
          type='primary'
          icon={<RefreshCw size={14} />}
          onClick={loadOverview}
          loading={loading}
        >
          查询
        </Button>

        <Button
          icon={<Download size={14} />}
          onClick={exportCSV}
          disabled={overviewData.length === 0}
        >
          导出 CSV
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4'>
        {statsCards.map((card) => (
          <Card key={card.label} className='hover:shadow-lg transition-shadow'>
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

      {/* 图表区域 */}
      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
        <Card title='用户消耗排行' bordered headerLine>
          <div style={{ height: 350 }}>
            {/* VChart render via spec */}
          </div>
        </Card>
        <Card title='全局用量趋势' bordered headerLine>
          <div style={{ height: 350 }}>
            {/* VChart render via spec */}
          </div>
        </Card>
      </div>

      {/* 用户列表 */}
      <Card title='用户列表' bordered headerLine>
        <Table
          columns={columns}
          dataSource={overviewData}
          loading={loading}
          pagination={{
            pageSize: 20,
            showTotal: true,
            showSizeChanger: true,
          }}
          rowKey='user_id'
          rowSelection={rowSelection}
          empty='暂无数据'
        />
      </Card>
    </div>
  );
};

export default MainDashboardView;
