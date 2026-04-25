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
import { Drawer, Card, Table, Tag, Typography, Progress, Empty, Spin, Tabs } from '@douyinfe/semi-ui';
import {
  IllustrationConstruction,
  IllustrationConstructionDark,
} from '@douyinfe/semi-illustrations';
import { Clock, AlertTriangle, BarChart3 } from 'lucide-react';
import { renderQuota, renderNumber } from '../../helpers';
import { DETAIL_TABS } from '../../constants/user-usage.constants';

const { Text, Title } = Typography;

const UserDetailView = ({
  drawerVisible,
  closeUserDetail,
  detailLoading,
  detailData,
  selectedUser,
  activeDetailTab,
  setActiveDetailTab,
  charts,
  t,
}) => {
  if (!selectedUser) return null;

  const { summary, model_distribution, time_distribution, error_distribution } = detailData || {};

  const renderContent = () => {
    if (detailLoading) {
      return (
        <div className='flex justify-center items-center py-20'>
          <Spin size='large' />
        </div>
      );
    }

    if (!detailData) {
      return (
        <Empty
          image={<IllustrationConstruction style={{ width: 150, height: 150 }} />}
          darkModeImage={<IllustrationConstructionDark style={{ width: 150, height: 150 }} />}
          title='暂无数据'
        />
      );
    }

    switch (activeDetailTab) {
      case 'model':
        return <ModelDistributionTab model_distribution={model_distribution} />;
      case 'time':
        return <TimeTrendTab time_distribution={time_distribution} />;
      case 'errors':
        return <ErrorsTab error_distribution={error_distribution} />;
      default:
        return null;
    }
  };

  return (
    <Drawer
      visible={drawerVisible}
      onCancel={closeUserDetail}
      onClose={closeUserDetail}
      title={
        <div className='flex items-center gap-2'>
          <BarChart3 size={18} />
          <span>
            {selectedUser.username || `用户${selectedUser.user_id}`}
            {selectedUser.display_name && (
              <Text type='tertiary' size='small' className='ml-2'>
                ({selectedUser.display_name})
              </Text>
            )}
            - 用量详情
          </span>
        </div>
      }
      placement='right'
      width={800}
      maskClosable={false}
    >
      {/* 汇总信息 */}
      {summary && (
        <div className='mb-6 grid grid-cols-2 lg:grid-cols-4 gap-3'>
          <Card size='small'>
            <div className='text-center'>
              <Text type='tertiary' size='small'>调用次数</Text>
              <div className='text-2xl font-semibold mt-1'>
                {renderNumber(summary.total_count || 0)}
              </div>
            </div>
          </Card>
          <Card size='small'>
            <div className='text-center'>
              <Text type='tertiary' size='small'>消耗额度</Text>
              <div className='text-2xl font-semibold mt-1'>
                {renderQuota(summary.total_quota || 0, 2)}
              </div>
            </div>
          </Card>
          <Card size='small'>
            <div className='text-center'>
              <Text type='tertiary' size='small'>Token 用量</Text>
              <div className='text-2xl font-semibold mt-1'>
                {renderNumber(summary.total_tokens || 0)}
              </div>
            </div>
          </Card>
          <Card size='small'>
            <div className='text-center'>
              <Text type='tertiary' size='small'>平均耗时</Text>
              <div className='text-2xl font-semibold mt-1'>
                {(summary.avg_use_time_ms / 1000).toFixed(2)}s
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Tab 切换 */}
      <Tabs
        type='button'
        activeKey={activeDetailTab}
        onChange={setActiveDetailTab}
      >
        {DETAIL_TABS.map((tab) => (
          <Tabs.TabPane itemKey={tab.key} tab={tab.label} key={tab.key}>
            {renderContent()}
          </Tabs.TabPane>
        ))}
      </Tabs>
    </Drawer>
  );
};

// ========== 子组件：模型分布 ==========
const ModelDistributionTab = ({ model_distribution = [] }) => {
  const totalQuota = model_distribution.reduce((s, m) => s + (m.quota || 0), 0);
  const totalCount = model_distribution.reduce((s, m) => s + (m.count || 0), 0);

  const columns = [
    {
      title: '模型',
      dataIndex: 'model_name',
      render: (v) => <Tag color='blue'>{v}</Tag>,
    },
    {
      title: '调用次数',
      dataIndex: 'count',
      render: (v) => renderNumber(v || 0),
    },
    {
      title: '占比',
      render: (_, record) => {
        const pct = totalCount > 0 ? ((record.count / totalCount) * 100).toFixed(1) : 0;
        return (
          <div className='flex items-center gap-2'>
            <Progress percent={parseFloat(pct)} showInfo={false} stroke='var(--semi-color-primary)' />
            <Text size='small'>{pct}%</Text>
          </div>
        );
      },
    },
    {
      title: '消耗额度',
      dataIndex: 'quota',
      render: (v) => renderQuota(v || 0, 2),
    },
    {
      title: '输入 Token',
      dataIndex: 'prompt_tokens',
      render: (v) => renderNumber(v || 0),
    },
    {
      title: '输出 Token',
      dataIndex: 'completion_tokens',
      render: (v) => renderNumber(v || 0),
    },
    {
      title: '错误数',
      dataIndex: 'error_count',
      render: (v) => (
        <Tag color={v > 0 ? 'orange' : 'green'} shape='circle'>{v || 0}</Tag>
      ),
    },
  ];

  return (
    <div>
      <Table
        columns={columns}
        dataSource={model_distribution}
        pagination={false}
        rowKey='model_name'
        empty='暂无模型数据'
      />
    </div>
  );
};

// ========== 子组件：时间趋势 ==========
const TimeTrendTab = ({ time_distribution = [] }) => {
  const columns = [
    {
      title: '时间',
      render: (_, record) => new Date(record.timestamp * 1000).toLocaleDateString('zh-CN'),
    },
    {
      title: '调用次数',
      dataIndex: 'count',
      render: (v) => renderNumber(v || 0),
    },
    {
      title: '消耗额度',
      dataIndex: 'quota',
      render: (v) => renderQuota(v || 0, 2),
    },
    {
      title: 'Token 用量',
      dataIndex: 'tokens',
      render: (v) => renderNumber(v || 0),
    },
  ];

  return (
    <div>
      <Table
        columns={columns}
        dataSource={time_distribution}
        pagination={{ pageSize: 10, showTotal: true }}
        rowKey='timestamp'
        empty='暂无时间数据'
      />
    </div>
  );
};

// ========== 子组件：错误统计 ==========
const ErrorsTab = ({ error_distribution = [] }) => {
  const columns = [
    {
      title: '模型',
      dataIndex: 'model_name',
      render: (v) => <Tag color='orange'>{v || '未知'}</Tag>,
    },
    {
      title: '错误信息',
      dataIndex: 'error_content',
      render: (v) => (
        <Text ellipsis={{ showTooltip: true }} style={{ maxWidth: 400 }}>
          {v || '无内容'}
        </Text>
      ),
    },
    {
      title: '次数',
      dataIndex: 'count',
      render: (v) => <Tag color='red' shape='circle'>{v || 0}</Tag>,
    },
    {
      title: '最近发生',
      dataIndex: 'latest_at',
      render: (v) => (
        <Text type='tertiary' size='small'>
          {v ? new Date(v * 1000).toLocaleString('zh-CN') : '-'}
        </Text>
      ),
    },
  ];

  return (
    <div>
      <Table
        columns={columns}
        dataSource={error_distribution}
        pagination={{ pageSize: 10, showTotal: true }}
        rowKey={(record) => `${record.model_name}-${record.error_content}`}
        empty='暂无错误数据'
      />
    </div>
  );
};

export default UserDetailView;
