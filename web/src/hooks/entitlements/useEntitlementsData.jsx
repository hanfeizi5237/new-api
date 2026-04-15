import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';

const PAGE_SIZE_KEY = 'marketplace-entitlements-page-size';

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

export const useEntitlementsData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode(
    'marketplace-entitlements',
  );
  const [entitlements, setEntitlements] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(getInitialPageSize);
  const [entitlementCount, setEntitlementCount] = useState(0);
  const [filters, setFilters] = useState({
    buyerUserId: '',
    modelName: '',
    status: '',
  });
  const [draftFilters, setDraftFilters] = useState({
    buyerUserId: '',
    modelName: '',
    status: '',
  });

  const fetchEntitlements = useCallback(
    async (page = activePage, size = pageSize, nextFilters = filters) => {
      setLoading(true);
      try {
        const query = buildQueryString({
          buyer_user_id: nextFilters.buyerUserId,
          model_name: nextFilters.modelName,
          status: nextFilters.status,
          p: page,
          page_size: size,
        });
        const res = await API.get(`/api/marketplace/admin/entitlements?${query}`);
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        const items = (data.items || []).map((item) => ({
          ...item,
          key: item.id,
        }));
        setEntitlements(items);
        setEntitlementCount(data.total || 0);
      } catch (error) {
        showError(error);
      } finally {
        setLoading(false);
      }
    },
    [activePage, filters, pageSize],
  );

  useEffect(() => {
    fetchEntitlements(activePage, pageSize, filters);
  }, [activePage, pageSize, filters, fetchEntitlements]);

  const refresh = useCallback(() => {
    fetchEntitlements(activePage, pageSize, filters);
  }, [activePage, fetchEntitlements, filters, pageSize]);

  const applyFilters = useCallback(() => {
    setActivePage(1);
    setFilters({ ...draftFilters });
  }, [draftFilters]);

  const resetFilters = useCallback(() => {
    const emptyFilters = {
      buyerUserId: '',
      modelName: '',
      status: '',
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
  };
};
