package backend

import (
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type imageMetaWarmupTask struct {
	Path  string
	Entry ImageMetaCacheEntry
}

func resolveGalleryLoadMode(settings GalleryPerformanceSettings, totalImages int) (string, string) {
	switch settings.Mode {
	case "performance":
		return "performance", "已手动切换到性能模式"
	case "standard":
		return "standard", "已手动保持标准模式"
	default:
		if totalImages >= 3000 {
			return "performance", "图库数量较大，已自动切换到性能模式"
		}
		return "standard", "图库规模尚可，继续使用标准模式"
	}
}

func matchesScopeRelPath(relPath, scopeRelPath string) bool {
	scope := normalizeRelPath(scopeRelPath)
	current := normalizeRelPath(relPath)
	if scope == "" {
		return true
	}
	return current == scope || strings.HasPrefix(current, scope+"/")
}

func hasFavoritePath(groups []FavoriteGroup, relPath string, favoriteGroupID string) bool {
	target := normalizeRelPath(relPath)
	if target == "" {
		return false
	}
	for _, group := range groups {
		if favoriteGroupID != "" && group.ID != favoriteGroupID {
			continue
		}
		for _, path := range group.Paths {
			if normalizeRelPath(path) == target {
				return true
			}
		}
	}
	return false
}

func getManagedImagesCount(a *App) int {
	count := 0
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		count++
		return nil
	})
	return count
}

