import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';

const PAGE_SIZE_KEY = 'marketplace-listings-page-size';

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

const listingActionPayloads = {
  submit: {
    audit_status: 'pending_review',
    audit_remark: 'submitted_from_marketplace_console',
  },
  approve: {
    status: 'active',
    audit_status: 'approved',
    audit_remark: 'approved_from_marketplace_console',
  },
  reject: {
    status: 'paused',
    audit_status: 'rejected',
    audit_remark: 'rejected_from_marketplace_console',
  },
  pause: {
    status: 'paused',
    audit_remark: 'paused_from_marketplace_console',
  },
  activate: {
    status: 'active',
    audit_remark: 'activated_from_marketplace_console',
  },
};

export const useListingsData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode(
    'marketplace-listings',
  );
  const [listings, setListings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actingId, setActingId] = useState(null);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(getInitialPageSize);
  const [listingCount, setListingCount] = useState(0);
  const [filters, setFilters] = useState({
    keyword: '',
    status: '',
    auditStatus: '',
    sellerId: '',
  });
  const [draftFilters, setDraftFilters] = useState({
    keyword: '',
    status: '',
    auditStatus: '',
    sellerId: '',
  });

  const fetchListings = useCallback(
    async (page = activePage, size = pageSize, nextFilters = filters) => {
      setLoading(true);
      try {
        const query = buildQueryString({
          keyword: nextFilters.keyword,
          status: nextFilters.status,
          audit_status: nextFilters.auditStatus,
          seller_id: nextFilters.sellerId,
          p: page,
          page_size: size,
        });
        const res = await API.get(`/api/listing/admin?${query}`);
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        const items = (data.items || []).map((item, index) => ({
          ...item,
          key: item?.listing?.id || index,
        }));
        setListings(items);
        setListingCount(data.total || 0);
      } catch (error) {
        showError(error);
      } finally {
        setLoading(false);
      }
    },
    [activePage, filters, pageSize],
  );

  useEffect(() => {
    fetchListings(activePage, pageSize, filters);
  }, [activePage, pageSize, filters, fetchListings]);

  const refresh = useCallback(() => {
    fetchListings(activePage, pageSize, filters);
  }, [activePage, fetchListings, filters, pageSize]);

  const applyFilters = useCallback(() => {
    setActivePage(1);
    setFilters({ ...draftFilters });
  }, [draftFilters]);

  const resetFilters = useCallback(() => {
    const emptyFilters = {
      keyword: '',
      status: '',
      auditStatus: '',
      sellerId: '',
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

  const updateListingStatus = useCallback(
    async (listingId, payload) => {
      setActingId(listingId);
      try {
        const res = await API.put(`/api/listing/admin/${listingId}/status`, payload);
        const { success, message } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        showSuccess(t('商品状态已更新'));
        refresh();
      } catch (error) {
        showError(error);
      } finally {
        setActingId(null);
      }
    },
    [refresh, t],
  );

  const runListingAction = useCallback(
    async (record, action) => {
      const payload = listingActionPayloads[action];
      if (!payload || !record?.listing?.id) {
        return;
      }
      await updateListingStatus(record.listing.id, payload);
    },
    [updateListingStatus],
  );

  return {
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
  };
};
