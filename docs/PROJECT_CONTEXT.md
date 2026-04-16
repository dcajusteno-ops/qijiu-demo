# Comfy Manager - 开发者上下文与项目总览

> 这份文档给后续维护者和 AI 编辑器快速建立上下文用，不需要每次全量扫仓库。

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2（Go + Vue 3）** 的桌面图像管理器，面向 ComfyUI 输出图片场景。

当前版本：`v1.7.0`

核心能力：

- 图片扫描、排序、分页展示
- PNG 元数据解析（ComfyUI / WebUI）
- 搜索 MVP：文件名、路径、Prompt、模型、LoRA、采样器、Negative、尺寸、标签、笔记
- 自动规则引擎：自动打标、加入收藏分组、移动目录
- 个人中心：资料编辑、默认进入页、展示图、规则控制入口
- 数据视界：生成历史时间线、趋势图、活跃热图、节点回溯
- 标签与收藏分组管理
- 回收站（软删除、恢复、清空、保留期）
- 自定义根目录（多目录纳管）
- 全局快捷键
- 外部工具启动器与提示词工具链接
- 缩略图 / 预览缓存管理

明确说明：

- `v1.7.0` 已删除“智能相册”模块，相关筛选能力改由“搜索 MVP + 自动规则引擎”承接

## 2. 技术栈与运行模型

### 2.1 后端

- 语言 / 框架：Go + Wails v2
- 关键依赖：
  - `fsnotify`：文件系统监听
  - `golang.org/x/image`：图像解码支持
  - `google/uuid`：ID 生成

后端职责：

- 文件扫描、移动、删除、恢复
- PNG 文本块解析与元数据建模
- JSON 数据持久化
- 目录监听并向前端发事件
- 通过自定义 `AssetServer` 读取磁盘图片
- 执行自动规则并发出进度事件

### 2.2 前端

- 框架：Vue 3（Composition API）
- 构建：Vite
- UI：Tailwind CSS 4 + shadcn-vue + lucide-vue-next + vue-sonner
- 状态组织：无 Vuex / Pinia，核心集中在 `useImages.js`

### 2.3 通信模型

- 前端通过 `frontend/src/api.js` 调用 `window.go.main.App.*`
- 后端通过 Wails runtime 推送事件：
  - `images:changed`
  - `auto-rules:progress`
- 前端在 `App.vue` 中监听刷新事件，在 `AutoRulesPanel.vue` 中监听规则进度事件

## 3. 目录结构

```text
comfy-manager/
├── docs/
│   ├── README.md
│   ├── RELEASE.md
│   ├── PROJECT_CONTEXT.md
│   └── V1.7_SEARCH_AND_RULES_PLAN.md
├── data/
├── .trash/
├── desktop-app.exe
└── desktop-source/
    ├── app.go
    ├── main.go
    ├── go.mod
    └── frontend/
        ├── src/
        │   ├── api.js
        │   ├── App.vue
        │   ├── composables/
        │   │   └── useImages.js
        │   └── components/
        └── wailsjs/
```

## 4. 后端核心架构

主文件：`desktop-source/app.go`

### 4.1 App 生命周期与目录绑定

`NewApp()` 会推导运行目录并确定：

- `appDir`：应用根目录
- `imageDir`：默认受管图片目录
- `rootDir`：用于相对路径安全解析的根目录
- `dataDir`：JSON 数据目录

### 4.2 路径安全

关键函数：

- `normalizeRelPath`
- `resolveRootPath`
- `isSubPath`
- `normalizeDir`

规则：

- 所有文件操作优先使用 `relPath`
- 所有磁盘访问必须经过后端路径校验

### 4.3 图片扫描与元数据缓存

主入口：`GetImages(sortBy, sortOrder)`

流程：

