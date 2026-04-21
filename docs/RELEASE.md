# Comfy Manager 发布说明

当前版本：`v3.0.0`

远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 1. v3.0.0 发布重点

`v3.0.0` 是一次完整的结构与发布收敛版本，重点不在单个新功能，而在于让项目进入更稳定、更易维护、更适合继续开发的状态。

### 本次核心变更

- 后端源码整理进入 `desktop-source/backend/`
- 后端按 `app_core_* / app_feature_* / app_support_* / app_types_*` 完成归类
- 修复前端对 Wails 绑定名的兼容问题
- 修复首页最新作品区域异常“路”字文案
- 重写根 README、`docs/README.md`、项目上下文、重构计划和软件内使用文档
- 软件内使用文档版本同步升级到 `v3.0.0`
- Windows 安装包重新打通并重建

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
wails build --nsis -clean
```

### 2.4 覆盖根目录发布产物

```powershell
Copy-Item .\desktop-source\build\bin\desktop-app.exe .\desktop-app.exe -Force
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force
```

### 2.5 Git 发布步骤

```powershell
git add -A
git commit -m "release: v3.0.0"
git tag -a v3.0.0 -m "v3.0.0"
git push origin main
git push origin v3.0.0
```

## 3. v3.0.0 验收清单

- 软件内使用文档版本显示为 `v3.0.0`
- GitHub README 版本显示为 `v3.0.0`
- 根目录 `desktop-app.exe` 已覆盖最新构建
- 根目录 `ComfyManager-amd64-installer.exe` 已覆盖最新安装包
- `wails build -clean` 成功
- `wails build --nsis -clean` 成功
- Git tag `v3.0.0` 已推送

## 4. 历史版本

| 版本 | 日期 | 说明 |
|---|---|---|
| v3.0.0 | 2026-04-21 | 后端目录整理、文档全面重写、软件内文档升级、安装链恢复、发布收敛 |
| v2.2.1 | 2026-04-21 | 软件内使用文档与外部文档版本同步、重新打包桌面端和安装包 |
| v2.2.0 | 2026-04-21 | 大型图库性能模式、目录健康中心、预览变体、自定义目录增强 |
| v2.1.6 | 2026-04-19 | Windows 安装版同步、Prompt 增强、分页与缓存修复 |

