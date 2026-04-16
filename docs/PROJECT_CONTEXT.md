# Comfy Manager 项目上下文

这份文档给后续维护者和 AI 助手快速建立项目认知使用，不需要每次都重新完整扫描整个仓库。

当前稳定版本：`v1.8.1`
更新时间：`2026-04-16`

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2 + Go + Vue 3** 的桌面图片管理器，主要服务于 ComfyUI 输出目录整理场景。

核心目标：

- 浏览 ComfyUI 输出图片
- 按日期回看近期产出
- 按模型 / LoRA 筛选图片
- 搜索 Prompt、模型、LoRA、标签、笔记等文本信息
- 通过自动规则完成打标、收藏、移动

## 2. v1.8.1 版本重点

### v1.8 主功能

- 新增“日期产出工作台”
- 新增“按模型 / LoRA 筛选”
- 新增日期统计卡片与最近活跃日期入口
- 优化侧边栏结构，保留更高频入口

### v1.8.1 修复与完善

- 修复主页、数据视界、个人中心在 ComfyUI 新出图后未自动刷新的问题
- 改为基于 `images:changed` 事件驱动刷新，去掉前端轮询，降低性能消耗
- 更新内置“使用文档”页面为完整中文说明
- 继续增强模型 / LoRA 筛选稳定性，减少旧筛选值残留导致的误判

## 3. 技术栈

### 后端

- Go
- Wails v2
- fsnotify
- golang.org/x/image
- google/uuid

### 前端

- Vue 3（Composition API）
- Vite
- Tailwind CSS 4
- shadcn-vue
- lucide-vue-next
- vue-sonner

### 通信模型

- 前端通过 `frontend/src/api.js` 调用 `window.go.main.App.*`
- 后端通过 Wails runtime 发送事件：
  - `images:changed`
  - `auto-rules:progress`

## 4. 关键目录

```text
comfy-manager/
├── docs/
│   ├── README.md
│   ├── RELEASE.md
│   ├── PROJECT_CONTEXT.md
│   ├── V1.8_DATE_MODEL_PLAN.md
│   └── V1.8_DATE_MODEL_IMPLEMENTATION.md
├── data/
├── .trash/
├── desktop-app.exe
└── desktop-source/
    ├── app.go
    ├── main.go
    ├── frontend/
    │   ├── src/
    │   │   ├── App.vue
    │   │   ├── api.js
    │   │   ├── composables/
    │   │   │   └── useImages.js
    │   │   ├── lib/
    │   │   │   └── dateWorkbench.js
    │   │   └── components/
    │   │       ├── AppSidebar.vue
    │   │       ├── DateWorkbench.vue
    │   │       ├── Documentation.vue
    │   │       ├── Home.vue
    │   │       ├── ImageGallery.vue
    │   │       ├── ProfileCenter.vue
    │   │       └── StatisticsDashboard.vue
    │   └── wailsjs/
    └── build/
```

## 5. 后端核心结构

主文件：`desktop-source/app.go`

主要职责：

- 扫描图片目录
- 读取并缓存图片 metadata
- 暴露 Wails 接口给前端
- 处理文件移动、删除、恢复
- 维护标签、收藏、规则、笔记、统计数据

关键数据结构：

- `ImageFile`
- `ImageMetadata`
- `ImageMetaCacheEntry`
- `AutoRule`
- `FavoriteGroup`

与 v1.8 相关的主要字段：

- `ImageFile.Model`
- `ImageFile.Loras`
- `ImageMetaCacheEntry.Model`
- `ImageMetaCacheEntry.Loras`

## 6. 前端核心结构

### `frontend/src/App.vue`

根组件，负责：

- 注入全局 `toast` 和 `confirm`
- 管理主视图切换
- 监听 `images:changed`
- 装配侧边栏、日期工作台、图库、文档页、个人中心

### `frontend/src/composables/useImages.js`

最核心的业务状态层，负责：

- 图片、标签、收藏、笔记
- 搜索、排序、分页
- 动态目录树
- 日期工作台筛选状态
- 模型 / LoRA 聚合与归一化
- 最终图库结果过滤

