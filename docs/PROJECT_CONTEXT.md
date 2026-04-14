# Comfy Manager - 开发者上下文与项目总览（可持续维护版）

> **致 AI 编辑器/开发者：**
> 本文档用于让 AI 在不全量扫描仓库的情况下快速理解 Comfy Manager 的架构、关键数据流、核心模块与维护方式。
>
> **使用方式建议：**
> 1) 先读本文件，再按“快速定位索引”跳转到目标代码；
> 2) 改动完成后，仅更新受影响章节与“变更记录模板”条目；
> 3) 避免每次重写整篇，保持文档长期可演进。

---

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2（Go + Vue 3）** 的桌面图像管理器，面向 ComfyUI 输出图片场景。

核心能力：
- 图片扫描、排序、分页展示
- PNG 元数据解析（ComfyUI / WebUI）
- 标签与收藏分组管理
- 回收站（软删除 / 恢复 / 清空 / 保留期）
- 自定义根目录（多目录纳管）
- 智能相册（模型 / 采样器 / LoRA / 尺寸）
- 外部工具启动器与提示词工具链接
- 缩略图/预览缓存管理

---

## 2. 技术栈与运行模型

### 2.1 后端
- **语言/框架**：Go + Wails v2
- **关键依赖**：
  - `fsnotify`：文件系统监听
  - `golang.org/x/image`：图像解码支持（含 webp）
  - `google/uuid`：ID 生成
- **后端职责**：
  - 文件扫描、移动、删除、恢复
  - PNG 文本块解析与元数据建模
  - JSON 数据持久化
  - 目录监听并向前端发事件
  - 通过自定义 `AssetServer` 读取磁盘图片

### 2.2 前端
- **框架**：Vue 3（Composition API）
- **构建**：Vite
- **UI**：Tailwind CSS 4 + shadcn-vue + lucide-vue-next + vue-sonner
- **状态组织**：无 Vuex/Pinia，核心集中在 `useImages.js`

### 2.3 通信模型
- 前端通过 `frontend/src/api.js` 调用 `window.go.main.App.*`
- 后端通过 Wails runtime 事件推送 `images:changed`
- 前端 `App.vue` 监听事件并刷新图片/标签关联

---

## 3. 目录结构（高价值部分）

```text
comfy-manager/
├── docs/
│   └── PROJECT_CONTEXT.md          # 本文档
├── data/                           # 运行期数据（持久化 JSON）
├── .trash/                         # 回收站目录（应用根下）
├── desktop-app.exe                 # 发布产物
└── desktop-source/
    ├── app.go                      # 后端核心业务（绝大多数接口）
    ├── main.go                     # Wails 启动与 AssetServer 配置
    ├── go.mod
    └── frontend/
        ├── src/
        │   ├── api.js              # JS API 桥
        │   ├── App.vue             # 根组件与跨模块编排
        │   ├── composables/
        │   │   ├── useImages.js    # 前端状态中枢
        │   │   └── useImageStacks.js
        │   ├── components/         # 业务组件
        │   ├── style.css           # 主题变量/Tailwind tokens
        │   └── theme.js            # 明暗切换与 ViewTransition
        └── wailsjs/                # 自动生成绑定（勿手改）
```

---

## 4. 后端核心架构（`desktop-source/app.go`）

### 4.1 App 生命周期与目录绑定

`NewApp()` 会推导运行目录并确定：
- `appDir`：应用根（comfy-manager）
- `imageDir`：默认受管图片目录（通常上级 output）
- `rootDir`：用于相对路径安全解析的根目录
- `dataDir`：JSON 数据目录

关键点：
- 支持旧/新发布目录结构
- 支持开发模式（临时目录启动时回退 `Getwd`）
- 启动时尝试迁移旧路径数据与旧回收站
- `startup()` 中启动 watcher、清理过期回收站、清理孤儿标签

### 4.2 路径安全与规范化

关键函数：
- `normalizeRelPath`
- `resolveRootPath`
- `isSubPath`
- `normalizeDir`

