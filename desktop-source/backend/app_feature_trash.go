package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (a *App) loadTrashMetadataRaw() (TrashMetadataMap, error) {
	var meta TrashMetadataMap
	data, err := os.ReadFile(a.trashMetadataFile())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		_ = json.Unmarshal(data, &meta)
	}
	if meta == nil {
		meta = make(TrashMetadataMap)
	}
	return meta, nil
}

func (a *App) loadTrashMetadata() (TrashMetadataMap, error) {
	meta, err := a.loadTrashMetadataRaw()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(a.trashDir())
	if err == nil {
		changed := false
		existingFiles := make(map[string]bool)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			existingFiles[name] = true

			if _, exists := meta[name]; !exists {
				info, _ := entry.Info()
				deletedAt := time.Now()
				if info != nil {
					deletedAt = info.ModTime()
				}
				dateFolder := deletedAt.Format("2006-01-02")
				meta[name] = TrashMetadata{
					OriginalPath: filepath.ToSlash(filepath.Join(dateFolder, name)),
					DeletedAt:    deletedAt.Format(time.RFC3339),
				}
				changed = true
			}
		}

		for name := range meta {
			if !existingFiles[name] {
				delete(meta, name)
				changed = true
			}
		}
		if changed {
			a.saveTrashMetadata(meta)
		}
	}

	return meta, nil
}

