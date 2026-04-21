package backend

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

func autoRulePromptText(metadata ImageMetadata) string {
	if strings.TrimSpace(metadata.Positive) != "" {
		return metadata.Positive
	}
	return metadata.Prompt
}

func buildImageSearchText(name, relPath string, width, height int, metadata ImageMetadata) string {
	parts := []string{
		name,
		relPath,
		metadata.Model,
		metadata.Sampler,
		autoRulePromptText(metadata),
		metadata.Negative,
	}
	if width > 0 && height > 0 {
		parts = append(parts, fmt.Sprintf("%dx%d", width, height))
	}
	parts = append(parts, metadata.Loras...)
	return strings.Join(uniqueNonEmptyStrings(parts), "\n")
}

func buildImageSearchTextFromCacheEntry(entry ImageMetaCacheEntry) string {
	metadata := ImageMetadata{
		RelPath:     entry.RelPath,
		Width:       entry.Width,
		Height:      entry.Height,
		HasMetadata: entry.HasMetadata,
		Prompt:      entry.Positive,
		Positive:    entry.Positive,
		Negative:    entry.Negative,
		Model:       entry.Model,
		Sampler:     entry.Sampler,
		Loras:       append([]string(nil), entry.Loras...),
	}
	return buildImageSearchText(entry.Name, entry.RelPath, entry.Width, entry.Height, metadata)
}

func normalizeAutoRuleCondition(condition AutoRuleCondition) AutoRuleCondition {
	field := strings.ToLower(strings.TrimSpace(condition.Field))
	switch field {
	case "model", "sampler", "lora", "dimensions", "filename", "prompt", "negative":
	default:
		field = "model"
	}

	operator := strings.ToLower(strings.TrimSpace(condition.Operator))
	switch operator {
	case "contains", "equals", "starts_with", "ends_with":
	default:
		operator = "contains"
	}

	return AutoRuleCondition{
		Field:    field,
		Operator: operator,
		Value:    repairLegacyMojibake(condition.Value),
	}
}

func normalizeAutoRuleAction(action AutoRuleAction) AutoRuleAction {
	actionType := strings.ToLower(strings.TrimSpace(action.Type))
	switch actionType {
	case "add_tag", "add_favorite_group", "move_to_folder":
	default:
		actionType = "add_tag"
	}

	return AutoRuleAction{
		Type:  actionType,
		Value: repairLegacyMojibake(action.Value),
	}
}

func normalizeAutoRule(rule AutoRule) AutoRule {
	normalized := AutoRule{
		ID:             strings.TrimSpace(rule.ID),
		Name:           repairLegacyMojibake(rule.Name),
		Enabled:        rule.Enabled,
		MatchMode:      strings.ToLower(strings.TrimSpace(rule.MatchMode)),
		LastRunAt:      strings.TrimSpace(rule.LastRunAt),
		LastMatchCount: rule.LastMatchCount,
		LastStatus:     strings.TrimSpace(rule.LastStatus),
		LastError:      repairLegacyMojibake(rule.LastError),
		CreatedAt:      strings.TrimSpace(rule.CreatedAt),
		UpdatedAt:      strings.TrimSpace(rule.UpdatedAt),
	}

	if normalized.MatchMode != "any" {
		normalized.MatchMode = "all"
	}
	if normalized.LastStatus == "" {
		normalized.LastStatus = "idle"
	}

	for _, condition := range rule.Conditions {
		next := normalizeAutoRuleCondition(condition)
		if next.Value == "" {
			continue
		}
		normalized.Conditions = append(normalized.Conditions, next)
	}

	for _, action := range rule.Actions {
		next := normalizeAutoRuleAction(action)
		if next.Value == "" {
			continue
		}
		normalized.Actions = append(normalized.Actions, next)
	}

	return normalized
}

