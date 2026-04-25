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

// ========== 页面配置常量 ==========
export const USER_USAGE_PAGE_SIZE = 20;
export const MAX_DATE_RANGE_DAYS = 31;

// ========== 时间粒度选项 ==========
export const GRANULARITY_OPTIONS = [
  { label: '天', value: 'day' },
  { label: '周', value: 'week' },
  { label: '月', value: 'month' },
];

// ========== 默认时间范围（最近7天） ==========
export const DEFAULT_DATE_RANGE = () => {
  const now = new Date();
  const start = new Date(now);
  start.setDate(start.getDate() - 7);
  return {
    start: start.toISOString().split('T')[0],
    end: now.toISOString().split('T')[0],
  };
};

// ========== 图表颜色 ==========
export const USER_COLORS = [
  '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
  '#ec4899', '#06b6d4', '#f97316', '#6366f1', '#14b8a6',
];

export const MODEL_COLORS = [
  '#1664FF', '#1AC6FF', '#FF8A00', '#3CC780', '#7442D4',
  '#FFC400', '#304D77', '#B48DEB', '#009488', '#FF7DDA',
];

// ========== 统计卡片配置 ==========
export const STATS_CARDS = [
  { key: 'totalUsers', label: '总用户数', icon: 'users', color: 'blue' },
  { key: 'totalCount', label: '总调用次数', icon: 'activity', color: 'green' },
  { key: 'totalQuota', label: '总消耗额度', icon: 'credit-card', color: 'orange' },
  { key: 'totalTokens', label: '总 Token 用量', icon: 'file-text', color: 'purple' },
  { key: 'totalErrors', label: '总错误数', icon: 'alert-circle', color: 'red' },
];

// ========== 详情抽屉 Tab 配置 ==========
export const DETAIL_TABS = [
  { key: 'model', label: '模型分布' },
  { key: 'time', label: '时间趋势' },
  { key: 'errors', label: '错误统计' },
];

// ========== 导出格式 ==========
export const EXPORT_FORMATS = [
  { label: 'CSV', value: 'csv' },
];
