package backend

import (
	"io/fs"
	"os"
	"strings"
	"time"
)

func (a *App) GetImageGallerySummary() (ImageGallerySummary, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return ImageGallerySummary{}, err
	}

	totalImages := getManagedImagesCount(a)
	mode, reason := resolveGalleryLoadMode(settingsToGalleryPerformanceSettings(settings), totalImages)
	thumbFiles, _, thumbBytes, thumbErr := measureDirectoryUsage(a.thumbVariantsDir())
	if thumbErr != nil && !os.IsNotExist(thumbErr) {
		return ImageGallerySummary{}, thumbErr
	}
	previewFiles, _, previewBytes, previewErr := measureDirectoryUsage(a.previewVariantsDir())
	if previewErr != nil && !os.IsNotExist(previewErr) {
		return ImageGallerySummary{}, previewErr
	}

	return ImageGallerySummary{
		TotalImages:       totalImages,
		ManagedRootCount:  len(a.managedImageRoots()),
		ActiveMode:        mode,
		ModeReason:        reason,
		ThumbCacheCount:   thumbFiles,
		ThumbCacheBytes:   thumbBytes,
		PreviewCacheCount: previewFiles,
		PreviewCacheBytes: previewBytes,
	}, nil
}

func (a *App) GetStatistics(period string) (*Stats, error) {
	stats := &Stats{
		ByDate: make(map[string]int),
		ByTag:  make(map[string]int),
	}

	imageTags, _ := a.loadImageTags()
	tagIDsToNames := make(map[string]string)
	tags, _ := a.loadTags()
	for _, t := range tags {
		tagIDsToNames[t.ID] = t.Name
	}

	now := time.Now().In(time.Local)
	today := now.Format("2006-01-02")

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		stats.TotalCount++
		stats.TotalSize += info.Size()

		modTime := info.ModTime().In(time.Local)
		dateStr := extractDateKeyFromRelPath(relPath, modTime)
		if strings.TrimSpace(dateStr) == "" {
			dateStr = modTime.Format("2006-01-02")
		}
		if dateStr == today {
			stats.TodayCount++
		}

		dateKey := dateStr
		if period == "month" {
			if parsedDate := parseDate(dateStr); !parsedDate.IsZero() {
				dateKey = parsedDate.Format("2006-01")
			} else {
				dateKey = modTime.Format("2006-01")
			}
		} else if period == "year" {
			if parsedDate := parseDate(dateStr); !parsedDate.IsZero() {
				dateKey = parsedDate.Format("2006")
			} else {
				dateKey = modTime.Format("2006")
			}
		}
		stats.ByDate[dateKey]++

		if tids, ok := imageTags[relPath]; ok {
			for _, tid := range tids {
				if name, ok := tagIDsToNames[tid]; ok {
					stats.ByTag[name]++
				}
			}
		}
		return nil
	})

	return stats, err
}
