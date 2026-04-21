package backend

import "sync"

type Tag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

type ImageTagsMap map[string][]string
type ImageNotesMap map[string]string

type TrashMetadataMap map[string]TrashMetadata

type TrashMetadata struct {
	OriginalPath string `json:"originalPath"`
	DeletedAt    string `json:"deletedAt"`
}

type FavoriteGroup struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

type favoriteGroupsStore struct {
	Groups []FavoriteGroup `json:"groups"`
}

type CustomRoot struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	Order     int    `json:"order,omitempty"`
	Pinned    bool   `json:"pinned,omitempty"`
	Enabled   bool   `json:"enabled"`
	Locked    bool   `json:"locked,omitempty"`
	IsBuiltin bool   `json:"isBuiltin,omitempty"`
}

type customRootsStore struct {
	Version int          `json:"version"`
	Roots   []CustomRoot `json:"roots"`
}

type TrashItem struct {
	Filename     string `json:"filename"`
	OriginalPath string `json:"originalPath"`
	DeletedAt    string `json:"deletedAt"`
	Path         string `json:"path"`
}

var tagMutex sync.Mutex