设计目标：
- 所有对外文件操作都基于 `relPath`
- 强制路径在 `rootDir` 范围内，避免越界访问

### 4.3 图片扫描 + 元数据缓存

主入口：`GetImages(sortBy, sortOrder)`

流程：
1. 遍历 `managedImageRoots()`（主目录 + custom roots）
2. 仅处理图片后缀（png/jpg/jpeg/webp/gif）
3. 尝试命中 `image-meta-cache.json`
4. 需要时读取尺寸、异步 warmup
5. 返回前端 `ImageFile[]`

缓存字段（`ImageMetaCacheEntry`）包含：
- 尺寸、基础文件信息
- 是否已扫描元数据
- positive/negative/model/sampler/loras/searchText

### 4.4 PNG 元数据解析

核心：`GetImageMetadata(relPath)`
- 非 png 仅返回基础信息
- png 读取 tEXt/zTXt/iTXt 等文本块后，构建统一 `ImageMetadata`
- 用于 Lightbox 详情展示与智能相册统计

### 4.5 文件监听与前端刷新

- `restartImageWatcher()` 递归监听所有纳管目录
- 忽略目录：`node_modules`、`.git`、`.trash`、应用自身目录
- `scheduleImagesChangedEvent()` 防抖 350ms 发 `images:changed`
- 前端收到后统一 `handleRefresh`

### 4.6 业务能力分组（后端接口）

1) **目录与绑定**
- `GetDirectoryBinding` / `SaveDirectoryBinding`
- `GetCustomRoots` / `AddCustomRoot` / `UpdateCustomRoot` / `DeleteCustomRoot`
- `GetRelativePath` / `GetSubFolders`

2) **图片与文件操作**
- `GetImages`
- `DeleteImage`（移动到 `.trash`）
- `BatchMove`
- `ExportImages` / `UploadImages`
- `OpenImageLocation` / `OpenFile`

3) **标签系统**
- `GetTags` / `CreateTag` / `UpdateTag` / `DeleteTag` / `BatchDeleteTags`
- `GetImageTags` / `AddTagToImage` / `RemoveTagFromImage`
- `BatchAddTag` / `BatchRemoveTag`
- `CleanupTags`

4) **收藏系统（分组化）**
- `GetFavoriteGroups` / `CreateFavoriteGroup` / `UpdateFavoriteGroup` / `DeleteFavoriteGroup`
- `SetImageFavoriteGroups` / `AddImageToFavoriteGroup` / `RemoveImageFromFavoriteGroup`
- 兼容接口：`GetFavorites` / `AddFavorite` / `RemoveFavorite`

5) **回收站**
- `GetTrashList`
- `RestoreTrash` / `BatchRestoreTrash`
- `BatchDeleteTrash` / `EmptyTrash`
- `GetTrashSettings` / `SaveTrashSettings`
- 启动时 `cleanExpiredTrash`

6) **扩展能力**
- 图片笔记：`GetImageNotes` / `SetImageNote` / `DeleteImageNote`
- 启动器：`GetLauncherTools` / `AddLauncherTool` / `UpdateLauncherTool` / `DeleteLauncherTool` / `RunLauncherTool` / `ExtractIcon`
- 提示词工具链接：`GetPromptToolLinks` / `AddPromptToolLink` / `UpdatePromptToolLink` / `DeletePromptToolLink`
- 统计：`GetStatistics`
- 智能相册：`GetSmartAlbumFields` / `GetSmartAlbums`
- 缓存管理：`ClearPreviewCache`
- 系统能力：`CopyText`

---

## 5. 前端架构与关键数据流

### 5.1 根组件编排（`frontend/src/App.vue`）

职责：
- 建立全局 confirm/toast 注入
- 组装侧栏、主页、文档页、画廊页
- 管理全局选择模式与批量删除
- 监听 Wails 事件并执行刷新
- 承接智能相册过滤状态

