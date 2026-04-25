/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useState, useCallback, useRef } from 'react';
import { API, showError, showSuccess, renderQuota, renderNumber, getQuotaWithUnit } from '../../helpers';
import { DEFAULT_DATE_RANGE, MAX_DATE_RANGE_DAYS, EXPORT_FORMATS } from '../../constants/user-usage.constants';

export const useUserUsageData = () => {
  const [loading, setLoading] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [overviewData, setOverviewData] = useState([]);
  const [detailData, setDetailData] = useState(null);
  const [dateRange, setDateRange] = useState(DEFAULT_DATE_RANGE());
  const [granularity, setGranularity] = useState('day');
  const [selectedUser, setSelectedUser] = useState(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [activeDetailTab, setActiveDetailTab] = useState('model');

  const abortControllerRef = useRef(null);

  // ========== 日期范围工具 ==========
  const dateRangeToTimestamps = useCallback((range) => {
    const startTs = Math.floor(new Date(range.start + 'T00:00:00').getTime() / 1000);
    const endTs = Math.floor(new Date(range.end + 'T23:59:59').getTime() / 1000);
    return { startTs, endTs };
  }, []);

  const validateDateRange = useCallback((range) => {
    if (!range.start || !range.end) {
      showError('请选择日期范围');
      return false;
    }
    const start = new Date(range.start);
    const end = new Date(range.end);
    const diffDays = (end - start) / (1000 * 60 * 60 * 24);
    if (diffDays < 0) {
      showError('结束日期不能早于开始日期');
      return false;
    }
    if (diffDays > MAX_DATE_RANGE_DAYS) {
      showError(`日期范围不能超过 ${MAX_DATE_RANGE_DAYS} 天`);
      return false;
    }
    return true;
  }, []);

  // ========== API 调用 ==========
  const loadOverview = useCallback(async () => {
    if (!validateDateRange(dateRange)) return;

    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    try {
      const { startTs, endTs } = dateRangeToTimestamps(dateRange);
      const res = await API.get(
        `/api/admin/usage/overview?start_timestamp=${startTs}&end_timestamp=${endTs}&granularity=${granularity}`,
        { signal: controller.signal },
      );
      const { success, message, data } = res.data;
      if (success) {
        setOverviewData(data || []);
      } else {
        showError(message || '加载数据失败');
        setOverviewData([]);
      }
    } catch (err) {
      if (err.name !== 'AbortError') {
        showError('加载概览数据失败: ' + err.message);
        setOverviewData([]);
      }
    } finally {
      setLoading(false);
    }
  }, [dateRange, granularity, dateRangeToTimestamps, validateDateRange]);

  const loadDetail = useCallback(async (userId) => {
    if (!validateDateRange(dateRange)) return;

    setDetailLoading(true);
    try {
      const { startTs, endTs } = dateRangeToTimestamps(dateRange);
      const res = await API.get(
        `/api/admin/usage/detail?user_id=${userId}&start_timestamp=${startTs}&end_timestamp=${endTs}&granularity=${granularity}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setDetailData(data);
      } else {
        showError(message || '加载详情失败');
        setDetailData(null);
      }
    } catch (err) {
      showError('加载详情数据失败: ' + err.message);
      setDetailData(null);
    } finally {
      setDetailLoading(false);
    }
  }, [dateRange, granularity, dateRangeToTimestamps, validateDateRange]);

  const openUserDetail = useCallback(async (user) => {
    setSelectedUser(user);
    setDrawerVisible(true);
    setActiveDetailTab('model');
    await loadDetail(user.user_id);
  }, [loadDetail]);

  const closeUserDetail = useCallback(() => {
    setDrawerVisible(false);
    setSelectedUser(null);
    setDetailData(null);
  }, []);

  // ========== 汇总计算 ==========
  const getSummary = useCallback(() => {
    let totalUsers = overviewData.length;
    let totalCount = 0;
    let totalQuota = 0;
    let totalTokens = 0;
    let totalErrors = 0;

    overviewData.forEach((user) => {
      totalCount += user.total_count || 0;
      totalQuota += user.total_quota || 0;
      totalTokens += user.total_tokens || 0;
      totalErrors += user.error_count || 0;
    });

    return { totalUsers, totalCount, totalQuota, totalTokens, totalErrors };
  }, [overviewData]);

  // ========== 导出 CSV ==========
  const exportCSV = useCallback(() => {
    if (overviewData.length === 0) {
      showError('没有可导出的数据');
      return;
    }

    const headers = ['用户', '调用次数', '消耗额度', 'Token 用量', '错误数'];
    const rows = overviewData.map((user) => [
      user.username || `用户${user.user_id}`,
      user.total_count || 0,
      (user.total_quota / 1000000).toFixed(4),
      user.total_tokens || 0,
      user.error_count || 0,
    ]);

    // 添加汇总行
    const summary = getSummary();
    rows.push([
      '总计',
      summary.totalCount,
      (summary.totalQuota / 1000000).toFixed(4),
      summary.totalTokens,
      summary.totalErrors,
    ]);

    // 构建 CSV 内容
    const csvContent = [
      headers.join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n');

    // 添加 BOM 以支持 Excel 中文
    const bom = '\uFEFF';
    const blob = new Blob([bom + csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `用户用量统计_${dateRange.start}_${dateRange.end}.csv`;
    link.click();
    URL.revokeObjectURL(url);

    showSuccess('导出成功');
  }, [overviewData, dateRange, getSummary]);

  // ========== 日期变更 ==========
  const handleDateRangeChange = useCallback((newRange) => {
    setDateRange(newRange);
  }, []);

  const handleGranularityChange = useCallback((newGranularity) => {
    setGranularity(newGranularity);
  }, []);

  return {
    // 状态
    loading,
    detailLoading,
    overviewData,
    detailData,
    dateRange,
    granularity,
    selectedUser,
    drawerVisible,
    activeDetailTab,

    // 操作
    loadOverview,
    loadDetail,
    openUserDetail,
    closeUserDetail,
    setActiveDetailTab,
    handleDateRangeChange,
    handleGranularityChange,
    exportCSV,
    getSummary,

    // 工具
    renderQuota,
    renderNumber,
    getQuotaWithUnit,
  };
};