func uniqueTrashFilename(name string, existing map[string]bool) string {
	if !existing[name] {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	timestamp := time.Now().Format("20060102_150405")
	candidate := fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	index := 1
	for existing[candidate] {
		candidate = fmt.Sprintf("%s_%s_%d%s", base, timestamp, index, ext)
		index++
	}
	return candidate
}

func (a *App) migrateLegacyTrash() error {
	currentDir := a.trashDir()
	if currentDir == "" {
		return nil
	}

	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return err
	}

	meta, err := a.loadTrashMetadataRaw()
	if err != nil {
		return err
	}

	existing := make(map[string]bool)
	if currentEntries, readErr := os.ReadDir(currentDir); readErr == nil {
		for _, entry := range currentEntries {
			if entry.IsDir() {
				continue
			}
			existing[entry.Name()] = true
		}
	}

	metaChanged := false

	for _, legacyDir := range a.legacyTrashDirs() {
		if legacyDir == "" || samePath(legacyDir, currentDir) {
			continue
		}

		entries, err := os.ReadDir(legacyDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		movedAny := false

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			sourceName := entry.Name()
			sourcePath := filepath.Join(legacyDir, sourceName)
			targetName := uniqueTrashFilename(sourceName, existing)
			targetPath := filepath.Join(currentDir, targetName)

			if err := moveFile(sourcePath, targetPath); err != nil {
				log.Printf("failed to migrate trash file %s: %v", sourceName, err)
				continue
			}

			existing[targetName] = true
			movedAny = true

			if targetName != sourceName {
				if item, ok := meta[sourceName]; ok {
					delete(meta, sourceName)
					meta[targetName] = item
				} else {
					info, _ := os.Stat(targetPath)
					deletedAt := time.Now()
					if info != nil {
						deletedAt = info.ModTime()
					}
					dateFolder := deletedAt.Format("2006-01-02")
					meta[targetName] = TrashMetadata{
						OriginalPath: filepath.ToSlash(filepath.Join(dateFolder, targetName)),
						DeletedAt:    deletedAt.Format(time.RFC3339),
					}
				}
				metaChanged = true
			}
		}

		if movedAny {
			if remaining, readErr := os.ReadDir(legacyDir); readErr == nil && len(remaining) == 0 {
				_ = os.Remove(legacyDir)
			}
		}
	}

	if metaChanged {
		if err := a.saveTrashMetadata(meta); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) saveTrashMetadata(meta TrashMetadataMap) error {
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(a.trashMetadataFile(), data, 0644)
}

func (a *App) cleanExpiredTrash() error {
	settings, _ := a.loadSettings()
	meta, _ := a.loadTrashMetadata()

	cutoffTime := time.Now().AddDate(0, 0, -settings.TrashRetentionDays)
	deletedCount := 0

	for trashFilename, item := range meta {
		itemTime, _ := time.Parse(time.RFC3339, item.DeletedAt)
		if itemTime.Before(cutoffTime) {
			trashPath := filepath.Join(a.trashDir(), trashFilename)
			if err := os.Remove(trashPath); err == nil || os.IsNotExist(err) {
				delete(meta, trashFilename)
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		return a.saveTrashMetadata(meta)
	}
	return nil
}

func (a *App) DeleteImage(relPath string) error {
	relPath = normalizeRelPath(relPath)
	targetPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid filename")
	}
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	fileName := filepath.Base(targetPath)
	trashPath := filepath.Join(a.trashDir(), fileName)

	if _, err := os.Stat(trashPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		trashPath = filepath.Join(a.trashDir(), fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := moveFile(targetPath, trashPath); err != nil {
		return fmt.Errorf("failed to move to trash: %v", err)
	}

	meta, _ := a.loadTrashMetadata()
	if meta == nil {
		meta = make(TrashMetadataMap)
	}
	trashFilename := filepath.Base(trashPath)
	meta[trashFilename] = TrashMetadata{
		OriginalPath: relPath,
		DeletedAt:    time.Now().Format(time.RFC3339),
	}
	a.saveTrashMetadata(meta)

	a.scheduleImagesChangedEvent()

	return nil
}

func (a *App) GetTrashList() ([]TrashItem, error) {
	meta, _ := a.loadTrashMetadata()
	items := []TrashItem{}

	for filename, itemMeta := range meta {
		items = append(items, TrashItem{
			Filename:     filename,
			OriginalPath: itemMeta.OriginalPath,
			DeletedAt:    itemMeta.DeletedAt,
			Path:         a.trashRelPath(filename),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DeletedAt > items[j].DeletedAt
	})

	return items, nil
}

func (a *App) RestoreTrash(filename string) (string, error) {
	meta, _ := a.loadTrashMetadata()
	itemMeta, exists := meta[filename]
	if !exists {
		return "", fmt.Errorf("file not found in trash")
	}

	trashPath := filepath.Join(a.trashDir(), filename)
	restorePath := filepath.Join(a.imageDir, itemMeta.OriginalPath)

	os.MkdirAll(filepath.Dir(restorePath), 0755)

	if _, err := os.Stat(restorePath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(restorePath)
		base := strings.TrimSuffix(restorePath, ext)
		restorePath = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	}

	if err := moveFile(trashPath, restorePath); err != nil {
		return "", err
	}

	delete(meta, filename)
	a.saveTrashMetadata(meta)

	return restorePath, nil
}

func (a *App) BatchRestoreTrash(filenames []string) (int, error) {
	meta, _ := a.loadTrashMetadata()
	successCount := 0

	for _, filename := range filenames {
		itemMeta, exists := meta[filename]
		if !exists {
			continue
		}

		trashPath := filepath.Join(a.trashDir(), filename)
		restorePath := filepath.Join(a.imageDir, itemMeta.OriginalPath)
		os.MkdirAll(filepath.Dir(restorePath), 0755)

		if _, err := os.Stat(restorePath); err == nil {
			timestamp := time.Now().Format("20060102_150405")
			ext := filepath.Ext(restorePath)
			name := strings.TrimSuffix(restorePath, ext)
			restorePath = fmt.Sprintf("%s_%s%s", name, timestamp, ext)
		}

		if err := moveFile(trashPath, restorePath); err == nil {
			delete(meta, filename)
			successCount++
		}
	}

	a.saveTrashMetadata(meta)
	return successCount, nil
}

func (a *App) BatchDeleteTrash(filenames []string) (int, error) {
	meta, _ := a.loadTrashMetadata()
	deletedCount := 0

	for _, filename := range filenames {
		trashPath := filepath.Join(a.trashDir(), filename)
		if err := os.RemoveAll(trashPath); err == nil || os.IsNotExist(err) {
			deletedCount++
			delete(meta, filename)
		}
	}

	a.saveTrashMetadata(meta)
	return deletedCount, nil
}

func (a *App) EmptyTrash() (int, error) {
	deletedCount := 0
	entries, err := os.ReadDir(a.trashDir())
	if err == nil {
		for _, entry := range entries {
			path := filepath.Join(a.trashDir(), entry.Name())
			if err := os.RemoveAll(path); err == nil {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		a.saveTrashMetadata(TrashMetadataMap{})
	}

	return deletedCount, nil
}
