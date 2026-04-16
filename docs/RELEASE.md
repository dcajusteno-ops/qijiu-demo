# Comfy Manager 发布指南

## 前置条件

- 已安装 [Wails CLI](https://wails.io/)
- 已安装 [GitHub CLI](https://cli.github.com/)
- `gh auth status` 可通过
- Git 远程仓库已配置

当前远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 版本规则

采用语义化版本：

- `MAJOR`：不兼容改动
- `MINOR`：向后兼容的新功能
- `PATCH`：向后兼容的问题修复

当前最新版本：`v1.8.1`

## 标准发布流程

### 1. 完成功能开发

在 `desktop-source/` 下完成前后端功能，并同步更新：

- `desktop-app.exe`
- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

### 2. 构建桌面端

```bash
cd desktop-source
wails build -clean
```

### 3. 覆盖根目录 exe

```bash
copy desktop-source\build\bin\desktop-app.exe desktop-app.exe
```

### 4. 提交代码

```bash
git add -A
git commit -m "release: v1.8.1"
```

### 5. 打版本标签

```bash
git tag -a v1.8.1 -m "v1.8.1"
```

### 6. 推送到 GitHub

```bash
git push origin main
git push origin v1.8.1
```

### 7. 创建 GitHub Release

```bash
gh release create v1.8.1 ^
  ./desktop-app.exe#Comfy^ Manager^ v1.8.1^ 桌面端 ^
  --title "v1.8.1" ^
  --notes "## v1.8.1 更新内容"
```

## v1.8.1 发布说明

Release 地址：

- [v1.8.1](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.8.1)

本次版本重点：

- 修复 ComfyUI 新出图后多个页面没有自动刷新的问题
- 去掉前端轮询，改为基于 `images:changed` 的事件驱动刷新
- 更新内置使用文档页，补齐完整中文说明
- 持续稳定日期产出工作台与模型 / LoRA 筛选体验

## 快速发布命令模板

```bash
set VER=v1.8.1

cd desktop-source
wails build -clean
cd ..

copy desktop-source\build\bin\desktop-app.exe desktop-app.exe

git add -A
git commit -m "release: %VER%"
git tag -a %VER% -m "%VER%"

git push origin main
git push origin %VER%

gh release create %VER% ^
  ./desktop-app.exe#Comfy^ Manager^ %VER%^ 桌面端 ^
  --title "%VER%" ^
  --notes "## %VER% 更新内容"
```

## 历史版本

| 版本 | 日期 | 说明 |
|---|---|---|
| v1.8.1 | 2026-04-16 | 自动刷新修复、去轮询、中文文档页更新 |
| v1.8.0 | 2026-04-16 | 日期产出工作台、模型 / LoRA 筛选 |
| v1.7.0 | 2026-04-16 | 搜索 MVP、自动规则引擎、个人中心规则入口 |
| v1.6.0 | 2026-04-15 | 个人中心与数据视界细化 |
| v1.5.0 | 2026-04-15 | 数据视界与全局快捷键 |
| v1.4.1 | 2026-04-14 | 统一发布产物名为 `desktop-app.exe` |
