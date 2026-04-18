# Comfy Manager 项目上下文

当前稳定版本：`v2.1.5`  
更新时间：`2026-04-18`

## 1. 项目定位

**Comfy Manager（灵动图库）** 是一个基于 **Wails v2 + Go + Vue 3** 的桌面图片管理器，服务于 ComfyUI 出图后的浏览、筛选、整理、归档与提示词复用场景。

项目当前的核心目标：

- 浏览 ComfyUI 输出图片与 PNG 元数据
- 绑定任意 ComfyUI `output` 目录，而不是固定跟随 exe 所在位置
- 按日期、模型、LoRA、标签、收藏、笔记等维度筛选图片
- 提供“日期产出工作台”，快速回看最近产出
- 提供自动规则引擎，自动打标、归类与后处理
- 提供提示词编辑器，完成“看图 -> 找词 -> 拼 Prompt -> 存模板”的本地闭环

## 2. v2.1.5 版本重点

### 本次新增

- 新增 Windows 安装程序（NSIS）
- 新增可选择安装目录的安装流程
- 新增安装器内置的 `data/prompt-library/` 分发能力
- 新增安装器中文界面与快捷方式创建流程
- 新增 `docs/WINDOWS_INSTALLER.md` 安装说明文档

### 本次重点修复

- 修复安装版把运行时数据写入系统盘其他目录的问题
- 修复安装版首次运行后数据目录不跟随安装目录的问题
- 修复安装器默认路径和项目预期不一致的问题
- 修复发布流程里缺少安装包产物的问题

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
│  │  ├─ all_prompts_merged.cleaned.json
│  │  └─ manifest.cleaned.json
│  ├─ custom-prompt-entries.json
│  ├─ prompt-assistant-state.json
│  └─ prompt-templates.json
├─ tools/
│  └─ build_prompt_library.py
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
   │  │  ├─ lib/
   │  │  │  ├─ dateWorkbench.js
   │  │  │  ├─ shortcuts.js
   │  │  │  └─ utilityMenu.js
   │  │  └─ components/
   │  │     ├─ AppSidebar.vue
   │  │     ├─ Documentation.vue
   │  │     ├─ ImageGallery.vue
   │  │     ├─ Lightbox.vue
   │  │     ├─ PromptAssistantPage.vue
   │  │     ├─ PromptFilterSelect.vue
   │  │     ├─ PromptTemplateDialog.vue
   │  │     └─ SettingsCenterDialog.vue
   │  └─ wailsjs/
   └─ build/
      └─ windows/
         └─ installer/
            ├─ project.nsi
            └─ wails_tools.nsh
```

## 5. 后端核心结构

主文件：`desktop-source/app.go`

主要职责：

- 扫描图片目录
- 解析 PNG 元数据、模型、LoRA、工作流节点数等
- 维护图片元数据缓存
- 维护目录绑定、自定义目录、收藏夹、标签、笔记、规则
- 维护提示词词库、自定义提示词、提示词模板、提示词工作台状态
- 提供图片删除、恢复、清理缓存、清理空目录、日期整理等能力

v2.1.5 关键数据结构：

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
- 统一装配侧边栏、主页、图库、工作台、设置相关页面
- 处理根级视图跳转
- 协调日期工作台、图库、提示词编辑器之间的入口联动

### `frontend/src/composables/useImages.js`

核心业务状态层，负责：

- 图片、收藏、标签、笔记、自定义目录状态
- 搜索、分页、排序、堆叠显示
- 日期工作台的日期范围、模型、LoRA 筛选
- 侧边栏目录树生成
- 当前图库结果集计算

### `frontend/src/components/PromptAssistantPage.vue`

v2.1.0 的提示词核心页面，负责：

- 正向 / 反向 Prompt 编辑区
- 词库搜索、筛选、分页浏览
- 常用预设词包
- 自定义提示词新增、删除
- 收藏、最近、模板保存
- 与图库 / Lightbox 上下文联动

### `frontend/src/components/PromptFilterSelect.vue`

用于替代系统原生下拉，解决：

- 下拉内容超出软件窗口
- 超长分类项显示不稳定
- 筛选下拉与整体页面风格不一致

### `frontend/src/components/PromptTemplateDialog.vue`

负责：

- 提示词模板查看、搜索、分类
- 复制模板内容
- 模板新增、编辑、删除
- 组合模板与正向 / 反向模板复用

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

v2.1.5 重点新增 / 调整：

- `data/prompt-library/` 作为运行时正式词库目录
- `custom-prompt-entries.json` 用于保存“我的词库”
- `prompt-assistant-state.json` 用于保存收藏、最近、分页和筛选状态
- 删除自定义提示词时，需要同步清理收藏与最近中的失效 id
- 安装版运行时数据固定写入安装目录自身的 `data/` 与 `.trash/`

## 8. 核心业务链路

### 图片刷新链路

`fsnotify` -> `images:changed` -> 前端页面订阅刷新

### 日期工作台链路

`GetImages()` -> `useImages.js` 提取日期目录 -> `dateWorkbenchSummary` 统计 -> `DateWorkbench.vue` 展示 -> 一键跳回图库

### 提示词编辑器链路

`data/prompt-library/all_prompts_merged.cleaned.json` -> `app.go` 加载词库 -> `PromptAssistantPage.vue` 搜索 / 筛选 / 分页 -> 拼装正向 / 反向 Prompt -> 保存模板

### 自定义提示词链路

用户新增词条 -> 后端查重 -> 写入 `custom-prompt-entries.json` -> 前端并入当前词库 -> 可收藏 / 可再次编辑使用

## 9. 开发注意事项

- 涉及筛选逻辑，优先检查 `useImages.js` 和 `PromptAssistantPage.vue`
- 涉及提示词状态持久化，优先检查 `app.go` 中的 `PromptAssistantState`
- 涉及提示词下拉溢出，优先检查 `PromptFilterSelect.vue`
- 涉及模板复制异常，优先检查 `PromptTemplateDialog.vue`
- 涉及中文乱码，优先检查文件编码和字符串字面量，不要先怀疑 JSON 数据损坏

## 10. 发布要求

发布版本时至少同步以下内容：

- 根目录 `desktop-app.exe`
- 根目录 `ComfyManager-amd64-installer.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

## 11. 最近变更记录

### 2026-04-18 | v2.1.5

- 新增 Windows 安装程序与安装目录选择页
- 安装包开始内置提示词词库目录
- 调整安装版运行时数据目录，统一跟随安装目录
- 更新 README、项目上下文、发布文档与安装器说明

### 2026-04-18 | v2.1.0

- 新增提示词编辑器独立页面
- 新增清洗后提示词词库副本与清洗脚本
- 新增自定义提示词、预设词包、分页、收藏、最近、模板复用
- 修复提示词编辑器多轮布局问题并统一删除确认弹窗
- 更新软件内置文档、README 与发布文档

### 2026-04-17 | v2.0.1

- 更新软件内置“使用文档”页面
- 修复内置文档页中文乱码和旧版本说明残留
- 同步更新 GitHub README 与 docs 文档

### 2026-04-17 | v2.0.0

- 新增任意位置绑定 ComfyUI `output` 目录
- 新增设置中心，统一承载工具与系统设置
- 新增工具菜单配置化能力
- 新增日期工作台日期范围筛选
