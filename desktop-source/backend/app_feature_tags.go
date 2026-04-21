package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

func (a *App) loadTags() ([]Tag, error) {
	var tags []Tag
	data, err := os.ReadFile(a.tagsFile())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		json.Unmarshal(data, &tags)
	}
	if tags == nil {
		tags = []Tag{}
	}
	return tags, nil
}

func (a *App) saveTags(tags []Tag) error {
	data, _ := json.MarshalIndent(tags, "", "  ")
	return os.WriteFile(a.tagsFile(), data, 0644)
}

func (a *App) GetTags() ([]Tag, error) {
	return a.loadTags()
}

func (a *App) CreateTag(name, color, category string) (Tag, error) {
	tagMutex.Lock()
	defer tagMutex.Unlock()

	tags, _ := a.loadTags()
	if category == "" {
		category = "default"
	}
	newTag := Tag{
		ID:       uuid.New().String(),
		Name:     name,
		Color:    color,
		Category: category,
	}
	tags = append(tags, newTag)
	err := a.saveTags(tags)
	return newTag, err
}

func (a *App) UpdateTag(id string, name, color, category *string) error {
	tagMutex.Lock()
	defer tagMutex.Unlock()

	tags, _ := a.loadTags()
	updated := false
	for i := range tags {
		if tags[i].ID == id {
			if name != nil {
				tags[i].Name = *name
			}
			if color != nil {
				tags[i].Color = *color
			}
			if category != nil {
				tags[i].Category = *category
			}
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("tag not found")
	}
	return a.saveTags(tags)
}

func (a *App) DeleteTag(id string) error {
	tagMutex.Lock()
	defer tagMutex.Unlock()

	tags, _ := a.loadTags()
	newTags := []Tag{}
	for _, tag := range tags {
		if tag.ID != id {
			newTags = append(newTags, tag)
		}
	}
	a.saveTags(newTags)

	imageTags, _ := a.loadImageTags()
	changed := false
	for relPath, tagIDs := range imageTags {
		newTagIDs := []string{}
		for _, tid := range tagIDs {
			if tid != id {
				newTagIDs = append(newTagIDs, tid)
			} else {
				changed = true
			}
		}
		imageTags[relPath] = newTagIDs
	}
	if changed {
		a.saveImageTags(imageTags)
	}

	return nil
}

func (a *App) BatchDeleteTags(tagIds []string) (int, error) {
	tagMutex.Lock()
	defer tagMutex.Unlock()

	tags, _ := a.loadTags()
	toDelete := make(map[string]bool)
	for _, id := range tagIds {
		toDelete[id] = true
	}

	newTags := []Tag{}
	for _, tag := range tags {
		if !toDelete[tag.ID] {
			newTags = append(newTags, tag)
		}
	}
	a.saveTags(newTags)

	imageTags, _ := a.loadImageTags()
	changed := false
	for relPath, tagIDs := range imageTags {
		newTagIDs := []string{}
		for _, id := range tagIDs {
			if !toDelete[id] {
				newTagIDs = append(newTagIDs, id)
			} else {
				changed = true
			}
		}
		imageTags[relPath] = newTagIDs
	}
	if changed {
		a.saveImageTags(imageTags)
	}

	return len(tagIds), nil
}

func (a *App) loadImageTags() (ImageTagsMap, error) {
	var imageTags ImageTagsMap
	data, err := os.ReadFile(a.imageTagsFile())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		json.Unmarshal(data, &imageTags)
	}
	if imageTags == nil {
		imageTags = make(ImageTagsMap)
	}
	normalized := make(ImageTagsMap, len(imageTags))
	changed := false
	for relPath, tagIDs := range imageTags {
		next := a.normalizeManagedReferencePath(relPath)
		if next == "" {
			continue
		}
		if next != normalizeRelPath(relPath) {
			changed = true
		}
		normalized[next] = uniqueNonEmptyStrings(append(normalized[next], tagIDs...))
	}
	if len(normalized) == 0 && len(imageTags) == 0 {
		return imageTags, nil
	}
	if changed {
		_ = a.saveImageTags(normalized)
	}
	return normalized, nil
}

