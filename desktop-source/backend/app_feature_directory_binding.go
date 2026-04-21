package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func (a *App) GetDirectoryBinding() (DirectoryBinding, error) {
	return DirectoryBinding{
		RootDir:       a.rootDir,
		OutputDir:     a.imageDir,
		OutputRelPath: a.outputRelPath(),
		Configured:    a.hasDirectoryBinding(),
	}, nil
}

func (a *App) SaveDirectoryBinding(rootDir, outputDir string) (DirectoryBinding, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return DirectoryBinding{}, err
	}

	previousRoot := a.rootDir
	previousOutput := a.imageDir
	previousSettingsRoot := settings.RootDir
	previousSettingsOutput := settings.OutputDir
	previousSettingsConfigured := settings.OutputConfigured
	previousSettingsVersion := settings.PathVersion

	nextRoot, nextOutput, err := a.validateDirectoryBinding(rootDir, outputDir)
	if err != nil {
		return DirectoryBinding{}, err
	}

	a.rootDir = nextRoot
	a.imageDir = nextOutput

	settings.RootDir = a.rootDir
	settings.OutputDir = a.imageDir
	settings.OutputConfigured = true
	settings.PathVersion = pathVersionRootRelative
	if settings.TrashRetentionDays <= 0 {
		settings.TrashRetentionDays = 30
	}

	if err := a.saveSettings(settings); err != nil {
		a.restoreDirectoryBinding(previousRoot, previousOutput)
		return DirectoryBinding{}, err
	}

	if err := os.MkdirAll(a.trashDir(), 0755); err != nil {
		settings.RootDir = previousSettingsRoot
		settings.OutputDir = previousSettingsOutput
		settings.OutputConfigured = previousSettingsConfigured
		settings.PathVersion = previousSettingsVersion
		_ = a.saveSettings(settings)
		a.restoreDirectoryBinding(previousRoot, previousOutput)
		return DirectoryBinding{}, err
	}

	if err := a.migrateLegacyTrash(); err != nil {
		log.Printf("failed to migrate legacy trash after rebinding: %v", err)
	}

	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return a.GetDirectoryBinding()
}

func (a *App) SaveOutputDirectory(outputDir string) (DirectoryBinding, error) {
	normalizedOutput, err := normalizeDir(outputDir)
	if err != nil {
		return DirectoryBinding{}, fmt.Errorf("invalid output directory: %w", err)
	}

	rootDir := filepath.Dir(normalizedOutput)
	return a.SaveDirectoryBinding(rootDir, normalizedOutput)
}
