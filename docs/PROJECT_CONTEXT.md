# Comfy Manager 项目上下文

当前稳定版本：`v2.1.6`  
更新时间：`2026-04-19`

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2 + Go + Vue 3** 的桌面图片管理器，服务于 ComfyUI 出图后的浏览、筛选、整理、归档与 Prompt 复用场景。

项目当前的核心目标：

- 浏览 ComfyUI 输出图片与 PNG 元数据
- 绑定任意 ComfyUI `output` 目录，而不是固定跟随 exe 所在位置
- 按日期、模型、LoRA、标签、收藏、笔记等维度筛选图片
- 提供日期产出工作台，快速回看最近出图
- 提供自动规则引擎，自动打标、归类与后处理
- 提供提示词提示器，完成“看图 -> 找词 -> 拼 Prompt -> 存模板”的本地闭环

## 2. v2.1.6 版本重点

### 本次新增

- Windows 安装程序（NSIS）
- 安装目录选择流程
- 安装包内置 `data/prompt-library/`
- Prompt 解析调试视图

### 本次重点修复

- 安装版运行时数据不再写入其他系统目录
- 提示词提示器分页支持固定显示
- 分页条重复省略号修复
- 图片删除后重新生成同名文件时的旧缓存显示修复
- 工作台总览中 Lightbox 打开提示词提示器无响应修复
- ComfyUI Prompt 提取逻辑增强，支持更复杂的节点链路
- Prompt 解析调试面板中的超长文本溢出修复

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
│  ├─ PROJECT_CONTEXT.md
│  ├─ WINDOWS_INSTALLER.md
│  └─ V2.1.0_PROMPT_ASSISTANT_TASK.md
├─ data/
│  ├─ prompt-library/
│  ├─ custom-prompt-entries.json
│  ├─ prompt-assistant-state.json
│  └─ prompt-templates.json
├─ .trash/
├─ desktop-app.exe
├─ ComfyManager-amd64-installer.exe
└─ desktop-source/
   ├─ app.go
   ├─ shortcuts.go
   ├─ main.go
   ├─ wails.json
   ├─ frontend/
   │  ├─ src/
   │  │  ├─ App.vue
   │  │  ├─ api.js
   │  │  ├─ composables/
   │  │  │  └─ useImages.js
   │  │  └─ components/
   │  │     ├─ Home.vue
   │  │     ├─ ImageGallery.vue
   │  │     ├─ Lightbox.vue
   │  │     ├─ PromptAssistantPage.vue
   │  │     └─ PromptTemplateDialog.vue
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
- 维护提示词词库、自定义提示词、模板与提示词状态
- 提供图片删除、恢复、清理缓存、清理空目录、日期整理等能力

v2.1.6 关键数据结构：

- `ImageMetadata`
- `PromptDebugInfo`
- `PromptLibraryEntry`
- `PromptAssistantState`
- `PromptTemplate`
- `FavoriteGroup`
- `CustomRoot`
- `DirectoryBinding`

## 6. 前端核心结构

### `frontend/src/App.vue`

根组件，负责：

- 注入全局 `toast`
- 装配侧边栏、主页、图库、工作台、设置等核心页面
- 处理根级视图切换
- 协调日期工作台、图库、提示词提示器之间的联动

### `frontend/src/composables/useImages.js`

核心状态层，负责：

- 图片、收藏、标签、笔记、自定义目录状态
- 搜索、分页、排序、堆叠显示
- 日期工作台的日期范围、模型、LoRA 筛选
- 侧边栏目录树生成
- 图片显示路径缓存处理

### `frontend/src/components/PromptAssistantPage.vue`

提示词提示器核心页面，负责：

- 正向 / 反向 Prompt 编辑区
- 词库搜索、筛选、分页浏览
- 预设词包与模板复用
- 自定义提示词新增、删除、收藏、最近使用
- 与图库 / Lightbox 上下文联动

### `frontend/src/components/Lightbox.vue`

图片详情与元数据视图，负责：

- PNG 元数据展示
- Prompt、Workflow、参数复制
- 打开提示词提示器
- Prompt 解析调试视图
- 标签、收藏、笔记操作

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
- `prompt-templates.json`
- `custom-prompt-entries.json`
- `prompt-assistant-state.json`

v2.1.6 重点调整：

- `data/prompt-library/` 作为运行时正式词库目录
- `custom-prompt-entries.json` 保存“我的词库”
- `prompt-assistant-state.json` 保存收藏、最近、分页和筛选状态
- 图片显示路径增加基于时间与大小的缓存版本参数

## 8. 核心业务链路

### 图片刷新链路

`fsnotify` -> `images:changed` -> 前端订阅 -> 图库 / 工作台刷新

### 日期工作台链路

`GetImages()` -> `useImages.js` 提取日期目录 -> `dateWorkbenchSummary` 统计 -> `DateWorkbench.vue` 展示

### 提示词提示器链路

`data/prompt-library/all_prompts_merged.cleaned.json` -> `app.go` 加载词库 -> `PromptAssistantPage.vue` 搜索 / 筛选 / 分页 -> 拼装正向 / 反向 Prompt -> 保存模板

### Prompt 解析链路

PNG `prompt` JSON -> `extractComfyPromptSummary()` -> 候选收集与打分 -> 正向 / 反向命中 -> 调试信息返回 -> `Lightbox.vue` 展示

## 9. 开发注意事项

- 涉及筛选逻辑，优先检查 `useImages.js` 和 `PromptAssistantPage.vue`
- 涉及提示词状态持久化，优先检查 `app.go` 中的 `PromptAssistantState`
- 涉及 Prompt 提取不准，优先检查 `extractComfyPromptSummary()` 与调试视图
- 涉及图片详情联动，优先检查 `Lightbox.vue`、`Home.vue`、`App.vue`
- 涉及中文乱码，优先检查文件编码和字符串字面量
- 涉及发布，必须同步根目录产物与 `README/docs`

## 10. 发布要求

发布版本时至少同步以下内容：

- 根目录 `desktop-app.exe`
- 根目录 `ComfyManager-amd64-installer.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

## 11. 最近变更记录

### 2026-04-19 | v2.1.6

- 基于已完成的修复内容，补发 `v2.1.6` 版本号与 GitHub 标签
- 同步根目录 README 与 docs 文档版本信息

### 2026-04-19 | v2.1.5

- 修复提示词提示器分页固定显示与重复省略号问题
- 修复图片删除后重新生成同名文件时的旧缓存显示问题
- 修复工作台总览中 Lightbox 打开提示词提示器无响应的问题
- 强化 ComfyUI Prompt 提取逻辑，支持更复杂的工作流路径
- 新增 Prompt 解析调试视图
- 修复 Prompt 调试卡片中长文本导致的布局溢出问题

### 2026-04-18 | v2.1.5

- 新增 Windows 安装程序与安装目录选择页
- 安装包开始内置提示词词库目录
- 安装版运行数据统一跟随安装目录
- 同步更新 README、项目上下文、发布文档与安装器说明

### 2026-04-18 | v2.1.0

- 新增提示词提示器独立页面
- 新增自定义提示词、预设词包、分页、收藏、最近、模板复用
- 同步更新 README 与项目文档
