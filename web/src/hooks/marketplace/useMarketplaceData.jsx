import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  API,
  showError,
  showInfo,
  showSuccess,
  showWarning,
} from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';

const LISTING_PAGE_SIZE = 6;
const ORDER_PAGE_SIZE = ITEMS_PER_PAGE;
const ENTITLEMENT_PAGE_SIZE = 6;
const WATCH_INTERVAL_MS = 4000;
const WATCH_MAX_TICKS = 15;

const buildQueryString = (params) => {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      query.set(key, `${value}`);
    }
  });
  return query.toString();
};

const dedupePaymentOptions = (options) => {
  const seen = new Set();
  return options.filter((option) => {
    if (!option?.value || seen.has(option.value)) {
      return false;
    }
    seen.add(option.value);
    return true;
  });
};

const buildPaymentOptions = (topupInfo, t) => {
  let payMethods = topupInfo?.pay_methods || [];
  if (typeof payMethods === 'string') {
    try {
      payMethods = JSON.parse(payMethods);
    } catch {
      payMethods = [];
    }
  }

  const options = (payMethods || [])
    .filter((method) => method?.type && method?.name)
    .map((method) => ({
      label: method.name,
      value: method.type,
    }));

  if (topupInfo?.enable_creem_topup) {
    options.push({ label: 'Creem', value: 'creem' });
  }
  if (topupInfo?.enable_waffo_topup) {
    options.push({ label: 'Waffo', value: 'waffo' });
  }

  if (options.length === 0) {
    options.push({ label: t('Stripe'), value: 'stripe' });
  }

  return dedupePaymentOptions(options);
};