导航特点：
- **未使用 Vue Router**
- 通过 `activeRoot / activeSub / activeChild` 实现多视图切换

### 5.2 状态中枢（`frontend/src/composables/useImages.js`）

这是最核心的前端业务层，管理：
- 图片、标签、图片-标签映射
- 收藏与收藏分组
- 自定义根目录
- 图片笔记
- 筛选、排序、分页、堆叠
- 动态 `fileTree` 构建

重要 computed：
- `fileTree`：按业务规则构建侧边树（output/日期归档/OXYZ/修复/收藏夹/custom roots）
- `currentImages`：根据当前导航节点收集图片
- `finalImages`：叠加标签与高级筛选
- `stackedImages`：按 prompt+model 或文件名启发式聚合
- `paginatedImages`

### 5.3 前端 API 桥（`frontend/src/api.js`）

- 提供统一 `callApp(method, ...args)`
- 抛出“非桌面环境”与“接口缺失”错误
- 暴露完整后端方法集合，供 composable/组件调用

### 5.4 核心组件职责

- `AppSidebar.vue`：主导航、标签入口、回收站/启动器/自定义根目录入口、智能相册入口
- `ImageGallery.vue`：画廊主体、卡片网格、选择模式、分页、批处理入口
- `ImageCard.vue`：单图卡片，含 3D 倾斜、收藏/标签/笔记状态
- `Lightbox.vue`：大图浏览、元数据详情、标签与收藏管理、笔记编辑
- `BatchActionsPanel.vue`：批量加标签/移除标签/收藏等
- `FilterPanel.vue`：日期/大小/尺寸过滤 + 堆叠开关
- `TrashDialog.vue`：回收站管理与设置
- `LauncherDialog.vue`：外部工具配置与运行
- `StatisticsDashboard.vue` / `Home.vue`：主页与统计展示

---

## 6. 数据持久化说明（`comfy-manager/data`）

常见文件：
- `favorites.json`：收藏分组与路径
- `tags.json`：标签定义
- `image-tags.json`：图片-标签映射
- `image-notes.json`：图片笔记
- `custom-roots.json`：自定义根目录
- `settings.json`：设置（含回收站保留期、路径绑定）
- `trash-metadata.json`：回收站原路径/删除时间
- `image-meta-cache.json`：图片元数据缓存
- `launcher-tools.json`：外部工具
- `prompt-tool-links.json`：提示词工具链接
- `prompt-templates.json`：提示词模板
- `icons/`：提取图标缓存
- `image-variants/preview|thumb/`：预览与缩略图缓存

回收站位置：
- 当前：`comfy-manager/.trash`
- 兼容：存在旧路径迁移逻辑（legacy trash）

---

## 7. 核心业务链路速查

### 7.1 图片刷新链路
`fsnotify 事件` → `scheduleImagesChangedEvent()` → Wails `images:changed` → `App.vue handleRefresh()` → `fetchImages + fetchImageTags`

### 7.2 删除链路（软删除）
前端触发删除 → `DeleteImage(relPath)` → 移入 `.trash` + 写 `trash-metadata.json` → 发刷新事件

### 7.3 智能相册链路
前端请求字段/分组 → 后端按 `imageMetaCache` 聚合 → model/sampler/lora 缺缓存时触发扫描 → 返回分组计数+路径

### 7.4 标签链路
创建/更新/删除标签（`tags.json`） + 图片打标（`image-tags.json`） + 前端过滤/计数联动

---

## 8. 开发与扩展指南（面向后续维护）

### 8.1 新增后端能力（推荐步骤）
1. 在 `app.go` 新增方法（注意参数/返回可 JSON 序列化）
2. 重新生成/编译 Wails 绑定
3. 在 `frontend/src/api.js` 暴露包装函数
4. 在 composable 或组件中接入
5. 若影响文件结构或缓存，补充刷新事件与文档

