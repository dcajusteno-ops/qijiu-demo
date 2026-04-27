# 后端文件地图

最后更新：`2026-04-27`  
适用版本：`v3.0.1`

这份文档说明 `desktop-source/backend/` 下当前 Go 文件的组织方式。

目标很简单：

- 不再让根目录堆满后端逻辑
- 不继续做过深的 package 拆分
- 让维护者光看文件名就能大概知道职责落点

## 1. 顶层入口

### `desktop-source/main.go`

Wails 桌面应用入口。

负责：

- 创建 `backend.App`
- 配置 Wails 启动参数
- 绑定 `OnStartup`、`OnShutdown`
- 绑定图片服务处理器

### `desktop-source/backend/exports.go`

后端对根入口暴露的薄包装层。

负责导出：

- `Startup`
- `Shutdown`
- `ServeImage`

## 2. `backend/app.go`

最小应用壳子。

这里主要保留：

- `App` 结构体
- 共享状态
- watcher、缓存、上下文等跨模块状态字段

原则：

- 不在这里堆业务逻辑
- 新增业务代码优先落到对应 `app_feature_*`

## 3. `backend/app_core_*`

核心运行时文件。

### `app_core_constants.go`

放跨模块常量，例如：

- 目录与路径版本常量
- 资源前缀
- 预览与缩略图规格

### `app_core_lifecycle.go`

放应用生命周期逻辑，例如：

- `NewApp`
- 启动初始化
- 关闭清理

### `app_core_runtime.go`

放运行时辅助能力，例如：

- 剪贴板复制

## 4. `backend/app_feature_*`

业务功能主入口。

### `app_feature_assets.go`

- 预览图 / 缩略图变体生成
- 图片服务响应

### `app_feature_auto_rules.go`

- 自动规则读取、保存
- 自动规则匹配与执行
- 自动规则运行结果汇总

### `app_feature_custom_roots.go`

- 自定义目录读取与维护
- 置顶、排序、启用状态控制

### `app_feature_directory_binding.go`

- output 绑定信息读取
- output 绑定保存
- output 重新切换

### `app_feature_favorites.go`

- 收藏夹
- 收藏分组
- 图片与收藏组的关系维护

### `app_feature_file_ops.go`

- 批量移动
- 导出图片
- 导入图片
- 打开目录 / 打开文件
- 目录整理

### `app_feature_gallery.go`

- 图库读取
- 分页结果生成
- 工作台聚合入口

### `app_feature_health.go`

- 目录健康汇总
- 空目录检查
- 失效标签 / 收藏引用检查与清理

### `app_feature_image_metadata.go`

- 图片元数据提取
- PNG 相关信息读取

### `app_feature_launcher.go`

- 外部工具维护
- 外部工具启动
- 图标提取

### `app_feature_profile.go`

- 用户资料读取与保存
- 头像选择、清除与写入

### `app_feature_prompt.go`

- 提示词库读取
- 自定义词条
- 模板与提示词助手状态

### `app_feature_settings.go`

- 设置项读取与保存
- 回收站、性能模式、工具菜单等设置接口

### `app_feature_statistics.go`

- 统计汇总
- 工作台相关指标入口

### `app_feature_tags.go`

- 标签管理
- 图片标签关系
- 图片笔记

### `app_feature_trash.go`

- 删除进入回收站
- 回收站列表
- 恢复、清空、批量删除

## 5. `backend/app_support_*`

内部支持文件。

### `app_support_helpers.go`

跨功能共享的小型 helper。

### `app_support_legacy.go`

历史兼容与迁移辅助：

- 旧路径兼容
- 旧数据迁移

### `app_support_metadata_cache.go`

- 元数据缓存读写
- 缓存替换与快照
- 预览缓存清理

### `app_support_paths.go`

- 路径拼装
- 目录校验
- 资源位置解析
- 运行数据文件路径

### `app_support_prompt_helpers.go`

- 提示词相关内部辅助逻辑

### `app_support_settings_helpers.go`

- 设置默认值和设置辅助逻辑

### `app_support_watcher.go`

- watcher 建立与停止
- 文件变化刷新节流
- 自动规则进度事件

## 6. `backend/app_types_*`

按领域分组的类型定义。

### `app_types_gallery.go`

- 图片、图库、分页、工作台相关类型

### `app_types_library.go`

- 共享库、收藏、标签等资料型结构

### `app_types_prompt.go`

- 提示词、模板、助手状态等类型

### `app_types_rules.go`

- 自动规则条件、动作、结果等类型

### `app_types_settings.go`

- 设置、快捷键、性能模式、工具菜单等类型

## 7. 当前维护规则

- 根目录保持轻量，只放入口层
- 后端主体统一留在 `desktop-source/backend/`
- 新增功能优先归入现有分组，不继续深拆 package
- `app.go` 继续保持“壳子化”
- 如果功能是对前端直接暴露的业务接口，优先归到 `app_feature_*`