与 v1.8.1 相关的状态：

- `activeDatePreset`
- `activeDateValue`
- `activeModelFilter`
- `activeLoraFilter`
- `availableModels`
- `availableLoras`
- `dateWorkbenchSummary`

v1.8.1 调整：

- `startPolling()` 不再持续轮询，改为一次性初始化抓取
- 自动刷新交由 `images:changed` 事件处理

### `frontend/src/lib/dateWorkbench.js`

日期辅助模块，负责：

- 识别 `YYYY-MM-DD` 日期目录
- 生成日期 key
- 计算预设日期范围
- 构建日期统计 map

### `frontend/src/components/DateWorkbench.vue`

日期工作台页面，负责：

- 显示今日 / 昨日 / 最近 7 天 / 本月统计
- 提供日期、模型、LoRA 快捷筛选
- 提供最近活跃日期入口

### `frontend/src/components/ImageGallery.vue`

图库主视图，负责：

- 搜索框
- 模型 / LoRA 下拉筛选
- 空状态提示
- 批量操作
- 图片网格与分页

### `frontend/src/components/Documentation.vue`

内置使用文档页面，当前为完整中文说明，包含：

- 核心功能介绍
- 快速上手
- 常用快捷操作
- 常见问题

## 7. 数据持久化

常见文件：

- `favorites.json`
- `tags.json`
- `image-tags.json`
- `image-notes.json`
- `custom-roots.json`
- `settings.json`
- `auto-rules.json`
- `trash-metadata.json`
- `image-meta-cache.json`

v1.8.x 注意点：

- 日期工作台本身不新增独立 JSON 文件
- 工作台筛选状态主要由前端 `localStorage` 持久化
- 旧版模型 / LoRA 筛选值会在前端尝试迁移或清理，避免升级后出现“结果被旧值卡住”

## 8. 核心业务链路

### 图片刷新链路

`fsnotify` -> `images:changed` -> 前端页面订阅刷新

当前主动订阅页面：

- `Home.vue`
- `StatisticsDashboard.vue`
- `ProfileCenter.vue`

### 日期工作台链路

`GetImages()` 返回图片与 metadata -> `useImages.js` 识别日期目录 -> `dateWorkbenchSummary` 计算统计 -> `DateWorkbench.vue` 渲染卡片和日期入口

### 模型 / LoRA 筛选链路

后端读取 metadata -> 前端做模型与 LoRA 归一化聚合 -> 用户选择筛选项 -> `finalImages` 与 `workbenchFilteredImages` 重新计算

### 搜索链路

`GetImages()` 生成 `searchText` -> 前端叠加文件名、路径、Prompt、模型、LoRA、标签、笔记等信息 -> 图库即时过滤

## 9. 开发注意事项

- 修改筛选逻辑时，优先检查 `useImages.js`
- 涉及模型 / LoRA 的改动要同时检查：
  - 后端 metadata 提取
  - `ImageFile` 返回字段
  - 前端归一化规则
  - 本地旧筛选值兼容
- 修改日期工作台时，优先检查：
  - `dateWorkbench.js`
  - `DateWorkbench.vue`
  - `App.vue`
- 修改自动刷新时，优先检查：
  - 后端是否正确发送 `images:changed`
  - 页面是否正确订阅和解绑事件
  - 是否误引入了高频轮询

## 10. 发布要求

发布版本时必须同步更新：

- `desktop-app.exe`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

## 11. 最近变更记录

### 2026-04-16 | v1.8.1

- 修复主页、数据视界、个人中心自动刷新缺失
- 去除前端轮询，改为事件驱动刷新
- 更新内置使用文档页为完整中文内容
- 稳定 v1.8 日期工作台与模型 / LoRA 筛选体验

### 2026-04-16 | v1.8.0

- 新增日期产出工作台
- 新增模型 / LoRA 筛选
- 新增日期统计卡片与最近活跃日期入口
- 优化侧边栏入口结构
- 修复模型筛选归一化与旧筛选值兼容问题

### 2026-04-16 | v1.7.0

- 新增搜索 MVP
- 新增自动规则引擎
- 自动规则整合进个人中心