func (a *App) saveImageTags(imageTags ImageTagsMap) error {
	data, _ := json.MarshalIndent(imageTags, "", "  ")
	return os.WriteFile(a.imageTagsFile(), data, 0644)
}

func (a *App) loadImageNotes() (ImageNotesMap, error) {
	var notes ImageNotesMap
	data, err := os.ReadFile(a.imageNotesFile())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		json.Unmarshal(data, &notes)
	}
	if notes == nil {
		notes = make(ImageNotesMap)
	}
	normalized := make(ImageNotesMap, len(notes))
	changed := false
	for relPath, note := range notes {
		next := a.normalizeManagedReferencePath(relPath)
		if next == "" {
			continue
		}
		if next != normalizeRelPath(relPath) {
			changed = true
		}
		if strings.TrimSpace(normalized[next]) == "" {
			normalized[next] = note
		}
	}
	if len(normalized) == 0 && len(notes) == 0 {
		return notes, nil
	}
	if changed {
		_ = a.saveImageNotes(normalized)
	}
	return normalized, nil
}

func (a *App) saveImageNotes(notes ImageNotesMap) error {
	data, _ := json.MarshalIndent(notes, "", "  ")
	return os.WriteFile(a.imageNotesFile(), data, 0644)
}

func (a *App) GetImageNotes() (ImageNotesMap, error) {
	return a.loadImageNotes()
}

func (a *App) SetImageNote(relPath, note string) error {
	relPath = normalizeRelPath(relPath)
	notes, err := a.loadImageNotes()
	if err != nil {
		return err
	}
	if strings.TrimSpace(note) == "" {
		delete(notes, relPath)
	} else {
		notes[relPath] = note
	}
	return a.saveImageNotes(notes)
}

func (a *App) DeleteImageNote(relPath string) error {
	relPath = normalizeRelPath(relPath)
	notes, err := a.loadImageNotes()
	if err != nil {
		return err
	}
	delete(notes, relPath)
	return a.saveImageNotes(notes)
}

func (a *App) GetImageTags() (ImageTagsMap, error) {
	return a.loadImageTags()
}

func (a *App) AddTagToImage(relPath, tagId string) ([]string, error) {
	relPath = strings.TrimPrefix(relPath, "/")
	imageTags, _ := a.loadImageTags()
	if imageTags[relPath] == nil {
		imageTags[relPath] = []string{}
	}
	exists := false
	for _, id := range imageTags[relPath] {
		if id == tagId {
			exists = true
			break
		}
	}
	if !exists {
		imageTags[relPath] = append(imageTags[relPath], tagId)
		err := a.saveImageTags(imageTags)
		return imageTags[relPath], err
	}
	return imageTags[relPath], nil
}

func (a *App) RemoveTagFromImage(relPath, tagId string) error {
	relPath = strings.TrimPrefix(relPath, "/")
	imageTags, _ := a.loadImageTags()
	if imageTags[relPath] != nil {
		newTagIDs := []string{}
		for _, id := range imageTags[relPath] {
			if id != tagId {
				newTagIDs = append(newTagIDs, id)
			}
		}
		imageTags[relPath] = newTagIDs
		return a.saveImageTags(imageTags)
	}
	return nil
}

func (a *App) BatchAddTag(paths []string, tagId string) (int, error) {
	imageTags, _ := a.loadImageTags()
	for _, relPath := range paths {
		if imageTags[relPath] == nil {
			imageTags[relPath] = []string{}
		}
		found := false
		for _, tid := range imageTags[relPath] {
			if tid == tagId {
				found = true
				break
			}
		}
		if !found {
			imageTags[relPath] = append(imageTags[relPath], tagId)
		}
	}
	err := a.saveImageTags(imageTags)
	return len(paths), err
}

func (a *App) BatchRemoveTag(paths []string, tagId string) (int, error) {
	imageTags, _ := a.loadImageTags()
	count := 0
	for _, relPath := range paths {
		if imageTags[relPath] != nil {
			newTagIDs := []string{}
			for _, tid := range imageTags[relPath] {
				if tid != tagId {
					newTagIDs = append(newTagIDs, tid)
				} else {
					count++
				}
			}
			imageTags[relPath] = newTagIDs
		}
	}
	err := a.saveImageTags(imageTags)
	return count, err
}
