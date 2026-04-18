# Windows Installer Notes

Current version: `v2.1.5`

## Goal

This project now supports a more standard Windows installation flow:

- The installer shows a directory selection page with a default path of `%USERPROFILE%\Comfy Manager`
- Program files and runtime data stay together under the selected install folder
- The app writes its `data` folder and trash folder inside the install directory

## Installed Layout

After installation, the app uses the selected install directory for everything, for example:

- `H:\Comfy Manager\desktop-app.exe`
- `H:\Comfy Manager\data\`
- `H:\Comfy Manager\.trash\`

No separate AppData or ProgramData storage is created by the installer.

## Build The Installer

From `desktop-source`:

```powershell
wails build --nsis
```

Expected output:

```text
desktop-source\build\bin\ComfyManager-amd64-installer.exe
```

## Installer Behavior

The NSIS installer now:

- shows a normal install-directory page and defaults to `%USERPROFILE%\Comfy Manager`
- creates Start Menu and desktop shortcuts
- bundles `data\prompt-library\` into the install directory
- keeps runtime data inside the chosen install directory
