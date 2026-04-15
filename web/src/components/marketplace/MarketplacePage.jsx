import React from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  Banner,
  Button,
  Card,
  Empty,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import CardPro from '../common/ui/CardPro';
import { createCardProPagination } from '../../helpers/utils';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { useMarketplaceData } from '../../hooks/marketplace/useMarketplaceData';
import MyMarketOrders from './MyMarketOrders';

const { Text, Title } = Typography;

const formatMinorAmount = (value, currency = 'CNY') =>
  `${currency} ${(Number(value || 0) / 100).toFixed(2)}`;

const renderInventoryTag = (availableAmount = 0) => {
  if (availableAmount <= 0) {
    return (
      <Tag color='red' shape='circle'>
        售罄
      </Tag>
    );
  }
  if (availableAmount < 1000) {
    return (
      <Tag color='orange' shape='circle'>
        库存紧张
      </Tag>
    );
  }
  return (
    <Tag color='green' shape='circle'>
      可购买
    </Tag>
  );
};

const calculateRemaining = (record) =>
  Math.max(
    Number(record.total_granted || 0) -
      Number(record.total_used || 0) -
      Number(record.total_frozen || 0) -
      Number(record.total_refunded || 0),
    0,
  );

const SummaryCard = ({ title, value, hint }) => (
  <Card className='!rounded-2xl h-full'>
    <div className='flex flex-col gap-1'>
      <Text type='tertiary'>{title}</Text>
      <Title heading={4} style={{ margin: 0 }}>
        {value}
      </Title>
      <Text type='tertiary' size='small'>
        {hint}
      </Text>
    </div>
  </Card>
);

const ListingCard = ({ item, onOpen, t }) => {
  const listing = item.listing || {};
  const primarySku = item.skus?.[0];

  return (
    <Card className='!rounded-2xl h-full'>
      <div className='flex flex-col gap-4 h-full'>
        <div className='flex items-start justify-between gap-2'>
          <div className='flex flex-col gap-1'>
            <Text strong>{listing.title || '-'}</Text>
            <Text type='tertiary' size='small'>
              {listing.model_name || '-'}
            </Text>
          </div>
          {renderInventoryTag(item.inventory?.available_amount || 0)}
        </div>

        <div className='flex flex-wrap gap-2'>
          <Tag color='blue' shape='circle'>
            {listing.sale_mode || 'fixed_price'}
          </Tag>
          <Tag color='grey' shape='circle'>
            {listing.pricing_unit || 'per_token_package'}
          </Tag>
          {listing.validity_days > 0 && (
            <Tag color='purple' shape='circle'>
              {`${listing.validity_days}${t('天有效')}`}
            </Tag>
          )}
        </div>

        <div className='flex flex-col gap-1 min-h-[52px]'>
          <Text>
            {primarySku
              ? `${primarySku.package_amount} ${primarySku.package_unit}`
              : t('暂无可售 SKU')}
          </Text>
          <Text type='tertiary' size='small'>
            {primarySku
              ? `${formatMinorAmount(primarySku.unit_price_minor)} / ${t('份')}`
              : listing.description || t('等待商品配置')}
          </Text>
        </div>

        <div className='mt-auto flex items-center justify-between'>
          <Text type='tertiary' size='small'>
            {item.skus?.length
              ? `${item.skus.length}${t('个 SKU 可选')}`
              : t('暂无 SKU')}
          </Text>
          <Button
            type='primary'
            disabled={!item.skus?.length}
            onClick={() => onOpen(item)}
          >
            {t('查看并购买')}
          </Button>
        </div>
      </div>
    </Card>
  );
};

