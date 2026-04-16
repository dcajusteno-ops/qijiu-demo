# 灵动图库 (Comfy Manager)

基于 **Wails v2（Go + Vue 3）** 的桌面图像管理器，面向 ComfyUI 输出图片场景。

当前稳定版本：`v1.7.0`

- Release 页面：[v1.7.0](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.7.0)
- 当前发布产物：`output/comfy-manager/desktop-app.exe`

## 功能特性

- **图片浏览**：自动扫描 ComfyUI 输出目录，按日期归档展示，支持多级目录导航
- **搜索 MVP**：支持按文件名、路径、Prompt、模型、LoRA、采样器、Negative、尺寸、标签、笔记快速搜索
- **自动规则引擎**：支持按模型、LoRA、采样器、Prompt、尺寸、文件名等条件自动执行打标、加入收藏分组、移动目录
- **规则进度反馈**：手动“立即执行一次”时显示执行进度、当前图片与执行结果摘要
- **个人中心**：维护昵称、简介、主页链接、每日目标、默认进入页、展示图，并集中管理自动规则开关
- **标签系统**：自定义标签分类、颜色，支持按标签筛选与批量打标
- **收藏分组**：创建多个收藏分组，自由管理精选图片
- **提示词模板**：从图片 Prompt 一键存为模板，分类管理、搜索过滤、一键复制
- **数据视界**：以生成历史时间线、趋势图、活跃热图回看创作过程，可快速定位任意时间段作品
- **全局快捷键**：支持系统级快捷键切换总览、图库、收藏、文档，并可刷新图库、折叠侧栏、切换批量模式
- **回收站**：软删除机制，支持恢复、清空、保留期自动清理
- **批量管理**：多选模式，批量删除、批量加标签、批量移动、批量导出
- **按日期整理**：将散落在根目录的图片按年/月自动归档到子文件夹
- **导出/上传**：复制或移动图片到指定目录，从外部目录导入图片
- **外部工具**：自定义启动器，一键打开外部编辑器或工具

## v1.7.0 重点更新

- 新增“搜索 MVP”，前端已接入后端 `searchText`，LoRA / 采样器 / Negative / 尺寸等元数据可直接搜索
- 新增“自动规则引擎”，支持默认规则、新建规则、编辑规则、启停规则、立即执行与进度展示
- 自动规则入口与控制项整合进个人中心页面，布局保持简约风
- 删除“智能相册”整块功能，统一收敛到搜索与规则两条主链路
- 修复个人中心头像路径不在 `rootDir` 时无法显示的问题

## 截图

| 工作室概览 | 图片浏览 |
|:---:|:---:|
| ![overview](images/overview.png) | ![archive_view](images/archive_view.png) |

| 批量管理 | 回收站 |
|:---:|:---:|
| ![batch_manage](images/batch_manage.png) | ![recycle_bin](images/recycle_bin.png) |

## 下载安装

1. 前往 [Releases](https://github.com/dcajusteno-ops/qijiu-demo/releases) 页面下载最新版 `desktop-app.exe`
2. 在 ComfyUI 的 `output/` 目录下创建 `comfy-manager` 文件夹
3. 将 `desktop-app.exe` 放入该文件夹中
4. 双击运行

最终目录结构如下：

```text
ComfyUI/
└── output/
    ├── 2026/
    └── comfy-manager/
        ├── desktop-app.exe
        ├── data/
        └── .trash/
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Wails v2 |
| 前端 | Vue 3 + Vite + Tailwind CSS 4 + shadcn-vue |
| 图标 | lucide-vue-next |
| 通信 | Wails Bindings + Runtime Events |

## 项目结构

```text
comfy-manager/
├── desktop-app.exe
├── docs/
│   ├── README.md
│   ├── RELEASE.md
│   ├── PROJECT_CONTEXT.md
│   └── V1.7_SEARCH_AND_RULES_PLAN.md
└── desktop-source/
    ├── app.go
    ├── main.go
    └── frontend/
        └── src/
            ├── App.vue
            ├── api.js
            ├── composables/
            └── components/
```

## 本地开发

需要安装 [Go](https://go.dev/)、[Node.js](https://nodejs.org/)、[Wails CLI](https://wails.io/docs/gettingstarted/installation)。

```bash
cd desktop-source

cd frontend && npm install && cd ..
wails dev
wails build
```

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.7.0 | 2026-04-16 | 新增搜索 MVP 与自动规则引擎；自动规则整合进个人中心并支持默认规则、执行进度；删除智能相册；修复头像资源路径与搜索命中链路 |
| v1.6.0 | 2026-04-15 | 新增简约风“个人中心”：支持资料编辑、默认进入页、单张展示图与创作摘要；继续优化“数据视界”节点详情与滚动布局 |
| v1.5.0 | 2026-04-15 | 新增“数据视界”生成历史时间线与活跃热图，支持趋势图悬浮读数与节点回溯；新增全局快捷键设置与系统级视图切换 |
| v1.4.1 | 2026-04-14 | 修复提示词模板弹窗：聚焦边框显示、连续添加、底部添加按钮固定显示，并统一发布产物为 `desktop-app.exe` |
| v1.4.0 | 2026-04-14 | 修复智能筛选路径叠加问题，改进侧边栏导航逻辑 |
| v1.3.0 | 2026-04-13 | 修复智能筛选不自动刷新、弹出框居中对齐，新增按日期整理文件、导出移动模式与图片上传 |
| v1.2.0 | 2026-04-12 | 新增提示词模板库：从图片 Prompt 一键存模板、分类管理、搜索过滤、一键复制 |
| v1.1.0 | 2026-04-12 | 修复文件变更自动刷新、Ctrl+S 保存笔记等交互细节 |
| v1.0.0 | 2025-04-11 | 初始发布 |

## 许可证

MIT
