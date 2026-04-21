package backend

import (
	"os"
	"path/filepath"
	"strings"
)

func prefixLegacyRelPath(pathValue, prefix string) string {
	cleaned := normalizeRelPath(pathValue)
	prefix = normalizeRelPath(prefix)

	if cleaned == "" || prefix == "" {
		return cleaned
	}
	if cleaned == prefix || strings.HasPrefix(cleaned, prefix+"/") {
		return cleaned
	}
	return filepath.ToSlash(filepath.Join(prefix, cleaned))
}

func (a *App) normalizeManagedReferencePath(pathValue string) string {
	cleaned := normalizeRelPath(pathValue)
	if cleaned == "" || !a.hasDirectoryBinding() {
		return cleaned
	}

	if resolved, err := a.resolveRootPath(cleaned); err == nil {
		if _, statErr := os.Stat(resolved); statErr == nil {
			return cleaned
		}
	}

	outputPrefix := a.outputRelPath()
	prefixed := prefixLegacyRelPath(cleaned, outputPrefix)
	if prefixed == cleaned {
		return cleaned
	}

	if resolved, err := a.resolveRootPath(prefixed); err == nil {
		if _, statErr := os.Stat(resolved); statErr == nil {
			return prefixed
		}
	}

	return cleaned
}

func (a *App) migrateLegacyPathData(settings *Settings) error {
	outputPrefix := a.outputRelPath()

	favoriteGroups, favErr := a.loadFavoriteGroups()
	if favErr == nil && len(favoriteGroups) > 0 {
		changed := false
		for i := range favoriteGroups {
			for j, rel := range favoriteGroups[i].Paths {
				next := prefixLegacyRelPath(rel, outputPrefix)
				if next != rel {
					favoriteGroups[i].Paths[j] = next
					changed = true
				}
			}
			favoriteGroups[i].Paths = uniqueNonEmptyStrings(favoriteGroups[i].Paths)
		}
		if changed {
			if err := a.saveFavoriteGroups(favoriteGroups); err != nil {
				return err
			}
		}
	}

	imageTags, tagsErr := a.loadImageTags()
	if tagsErr == nil && len(imageTags) > 0 {
		migrated := make(ImageTagsMap, len(imageTags))
		changed := false
		for relPath, tagIDs := range imageTags {
			next := prefixLegacyRelPath(relPath, outputPrefix)
			if next != relPath {
				changed = true
			}
			migrated[next] = tagIDs
		}
		if changed {
			if err := a.saveImageTags(migrated); err != nil {
				return err
			}
		}
	}

	meta, metaErr := a.loadTrashMetadata()
	if metaErr == nil && len(meta) > 0 {
		changed := false
		for filename, item := range meta {
			next := prefixLegacyRelPath(item.OriginalPath, outputPrefix)
			if next != item.OriginalPath {
				item.OriginalPath = next
				meta[filename] = item
				changed = true
			}
		}
		if changed {
			if err := a.saveTrashMetadata(meta); err != nil {
				return err
			}
		}
	}

	customRoots, customErr := a.loadCustomRoots()
	if customErr == nil && len(customRoots) > 0 {
		changed := false
		for i, root := range customRoots {
			next := prefixLegacyRelPath(root.Path, outputPrefix)
			if next != root.Path {
				customRoots[i].Path = next
				changed = true
			}
		}
		if changed {
			if err := a.saveCustomRoots(customRoots); err != nil {
				return err
			}
		}
	}

	return nil
}
