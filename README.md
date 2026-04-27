# Comfy Manager

当前稳定版本：`v3.0.1`

Comfy Manager 是一个面向 **ComfyUI output 目录** 的桌面整理工具，基于 **Wails v2 + Go + Vue 3** 构建。  
它不是单纯的“看图器”，而是把 ComfyUI 出图之后真正高频的事情串成一个完整工作流：

- 回看最近产出
- 按日期、模型、LoRA、标签、收藏、笔记快速筛图
- 查看 PNG 元数据、Prompt 和 Workflow
- 整理目录、清理缓存、治理侧边栏目录结构
- 通过提示词助手、模板和自动规则提高复用效率

## v3.0.1 这次有什么变化

`v3.0.1` 是在 `v3.0.0` 基础上的稳定性修复版本，重点是把当前发现的归档目录与发布产物问题一次收口：

- 后端 Go 源码从根目录整理进 `desktop-source/backend/`
- 后端文件按 `app_core_* / app_feature_* / app_support_* / app_types_*` 分组
- 修复重构后前端对 Wails 绑定名的兼容问题
- 修复首页最新作品卡片里异常显示的“路”字文案
- 修复“日期归档目录”进入后黑屏的运行时错误
- 修复日期归档树中年份分类无法正常折叠的交互问题
- 确认根目录 `desktop-app.exe` 与 `ComfyManager-amd64-installer.exe` 和 `build/bin` 产物哈希一致
- 重写软件内使用文档、GitHub README 与发布文档
- 重新打通 Windows 安装包生成链路
- 统一重打 `desktop-app.exe` 与 `ComfyManager-amd64-installer.exe`

## 核心能力

- 绑定任意 ComfyUI `output` 目录，而不是依赖 exe 的相对位置
- 同时浏览默认目录、日期归档目录和自定义目录
- 工作台总览、日期产出工作台、收藏夹、数据视界
- 模型 / LoRA / 标签 / 笔记 / 收藏联合筛图
- PNG 元数据、Prompt、Workflow 查看与复制
- 提示词助手、提示词模板与本地词库复用
- 自动规则：按模型、LoRA、Prompt、文件名自动打标或归类
- 目录健康中心、空文件夹清理、缓存治理
- 回收站保护与恢复
- 快捷键、工具菜单、性能模式等设置中心能力

## 适合谁

- 长期使用 ComfyUI，本地 output 已经很大的人
- 需要经常回看“今天、昨天、最近 7 天、本月”产出的用户
- 习惯按模型、LoRA、题材、风格追踪作品的用户
- 想把 Prompt、模板、收藏和自动整理流程统一起来的用户

## 快速开始

### 1. 安装或直接运行

- 安装版：运行根目录的 [ComfyManager-amd64-installer.exe](./ComfyManager-amd64-installer.exe)
- 便携版：直接运行根目录的 [desktop-app.exe](./desktop-app.exe)

### 2. 首次进入先绑定 output

从 `v2.x` 开始，程序不再猜测 exe 的上级目录。首次进入会要求你手动选择 **ComfyUI 的真实 output 文件夹**。

正确示例：

```text
D:\AiImg\ComfyUI-aki-v3\ComfyUI\output
```

不是选择：

```text
D:\AiImg\ComfyUI-aki-v3\ComfyUI\output\comfy-manager
```

绑定完成后，程序会自动识别：

- `rootDir`：`output` 的上一级目录
- `outputDir`：当前绑定的真实输出目录

### 3. 建议第一轮使用顺序

1. 先看“工作台总览”，确认最新作品、总量和占用是否正常
2. 再进“日期产出”，按今天、昨天、最近 7 天回看近期创作
3. 需要细筛时进入“默认目录”或自定义目录图库
4. 常用 Prompt 与模板建议放进“提示词助手”统一复用

## 文档导航

- [详细使用文档](./docs/README.md)
- [项目上下文](./docs/PROJECT_CONTEXT.md)
- [后端文件地图](./docs/BACKEND_FILE_MAP.md)
- [重构收敛计划](./docs/REFACTOR_PLAN.md)
- [发布说明](./docs/RELEASE.md)
- [Windows 安装器说明](./docs/WINDOWS_INSTALLER.md)

## 项目结构

```text
comfy-manager/
├─ README.md
├─ docs/
├─ data/
├─ .trash/
├─ desktop-app.exe
├─ ComfyManager-amd64-installer.exe
└─ desktop-source/
   ├─ main.go
   ├─ backend/
   ├─ frontend/
   ├─ build/
   └─ wails.json
```

## 仓库信息

- GitHub: [dcajusteno-ops/qijiu-demo](https://github.com/dcajusteno-ops/qijiu-demo)
- 推荐标签：`v3.0.1`

## 发布产物

当前根目录应始终保持以下发布产物与文档同步：

- `desktop-app.exe`
- `ComfyManager-amd64-installer.exe`
- `README.md`
- `docs/README.md`
- `docs/PROJECT_CONTEXT.md`
- `docs/RELEASE.md`