1. 遍历 `managedImageRoots()`
2. 仅处理图片后缀（png / jpg / jpeg / webp / gif）
3. 尝试命中 `image-meta-cache.json`
4. 需要时读取尺寸或补扫元数据
5. 生成 `ImageFile[]`
6. 同步填充 `searchText`

缓存字段 `ImageMetaCacheEntry` 现在包含：

- 尺寸、基础文件信息
- `positive / negative / model / sampler / loras`
- `searchText`

### 4.4 PNG 元数据解析

主入口：`GetImageMetadata(relPath)`

- 非 PNG 返回基础信息
- PNG 会解析文本块并构建统一 `ImageMetadata`
- 解析结果会回写缓存，供搜索与自动规则使用

### 4.5 文件监听与前端刷新

- `restartImageWatcher()` 递归监听所有纳管目录
- `scheduleImagesChangedEvent()` 防抖 350ms 发 `images:changed`
- 前端收到后统一执行 `handleRefresh()`

### 4.6 自动规则引擎

核心结构：

- `AutoRuleCondition`
- `AutoRuleAction`
- `AutoRule`
- `AutoRulesStore`
- `AutoRulesRunSummary`
- `AutoRulesRunProgress`

核心方法：

- `GetAutoRules`
- `SetAutoRulesEnabled`
- `CreateAutoRule`
- `UpdateAutoRule`
- `DeleteAutoRule`
- `RunAutoRulesNow`
- `runAutoRulesForPaths`
- `executeAutoRuleActions`

支持条件字段：

- `model`
- `sampler`
- `lora`
- `dimensions`
- `filename`
- `prompt`
- `negative`

支持动作：

- `add_tag`
- `add_favorite_group`
- `move_to_folder`

### 4.7 个人中心展示图

展示图不再依赖普通图库 `rootDir` 解析，当前通过独立资源前缀处理：

- `__profile__/...`

这样即使展示图落在 `data/profile/` 目录，也能正常显示。

## 5. 前端架构

### 5.1 根组件编排（`frontend/src/App.vue`）

职责：

- 注入全局 `confirm` / `toast`
- 组装侧栏、首页、文档页、画廊页、个人中心页
- 监听 `images:changed` 事件统一刷新
- 管理全局选择模式与批量删除

说明：

- 未使用 Vue Router
- 通过 `activeRoot / activeSub / activeChild` 实现多视图切换

### 5.2 状态中枢（`frontend/src/composables/useImages.js`）

这是最核心的前端业务层，管理：

- 图片、标签、图片-标签映射
- 收藏与收藏分组
- 自定义根目录
- 图片笔记
- 搜索、排序、分页
- 动态 `fileTree` 构建

`v1.7.0` 关键点：

- 搜索逻辑已接入 `img.searchText`
- 搜索覆盖 LoRA / 采样器 / Negative / 尺寸等元数据

### 5.3 前端 API 桥（`frontend/src/api.js`）

- 统一 `callApp(method, ...args)`
- 对外暴露完整 Wails 调用
- 已包含自动规则与个人中心相关接口

### 5.4 核心组件职责

- `AppSidebar.vue`：主导航、标签入口、常用操作、自定义根目录入口
- `ImageGallery.vue`：画廊主体、搜索框、卡片网格、选择模式、分页、批处理入口
- `ImageCard.vue`：单图卡片
- `Lightbox.vue`：大图浏览、元数据详情、标签与收藏管理、笔记编辑
- `AutoRulesPanel.vue`：自动规则列表、执行进度、搜索、过滤、立即执行
- `AutoRulesDialog.vue`：规则创建 / 编辑弹窗
- `ProfileCenter.vue`：资料编辑、默认进入页、展示图、自动规则入口
- `StatisticsDashboard.vue` / `Home.vue`：首页与数据视界

## 6. 数据持久化

常见文件：

