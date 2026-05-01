/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, or
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
import { SideSheet, Card, Table, Tag, Typography, Progress, Empty, Spin, Tabs } from '@douyinfe/semi-ui';
import { VChart } from '@visactor/react-vchart';
import { IllustrationConstruction, IllustrationConstructionDark } from '@douyinfe/semi-illustrations';
import { BarChart3, PieChart, TrendingUp, AlertTriangle, Timer, FileBarChart, CalendarRange } from 'lucide-react';
import { renderQuota, renderNumber, getQuotaWithUnit } from '../../helpers';
import { DETAIL_TABS } from '../../constants/user-usage.constants';

const { Text } = Typography;
const CHART_CONFIG = { mode: 'desktop-browser' };
const tokensToMillions = (tokens) => `${Number((Number(tokens || 0) / 1000000).toFixed(4))}M`;

const UserDetailView = ({ drawerVisible, closeUserDetail, detailLoading, detailData, selectedUser, activeDetailTab, setActiveDetailTab, charts, dateRange }) => {
  if (!selectedUser) return null;
  const { summary, model_distribution, time_distribution, error_distribution } = detailData || {};

  const renderContent = () => {
    if (detailLoading) return <div className='flex justify-center items-center py-20'><Spin size='large' /></div>;
    if (!detailData) {
      return <Empty image={<IllustrationConstruction style={{ width: 150, height: 150 }} />} darkModeImage={<IllustrationConstructionDark style={{ width: 150, height: 150 }} />} title='暂无数据' />;
    }
    switch (activeDetailTab) {
      case 'model':
        return (
          <div className='space-y-4'>
            <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
              <Card title={<div className='flex items-center gap-2'><PieChart size={16} />模型调用次数占比</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specModelPie} option={CHART_CONFIG} /></div></Card>
              <Card title={<div className='flex items-center gap-2'><PieChart size={16} />模型消耗占比</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specDetailModelQuotaPie} option={CHART_CONFIG} /></div></Card>
            </div>
            <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
              <Card title='模型调用次数排行' bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specModelRank} option={CHART_CONFIG} /></div></Card>
              <Card title={<div className='flex items-center gap-2'><FileBarChart size={16} />Token 分布</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specTokenDistribution} option={CHART_CONFIG} /></div></Card>
            </div>
            <ModelDistributionTab model_distribution={model_distribution} />
          </div>
        );
      case 'time':
        return (
          <div className='space-y-4'>
            <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
              <Card title={<div className='flex items-center gap-2'><TrendingUp size={16} />时间趋势</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specDetailTimeTrend} option={CHART_CONFIG} /></div></Card>
              <Card title={<div className='flex items-center gap-2'><Timer size={16} />平均耗时分布</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-80 p-2'><VChart spec={charts.specLatencyDistribution} option={CHART_CONFIG} /></div></Card>
            </div>
            <TimeTrendTab time_distribution={time_distribution} />
          </div>
        );
      case 'errors':
        return (
          <div className='space-y-4'>
            <Card title={<div className='flex items-center gap-2'><AlertTriangle size={16} />失败原因排行</div>} bordered headerLine bodyStyle={{ padding: 0 }}><div className='h-72 p-2'><VChart spec={charts.specErrorRank} option={CHART_CONFIG} /></div></Card>
            <ErrorsTab error_distribution={error_distribution} />
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <SideSheet visible={drawerVisible} onCancel={closeUserDetail} onClose={closeUserDetail} title={<div className='flex items-center gap-2'><BarChart3 size={18} /><span>{selectedUser.username || `用户${selectedUser.user_id}`}{selectedUser.display_name && <Text type='tertiary' size='small' className='ml-2'>({selectedUser.display_name})</Text>}- 用量详情</span></div>} placement='right' width={1100} maskClosable={false}>
      {summary && (
        <>
          <Card size='small' className='mb-4 !rounded-xl'>
            <div className='flex items-center gap-2 text-sm text-gray-600'>
              <CalendarRange size={16} />
              <span>当前查询日期范围：{dateRange?.start} ~ {dateRange?.end}</span>
            </div>
          </Card>
          <div className='mb-6 grid grid-cols-2 lg:grid-cols-5 gap-3'>
            <MetricCard label='调用次数' value={renderNumber(summary.total_count || 0)} />
            <MetricCard label='消耗额度' value={renderQuota(summary.total_quota || 0, 2)} />
            <MetricCard label='Token 用量' value={tokensToMillions(summary.total_tokens || 0)} />
            <MetricCard label='平均耗时' value={`${(summary.avg_use_time_ms / 1000).toFixed(2)}s`} />
            <MetricCard label='失败数' value={renderNumber(summary.error_count || 0)} />
          </div>
        </>
      )}
      <Tabs type='button' activeKey={activeDetailTab} onChange={setActiveDetailTab}>{DETAIL_TABS.map((tab) => <Tabs.TabPane itemKey={tab.key} tab={tab.label} key={tab.key}>{renderContent()}</Tabs.TabPane>)}</Tabs>
    </SideSheet>
  );
};

const MetricCard = ({ label, value }) => <Card size='small'><div className='text-center'><Text type='tertiary' size='small'>{label}</Text><div className='text-2xl font-semibold mt-1'>{value}</div></div></Card>;

const ModelDistributionTab = ({ model_distribution = [] }) => {
  const totalCount = model_distribution.reduce((s, m) => s + (m.count || 0), 0);
  const columns = [
    { title: '模型', dataIndex: 'model_name', render: (v) => <Tag color='blue'>{v}</Tag> },
    { title: '调用次数', dataIndex: 'count', render: (v) => renderNumber(v || 0) },
    { title: '占比', render: (_, record) => { const pct = totalCount > 0 ? ((record.count / totalCount) * 100).toFixed(1) : 0; return <div className='flex items-center gap-2'><Progress percent={parseFloat(pct)} showInfo={false} stroke='var(--semi-color-primary)' /><Text size='small'>{pct}%</Text></div>; } },
    { title: '消耗额度', dataIndex: 'quota', render: (v) => renderQuota(v || 0, 2) },
    { title: '输入 Token', dataIndex: 'prompt_tokens', render: (v) => tokensToMillions(v || 0) },
    { title: '输出 Token', dataIndex: 'completion_tokens', render: (v) => tokensToMillions(v || 0) },
    { title: '失败数', dataIndex: 'error_count', render: (v) => <Tag color={v > 0 ? 'orange' : 'green'} shape='circle'>{v || 0}</Tag> },
  ];
  return <Table columns={columns} dataSource={model_distribution} pagination={false} rowKey='model_name' empty='暂无模型数据' />;
};

const TimeTrendTab = ({ time_distribution = [] }) => {
  const columns = [
    { title: '时间', render: (_, record) => new Date(record.timestamp * 1000).toLocaleDateString('zh-CN') },
    { title: '调用次数', dataIndex: 'count', render: (v) => renderNumber(v || 0) },
    { title: '消耗额度', dataIndex: 'quota', render: (v) => renderQuota(v || 0, 2) },
    { title: 'Token 用量', dataIndex: 'tokens', render: (v) => tokensToMillions(v || 0) },
    { title: '平均耗时', dataIndex: 'avg_use_ms', render: (v) => `${((v || 0) / 1000).toFixed(2)}s` },
  ];
  return <Table columns={columns} dataSource={time_distribution} pagination={{ pageSize: 10, showTotal: true }} rowKey='timestamp' empty='暂无时间数据' />;
};

const ErrorsTab = ({ error_distribution = [] }) => {
  const columns = [
    { title: '模型', dataIndex: 'model_name', render: (v) => <Tag color='orange'>{v || '未知模型'}</Tag> },
    { title: '错误信息', dataIndex: 'error_content', render: (v) => <Text ellipsis={{ showTooltip: true }} style={{ maxWidth: 420 }}>{v || '无内容'}</Text> },
    { title: '次数', dataIndex: 'count', render: (v) => <Tag color='red' shape='circle'>{v || 0}</Tag> },
    { title: '最近发生', dataIndex: 'latest_at', render: (v) => <Text type='tertiary' size='small'>{v ? new Date(v * 1000).toLocaleString('zh-CN') : '-'}</Text> },
  ];
  return <Table columns={columns} dataSource={error_distribution} pagination={{ pageSize: 10, showTotal: true }} rowKey={(record) => `${record.model_name}-${record.error_content}`} empty='暂无错误数据' />;
};

export default UserDetailView;
