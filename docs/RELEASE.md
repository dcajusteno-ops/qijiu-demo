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

当前最新版本：`v1.8.0`

## 标准发布流程

### 1. 完成功能开发

在 `desktop-source/` 下完成前后端功能，必要时同步更新：

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
copy desktop-source\\build\\bin\\desktop-app.exe desktop-app.exe
```

### 4. 提交代码

```bash
git add -A
git commit -m "release: v1.8.0"
```

### 5. 打版本标签

```bash
git tag -a v1.8.0 -m "v1.8.0"
```

### 6. 推送到 GitHub

```bash
git push origin main
git push origin v1.8.0
```

### 7. 创建 GitHub Release

```bash
gh release create v1.8.0 ^
  ./desktop-app.exe#Comfy^ Manager^ v1.8.0^ 桌面端 ^
  --title "v1.8.0" ^
  --notes "## v1.8.0 更新内容"
```

## v1.8.0 发布说明

Release 地址：

- [v1.8.0](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.8.0)

本次版本重点：

- 新增日期产出工作台
- 新增模型 / LoRA 筛选
- 新增日期统计与最近活跃日期入口
- 优化侧边栏顶部入口结构
- 修复模型筛选归一化兼容与旧筛选残留问题

## 快速发布命令模板

```bash
set VER=v1.8.0

cd desktop-source
wails build -clean
cd ..

copy desktop-source\\build\\bin\\desktop-app.exe desktop-app.exe

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
| v1.8.0 | 2026-04-16 | 日期产出工作台、模型 / LoRA 筛选、筛选兼容修复 |
| v1.7.0 | 2026-04-16 | 搜索 MVP、自动规则引擎、个人中心规则入口 |
| v1.6.0 | 2026-04-15 | 个人中心与数据视界细化 |
| v1.5.0 | 2026-04-15 | 数据视界与全局快捷键 |
| v1.4.1 | 2026-04-14 | 统一发布产物名为 `desktop-app.exe` |
