//go:build !windows

package main

type noopShortcutManager struct{}

func newShortcutManager() shortcutManager {
	return &noopShortcutManager{}
}

func (m *noopShortcutManager) Apply(app *App, settings ShortcutSettings) error {
	return nil
}

func (m *noopShortcutManager) Close() error {
	return nil
}
