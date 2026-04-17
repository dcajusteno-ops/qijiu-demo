# Comfy Manager 发布指南

## 前置条件

- 已安装 [Wails CLI](https://wails.io/)
- 已安装 [GitHub CLI](https://cli.github.com/)
- `gh auth status` 可通过
- 远程仓库已配置

当前远程仓库：

- `origin`: `https://github.com/dcajusteno-ops/qijiu-demo.git`

## 版本规则

采用语义化版本：

- `MAJOR`：不兼容变更或大的结构升级
- `MINOR`：向后兼容的新功能
- `PATCH`：向后兼容的问题修复

当前最新版本：`v2.0.1`

## 标准发布流程

### 1. 完成功能开发

确保以下内容已同步：

- 根目录 `desktop-app.exe`
- 根目录 `README.md`
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
git commit -m "release: v2.0.1"
```

### 5. 打标签

```bash
git tag -a v2.0.1 -m "v2.0.1"
```

### 6. 推送到 GitHub

```bash
git push origin main
git push origin v2.0.1
```

### 7. 创建 GitHub Release

```bash
gh release create v2.0.1 ^
  ./desktop-app.exe#Comfy^ Manager^ v2.0.1^ 桌面端 ^
  --title "v2.0.1" ^
  --notes "## v2.0.1 更新内容"
```

## v2.0.1 发布说明

Release 地址：

- [v2.0.1](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v2.0.1)

本次版本重点：

- 更新软件内置“使用文档”页面
- 修复内置文档页中的中文乱码与旧版本残留
- 同步更新根 README 和 docs 文档

## 快速发布命令模板

```bash
set VER=v2.0.1

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
| v2.0.1 | 2026-04-17 | 内置文档页更新、乱码修复、文档同步 |
| v2.0.0 | 2026-04-17 | 目录绑定升级、设置中心、工具菜单配置、日期范围筛选 |
| v1.8.1 | 2026-04-16 | 自动刷新修复、去轮询、中文文档页更新 |
| v1.8.0 | 2026-04-16 | 日期产出工作台、模型 / LoRA 筛选 |
| v1.7.0 | 2026-04-16 | 搜索 MVP、自动规则引擎、个人中心规则入口 |
| v1.6.0 | 2026-04-15 | 个人中心与数据视界细化 |
