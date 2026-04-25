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
import { renderQuota, renderNumber, modelToColor } from '../../helpers';
import { USER_COLORS } from '../../constants/user-usage.constants';

const emptyPie = [{ type: '无数据', value: 0 }];
const tsLabel = (ts) => new Date(ts * 1000).toLocaleDateString('zh-CN');

export const useUserUsageCharts = () => {
  const [specUserRank, setSpecUserRank] = useState({
    type: 'bar',
    data: [{ id: 'userRankData', values: [] }],
    xField: 'rawQuota',
    yField: 'User',
    seriesField: 'User',
    direction: 'horizontal',
    legends: { visible: false },
    title: { visible: true, text: '用户消耗排行', subtext: '' },
    label: { visible: true, position: 'outside', formatMethod: (v, d) => renderQuota(d.rawQuota || 0, 2) },
    axes: [{ orient: 'left', type: 'band', label: { visible: true } }, { orient: 'bottom', type: 'linear', visible: false }],
    color: { type: 'ordinal', range: USER_COLORS },
  });

  const [specUserTrend, setSpecUserTrend] = useState({
    type: 'line',
    data: [{ id: 'userTrendData', values: [] }],
    xField: 'Time',
    yField: 'rawQuota',
    seriesField: 'User',
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '用户用量趋势', subtext: '' },
    axes: [{ orient: 'left', label: { formatMethod: (v) => renderQuota(v, 2) } }],
    color: { type: 'ordinal', range: USER_COLORS },
  });

  const [specCountRank, setSpecCountRank] = useState({
    type: 'bar',
    data: [{ id: 'countRankData', values: [] }],
    xField: 'User',
    yField: 'Count',
    seriesField: 'User',
    legends: { visible: false },
    title: { visible: true, text: '调用次数排行', subtext: '' },
    color: { type: 'ordinal', range: USER_COLORS },
  });

  const [specErrorUserRank, setSpecErrorUserRank] = useState({
    type: 'bar',
    data: [{ id: 'errorUserRankData', values: [] }],
    xField: 'User',
    yField: 'Errors',
    seriesField: 'User',
    legends: { visible: false },
    title: { visible: true, text: '失败用户排行', subtext: '' },
    color: { type: 'ordinal', range: ['#ef4444', '#f97316', '#f59e0b', '#8b5cf6'] },
  });

  const [specModelPie, setSpecModelPie] = useState({
    type: 'pie',
    data: [{ id: 'modelPieData', values: emptyPie }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    legends: { visible: true, orient: 'left' },
    title: { visible: true, text: '模型调用次数占比', subtext: '' },
    color: { specified: {} },
  });

  const [specDetailModelQuotaPie, setSpecDetailModelQuotaPie] = useState({
    type: 'pie',
    data: [{ id: 'detailQuotaPie', values: emptyPie }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    legends: { visible: true, orient: 'left' },
    title: { visible: true, text: '模型消耗占比', subtext: '' },
    color: { specified: {} },
  });

  const [specModelRank, setSpecModelRank] = useState({
    type: 'bar',
    data: [{ id: 'modelRankData', values: [] }],
    xField: 'Model',
    yField: 'Count',
    seriesField: 'Model',
    legends: { visible: false },
    title: { visible: true, text: '模型调用次数排行', subtext: '' },
    color: { specified: {} },
  });

  const [specDetailTimeTrend, setSpecDetailTimeTrend] = useState({
    type: 'line',
    data: [{ id: 'detailTimeTrend', values: [] }],
    xField: 'Time',
    yField: 'value',
    seriesField: 'Metric',
    legends: { visible: true, selectMode: 'single' },
    title: { visible: true, text: '时间趋势', subtext: '' },
    axes: [{ orient: 'left' }],
    color: { specified: { '消耗额度': '#3b82f6', '调用次数': '#10b981', 'Token 用量': '#8b5cf6' } },
  });

  const [specErrorRank, setSpecErrorRank] = useState({
    type: 'bar',
    data: [{ id: 'errorRankData', values: [] }],
    xField: 'Error',
    yField: 'Count',
    seriesField: 'Error',
    legends: { visible: false },
    title: { visible: true, text: '失败原因排行', subtext: '' },
    color: { type: 'ordinal', range: ['#ef4444', '#f97316', '#f59e0b', '#8b5cf6'] },
  });

  const [specTokenDistribution, setSpecTokenDistribution] = useState({
    type: 'bar',
    data: [{ id: 'tokenDistributionData', values: [] }],
    xField: 'Model',
    yField: 'Tokens',
    seriesField: 'Type',
    stack: true,
    legends: { visible: true },
    title: { visible: true, text: 'Token 分布', subtext: '' },
    color: { specified: { '输入 Token': '#3b82f6', '输出 Token': '#8b5cf6' } },
  });

  const [specLatencyDistribution, setSpecLatencyDistribution] = useState({
    type: 'bar',
    data: [{ id: 'latencyData', values: [] }],
    xField: 'Time',
    yField: 'Latency',
    seriesField: 'Metric',
    legends: { visible: false },
    title: { visible: true, text: '平均耗时分布', subtext: '' },
    color: { specified: { '平均耗时(ms)': '#f59e0b' } },
  });

  const updateOverviewCharts = useCallback((overviewData) => {
    if (!overviewData || overviewData.length === 0) return;

    const sortedByQuota = [...overviewData].sort((a, b) => (b.total_quota || 0) - (a.total_quota || 0));
    const topUsersByQuota = sortedByQuota.slice(0, 10);
    const totalQuota = sortedByQuota.reduce((s, i) => s + (i.total_quota || 0), 0);

    setSpecUserRank((prev) => ({
      ...prev,
      data: [{ id: 'userRankData', values: topUsersByQuota.map((u) => ({ User: u.display_name || u.username || `用户${u.user_id}`, rawQuota: u.total_quota || 0 })) }],
      title: { ...prev.title, subtext: `总计：${renderQuota(totalQuota, 2)}` },
    }));

    const sortedByCount = [...overviewData].sort((a, b) => (b.total_count || 0) - (a.total_count || 0)).slice(0, 10);
    setSpecCountRank((prev) => ({
      ...prev,
      data: [{ id: 'countRankData', values: sortedByCount.map((u) => ({ User: u.display_name || u.username || `用户${u.user_id}`, Count: u.total_count || 0 })) }],
      title: { ...prev.title, subtext: `总计：${renderNumber(overviewData.reduce((s, i) => s + (i.total_count || 0), 0))}` },
    }));

    const sortedByErrors = [...overviewData].filter((u) => (u.error_count || 0) > 0).sort((a, b) => (b.error_count || 0) - (a.error_count || 0)).slice(0, 10);
    setSpecErrorUserRank((prev) => ({
      ...prev,
      data: [{ id: 'errorUserRankData', values: sortedByErrors.map((u) => ({ User: u.display_name || u.username || `用户${u.user_id}`, Errors: u.error_count || 0 })) }],
      title: { ...prev.title, subtext: `总计：${renderNumber(overviewData.reduce((s, i) => s + (i.error_count || 0), 0))}` },
    }));

    const trendMap = new Map();
    topUsersByQuota.forEach((u) => {
      (u.time_series || []).forEach((point) => {
        const key = `${point.timestamp}-${u.username}`;
        trendMap.set(key, {
          Time: tsLabel(point.timestamp),
          User: u.display_name || u.username || `用户${u.user_id}`,
          rawQuota: point.quota || 0,
        });
      });
    });

    setSpecUserTrend((prev) => ({
      ...prev,
      data: [{ id: 'userTrendData', values: Array.from(trendMap.values()) }],
      title: { ...prev.title, subtext: `Top ${topUsersByQuota.length} 用户` },
    }));
  }, []);

  const updateDetailCharts = useCallback((detailData) => {
    if (!detailData) return;
    const modelDistribution = detailData.model_distribution || [];
    const timeDistribution = detailData.time_distribution || [];
    const errorDistribution = detailData.error_distribution || [];

    const colors = {};
    modelDistribution.forEach((m) => {
      colors[m.model_name] = modelToColor(m.model_name || '未知模型');
    });

    setSpecModelPie((prev) => ({
      ...prev,
      data: [{ id: 'modelPieData', values: modelDistribution.length > 0 ? modelDistribution.map((m) => ({ type: m.model_name, value: m.count || 0 })) : emptyPie }],
      title: { ...prev.title, subtext: `总计：${renderNumber(modelDistribution.reduce((s, m) => s + (m.count || 0), 0))}` },
      color: { specified: colors },
    }));

    setSpecDetailModelQuotaPie((prev) => ({
      ...prev,
      data: [{ id: 'detailQuotaPie', values: modelDistribution.length > 0 ? modelDistribution.map((m) => ({ type: m.model_name, value: m.quota || 0 })) : emptyPie }],
      title: { ...prev.title, subtext: `总计：${renderQuota(modelDistribution.reduce((s, m) => s + (m.quota || 0), 0), 2)}` },
      color: { specified: colors },
    }));

    setSpecModelRank((prev) => ({
      ...prev,
      data: [{ id: 'modelRankData', values: modelDistribution.map((m) => ({ Model: m.model_name, Count: m.count || 0 })) }],
      color: { specified: colors },
    }));

    const timeValues = [];
    const latencyValues = [];
    timeDistribution.forEach((t) => {
      const time = tsLabel(t.timestamp);
      timeValues.push({ Time: time, Metric: '消耗额度', value: t.quota || 0 });
      timeValues.push({ Time: time, Metric: '调用次数', value: t.count || 0 });
      timeValues.push({ Time: time, Metric: 'Token 用量', value: t.tokens || 0 });
      latencyValues.push({ Time: time, Metric: '平均耗时(ms)', Latency: t.avg_use_ms || 0 });
    });
    setSpecDetailTimeTrend((prev) => ({ ...prev, data: [{ id: 'detailTimeTrend', values: timeValues }] }));
    setSpecLatencyDistribution((prev) => ({ ...prev, data: [{ id: 'latencyData', values: latencyValues }] }));

    const tokenValues = [];
    modelDistribution.forEach((m) => {
      tokenValues.push({ Model: m.model_name, Type: '输入 Token', Tokens: m.prompt_tokens || 0 });
      tokenValues.push({ Model: m.model_name, Type: '输出 Token', Tokens: m.completion_tokens || 0 });
    });
    setSpecTokenDistribution((prev) => ({ ...prev, data: [{ id: 'tokenDistributionData', values: tokenValues }] }));

    setSpecErrorRank((prev) => ({
      ...prev,
      data: [{ id: 'errorRankData', values: errorDistribution.map((e) => ({ Error: e.error_content || '未知错误', Count: e.count || 0 })) }],
      title: { ...prev.title, subtext: `总计：${renderNumber(errorDistribution.reduce((s, e) => s + (e.count || 0), 0))}` },
    }));
  }, []);

  useEffect(() => {
    initVChartSemiTheme({ isWatchingThemeSwitch: true });
  }, []);

  return {
    specUserRank,
    specUserTrend,
    specCountRank,
    specErrorUserRank,
    specModelPie,
    specDetailModelQuotaPie,
    specModelRank,
    specDetailTimeTrend,
    specErrorRank,
    specTokenDistribution,
    specLatencyDistribution,
    updateOverviewCharts,
    updateDetailCharts,
  };
};
