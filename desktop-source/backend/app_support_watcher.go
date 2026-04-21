package backend

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) shouldSkipDir(path, name string) bool {
	switch name {
	case "node_modules", ".git", ".trash":
		return true
	}

	lowerName := strings.ToLower(name)
	if strings.HasPrefix(lowerName, "comfy-manager") {
		return true
	}

	return a.appDir != "" && (samePath(path, a.appDir) || isSubPath(a.appDir, path))
}

func (a *App) managedImageRoots() []string {
	if !a.hasDirectoryBinding() {
		return []string{}
	}

	candidates := []string{a.imageDir}

	customRoots, err := a.loadCustomRoots()
	if err == nil {
		for _, root := range customRoots {
			if !root.Enabled {
				continue
			}
			absPath, resolveErr := a.resolveRootPath(root.Path)
			if resolveErr != nil {
				continue
			}
			info, statErr := os.Stat(absPath)
			if statErr != nil || !info.IsDir() {
				continue
			}
			candidates = append(candidates, absPath)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return len(filepath.Clean(candidates[i])) < len(filepath.Clean(candidates[j]))
	})

	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		normalized := filepath.Clean(candidate)
		covered := false
		for _, existing := range roots {
			if isSubPath(existing, normalized) {
				covered = true
				break
			}
		}
		if !covered {
			roots = append(roots, normalized)
		}
	}

	return roots
}

func (a *App) stopImageWatcher() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.imageWatchDebounce != nil {
		a.imageWatchDebounce.Stop()
		a.imageWatchDebounce = nil
	}
	if a.imageWatchStop != nil {
		close(a.imageWatchStop)
		a.imageWatchStop = nil
	}
	if a.imageWatcher != nil {
		_ = a.imageWatcher.Close()
		a.imageWatcher = nil
	}
}

func (a *App) scheduleImagesChangedEvent() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.ctx == nil {
		return
	}

	if a.imageWatchDebounce != nil {
		a.imageWatchDebounce.Stop()
	}
	a.imageWatchDebounce = time.AfterFunc(350*time.Millisecond, func() {
		runtime.EventsEmit(a.ctx, "images:changed")
	})
}

func (a *App) emitAutoRulesProgress(progress AutoRulesRunProgress) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "auto-rules:progress", progress)
}

func (a *App) addWatchTree(watcher *fsnotify.Watcher, root string, seen map[string]struct{}) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if a.shouldSkipDir(path, d.Name()) {
			if samePath(path, root) {
				return nil
			}
			return fs.SkipDir
		}

		cleaned := filepath.Clean(path)
		if _, ok := seen[cleaned]; ok {
			return nil
		}
		if err := watcher.Add(cleaned); err != nil {
			return nil
		}
		seen[cleaned] = struct{}{}
		return nil
	})
}

func shouldReactToWatchEvent(event fsnotify.Event) bool {
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0
}

func (a *App) restartImageWatcher() {
	a.stopImageWatcher()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("failed to create image watcher: %v", err)
		return
	}

	seen := make(map[string]struct{})
	for _, root := range a.managedImageRoots() {
		if err := a.addWatchTree(watcher, root, seen); err != nil {
			log.Printf("failed to watch root %s: %v", root, err)
		}
	}

	stop := make(chan struct{})

	a.watchMu.Lock()
	a.imageWatcher = watcher
	a.imageWatchStop = stop
	a.watchMu.Unlock()

	go func(localWatcher *fsnotify.Watcher, stopCh chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case event, ok := <-localWatcher.Events:
				if !ok {
					return
				}
				if !shouldReactToWatchEvent(event) {
					continue
				}

				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						a.restartImageWatcher()
						a.scheduleImagesChangedEvent()
						continue
					}
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					a.restartImageWatcher()
				}

				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					a.scheduleImagesChangedEvent()
					continue
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 || isImageExt(filepath.Ext(event.Name)) {
					a.scheduleImagesChangedEvent()
				}
			case err, ok := <-localWatcher.Errors:
				if !ok {
					return
				}
				log.Printf("image watcher error: %v", err)
			}
		}
	}(watcher, stop)
}

func (a *App) walkManagedImages(visitor func(absPath, relPath string, info fs.FileInfo) error) error {
	if !a.hasDirectoryBinding() {
		return nil
	}

	seen := make(map[string]bool)

	for _, root := range a.managedImageRoots() {
		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				if a.shouldSkipDir(path, d.Name()) {
					return fs.SkipDir
				}
				return nil
			}

			if !isImageExt(filepath.Ext(path)) {
				return nil
			}

			relPath, relErr := filepath.Rel(a.rootDir, path)
			if relErr != nil {
				return nil
			}
			relPath = normalizeRelPath(relPath)

			if seen[relPath] {
				return nil
			}
			seen[relPath] = true

			info, infoErr := d.Info()
			if infoErr != nil {
				return nil
			}

			return visitor(path, relPath, info)
		})
		if walkErr != nil {
			return walkErr
		}
	}

	return nil
}
