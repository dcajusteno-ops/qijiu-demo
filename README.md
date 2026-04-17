# Comfy Manager

Comfy Manager（灵动图库）是一个面向 **ComfyUI output 目录** 的桌面图片管理器，基于 **Wails v2 + Go + Vue 3** 构建。

它围绕 ComfyUI 的真实使用流程，解决三类高频问题：

- output 目录越来越乱，历史图片难回看
- PNG 元数据、模型、LoRA、Prompt 信息分散，筛选效率低
- 看图、找词、拼 Prompt、存模板之间缺少顺手的本地工作流

当前版本：`v2.1.0`

## 核心功能

- 绑定任意 ComfyUI `output` 目录
- 默认目录、日期归档目录、自定义目录并行浏览
- 日期产出工作台
- 模型 / LoRA / 标签 / 笔记搜索与筛选
- PNG 元数据查看与复制
- 收藏夹与收藏分组
- 标签、笔记、批量模式
- 自动规则引擎
- 提示词编辑器（本地提示词词库、Prompt 拼装、模板沉淀）
- 设置中心与工具菜单配置

## v2.1.0 更新重点

- 新增独立页面式 **提示词编辑器**
- 运行时改为读取 `data/prompt-library/` 下的清洗词库副本
- 支持中文 / 英文搜索、来源 / 分类 / 子分类 / 作用域筛选
- 支持正向 / 反向 Prompt 拼装、分页浏览、去重、顺序调整、复制
- 新增常用预设词包、自定义提示词持久化、查重、删除
- 新增提示词收藏、最近使用、模板保存与组合复用
- 图库页、Lightbox、工具菜单已接入提示词编辑器入口
- 更新软件内置“使用文档”与外部 docs，统一到 `v2.1.0`

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
- 你希望在回图时顺手补 Prompt、存模板、继续出图
- 你希望桌面管理器可以脱离固定 output 路径使用

## 文档

- [使用文档](./docs/README.md)
- [项目上下文](./docs/PROJECT_CONTEXT.md)
- [发布说明](./docs/RELEASE.md)
- [v2.1.0 提示词编辑器任务书](./docs/V2.1.0_PROMPT_ASSISTANT_TASK.md)

## 发布信息

- 仓库地址：<https://github.com/dcajusteno-ops/qijiu-demo>
- Release：<https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.1.0>
