# Comfy Manager

Comfy Manager（灵动图库）是一个面向 **ComfyUI output 目录** 的桌面图片管理器，基于 **Wails v2 + Go + Vue 3** 构建。

它围绕 ComfyUI 的真实使用流程，解决几个高频问题：

- output 目录越用越乱，历史图片难回看
- PNG 元数据、模型、LoRA、Prompt 信息分散，筛选效率低
- 看图、找词、拼 Prompt、存模板之间缺少顺手的本地工作流

当前稳定版本：`v2.1.6`

## 核心能力

- 绑定任意 ComfyUI `output` 目录
- 默认目录、日期归档目录、自定义目录并行浏览
- 日期产出工作台
- 模型 / LoRA / 标签 / 笔记搜索与筛选
- PNG 元数据查看、复制与工作流分析
- 收藏夹与收藏分组
- 标签、笔记、批量模式
- 自动规则引擎
- 提示词提示器与本地 Prompt 词库
- 设置中心与工具菜单配置

## v2.1.6 重点更新

- 新增 Windows 安装程序，支持选择安装目录
- 安装版运行时数据统一写入安装目录内的 `data/` 与 `.trash/`
- 安装包内置 `data/prompt-library/`，首次安装后即可使用提示词词库
- 提示词提示器分页支持固定显示，翻页不需要反复滚回顶部
- 修复分页条重复省略号问题
- 修复删除图片后重新生成同名文件时仍显示旧图的缓存问题
- 修复工作台总览中图片详情页“提示词提示器”按钮无响应的问题
- 优化 ComfyUI Prompt 提取逻辑，支持更复杂的工作流链路与候选排序
- 新增 Prompt 解析调试视图，便于排查不同工作流下的命中来源
- 修复元数据调试面板中超长 LoRA / Prompt 文本导致布局溢出的问题

## 适合的使用场景

- 长期使用 ComfyUI，需要稳定回看历史图片
- 经常按模型、LoRA、日期、标签回看结果
- 想在看图时顺手补 Prompt、存模板、继续出图
- 希望桌面管理器脱离固定 output 路径使用

## 文档

- [使用文档](./docs/README.md)
- [项目上下文](./docs/PROJECT_CONTEXT.md)
- [发布说明](./docs/RELEASE.md)
- [Windows 安装器说明](./docs/WINDOWS_INSTALLER.md)
- [v2.1.0 提示词提示器任务文档](./docs/V2.1.0_PROMPT_ASSISTANT_TASK.md)

## 发布信息

- 仓库地址：<https://github.com/dcajusteno-ops/qijiu-demo>
- Release：<https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.1.6>
