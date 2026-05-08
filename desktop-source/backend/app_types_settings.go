package backend

type Settings struct {
	TrashRetentionDays             int              `json:"trashRetentionDays"`
	RootDir                        string           `json:"rootDir,omitempty"`
	OutputDir                      string           `json:"outputDir,omitempty"`
	OutputConfigured               bool             `json:"outputConfigured,omitempty"`
	PathVersion                    int              `json:"pathVersion,omitempty"`
	ShortcutSettings               ShortcutSettings `json:"shortcutSettings,omitempty"`
	UserProfile                    UserProfile      `json:"userProfile,omitempty"`
	UtilityMenu                    UtilityMenuState `json:"utilityMenu,omitempty"`
	GalleryPerformanceMode         string           `json:"galleryPerformanceMode,omitempty"`
	GalleryInitialBatchSize        int              `json:"galleryInitialBatchSize,omitempty"`
	GalleryPageSize                int              `json:"galleryPageSize,omitempty"`
	GalleryThumbPreferred          bool             `json:"galleryThumbPreferred,omitempty"`
	GalleryBackgroundVariantWarmup bool             `json:"galleryBackgroundVariantWarmup,omitempty"`
	GalleryMetadataLazy            bool             `json:"galleryMetadataLazy,omitempty"`
	AlwaysOnTop                    bool             `json:"alwaysOnTop,omitempty"`
}

type GalleryPerformanceSettings struct {
	Mode                    string `json:"mode"`
	InitialBatchSize        int    `json:"initialBatchSize"`
	PageSize                int    `json:"pageSize"`
	ThumbPreferred          bool   `json:"thumbPreferred"`
	BackgroundVariantWarmup bool   `json:"backgroundVariantWarmup"`
	MetadataLazy            bool   `json:"metadataLazy"`
}

type WindowBehaviorSettings struct {
	AlwaysOnTop bool `json:"alwaysOnTop"`
}

type UserProfile struct {
	DisplayName        string `json:"displayName,omitempty"`
	Headline           string `json:"headline,omitempty"`
	Bio                string `json:"bio,omitempty"`
	Location           string `json:"location,omitempty"`
	Website            string `json:"website,omitempty"`
	DailyGoal          int    `json:"dailyGoal,omitempty"`
	PreferredStartPage string `json:"preferredStartPage,omitempty"`
	ImagePath          string `json:"imagePath,omitempty"`
}

type UtilityMenuItem struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
	Order   int    `json:"order,omitempty"`
}

type UtilityMenuState struct {
	Items []UtilityMenuItem `json:"items,omitempty"`
}

type LauncherTool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Icon string `json:"icon"`
	Args string `json:"args"`
}

type DirectoryBinding struct {
	RootDir       string `json:"rootDir"`
	OutputDir     string `json:"outputDir"`
	OutputRelPath string `json:"outputRelPath"`
	Configured    bool   `json:"configured"`
}
