package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func defaultDateArchiveCustomRoot() CustomRoot {
	return CustomRoot{
		ID:        builtinDateArchiveRootID,
		Name:      "日期归档目录",
		Path:      "日期归档",
		Icon:      "Calendar",
		Order:     0,
		Enabled:   true,
		Locked:    true,
		IsBuiltin: true,
	}
}

func normalizeCustomRootDisplayName(pathValue, displayName string) string {
	name := strings.TrimSpace(displayName)
	if name != "" && !strings.Contains(name, "\ufffd") && !strings.Contains(name, "?") {
		return name
	}

	parts := strings.Split(normalizeRelPath(pathValue), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func normalizeCustomRoots(roots []CustomRoot, preserveEnabled bool) []CustomRoot {
	builtin := defaultDateArchiveCustomRoot()
	builtinEnabled := builtin.Enabled
	builtinOrder := 1
	builtinPinned := builtin.Pinned
	allRoots := make([]CustomRoot, 0, len(roots)+1)
	seenPaths := map[string]bool{}

	for _, root := range roots {
		pathValue := normalizeRelPath(repairLegacyMojibake(root.Path))
		if pathValue == "" {
			continue
		}

		if root.ID == builtinDateArchiveRootID || pathValue == builtin.Path {
			if preserveEnabled {
				builtinEnabled = root.Enabled
			}
			if root.Order > 0 {
				builtinOrder = root.Order
			}
			builtinPinned = root.Pinned
			continue
		}

		if seenPaths[pathValue] {
			continue
		}
		seenPaths[pathValue] = true

		enabled := true
		if preserveEnabled {
			enabled = root.Enabled
		}

		allRoots = append(allRoots, CustomRoot{
			ID:        strings.TrimSpace(root.ID),
			Name:      normalizeCustomRootDisplayName(pathValue, repairLegacyMojibake(root.Name)),
			Path:      pathValue,
			Icon:      strings.TrimSpace(root.Icon),
			Order:     root.Order,
			Pinned:    root.Pinned,
			Enabled:   enabled,
			Locked:    false,
			IsBuiltin: false,
		})
	}

	builtin.Enabled = builtinEnabled
	builtin.Order = builtinOrder
	builtin.Pinned = builtinPinned
	allRoots = append(allRoots, builtin)

	sort.SliceStable(allRoots, func(i, j int) bool {
		if allRoots[i].Pinned != allRoots[j].Pinned {
			return allRoots[i].Pinned
		}
		leftOrder := allRoots[i].Order
		rightOrder := allRoots[j].Order
		if leftOrder <= 0 {
			leftOrder = 1000000 + i
		}
		if rightOrder <= 0 {
			rightOrder = 1000000 + j
		}
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		return allRoots[i].Name < allRoots[j].Name
	})

	for i := range allRoots {
		if strings.TrimSpace(allRoots[i].ID) == "" {
			allRoots[i].ID = uuid.New().String()
		}
		if strings.TrimSpace(allRoots[i].Icon) == "" {
			allRoots[i].Icon = "FolderSymlink"
		}
		allRoots[i].Order = i + 1
	}

	return allRoots
}

func (a *App) loadCustomRoots() ([]CustomRoot, error) {
	data, err := os.ReadFile(a.customRootsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return normalizeCustomRoots(nil, true), nil
		}
		return nil, err
	}
	data = stripUTF8BOM(data)

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return normalizeCustomRoots(nil, true), nil
	}

	var store customRootsStore
	if err := json.Unmarshal(data, &store); err == nil && (store.Version > 0 || store.Roots != nil) {
		normalized := normalizeCustomRoots(store.Roots, true)
		if store.Version != customRootsVersion || !reflect.DeepEqual(normalized, store.Roots) {
			_ = a.saveCustomRoots(normalized)
		}
		return normalized, nil
	}

	var roots []CustomRoot
	if err := json.Unmarshal(data, &roots); err != nil {
		normalized := normalizeCustomRoots(nil, true)
		_ = a.saveCustomRoots(normalized)
		return normalized, nil
	}
	normalized := normalizeCustomRoots(roots, false)
	if !reflect.DeepEqual(normalized, roots) {
		_ = a.saveCustomRoots(normalized)
	}
	return normalized, nil
}

func (a *App) saveCustomRoots(roots []CustomRoot) error {
	store := customRootsStore{
		Version: customRootsVersion,
		Roots:   normalizeCustomRoots(roots, true),
	}
	data, _ := json.MarshalIndent(store, "", "  ")
	return os.WriteFile(a.customRootsFile(), data, 0644)
}

func (a *App) GetCustomRoots() ([]CustomRoot, error) {
	return a.loadCustomRoots()
}

