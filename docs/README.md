# 灵动图库（Comfy Manager）

面向 ComfyUI `output/` 场景的桌面图片管理器，基于 Wails v2、Go 和 Vue 3 构建。

当前稳定版本：`v1.8.0`

- Release 页面：[v1.8.0](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.8.0)
- 当前发布产物：`output/comfy-manager/desktop-app.exe`

## 核心能力

- 图片浏览：自动扫描 ComfyUI 输出目录，支持日期归档、多级目录和收藏分组。
- 日期产出工作台：新增“今天 / 昨天 / 最近7天 / 本月 / 指定日期”入口，适合直接回看近期产出。
- 模型 / LoRA 筛选：从图片 metadata 中提取模型和 LoRA，并与图库视图联动。
- 搜索 MVP：支持按文件名、路径、Prompt、模型、LoRA、Negative、标签、笔记等内容搜索。
- 自动规则引擎：支持按模型、LoRA、Prompt、采样器、尺寸、文件名等条件自动打标签、加收藏组、移动目录。
- 个人中心：管理资料、默认进入页、展示图和常用设置。
- 数据视界：查看生成历史时间线、趋势图和活跃热图。
- 标签和收藏分组：支持自定义管理、筛选和批量操作。
- 回收站：软删除、恢复、清空、保留期清理。
- 批量管理：支持多选、导出、移动、批量加标签和批量删除。

## v1.8.0 重点更新

- 新增“日期产出工作台”，直接围绕日期目录查看产出。
- 新增模型 / LoRA 筛选栏，并与图库结果联动。
- 新增日期工作台统计卡片与最近活跃日期入口。
- 优化侧边栏入口结构，减少顶部高频按钮数量。
- 改进模型名归一化匹配，尽量消除路径、扩展名、大小写差异带来的筛选误差。
- 修复旧版筛选值残留导致新版界面看似“全部条件”、实际仍被过滤的问题。

## 安装方式

1. 前往 [Releases](https://github.com/dcajusteno-ops/qijiu-demo/releases) 页面下载最新 `desktop-app.exe`
2. 在 ComfyUI 的 `output/` 目录下创建 `comfy-manager` 文件夹
3. 将 `desktop-app.exe` 放入该目录
4. 双击运行

目录示例：

```text
ComfyUI/
└─ output/
   ├─ 2026-04-16/
   ├─ 2026-04-14/
   └─ comfy-manager/
      ├─ desktop-app.exe
      ├─ data/
      └─ .trash/
```

## 技术栈

| 层级 | 技术 |
|---|---|
| 后端 | Go + Wails v2 |
| 前端 | Vue 3 + Vite + Tailwind CSS 4 + shadcn-vue |
| 图标 | lucide-vue-next |
| 通信 | Wails Bindings + Runtime Events |

## 项目结构

```text
comfy-manager/
├─ desktop-app.exe
├─ docs/
│  ├─ README.md
│  ├─ RELEASE.md
│  ├─ PROJECT_CONTEXT.md
│  ├─ V1.8_DATE_MODEL_PLAN.md
│  └─ V1.8_DATE_MODEL_IMPLEMENTATION.md
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
      │     ├─ ImageGallery.vue
      │     ├─ DateWorkbench.vue
      │     └─ ProfileCenter.vue
      └─ wailsjs/
```

## 本地开发

依赖：

- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
cd desktop-source
cd frontend && npm install && cd ..
wails dev
wails build
```

## 版本历史

| 版本 | 日期 | 说明 |
|---|---|---|
| v1.8.0 | 2026-04-16 | 新增日期产出工作台、模型 / LoRA 筛选、日期统计入口，优化侧边栏并修复筛选残留问题 |
| v1.7.0 | 2026-04-16 | 新增搜索 MVP 与自动规则引擎，自动规则整合进个人中心 |
| v1.6.0 | 2026-04-15 | 新增个人中心并细化数据视界 |
| v1.5.0 | 2026-04-15 | 新增数据视界与全局快捷键 |
| v1.4.1 | 2026-04-14 | 统一发布产物为 `desktop-app.exe` |
| v1.0.0 | 2025-04-11 | 初始发布 |

## 许可证

MIT
