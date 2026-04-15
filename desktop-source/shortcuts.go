package main

import (
	"fmt"
	"strings"
)

type ShortcutBinding struct {
	Action      string `json:"action"`
	Accelerator string `json:"accelerator"`
}

type ShortcutSettings struct {
	Enabled  bool              `json:"enabled"`
	Bindings []ShortcutBinding `json:"bindings"`
}

type ShortcutAction struct {
	ID                 string `json:"id"`
	Label              string `json:"label"`
	Description        string `json:"description"`
	DefaultAccelerator string `json:"defaultAccelerator"`
}

type shortcutManager interface {
	Apply(app *App, settings ShortcutSettings) error
	Close() error
}

var shortcutCatalog = []ShortcutAction{
	{
		ID:                 "switch_dashboard",
		Label:              "切换到总览",
		Description:        "快速进入工作室总览页面",
		DefaultAccelerator: "Ctrl+Alt+1",
	},
	{
		ID:                 "switch_gallery",
		Label:              "切换到图库",
		Description:        "快速回到主图库视图",
		DefaultAccelerator: "Ctrl+Alt+2",
	},
	{
		ID:                 "switch_favorites",
		Label:              "切换到收藏",
		Description:        "快速查看收藏图片",
		DefaultAccelerator: "Ctrl+Alt+3",
	},
	{
		ID:                 "switch_documentation",
		Label:              "切换到文档",
		Description:        "快速打开使用文档",
		DefaultAccelerator: "Ctrl+Alt+4",
	},
	{
		ID:                 "refresh_images",
		Label:              "刷新图库",
		Description:        "立即重新扫描当前图片列表",
		DefaultAccelerator: "Ctrl+Alt+R",
	},
	{
		ID:                 "toggle_sidebar",
		Label:              "折叠侧边栏",
		Description:        "切换左侧导航栏展开状态",
		DefaultAccelerator: "Ctrl+Alt+B",
	},
	{
		ID:                 "toggle_selection_mode",
		Label:              "切换批量模式",
		Description:        "快速进入或退出批量选择模式",
		DefaultAccelerator: "Ctrl+Alt+M",
	},
}

var shortcutCatalogByID = func() map[string]ShortcutAction {
	result := make(map[string]ShortcutAction, len(shortcutCatalog))
	for _, action := range shortcutCatalog {
		result[action.ID] = action
	}
	return result
}()

func defaultShortcutSettings() ShortcutSettings {
	bindings := make([]ShortcutBinding, 0, len(shortcutCatalog))
	for _, action := range shortcutCatalog {
		bindings = append(bindings, ShortcutBinding{
			Action:      action.ID,
			Accelerator: action.DefaultAccelerator,
		})
	}
	return ShortcutSettings{
		Enabled:  true,
		Bindings: bindings,
	}
}

func normalizeShortcutSettings(settings ShortcutSettings) ShortcutSettings {
	if len(settings.Bindings) == 0 {
		return defaultShortcutSettings()
	}

	custom := make(map[string]string, len(settings.Bindings))
	for _, binding := range settings.Bindings {
		actionID := strings.TrimSpace(binding.Action)
		if actionID == "" {
			continue
		}
		if _, ok := shortcutCatalogByID[actionID]; !ok {
			continue
		}
		custom[actionID] = normalizeAccelerator(binding.Accelerator)
	}

	normalized := ShortcutSettings{
		Enabled:  settings.Enabled,
		Bindings: make([]ShortcutBinding, 0, len(shortcutCatalog)),
	}

	for _, action := range shortcutCatalog {
		accelerator, ok := custom[action.ID]
		if !ok {
			accelerator = action.DefaultAccelerator
		}
		normalized.Bindings = append(normalized.Bindings, ShortcutBinding{
			Action:      action.ID,
			Accelerator: accelerator,
		})
	}

	return normalized
}

func normalizeAccelerator(accelerator string) string {
	rawParts := strings.Split(strings.TrimSpace(accelerator), "+")
	hasCtrl := false
	hasAlt := false
	hasShift := false
	hasWin := false
	key := ""

	for _, part := range rawParts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		switch strings.ToLower(token) {
		case "ctrl", "control":
			hasCtrl = true
		case "alt":
			hasAlt = true
		case "shift":
			hasShift = true
		case "win", "meta", "cmd", "super":
			hasWin = true
		default:
			key = strings.ToUpper(token)
		}
	}

	parts := make([]string, 0, 5)
	if hasCtrl {
		parts = append(parts, "Ctrl")
	}
	if hasAlt {
		parts = append(parts, "Alt")
	}
	if hasShift {
		parts = append(parts, "Shift")
	}
	if hasWin {
		parts = append(parts, "Win")
	}
	if key != "" {
		parts = append(parts, key)
	}

	return strings.Join(parts, "+")
}

func (a *App) GetShortcutSettings() (ShortcutSettings, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultShortcutSettings(), err
	}
	return normalizeShortcutSettings(settings.ShortcutSettings), nil
}

func (a *App) GetShortcutActions() []ShortcutAction {
	actions := make([]ShortcutAction, len(shortcutCatalog))
	copy(actions, shortcutCatalog)
	return actions
}

func (a *App) SaveShortcutSettings(input ShortcutSettings) (ShortcutSettings, error) {
	normalized := normalizeShortcutSettings(input)
	if err := validateShortcutBindings(normalized); err != nil {
		return normalized, err
	}

	if a.shortcutManager != nil {
		if err := a.shortcutManager.Apply(a, normalized); err != nil {
			return normalized, err
		}
	}

	settings, err := a.loadSettings()
	if err != nil {
		return normalized, err
	}
	settings.ShortcutSettings = normalized
	if err := a.saveSettings(settings); err != nil {
		return normalized, err
	}

	return normalized, nil
}

func (a *App) registerConfiguredShortcuts() error {
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	if a.shortcutManager == nil {
		return nil
	}
	return a.shortcutManager.Apply(a, settings.ShortcutSettings)
}

func validateShortcutBindings(settings ShortcutSettings) error {
	seenActions := make(map[string]struct{}, len(shortcutCatalog))
	seenAccelerators := make(map[string]string, len(shortcutCatalog))

	for _, binding := range settings.Bindings {
		actionID := strings.TrimSpace(binding.Action)
		if actionID == "" {
			return fmt.Errorf("快捷键动作不能为空")
		}
		if _, ok := shortcutCatalogByID[actionID]; !ok {
			return fmt.Errorf("未知快捷键动作: %s", actionID)
		}
		if _, exists := seenActions[actionID]; exists {
			return fmt.Errorf("快捷键动作重复: %s", actionID)
		}
		seenActions[actionID] = struct{}{}

		accelerator := normalizeAccelerator(binding.Accelerator)
		if accelerator == "" {
			continue
		}

		if otherAction, exists := seenAccelerators[accelerator]; exists {
			return fmt.Errorf("快捷键 %s 同时分配给了 %s 和 %s", accelerator, otherAction, actionID)
		}
		seenAccelerators[accelerator] = actionID
	}

	return nil
}
