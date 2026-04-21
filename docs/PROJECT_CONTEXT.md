# Comfy Manager 项目上下文

当前稳定版本：`v2.2.1`  
更新时间：`2026-04-21`

## 1. 项目定位

Comfy Manager（灵动图库）是一套基于 **Wails v2 + Go + Vue 3** 的桌面图库与创作管理工具，服务于 ComfyUI 出图后的浏览、筛选、整理、归档与 Prompt 复用场景。

当前项目主线：

- 浏览 ComfyUI 输出图片和 PNG 元数据
- 绑定任意 ComfyUI `output` 目录，而不是强依赖 exe 所在位置
- 按日期、模型、LoRA、标签、收藏、笔记等维度筛图
- 提供日期工作台、统计视图和提示词助手
- 提供自动规则、目录治理、缓存治理等整理能力

## 2. v2.2.1 版本重点

### 本次新增

- 内置“使用文档”版本与内容同步到 v2.2.1
- 大型图库性能模式
- 目录健康中心
- 缩略图 / 预览图变体缓存
- 自定义目录置顶状态与侧边栏联动

### 本次重点修复

- 日期工作台、统计页、工作台总览的数据口径问题
- 性能模式分页请求乱序覆盖
- 性能模式与标准模式切换后的数据完整性问题
- 删除后分页列表不同步
- 自定义目录弹窗布局与编辑区滚动问题

## 3. 技术栈

### 后端

- Go
- Wails v2
- fsnotify
- golang.org/x/image
- google/uuid

### 前端

- Vue 3（Composition API）
- Vite
- Tailwind CSS
- shadcn-vue
- lucide-vue-next
- vue-sonner

## 4. 关键目录

```text
comfy-manager/
├─ README.md
├─ docs/
├─ data/
├─ .trash/
├─ desktop-app.exe
├─ ComfyManager-amd64-installer.exe
└─ desktop-source/
   ├─ app.go
   ├─ main.go
   ├─ shortcuts.go
   ├─ wails.json
   ├─ build/
   └─ frontend/
      └─ src/
```

## 5. 后端核心结构

主文件：`desktop-source/app.go`

当前主要职责：

- 扫描和筛选受管图片
- 解析 PNG 元数据、模型、LoRA、Workflow
- 维护图片元数据缓存
- 维护缩略图 / 预览图变体缓存
- 维护目录绑定、自定义目录、收藏夹、标签、笔记、自动规则
- 提供统计、工作台、目录健康和清理接口

## 6. 前端核心结构

### `frontend/src/App.vue`

- 根级页面装配
- 视图切换
- 图库与工作台联动
- 刷新链路协调

### `frontend/src/composables/useImages.js`

- 图片、收藏、标签、笔记、自定义目录状态
- 搜索、排序、分页
- 日期工作台筛选状态
- 目录树构建
- 性能模式与标准模式切换

### `frontend/src/components/ImageGallery.vue`

- 图库视图
- 悬浮分页
- 性能模式状态提示

### `frontend/src/components/CustomRootDialog.vue`

- 自定义目录新增、排序、置顶、显示控制
- 编辑区状态管理

## 7. 数据持久化

常见数据文件：

- `favorites.json`
- `tags.json`
- `image-tags.json`
- `image-notes.json`
- `custom-roots.json`
- `settings.json`
- `auto-rules.json`
- `trash-metadata.json`
- `image-meta-cache.json`

## 8. 关键业务链路

### 图库刷新

`fsnotify` -> `images:changed` -> 前端订阅 -> 图库 / 工作台 / 统计刷新

### 性能模式

目录摘要 / 轻量索引 -> 分页接口 -> 缩略图或预览图变体 -> Lightbox 原图回落

### 日期工作台

后端聚合接口 -> 日期摘要 / 活跃日期 -> 打开图库继续筛图

## 9. 开发注意事项

- 涉及筛选、分页、目录树时，优先检查 `useImages.js`
- 涉及自定义目录顺序和置顶时，优先检查 `app.go` 与 `CustomRootDialog.vue`
- 涉及性能模式时，注意轻量索引和完整图片数据的切换边界
- 发布时必须同步根目录产物和 `docs`

## 10. 最近变更记录

### 2026-04-21 | v2.2.1

- 修正软件内使用文档仍停留在 v2.1.0 的问题
- 同步更新 README、发布说明与 Windows 安装器版本号
- 重新构建并发布 v2.2.1 桌面端与安装程序

### 2026-04-21 | v2.2.0

- 发布大型图库性能模式与目录健康中心
- 接入缩略图 / 预览图缓存变体
- 升级分页交互与页码直跳
- 增强自定义目录管理交互
- 修复多项性能模式与工作台回归问题

### 2026-04-19 | v2.1.6

- 同步 Windows 安装版
- 增强 Prompt 提取和调试视图
- 修复分页、缓存和布局相关问题
