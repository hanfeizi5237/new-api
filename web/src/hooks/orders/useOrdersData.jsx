import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';

const PAGE_SIZE_KEY = 'marketplace-orders-page-size';

const getInitialPageSize = () =>
  parseInt(localStorage.getItem(PAGE_SIZE_KEY), 10) || ITEMS_PER_PAGE;

const buildQueryString = (params) => {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      query.set(key, `${value}`);
    }
  });
  return query.toString();
};

export const useOrdersData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode(
    'marketplace-orders',
  );
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(getInitialPageSize);
  const [orderCount, setOrderCount] = useState(0);
  const [filters, setFilters] = useState({
    keyword: '',
    buyerUserId: '',
    orderStatus: '',
    paymentStatus: '',
    entitlementStatus: '',
  });
  const [draftFilters, setDraftFilters] = useState({
    keyword: '',
    buyerUserId: '',
    orderStatus: '',
    paymentStatus: '',
    entitlementStatus: '',
  });

  const fetchOrders = useCallback(
    async (page = activePage, size = pageSize, nextFilters = filters) => {
      setLoading(true);
      try {
        const query = buildQueryString({
          keyword: nextFilters.keyword,
          buyer_user_id: nextFilters.buyerUserId,
          order_status: nextFilters.orderStatus,
          payment_status: nextFilters.paymentStatus,
          entitlement_status: nextFilters.entitlementStatus,
          p: page,
          page_size: size,
        });
        const res = await API.get(`/api/marketplace/admin/orders?${query}`);
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        const items = (data.items || []).map((item, index) => ({
          ...item,
          key: item?.order?.id || index,
        }));
        setOrders(items);
        setOrderCount(data.total || 0);
      } catch (error) {
        showError(error);
      } finally {
        setLoading(false);
      }
    },
    [activePage, filters, pageSize],
  );

  useEffect(() => {
    fetchOrders(activePage, pageSize, filters);
  }, [activePage, pageSize, filters, fetchOrders]);

  const refresh = useCallback(() => {
    fetchOrders(activePage, pageSize, filters);
  }, [activePage, fetchOrders, filters, pageSize]);

  const applyFilters = useCallback(() => {
    setActivePage(1);
    setFilters({ ...draftFilters });
  }, [draftFilters]);

  const resetFilters = useCallback(() => {
    const emptyFilters = {
      keyword: '',
      buyerUserId: '',
      orderStatus: '',
      paymentStatus: '',
      entitlementStatus: '',
    };
    setDraftFilters(emptyFilters);
    setActivePage(1);
    setFilters(emptyFilters);
  }, []);

  const handlePageChange = useCallback((page) => {
    setActivePage(page);
  }, []);

  const handlePageSizeChange = useCallback((size) => {
    localStorage.setItem(PAGE_SIZE_KEY, `${size}`);
    setPageSize(size);
    setActivePage(1);
  }, []);

  return {
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
  };
};
