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
import { useListingsData } from '../../../hooks/listings/useListingsData';

const { Text } = Typography;

const renderTag = (text, color = 'grey') => (
  <Tag color={color} shape='circle'>
    {text || '-'}
  </Tag>
);

const getListingStatusColor = (status) => {
  switch (status) {
    case 'active':
      return 'green';
    case 'paused':
      return 'yellow';
    case 'sold_out':
      return 'orange';
    default:
      return 'grey';
  }
};

const getAuditStatusColor = (status) => {
  switch (status) {
    case 'approved':
      return 'green';
    case 'pending_review':
      return 'blue';
    case 'rejected':
      return 'red';
    default:
      return 'grey';
  }
};

const formatMinorAmount = (value) => `¥ ${(Number(value || 0) / 100).toFixed(2)}`;

const renderSkuSummary = (skus = []) => {
  if (!skus.length) {
    return '-';
  }
  const preview = skus.slice(0, 2).map((sku) => (
    <div key={sku.id || sku.sku_code} className='flex flex-col'>
      <Text>{sku.sku_code}</Text>
      <Text type='tertiary' size='small'>
        {`${sku.package_amount} ${sku.package_unit} / ${formatMinorAmount(sku.unit_price_minor)}`}
      </Text>
    </div>
  ));
  return (
    <div className='flex flex-col gap-1'>
      {preview}
      {skus.length > 2 && (
        <Text type='tertiary' size='small'>
          {`+${skus.length - 2}`}
        </Text>
      )}
    </div>
  );
};

const ListingActions = ({ record, actingId, runListingAction, t }) => {
  const listing = record.listing || {};
  const loading = actingId === listing.id;

  if (listing.audit_status === 'draft' || listing.audit_status === 'rejected') {
    return (
      <Button
        size='small'
        type='tertiary'
        loading={loading}
        onClick={() => runListingAction(record, 'submit')}
      >
        {t('提审')}
      </Button>
    );
  }

  if (listing.audit_status === 'pending_review') {
    return (
      <Space wrap>
        <Button
          size='small'
          type='primary'
          loading={loading}
          onClick={() => runListingAction(record, 'approve')}
        >
          {t('通过')}
        </Button>
        <Button
          size='small'
          type='danger'
          theme='light'
          loading={loading}
          onClick={() => runListingAction(record, 'reject')}
        >
          {t('驳回')}
        </Button>
      </Space>
    );
  }

  if (listing.audit_status === 'approved' && listing.status === 'active') {
    return (
      <Button
        size='small'
        type='tertiary'
        loading={loading}
        onClick={() => runListingAction(record, 'pause')}
      >
        {t('暂停')}
      </Button>
    );
  }

  if (listing.audit_status === 'approved') {
    return (
      <Button
        size='small'
        type='tertiary'
        loading={loading}
        onClick={() => runListingAction(record, 'activate')}
      >
        {t('上架')}
      </Button>
    );
  }

  return '-';
};

const ListingsPage = () => {
  const listingsData = useListingsData();
  const isMobile = useIsMobile();

  const {
    t,
    compactMode,
    setCompactMode,
    listings,
    loading,
    actingId,
    activePage,
    pageSize,
    listingCount,
    draftFilters,
    setDraftFilters,
    applyFilters,
    resetFilters,
    refresh,
    handlePageChange,
    handlePageSizeChange,
    runListingAction,
  } = listingsData;

  const columns = useMemo(
    () => [
      {
        title: t('商品'),
        key: 'listing',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{record.listing?.title || '-'}</Text>
            <Text type='tertiary' size='small'>
              {record.listing?.listing_code || '-'}
            </Text>
          </div>
        ),
      },
      {
        title: t('卖家 / 供给'),
        key: 'seller_supply',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{`seller #${record.listing?.seller_id || '-'}`}</Text>
            <Text type='tertiary' size='small'>
              {`supply #${record.listing?.supply_account_id || '-'}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('模型'),
        key: 'model',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{record.listing?.model_name || '-'}</Text>
            <Text type='tertiary' size='small'>
              {`vendor #${record.listing?.vendor_id || '-'}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('SKU'),
        key: 'skus',
        render: (_, record) => renderSkuSummary(record.skus),
      },
      {
        title: t('审核 / 状态'),
        key: 'status',
        render: (_, record) => (
          <div className='flex flex-wrap gap-1 justify-end md:justify-start'>
            {renderTag(
              record.listing?.audit_status,
              getAuditStatusColor(record.listing?.audit_status),
            )}
            {renderTag(
              record.listing?.status,
              getListingStatusColor(record.listing?.status),
            )}
          </div>
        ),
      },
      {
        title: t('更新时间'),
        key: 'updated_at',
        render: (_, record) =>
          record.listing?.updated_at
            ? timestamp2string(record.listing.updated_at)
            : '-',
      },
      {
        title: t('操作'),
        key: 'actions',
        render: (_, record) => (
          <ListingActions
            record={record}
            actingId={actingId}
            runListingAction={runListingAction}
            t={t}
          />
        ),
      },
    ],
    [actingId, runListingAction, t],
  );

  return (
    <CardPro
      type='type1'
      descriptionArea={
        <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-2'>
          <div className='flex flex-col'>
            <Text strong>{t('商品管理')}</Text>
            <Text type='tertiary'>
              {t('管理商品审核、上下架状态与 SKU 定价包')}
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
            placeholder={t('搜索商品标题 / 编码 / 模型')}
            showClear
          />
          <Input
            value={draftFilters.sellerId}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, sellerId: value }))
            }
            placeholder={t('卖家 ID')}
            showClear
          />
          <Select
            value={draftFilters.auditStatus}
            optionList={[
              { label: t('全部审核状态'), value: '' },
              { label: t('草稿'), value: 'draft' },
              { label: t('待审核'), value: 'pending_review' },
              { label: t('已通过'), value: 'approved' },
              { label: t('已驳回'), value: 'rejected' },
            ]}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, auditStatus: value || '' }))
            }
          />
          <Select
            value={draftFilters.status}
            optionList={[
              { label: t('全部商品状态'), value: '' },
              { label: t('上架中'), value: 'active' },
              { label: t('已暂停'), value: 'paused' },
              { label: t('已售罄'), value: 'sold_out' },
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
        total: listingCount,
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
        dataSource={listings}
        loading={loading}
        hidePagination={true}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        empty={<Empty description={t('暂无商品数据')} style={{ padding: 32 }} />}
        className='rounded-xl overflow-hidden'
      />
    </CardPro>
  );
};

export default ListingsPage;
