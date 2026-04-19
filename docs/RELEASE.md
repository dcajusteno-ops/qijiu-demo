# Comfy Manager 发布指南

当前版本：`v2.1.6`

远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 版本规则

采用语义化版本：

- `MAJOR`：不兼容变更或大的结构升级
- `MINOR`：向后兼容的新功能
- `PATCH`：向后兼容的问题修复

## 标准发布流程

### 1. 完成功能与文档同步

确保以下内容已同步：

- 根目录 `desktop-app.exe`
- 根目录 `ComfyManager-amd64-installer.exe`
- 根目录 `README.md`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

### 2. 构建桌面端

在 `desktop-source` 下执行：

```powershell
wails build --nsis
```

### 3. 覆盖根目录产物

```powershell
Copy-Item .\desktop-source\build\bin\desktop-app.exe .\desktop-app.exe -Force
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force
```

### 4. 提交代码

```powershell
git add -A
git commit -m "release: v2.1.6"
```

### 5. 处理版本标签

如果当前版本标签尚不存在：

```powershell
git tag -a v2.1.6 -m "v2.1.6"
```

如果 `v2.1.6` 已经存在，不要默认强推覆盖标签；优先继续推送 `main`，必要时明确创建新 patch 版本。

### 6. 推送到 GitHub

```powershell
git push origin main
git push origin v2.1.6
```

如果标签已存在于远端且无需改写，只推送 `main` 即可。

### 7. 创建 GitHub Release

```powershell
gh release create v2.1.6 `
  ./desktop-app.exe#Comfy Manager v2.1.6 桌面端 `
  ./ComfyManager-amd64-installer.exe#Comfy Manager v2.1.6 Windows 安装程序 `
  --title "v2.1.6" `
  --notes "## v2.1.6 更新内容"
```

## v2.1.6 发布说明

Release 地址：

- [v2.1.6](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.1.6)

本次版本重点：

- 新增 Windows 安装程序，支持安装目录选择
- 安装包内置 `data/prompt-library/`
- 安装版运行时数据统一落在安装目录内
- 提示词提示器分页固定显示与重复省略号修复
- 图片删除后重新生成同名文件时的旧缓存显示修复
- 工作台总览打开提示词提示器无响应修复
- ComfyUI Prompt 提取逻辑增强，支持更复杂工作流
- 新增 Prompt 解析调试视图
- 修复调试面板中的长文本溢出问题

## 快速发布命令模板

```powershell
$VER = "v2.1.6"

cd desktop-source
wails build --nsis
cd ..

Copy-Item .\desktop-source\build\bin\desktop-app.exe .\desktop-app.exe -Force
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force

git add -A
git commit -m "release: $VER"
git push origin main
```

如需首次创建标签，再追加：

```powershell
git tag -a $VER -m $VER
git push origin $VER
```

## 历史版本

| 版本 | 日期 | 说明 |
|---|---|---|
| v2.1.6 | 2026-04-19 | 基于 v2.1.5 的正式补发版本，包含安装版同步、Prompt 提取增强、调试视图、缓存修复、分页与布局修复 |
| v2.1.5 | 2026-04-19 | 安装版同步、Prompt 提取增强、调试视图、缓存修复、分页与布局修复 |
| v2.1.0 | 2026-04-18 | 提示词提示器、自定义提示词、模板复用、文档同步 |
| v2.0.1 | 2026-04-17 | 内置文档页更新、乱码修复、文档同步 |
| v2.0.0 | 2026-04-17 | 目录绑定升级、设置中心、工具菜单配置、日期筛选 |
| v1.8.1 | 2026-04-16 | 自动刷新修复、中文文档更新 |
| v1.8.0 | 2026-04-16 | 日期产出工作台、模型 / LoRA 筛选 |
