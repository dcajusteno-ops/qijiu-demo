# Comfy Manager

Comfy Manager（灵动图库）是一个面向 **ComfyUI 输出目录** 的桌面图片管理器，基于 **Wails v2 + Go + Vue 3** 构建。

它专门解决 ComfyUI 长期使用过程中常见的几个问题：

- output 目录越来越乱，图片回看困难
- 模型、LoRA、Prompt 信息分散，筛选效率低
- 想按日期、目录、收藏、标签统一整理，但现有方案不顺手
- 程序不应被固定在某一个 output 路径里

当前版本：`v2.0.1`

## 核心功能

- 绑定任意 ComfyUI `output` 目录
- 默认目录、日期归档目录、自定义目录并行浏览
- 日期产出工作台
- 模型 / LoRA 筛选
- PNG 元数据查看与复制
- 收藏夹与收藏分组
- 标签、笔记、批量模式
- 自动规则引擎
- 设置中心与工具菜单配置

## v2.0.1 更新重点

- 更新软件内置“使用文档”页面，和外部 README / docs 文档保持一致
- 修正内置文档页的中文乱码与旧版本说明残留
- 新增任意位置绑定 ComfyUI `output` 目录
- 新增设置中心，统一收纳快捷键、缓存、文件夹维护、主题等功能
- 新增工具菜单显示/隐藏与顺序调整
- 日期产出工作台支持日期范围筛选
- 优化首页跳转、侧边栏结构和多处中文乱码问题

## 技术栈

- Go
- Wails v2
- Vue 3
- Vite
- Tailwind CSS 4
- shadcn-vue

## 适合的使用场景

- 你长期使用 ComfyUI，需要稳定浏览历史作品
- 你经常按模型、LoRA、日期回看结果
- 你希望用自动规则做打标或归类
- 你希望桌面管理器可以脱离固定 output 路径使用

## 文档

- [使用文档](./docs/README.md)
- [项目上下文](./docs/PROJECT_CONTEXT.md)
- [发布说明](./docs/RELEASE.md)

## 发布信息

- 仓库地址：<https://github.com/dcajusteno-ops/qijiu-demo>
- Release：<https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.0.1>
