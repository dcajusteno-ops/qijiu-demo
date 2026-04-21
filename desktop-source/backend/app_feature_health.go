package backend

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (a *App) countEmptyFolders() int {
	if !a.hasDirectoryBinding() {
		return 0
	}

	emptyDirs := 0
	filepath.WalkDir(a.imageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".trash" || d.Name() == "desktop-source" {
			return fs.SkipDir
		}
		if path == a.imageDir {
			return nil
		}

		entries, readErr := os.ReadDir(path)
		if readErr != nil {
			return nil
		}
		isEmpty := true
		for _, entry := range entries {
			if entry.IsDir() {
				isEmpty = false
				break
			}
			nameUpper := strings.ToUpper(entry.Name())
			if nameUpper != "THUMBS.DB" && nameUpper != ".DS_STORE" && nameUpper != "DESKTOP.INI" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			emptyDirs++
		}
		return nil
	})

	return emptyDirs
}

func (a *App) countInvalidTagReferences() int {
	imageTags, _ := a.loadImageTags()
	validPaths := make(map[string]bool)
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		validPaths[relPath] = true
		return nil
	})

	invalid := 0
	for relPath := range imageTags {
		if !validPaths[relPath] {
			invalid++
		}
	}
	return invalid
}

func (a *App) countInvalidFavoriteReferences() int {
	groups, _ := a.loadFavoriteGroups()
	validPaths := make(map[string]bool)
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		validPaths[relPath] = true
		return nil
	})

	invalid := 0
	for _, group := range groups {
		for _, relPath := range group.Paths {
			if !validPaths[normalizeRelPath(relPath)] {
				invalid++
			}
		}
	}
	return invalid
}

func (a *App) buildDirectoryHealthSummary() (DirectoryHealthSummary, error) {
	totalImages := getManagedImagesCount(a)
	thumbFiles, _, thumbBytes, thumbErr := measureDirectoryUsage(a.thumbVariantsDir())
	if thumbErr != nil && !os.IsNotExist(thumbErr) {
		return DirectoryHealthSummary{}, thumbErr
	}
	previewFiles, _, previewBytes, previewErr := measureDirectoryUsage(a.previewVariantsDir())
	if previewErr != nil && !os.IsNotExist(previewErr) {
		return DirectoryHealthSummary{}, previewErr
	}

	summary := DirectoryHealthSummary{
		TotalImages:                   totalImages,
		EmptyFolderCount:              a.countEmptyFolders(),
		InvalidTagReferenceCount:      a.countInvalidTagReferences(),
		InvalidFavoriteReferenceCount: a.countInvalidFavoriteReferences(),
		ThumbCacheCount:               thumbFiles,
		ThumbCacheBytes:               thumbBytes,
		PreviewCacheCount:             previewFiles,
		PreviewCacheBytes:             previewBytes,
		Issues:                        []DirectoryHealthIssue{},
	}

	cacheCount := summary.ThumbCacheCount + summary.PreviewCacheCount
	if summary.EmptyFolderCount > 0 {
		summary.Issues = append(summary.Issues, DirectoryHealthIssue{
			Key:         "empty_folders",
			Level:       "warning",
			Title:       "存在空文件夹",
			Description: "可以清理无内容目录，减少层级噪音，让图库结构更整洁。",
			Count:       summary.EmptyFolderCount,
			Action:      "clean_empty_folders",
		})
	}
	if summary.InvalidTagReferenceCount > 0 {
		summary.Issues = append(summary.Issues, DirectoryHealthIssue{
			Key:         "invalid_tag_refs",
			Level:       "warning",
			Title:       "存在失效标签引用",
			Description: "部分标签仍指向已不存在的图片，可以执行清理来移除这些无效关联。",
			Count:       summary.InvalidTagReferenceCount,
			Action:      "cleanup_tags",
		})
	}
	if summary.InvalidFavoriteReferenceCount > 0 {
		summary.Issues = append(summary.Issues, DirectoryHealthIssue{
			Key:         "invalid_favorite_refs",
			Level:       "warning",
			Title:       "存在失效收藏引用",
			Description: "部分收藏分组仍引用已不存在的图片，可以执行清理来同步这些失效路径。",
			Count:       summary.InvalidFavoriteReferenceCount,
			Action:      "cleanup_favorites",
		})
	}
	if cacheCount > 0 {
		summary.Issues = append(summary.Issues, DirectoryHealthIssue{
			Key:         "cache_usage",
			Level:       "info",
			Title:       "存在缓存占用",
			Description: "可以清理预览缓存以回收空间，缩略图会在后续浏览时重新生成。",
			Count:       cacheCount,
			Action:      "clear_preview_cache",
		})
	}

	return summary, nil
}