func sortImageFiles(images []ImageFile, sortBy, sortOrder string) {
	sort.Slice(images, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "time":
			less = images[i].ModTime < images[j].ModTime
		case "size":
			less = images[i].Size < images[j].Size
		case "name":
			less = images[i].Name < images[j].Name
		case "dimensions":
			less = (images[i].Width * images[i].Height) < (images[j].Width * images[j].Height)
		default:
			less = images[i].ModTime < images[j].ModTime
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

func aggregateWorkbenchFilterOptions(values []string) []WorkbenchFilterOption {
	type aggregateEntry struct {
		Value   string
		Label   string
		Count   int
		Aliases map[string]struct{}
	}

	grouped := make(map[string]*aggregateEntry)
	for _, raw := range values {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		key := normalizeAssetKey(trimmed)
		if key == "" {
			continue
		}

		label := prettifyAssetLabel(trimmed)
		if label == "" {
			label = trimmed
		}

		entry, exists := grouped[key]
		if !exists {
			entry = &aggregateEntry{
				Value:   key,
				Label:   label,
				Count:   0,
				Aliases: map[string]struct{}{},
			}
			grouped[key] = entry
		}

		entry.Count++
		entry.Aliases[trimmed] = struct{}{}
		if len(label) < len(entry.Label) {
			entry.Label = label
		}
	}

	options := make([]WorkbenchFilterOption, 0, len(grouped))
	for _, entry := range grouped {
		aliases := make([]string, 0, len(entry.Aliases))
		for alias := range entry.Aliases {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		options = append(options, WorkbenchFilterOption{
			Value:   entry.Value,
			Label:   entry.Label,
			Count:   entry.Count,
			Aliases: aliases,
		})
	}

	sort.Slice(options, func(i, j int) bool {
		if options[i].Count != options[j].Count {
			return options[i].Count > options[j].Count
		}
		return options[i].Label < options[j].Label
	})

	return options
}

func (a *App) GetImagesIndex(sortBy, sortOrder string) ([]ImageFile, error) {
	if !a.hasDirectoryBinding() {
		return []ImageFile{}, nil
	}

	images := make([]ImageFile, 0, 512)
	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		images = append(images, ImageFile{
			Name:        filepath.Base(path),
			Path:        relPath,
			ThumbPath:   a.imageVariantURL("thumb", relPath),
			PreviewPath: a.imageVariantURL("preview", relPath),
			RelPath:     relPath,
			ModTime:     info.ModTime().UTC().Format(time.RFC3339Nano),
			Size:        info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sortImageFiles(images, sortBy, sortOrder)
	return images, nil
}

func (a *App) GetImages(sortBy, sortOrder string) ([]ImageFile, error) {
	if !a.hasDirectoryBinding() {
		return []ImageFile{}, nil
	}

	a.ensureImageMetaCacheLoaded()
	cachedMeta := a.snapshotImageMetaCache()

	images := []ImageFile{}
	newCache := make(ImageMetaCache, len(cachedMeta))
	warmupTasks := make([]imageMetaWarmupTask, 0)
	autoRuleCandidates := make([]string, 0)
	cacheChanged := len(cachedMeta) == 0

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
		name := filepath.Base(path)
		width, height := 0, 0
		needsAutoRuleCheck := false
		shouldWarmupMetadata := false

		if cached, ok := cachedMeta[relPath]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
			shouldWarmupMetadata = !cached.MetadataScanned
			queuedWarmup := false
			if cached.Width > 0 || cached.Height > 0 {
				width = cached.Width
				height = cached.Height
			} else if sortBy == "dimensions" {
				width, height = readImageDimensions(path)
				cacheChanged = true
			} else {
				warmupTasks = append(warmupTasks, imageMetaWarmupTask{
					Path: path,
					Entry: ImageMetaCacheEntry{
						Name:    name,
						RelPath: relPath,
						ModTime: modTime,
						Size:    info.Size(),
					},
				})
				queuedWarmup = true
			}
			if shouldWarmupMetadata && !queuedWarmup {
				warmupTasks = append(warmupTasks, imageMetaWarmupTask{
					Path: path,
					Entry: ImageMetaCacheEntry{
						Name:    name,
						RelPath: relPath,
						ModTime: modTime,
						Size:    info.Size(),
					},
				})
			}
		} else if sortBy == "dimensions" {
			width, height = readImageDimensions(path)
			cacheChanged = true
			needsAutoRuleCheck = true
		} else {
			warmupTasks = append(warmupTasks, imageMetaWarmupTask{
				Path: path,
				Entry: ImageMetaCacheEntry{
					Name:    name,
					RelPath: relPath,
					ModTime: modTime,
					Size:    info.Size(),
				},
			})
			cacheChanged = true
			needsAutoRuleCheck = true
		}

		entry := ImageMetaCacheEntry{
			Name:    name,
			RelPath: relPath,
			ModTime: modTime,
			Size:    info.Size(),
			Width:   width,
			Height:  height,
		}
		if cached, ok := cachedMeta[relPath]; ok {
			entry.MetadataScanned = cached.MetadataScanned
			entry.HasMetadata = cached.HasMetadata
			entry.HasWorkflow = cached.HasWorkflow
			entry.Positive = cached.Positive
			entry.Negative = cached.Negative
			entry.Model = cached.Model
			entry.Sampler = cached.Sampler
			if len(cached.Loras) > 0 {
				entry.Loras = append([]string(nil), cached.Loras...)
			}
			entry.SearchText = cached.SearchText

			if cached.Name != entry.Name ||
				cached.RelPath != entry.RelPath ||
				cached.ModTime != entry.ModTime ||
				cached.Size != entry.Size ||
				cached.Width != entry.Width ||
				cached.Height != entry.Height {
				cacheChanged = true
			}
		} else {
			cacheChanged = true
		}
		if entry.SearchText == "" {
			entry.SearchText = buildImageSearchTextFromCacheEntry(entry)
			cacheChanged = true
		}
		newCache[relPath] = entry
		if needsAutoRuleCheck {
			autoRuleCandidates = append(autoRuleCandidates, relPath)
		}

		images = append(images, ImageFile{
			Name:        name,
			Path:        relPath,
			ThumbPath:   a.imageVariantURL("thumb", relPath),
			PreviewPath: a.imageVariantURL("preview", relPath),
			RelPath:     relPath,
			ModTime:     info.ModTime().UTC().Format(time.RFC3339Nano),
			Size:        info.Size(),
			Width:       width,
			Height:      height,
			Prompt:      entry.Positive,
			Model:       entry.Model,
			Loras:       append([]string(nil), entry.Loras...),
			SearchText:  entry.SearchText,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(newCache) != len(cachedMeta) {
		cacheChanged = true
	}

	a.replaceImageMetaCache(newCache)
	if cacheChanged {
		if err := a.saveImageMetaCache(newCache); err != nil {
			log.Printf("failed to save image metadata cache: %v", err)
		}
	}
	if sortBy != "dimensions" {
		a.scheduleImageMetaWarmup(warmupTasks)
	}
	if len(autoRuleCandidates) > 0 {
		a.scheduleAutoRulesRun(autoRuleCandidates)
	}

	sortImageFiles(images, sortBy, sortOrder)

	return images, nil
}

func (a *App) scheduleImageMetaWarmup(tasks []imageMetaWarmupTask) {
	if len(tasks) == 0 {
		return
	}

	a.imageMetaMu.Lock()
	if a.imageMetaWarmupRunning {
		a.imageMetaMu.Unlock()
		return
	}
	a.imageMetaWarmupRunning = true
	a.imageMetaMu.Unlock()

	go func(pending []imageMetaWarmupTask) {
		defer func() {
			a.imageMetaMu.Lock()
			a.imageMetaWarmupRunning = false
			a.imageMetaMu.Unlock()
		}()

		updated := false
		autoRuleCandidates := make([]string, 0)
		for _, task := range pending {
			entryNeedsMetadata := false
			a.imageMetaMu.RLock()
			entry, ok := a.imageMetaCache[task.Entry.RelPath]
			if ok && entry.ModTime == task.Entry.ModTime && entry.Size == task.Entry.Size {
				entryNeedsMetadata = !entry.MetadataScanned
			}
			a.imageMetaMu.RUnlock()

			if entryNeedsMetadata {
				metadata, err := a.GetImageMetadata(task.Entry.RelPath)
				if err == nil {
					updated = true
					if metadata.HasMetadata || strings.TrimSpace(metadata.Model) != "" || strings.TrimSpace(metadata.Sampler) != "" || len(metadata.Loras) > 0 {
						autoRuleCandidates = append(autoRuleCandidates, task.Entry.RelPath)
					}
				}
				continue
			}

			width, height := readImageDimensions(task.Path)
			if width == 0 && height == 0 {
				continue
			}

			a.imageMetaMu.Lock()
			entry, ok = a.imageMetaCache[task.Entry.RelPath]
			if ok && entry.ModTime == task.Entry.ModTime && entry.Size == task.Entry.Size {
				if entry.Width != width || entry.Height != height {
					entry.Width = width
					entry.Height = height
					a.imageMetaCache[task.Entry.RelPath] = entry
					updated = true
				}
			}
			a.imageMetaMu.Unlock()
		}

		if updated {
			if err := a.saveImageMetaCache(a.snapshotImageMetaCache()); err != nil {
				log.Printf("failed to save warmed image metadata cache: %v", err)
			}
			a.scheduleImagesChangedEvent()
		}
		if len(autoRuleCandidates) > 0 {
			a.scheduleAutoRulesRun(autoRuleCandidates)
		}
	}(tasks)
}

func shouldUseLightweightPagedScan(query GetImagesPageQuery, settings GalleryPerformanceSettings) bool {
	if !settings.MetadataLazy {
		return false
	}
	if strings.TrimSpace(query.SearchQuery) != "" {
		return false
	}
	if strings.TrimSpace(query.ActiveModelFilter) != "" || strings.TrimSpace(query.ActiveLoraFilter) != "" {
		return false
	}
	if strings.TrimSpace(query.SortBy) == "dimensions" {
		return false
	}
	return true
}

func (a *App) buildPagedResultFromItems(items []ImageFile, query GetImagesPageQuery, mode, reason string) GetImagesPageResult {
	total := len(items)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	if totalPages == 0 {
		query.Page = 1
	} else if query.Page > totalPages {
		query.Page = totalPages
	}

	startIndex := (query.Page - 1) * query.PageSize
	if startIndex < 0 {
		startIndex = 0
	}
	endIndex := startIndex + query.PageSize
	if endIndex > total {
		endIndex = total
	}

	pageItems := []ImageFile{}
	if startIndex < total && startIndex < endIndex {
		pageItems = items[startIndex:endIndex]
	}

	return GetImagesPageResult{
		Items:      pageItems,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
		HasMore:    endIndex < total,
		Mode:       mode,
		ModeReason: reason,
	}
}

func (a *App) getImagesPageLightweight(query GetImagesPageQuery, settings GalleryPerformanceSettings, mode, reason string) (GetImagesPageResult, error) {
	a.ensureImageMetaCacheLoaded()
	cachedMeta := a.snapshotImageMetaCache()
	newCache := make(ImageMetaCache, len(cachedMeta))
	imageTags, _ := a.loadImageTags()
	favoriteGroups, _ := a.loadFavoriteGroups()

	filtered := make([]ImageFile, 0, 256)
	warmupTaskMap := make(map[string]imageMetaWarmupTask)
	cacheChanged := len(cachedMeta) == 0

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
		name := filepath.Base(path)

		entry := ImageMetaCacheEntry{
			Name:    name,
			RelPath: relPath,
			ModTime: modTime,
			Size:    info.Size(),
		}

		taskNeeded := false
		if cached, ok := cachedMeta[relPath]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
			entry.Width = cached.Width
			entry.Height = cached.Height
			entry.MetadataScanned = cached.MetadataScanned
			entry.HasMetadata = cached.HasMetadata
			entry.HasWorkflow = cached.HasWorkflow
			entry.Positive = cached.Positive
			entry.Negative = cached.Negative
			entry.Model = cached.Model
			entry.Sampler = cached.Sampler
			if len(cached.Loras) > 0 {
				entry.Loras = append([]string(nil), cached.Loras...)
			}
			entry.SearchText = cached.SearchText
			if entry.Width == 0 && entry.Height == 0 {
				taskNeeded = true
			} else if !entry.MetadataScanned {
				taskNeeded = true
			}
		} else {
			cacheChanged = true
			taskNeeded = true
		}

		if entry.SearchText == "" {
			entry.SearchText = buildImageSearchTextFromCacheEntry(entry)
			cacheChanged = true
		}
		newCache[relPath] = entry

		if taskNeeded {
			warmupTaskMap[relPath] = imageMetaWarmupTask{
				Path: path,
				Entry: ImageMetaCacheEntry{
					Name:    name,
					RelPath: relPath,
					ModTime: modTime,
					Size:    info.Size(),
				},
			}
		}

		if !matchesScopeRelPath(relPath, query.ScopeRelPath) {
			return nil
		}
		if query.FavoritesOnly && !hasFavoritePath(favoriteGroups, relPath, query.FavoriteGroupID) {
			return nil
		}
		if query.ActiveTagID != "" && !contains(imageTags[relPath], query.ActiveTagID) {
			return nil
		}

		dateKey := extractDateKeyFromRelPath(relPath, info.ModTime())
		if !matchesDatePreset(dateKey, query.ActiveDatePreset, query.ActiveDateStart, query.ActiveDateEnd) {
			return nil
		}

		filtered = append(filtered, ImageFile{
			Name:        name,
			Path:        relPath,
			ThumbPath:   a.imageVariantURL("thumb", relPath),
			PreviewPath: a.imageVariantURL("preview", relPath),
			RelPath:     relPath,
			ModTime:     modTime,
			Size:        info.Size(),
			Width:       entry.Width,
			Height:      entry.Height,
			Prompt:      entry.Positive,
			Model:       entry.Model,
			Loras:       append([]string(nil), entry.Loras...),
			SearchText:  entry.SearchText,
		})
		return nil
	})
	if err != nil {
		return GetImagesPageResult{}, err
	}

	if len(newCache) != len(cachedMeta) {
		cacheChanged = true
	}

	a.replaceImageMetaCache(newCache)
	if cacheChanged {
		if err := a.saveImageMetaCache(newCache); err != nil {
			log.Printf("failed to save lightweight image metadata cache: %v", err)
		}
	}

	sortImageFiles(filtered, query.SortBy, query.SortOrder)
	result := a.buildPagedResultFromItems(filtered, query, mode, reason)

	if len(result.Items) > 0 {
		warmupTasks := make([]imageMetaWarmupTask, 0, len(result.Items))
		for _, item := range result.Items {
			if task, ok := warmupTaskMap[item.RelPath]; ok {
				warmupTasks = append(warmupTasks, task)
			}
		}
		a.scheduleImageMetaWarmup(warmupTasks)
	}
	a.warmImageVariantsAsync(result.Items, settings)

	return result, nil
}

func (a *App) GetWorkbenchAggregate(query WorkbenchSummaryQuery) (WorkbenchAggregateResult, error) {
	if !a.hasDirectoryBinding() {
		return WorkbenchAggregateResult{
			AvailableModels: []WorkbenchFilterOption{},
			AvailableLoras:  []WorkbenchFilterOption{},
			Summary:         WorkbenchSummary{RecentDates: []WorkbenchRecentDate{}},
			FilteredCount:   0,
		}, nil
	}

	a.ensureImageMetaCacheLoaded()
	cachedMeta := a.snapshotImageMetaCache()
	newCache := make(ImageMetaCache, len(cachedMeta))
	warmupTaskMap := make(map[string]imageMetaWarmupTask)
	modelValues := make([]string, 0, 256)
	loraValues := make([]string, 0, 256)
	dateCountMap := make(map[string]int)
	cacheChanged := len(cachedMeta) == 0
	filteredCount := 0
	datedTotal := 0
	todayCount := 0
	yesterdayCount := 0
	last7Count := 0
	monthCount := 0

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
		name := filepath.Base(path)

		entry := ImageMetaCacheEntry{
			Name:    name,
			RelPath: relPath,
			ModTime: modTime,
			Size:    info.Size(),
		}

		taskNeeded := false
		if cached, ok := cachedMeta[relPath]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
			entry.Width = cached.Width
			entry.Height = cached.Height
			entry.MetadataScanned = cached.MetadataScanned
			entry.HasMetadata = cached.HasMetadata
			entry.HasWorkflow = cached.HasWorkflow
			entry.Positive = cached.Positive
			entry.Negative = cached.Negative
			entry.Model = cached.Model
			entry.Sampler = cached.Sampler
			if len(cached.Loras) > 0 {
				entry.Loras = append([]string(nil), cached.Loras...)
			}
			entry.SearchText = cached.SearchText
			if entry.Width == 0 && entry.Height == 0 {
				taskNeeded = true
			} else if !entry.MetadataScanned {
				taskNeeded = true
			}
		} else {
			cacheChanged = true
			taskNeeded = true
		}

		if entry.SearchText == "" {
			entry.SearchText = buildImageSearchTextFromCacheEntry(entry)
			cacheChanged = true
		}
		newCache[relPath] = entry

		if taskNeeded {
			warmupTaskMap[relPath] = imageMetaWarmupTask{
				Path: path,
				Entry: ImageMetaCacheEntry{
					Name:    name,
					RelPath: relPath,
					ModTime: modTime,
					Size:    info.Size(),
				},
			}
		}

		if strings.TrimSpace(entry.Model) != "" {
			modelValues = append(modelValues, entry.Model)
		}
		if len(entry.Loras) > 0 {
			loraValues = append(loraValues, entry.Loras...)
		}

		dateKey := extractDateKeyFromRelPath(relPath, info.ModTime().In(time.Local))

		modelMatched := true
		if query.ActiveModelFilter != "" {
			modelMatched = normalizeAssetKey(entry.Model) == normalizeAssetKey(query.ActiveModelFilter)
		}
		loraMatched := true
		if query.ActiveLoraFilter != "" {
			target := normalizeAssetKey(query.ActiveLoraFilter)
			loraMatched = false
			for _, lora := range entry.Loras {
				if normalizeAssetKey(lora) == target {
					loraMatched = true
					break
				}
			}
		}

		if modelMatched && loraMatched {
			dateMatched := matchesDatePreset(dateKey, query.ActiveDatePreset, query.ActiveDateStart, query.ActiveDateEnd)
			if dateMatched {
				filteredCount++
				if dateKey != "" {
					dateCountMap[dateKey]++
					datedTotal++
				}
			}
			if matchesDatePreset(dateKey, "today", "", "") {
				todayCount++
			}
			if matchesDatePreset(dateKey, "yesterday", "", "") {
				yesterdayCount++
			}
			if matchesDatePreset(dateKey, "last7", "", "") {
				last7Count++
			}
			if matchesDatePreset(dateKey, "month", "", "") {
				monthCount++
			}
		}

		return nil
	})
	if err != nil {
		return WorkbenchAggregateResult{}, err
	}

	if len(newCache) != len(cachedMeta) {
		cacheChanged = true
	}
	a.replaceImageMetaCache(newCache)
	if cacheChanged {
		if err := a.saveImageMetaCache(newCache); err != nil {
			log.Printf("failed to save workbench image metadata cache: %v", err)
		}
	}

	if len(warmupTaskMap) > 0 {
		warmupTasks := make([]imageMetaWarmupTask, 0, len(warmupTaskMap))
		for _, task := range warmupTaskMap {
			warmupTasks = append(warmupTasks, task)
		}
		a.scheduleImageMetaWarmup(warmupTasks)
	}

	recentDates := make([]WorkbenchRecentDate, 0, len(dateCountMap))
	for dateKey, count := range dateCountMap {
		recentDates = append(recentDates, WorkbenchRecentDate{
			Date:  dateKey,
			Count: count,
		})
	}
	sort.Slice(recentDates, func(i, j int) bool {
		return recentDates[i].Date > recentDates[j].Date
	})
	if len(recentDates) > 12 {
		recentDates = recentDates[:12]
	}

	return WorkbenchAggregateResult{
		AvailableModels: aggregateWorkbenchFilterOptions(modelValues),
		AvailableLoras:  aggregateWorkbenchFilterOptions(loraValues),
		Summary: WorkbenchSummary{
			Total:       filteredCount,
			DatedTotal:  datedTotal,
			Today:       todayCount,
			Yesterday:   yesterdayCount,
			Last7:       last7Count,
			Month:       monthCount,
			RecentDates: recentDates,
		},
		FilteredCount: filteredCount,
	}, nil
}

