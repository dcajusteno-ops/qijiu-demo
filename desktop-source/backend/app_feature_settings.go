package backend

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
)

func (a *App) loadSettings() (Settings, error) {
	var settings Settings
	data, err := os.ReadFile(a.settingsFile())
	if err != nil {
		return defaultSettings(), nil
	}
	data = stripUTF8BOM(data)
	if err := json.Unmarshal(data, &settings); err != nil {
		settings = defaultSettings()
		_ = a.saveSettings(settings)
		return settings, nil
	}
	originalShortcutSettings := settings.ShortcutSettings
	originalUserProfile := settings.UserProfile
	originalUtilityMenu := settings.UtilityMenu
	originalPerformanceSettings := settingsToGalleryPerformanceSettings(settings)
	originalOutputConfigured := settings.OutputConfigured
	originalTrashRetention := settings.TrashRetentionDays
	if settings.TrashRetentionDays <= 0 {
		settings.TrashRetentionDays = 30
	}
	if settings.OutputConfigured || strings.TrimSpace(settings.RootDir) != "" || strings.TrimSpace(settings.OutputDir) != "" {
		settings.OutputConfigured = true
	}
	settings.ShortcutSettings = normalizeShortcutSettings(settings.ShortcutSettings)
	settings.UserProfile = normalizeUserProfile(settings.UserProfile)
	settings.UtilityMenu = normalizeUtilityMenuState(settings.UtilityMenu)
	applyGalleryPerformanceSettings(&settings, settingsToGalleryPerformanceSettings(settings))
	if settings.TrashRetentionDays != originalTrashRetention ||
		settings.OutputConfigured != originalOutputConfigured ||
		!reflect.DeepEqual(settings.ShortcutSettings, originalShortcutSettings) ||
		!reflect.DeepEqual(settings.UserProfile, originalUserProfile) ||
		!reflect.DeepEqual(settings.UtilityMenu, originalUtilityMenu) ||
		!reflect.DeepEqual(settingsToGalleryPerformanceSettings(settings), originalPerformanceSettings) {
		_ = a.saveSettings(settings)
	}
	return settings, nil
}

func (a *App) saveSettings(settings Settings) error {
	data, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(a.settingsFile(), data, 0644)
}

func (a *App) GetTrashSettings() (Settings, error) {
	return a.loadSettings()
}

func (a *App) SaveTrashSettings(settings Settings) error {
	current, err := a.loadSettings()
	if err != nil {
		return err
	}
	current.TrashRetentionDays = settings.TrashRetentionDays
	if current.TrashRetentionDays <= 0 {
		current.TrashRetentionDays = 30
	}
	return a.saveSettings(current)
}

func (a *App) GetUtilityMenuSettings() (UtilityMenuState, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUtilityMenuState(), err
	}
	return normalizeUtilityMenuState(settings.UtilityMenu), nil
}

func (a *App) SaveUtilityMenuSettings(state UtilityMenuState) (UtilityMenuState, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUtilityMenuState(), err
	}

	settings.UtilityMenu = normalizeUtilityMenuState(state)
	if err := a.saveSettings(settings); err != nil {
		return defaultUtilityMenuState(), err
	}
	return settings.UtilityMenu, nil
}

func (a *App) GetGalleryPerformanceSettings() (GalleryPerformanceSettings, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultGalleryPerformanceSettings(), err
	}
	return settingsToGalleryPerformanceSettings(settings), nil
}

func (a *App) SaveGalleryPerformanceSettings(performance GalleryPerformanceSettings) (GalleryPerformanceSettings, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultGalleryPerformanceSettings(), err
	}
	applyGalleryPerformanceSettings(&settings, performance)
	if err := a.saveSettings(settings); err != nil {
		return defaultGalleryPerformanceSettings(), err
	}
	return settingsToGalleryPerformanceSettings(settings), nil
}
