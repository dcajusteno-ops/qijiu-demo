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
	{ID: "switch_dashboard", Label: "工作台总览", Description: "打开工作台总览页面", DefaultAccelerator: "Ctrl+Shift+F1"},
	{ID: "switch_gallery", Label: "默认图库", Description: "打开默认目录图库", DefaultAccelerator: "Ctrl+Shift+F2"},
	{ID: "switch_favorites", Label: "收藏夹", Description: "打开收藏夹视图", DefaultAccelerator: "Ctrl+Shift+F3"},
	{ID: "switch_documentation", Label: "使用文档", Description: "打开使用文档页面", DefaultAccelerator: "Ctrl+Shift+F4"},
	{ID: "refresh_images", Label: "刷新图库", Description: "重新加载图片与元数据", DefaultAccelerator: "Ctrl+Shift+F5"},
	{ID: "toggle_sidebar", Label: "切换侧边栏", Description: "折叠或展开左侧边栏", DefaultAccelerator: "Ctrl+Shift+F6"},
	{ID: "toggle_selection_mode", Label: "切换批量模式", Description: "进入或退出批量选择模式", DefaultAccelerator: "Ctrl+Shift+F7"},
	{ID: "switch_auto_rules", Label: "自动规则引擎", Description: "打开自动规则引擎页面", DefaultAccelerator: "Ctrl+Shift+F8"},
	{ID: "switch_date_workbench", Label: "日期产出工作台", Description: "打开日期产出工作台", DefaultAccelerator: "Ctrl+Shift+F9"},
}

var shortcutCatalogByID = func() map[string]ShortcutAction {
	result := make(map[string]ShortcutAction, len(shortcutCatalog))
	for _, action := range shortcutCatalog {
		result[action.ID] = action
	}
	return result
}()

var legacyShortcutDefaults = map[string][]string{
	"switch_dashboard":      {"Ctrl+Alt+1"},
	"switch_gallery":        {"Ctrl+Alt+2"},
	"switch_favorites":      {"Ctrl+Alt+3"},
	"switch_documentation":  {"Ctrl+Alt+4"},
	"refresh_images":        {"Ctrl+Alt+R", "Ctrl+Alt+F5"},
	"toggle_sidebar":        {"Ctrl+Alt+B"},
	"toggle_selection_mode": {"Ctrl+Alt+M"},
	"switch_auto_rules":     {},
	"switch_date_workbench": {},
}

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

func migrateLegacyShortcutAccelerator(actionID, accelerator string) string {
	normalized := normalizeAccelerator(accelerator)
	for _, action := range shortcutCatalog {
		if action.ID != actionID {
			continue
		}
		for _, legacy := range legacyShortcutDefaults[actionID] {
			if normalized == normalizeAccelerator(legacy) {
				return action.DefaultAccelerator
			}
		}
		return normalized
	}
	return normalized
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
		custom[actionID] = migrateLegacyShortcutAccelerator(actionID, binding.Accelerator)
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
	normalized := normalizeShortcutSettings(settings.ShortcutSettings)
	if err := validateShortcutBindings(normalized); err != nil {
		return err
	}
	if a.shortcutManager == nil {
		return nil
	}
	return a.shortcutManager.Apply(a, normalized)
}

func validateShortcutBindings(settings ShortcutSettings) error {
	seenActions := make(map[string]struct{}, len(shortcutCatalog))
	seenAccelerators := make(map[string]string, len(shortcutCatalog))

	for _, binding := range settings.Bindings {
		actionID := strings.TrimSpace(binding.Action)
		if actionID == "" {
			return fmt.Errorf("shortcut action cannot be empty")
		}
		if _, ok := shortcutCatalogByID[actionID]; !ok {
			return fmt.Errorf("unknown shortcut action: %s", actionID)
		}
		if _, exists := seenActions[actionID]; exists {
			return fmt.Errorf("duplicate shortcut action: %s", actionID)
		}
		seenActions[actionID] = struct{}{}

		accelerator := normalizeAccelerator(binding.Accelerator)
		if accelerator == "" {
			continue
		}
		if owner, exists := seenAccelerators[accelerator]; exists {
			ownerLabel := shortcutCatalogByID[owner].Label
			currentLabel := shortcutCatalogByID[actionID].Label
			return fmt.Errorf("%s and %s use the same accelerator %s", ownerLabel, currentLabel, accelerator)
		}
		seenAccelerators[accelerator] = actionID
	}

	for _, action := range shortcutCatalog {
		if _, exists := seenActions[action.ID]; !exists {
			return fmt.Errorf("missing shortcut action: %s", action.Label)
		}
	}

	return nil
}