func (a *App) AddCustomRoot(name, relPath, icon string) (CustomRoot, error) {
	normalizedPath := normalizeRelPath(relPath)
	abs, err := a.resolveRootPath(normalizedPath)
	if err != nil {
		return CustomRoot{}, fmt.Errorf("自定义目录路径无效: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return CustomRoot{}, fmt.Errorf("目录不存在或不是文件夹: %s", relPath)
	}

	roots, _ := a.loadCustomRoots()
	for _, r := range roots {
		if normalizeRelPath(r.Path) == normalizedPath {
			return CustomRoot{}, fmt.Errorf("该目录已经被添加为自定义根目录")
		}
	}

	if strings.TrimSpace(name) == "" {
		name = normalizeCustomRootDisplayName(normalizedPath, "")
	}
	if strings.TrimSpace(icon) == "" {
		icon = "FolderSymlink"
	}

	newRoot := CustomRoot{
		ID:      uuid.New().String(),
		Name:    name,
		Path:    normalizedPath,
		Icon:    icon,
		Order:   len(roots),
		Enabled: true,
	}
	roots = append(roots, newRoot)
	if err := a.saveCustomRoots(roots); err != nil {
		return CustomRoot{}, err
	}
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return newRoot, nil
}

func (a *App) UpdateCustomRoot(id, name, icon string) error {
	roots, _ := a.loadCustomRoots()
	updated := false

	for i, root := range roots {
		if root.ID != id {
			continue
		}
		if root.Locked || root.IsBuiltin {
			return fmt.Errorf("内置根目录不允许修改")
		}

		displayName := strings.TrimSpace(name)
		if displayName == "" {
			displayName = normalizeCustomRootDisplayName(root.Path, "")
		}
		if strings.TrimSpace(icon) == "" {
			icon = "FolderSymlink"
		}

		roots[i].Name = displayName
		roots[i].Icon = icon
		updated = true
		break
	}

	if !updated {
		return fmt.Errorf("未找到要更新的自定义根目录")
	}
	if err := a.saveCustomRoots(roots); err != nil {
		return err
	}
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) DeleteCustomRoot(id string) error {
	roots, _ := a.loadCustomRoots()
	newRoots := make([]CustomRoot, 0, len(roots))
	deleted := false

	for _, root := range roots {
		if root.ID != id {
			newRoots = append(newRoots, root)
			continue
		}
		if root.Locked || root.IsBuiltin {
			return fmt.Errorf("内置根目录不允许删除")
		}
		deleted = true
	}

	if !deleted {
		return fmt.Errorf("未找到要删除的自定义根目录")
	}
	if err := a.saveCustomRoots(newRoots); err != nil {
		return err
	}
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) UpdateCustomRootEnabled(id string, enabled bool) error {
	roots, _ := a.loadCustomRoots()
	updated := false

	for i := range roots {
		if roots[i].ID != id {
			continue
		}
		roots[i].Enabled = enabled
		updated = true
		break
	}

	if !updated {
		return fmt.Errorf("未找到要更新的自定义根目录")
	}
	if err := a.saveCustomRoots(roots); err != nil {
		return err
	}
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) MoveCustomRoot(id, direction string) error {
	roots, _ := a.loadCustomRoots()
	index := -1
	for i := range roots {
		if roots[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("未找到要移动的自定义根目录")
	}

	target := index
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "up":
		target = index - 1
	case "down":
		target = index + 1
	default:
		return fmt.Errorf("unsupported move direction")
	}

	if target < 0 || target >= len(roots) {
		return nil
	}

	roots[index].Order, roots[target].Order = roots[target].Order, roots[index].Order
	if err := a.saveCustomRoots(roots); err != nil {
		return err
	}
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) PinCustomRoot(id string) error {
	roots, _ := a.loadCustomRoots()
	index := -1
	for i := range roots {
		if roots[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("未找到要置顶的自定义根目录")
	}
	if roots[index].Pinned {
		roots[index].Pinned = false
		roots[index].Order = len(roots) + 1
		if err := a.saveCustomRoots(roots); err != nil {
			return err
		}
		a.scheduleImagesChangedEvent()
		return nil
	}

	for i := range roots {
		if roots[i].ID != id {
			roots[i].Order++
		}
	}
	roots[index].Pinned = true
	roots[index].Order = 1
	if err := a.saveCustomRoots(roots); err != nil {
		return err
	}
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) GetRelativePath(absPath string) (string, error) {
	if !a.hasDirectoryBinding() {
		return "", fmt.Errorf("output directory is not configured")
	}
	abs, err := normalizeExistingPath(absPath)
	if err != nil {
		return "", err
	}
	if !isSubPath(a.rootDir, abs) {
		return "", fmt.Errorf("path is outside the root directory")
	}
	rel, err := filepath.Rel(a.rootDir, abs)
	if err != nil {
		return "", err
	}
	return normalizeRelPath(rel), nil
}

func (a *App) GetSubFolders(relPath string) ([]string, error) {
	var base string
	if relPath == "" {
		base = a.rootDir
	} else {
		resolved, err := a.resolveRootPath(relPath)
		if err != nil {
			return nil, fmt.Errorf("无法解析目录路径: %w", err)
		}
		base = resolved
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}

	var folders []string
	skipNames := map[string]bool{
		"node_modules": true,
		".git":         true,
		".trash":       true,
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if skipNames[name] || name == "desktop-source" {
			continue
		}
		var folderRel string
		if relPath == "" {
			folderRel = name
		} else {
			folderRel = filepath.ToSlash(filepath.Join(normalizeRelPath(relPath), name))
		}
		folders = append(folders, folderRel)
	}
	return folders, nil
}
