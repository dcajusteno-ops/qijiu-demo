# 灵动图库 (Comfy Manager)

基于 **Wails v2 (Go + Vue 3)** 的桌面图像管理器，专为 ComfyUI 输出图片场景设计。

当前稳定版本：`v1.5`

- Release 页面：[v1.5](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.5)
- 当前发布产物：`output/comfy-manager/desktop-app.exe`

## 功能特性

- **图片浏览** — 自动扫描 ComfyUI 输出目录，按日期归档展示，支持多级目录导航
- **元数据解析** — 解析 PNG 内嵌的 ComfyUI / WebUI 工作流参数（模型、采样器、LoRA、Prompt 等）
- **智能相册** — 按模型、采样器、LoRA、尺寸自动归类，一键筛选同参数图片
- **标签系统** — 自定义标签分类、颜色，支持按标签筛选与批量打标
- **收藏分组** — 创建多个收藏分组，自由管理精选图片
- **提示词模板** — 从图片 Prompt 一键存为模板，分类管理、搜索过滤、一键复制
- **数据视界** — 以生成历史时间线、趋势图、活跃热图回看创作过程，可快速定位任意时间段作品
- **全局快捷键** — 支持系统级快捷键切换总览/图库/收藏/文档，并可刷新图库、折叠侧栏、切换批量模式
- **回收站** — 软删除机制，支持恢复、清空、保留期自动清理
- **批量管理** — 多选模式，批量删除、批量加标签、批量移动、批量导出
- **按日期整理** — 将散落在根目录的图片按年/月自动归档到子文件夹
- **导出/上传** — 复制或移动图片到指定目录，从外部目录导入图片
- **外部工具** — 自定义启动器，一键打开外部编辑器或工具
- **深色模式** — 支持亮色/暗色主题切换

## 截图

| 工作室概览 | 图片浏览 |
|:---:|:---:|
| ![overview](images/overview.png) | ![archive_view](images/archive_view.png) |

| 批量管理 | 回收站 |
|:---:|:---:|
| ![batch_manage](images/batch_manage.png) | ![recycle_bin](images/recycle_bin.png) |

| 深色模式 |
|:---:|
| ![dark_mode](images/dark_mode.png) |

## 下载安装

1. 前往 [Releases](https://github.com/dcajusteno-ops/qijiu-demo/releases) 页面下载最新版 `desktop-app.exe`
2. 在 ComfyUI 的 `output/` 目录下创建 `comfy-manager` 文件夹
3. 将 `desktop-app.exe` 放入该文件夹中
4. 双击运行

最终目录结构如下：

```
ComfyUI/
└── output/                  ← ComfyUI 图片输出目录
    ├── 2026/                 ← 生成的图片
    └── comfy-manager/        ← 新建此文件夹
        ├── desktop-app.exe   ← 放在这里
        ├── data/             ← 运行后自动生成（配置与缓存）
        └── .trash/           ← 运行后自动生成（回收站）
```

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Wails v2 |
| 前端 | Vue 3 + Vite + Tailwind CSS 4 + shadcn-vue |
| 图标 | lucide-vue-next |
| 通信 | Wails Bindings + Runtime Events |

## 项目结构

```
comfy-manager/
├── desktop-app.exe              # 发布产物
├── docs/
│   ├── RELEASE.md               # 版本发布指南
│   └── PROJECT_CONTEXT.md       # 开发者上下文文档
└── desktop-source/
    ├── app.go                   # 后端核心业务
    ├── main.go                  # Wails 启动入口
    └── frontend/
        └── src/
            ├── App.vue          # 根组件
            ├── api.js           # JS API 桥
            ├── composables/     # 状态管理
            └── components/      # 业务组件
```

## 本地开发

需要安装 [Go](https://go.dev/)、[Node.js](https://nodejs.org/)、[Wails CLI](https://wails.io/docs/gettingstarted/installation)。

```bash
cd desktop-source

# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式运行
wails dev

# 编译打包
wails build
```

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.5 | 2026-04-15 | 新增“数据视界”生成历史时间线与活跃热图，支持趋势图悬浮读数与节点回溯；新增全局快捷键设置与系统级视图切换 |
| v1.4.1 | 2026-04-14 | 修复提示词模板弹窗：聚焦边框显示、连续添加、底部添加按钮固定显示，并统一发布产物为 `desktop-app.exe` |
| v1.4.0 | 2026-04-14 | 修复智能筛选路径叠加问题，改进侧边栏导航逻辑 |
| v1.3.0 | 2026-04-13 | 修复智能筛选不自动刷新、弹出框居中对齐、按日期整理文件、导出移动模式、图片上传 |
| v1.2.0 | 2026-04-12 | 新增提示词模板库：从图片 Prompt 一键存模板、分类管理、搜索过滤、一键复制 |
| v1.1.0 | 2026-04-12 | 智能相册滚动修复、文件变更自动刷新、Ctrl+S 保存笔记 |
| v1.0.0 | 2025-04-11 | 初始发布，提示词辅助工具改进 + 自定义链接管理 |

## 许可证

MIT
