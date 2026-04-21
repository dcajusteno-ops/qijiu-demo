# Comfy Manager 发布说明

当前版本：`v2.2.1`

远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 标准发布流程

### 1. 同步文档与产物

发布前至少同步这些文件：

- 根目录 `desktop-app.exe`
- 根目录 `ComfyManager-amd64-installer.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

### 2. 构建桌面程序与安装器

在 `desktop-source` 下执行：

```powershell
wails build --nsis
```

### 3. 覆盖根目录产物

```powershell
Copy-Item .\desktop-source\build\bin\desktop-app.exe .\desktop-app.exe -Force
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force
```

### 4. 提交与推送

```powershell
git add -A
git commit -m "release: v2.2.1"
git tag -a v2.2.1 -m "v2.2.1"
git push origin main
git push origin v2.2.1
```

## v2.2.1 发布内容

### 修复

- 修正软件内“使用文档”仍显示 v2.1.0 的版本错位问题
- 同步软件内文档、README、项目上下文和安装器版本号
- 重新打包桌面端与 Windows 安装程序，确保发布产物与文档一致

## v2.2.0 发布内容

### 新增

- 大型图库性能模式
- 目录健康中心
- 缩略图 / 预览图变体缓存
- 自定义目录置顶状态与侧边栏联动

### 优化

- 悬浮式分页
- 手动输入页码
- 自定义目录弹窗布局与编辑区滚动
- 工作台、统计页、日期产出页的刷新链路

### 修复

- 性能模式分页请求乱序覆盖
- 标准 / 性能模式切换后数据不完整
- 删除后分页列表不同步
- 工作台统计与日期口径问题

## 历史版本

| 版本 | 日期 | 说明 |
|---|---|---|
| v2.2.1 | 2026-04-21 | 修正软件内使用文档版本滞后问题，并同步重打包桌面端与安装器 |
| v2.2.0 | 2026-04-21 | 大型图库性能模式、目录健康中心、缩略图变体、自定义目录增强、分页交互升级与多项回归修复 |
| v2.1.6 | 2026-04-19 | Windows 安装版同步、Prompt 提取增强、分页与缓存修复 |
| v2.1.0 | 2026-04-18 | 提示词助手独立页面与本地 Prompt 工作流 |
