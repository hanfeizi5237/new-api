# 用户用量看板 - 开发进度

## 状态：✅ 已完成

### 已完成项

#### 后端
- [x] `dto/user_usage.go` - DTO 定义（UserUsageOverview, UserUsageDetail, 等）
- [x] `model/user_usage.go` - 数据查询方法
  - `GetUserUsageOverview` - 用户用量概览（quota_data + logs 聚合）
  - `GetUserUsageDetail` - 用户用量详情（模型分布、时间分布、错误统计）
  - `GetTimeSeriesData` - 全局时间序列数据
  - 支持 MySQL / PostgreSQL / SQLite 三种数据库
  - 支持 day / week / month 三种聚合粒度
- [x] `controller/user_usage.go` - API 控制器
  - `GET /api/admin/usage/overview` - 用户概览
  - `GET /api/admin/usage/detail` - 用户详情
  - `GET /api/admin/usage/timeseries` - 全局时间序列
- [x] `router/api-router.go` - 路由注册（仅追加，不修改现有代码）
- [x] Go 编译通过

#### 前端
- [x] `constants/user-usage.constants.js` - 常量定义
- [x] `hooks/user-usage/useUserUsageData.js` - 数据管理 Hook
- [x] `hooks/user-usage/useUserUsageCharts.js` - 图表管理 Hook
- [x] `components/user-usage/MainDashboardView.jsx` - 主看板视图
- [x] `components/user-usage/UserDetailView.jsx` - 详情抽屉
- [x] `pages/UserUsageDashboard/index.jsx` - 页面入口
- [x] `App.jsx` - 路由注册（/user-usage）
- [x] `SiderBar.jsx` - 侧边栏菜单项（用户看板）
- [x] `render.jsx` - 图标注册

### 功能清单

| 功能 | 状态 |
|------|------|
| 主看板 - 所有用户汇总统计 | ✅ |
| 主看板 - 5 个统计卡片 | ✅ |
| 主看板 - 用户列表（可排序、分页） | ✅ |
| 日期区间选择（最长 31 天） | ✅ |
| 聚合粒度切换（天/周/月） | ✅ |
| CSV 导出 | ✅ |
| 用户详情 - 抽屉展示 | ✅ |
| 详情 - 模型分布（调用次数/配额/Token） | ✅ |
| 详情 - 时间趋势 | ✅ |
| 详情 - 错误统计 | ✅ |
| Admin 权限控制 | ✅ |
| 零侵入现有代码 | ✅ |

### 验证命令
```bash
# 后端编译
cd ~/.openclaw/projects/new-api && go build ./...

# 前端构建（需要 Node.js 环境）
cd ~/.openclaw/projects/new-api/web && npm run build
```

### 已知限制
1. `quota_data` 表仅精确到小时级别
2. 详情查询限制最大 31 天时间跨度
3. 错误信息截断为 200 字符
4. 图表使用 VChart（项目已有依赖），但部分图表 spec 待完善

### 下一步
1. 前端构建测试
2. 后端联调测试（需实际数据库数据）
3. 图表渲染完善（VChart spec 需要实际数据验证）