func normalizeAutoRulesStore(store AutoRulesStore) AutoRulesStore {
	normalized := AutoRulesStore{
		Enabled: store.Enabled,
		Rules:   make([]AutoRule, 0, len(store.Rules)),
	}
	for _, rule := range store.Rules {
		next := normalizeAutoRule(rule)
		if next.Name == "" || len(next.Conditions) == 0 || len(next.Actions) == 0 {
			continue
		}
		normalized.Rules = append(normalized.Rules, next)
	}
	return normalized
}

func defaultAutoRulesStore() AutoRulesStore {
	now := time.Now().Format(time.RFC3339)
	return AutoRulesStore{
		Enabled: true,
		Rules: []AutoRule{
			{
				ID:        "default-rule-pony-tag",
				Name:      "Pony Auto Tag",
				Enabled:   true,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "model", Operator: "contains", Value: "pony"},
				},
				Actions: []AutoRuleAction{
					{Type: "add_tag", Value: "Pony"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:        "default-rule-sdxl-tag",
				Name:      "SDXL Auto Tag",
				Enabled:   true,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "model", Operator: "contains", Value: "sdxl"},
				},
				Actions: []AutoRuleAction{
					{Type: "add_tag", Value: "SDXL"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:        "default-rule-detail-lora",
				Name:      "Detail LoRA Auto Tag",
				Enabled:   true,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "lora", Operator: "contains", Value: "detail"},
				},
				Actions: []AutoRuleAction{
					{Type: "add_tag", Value: "Detail Boost"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:        "default-rule-euler-tag",
				Name:      "Euler a Auto Tag",
				Enabled:   true,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "sampler", Operator: "equals", Value: "Euler a"},
				},
				Actions: []AutoRuleAction{
					{Type: "add_tag", Value: "Euler a"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:        "default-rule-portrait-folder",
				Name:      "Portrait Folder Rule",
				Enabled:   false,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "dimensions", Operator: "equals", Value: "1024x1536"},
				},
				Actions: []AutoRuleAction{
					{Type: "move_to_folder", Value: "Auto Archive/Portrait"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			{
				ID:        "default-rule-landscape-folder",
				Name:      "Landscape Folder Rule",
				Enabled:   false,
				MatchMode: "all",
				Conditions: []AutoRuleCondition{
					{Field: "dimensions", Operator: "equals", Value: "1536x1024"},
				},
				Actions: []AutoRuleAction{
					{Type: "move_to_folder", Value: "Auto Archive/Landscape"},
				},
				LastStatus: "idle",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}
}

func (a *App) loadAutoRulesStoreUnlocked() (AutoRulesStore, error) {
	data, err := os.ReadFile(a.autoRulesFile())
	if err != nil {
		if os.IsNotExist(err) {
			store := defaultAutoRulesStore()
			if saveErr := a.saveAutoRulesStoreUnlocked(store); saveErr != nil {
				return store, nil
			}
			return store, nil
		}
		return AutoRulesStore{}, err
	}
	data = stripUTF8BOM(data)

	store := AutoRulesStore{}
	if err := json.Unmarshal(data, &store); err != nil {
		return AutoRulesStore{}, err
	}
	normalized := normalizeAutoRulesStore(store)
	if !reflect.DeepEqual(normalized, store) {
		_ = a.saveAutoRulesStoreUnlocked(normalized)
	}
	return normalized, nil
}

func (a *App) saveAutoRulesStoreUnlocked(store AutoRulesStore) error {
	normalized := normalizeAutoRulesStore(store)
	data, _ := json.MarshalIndent(normalized, "", "  ")
	return os.WriteFile(a.autoRulesFile(), data, 0644)
}

func (a *App) GetAutoRules() (AutoRulesStore, error) {
	a.autoRulesMu.Lock()
	defer a.autoRulesMu.Unlock()

	return a.loadAutoRulesStoreUnlocked()
}

func (a *App) SetAutoRulesEnabled(enabled bool) (AutoRulesStore, error) {
	a.autoRulesMu.Lock()
	defer a.autoRulesMu.Unlock()

	store, err := a.loadAutoRulesStoreUnlocked()
	if err != nil {
		return AutoRulesStore{}, err
	}
	store.Enabled = enabled
	if err := a.saveAutoRulesStoreUnlocked(store); err != nil {
		return AutoRulesStore{}, err
	}
	return normalizeAutoRulesStore(store), nil
}

func (a *App) CreateAutoRule(rule AutoRule) (AutoRule, error) {
	a.autoRulesMu.Lock()
	defer a.autoRulesMu.Unlock()

	store, err := a.loadAutoRulesStoreUnlocked()
	if err != nil {
		return AutoRule{}, err
	}

	normalized := normalizeAutoRule(rule)
	if normalized.Name == "" {
		return AutoRule{}, fmt.Errorf("rule name is required")
	}
	if len(normalized.Conditions) == 0 {
		return AutoRule{}, fmt.Errorf("at least one condition is required")
	}
	if len(normalized.Actions) == 0 {
		return AutoRule{}, fmt.Errorf("at least one action is required")
	}
	now := time.Now().Format(time.RFC3339)
	normalized.ID = uuid.New().String()
	normalized.CreatedAt = now
	normalized.UpdatedAt = now
	normalized.LastStatus = "idle"

	store.Rules = append([]AutoRule{normalized}, store.Rules...)
	if err := a.saveAutoRulesStoreUnlocked(store); err != nil {
		return AutoRule{}, err
	}
	return normalized, nil
}

func (a *App) UpdateAutoRule(rule AutoRule) (AutoRule, error) {
	a.autoRulesMu.Lock()
	defer a.autoRulesMu.Unlock()

	store, err := a.loadAutoRulesStoreUnlocked()
	if err != nil {
		return AutoRule{}, err
	}

	normalized := normalizeAutoRule(rule)
	if normalized.ID == "" {
		return AutoRule{}, fmt.Errorf("rule id is required")
	}
	if normalized.Name == "" {
		return AutoRule{}, fmt.Errorf("rule name is required")
	}
	if len(normalized.Conditions) == 0 {
		return AutoRule{}, fmt.Errorf("at least one condition is required")
	}
	if len(normalized.Actions) == 0 {
		return AutoRule{}, fmt.Errorf("at least one action is required")
	}

	found := false
	for i := range store.Rules {
		if store.Rules[i].ID != normalized.ID {
			continue
		}
		normalized.CreatedAt = store.Rules[i].CreatedAt
		normalized.LastRunAt = store.Rules[i].LastRunAt
		normalized.LastMatchCount = store.Rules[i].LastMatchCount
		normalized.LastStatus = store.Rules[i].LastStatus
		normalized.LastError = store.Rules[i].LastError
		normalized.UpdatedAt = time.Now().Format(time.RFC3339)
		store.Rules[i] = normalized
		found = true
		break
	}
	if !found {
		return AutoRule{}, fmt.Errorf("rule not found")
	}

	if err := a.saveAutoRulesStoreUnlocked(store); err != nil {
		return AutoRule{}, err
	}
	return normalized, nil
}

func (a *App) DeleteAutoRule(id string) error {
	a.autoRulesMu.Lock()
	defer a.autoRulesMu.Unlock()

	store, err := a.loadAutoRulesStoreUnlocked()
	if err != nil {
		return err
	}

	filtered := make([]AutoRule, 0, len(store.Rules))
	found := false
	for _, rule := range store.Rules {
		if rule.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, rule)
	}
	if !found {
		return fmt.Errorf("rule not found")
	}

	store.Rules = filtered
	return a.saveAutoRulesStoreUnlocked(store)
}

type autoRuleMatchContext struct {
	RelPath  string
	FileName string
	Width    int
	Height   int
	Metadata ImageMetadata
}

func autoRuleFieldValue(ctx autoRuleMatchContext, field string) string {
	switch field {
	case "model":
		return ctx.Metadata.Model
	case "sampler":
		return ctx.Metadata.Sampler
	case "lora":
		return strings.Join(ctx.Metadata.Loras, "\n")
	case "dimensions":
		if ctx.Width > 0 && ctx.Height > 0 {
			return fmt.Sprintf("%dx%d", ctx.Width, ctx.Height)
		}
		return ""
	case "filename":
		return ctx.FileName
	case "prompt":
		return autoRulePromptText(ctx.Metadata)
	case "negative":
		return ctx.Metadata.Negative
	default:
		return ""
	}
}

func matchesAutoRuleCondition(ctx autoRuleMatchContext, condition AutoRuleCondition) bool {
	left := normalizeSearchValue(autoRuleFieldValue(ctx, condition.Field))
	right := normalizeSearchValue(condition.Value)
	if left == "" || right == "" {
		return false
	}

	switch condition.Operator {
	case "equals":
		return left == right
	case "starts_with":
		return strings.HasPrefix(left, right)
	case "ends_with":
		return strings.HasSuffix(left, right)
	default:
		return strings.Contains(left, right)
	}
}

func matchesAutoRule(ctx autoRuleMatchContext, rule AutoRule) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	if rule.MatchMode == "any" {
		for _, condition := range rule.Conditions {
			if matchesAutoRuleCondition(ctx, condition) {
				return true
			}
		}
		return false
	}

	for _, condition := range rule.Conditions {
		if !matchesAutoRuleCondition(ctx, condition) {
			return false
		}
	}
	return true
}

func hasMoveAction(actions []AutoRuleAction) bool {
	for _, action := range actions {
		if action.Type == "move_to_folder" {
			return true
		}
	}
	return false
}

func (a *App) ensureTagAssigned(relPath, tagName string) (bool, error) {
	normalizedPath := normalizeRelPath(relPath)
	trimmedName := strings.TrimSpace(tagName)
	if normalizedPath == "" || trimmedName == "" {
		return false, fmt.Errorf("tag target is invalid")
	}

	tagMutex.Lock()
	defer tagMutex.Unlock()

	tags, err := a.loadTags()
	if err != nil {
		return false, err
	}

	tagID := ""
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag.Name), trimmedName) {
			tagID = tag.ID
			break
		}
	}
	if tagID == "" {
		newTag := Tag{
			ID:       uuid.New().String(),
			Name:     trimmedName,
			Color:    "#64748b",
			Category: "default",
		}
		tags = append(tags, newTag)
		if err := a.saveTags(tags); err != nil {
			return false, err
		}
		tagID = newTag.ID
	}

	imageTags, err := a.loadImageTags()
	if err != nil {
		return false, err
	}
	current := imageTags[normalizedPath]
	if contains(current, tagID) {
		return false, nil
	}
	imageTags[normalizedPath] = append(current, tagID)
	return true, a.saveImageTags(imageTags)
}