const createMarketplaceIdempotencyKey = () =>
  `market-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

const submitPaymentForm = (formURL, formParams = {}) => {
  const form = document.createElement('form');
  form.action = formURL;
  form.method = 'POST';
  form.target = '_blank';
  Object.entries(formParams).forEach(([key, value]) => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = key;
    input.value = value;
    form.appendChild(input);
  });
  document.body.appendChild(form);
  form.submit();
  document.body.removeChild(form);
};

const normalizeListingItems = (items = []) =>
  items.map((item, index) => ({
    ...item,
    key: item?.listing?.id || index,
  }));

const normalizeOrderItems = (items = []) =>
  items.map((item, index) => ({
    ...item,
    key: item?.order?.id || index,
  }));

const normalizeEntitlementItems = (items = []) =>
  items.map((item) => ({
    ...item,
    key: item.id,
  }));

export const useMarketplaceData = () => {
  const { t } = useTranslation();

  const [listingKeyword, setListingKeyword] = useState('');
  const [listingDraftKeyword, setListingDraftKeyword] = useState('');
  const [listingPage, setListingPage] = useState(1);
  const [listings, setListings] = useState([]);
  const [listingCount, setListingCount] = useState(0);
  const [listingsLoading, setListingsLoading] = useState(true);

  const [orders, setOrders] = useState([]);
  const [ordersLoading, setOrdersLoading] = useState(true);
  const [orderCount, setOrderCount] = useState(0);

  const [entitlements, setEntitlements] = useState([]);
  const [entitlementsLoading, setEntitlementsLoading] = useState(true);

  const [paymentOptions, setPaymentOptions] = useState([]);
  const [paymentOptionsLoading, setPaymentOptionsLoading] = useState(true);

  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedListing, setSelectedListing] = useState(null);
  const [selectedSkuId, setSelectedSkuId] = useState(null);
  const [purchaseQuantity, setPurchaseQuantity] = useState(1);
  const [paymentMethod, setPaymentMethod] = useState('');

  const [purchaseLoading, setPurchaseLoading] = useState(false);
  const [payingOrderId, setPayingOrderId] = useState(null);
  const [watchingOrderId, setWatchingOrderId] = useState(null);
  const [watchStatusText, setWatchStatusText] = useState('');

  const watchTickRef = useRef(0);

  const selectedSku = useMemo(() => {
    if (!selectedListing?.skus?.length) {
      return null;
    }
    return (
      selectedListing.skus.find((sku) => sku.id === selectedSkuId) ||
      selectedListing.skus[0]
    );
  }, [selectedListing, selectedSkuId]);

  const purchaseTotalMinor = useMemo(() => {
    if (!selectedSku) {
      return 0;
    }
    return (
      Number(selectedSku.unit_price_minor || 0) * Number(purchaseQuantity || 1)
    );
  }, [purchaseQuantity, selectedSku]);

  useEffect(() => {
    if (!selectedSku) {
      return;
    }
    const nextMin = Number(selectedSku.min_quantity || 1);
    const nextMax = Number(selectedSku.max_quantity || nextMin);
    if (purchaseQuantity < nextMin) {
      setPurchaseQuantity(nextMin);
      return;
    }
    if (purchaseQuantity > nextMax) {
      setPurchaseQuantity(nextMax);
    }
  }, [purchaseQuantity, selectedSku]);

  const loadListings = useCallback(
    async (page = listingPage, keyword = listingKeyword) => {
      setListingsLoading(true);
      try {
        const query = buildQueryString({
          keyword,
          p: page,
          page_size: LISTING_PAGE_SIZE,
        });
        const res = await API.get(`/api/market/listings?${query}`);
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return [];
        }
        const items = normalizeListingItems(data.items || []);
        setListings(items);
        setListingCount(data.total || 0);
        return items;
      } catch (error) {
        showError(error);
        return [];
      } finally {
        setListingsLoading(false);
      }
    },
    [listingKeyword, listingPage],
  );

  const loadOrders = useCallback(async () => {
    setOrdersLoading(true);
    try {
      const query = buildQueryString({
        p: 1,
        page_size: ORDER_PAGE_SIZE,
      });
      const res = await API.get(`/api/market/orders?${query}`);
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return [];
      }
      const items = normalizeOrderItems(data.items || []);
      setOrders(items);
      setOrderCount(data.total || 0);
      return items;
    } catch (error) {
      showError(error);
      return [];
    } finally {
      setOrdersLoading(false);
    }
  }, []);

  const loadEntitlements = useCallback(async () => {
    setEntitlementsLoading(true);
    try {
      const query = buildQueryString({
        p: 1,
        page_size: ENTITLEMENT_PAGE_SIZE,
      });
      const res = await API.get(`/api/market/entitlements?${query}`);
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return [];
      }
      const items = normalizeEntitlementItems(data.items || []);
      setEntitlements(items);
      return items;
    } catch (error) {
      showError(error);
      return [];
    } finally {
      setEntitlementsLoading(false);
    }
  }, []);

  const loadPaymentOptions = useCallback(async () => {
    setPaymentOptionsLoading(true);
    try {
      const res = await API.get('/api/user/topup/info');
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return [];
      }
      const options = buildPaymentOptions(data, t);
      setPaymentOptions(options);
      return options;
    } catch (error) {
      showError(error);
      return [];
    } finally {
      setPaymentOptionsLoading(false);
    }
  }, [t]);

  const refreshAll = useCallback(async () => {
    const [, nextOrders] = await Promise.all([
      loadListings(listingPage, listingKeyword),
      loadOrders(),
      loadEntitlements(),
    ]);
    return nextOrders;
  }, [listingKeyword, listingPage, loadEntitlements, loadListings, loadOrders]);

  useEffect(() => {
    loadListings(listingPage, listingKeyword);
  }, [listingKeyword, listingPage, loadListings]);

  useEffect(() => {
    loadOrders();
    loadEntitlements();
    loadPaymentOptions();
  }, [loadEntitlements, loadOrders, loadPaymentOptions]);

  useEffect(() => {
    if (!selectedListing?.skus?.length) {
      setSelectedSkuId(null);
      setPurchaseQuantity(1);
      return;
    }
    const fallbackSku =
      selectedListing.skus.find((sku) => sku.status === 'active') ||
      selectedListing.skus[0];
    setSelectedSkuId(fallbackSku?.id || null);
    setPurchaseQuantity(fallbackSku?.min_quantity || 1);
  }, [selectedListing]);

  useEffect(() => {
    if (!paymentOptions.length) {
      return;
    }
    if (!paymentMethod) {
      setPaymentMethod(paymentOptions[0].value);
    }
  }, [paymentMethod, paymentOptions]);

  const upsertOrder = useCallback((orderView) => {
    if (!orderView?.order?.id) {
      return;
    }
    setOrders((prev) => {
      const next = [...prev];
      const index = next.findIndex(
        (item) => item.order?.id === orderView.order.id,
      );
      if (index >= 0) {
        next[index] = {
          ...orderView,
          key: orderView.order.id,
        };
        return next;
      }
      return [{ ...orderView, key: orderView.order.id }, ...next].slice(
        0,
        ORDER_PAGE_SIZE,
      );
    });
  }, []);

  const fetchOrderDetail = useCallback(
    async (orderId, { silent = false } = {}) => {
      try {
        const res = await API.get(`/api/market/orders/${orderId}`, {
          disableDuplicate: true,
        });
        const { success, message, data } = res.data;
        if (!success) {
          if (!silent) {
            showError(message);
          }
          return null;
        }
        const orderView = {
          ...data,
          key: data?.order?.id || orderId,
        };
        upsertOrder(orderView);
        return orderView;
      } catch (error) {
        if (!silent) {
          showError(error);
        }
        return null;
      }
    },
    [upsertOrder],
  );

  const processPaymentIntent = useCallback(
    (intent) => {
      if (intent?.redirect_url) {
        window.open(intent.redirect_url, '_blank');
        showSuccess(t('已打开支付页面'));
        return;
      }
      if (intent?.form_url) {
        submitPaymentForm(intent.form_url, intent.form_params || {});
        showSuccess(t('已发起支付'));
        return;
      }
      showInfo(t('订单已创建，请等待支付状态刷新'));
    },
    [t],
  );

  const startWatchingOrder = useCallback(
    (orderId) => {
      if (!orderId || orderId <= 0) {
        return;
      }
      // Reset the polling budget every time we launch a fresh payment/return reconciliation cycle.
      watchTickRef.current = 0;
      setWatchingOrderId(orderId);
      setWatchStatusText(t('正在同步支付与权益状态...'));
    },
    [t],
  );

  useEffect(() => {
    if (!watchingOrderId) {
      return undefined;
    }

    const timer = window.setInterval(async () => {
      watchTickRef.current += 1;
      const orderView = await fetchOrderDetail(watchingOrderId, {
        silent: true,
      });
      if (!orderView?.order) {
        if (watchTickRef.current >= WATCH_MAX_TICKS) {
          setWatchStatusText('');
          setWatchingOrderId(null);
        }
        return;
      }

      const order = orderView.order;
      // Marketplace orders are only "done" after both payment and entitlement grant have landed.
      const isCompleted =
        order.payment_status === 'paid' &&
        order.entitlement_status === 'created';
      const isFailed =
        order.payment_status === 'failed' ||
        order.entitlement_status === 'failed';

      if (isCompleted) {
        await loadEntitlements();
        setWatchStatusText('');
        setWatchingOrderId(null);
        showSuccess(t('支付与权益已同步完成'));
        return;
      }
      if (isFailed || watchTickRef.current >= WATCH_MAX_TICKS) {
        await refreshAll();
        setWatchStatusText('');
        setWatchingOrderId(null);
        if (isFailed) {
          showWarning(t('订单处理未完成，请稍后刷新或重新支付'));
        }
      }
    }, WATCH_INTERVAL_MS);

    return () => {
      window.clearInterval(timer);
    };
  }, [fetchOrderDetail, loadEntitlements, refreshAll, t, watchingOrderId]);

  const openListingDetail = useCallback((listingDetail) => {
    setSelectedListing(listingDetail);
    setDetailVisible(true);
  }, []);

  const closeListingDetail = useCallback(() => {
    setDetailVisible(false);
    setSelectedListing(null);
  }, []);

  const applyListingKeyword = useCallback(() => {
    setListingPage(1);
    setListingKeyword(listingDraftKeyword);
  }, [listingDraftKeyword]);

  const resetListingKeyword = useCallback(() => {
    setListingDraftKeyword('');
    setListingKeyword('');
    setListingPage(1);
  }, []);

  const payExistingOrder = useCallback(
    async (orderId, customPaymentMethod = '') => {
      const method = customPaymentMethod || paymentMethod;
      if (!method) {
        showInfo(t('请选择支付方式'));
        return;
      }
      setPayingOrderId(orderId);
      try {
        const res = await API.post(`/api/market/orders/${orderId}/pay`, {
          payment_method: method,
        });
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        processPaymentIntent(data);
        startWatchingOrder(orderId);
        await loadOrders();
      } catch (error) {
        showError(error);
      } finally {
        setPayingOrderId(null);
      }
    },
    [loadOrders, paymentMethod, processPaymentIntent, startWatchingOrder, t],
  );

  const createOrderAndPay = useCallback(async () => {
    if (!selectedListing?.listing?.id || !selectedSku?.id) {
      showInfo(t('请选择可购买的商品与 SKU'));
      return;
    }
    if (!paymentMethod) {
      showInfo(t('请选择支付方式'));
      return;
    }
    setPurchaseLoading(true);
    try {
      const createRes = await API.post('/api/market/orders', {
        listing_id: selectedListing.listing.id,
        sku_id: selectedSku.id,
        quantity: Number(purchaseQuantity || 1),
        idempotency_key: createMarketplaceIdempotencyKey(),
      });
      const { success, message, data } = createRes.data;
      if (!success) {
        showError(message);
        return;
      }
      upsertOrder({
        ...data,
        key: data?.order?.id,
      });
      closeListingDetail();
      await payExistingOrder(data.order.id, paymentMethod);
    } catch (error) {
      showError(error);
    } finally {
      setPurchaseLoading(false);
    }
  }, [
    closeListingDetail,
    payExistingOrder,
    paymentMethod,
    purchaseQuantity,
    selectedListing,
    selectedSku,
    t,
    upsertOrder,
  ]);

  const handlePaymentReturn = useCallback(
    async (kind = 'success') => {
      if (kind === 'cancel') {
        showInfo(t('已返回集市页面，可重新发起支付'));
      } else {
        showSuccess(t('已返回集市页面，正在同步最新支付结果'));
      }
      const nextOrders = await refreshAll();
      // PSP return pages may come back before the webhook commits, so resume polling on the newest unfinished order.
      const pendingOrder = nextOrders.find(
        (item) =>
          item?.order?.payment_status !== 'paid' ||
          item?.order?.entitlement_status !== 'created',
      );
      if (pendingOrder?.order?.id) {
        startWatchingOrder(pendingOrder.order.id);
      }
    },
    [refreshAll, startWatchingOrder, t],
  );

  const entitlementSummary = useMemo(() => {
    const totalGranted = entitlements.reduce(
      (sum, item) => sum + Number(item.total_granted || 0),
      0,
    );
    const totalUsed = entitlements.reduce(
      (sum, item) => sum + Number(item.total_used || 0),
      0,
    );
    const totalFrozen = entitlements.reduce(
      (sum, item) => sum + Number(item.total_frozen || 0),
      0,
    );
    const totalRemaining = entitlements.reduce((sum, item) => {
      const remaining =
        Number(item.total_granted || 0) -
        Number(item.total_used || 0) -
        Number(item.total_frozen || 0) -
        Number(item.total_refunded || 0);
      return sum + Math.max(remaining, 0);
    }, 0);
    return {
      totalGranted,
      totalUsed,
      totalFrozen,
      totalRemaining,
    };
  }, [entitlements]);

  return {
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
    orderCount,
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
    watchingOrderId,
    watchStatusText,
    refreshAll,
    openListingDetail,
    closeListingDetail,
    createOrderAndPay,
    payExistingOrder,
    handlePaymentReturn,
  };
};
