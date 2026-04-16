# Comfy Manager 项目上下文

这份文档给后续维护者和 AI 助手快速建立上下文使用，不需要每次重新扫完整个仓库。

当前稳定版本：`v1.8.0`

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2（Go + Vue 3）** 的桌面图片管理器，主要服务于 ComfyUI 输出目录整理场景。

当前核心链路：

- 浏览 ComfyUI 输出图片
- 按日期回看近期产出
- 按模型 / LoRA 过滤图片
- 搜索 Prompt、模型、LoRA、标签、笔记等文本信息
- 通过自动规则完成打标、收藏、移动

## 2. v1.8.0 版本重点

本版本新增：

- 日期产出工作台
- 模型 / LoRA 筛选
- 日期统计卡片和最近活跃日期入口
- 模型名归一化筛选逻辑

本版本调整：

- 侧边栏顶部入口做减法，只保留更高频的工作台入口
- 数据视界收进工具菜单
- 图库头部布局优化，避免标题被筛选区挤压

本版本修复：

- 旧版模型 / LoRA 筛选值残留导致“界面看似默认、结果却被过滤”
- 模型名因路径、扩展名、大小写差异导致匹配不稳定

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
├─ docs/
│  ├─ README.md
│  ├─ RELEASE.md
│  ├─ PROJECT_CONTEXT.md
│  ├─ V1.8_DATE_MODEL_PLAN.md
│  └─ V1.8_DATE_MODEL_IMPLEMENTATION.md
├─ data/
├─ .trash/
├─ desktop-app.exe
└─ desktop-source/
   ├─ app.go
   ├─ main.go
   └─ frontend/
      ├─ src/
      │  ├─ App.vue
      │  ├─ api.js
      │  ├─ composables/
      │  │  └─ useImages.js
      │  ├─ lib/
      │  │  └─ dateWorkbench.js
      │  └─ components/
      │     ├─ AppSidebar.vue
      │     ├─ DateWorkbench.vue
      │     ├─ ImageGallery.vue
      │     ├─ ProfileCenter.vue
      │     └─ AutoRulesPanel.vue
      └─ wailsjs/
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

v1.8.0 相关字段：

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
- 装配侧边栏、日期工作台、个人中心、图库和文档页

### `frontend/src/composables/useImages.js`

这是最核心的业务状态层，负责：

- 图片、标签、收藏、笔记
- 搜索、排序、分页
- 动态目录树
- 日期工作台筛选状态
- 模型 / LoRA 聚合与归一化
- 最终图库结果过滤

v1.8.0 相关状态：

- `activeDatePreset`
- `activeDateValue`
- `activeModelFilter`
- `activeLoraFilter`
- `availableModels`
- `availableLoras`
- `dateWorkbenchSummary`

### `frontend/src/lib/dateWorkbench.js`

新增的日期辅助模块，负责：

- 识别 `YYYY-MM-DD` 日期目录
- 生成日期 key
- 计算预设日期范围
- 构建日期统计 map

### `frontend/src/components/DateWorkbench.vue`

新增的日期工作台页面，负责：

- 显示今日 / 昨日 / 最近7天 / 本月统计
- 提供日期、模型、LoRA 快捷筛选
- 提供最近活跃日期入口

### `frontend/src/components/ImageGallery.vue`

图库主视图，负责：

- 搜索框
- 模型 / LoRA 下拉筛选
- 空状态提示
- 批量操作
- 图片网格与分页

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

v1.8.0 相关注意点：

- 日期工作台本身不新增单独 JSON 文件
- 工作台筛选状态主要由前端 `localStorage` 持久化
- 升级到 v1.8.0 后，旧版模型 / LoRA 筛选值会在前端自动迁移或清空

## 8. 核心业务链路

### 图片刷新链路

`fsnotify` -> `images:changed` -> `App.vue handleRefresh()` -> `fetchImages / fetchTags / fetchImageTags`

### 日期工作台链路

`GetImages()` 返回图片与 metadata -> `useImages.js` 识别日期目录 -> `dateWorkbenchSummary` 计算统计 -> `DateWorkbench.vue` 渲染卡片和日期入口

### 模型 / LoRA 筛选链路

后端读取 metadata -> 前端对模型和 LoRA 做归一化聚合 -> 用户选择筛选项 -> `finalImages` 和 `workbenchFilteredImages` 重新计算

### 搜索链路

`GetImages()` 生成 `searchText` -> 前端叠加文件名、路径、Prompt、模型、LoRA、标签、笔记等信息 -> 图库即时过滤

## 9. 开发注意事项

- 修改筛选逻辑时，优先检查 `useImages.js`
- 模型 / LoRA 相关改动要同时检查：
  - 后端 metadata 提取
  - `ImageFile` 返回字段
  - 前端归一化规则
  - 本地旧筛选值兼容
- 修改日期工作台时，优先检查：
  - `dateWorkbench.js`
  - `DateWorkbench.vue`
  - `App.vue`
- 发布版本时必须同步更新：
  - `desktop-app.exe`
  - `docs/README.md`
  - `docs/RELEASE.md`
  - `docs/PROJECT_CONTEXT.md`

## 10. 最近变更记录

### 2026-04-16 | v1.8.0

- 新增日期产出工作台
- 新增模型 / LoRA 筛选
- 新增日期统计卡片与最近活跃日期
- 优化侧边栏入口结构
- 修复模型筛选归一化与旧筛选值兼容问题

### 2026-04-16 | v1.7.0

- 新增搜索 MVP
- 新增自动规则引擎
- 自动规则整合进个人中心

Updated for v1.8.0 on 2026-04-16
