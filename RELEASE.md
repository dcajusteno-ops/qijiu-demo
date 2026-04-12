# Comfy Manager 版本发布指南

## 前置条件

- [Wails CLI](https://wails.io/) 已安装（`wails` 命令可用）
- [GitHub CLI](https://cli.github.com/) 已安装并登录（`gh auth status` 验证）
- Git 远程仓库已配置（当前：`https://github.com/dcajusteno-ops/qijiu-demo.git`）

---

## 版本号规范

采用 [语义化版本](https://semver.org/lang/zh-CN/)：`MAJOR.MINOR.PATCH`

| 类型 | 含义 | 示例 |
|------|------|------|
| MAJOR | 不兼容的 API 变更 | 1.0.0 → 2.0.0 |
| MINOR | 向后兼容的功能新增 | 1.0.0 → 1.1.0 |
| PATCH | 向后兼容的问题修复 | 1.0.0 → 1.0.1 |

---

## 发布流程

### 1. 修改代码

在 `desktop-source/` 下进行开发，完成后确认功能正常。

### 2. 编译打包

```bash
cd desktop-source
wails build
```

将生成的可执行文件复制到项目根目录：

```bash
cp build/bin/comfy-manager-wails.exe ../desktop-app.exe
```

### 3. 提交代码

```bash
# 查看变更
git status
git diff

# 暂存文件（按需添加，排除 data/ 等运行时文件）
git add <相关文件>

# 提交
git commit -m "feat: 简要描述变更内容"
```

### 4. 打标签

```bash
git tag -a v<版本号> -m "v<版本号> - 版本描述"
```

示例：
```bash
git tag -a v1.1.0 -m "v1.1.0 - 新增XXX功能"
```

### 5. 推送到 GitHub

```bash
# 推送代码
git push origin main

# 推送标签
git push origin v<版本号>
```

### 6. 创建 Release（含可下载的 exe）

```bash
gh release create v<版本号> \
  ./desktop-app.exe#Comfy\ Manager\ v<版本号>\ 桌面端 \
  --title "v<版本号>" \
  --notes "## 更新内容

### 新功能
- ...

### 修复
- ...
"
```

示例：
```bash
gh release create v1.1.0 \
  ./desktop-app.exe#Comfy\ Manager\ v1.1.0\ 桌面端 \
  --title "v1.1.0" \
  --notes "## v1.1.0

### 新功能
- 新增 XXX 功能

### 修复
- 修复 XXX 问题
"
```

---

## .gitignore 说明

以下目录/文件已被排除，不会提交到仓库：

| 排除项 | 原因 |
|--------|------|
| `data/` | 运行时数据，每台机器不同 |
| `.trash/` | 回收站图片文件 |
| `images/` | 项目截图，体积大 |
| `desktop-source/frontend/node_modules/` | 依赖包，可 `npm install` 重建 |
| `desktop-source/frontend/dist/` | 前端构建产物，`wails build` 时重新生成 |
| `desktop-source/build/bin/` | 开发构建产物 |
| `Sanrio Cinnamoroll White Arrow & Head.zip` | 非项目资源 |

`desktop-app.exe` 会通过 **GitHub Release** 附件分发，不再直接提交到仓库代码中（如需调整可修改 `.gitignore`）。

---

## 快速一键发布（参考命令）

```bash
# 设置版本号
VER=v1.1.0

# 编译
cd desktop-source && wails build && cd ..

# 复制 exe
cp desktop-source/build/bin/comfy-manager-wails.exe desktop-app.exe

# 提交
git add -A
git commit -m "release: $VER"
git tag -a $VER -m "$VER"

# 推送
git push origin main
git push origin $VER

# 创建 Release
gh release create $VER \
  ./desktop-app.exe#Comfy\ Manager\ $VER\ 桌面端 \
  --title "$VER" \
  --notes "## $VER 更新内容" \
```

---

## 历史版本

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.1.0 | 2026-04-12 | 智能相册滚动修复、文件变更自动刷新、Ctrl+S保存笔记 |
| v1.0.0 | 2025-04-11 | 初始发布，提示词辅助工具改进 + 自定义链接管理 |
