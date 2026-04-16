# Comfy Manager 文档总览

当前稳定版本：`v1.8.1`

## 项目简介

Comfy Manager（灵动图库）是一个面向 ComfyUI 输出目录的桌面图片管理器，基于 **Wails v2 + Go + Vue 3** 构建。

它的核心用途是：

- 浏览和整理 ComfyUI 生成图片
- 通过日期快速回看最近产出
- 通过模型 / LoRA 快速筛选结果
- 管理标签、收藏、笔记和自动规则

## v1.8.1 更新概览

### 核心功能

- 日期产出工作台
- 模型 / LoRA 筛选
- 日期统计卡片与最近活跃日期入口
- 图库、收藏夹、数据视界、个人中心

### 本次补丁修复

- 修复 ComfyUI 新出图后多个页面不自动刷新的问题
- 去掉前端轮询，改成事件驱动刷新，降低性能占用
- 更新内置“使用文档”页面为完整中文内容
- 持续增强模型 / LoRA 筛选稳定性

## 文档索引

- [项目上下文](./PROJECT_CONTEXT.md)
- [发布指南](./RELEASE.md)
- [v1.8 功能规划](./V1.8_DATE_MODEL_PLAN.md)
- [v1.8 实现说明](./V1.8_DATE_MODEL_IMPLEMENTATION.md)

## 版本信息

- 当前版本：`v1.8.1`
- Release：<https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.8.1>
- 仓库地址：<https://github.com/dcajusteno-ops/qijiu-demo>

## 当前推荐使用方式

1. 用 ComfyUI 正常出图。
2. 打开 Comfy Manager 查看最新图片。
3. 在“日期产出”中按日期、模型、LoRA 组合筛选。
4. 回到图库继续浏览、收藏、打标或批量处理。

## 维护约定

每次正式发布版本时，请至少同步以下内容：

- 更新根目录 `desktop-app.exe`
- 更新 `docs/README.md`
- 更新 `docs/RELEASE.md`
- 更新 `docs/PROJECT_CONTEXT.md`
