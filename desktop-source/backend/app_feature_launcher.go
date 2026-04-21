package backend

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (a *App) loadLauncherTools() ([]LauncherTool, error) {
	var tools []LauncherTool
	data, err := os.ReadFile(a.launcherToolsFile())
	if err != nil {
		return []LauncherTool{}, nil
	}
	json.Unmarshal(data, &tools)
	return tools, nil
}

func (a *App) saveLauncherTools(tools []LauncherTool) error {
	data, _ := json.MarshalIndent(tools, "", "  ")
	return os.WriteFile(a.launcherToolsFile(), data, 0644)
}

func (a *App) GetLauncherTools() ([]LauncherTool, error) {
	return a.loadLauncherTools()
}

func (a *App) AddLauncherTool(tool LauncherTool) (LauncherTool, error) {
	tools, _ := a.loadLauncherTools()
	tool.ID = uuid.New().String()
	tools = append(tools, tool)
	err := a.saveLauncherTools(tools)
	return tool, err
}

func (a *App) UpdateLauncherTool(id string, tool LauncherTool) error {
	tools, _ := a.loadLauncherTools()
	updated := false
	for i, t := range tools {
		if t.ID == id {
			tools[i].Name = tool.Name
			tools[i].Path = tool.Path
			tools[i].Icon = tool.Icon
			tools[i].Args = tool.Args
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("tool not found")
	}
	return a.saveLauncherTools(tools)
}

func (a *App) DeleteLauncherTool(id string) error {
	tools, _ := a.loadLauncherTools()
	newTools := []LauncherTool{}
	for _, t := range tools {
		if t.ID != id {
			newTools = append(newTools, t)
		}
	}
	return a.saveLauncherTools(newTools)
}

func (a *App) RunLauncherTool(id string) error {
	tools, _ := a.loadLauncherTools()
	var tool *LauncherTool
	for _, t := range tools {
		if t.ID == id {
			tool = &t
			break
		}
	}

	if tool == nil {
		return fmt.Errorf("tool not found")
	}

	targetPath := strings.TrimSpace(tool.Path)
	if targetPath == "" {
		return fmt.Errorf("tool path is empty")
	}
	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tool file not found")
		}
		return err
	}

	args := []string{}
	if strings.TrimSpace(tool.Args) != "" {
		args = strings.Fields(tool.Args)
	}

	ext := strings.ToLower(filepath.Ext(targetPath))
	var cmd *exec.Cmd
	switch ext {
	case ".bat", ".cmd", ".html", ".htm", ".url":
		cmdArgs := append([]string{"/c", "start", "", targetPath}, args...)
		cmd = exec.Command("cmd.exe", cmdArgs...)
	default:
		cmd = exec.Command(targetPath, args...)
	}

	cmd.Dir = filepath.Dir(targetPath)
	return cmd.Start()
}

func (a *App) ExtractIcon(path string) (string, error) {
	if _, err := os.Stat(a.iconsDir()); os.IsNotExist(err) {
		os.MkdirAll(a.iconsDir(), 0755)
	}

	hash := md5.Sum([]byte(path))
	iconFilename := hex.EncodeToString(hash[:]) + ".png"
	iconPath := filepath.Join(a.iconsDir(), iconFilename)

	generateBase64 := func(ip string) (string, error) {
		bytes, err := os.ReadFile(ip)
		if err != nil {
			return "", err
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes), nil
	}

	if _, err := os.Stat(iconPath); err == nil {
		return generateBase64(iconPath)
	}

	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Drawing
		$icon = [System.Drawing.Icon]::ExtractAssociatedIcon('%s')
		if ($icon) {
			$bitmap = $icon.ToBitmap()
			$bitmap.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png)
			$bitmap.Dispose()
			$icon.Dispose()
		}
	`, path, iconPath)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", "& {"+psScript+"}")
	if _, err := cmd.CombinedOutput(); err != nil {
		return "", err
	}

	return generateBase64(iconPath)
}