- `favorites.json`：收藏分组与路径
- `tags.json`：标签定义
- `image-tags.json`：图片-标签映射
- `image-notes.json`：图片笔记
- `custom-roots.json`：自定义根目录
- `settings.json`：设置、快捷键、个人中心资料
- `auto-rules.json`：自动规则配置
- `trash-metadata.json`：回收站原路径、删除时间
- `image-meta-cache.json`：图片元数据缓存

## 7. 核心业务链路速查

### 7.1 图片刷新链路

`fsnotify` 事件 -> `scheduleImagesChangedEvent()` -> Wails `images:changed` -> `App.vue handleRefresh()` -> `fetchImages / fetchTags / fetchImageTags`

### 7.2 搜索链路

`GetImages()` 生成 `searchText` -> 前端 `useImages.js` 汇总文件名 / Prompt / 模型 / `searchText` / 标签 / 笔记 -> `ImageGallery.vue` 输入框触发过滤

### 7.3 自动规则链路

导入图片或手动执行 -> `runAutoRulesForPaths()` -> 条件匹配 -> 动作执行 -> 发出 `auto-rules:progress` -> 有更新时发出 `images:changed`

### 7.4 标签链路

创建 / 更新 / 删除标签（`tags.json`） + 图片打标（`image-tags.json`） + 前端统计与筛选联动

## 8. 开发注意事项

### 8.1 修改前端状态逻辑

- 优先改 `useImages.js`
- 不要把同一份筛选逻辑复制到多个组件

### 8.2 修改自动规则

- 规则条件、动作、默认规则都在 `app.go`
- 改规则时同时检查：
  - `auto-rules.json` 兼容性
  - 前端表单字段
  - `auto-rules:progress` 事件结构

### 8.3 修改个人中心展示图

- 展示图资源路径走 `__profile__/...`
- 不要再假设展示图一定在图库根目录内

## 9. 快速定位索引

### 后端

- 核心业务：`desktop-source/app.go`
- 启动入口：`desktop-source/main.go`

### 前端

- 根编排：`desktop-source/frontend/src/App.vue`
- 状态中枢：`desktop-source/frontend/src/composables/useImages.js`
- API 桥：`desktop-source/frontend/src/api.js`
- 侧栏：`desktop-source/frontend/src/components/AppSidebar.vue`
- 画廊：`desktop-source/frontend/src/components/ImageGallery.vue`
- 自动规则：`desktop-source/frontend/src/components/AutoRulesPanel.vue`
- 自动规则弹窗：`desktop-source/frontend/src/components/AutoRulesDialog.vue`
- 个人中心：`desktop-source/frontend/src/components/ProfileCenter.vue`

## 10. 变更记录

- **2026-04-16 | v1.7.0 - 搜索 MVP 与自动规则引擎**
  - 影响范围：后端 / 前端 / 数据结构 / 事件 / 发布产物 / 文档
  - 变更重点：
    - 新增搜索 MVP，前端正式接入 `searchText`
    - 新增自动规则引擎、默认规则、立即执行进度反馈
    - 自动规则控制收敛进个人中心页面
    - 删除智能相册模块与相关接口
    - 修复个人中心展示图资源路径问题
  - 关键文件：
    - `desktop-source/app.go`
    - `desktop-source/frontend/src/composables/useImages.js`
    - `desktop-source/frontend/src/components/AutoRulesPanel.vue`
    - `desktop-source/frontend/src/components/AutoRulesDialog.vue`
    - `desktop-source/frontend/src/components/ProfileCenter.vue`
    - `docs/README.md`
    - `docs/RELEASE.md`
    - `docs/PROJECT_CONTEXT.md`
    - `desktop-app.exe`

- **2026-04-15 | v1.6.0 - 个人中心与数据视界细化**
  - 新增个人中心页面与资料编辑
  - 优化数据视界节点详情与布局

- **2026-04-15 | v1.5.0 - 数据视界与全局快捷键**
  - 新增数据视界
  - 新增系统级快捷键设置与动作

*Updated for v1.7.0 on 2026-04-16*
