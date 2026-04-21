package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func defaultFavoriteGroup() FavoriteGroup {
	return FavoriteGroup{
		ID:    defaultFavoriteGroupID,
		Name:  defaultFavoriteGroupName,
		Paths: []string{},
	}
}

func normalizeFavoriteGroups(groups []FavoriteGroup) []FavoriteGroup {
	if len(groups) == 0 {
		return []FavoriteGroup{defaultFavoriteGroup()}
	}

	normalized := make([]FavoriteGroup, 0, len(groups)+1)
	seenIDs := make(map[string]struct{}, len(groups))
	hasDefault := false

	for _, group := range groups {
		id := strings.TrimSpace(group.ID)
		if id == "" {
			id = uuid.New().String()
		}
		if _, exists := seenIDs[id]; exists {
			id = uuid.New().String()
		}
		seenIDs[id] = struct{}{}

		name := repairLegacyMojibake(group.Name)
		if name == "" {
			if id == defaultFavoriteGroupID {
				name = defaultFavoriteGroupName
			} else {
				name = "Untitled Group"
			}
		}

		normalizedPaths := make([]string, 0, len(group.Paths))
		for _, pathValue := range group.Paths {
			if normalized := normalizeRelPath(repairLegacyMojibake(pathValue)); normalized != "" {
				normalizedPaths = append(normalizedPaths, normalized)
			}
		}
		paths := uniqueNonEmptyStrings(normalizedPaths)
		normalized = append(normalized, FavoriteGroup{
			ID:    id,
			Name:  name,
			Paths: paths,
		})
		if id == defaultFavoriteGroupID {
			hasDefault = true
		}
	}

	if !hasDefault {
		normalized = append([]FavoriteGroup{defaultFavoriteGroup()}, normalized...)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].ID == defaultFavoriteGroupID {
			return true
		}
		if normalized[j].ID == defaultFavoriteGroupID {
			return false
		}
		return normalized[i].Name < normalized[j].Name
	})

	return normalized
}

func collectFavoritePaths(groups []FavoriteGroup) []string {
	all := make([]string, 0)
	for _, group := range groups {
		all = append(all, group.Paths...)
	}
	return uniqueNonEmptyStrings(all)
}

func findFavoriteGroupIndex(groups []FavoriteGroup, id string) int {
	for i, group := range groups {
		if group.ID == id {
			return i
		}
	}
	return -1
}

func (a *App) loadFavoriteGroups() ([]FavoriteGroup, error) {
	data, err := os.ReadFile(a.favoritesFile())
	if err != nil {
		if os.IsNotExist(err) {
			return []FavoriteGroup{defaultFavoriteGroup()}, nil
		}
		return nil, err
	}

	var store favoriteGroupsStore
	if err := json.Unmarshal(data, &store); err == nil && len(store.Groups) > 0 {
		return normalizeFavoriteGroups(store.Groups), nil
	}

	var legacyPaths []string
	if err := json.Unmarshal(data, &legacyPaths); err == nil {
		if len(legacyPaths) > 0 {
			return []FavoriteGroup{
				{
					ID:    defaultFavoriteGroupID,
					Name:  defaultFavoriteGroupName,
					Paths: uniqueNonEmptyStrings(legacyPaths),
				},
			}, nil
		}
	}

	return []FavoriteGroup{defaultFavoriteGroup()}, nil
}

func (a *App) saveFavoriteGroups(groups []FavoriteGroup) error {
	store := favoriteGroupsStore{Groups: normalizeFavoriteGroups(groups)}
	data, _ := json.MarshalIndent(store, "", "  ")
	return os.WriteFile(a.favoritesFile(), data, 0644)
}

func (a *App) loadFavorites() ([]string, error) {
	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return nil, err
	}
	return collectFavoritePaths(groups), nil
}

func (a *App) saveFavorites(favs []string) error {
	return a.saveFavoriteGroups([]FavoriteGroup{
		{
			ID:    defaultFavoriteGroupID,
			Name:  defaultFavoriteGroupName,
			Paths: favs,
		},
	})
}

func (a *App) GetFavorites() ([]string, error) {
	return a.loadFavorites()
}

func (a *App) GetFavoriteGroups() ([]FavoriteGroup, error) {
	return a.loadFavoriteGroups()
}

func (a *App) CreateFavoriteGroup(name string) (FavoriteGroup, error) {
	groupName := strings.TrimSpace(name)
	if groupName == "" {
		return FavoriteGroup{}, fmt.Errorf("group name is required")
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return FavoriteGroup{}, err
	}
	for _, group := range groups {
		if strings.EqualFold(strings.TrimSpace(group.Name), groupName) {
			return FavoriteGroup{}, fmt.Errorf("group already exists")
		}
	}

	group := FavoriteGroup{
		ID:    uuid.New().String(),
		Name:  groupName,
		Paths: []string{},
	}
	groups = append(groups, group)
	if err := a.saveFavoriteGroups(groups); err != nil {
		return FavoriteGroup{}, err
	}
	return group, nil
}

