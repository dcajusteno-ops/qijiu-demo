# Comfy Manager 项目上下文

当前版本：`v3.0.0`  
更新时间：`2026-04-21`

## 1. 项目定位

Comfy Manager 是一个基于 **Wails v2 + Go + Vue 3** 的桌面应用，面向 ComfyUI 出图后的本地整理场景。

当前主线很明确：

- 浏览 ComfyUI output 图片
- 按日期、模型、LoRA、标签、收藏和笔记回看作品
- 管理目录、回收站、缓存和自定义目录
- 提供提示词助手、模板与自动规则，提升复用效率

这不是云相册，也不是远程协作系统。  
它当前的产品重心仍然是：

`ComfyUI output 的本地整理器`

## 2. 技术栈

### 后端

- Go
- Wails v2
- fsnotify
- golang.org/x/image
- github.com/google/uuid

### 前端

- Vue 3
- Vite
- Tailwind CSS
- shadcn-vue
- lucide-vue-next
- vue-sonner

## 3. 关键目录

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

## 4. 当前后端结构

当前后端已经不再把主要逻辑堆在根目录 `app.go`。

新的结构是：

- `desktop-source/main.go`
  Wails 入口，只负责启动应用和绑定 `backend.App`
- `desktop-source/backend/`
  后端主要实现目录
- `desktop-source/backend/exports.go`
  为根入口暴露启动、关闭和图片服务包装

后端分组规则：

- `app.go`
  `App` 壳子与共享状态
- `app_core_*`
  生命周期、运行时、常量
- `app_feature_*`
  业务功能实现
- `app_support_*`
  内部辅助与基础设施
- `app_types_*`
  类型定义

详细映射见：[BACKEND_FILE_MAP.md](./BACKEND_FILE_MAP.md)

## 5. 当前前端结构

### 根级页面

- `frontend/src/App.vue`
  根级页面装配、视图切换、绑定刷新链路

### 核心页面 / 组件

- `frontend/src/components/Home.vue`
  工作台总览
- `frontend/src/components/ImageGallery.vue`
  图库主视图
- `frontend/src/components/DateWorkbench.vue`
  日期产出工作台
- `frontend/src/components/StatisticsDashboard.vue`
  数据视界
- `frontend/src/components/Documentation.vue`
  软件内使用文档
- `frontend/src/components/PromptAssistantPage.vue`
  提示词助手
- `frontend/src/components/AutoRulesPanel.vue`
  自动规则引擎

### 当前已拆分的 composables

- `useImages.js`
  主图库状态入口
- `useGalleryData.js`
  图库数据与分页请求
- `useGalleryHelpers.js`
  辅助格式化与筛选工具
- `useLibraryMeta.js`
  标签、收藏、笔记等资料元数据
- `useWorkbenchFilters.js`
  工作台日期 / 模型 / LoRA 筛选

## 6. 数据持久化

常见持久化文件位于 `data/`：

- `favorites.json`
- `tags.json`
- `image-tags.json`
- `image-notes.json`
- `custom-roots.json`
- `settings.json`
- `auto-rules.json`
- `trash-metadata.json`
- `image-meta-cache.json`
- `prompt-library/`

## 7. 关键业务链路

### 图片刷新链

`fsnotify` -> `images:changed` 事件 -> 前端订阅 -> 图库 / 工作台 / 统计刷新

### output 绑定链

用户选择 output -> 后端校验目录 -> 保存 settings -> 重启 watcher -> 前端刷新当前视图

### 大图库性能链

目录扫描 -> 轻量分页 / 预览变体 -> 延迟元数据读取 -> Lightbox 查看原图与细节

### 提示词复用链

本地词库 -> 搜索 / 分类 / 模板 -> 拼装 Prompt -> 回写使用上下文

## 8. v3.0.0 当前状态总结

`v3.0.0` 的主要成果：

- 后端源码收纳进 `desktop-source/backend/`
- 前端与后端绑定兼容问题修复
- 首页异常文案问题修复
- 软件内文档、GitHub README 与发布说明重写
- Windows 安装包构建链恢复
- 根目录 exe 和安装包重新打包同步

## 9. 当前边界与原则

- 当前阶段不继续做更深的 Go 包拆分
- 以后新增后端能力，优先继续放进 `backend/` 现有分组
- 优先维持外部行为稳定，再做结构优化
- 发布时必须同时同步：
  - 根目录 exe
  - 安装程序
  - 根 README
  - `docs/README.md`
  - `docs/PROJECT_CONTEXT.md`
  - `docs/RELEASE.md`

## 10. 下一步建议

若继续推进，优先级建议如下：

1. 清理剩余编码历史包袱，统一前端中文文案
2. 继续收敛前端大组件的职责边界
3. 检查设置中心与工具菜单的统一入口体验
4. 评估是否需要补一份图库 / 工作台数据流图

