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
import { useEntitlementsData } from '../../../hooks/entitlements/useEntitlementsData';

const { Text } = Typography;

const renderTag = (text, color = 'grey') => (
  <Tag color={color} shape='circle'>
    {text || '-'}
  </Tag>
);

const getStatusColor = (status) => {
  switch (status) {
    case 'active':
      return 'green';
    case 'expired':
      return 'grey';
    case 'paused':
      return 'yellow';
    default:
      return 'blue';
  }
};

const calculateRemaining = (record) =>
  Math.max(
    Number(record.total_granted || 0) -
      Number(record.total_used || 0) -
      Number(record.total_frozen || 0) -
      Number(record.total_refunded || 0),
    0,
  );

const EntitlementsPage = () => {
  const entitlementsData = useEntitlementsData();
  const isMobile = useIsMobile();

  const {
    t,
    compactMode,
    setCompactMode,
    entitlements,
    loading,
    activePage,
    pageSize,
    entitlementCount,
    draftFilters,
    setDraftFilters,
    applyFilters,
    resetFilters,
    refresh,
    handlePageChange,
    handlePageSizeChange,
  } = entitlementsData;

  const columns = useMemo(
    () => [
      {
        title: t('买家'),
        key: 'buyer',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{`user #${record.buyer_user_id}`}</Text>
            <Text type='tertiary' size='small'>
              {`entitlement #${record.id}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('模型'),
        key: 'model',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{record.model_name || '-'}</Text>
            <Text type='tertiary' size='small'>
              {`vendor #${record.vendor_id || '-'}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('额度'),
        key: 'amounts',
        render: (_, record) => (
          <div className='flex flex-col gap-1'>
            <Text>{`${t('授予')} ${record.total_granted || 0}`}</Text>
            <Text type='tertiary' size='small'>
              {`${t('已用')} ${record.total_used || 0} / ${t('冻结')} ${record.total_frozen || 0}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('可用余额'),
        key: 'remaining',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{calculateRemaining(record)}</Text>
            <Text type='tertiary' size='small'>
              {`${t('已退款')} ${record.total_refunded || 0}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('状态'),
        key: 'status',
        render: (_, record) => renderTag(record.status, getStatusColor(record.status)),
      },
      {
        title: t('更新时间'),
        key: 'updated_at',
        render: (_, record) =>
          record.updated_at ? timestamp2string(record.updated_at) : '-',
      },
    ],
    [t],
  );

  return (
    <CardPro
      type='type1'
      descriptionArea={
        <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-2'>
          <div className='flex flex-col'>
            <Text strong>{t('权益管理')}</Text>
            <Text type='tertiary'>
              {t('查看买家 entitlement 聚合余额与可用状态')}
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
            value={draftFilters.buyerUserId}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, buyerUserId: value }))
            }
            placeholder={t('买家 ID')}
            showClear
          />
          <Input
            value={draftFilters.modelName}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, modelName: value }))
            }
            placeholder={t('模型名称')}
            showClear
          />
          <Select
            value={draftFilters.status}
            optionList={[
              { label: t('全部状态'), value: '' },
              { label: t('活跃'), value: 'active' },
              { label: t('过期'), value: 'expired' },
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
        total: entitlementCount,
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
        dataSource={entitlements}
        loading={loading}
        hidePagination={true}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        empty={
          <Empty description={t('暂无权益数据')} style={{ padding: 32 }} />
        }
        className='rounded-xl overflow-hidden'
      />
    </CardPro>
  );
};

export default EntitlementsPage;