func (a *App) ensureFavoriteGroupAssigned(relPath, groupName string) (bool, error) {
	normalizedPath := normalizeRelPath(relPath)
	trimmedName := strings.TrimSpace(groupName)
	if normalizedPath == "" || trimmedName == "" {
		return false, fmt.Errorf("favorite group target is invalid")
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return false, err
	}

	index := -1
	for i, group := range groups {
		if strings.EqualFold(strings.TrimSpace(group.Name), trimmedName) {
			index = i
			break
		}
	}
	if index < 0 {
		groups = append(groups, FavoriteGroup{
			ID:    uuid.New().String(),
			Name:  trimmedName,
			Paths: []string{},
		})
		index = len(groups) - 1
	}

	if contains(groups[index].Paths, normalizedPath) {
		if len(groups) == 1 {
			return false, nil
		}
		return false, a.saveFavoriteGroups(groups)
	}

	groups[index].Paths = append(groups[index].Paths, normalizedPath)
	groups[index].Paths = uniqueNonEmptyStrings(groups[index].Paths)
	return true, a.saveFavoriteGroups(groups)
}

func (a *App) remapManagedImageReferences(oldRelPath, newRelPath string) error {
	oldRelPath = normalizeRelPath(oldRelPath)
	newRelPath = normalizeRelPath(newRelPath)
	if oldRelPath == "" || newRelPath == "" || oldRelPath == newRelPath {
		return nil
	}

	imageTags, err := a.loadImageTags()
	if err != nil {
		return err
	}
	if tagIDs, ok := imageTags[oldRelPath]; ok {
		imageTags[newRelPath] = uniqueNonEmptyStrings(append(imageTags[newRelPath], tagIDs...))
		delete(imageTags, oldRelPath)
		if err := a.saveImageTags(imageTags); err != nil {
			return err
		}
	}

	notes, err := a.loadImageNotes()
	if err != nil {
		return err
	}
	if note, ok := notes[oldRelPath]; ok {
		if strings.TrimSpace(notes[newRelPath]) == "" {
			notes[newRelPath] = note
		}
		delete(notes, oldRelPath)
		if err := a.saveImageNotes(notes); err != nil {
			return err
		}
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}
	groupsChanged := false
	for i := range groups {
		replaced := false
		paths := make([]string, 0, len(groups[i].Paths))
		for _, path := range groups[i].Paths {
			if path == oldRelPath {
				paths = append(paths, newRelPath)
				replaced = true
				continue
			}
			paths = append(paths, path)
		}
		if replaced {
			groups[i].Paths = uniqueNonEmptyStrings(paths)
			groupsChanged = true
		}
	}
	if groupsChanged {
		if err := a.saveFavoriteGroups(groups); err != nil {
			return err
		}
	}

	a.ensureImageMetaCacheLoaded()
	cache := a.snapshotImageMetaCache()
	if entry, ok := cache[oldRelPath]; ok {
		delete(cache, oldRelPath)
		entry.RelPath = newRelPath
		entry.Name = filepath.Base(newRelPath)
		cache[newRelPath] = entry
		a.replaceImageMetaCache(cache)
		if err := a.saveImageMetaCache(cache); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) moveManagedImageToFolder(relPath, targetFolder string) (string, bool, error) {
	normalizedPath := normalizeRelPath(relPath)
	if normalizedPath == "" {
		return "", false, fmt.Errorf("invalid image path")
	}

	sourcePath, err := a.resolveRootPath(normalizedPath)
	if err != nil {
		return "", false, err
	}
	targetFolder = normalizeRelPath(targetFolder)

	targetPath, err := a.resolveRootPath(targetFolder)
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return "", false, err
	}

	fileName := filepath.Base(sourcePath)
	destPath := filepath.Join(targetPath, fileName)
	if samePath(sourcePath, destPath) {
		return normalizedPath, false, nil
	}

	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		timestamp := time.Now().Format("20060102_150405")
		destPath = filepath.Join(targetPath, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := moveFile(sourcePath, destPath); err != nil {
		return "", false, err
	}

	relative, err := filepath.Rel(a.rootDir, destPath)
	if err != nil {
		return "", false, err
	}
	newRelPath := normalizeRelPath(relative)
	if err := a.remapManagedImageReferences(normalizedPath, newRelPath); err != nil {
		return "", false, err
	}

	return newRelPath, true, nil
}

func (a *App) updateImageMetaCacheEntry(relPath string, info fs.FileInfo, metadata ImageMetadata) error {
	normalized := normalizeRelPath(relPath)
	if normalized == "" || info == nil {
		return nil
	}

	a.ensureImageMetaCacheLoaded()
	cache := a.snapshotImageMetaCache()
	entry := cache[normalized]
	entry.Name = filepath.Base(normalized)
	entry.RelPath = normalized
	entry.ModTime = info.ModTime().UTC().Format(time.RFC3339Nano)
	entry.Size = info.Size()
	entry.Width = metadata.Width
	entry.Height = metadata.Height
	entry.MetadataScanned = true
	entry.HasMetadata = metadata.HasMetadata
	entry.HasWorkflow = strings.TrimSpace(metadata.Workflow) != ""
	entry.Positive = metadata.Positive
	entry.Negative = metadata.Negative
	entry.Model = metadata.Model
	entry.Sampler = metadata.Sampler
	entry.Loras = append([]string(nil), metadata.Loras...)
	entry.SearchText = buildImageSearchText(entry.Name, normalized, metadata.Width, metadata.Height, metadata)
	cache[normalized] = entry
	a.replaceImageMetaCache(cache)
	return a.saveImageMetaCache(cache)
}

func (a *App) buildAutoRuleMatchContext(relPath string) (autoRuleMatchContext, error) {
	metadata, err := a.GetImageMetadata(relPath)
	if err != nil {
		return autoRuleMatchContext{}, err
	}
	return autoRuleMatchContext{
		RelPath:  metadata.RelPath,
		FileName: filepath.Base(metadata.RelPath),
		Width:    metadata.Width,
		Height:   metadata.Height,
		Metadata: metadata,
	}, nil
}

func (a *App) executeAutoRuleActions(relPath string, actions []AutoRuleAction) (string, bool, bool, error) {
	currentPath := normalizeRelPath(relPath)
	updated := false
	stopAfterRule := false

	for _, action := range actions {
		switch action.Type {
		case "add_tag":
			changed, err := a.ensureTagAssigned(currentPath, action.Value)
			if err != nil {
				return currentPath, updated, stopAfterRule, err
			}
			updated = updated || changed
		case "add_favorite_group":
			changed, err := a.ensureFavoriteGroupAssigned(currentPath, action.Value)
			if err != nil {
				return currentPath, updated, stopAfterRule, err
			}
			updated = updated || changed
		case "move_to_folder":
			nextPath, moved, err := a.moveManagedImageToFolder(currentPath, action.Value)
			if err != nil {
				return currentPath, updated, true, err
			}
			currentPath = nextPath
			updated = updated || moved
			stopAfterRule = true
		default:
			return currentPath, updated, stopAfterRule, fmt.Errorf("unsupported action: %s", action.Type)
		}
	}

	if hasMoveAction(actions) {
		stopAfterRule = true
	}

	return currentPath, updated, stopAfterRule, nil
}

func (a *App) runAutoRulesForPaths(paths []string, source string) (AutoRulesRunSummary, error) {
	summary := AutoRulesRunSummary{
		RanAt: time.Now().Format(time.RFC3339),
	}

	normalizedPaths := uniqueNonEmptyStrings(paths)
	summary.TotalCount = len(normalizedPaths)
	emitProgress := func(stage string, running bool, currentRelPath, currentRuleName, message string) {
		a.emitAutoRulesProgress(AutoRulesRunProgress{
			Source:          source,
			Stage:           stage,
			Running:         running,
			TotalCount:      summary.TotalCount,
			ProcessedCount:  summary.ProcessedCount,
			MatchedCount:    summary.MatchedCount,
			UpdatedCount:    summary.UpdatedCount,
			ErrorCount:      summary.ErrorCount,
			CurrentRelPath:  currentRelPath,
			CurrentRuleName: currentRuleName,
			RanAt:           summary.RanAt,
			Message:         message,
		})
	}
	if len(normalizedPaths) == 0 {
		emitProgress("completed", false, "", "", "No images to process")
		return summary, nil
	}

	a.autoRulesRunMu.Lock()
	defer a.autoRulesRunMu.Unlock()
	emitProgress("started", true, "", "", "Starting auto rules run")

	a.autoRulesMu.Lock()
	store, err := a.loadAutoRulesStoreUnlocked()
	a.autoRulesMu.Unlock()
	if err != nil {
		emitProgress("failed", false, "", "", err.Error())
		return summary, err
	}
	if !store.Enabled || len(store.Rules) == 0 {
		emitProgress("completed", false, "", "", "Auto rules are disabled or no rules are configured")
		return summary, nil
	}

	for i := range store.Rules {
		if !store.Rules[i].Enabled {
			continue
		}
		store.Rules[i].LastRunAt = summary.RanAt
		store.Rules[i].LastMatchCount = 0
		store.Rules[i].LastStatus = "success"
		store.Rules[i].LastError = ""
		store.Rules[i].UpdatedAt = time.Now().Format(time.RFC3339)
	}

	lastProgressEmitAt := time.Time{}
	emitRunningProgress := func(force bool, currentRelPath, currentRuleName string) {
		if !force && !lastProgressEmitAt.IsZero() && time.Since(lastProgressEmitAt) < 120*time.Millisecond {
			return
		}
		lastProgressEmitAt = time.Now()
		emitProgress("running", true, currentRelPath, currentRuleName, "")
	}

	for _, relPath := range normalizedPaths {
		currentRelPath := normalizeRelPath(relPath)
		currentRuleName := ""
		ctx, err := a.buildAutoRuleMatchContext(relPath)
		if err != nil {
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %v", relPath, err))
			summary.ProcessedCount++
			emitRunningProgress(summary.ProcessedCount >= summary.TotalCount, currentRelPath, currentRuleName)
			continue
		}

		imageUpdated := false
		currentPath := ctx.RelPath

		for i := range store.Rules {
			rule := &store.Rules[i]
			if !rule.Enabled {
				continue
			}
			if !matchesAutoRule(ctx, *rule) {
				continue
			}

			summary.MatchedCount++
			rule.LastMatchCount++
			currentRuleName = rule.Name

			nextPath, updated, stopAfterRule, actionErr := a.executeAutoRuleActions(currentPath, rule.Actions)
			if actionErr != nil {
				rule.LastStatus = "error"
				rule.LastError = actionErr.Error()
				summary.ErrorCount++
				summary.Errors = append(summary.Errors, fmt.Sprintf("%s / %s: %v", rule.Name, currentPath, actionErr))
				if stopAfterRule {
					break
				}
				continue
			}

			currentPath = nextPath
			imageUpdated = imageUpdated || updated

			if currentPath != ctx.RelPath {
				nextCtx, rebuildErr := a.buildAutoRuleMatchContext(currentPath)
				if rebuildErr != nil {
					rule.LastStatus = "error"
					rule.LastError = rebuildErr.Error()
					summary.ErrorCount++
					summary.Errors = append(summary.Errors, fmt.Sprintf("%s / %s: %v", rule.Name, currentPath, rebuildErr))
					break
				}
				ctx = nextCtx
			}

			if stopAfterRule {
				break
			}
		}

		if imageUpdated {
			summary.UpdatedCount++
		}
		summary.ProcessedCount++
		emitRunningProgress(summary.ProcessedCount >= summary.TotalCount, currentPath, currentRuleName)
	}

	a.autoRulesMu.Lock()
	saveErr := a.saveAutoRulesStoreUnlocked(store)
	a.autoRulesMu.Unlock()
	if saveErr != nil {
		emitProgress("failed", false, "", "", saveErr.Error())
		return summary, saveErr
	}

	if summary.UpdatedCount > 0 {
		a.scheduleImagesChangedEvent()
	}
	emitProgress("completed", false, "", "", "")

	return summary, nil
}

func (a *App) scheduleAutoRulesRun(paths []string) {
	normalizedPaths := uniqueNonEmptyStrings(paths)
	if len(normalizedPaths) == 0 {
		return
	}

	go func(items []string) {
		if _, err := a.runAutoRulesForPaths(items, "background"); err != nil {
			log.Printf("auto rules run failed: %v", err)
		}
	}(normalizedPaths)
}

func (a *App) RunAutoRulesNow() (AutoRulesRunSummary, error) {
	paths := make([]string, 0)
	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		paths = append(paths, relPath)
		return nil
	})
	if err != nil {
		return AutoRulesRunSummary{}, err
	}
	return a.runAutoRulesForPaths(paths, "manual")
}
