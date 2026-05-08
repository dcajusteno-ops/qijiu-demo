# Comfy Manager 发布说明

当前版本：`v3.1`

远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 1. v3.1 发布重点

`v3.1` 是在 `v3.0.1` 基础上的维护发布，重点是把灵动图库小窗体验、文档版本和安装器发布链路一次同步。

### 本次核心变更

- 后端源码整理进入 `desktop-source/backend/`
- 后端按 `app_core_* / app_feature_* / app_support_* / app_types_*` 完成归类
- 修复前端对 Wails 绑定名的兼容问题
- 修复首页最新作品区域异常“路”字文案
- 重写根 README、`docs/README.md`、项目上下文、重构计划和软件内使用文档
- 修复“日期归档目录”进入后黑屏问题
- 修复日期归档树中年份分类无法正常折叠的问题
- 新增灵动图库小窗：支持置顶、恢复主窗口、刷新、搜索、目录切换、图片详情和批量操作
- 优化小窗窗口模式布局，将筛选、排序、Output、清缓存等低频功能收进工具面板
- 小窗图片区新增分页，支持每页 60 / 120 / 240 张
- 小窗批量删除与清缓存确认改为应用内暗色确认弹层
- 软件内使用文档版本同步升级到 `v3.1`
- 发布前校验根目录 exe 与安装包和 `desktop-source/build/bin/` 哈希一致

## 2. 标准发布流程

### 2.1 同步文档

正式发布前至少同步这些文件：

- 根目录 `README.md`
- `docs/README.md`
- `docs/PROJECT_CONTEXT.md`
- `docs/BACKEND_FILE_MAP.md`
- `docs/REFACTOR_PLAN.md`
- `docs/RELEASE.md`
- `docs/WINDOWS_INSTALLER.md`
- 软件内 `frontend/src/components/Documentation.vue`

### 2.2 构建桌面程序

在 `desktop-source` 下执行：

```powershell
wails build -clean
```

### 2.3 构建安装程序

确保系统可用 `makensis` 后执行：

```powershell
wails build -clean -nsis
```

### 2.4 覆盖根目录发布产物

```powershell
Copy-Item .\desktop-source\build\bin\desktop-app.exe .\desktop-app.exe -Force
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force
```

### 2.5 Git 发布步骤

```powershell
git add -A
git commit -m "release: v3.1"
git tag -a v3.1 -m "v3.1"
git push origin main
git push origin v3.1
```

## 3. v3.1 验收清单

- 软件内使用文档版本显示为 `v3.1`
- GitHub README 版本显示为 `v3.1`
- 根目录 `desktop-app.exe` 已覆盖最新构建
- 根目录 `ComfyManager-amd64-installer.exe` 已覆盖最新安装包
- `wails build -clean` 成功
- `wails build -clean -nsis` 成功
- 根目录发布产物与 `desktop-source/build/bin/` 哈希一致
- Git tag `v3.1` 已推送

## 4. 历史版本

| 版本 | 日期 | 说明 |
|---|---|---|
| v3.1 | 2026-05-08 | 灵动图库小窗、分页、应用内确认弹层、安装程序与文档同步更新 |
| v3.0.1 | 2026-04-27 | 日期归档黑屏修复、归档树折叠修复、根目录发布产物哈希校验、版本同步 |
| v3.0.0 | 2026-04-21 | 后端目录整理、文档全面重写、软件内文档升级、安装链恢复、发布收敛 |
| v2.2.1 | 2026-04-21 | 软件内使用文档与外部文档版本同步、重新打包桌面端和安装包 |
| v2.2.0 | 2026-04-21 | 大型图库性能模式、目录健康中心、预览变体、自定义目录增强 |
| v2.1.6 | 2026-04-19 | Windows 安装版同步、Prompt 增强、分页与缓存修复 |