func (a *App) GetImagesPage(query GetImagesPageQuery) (GetImagesPageResult, error) {
	if !a.hasDirectoryBinding() {
		return GetImagesPageResult{
			Items:      []ImageFile{},
			Total:      0,
			Page:       1,
			PageSize:   1,
			TotalPages: 0,
			HasMore:    false,
			Mode:       "standard",
		}, nil
	}

	if strings.TrimSpace(query.SortBy) == "" {
		query.SortBy = "time"
	}
	if strings.TrimSpace(query.SortOrder) == "" {
		query.SortOrder = "desc"
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = defaultGalleryPerformanceSettings().PageSize
	}
	if query.PageSize > 500 {
		query.PageSize = 500
	}

	settings, _ := a.loadSettings()
	performanceSettings := settingsToGalleryPerformanceSettings(settings)
	mode, reason := resolveGalleryLoadMode(performanceSettings, getManagedImagesCount(a))
	if shouldUseLightweightPagedScan(query, performanceSettings) {
		return a.getImagesPageLightweight(query, performanceSettings, mode, reason)
	}

	images, err := a.GetImages(query.SortBy, query.SortOrder)
	if err != nil {
		return GetImagesPageResult{}, err
	}

	imageTags, _ := a.loadImageTags()
	notes, _ := a.loadImageNotes()
	tags, _ := a.loadTags()
	tagNameMap := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagNameMap[tag.ID] = tag.Name
	}
	favoriteGroups, _ := a.loadFavoriteGroups()

	filtered := make([]ImageFile, 0, len(images))
	for _, img := range images {
		if !matchesScopeRelPath(img.RelPath, query.ScopeRelPath) {
			continue
		}
		if query.FavoritesOnly && !hasFavoritePath(favoriteGroups, img.RelPath, query.FavoriteGroupID) {
			continue
		}
		if query.ActiveTagID != "" && !contains(imageTags[img.RelPath], query.ActiveTagID) {
			continue
		}
		if query.ActiveModelFilter != "" && normalizeAssetKey(img.Model) != normalizeAssetKey(query.ActiveModelFilter) {
			continue
		}
		if query.ActiveLoraFilter != "" {
			target := normalizeAssetKey(query.ActiveLoraFilter)
			matched := false
			for _, lora := range img.Loras {
				if normalizeAssetKey(lora) == target {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		modTime, _ := time.Parse(time.RFC3339Nano, img.ModTime)
		modTime = modTime.In(time.Local)
		dateKey := extractDateKeyFromRelPath(img.RelPath, modTime)
		if !matchesDatePreset(dateKey, query.ActiveDatePreset, query.ActiveDateStart, query.ActiveDateEnd) {
			continue
		}

		normalizedQuery := normalizeSearchValue(query.SearchQuery)
		if normalizedQuery != "" {
			tagTexts := make([]string, 0)
			for _, tagID := range imageTags[img.RelPath] {
				if name := strings.TrimSpace(tagNameMap[tagID]); name != "" {
					tagTexts = append(tagTexts, name)
				}
			}
			searchParts := []string{
				img.Name,
				img.RelPath,
				img.Prompt,
				img.Model,
				img.SearchText,
				notes[img.RelPath],
				strings.Join(img.Loras, " "),
				strings.Join(tagTexts, " "),
			}
			matched := false
			for _, part := range searchParts {
				if strings.Contains(normalizeSearchValue(part), normalizedQuery) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, img)
	}

	result := a.buildPagedResultFromItems(filtered, query, mode, reason)
	a.warmImageVariantsAsync(result.Items, performanceSettings)
	return result, nil
}
