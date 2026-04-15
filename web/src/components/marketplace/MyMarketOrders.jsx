import React, { useMemo } from 'react';
import { Button, Empty, Space, Tag, Typography } from '@douyinfe/semi-ui';
import CardPro from '../common/ui/CardPro';
import CardTable from '../common/ui/CardTable';
import { timestamp2string } from '../../helpers/utils';

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

const renderItemSummary = (items = []) => {
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

const MyMarketOrders = ({
  orders,
  loading,
  payingOrderId,
  onPayOrder,
  onRefresh,
  watchStatusText,
  t,
}) => {
  const columns = useMemo(
    () => [
      {
        title: t('订单'),
        key: 'order',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text strong>{record.order?.order_no || '-'}</Text>
            <Text type='tertiary' size='small'>
              {record.order?.payment_trade_no ||
                record.order?.payment_method ||
                '-'}
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
              {record.order?.payment_method || t('未选择支付方式')}
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
        title: t('商品'),
        key: 'items',
        render: (_, record) => renderItemSummary(record.items),
      },
      {
        title: t('时间'),
        key: 'created_at',
        render: (_, record) =>
          record.order?.created_at
            ? timestamp2string(record.order.created_at)
            : '-',
      },
      {
        title: t('操作'),
        key: 'actions',
        render: (_, record) => {
          const order = record.order || {};
          const canRepay =
            order.order_status === 'pending_payment' &&
            order.payment_status !== 'paid';
          if (!canRepay) {
            return '-';
          }
          return (
            <Button
              size='small'
              type='tertiary'
              loading={payingOrderId === order.id}
              onClick={() => onPayOrder(order.id, order.payment_method)}
            >
              {t('继续支付')}
            </Button>
          );
        },
      },
    ],
    [onPayOrder, payingOrderId, t],
  );

  return (
    <CardPro
      type='type1'
      descriptionArea={
        <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-2'>
          <div className='flex flex-col'>
            <Text strong>{t('我的订单')}</Text>
            <Text type='tertiary'>
              {watchStatusText ||
                t('支付完成后会自动刷新订单与 entitlement 状态')}
            </Text>
          </div>
          <Space>
            <Button type='tertiary' onClick={onRefresh}>
              {t('刷新订单')}
            </Button>
          </Space>
        </div>
      }
      t={t}
    >
      <CardTable
        rowKey='key'
        columns={columns}
        dataSource={orders}
        loading={loading}
        hidePagination={true}
        scroll={{ x: 'max-content' }}
        empty={<Empty description={t('暂无订单')} style={{ padding: 24 }} />}
        className='rounded-xl overflow-hidden'
      />
    </CardPro>
  );
};

export default MyMarketOrders;
