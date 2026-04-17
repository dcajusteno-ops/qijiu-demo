# Comfy Manager 项目上下文

当前稳定版本：`v2.0.1`  
更新时间：`2026-04-17`

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2 + Go + Vue 3** 的桌面图片管理器，服务于 ComfyUI 出图后的浏览、筛选、整理与归档场景。

项目当前的核心目标：

- 浏览 ComfyUI 输出图片与 PNG 元数据
- 绑定任意 ComfyUI `output` 目录，而不是固定跟随 exe 所在位置
- 按日期、模型、LoRA、标签、收藏、笔记等维度筛选图片
- 提供“日期产出工作台”，快速回看最近产出
- 提供自动规则引擎，自动打标、归类与后处理
- 通过自定义目录、日期归档目录和默认目录统一组织侧边栏结构

## 2. v2.0.1 版本重点

### 2.0 核心升级

- 新增 **任意位置绑定 ComfyUI output 目录**，首次进入强制用户选择真实输出目录
- 新增 **设置中心**，把主题、快捷键、缓存、文件夹维护、工具菜单配置整合到统一浮层
- 新增 **工具菜单自定义**：支持顺序调整、显示/隐藏，设置项固定置顶
- 新增 **日期产出工作台日期范围筛选**，支持开始日期与结束日期
- 新增 **默认目录 / 日期归档目录 / 自定义目录** 并行管理模型
- 自定义目录支持侧边栏显示开关、顺序调整、多层折叠显示

### 本次重点修复

- 修复软件内置“使用文档”页面未更新的问题
- 修复内置文档页中的中文乱码与旧版本内容残留
- 修复首页“查看全部 / 查看更多”跳转到空页面的问题
- 修复日期工作台与图库页之间的筛选联动断裂
- 修复工具菜单与设置中心入口分散的问题
- 修复部分中文乱码、错误提示乱码、自定义目录名称乱码
- 修复 Lightbox 详情中的尺寸文本与按钮文案异常
- 修复全局快捷键文案和配置页的中英文混杂问题

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
- 后端通过 Wails runtime 事件通知前端
  - `images:changed`
  - `shortcut:triggered`
  - `auto-rules:progress`

## 4. 关键目录

```text
comfy-manager/
├─ README.md
├─ docs/
│  ├─ README.md
│  ├─ RELEASE.md
│  └─ PROJECT_CONTEXT.md
├─ data/
├─ .trash/
├─ desktop-app.exe
└─ desktop-source/
   ├─ app.go
   ├─ shortcuts.go
   ├─ main.go
   ├─ frontend/
   │  ├─ src/
   │  │  ├─ App.vue
   │  │  ├─ api.js
   │  │  ├─ composables/
   │  │  │  └─ useImages.js
   │  │  ├─ lib/
   │  │  │  ├─ dateWorkbench.js
   │  │  │  ├─ shortcuts.js
   │  │  │  └─ utilityMenu.js
   │  │  └─ components/
   │  │     ├─ AppSidebar.vue
   │  │     ├─ SettingsCenterDialog.vue
   │  │     ├─ DirectoryBindingDialog.vue
   │  │     ├─ DateWorkbench.vue
   │  │     ├─ Home.vue
   │  │     ├─ ImageGallery.vue
   │  │     ├─ Lightbox.vue
   │  │     ├─ Documentation.vue
   │  │     ├─ ProfileCenter.vue
   │  │     ├─ StatisticsDashboard.vue
   │  │     └─ AutoRulesPanel.vue
   │  └─ wailsjs/
   └─ build/
```

## 5. 后端核心结构

主文件：`desktop-source/app.go`

主要职责：

- 扫描图片目录
- 解析 PNG 元数据、模型、LoRA、工作流节点数等
- 维护图片元数据缓存
- 维护目录绑定、自定义目录、收藏夹、标签、笔记、规则
- 提供图片删除、恢复、清理缓存、清理空目录、日期整理等能力

关键数据结构：

- `ImageFile`
- `ImageMetadata`
- `ImageMetaCacheEntry`
- `AutoRule`
- `FavoriteGroup`
- `CustomRoot`
- `DirectoryBinding`
- `UtilityMenuState`
- `ShortcutSettings`

## 6. 前端核心结构

### `frontend/src/App.vue`

根组件，负责：

- 注入全局 `toast` 与 `confirm`
- 统一装配侧边栏、主页、图库、工作台、设置相关页面
- 处理根级视图跳转
- 监听 `images:changed` 和 `shortcut:triggered`
- 协调日期工作台与图库筛选状态