### 8.2 修改前端状态逻辑的注意点
- 优先改 `useImages.js`，不要在组件中分散复制状态逻辑
- 涉及列表规模逻辑时注意 computed 成本（分页默认 50）
- 对 `activeRoot/activeSub/activeChild` 的变更需同步考虑分页重置与选择模式清空

### 8.3 路径与安全约束
- 前后端传输优先使用 `relPath`
- 涉及磁盘访问必须通过后端 `resolveRootPath` 系列校验
- 禁止在前端拼接绝对路径执行敏感操作

### 8.4 缓存一致性
- 文件操作后应触发刷新（或依赖 watcher 事件）
- 元数据缓存写回采用“按需 + 异步 warmup”策略
- 智能相册依赖 metadata 缓存，不命中时会补扫

---

## 9. AI 快速定位索引

### 后端
- 启动与绑定：`desktop-source/main.go`
- 核心业务：`desktop-source/app.go`
  - 图片列表：`GetImages`
  - 元数据：`GetImageMetadata`
  - 删除/回收站：`DeleteImage` / `GetTrashList` / `RestoreTrash`
  - 标签：`GetTags` / `AddTagToImage`
  - 收藏分组：`GetFavoriteGroups` / `AddImageToFavoriteGroup`
  - 智能相册：`GetSmartAlbums`

### 前端
- 根编排：`desktop-source/frontend/src/App.vue`
- 状态核心：`desktop-source/frontend/src/composables/useImages.js`
- API 桥：`desktop-source/frontend/src/api.js`
- 侧栏：`desktop-source/frontend/src/components/AppSidebar.vue`
- 画廊：`desktop-source/frontend/src/components/ImageGallery.vue`
- 灯箱：`desktop-source/frontend/src/components/Lightbox.vue`

---

## 10. 可持续编辑机制（每次改动后只更新这里）

> 目标：避免“整篇重写”，只增量维护。

### 10.1 变更记录模板

在每次功能开发/重构后，追加一条：

```markdown
- **YYYY-MM-DD | 主题**
  - 影响范围：后端/前端/数据结构/缓存/事件
  - 变更文件：
    - path/to/fileA
    - path/to/fileB
  - 行为变化：
    - 旧行为 -> 新行为
  - 兼容性：是否影响旧数据/旧路径/旧接口
  - AI 提示：下次若修改 X，优先查看某函数/组件
```

### 10.2 文档同步清单

发生以下变化时，务必同步本文件：
- 新增/删除后端公开方法
- 新增/变更 data 目录 JSON 结构
- 新增 watcher 行为、事件名或刷新机制
- `useImages.js` 的核心状态模型变更
- 导航根节点规则（fileTree）变更
- 智能相册字段或聚合逻辑变更

---

## 11. 最近变更记录

- **2026-04-14 | v1.4.1 - 修复提示词模板弹窗交互，统一发布产物命名**
  - 影响范围：前端/发布流程/桌面产物
  - 变更文件：
    - `desktop-source/frontend/src/components/PromptTemplateDialog.vue`
    - `desktop-source/frontend/src/components/ui/input/Input.vue`
    - `desktop-source/wails.json`
    - `desktop-app.exe`
  - 行为变化：
    - 修复提示词输入框聚焦时左右边框看起来消失的问题
    - 新增模板后保持表单打开，支持连续添加
    - 模板数量增多时，左侧底部“添加模板”按钮保持固定可见
    - 隐藏右侧表单滚动条，但保留滚轮滚动能力
    - Wails 构建输出文件名统一为 `desktop-app.exe`，与仓库根目录发布文件一致
  - 兼容性：
    - 不影响既有 `data/prompt-templates.json` 数据
    - 发布流程中复制源文件由 `comfy-manager-wails.exe` 调整为 `desktop-app.exe`
  - AI 提示：
    - 以后改提示词模板弹窗时，优先检查左侧列表容器的 `min-h-0/flex` 约束与右侧表单的保存后状态切换逻辑

