# Windows 安装器说明

当前版本：`v3.0.1`

## 1. 安装器目标

本项目提供标准 Windows 安装流程，目标是：

- 用户能选择安装目录
- 程序文件和运行数据保持在同一安装目录附近
- 安装完成后可直接从桌面或开始菜单启动

## 2. 默认安装布局

安装完成后，典型目录结构如下：

```text
H:\Comfy Manager\
├─ desktop-app.exe
├─ data\
└─ .trash\
```

说明：

- 程序主文件：`desktop-app.exe`
- 运行数据：`data\`
- 回收站：`.trash\`

## 3. 如何构建安装器

在 `desktop-source/` 目录执行：

```powershell
wails build --nsis -clean
```

前提：

- 系统已安装 NSIS
- `makensis` 可在命令行中调用

构建成功后，输出位于：

```text
desktop-source\build\bin\ComfyManager-amd64-installer.exe
```

## 4. 发布时如何覆盖根目录安装器

```powershell
Copy-Item .\desktop-source\build\bin\ComfyManager-amd64-installer.exe .\ComfyManager-amd64-installer.exe -Force
```

## 5. 安装器行为

当前安装器会：

- 显示欢迎页
- 允许用户选择安装目录
- 自动安装或检测 WebView2 运行时
- 复制程序文件
- 复制 `data\prompt-library\` 到安装目录
- 创建开始菜单快捷方式
- 创建桌面快捷方式

## 6. 当前版本注意事项

- `v3.0.1` 已重新验证安装器生成链，并确认根目录安装包与 `build/bin` 产物一致
- 发布前必须确保根目录安装包是最新构建结果
- 如果安装器构建失败，优先检查 `makensis` 是否可用