func (a *App) GetDirectoryHealthSummary() (DirectoryHealthSummary, error) {
	return a.buildDirectoryHealthSummary()
}

func (a *App) RunDirectoryHealthAction(action string) (DirectoryHealthSummary, error) {
	switch strings.TrimSpace(action) {
	case "clear_preview_cache":
		if _, err := a.ClearPreviewCache(); err != nil {
			return DirectoryHealthSummary{}, err
		}
	case "clean_empty_folders":
		if _, err := a.CleanEmptyFolders(); err != nil {
			return DirectoryHealthSummary{}, err
		}
	case "cleanup_tags":
		if _, err := a.CleanupTags(); err != nil {
			return DirectoryHealthSummary{}, err
		}
	case "cleanup_favorites":
		if _, err := a.CleanupFavoriteReferences(); err != nil {
			return DirectoryHealthSummary{}, err
		}
	default:
		return DirectoryHealthSummary{}, fmt.Errorf("unsupported health action")
	}

	a.scheduleImagesChangedEvent()
	return a.buildDirectoryHealthSummary()
}

func (a *App) CleanEmptyFolders() (int, error) {
	var emptyDirs []string
	filepath.WalkDir(a.imageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".trash" || d.Name() == "desktop-source" {
				return fs.SkipDir
			}
			if path != a.imageDir {
				emptyDirs = append(emptyDirs, path)
			}
		}
		return nil
	})

	sort.Slice(emptyDirs, func(i, j int) bool {
		return len(emptyDirs[i]) > len(emptyDirs[j])
	})

	removedCount := 0
	for _, dirPath := range emptyDirs {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		isEmpty := true
		for _, entry := range entries {
			if entry.IsDir() {
				isEmpty = false
				break
			}
			nameUpper := strings.ToUpper(entry.Name())
			if nameUpper != "THUMBS.DB" && nameUpper != ".DS_STORE" && nameUpper != "DESKTOP.INI" {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			if os.RemoveAll(dirPath) == nil {
				removedCount++
			}
		}
	}

	return removedCount, nil
}

func (a *App) CleanupTags() (int, error) {
	return a.cleanupTagsSilent()
}

func (a *App) CleanupFavoriteReferences() (int, error) {
	return a.cleanupFavoriteReferencesSilent()
}

func (a *App) cleanupTagsSilent() (int, error) {
	imageTags, _ := a.loadImageTags()

	validPaths := make(map[string]bool)
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		validPaths[relPath] = true
		return nil
	})

	removedCount := 0
	newImageTags := make(ImageTagsMap)
	for relPath, tags := range imageTags {
		if validPaths[relPath] {
			newImageTags[relPath] = tags
		} else {
			removedCount++
		}
	}

	if removedCount > 0 {
		a.saveImageTags(newImageTags)
	}

	return removedCount, nil
}

func (a *App) cleanupFavoriteReferencesSilent() (int, error) {
	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return 0, err
	}

	validPaths := make(map[string]bool)
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		validPaths[relPath] = true
		return nil
	})

	removedCount := 0
	changed := false
	for i := range groups {
		filtered := make([]string, 0, len(groups[i].Paths))
		for _, relPath := range groups[i].Paths {
			normalized := normalizeRelPath(relPath)
			if normalized == "" || !validPaths[normalized] {
				removedCount++
				changed = true
				continue
			}
			filtered = append(filtered, normalized)
		}
		filtered = uniqueNonEmptyStrings(filtered)
		if len(filtered) != len(groups[i].Paths) {
			changed = true
		}
		groups[i].Paths = filtered
	}

	if changed {
		if err := a.saveFavoriteGroups(groups); err != nil {
			return 0, err
		}
	}

	return removedCount, nil
}
