import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';

const PAGE_SIZE_KEY = 'marketplace-sellers-page-size';

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

export const useSellersData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode(
    'marketplace-sellers',
  );
  const [sellers, setSellers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actingId, setActingId] = useState(null);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(getInitialPageSize);
  const [sellerCount, setSellerCount] = useState(0);
  const [filters, setFilters] = useState({
    keyword: '',
    status: '',
  });
  const [draftFilters, setDraftFilters] = useState({
    keyword: '',
    status: '',
  });

  const fetchSellers = useCallback(
    async (page = activePage, size = pageSize, nextFilters = filters) => {
      setLoading(true);
      try {
        const query = buildQueryString({
          keyword: nextFilters.keyword,
          status: nextFilters.status,
          p: page,
          page_size: size,
        });
        const res = await API.get(`/api/seller/admin?${query}`);
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        const items = (data.items || []).map((seller) => ({
          ...seller,
          key: seller.id,
        }));
        setSellers(items);
        setSellerCount(data.total || 0);
      } catch (error) {
        showError(error);
      } finally {
        setLoading(false);
      }
    },
    [activePage, filters, pageSize],
  );

  useEffect(() => {
    fetchSellers(activePage, pageSize, filters);
  }, [activePage, pageSize, filters, fetchSellers]);

  const refresh = useCallback(() => {
    fetchSellers(activePage, pageSize, filters);
  }, [activePage, fetchSellers, filters, pageSize]);

  const applyFilters = useCallback(() => {
    setActivePage(1);
    setFilters({ ...draftFilters });
  }, [draftFilters]);

  const resetFilters = useCallback(() => {
    const emptyFilters = { keyword: '', status: '' };
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

  const updateSellerStatus = useCallback(
    async (seller, status) => {
      setActingId(seller.id);
      try {
        const res = await API.put(`/api/seller/admin/${seller.id}/status`, {
          status,
          remark:
            status === 'disabled'
              ? 'updated_from_marketplace_console'
              : 'reactivated_from_marketplace_console',
        });
        const { success, message } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        showSuccess(t('卖家状态已更新'));
        refresh();
      } catch (error) {
        showError(error);
      } finally {
        setActingId(null);
      }
    },
    [refresh, t],
  );

  return {
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
  };
};