- **2026-04-14 | v1.4.0 - 修复智能筛选路径叠加问题，改进侧边栏导航逻辑**
  - 影响范围：前端
  - 变更文件：
    - `desktop-source/frontend/src/App.vue`
    - `desktop-source/frontend/src/components/AppSidebar.vue`
    - `desktop-source/frontend/src/composables/useImages.js`
  - 行为变化：
    - 修复智能筛选激活状态下点击侧边栏目录时路径叠加问题
    - 侧边栏根目录切换逻辑优化：如果已在该根目录，则切换到dashboard而非空值
    - 切换导航节点时自动清除智能筛选状态，避免残留筛选影响新视图
    - 智能相册弹出框增加 `ml-2 mb-2` 边距，优化视觉对齐
  - 兼容性：
    - 不影响旧数据/旧接口
  - AI 提示：
    - 修改侧边栏导航逻辑时注意 `toggleRoot` 函数中的状态切换策略，确保清除所有可能残留的筛选状态

- **2026-04-13 | v1.3.0 - 智能筛选自动刷新、按日期整理文件、导出移动模式、图片上传**
  - 影响范围：后端/前端/事件刷新
  - 变更文件：
    - `desktop-source/app.go`（新增 `UploadImages`、`OrganizeFiles`）
    - `desktop-source/frontend/src/App.vue`（新增 `refreshSmartAlbumFilter`、`handleOrganizeFiles`，`handleRefresh` 调用刷新智能筛选）
    - `desktop-source/frontend/src/api.js`（新增 `OrganizeFiles`、`UploadImages`）
    - `desktop-source/frontend/src/components/AppSidebar.vue`（智能相册弹出框居中对齐+边距，新增”按日期整理文件”按钮）
    - `desktop-source/frontend/src/components/ExportDialog.vue`（导出支持移动模式，含二次确认）
  - 行为变化：
    - 智能筛选激活时，`images:changed` 事件触发后会自动重新获取筛选路径，新图片即时显示
    - 智能相册弹出框改为 `align=”center”` 居中对齐，增加 `ml-2 mb-2` 边距
    - 新增按日期整理文件功能（散落图片移动到年/月子文件夹）
    - 导出弹窗恢复移动模式选项，含二次确认弹窗
    - 新增 `UploadImages` 后端接口（从外部目录导入图片）
  - 兼容性：
    - 不影响旧数据/旧接口
  - AI 提示：
    - 修改智能筛选相关逻辑时，注意 `refreshSmartAlbumFilter` 在 `handleRefresh` 中被调用；若筛选不存在匹配项会自动清除

- **2026-04-12 | v1.2.0 - 提示词模板库**
  - 影响范围：后端/前端/数据结构
  - 变更文件：
    - `desktop-source/app.go`
    - `desktop-source/frontend/src/api.js`
    - `desktop-source/frontend/src/components/PromptTemplateDialog.vue`（新增）
    - `desktop-source/frontend/src/components/Lightbox.vue`
    - `desktop-source/frontend/src/components/AppSidebar.vue`
  - 行为变化：
    - 新增 PromptTemplate 实体（CRUD + JSON 持久化）
    - Lightbox 正向/反向 Prompt 区域新增”存为模板”按钮
    - 侧栏工具菜单新增”提示词模板”入口
    - 新增 PromptTemplateDialog 模板管理弹窗（搜索/分类/复制/编辑/删除）
  - 兼容性：
    - 新增 `data/prompt-templates.json`，不影响旧数据
  - AI 提示：
    - 模板 CRUD 模式同 PromptToolLink，无 mutex

- **2026-04-12 | 文档重建（结构化可持续维护版）**
  - 影响范围：文档
  - 变更文件：
    - `docs/PROJECT_CONTEXT.md`
  - 行为变化：
    - 旧版偏概览 -> 新版包含架构、链路、索引、维护模板
  - 兼容性：
    - 无代码行为变更
  - AI 提示：
    - 先读”第9节快速定位索引”，再进入代码修改。

---

*Generated/Updated by Claude Code (Opus 4.6) on 2026-04-14*
