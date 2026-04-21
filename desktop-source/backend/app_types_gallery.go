package backend

type ImageFile struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	ThumbPath   string   `json:"thumbPath,omitempty"`
	PreviewPath string   `json:"previewPath,omitempty"`
	RelPath     string   `json:"relPath"`
	ModTime     string   `json:"modTime"`
	Size        int64    `json:"size"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Prompt      string   `json:"prompt,omitempty"`
	Model       string   `json:"model,omitempty"`
	Loras       []string `json:"loras,omitempty"`
	SearchText  string   `json:"searchText,omitempty"`
}

type ImageMetaCacheEntry struct {
	Name            string   `json:"name"`
	RelPath         string   `json:"relPath"`
	ModTime         string   `json:"modTime"`
	Size            int64    `json:"size"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	MetadataScanned bool     `json:"metadataScanned,omitempty"`
	HasMetadata     bool     `json:"hasMetadata,omitempty"`
	HasWorkflow     bool     `json:"hasWorkflow,omitempty"`
	Positive        string   `json:"positive,omitempty"`
	Negative        string   `json:"negative,omitempty"`
	Model           string   `json:"model,omitempty"`
	Sampler         string   `json:"sampler,omitempty"`
	Loras           []string `json:"loras,omitempty"`
	SearchText      string   `json:"searchText,omitempty"`
}

type ImageMetaCache map[string]ImageMetaCacheEntry

type ImageGallerySummary struct {
	TotalImages       int    `json:"totalImages"`
	ManagedRootCount  int    `json:"managedRootCount"`
	ActiveMode        string `json:"activeMode"`
	ModeReason        string `json:"modeReason"`
	ThumbCacheCount   int    `json:"thumbCacheCount"`
	ThumbCacheBytes   int64  `json:"thumbCacheBytes"`
	PreviewCacheCount int    `json:"previewCacheCount"`
	PreviewCacheBytes int64  `json:"previewCacheBytes"`
}

type DirectoryHealthIssue struct {
	Key         string `json:"key"`
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Action      string `json:"action,omitempty"`
}

type DirectoryHealthSummary struct {
	TotalImages                   int                    `json:"totalImages"`
	EmptyFolderCount              int                    `json:"emptyFolderCount"`
	InvalidTagReferenceCount      int                    `json:"invalidTagReferenceCount"`
	InvalidFavoriteReferenceCount int                    `json:"invalidFavoriteReferenceCount"`
	ThumbCacheCount               int                    `json:"thumbCacheCount"`
	ThumbCacheBytes               int64                  `json:"thumbCacheBytes"`
	PreviewCacheCount             int                    `json:"previewCacheCount"`
	PreviewCacheBytes             int64                  `json:"previewCacheBytes"`
	Issues                        []DirectoryHealthIssue `json:"issues"`
}

type GetImagesPageQuery struct {
	SortBy            string `json:"sortBy"`
	SortOrder         string `json:"sortOrder"`
	Page              int    `json:"page"`
	PageSize          int    `json:"pageSize"`
	ScopeRelPath      string `json:"scopeRelPath,omitempty"`
	FavoritesOnly     bool   `json:"favoritesOnly,omitempty"`
	FavoriteGroupID   string `json:"favoriteGroupId,omitempty"`
	SearchQuery       string `json:"searchQuery,omitempty"`
	ActiveTagID       string `json:"activeTagId,omitempty"`
	ActiveModelFilter string `json:"activeModelFilter,omitempty"`
	ActiveLoraFilter  string `json:"activeLoraFilter,omitempty"`
	ActiveDatePreset  string `json:"activeDatePreset,omitempty"`
	ActiveDateStart   string `json:"activeDateStart,omitempty"`
	ActiveDateEnd     string `json:"activeDateEnd,omitempty"`
}

type GetImagesPageResult struct {
	Items      []ImageFile `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
	HasMore    bool        `json:"hasMore"`
	Mode       string      `json:"mode"`
	ModeReason string      `json:"modeReason,omitempty"`
}

type WorkbenchSummaryQuery struct {
	ActiveDatePreset  string `json:"activeDatePreset,omitempty"`
	ActiveDateStart   string `json:"activeDateStart,omitempty"`
	ActiveDateEnd     string `json:"activeDateEnd,omitempty"`
	ActiveModelFilter string `json:"activeModelFilter,omitempty"`
	ActiveLoraFilter  string `json:"activeLoraFilter,omitempty"`
}

type WorkbenchFilterOption struct {
	Value   string   `json:"value"`
	Label   string   `json:"label"`
	Count   int      `json:"count"`
	Aliases []string `json:"aliases,omitempty"`
}

type WorkbenchRecentDate struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type WorkbenchSummary struct {
	Total       int                   `json:"total"`
	DatedTotal  int                   `json:"datedTotal"`
	Today       int                   `json:"today"`
	Yesterday   int                   `json:"yesterday"`
	Last7       int                   `json:"last7"`
	Month       int                   `json:"month"`
	RecentDates []WorkbenchRecentDate `json:"recentDates"`
}

type WorkbenchAggregateResult struct {
	AvailableModels []WorkbenchFilterOption `json:"availableModels"`
	AvailableLoras  []WorkbenchFilterOption `json:"availableLoras"`
	Summary         WorkbenchSummary        `json:"summary"`
	FilteredCount   int                     `json:"filteredCount"`
}

type CacheClearResult struct {
	DeletedFiles int   `json:"deletedFiles"`
	DeletedDirs  int   `json:"deletedDirs"`
	BytesFreed   int64 `json:"bytesFreed"`
}

type Stats struct {
	TotalCount int            `json:"totalCount"`
	TodayCount int            `json:"todayCount"`
	TotalSize  int64          `json:"totalSize"`
	ByDate     map[string]int `json:"byDate"`
	ByTag      map[string]int `json:"byTag"`
}
