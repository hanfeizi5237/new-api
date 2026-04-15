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
import { useOrdersData } from '../../../hooks/orders/useOrdersData';

const { Text } = Typography;

const renderTag = (text, color = 'grey') => (
  <Tag color={color} shape='circle'>
    {text || '-'}
  </Tag>
);

const getOrderStatusColor = (status) => {
  switch (status) {
    case 'paid':
      return 'green';
    case 'pending_payment':
      return 'blue';
    case 'closed':
      return 'grey';
    default:
      return 'grey';
  }
};

const getPaymentStatusColor = (status) => {
  switch (status) {
    case 'paid':
      return 'green';
    case 'unpaid':
      return 'orange';
    case 'failed':
      return 'red';
    default:
      return 'grey';
  }
};

const getEntitlementStatusColor = (status) => {
  switch (status) {
    case 'created':
      return 'green';
    case 'pending':
      return 'blue';
    case 'failed':
      return 'red';
    default:
      return 'grey';
  }
};

const formatMinorAmount = (value, currency = 'CNY') =>
  `${currency} ${(Number(value || 0) / 100).toFixed(2)}`;

const renderItemsSummary = (items = []) => {
  if (!items.length) {
    return '-';
  }
  return (
    <div className='flex flex-col gap-1'>
      {items.slice(0, 2).map((item) => (
        <div key={item.id} className='flex flex-col'>
          <Text>{item.model_name || '-'}</Text>
          <Text type='tertiary' size='small'>
            {`${item.package_amount} ${item.package_unit} x ${item.quantity}`}
          </Text>
        </div>
      ))}
      {items.length > 2 && (
        <Text type='tertiary' size='small'>
          {`+${items.length - 2}`}
        </Text>
      )}
    </div>
  );
};

const OrdersPage = () => {
  const ordersData = useOrdersData();
  const isMobile = useIsMobile();

  const {
    t,
    compactMode,
    setCompactMode,
    orders,
    loading,
    activePage,
    pageSize,
    orderCount,
    draftFilters,
    setDraftFilters,
    applyFilters,
    resetFilters,
    refresh,
    handlePageChange,
    handlePageSizeChange,
  } = ordersData;

  const columns = useMemo(
    () => [
      {
        title: t('订单'),
        key: 'order',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{record.order?.order_no || '-'}</Text>
            <Text type='tertiary' size='small'>
              {record.order?.payment_trade_no || record.order?.payment_method || '-'}
            </Text>
          </div>
        ),
      },
      {
        title: t('买家'),
        key: 'buyer',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{`user #${record.order?.buyer_user_id || '-'}`}</Text>
            <Text type='tertiary' size='small'>
              {record.order?.payment_method || '-'}
            </Text>
          </div>
        ),
      },
      {
        title: t('金额'),
        key: 'amount',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>
              {formatMinorAmount(
                record.order?.payable_amount_minor,
                record.order?.currency,
              )}
            </Text>
            <Text type='tertiary' size='small'>
              {`${t('原价')} ${formatMinorAmount(
                record.order?.total_amount_minor,
                record.order?.currency,
              )}`}
            </Text>
          </div>
        ),
      },
      {
        title: t('状态'),
        key: 'status',
        render: (_, record) => (
          <div className='flex flex-wrap gap-1 justify-end md:justify-start'>
            {renderTag(
              record.order?.order_status,
              getOrderStatusColor(record.order?.order_status),
            )}
            {renderTag(
              record.order?.payment_status,
              getPaymentStatusColor(record.order?.payment_status),
            )}
            {renderTag(
              record.order?.entitlement_status,
              getEntitlementStatusColor(record.order?.entitlement_status),
            )}
          </div>
        ),
      },
      {
        title: t('商品明细'),
        key: 'items',
        render: (_, record) => renderItemsSummary(record.items),
      },
      {
        title: t('时间'),
        key: 'time',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>
              {record.order?.created_at
                ? timestamp2string(record.order.created_at)
                : '-'}
            </Text>
            <Text type='tertiary' size='small'>
              {record.order?.paid_at
                ? `${t('支付')} ${timestamp2string(record.order.paid_at)}`
                : t('未支付')}
            </Text>
          </div>
        ),
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
            <Text strong>{t('订单管理')}</Text>
            <Text type='tertiary'>
              {t('查看交易订单、支付状态与发权执行结果')}
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
            placeholder={t('搜索订单号 / 支付单号 / 支付方式')}
            showClear
          />
          <Input
            value={draftFilters.buyerUserId}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, buyerUserId: value }))
            }
            placeholder={t('买家 ID')}
            showClear
          />
          <Select
            value={draftFilters.orderStatus}
            optionList={[
              { label: t('全部订单状态'), value: '' },
              { label: t('待支付'), value: 'pending_payment' },
              { label: t('已支付'), value: 'paid' },
              { label: t('已关闭'), value: 'closed' },
            ]}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, orderStatus: value || '' }))
            }
          />
          <Select
            value={draftFilters.paymentStatus}
            optionList={[
              { label: t('全部支付状态'), value: '' },
              { label: t('未支付'), value: 'unpaid' },
              { label: t('已支付'), value: 'paid' },
              { label: t('失败'), value: 'failed' },
            ]}
            onChange={(value) =>
              setDraftFilters((prev) => ({ ...prev, paymentStatus: value || '' }))
            }
          />
          <Select
            value={draftFilters.entitlementStatus}
            optionList={[
              { label: t('全部发权状态'), value: '' },
              { label: t('待发权'), value: 'pending' },
              { label: t('已发权'), value: 'created' },
              { label: t('发权失败'), value: 'failed' },
            ]}
            onChange={(value) =>
              setDraftFilters((prev) => ({
                ...prev,
                entitlementStatus: value || '',
              }))
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
        total: orderCount,
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
        dataSource={orders}
        loading={loading}
        hidePagination={true}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        empty={<Empty description={t('暂无订单数据')} style={{ padding: 32 }} />}
        className='rounded-xl overflow-hidden'
      />
    </CardPro>
  );
};

export default OrdersPage;
