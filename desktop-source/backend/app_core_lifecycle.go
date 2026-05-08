package backend

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func NewApp() *App {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(exePath)

	unifiedRoot := exeDir

	exeBase := filepath.Base(exeDir)
	exeParent := filepath.Dir(exeDir)
	if exeBase == "bin" && filepath.Base(exeParent) == "build" {
		unifiedRoot = filepath.Dir(filepath.Dir(exeParent))
	} else if strings.Contains(exePath, os.TempDir()) || strings.EqualFold(exeBase, "tmp") {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			unifiedRoot = filepath.Dir(wd)
		}
	}

	defaultOutputDir := filepath.Dir(unifiedRoot)
	defaultRootDir := filepath.Dir(defaultOutputDir)
	dataDir := filepath.Join(unifiedRoot, "data")

	app := &App{
		rootDir:         defaultRootDir,
		imageDir:        defaultOutputDir,
		dataDir:         dataDir,
		appDir:          unifiedRoot,
		shortcutManager: newShortcutManager(),
	}

	if _, err := os.Stat(app.dataDir); os.IsNotExist(err) {
		os.MkdirAll(app.dataDir, 0755)
	}

	settings, _ := app.loadSettings()
	if settings.OutputConfigured {
		if err := app.applyDirectoryBinding(settings.RootDir, settings.OutputDir); err != nil {
			log.Printf("failed to apply saved directory binding: %v", err)
			app.rootDir = ""
			app.imageDir = ""
			settings.OutputConfigured = false
			settings.RootDir = ""
			settings.OutputDir = ""
			_ = app.saveSettings(settings)
		}
	} else {
		app.rootDir = ""
		app.imageDir = ""
	}

	if settings.OutputConfigured && (strings.TrimSpace(settings.RootDir) != "" || strings.TrimSpace(settings.OutputDir) != "") && settings.PathVersion < pathVersionRootRelative {
		if err := app.migrateLegacyPathData(&settings); err != nil {
			log.Printf("failed to migrate legacy paths: %v", err)
		} else {
			if strings.TrimSpace(settings.RootDir) == "" {
				settings.RootDir = app.rootDir
			}
			if strings.TrimSpace(settings.OutputDir) == "" {
				settings.OutputDir = app.imageDir
			}
			settings.PathVersion = pathVersionRootRelative
			_ = app.saveSettings(settings)
		}
	}

	if err := app.migrateLegacyTrash(); err != nil {
		log.Printf("failed to migrate legacy trash: %v", err)
	}

	if err := os.MkdirAll(app.trashDir(), 0755); err != nil {
		log.Printf("failed to ensure trash directory: %v", err)
	}

	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		_ = a.cleanExpiredTrash()
		_, _ = a.cleanupTagsSilent()
	}()
	if err := a.registerConfiguredShortcuts(); err != nil {
		log.Printf("failed to register shortcuts: %v", err)
	}
	a.applySavedWindowSettings(ctx)
	a.restartImageWatcher()
}

func (a *App) shutdown(ctx context.Context) {
	a.stopImageWatcher()
	if a.shortcutManager != nil {
		if err := a.shortcutManager.Close(); err != nil {
			log.Printf("failed to stop shortcut manager: %v", err)
		}
	}
}
