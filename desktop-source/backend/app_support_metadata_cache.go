package backend

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func (a *App) loadImageMetaCache() (ImageMetaCache, error) {
	var cache ImageMetaCache
	data, err := os.ReadFile(a.imageMetaCacheFile())
	if err != nil {
		if os.IsNotExist(err) {
			return ImageMetaCache{}, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		return ImageMetaCache{}, nil
	}
	if cache == nil {
		cache = ImageMetaCache{}
	}
	return cache, nil
}

func (a *App) saveImageMetaCache(cache ImageMetaCache) error {
	data, _ := json.MarshalIndent(cache, "", "  ")
	return os.WriteFile(a.imageMetaCacheFile(), data, 0644)
}

func measureDirectoryUsage(root string) (int, int, int64, error) {
	files := 0
	dirs := 0
	var bytesFreed int64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			dirs++
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		files++
		bytesFreed += info.Size()
		return nil
	})

	return files, dirs, bytesFreed, err
}

func (a *App) ClearPreviewCache() (CacheClearResult, error) {
	result := CacheClearResult{}
	var firstErr error

	if _, err := os.Stat(a.imageVariantsDir()); err == nil {
		files, dirs, bytesFreed, measureErr := measureDirectoryUsage(a.imageVariantsDir())
		if measureErr != nil {
			firstErr = measureErr
		} else {
			result.DeletedFiles += files
			result.DeletedDirs += dirs
			result.BytesFreed += bytesFreed
		}

		if err := os.RemoveAll(a.imageVariantsDir()); err != nil && firstErr == nil {
			firstErr = err
		}
	} else if err != nil && !os.IsNotExist(err) {
		firstErr = err
	}

	if err := os.MkdirAll(a.previewVariantsDir(), 0755); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := os.MkdirAll(a.thumbVariantsDir(), 0755); err != nil && firstErr == nil {
		firstErr = err
	}

	return result, firstErr
}

func (a *App) ensureImageMetaCacheLoaded() {
	a.imageMetaMu.Lock()
	defer a.imageMetaMu.Unlock()

	if a.imageMetaLoaded {
		return
	}

	cache, err := a.loadImageMetaCache()
	if err != nil {
		log.Printf("failed to load image metadata cache: %v", err)
		cache = ImageMetaCache{}
	}

	a.imageMetaCache = cache
	a.imageMetaLoaded = true
}

func (a *App) snapshotImageMetaCache() ImageMetaCache {
	a.imageMetaMu.RLock()
	defer a.imageMetaMu.RUnlock()

	cache := make(ImageMetaCache, len(a.imageMetaCache))
	for key, value := range a.imageMetaCache {
		cache[key] = value
	}
	return cache
}

func (a *App) replaceImageMetaCache(cache ImageMetaCache) {
	a.imageMetaMu.Lock()
	defer a.imageMetaMu.Unlock()

	a.imageMetaCache = cache
	a.imageMetaLoaded = true
}
