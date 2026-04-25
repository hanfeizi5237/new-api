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

import { useState, useCallback, useEffect } from 'react';
import { initVChartSemiTheme } from '@visactor/vchart-semi-theme';
import { renderQuota, renderNumber, getQuotaWithUnit, modelToColor } from '../../helpers';
import { USER_COLORS } from '../../constants/user-usage.constants';

export const useUserUsageCharts = (t) => {
  // ========== 模型分布（饼图） ==========
  const [specModelPie, setSpecModelPie] = useState({
    type: 'pie',
    data: [{ id: 'modelPieData', values: [{ type: '无数据', value: 0 }] }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    pie: { style: { cornerRadius: 10 } },
    title: { visible: true, text: '模型调用次数占比', subtext: '' },
    legends: { visible: true, orient: 'left' },
    label: { visible: true },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['type'], value: (d) => renderNumber(d['value']) },
        ],
      },
    },
    color: { specified: {} },
  });

  // ========== 模型用量排行（柱状图） ==========
  const [specModelRank, setSpecModelRank] = useState({
    type: 'bar',
    data: [{ id: 'modelRankData', values: [] }],
    xField: 'Model',
    yField: 'Count',
    seriesField: 'Model',
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '模型调用次数排行', subtext: '' },
    bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['Model'], value: (d) => renderNumber(d['Count']) },
        ],
      },
    },
    color: { specified: {} },
  });

  // ========== 模型消耗分布（堆叠柱状图） ==========
  const [specModelQuota, setSpecModelQuota] = useState({
    type: 'bar',
    data: [{ id: 'modelQuotaData', values: [] }],
    xField: 'Time',
    yField: 'Usage',
    seriesField: 'Model',
    stack: true,
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '模型消耗分布', subtext: '' },
    bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['Model'], value: (d) => renderQuota(d['rawQuota'] || 0, 4) },
        ],
      },
    },
    color: { specified: {} },
  });

  // ========== 用户详情：模型分布（饼图 - 按配额） ==========
  const [specDetailModelQuotaPie, setSpecDetailModelQuotaPie] = useState({
    type: 'pie',
    data: [{ id: 'detailQuotaPie', values: [{ type: '无数据', value: 0 }] }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    pie: { style: { cornerRadius: 10 } },
    title: { visible: true, text: '模型消耗占比', subtext: '' },
    legends: { visible: true, orient: 'left' },
    label: { visible: true },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['type'], value: (d) => renderQuota(d['value'], 4) },
        ],
      },
    },
    color: { specified: {} },
  });

  // ========== 用户详情：时间趋势（折线图） ==========
  const [specDetailTimeTrend, setSpecDetailTimeTrend] = useState({
    type: 'line',
    data: [{ id: 'detailTimeTrend', values: [] }],
    xField: 'Time',
    yField: 'Quota',
    seriesField: 'Metric',
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '用量趋势', subtext: '' },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['Metric'], value: (d) => d['Quota'] },
        ],
      },
    },
    color: { specified: { '消耗额度': '#3b82f6', '调用次数': '#10b981' } },
  });

  // ========== 用户详情：时间趋势（堆叠柱状图 - 按模型） ==========
  const [specDetailModelTime, setSpecDetailModelTime] = useState({
    type: 'bar',
    data: [{ id: 'detailModelTime', values: [] }],
    xField: 'Time',
    yField: 'Usage',
    seriesField: 'Model',
    stack: true,
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '按模型时间分布', subtext: '' },
    bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['Model'], value: (d) => renderQuota(d['rawQuota'] || 0, 4) },
        ],
      },
    },
    color: { specified: {} },
  });

  // ========== 主看板：用户消耗排行 ==========
  const [specUserRank, setSpecUserRank] = useState({
    type: 'bar',
    data: [{ id: 'userRankData', values: [] }],
    xField: 'rawQuota',
    yField: 'User',
    seriesField: 'User',
    direction: 'horizontal',
    legends: { visible: false },
    title: { visible: true, text: '用户消耗排行', subtext: '' },
    bar: { state: { hover: { stroke: '#000', lineWidth: 1 } } },
    label: {
      visible: true,
      position: 'outside',
      formatMethod: (value, datum) => renderQuota(datum['rawQuota'] || 0, 2),
    },
    axes: [
      { orient: 'left', type: 'band', label: { visible: true } },
      { orient: 'bottom', type: 'linear', visible: false },
    ],
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['User'], value: (d) => renderQuota(d['rawQuota'] || 0, 4) },
        ],
      },
    },
    color: { type: 'ordinal', range: USER_COLORS },
  });

  // ========== 主看板：时间趋势 ==========
  const [specGlobalTrend, setSpecGlobalTrend] = useState({
    type: 'area',
    data: [{ id: 'globalTrendData', values: [] }],
    xField: 'Time',
    yField: 'rawQuota',
    seriesField: 'Metric',
    stack: false,
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '全局用量趋势', subtext: '' },
    area: { style: { fillOpacity: 0.15 } },
    line: { style: { lineWidth: 2 } },
    point: { visible: false },
    axes: [{
      orient: 'left',
      label: { formatMethod: (v) => renderQuota(v, 2) },
    }],
    tooltip: {
      mark: {
        content: [
          { key: (d) => d['Metric'], value: (d) => renderQuota(d['rawQuota'] || 0, 4) },
        ],
      },
    },
    color: { specified: { '消耗额度': '#3b82f6', '调用次数': '#10b981' } },
  });

  // ========== 更新主看板图表 ==========
  const updateOverviewCharts = useCallback((overviewData) => {
    // 用户消耗排行
    const MAX_USERS = 10;
    const sorted = [...overviewData].sort((a, b) => b.total_quota - a.total_quota);
    const topUsers = sorted.slice(0, MAX_USERS);

    const userRankValues = topUsers.map((u) => ({
      User: u.username || `用户${u.user_id}`,
      rawQuota: u.total_quota,
      Quota: getQuotaWithUnit(u.total_quota, 4),
    }));

    const totalQuota = topUsers.reduce((s, u) => s + u.total_quota, 0);
    setSpecUserRank((prev) => ({
      ...prev,
      data: [{ id: 'userRankData', values: userRankValues }],
      title: { ...prev.title, subtext: `总计：${renderQuota(totalQuota, 2)}` },
    }));

    // 全局用量趋势（聚合所有用户）
    // 这里简化处理：使用各用户的总用量作为趋势
    const trendValues = overviewData.map((u) => ({
      Time: u.username || `用户${u.user_id}`,
      Metric: '消耗额度',
      rawQuota: u.total_quota,
      Quota: getQuotaWithUnit(u.total_quota, 4),
    }));

    setSpecGlobalTrend((prev) => ({
      ...prev,
      data: [{ id: 'globalTrendData', values: trendValues }],
      title: { ...prev.title, subtext: `总计：${renderQuota(totalQuota, 2)}` },
    }));
  }, []);

  // ========== 更新详情图表 ==========
  const updateDetailCharts = useCallback((detailData) => {
    if (!detailData) return;

    const { model_distribution, time_distribution } = detailData;

    // 模型调用次数饼图
    if (model_distribution && model_distribution.length > 0) {
      const pieValues = model_distribution.map((m) => ({
        type: m.model_name,
        value: m.count,
      }));

      const colors = {};
      model_distribution.forEach((m) => {
        colors[m.model_name] = modelToColor(m.model_name);
      });

      setSpecModelPie((prev) => ({
        ...prev,
        data: [{ id: 'modelPieData', values: pieValues }],
        title: { ...prev.title, subtext: `总计：${renderNumber(pieValues.reduce((s, v) => s + v.value, 0))}` },
        color: { specified: colors },
      }));

      // 模型调用次数排行
      const rankData = model_distribution.map((m) => ({
        Model: m.model_name,
        Count: m.count,
      }));
      setSpecModelRank((prev) => ({
        ...prev,
        data: [{ id: 'modelRankData', values: rankData }],
        color: { specified: colors },
      }));

      // 模型消耗占比饼图
      const quotaPieValues = model_distribution.map((m) => ({
        type: m.model_name,
        value: m.quota,
      }));
      setSpecDetailModelQuotaPie((prev) => ({
        ...prev,
        data: [{ id: 'detailQuotaPie', values: quotaPieValues }],
        title: { ...prev.title, subtext: `总计：${renderQuota(quotaPieValues.reduce((s, v) => s + v.value, 0), 2)}` },
        color: { specified: colors },
      }));
    }

    // 时间趋势
    if (time_distribution && time_distribution.length > 0) {
      // 用量趋势线
      const timeValues = time_distribution.flatMap((t) => {
        const timeLabel = new Date(t.timestamp * 1000).toLocaleDateString('zh-CN');
        return [
          { Time: timeLabel, Metric: '消耗额度', rawQuota: t.quota, Quota: getQuotaWithUnit(t.quota, 4) },
          { Time: timeLabel, Metric: '调用次数', rawQuota: t.count, Quota: t.count },
        ];
      });
      setSpecDetailTimeTrend((prev) => ({
        ...prev,
        data: [{ id: 'detailTimeTrend', values: timeValues }],
      }));

      // 按模型的时间分布（简化为总用量）
      const modelTimeValues = time_distribution.map((t) => ({
        Time: new Date(t.timestamp * 1000).toLocaleDateString('zh-CN'),
        Model: '总用量',
        rawQuota: t.quota,
        Usage: getQuotaWithUnit(t.quota, 4),
      }));
      setSpecDetailModelTime((prev) => ({
        ...prev,
        data: [{ id: 'detailModelTime', values: modelTimeValues }],
      }));
    }
  }, []);

  // ========== 初始化主题 ==========
  useEffect(() => {
    initVChartSemiTheme({ isWatchingThemeSwitch: true });
  }, []);

  return {
    specModelPie,
    specModelRank,
    specModelQuota,
    specDetailModelQuotaPie,
    specDetailTimeTrend,
    specDetailModelTime,
    specUserRank,
    specGlobalTrend,
    updateOverviewCharts,
    updateDetailCharts,
  };
};
