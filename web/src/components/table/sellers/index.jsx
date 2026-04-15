import React, { useMemo } from 'react';
import {
  Button,
  Empty,
  Input,
  Select,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconRefresh } from '@douyinfe/semi-icons';
import CardPro from '../../common/ui/CardPro';
import CardTable from '../../common/ui/CardTable';
import { createCardProPagination, timestamp2string } from '../../../helpers/utils';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { useSellersData } from '../../../hooks/sellers/useSellersData';

const { Text } = Typography;

const formatMinorAmount = (value) => `¥ ${(Number(value || 0) / 100).toFixed(2)}`;

const renderTag = (text, color = 'grey') => (
  <Tag color={color} shape='circle'>
    {text || '-'}
  </Tag>
);

const getSellerStatusColor = (status) => {
  switch (status) {
    case 'active':
      return 'green';
    case 'disabled':
      return 'red';
    case 'paused':
      return 'yellow';
    default:
      return 'grey';
  }
};

const getRiskColor = (riskLevel) => {
  switch (riskLevel) {
    case 'high':
      return 'red';
    case 'medium':
      return 'yellow';
    default:
      return 'blue';
  }
};

const SellersPage = () => {
  const sellersData = useSellersData();
  const isMobile = useIsMobile();

  const {
    t,
    compactMode,
    setCompactMode,
    sellers,
    loading,
    actingId,
    activePage,
    pageSize,
    sellerCount,
    draftFilters,
    setDraftFilters,
    applyFilters,
    resetFilters,
    refresh,
    handlePageChange,
    handlePageSizeChange,
    updateSellerStatus,
  } = sellersData;

  const columns = useMemo(
    () => [
      {
        title: t('卖家'),
        key: 'seller',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{record.display_name || '-'}</Text>
            <Text type='tertiary' size='small'>
              {record.seller_code}
            </Text>
          </div>
        ),
      },
      {
        title: t('用户'),
        key: 'user',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{`#${record.user_id}`}</Text>
            <Text type='tertiary' size='small'>
              {record.company_name || t('个人卖家')}
            </Text>
          </div>
        ),
      },
      {
        title: t('类型与风控'),
        key: 'profile',
        render: (_, record) => (
          <div className='flex flex-wrap gap-1 justify-end md:justify-start'>
            {renderTag(record.seller_type || t('未设置'), 'blue')}
            {renderTag(record.kyc_status || t('未实名'), 'orange')}
            {renderTag(record.risk_level || t('low'), getRiskColor(record.risk_level))}
          </div>
        ),
      },
      {
        title: t('保证金'),
        key: 'deposit',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{formatMinorAmount(record.deposit_balance_minor)}</Text>
            <Text type='tertiary' size='small'>
              {`${t('冻结')} ${formatMinorAmount(record.frozen_deposit_minor)}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('状态'),
        key: 'status',
        render: (_, record) => renderTag(record.status, getSellerStatusColor(record.status)),
      },
      {
        title: t('更新时间'),
        key: 'updated_at',
        render: (_, record) =>
          record.updated_at ? timestamp2string(record.updated_at) : '-',
      },
      {
        title: t('操作'),
        key: 'actions',
        render: (_, record) => {
          const nextStatus = record.status === 'active' ? 'disabled' : 'active';
          return (
            <Space wrap>
              <Button
                size='small'
                type='tertiary'
                loading={actingId === record.id}
                onClick={() => updateSellerStatus(record, nextStatus)}
              >
                {record.status === 'active' ? t('禁用') : t('启用')}
              </Button>
            </Space>
          );
        },
      },
    ],
    [actingId, t, updateSellerStatus],
  );

  return (
    <CardPro
      type='type1'
      descriptionArea={
        <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-2'>
          <div className='flex flex-col'>
            <Text strong>{t('卖家管理')}</Text>
            <Text type='tertiary'>
              {t('管理卖家主体、准入状态与基础风控信息')}
            </Text>
          </div>
          <Space wrap>
            <Button type='tertiary' icon={<IconRefresh />} onClick={refresh}>
              {t('刷新')}
            </Button>
            <Button type='tertiary' onClick={() => setCompactMode(!compactMode)}>
              {compactMode ? t('舒展视图') : t('紧凑视图')}
            </Button>
          </Space>
        </div>
      }
      searchArea={
        <div className='flex flex-col md:flex-row gap-2'>
          <Input
            value={draftFilters.keyword}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, keyword: value }))
            }
            placeholder={t('搜索卖家名称 / 编码 / 邮箱')}
            showClear
          />
          <Select
            value={draftFilters.status}
            optionList={[
              { label: t('全部状态'), value: '' },
              { label: t('活跃'), value: 'active' },
              { label: t('禁用'), value: 'disabled' },
              { label: t('暂停'), value: 'paused' },
            ]}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, status: value || '' }))
            }
          />
          <Space wrap>
            <Button type='primary' onClick={applyFilters}>
              {t('查询')}
            </Button>
            <Button type='tertiary' onClick={resetFilters}>
              {t('重置')}
            </Button>
          </Space>
        </div>
      }
      paginationArea={createCardProPagination({
        currentPage: activePage,
        pageSize,
        total: sellerCount,
        onPageChange: handlePageChange,
        onPageSizeChange: handlePageSizeChange,
        isMobile,
        t,
      })}
      t={t}
    >
      <CardTable
        rowKey='key'
        columns={columns}
        dataSource={sellers}
        loading={loading}
        hidePagination={true}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        empty={<Empty description={t('暂无卖家数据')} style={{ padding: 32 }} />}
        className='rounded-xl overflow-hidden'
      />
    </CardPro>
  );
};

export default SellersPage;