### `frontend/src/composables/useImages.js`

核心业务状态层，负责：

- 图片、收藏、标签、笔记、自定义目录状态
- 搜索、分页、排序、堆叠显示
- 日期工作台的日期范围、模型、LoRA 筛选
- 侧边栏目录树生成
- 当前图库结果集计算

v2.0.1 关键状态：

- `activeDatePreset`
- `activeDateStart`
- `activeDateEnd`
- `activeModelFilter`
- `activeLoraFilter`
- `availableModels`
- `availableLoras`
- `dateWorkbenchSummary`

### `frontend/src/components/AppSidebar.vue`

负责：

- 主导航与目录树显示
- 标签区与标签筛选
- 工具菜单浮层
- 设置中心、自定义目录、目录绑定等入口

v2.0.1 调整：

- 标签区新增高度限制与独立滚动
- 工具菜单从静态列表改为配置化渲染
- “按日期整理文件”移入设置中心

### `frontend/src/components/SettingsCenterDialog.vue`

负责：

- 外观模式
- 收藏分组
- 快捷键设置
- 缓存清理
- 文件夹维护
- 工具菜单顺序与显示配置

### `frontend/src/components/DateWorkbench.vue`

负责：

- 日期产出统计卡片
- 日期范围选择
- 模型 / LoRA 快捷筛选
- 最近活跃日期入口
- 一键跳回图库继续浏览

## 7. 数据持久化

常见数据文件：

- `favorites.json`
- `tags.json`
- `image-tags.json`
- `image-notes.json`
- `custom-roots.json`
- `settings.json`
- `auto-rules.json`
- `trash-metadata.json`
- `image-meta-cache.json`

v2.0.1 重点新增 / 调整：

- `settings.json` 中新增工具菜单配置
- `custom-roots.json` 中维护内置日期归档目录和自定义目录顺序
- 目录绑定信息不再依赖 exe 上级目录推断，而是显式配置

## 8. 核心业务链路

### 图片刷新链路

`fsnotify` -> `images:changed` -> 前端页面订阅刷新

当前主要依赖该事件的页面：

- `Home.vue`
- `ProfileCenter.vue`
- `StatisticsDashboard.vue`
- 主图库页

### 日期工作台链路

`GetImages()` -> `useImages.js` 提取日期目录 -> `dateWorkbenchSummary` 统计 -> `DateWorkbench.vue` 展示 -> 一键跳回图库

### 模型 / LoRA 筛选链路

后端解析 metadata -> 前端归一化聚合模型与 LoRA -> 工作台或图库选择筛选 -> 结果重新计算

### 自定义目录链路

目录绑定确定根路径 -> `custom-roots.json` 提供自定义目录定义 -> `useImages.js` 构建侧边栏树 -> 用户切换目录查看

## 9. 开发注意事项

- 涉及筛选逻辑，优先检查 `useImages.js`
- 涉及日期产出工作台，优先检查：
  - `DateWorkbench.vue`
  - `dateWorkbench.js`
  - `App.vue`
- 涉及工具菜单显示，优先检查：
  - `AppSidebar.vue`
  - `SettingsCenterDialog.vue`
  - `utilityMenu.js`
- 涉及中文乱码，优先检查：
  - Go 源文件字符串字面量
  - 前端新增组件编码
  - 历史数据文件中的目录名与显示名回退逻辑

## 10. 发布要求

发布版本时至少同步以下内容：

- 根目录 `desktop-app.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

## 11. 最近变更记录

### 2026-04-17 | v2.0.1

- 更新软件内置“使用文档”页面
- 修复内置文档页中文乱码和旧版本说明残留
- 同步更新 GitHub README 与 docs 文档

### 2026-04-17 | v2.0.0

- 新增任意位置绑定 ComfyUI `output` 目录
- 新增设置中心，统一承载工具与系统设置
- 新增工具菜单配置化能力
- 新增日期工作台日期范围筛选
- 优化默认目录、日期归档目录、自定义目录的侧边栏结构
- 修复首页跳转空页、部分筛选联动异常和多处中文乱码

### 2026-04-16 | v1.8.1

- 修复自动刷新问题
- 去除前端轮询，改为事件驱动刷新
- 更新中文文档页面

### 2026-04-16 | v1.8.0

- 新增日期产出工作台
- 新增模型 / LoRA 筛选

### 2026-04-16 | v1.7.0

- 新增搜索 MVP
- 新增自动规则引擎
