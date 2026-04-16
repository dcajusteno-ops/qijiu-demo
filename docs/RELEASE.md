# Comfy Manager 版本发布指南

## 前置条件

- [Wails CLI](https://wails.io/) 已安装，`wails` 命令可用
- [GitHub CLI](https://cli.github.com/) 已安装并登录，执行 `gh auth status` 可通过
- Git 远程仓库已配置（当前：`https://github.com/dcajusteno-ops/qijiu-demo.git`）

## 版本号规范

采用 [语义化版本](https://semver.org/lang/zh-CN/)：`MAJOR.MINOR.PATCH`

| 类型 | 含义 | 示例 |
|------|------|------|
| MAJOR | 不兼容的 API 变更 | `1.0.0 -> 2.0.0` |
| MINOR | 向后兼容的功能新增 | `1.6.0 -> 1.7.0` |
| PATCH | 向后兼容的问题修复 | `1.7.0 -> 1.7.1` |

## 发布流程

### 1. 修改代码

在 `desktop-source/` 下完成开发，并确认桌面端关键功能可用。

### 2. 编译打包

```bash
cd desktop-source
wails build -clean
```

### 3. 更新项目根目录 exe

将编译产物复制到项目根目录，确保仓库中的 `desktop-app.exe` 与当前版本一致：

```bash
copy desktop-source\\build\\bin\\desktop-app.exe desktop-app.exe
```

> 每次发版都必须更新并提交 `desktop-app.exe`。

当前最新 Release：[`v1.7.0`](https://github.com/dcajusteno-ops/qijiu-demo/releases/tag/v1.7.0)

### 4. 更新文档

至少同步以下文档：

- `docs/README.md`
- `docs/RELEASE.md`
- `docs/PROJECT_CONTEXT.md`

建议同步内容：

- 当前稳定版本号
- 本次新增功能 / 删除功能 / 修复项
- GitHub Release 链接
- 历史版本表

### 5. 提交代码

```bash
git status
git add -A
git commit -m "release: v1.7.0"
```

### 6. 打标签

```bash
git tag -a v1.7.0 -m "v1.7.0"
```

### 7. 推送到 GitHub

```bash
git push origin main
git push origin v1.7.0
```

### 8. 创建 GitHub Release

```bash
gh release create v1.7.0 ^
  ./desktop-app.exe#Comfy^ Manager^ v1.7.0^ 桌面端 ^
  --title "v1.7.0" ^
  --notes "## v1.7.0

### 新功能
- 新增搜索 MVP，支持按文件名、Prompt、模型、LoRA、采样器、Negative、尺寸、标签、笔记搜索
- 新增自动规则引擎，支持自动打标、加入收藏分组、移动目录
- 自动规则整合进个人中心，并支持手动执行进度反馈

### 调整
- 删除智能相册模块，统一收敛到搜索与自动规则链路

### 修复
- 修复个人中心头像路径不在 rootDir 时无法显示的问题
- 修复搜索未命中 searchText 导致 LoRA 等元数据无法被搜到的问题
"
```

## .gitignore 说明

以下目录或文件默认不提交：

| 排除项 | 原因 |
|--------|------|
| `data/` | 运行时数据，每台机器不同 |
| `.trash/` | 回收站图片文件 |
| `images/` | 项目截图，体积大 |
| `desktop-source/frontend/node_modules/` | 依赖包，可重建 |
| `desktop-source/frontend/dist/` | 前端构建产物 |
| `desktop-source/build/bin/` | 本地构建产物 |
| `Sanrio Cinnamoroll White Arrow & Head.zip` | 非项目资源 |

`desktop-app.exe` 需要随版本一起提交，并通过 GitHub Release 作为附件分发。

## 快速发布命令

```bash
set VER=v1.7.0

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
|------|------|------|
| v1.7.0 | 2026-04-16 | 新增搜索 MVP 与自动规则引擎；自动规则整合进个人中心并支持默认规则、执行进度；删除智能相册；修复头像资源路径与搜索命中链路 |
| v1.6.0 | 2026-04-15 | 新增简约风“个人中心”，支持资料编辑、默认进入页、单张展示图与创作摘要；优化“数据视界”节点详情与滚动布局 |
| v1.5.0 | 2026-04-15 | 新增“数据视界”生成历史时间线、趋势图与活跃热图；新增全局快捷键设置与系统级视图切换 |
| v1.4.1 | 2026-04-14 | 修复提示词模板弹窗交互与边框显示，统一 Wails 发布产物文件名为 `desktop-app.exe` |
| v1.4.0 | 2026-04-14 | 修复智能筛选路径叠加问题，改进侧边栏导航逻辑 |
| v1.3.0 | 2026-04-13 | 修复智能筛选不自动刷新，新增按日期整理文件、导出支持移动模式、新增图片上传 |
| v1.2.0 | 2026-04-12 | 新增提示词模板库 |
| v1.1.0 | 2026-04-12 | 修复文件变更自动刷新、Ctrl+S 保存笔记 |
| v1.0.0 | 2025-04-11 | 初始发布 |