func (a *App) UpdateFavoriteGroup(id, name string) error {
	groupName := strings.TrimSpace(name)
	if groupName == "" {
		return fmt.Errorf("group name is required")
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}
	index := findFavoriteGroupIndex(groups, id)
	if index < 0 {
		return fmt.Errorf("group not found")
	}
	groups[index].Name = groupName
	return a.saveFavoriteGroups(groups)
}

func (a *App) DeleteFavoriteGroup(id string) error {
	if id == defaultFavoriteGroupID {
		return fmt.Errorf("default group cannot be deleted")
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}

	filtered := make([]FavoriteGroup, 0, len(groups))
	found := false
	for _, group := range groups {
		if group.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, group)
	}
	if !found {
		return fmt.Errorf("group not found")
	}
	return a.saveFavoriteGroups(filtered)
}

func (a *App) SetImageFavoriteGroups(path string, groupIDs []string) error {
	normalizedPath := normalizeRelPath(path)
	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}

	validIDs := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		validIDs[group.ID] = struct{}{}
	}

	targetIDs := make(map[string]struct{})
	for _, groupID := range uniqueNonEmptyStrings(groupIDs) {
		if _, ok := validIDs[groupID]; ok {
			targetIDs[groupID] = struct{}{}
		}
	}

	for i := range groups {
		filtered := make([]string, 0, len(groups[i].Paths))
		for _, item := range groups[i].Paths {
			if item != normalizedPath {
				filtered = append(filtered, item)
			}
		}
		groups[i].Paths = filtered
		if _, ok := targetIDs[groups[i].ID]; ok {
			groups[i].Paths = append(groups[i].Paths, normalizedPath)
		}
		groups[i].Paths = uniqueNonEmptyStrings(groups[i].Paths)
	}

	return a.saveFavoriteGroups(groups)
}

func (a *App) AddImageToFavoriteGroup(path, groupID string) error {
	normalizedPath := normalizeRelPath(path)
	targetGroupID := strings.TrimSpace(groupID)
	if targetGroupID == "" {
		targetGroupID = defaultFavoriteGroupID
	}

	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}
	index := findFavoriteGroupIndex(groups, targetGroupID)
	if index < 0 {
		return fmt.Errorf("group not found")
	}
	if !contains(groups[index].Paths, normalizedPath) {
		groups[index].Paths = append(groups[index].Paths, normalizedPath)
	}
	return a.saveFavoriteGroups(groups)
}

func (a *App) RemoveImageFromFavoriteGroup(path, groupID string) error {
	normalizedPath := normalizeRelPath(path)
	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}
	index := findFavoriteGroupIndex(groups, groupID)
	if index < 0 {
		return fmt.Errorf("group not found")
	}
	filtered := make([]string, 0, len(groups[index].Paths))
	for _, item := range groups[index].Paths {
		if item != normalizedPath {
			filtered = append(filtered, item)
		}
	}
	groups[index].Paths = filtered
	return a.saveFavoriteGroups(groups)
}

func (a *App) AddFavorite(path string) error {
	return a.AddImageToFavoriteGroup(path, defaultFavoriteGroupID)
}

func (a *App) RemoveFavorite(path string) error {
	normalizedPath := normalizeRelPath(path)
	groups, err := a.loadFavoriteGroups()
	if err != nil {
		return err
	}
	changed := false
	for i := range groups {
		filtered := make([]string, 0, len(groups[i].Paths))
		for _, item := range groups[i].Paths {
			if item == normalizedPath {
				changed = true
				continue
			}
			filtered = append(filtered, item)
		}
		groups[i].Paths = filtered
	}
	if !changed {
		return nil
	}
	return a.saveFavoriteGroups(groups)
}

func (a *App) BatchFavorites(paths []string, action string) (int, error) {
	if action == "add" {
		groups, err := a.loadFavoriteGroups()
		if err != nil {
			return 0, err
		}
		index := findFavoriteGroupIndex(groups, defaultFavoriteGroupID)
		if index < 0 {
			groups = append([]FavoriteGroup{defaultFavoriteGroup()}, groups...)
			index = 0
		}
		for _, path := range paths {
			normalizedPath := normalizeRelPath(path)
			if !contains(groups[index].Paths, normalizedPath) {
				groups[index].Paths = append(groups[index].Paths, normalizedPath)
			}
		}
		return len(paths), a.saveFavoriteGroups(groups)
	}
	if action == "remove" {
		targets := make(map[string]struct{}, len(paths))
		for _, path := range paths {
			targets[normalizeRelPath(path)] = struct{}{}
		}
		groups, err := a.loadFavoriteGroups()
		if err != nil {
			return 0, err
		}
		for i := range groups {
			filtered := make([]string, 0, len(groups[i].Paths))
			for _, item := range groups[i].Paths {
				if _, ok := targets[item]; ok {
					continue
				}
				filtered = append(filtered, item)
			}
			groups[i].Paths = filtered
		}
		return len(paths), a.saveFavoriteGroups(groups)
	}
	return 0, fmt.Errorf("invalid action")
}