const MarketplacePage = () => {
  const marketplaceData = useMarketplaceData();
  const isMobile = useIsMobile();
  const [searchParams, setSearchParams] = useSearchParams();

  const {
    t,
    listings,
    listingCount,
    listingPage,
    listingDraftKeyword,
    setListingDraftKeyword,
    setListingPage,
    applyListingKeyword,
    resetListingKeyword,
    listingsLoading,
    orders,
    ordersLoading,
    entitlements,
    entitlementsLoading,
    entitlementSummary,
    paymentOptions,
    paymentOptionsLoading,
    detailVisible,
    selectedListing,
    selectedSku,
    selectedSkuId,
    setSelectedSkuId,
    purchaseQuantity,
    setPurchaseQuantity,
    paymentMethod,
    setPaymentMethod,
    purchaseTotalMinor,
    purchaseLoading,
    payingOrderId,
    watchStatusText,
    refreshAll,
    openListingDetail,
    closeListingDetail,
    createOrderAndPay,
    payExistingOrder,
    handlePaymentReturn,
  } = marketplaceData;

  React.useEffect(() => {
    const paymentReturned = searchParams.get('payment_return');
    const paymentCancelled = searchParams.get('payment_cancel');
    if (!paymentReturned && !paymentCancelled) {
      return;
    }
    handlePaymentReturn(paymentCancelled ? 'cancel' : 'success');
    setSearchParams({}, { replace: true });
  }, [handlePaymentReturn, searchParams, setSearchParams]);

  const skuOptions = (selectedListing?.skus || []).map((sku) => ({
    label: `${sku.sku_code} · ${sku.package_amount} ${sku.package_unit} · ${formatMinorAmount(sku.unit_price_minor)}`,
    value: sku.id,
  }));

  const purchaseMin = selectedSku?.min_quantity || 1;
  const purchaseMax = selectedSku?.max_quantity || 1;

  return (
    <div className='mt-[60px] px-2 space-y-4'>
      <CardPro
        type='type1'
        descriptionArea={
          <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-2'>
            <div className='flex flex-col'>
              <Text strong>{t('AI Token 集市')}</Text>
              <Text type='tertiary'>
                {t('浏览商品、完成购买，并在同一页查看订单与 entitlement 结果')}
              </Text>
            </div>
            <Space wrap>
              <Button type='tertiary' onClick={refreshAll}>
                {t('刷新状态')}
              </Button>
            </Space>
          </div>
        }
        searchArea={
          <div className='flex flex-col md:flex-row gap-2'>
            <Input
              value={listingDraftKeyword}
              onChange={setListingDraftKeyword}
              placeholder={t('搜索商品标题 / 模型')}
              showClear
            />
            <Space wrap>
              <Button type='primary' onClick={applyListingKeyword}>
                {t('查询商品')}
              </Button>
              <Button type='tertiary' onClick={resetListingKeyword}>
                {t('重置')}
              </Button>
            </Space>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: listingPage,
          pageSize: 6,
          total: listingCount,
          onPageChange: setListingPage,
          onPageSizeChange: () => {},
          showSizeChanger: false,
          isMobile,
          t,
        })}
        t={t}
      >
        {(watchStatusText || paymentOptionsLoading) && (
          <Banner
            type='info'
            closeIcon={null}
            className='!rounded-xl mb-4'
            description={
              watchStatusText || t('正在加载支付方式配置，请稍候...')
            }
          />
        )}

        <div className='grid grid-cols-1 lg:grid-cols-4 gap-3 mb-4'>
          <SummaryCard
            title={t('累计授予')}
            value={entitlementSummary.totalGranted}
            hint={t('所有已购 entitlement 授予总额')}
          />
          <SummaryCard
            title={t('累计消耗')}
            value={entitlementSummary.totalUsed}
            hint={t('已通过 relay 使用的额度')}
          />
          <SummaryCard
            title={t('当前冻结')}
            value={entitlementSummary.totalFrozen}
            hint={t('进行中的请求预冻结额度')}
          />
          <SummaryCard
            title={t('当前可用')}
            value={entitlementSummary.totalRemaining}
            hint={t('可继续消费的 entitlement 余额')}
          />
        </div>

        {entitlements.length > 0 && (
          <div className='grid grid-cols-1 lg:grid-cols-3 gap-3 mb-4'>
            {entitlements.map((entitlement) => (
              <Card className='!rounded-2xl' key={entitlement.id}>
                <div className='flex flex-col gap-1'>
                  <Text strong>{entitlement.model_name || '-'}</Text>
                  <Text type='tertiary' size='small'>
                    {`vendor #${entitlement.vendor_id || '-'}`}
                  </Text>
                  <div className='flex flex-wrap gap-2 mt-2'>
                    <Tag color='green' shape='circle'>
                      {`${t('可用')} ${calculateRemaining(entitlement)}`}
                    </Tag>
                    <Tag color='blue' shape='circle'>
                      {`${t('授予')} ${entitlement.total_granted || 0}`}
                    </Tag>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}

        {listings.length === 0 && !listingsLoading ? (
          <Empty description={t('暂无可购买商品')} style={{ padding: 32 }} />
        ) : (
          <div className='grid grid-cols-1 xl:grid-cols-3 gap-3'>
            {listings.map((item) => (
              <ListingCard
                key={item.key}
                item={item}
                onOpen={openListingDetail}
                t={t}
              />
            ))}
          </div>
        )}
      </CardPro>

      <MyMarketOrders
        orders={orders}
        loading={ordersLoading || entitlementsLoading}
        payingOrderId={payingOrderId}
        onPayOrder={payExistingOrder}
        onRefresh={refreshAll}
        watchStatusText={watchStatusText}
        t={t}
      />

      <Modal
        title={selectedListing?.listing?.title || t('商品详情')}
        visible={detailVisible}
        onCancel={closeListingDetail}
        onOk={createOrderAndPay}
        okText={t('创建订单并支付')}
        cancelText={t('关闭')}
        confirmLoading={purchaseLoading}
        okButtonProps={{
          disabled: !selectedSku || paymentOptionsLoading || !paymentMethod,
        }}
        width={720}
      >
        <div className='flex flex-col gap-4'>
          <div className='flex flex-col gap-1'>
            <Text strong>{selectedListing?.listing?.model_name || '-'}</Text>
            <Text type='tertiary'>
              {selectedListing?.listing?.description ||
                t('该商品为平台代理消费权益包，购买后可直接用于模型请求。')}
            </Text>
          </div>

          <div className='grid grid-cols-1 md:grid-cols-2 gap-3'>
            <div className='flex flex-col gap-2'>
              <Text type='tertiary'>{t('选择 SKU')}</Text>
              <Select
                value={selectedSkuId}
                optionList={skuOptions}
                onChange={(value) => setSelectedSkuId(value)}
              />
            </div>
            <div className='flex flex-col gap-2'>
              <Text type='tertiary'>{t('购买数量')}</Text>
              <InputNumber
                value={purchaseQuantity}
                min={purchaseMin}
                max={purchaseMax}
                onChange={(value) =>
                  setPurchaseQuantity(
                    Number.isFinite(Number(value))
                      ? Number(value)
                      : purchaseMin,
                  )
                }
              />
            </div>
          </div>

          <div className='grid grid-cols-1 md:grid-cols-2 gap-3'>
            <div className='flex flex-col gap-2'>
              <Text type='tertiary'>{t('支付方式')}</Text>
              <Select
                value={paymentMethod}
                optionList={paymentOptions}
                loading={paymentOptionsLoading}
                onChange={(value) => setPaymentMethod(value)}
              />
            </div>
            <div className='flex flex-col gap-2'>
              <Text type='tertiary'>{t('订单金额')}</Text>
              <Card className='!rounded-xl'>
                <div className='flex flex-col gap-1'>
                  <Text strong>
                    {formatMinorAmount(
                      purchaseTotalMinor,
                      selectedListing?.listing?.currency || 'CNY',
                    )}
                  </Text>
                  <Text type='tertiary' size='small'>
                    {selectedSku
                      ? `${selectedSku.package_amount} ${selectedSku.package_unit} x ${purchaseQuantity}`
                      : '-'}
                  </Text>
                </div>
              </Card>
            </div>
          </div>

          <div className='flex flex-wrap gap-2'>
            <Tag color='blue' shape='circle'>
              {`seller #${selectedListing?.listing?.seller_id || '-'}`}
            </Tag>
            <Tag color='grey' shape='circle'>
              {`listing #${selectedListing?.listing?.id || '-'}`}
            </Tag>
            {selectedListing?.inventory &&
              renderInventoryTag(selectedListing.inventory.available_amount)}
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default MarketplacePage;
