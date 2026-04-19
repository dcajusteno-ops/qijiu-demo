# Comfy Manager 使用文档

当前稳定版本：`v2.1.6`

## 项目简介

Comfy Manager（灵动图库）是一个面向 **ComfyUI output 目录** 的桌面图片管理器，使用 **Wails v2 + Go + Vue 3** 构建。

它覆盖的核心工作流包括：

- 回看图片
- 筛选模型 / LoRA / 标签 / 日期
- 查看 PNG 元数据
- 收藏与整理
- 查找提示词、拼 Prompt、保存模板

## v2.1.6 功能概览

### 核心能力

- 绑定任意 ComfyUI `output` 目录
- 默认目录 / 日期归档目录 / 自定义目录并行浏览
- 日期产出工作台
- 模型 / LoRA 高级筛选
- 搜索文件名、Prompt、模型、LoRA、标签、笔记
- PNG 元数据查看与复制
- 收藏夹与收藏分组
- 自动规则引擎
- 提示词提示器

### 本次版本新增重点

- 新增 Windows 安装程序与安装目录选择流程
- 安装版数据统一落在安装目录内，便于迁移和备份
- 安装包内置提示词词库目录
- 补齐 GitHub 发布流程与安装器文档
- 提示词提示器分页支持固定显示
- 修复分页重复省略号
- 修复图片删除后重新生成同名文件时的旧缓存显示问题
- 修复工作台总览里打开提示词提示器无响应的问题
- 强化 ComfyUI Prompt 提取逻辑，适配更复杂的工作流结构
- 新增 Prompt 解析调试视图
- 修复元数据调试卡片中的长文本溢出问题

## 常用入口说明

### 工作台总览

用于查看：

- 最新作品
- 今日新增
- 图片总数
- 存储占用
- 最近更新入口

### 日期产出工作台

适合围绕日期回看近期出图：

- 今天 / 昨天 / 最近 7 天 / 本月
- 自定义日期范围
- 叠加模型筛选
- 叠加 LoRA 筛选
- 点击活跃日期后快速跳回图库

### 图库页

适合进行细看和批量整理：

- 搜索
- 模型 / LoRA / 标签 / 收藏筛选
- 排序与分页
- Lightbox 详情查看

### 提示词提示器

适合做本地 Prompt 工作流：

- 从工具菜单、图库顶部、Lightbox 直接打开
- 搜索提示词词库
- 按来源 / 分类 / 子分类 / 作用域筛选
- 分页浏览词条
- 加入正向或反向 Prompt
- 使用预设词包快速起稿
- 保存组合模板或正向 / 反向模板
- 维护“我的词库”、收藏、最近使用

### 自动规则引擎

适合做自动化整理：

- 按模型自动打标
- 按 LoRA 自动打标
- 按目录归类
- 自动收藏或补充信息

## 使用建议

1. 首次启动后，先绑定真实的 ComfyUI `output` 目录。
2. 回看近期产出时，优先使用“日期产出工作台”。
3. 词条过多时，先用分类筛选，再配合分页浏览，不要直接全量翻。
4. 高频使用的词条建议存入“我的词库”或模板，后续复用更快。
5. 如果某张图的 Prompt 提取不准确，可以在 Lightbox 的“提示词解析调试”里查看命中来源。

## 文档索引

- [项目上下文](./PROJECT_CONTEXT.md)
- [发布说明与流程](./RELEASE.md)
- [Windows 安装器说明](./WINDOWS_INSTALLER.md)
- [v2.1.0 提示词提示器任务文档](./V2.1.0_PROMPT_ASSISTANT_TASK.md)
- [v1.8 功能规划](./V1.8_DATE_MODEL_PLAN.md)
- [v1.8 实现说明](./V1.8_DATE_MODEL_IMPLEMENTATION.md)

## 版本信息

- 当前版本：`v2.1.6`
- 仓库地址：<https://github.com/dcajusteno-ops/qijiu-demo>
- Release：<https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.1.6>

## 维护约定

每次正式发布时，至少同步以下内容：

- 根目录 `desktop-app.exe`
- 根目录 `ComfyManager-amd64-installer.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`
