package backend

type ImageMetadata struct {
	RelPath     string            `json:"relPath"`
	Format      string            `json:"format"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	HasMetadata bool              `json:"hasMetadata"`
	Prompt      string            `json:"prompt"`
	Workflow    string            `json:"workflow"`
	Positive    string            `json:"positive"`
	Negative    string            `json:"negative"`
	Model       string            `json:"model"`
	Sampler     string            `json:"sampler"`
	Scheduler   string            `json:"scheduler"`
	Seed        string            `json:"seed"`
	Steps       string            `json:"steps"`
	CFG         string            `json:"cfg"`
	Loras       []string          `json:"loras"`
	NodeCount   int               `json:"nodeCount"`
	ExtraFields map[string]string `json:"extraFields"`
	PromptDebug *PromptDebugInfo  `json:"promptDebug,omitempty"`
}

type PromptCandidateDebug struct {
	Text         string `json:"text"`
	Score        int    `json:"score"`
	SourceNodeID string `json:"sourceNodeId,omitempty"`
	SourceClass  string `json:"sourceClass,omitempty"`
	SourceTitle  string `json:"sourceTitle,omitempty"`
	SourceKey    string `json:"sourceKey,omitempty"`
	Strategy     string `json:"strategy,omitempty"`
	Depth        int    `json:"depth,omitempty"`
}

type PromptSelectionDebug struct {
	SelectedText string                 `json:"selectedText,omitempty"`
	Strategy     string                 `json:"strategy,omitempty"`
	SourceNodeID string                 `json:"sourceNodeId,omitempty"`
	SourceClass  string                 `json:"sourceClass,omitempty"`
	SourceTitle  string                 `json:"sourceTitle,omitempty"`
	SourceKey    string                 `json:"sourceKey,omitempty"`
	Candidates   []PromptCandidateDebug `json:"candidates,omitempty"`
}

type PromptDebugInfo struct {
	Positive PromptSelectionDebug `json:"positive"`
	Negative PromptSelectionDebug `json:"negative"`
}

type comfyPromptNode struct {
	ClassType string         `json:"class_type"`
	Inputs    map[string]any `json:"inputs"`
	Meta      struct {
		Title string `json:"title"`
	} `json:"_meta"`
}

type PromptToolLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

type PromptTemplate struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Category   string `json:"category"`
	SourcePath string `json:"sourcePath"`
	CreatedAt  string `json:"createdAt"`
}

type PromptLibraryEntry struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Scope       string `json:"scope"`
	TextEN      string `json:"text_en"`
	TextZH      string `json:"text_zh"`
	Preview     string `json:"preview"`
	ExtraID     string `json:"extra_id"`
	SearchText  string `json:"search_text"`
}

type PromptAssistantState struct {
	FavoriteIDs       []string `json:"favoriteIds"`
	RecentIDs         []string `json:"recentIds"`
	ActiveSource      string   `json:"activeSource,omitempty"`
	ActiveCategory    string   `json:"activeCategory,omitempty"`
	ActiveSubcategory string   `json:"activeSubcategory,omitempty"`
	ActiveScope       string   `json:"activeScope,omitempty"`
	ViewMode          string   `json:"viewMode,omitempty"`
	ActiveEditor      string   `json:"activeEditor,omitempty"`
	ItemsPerPage      int      `json:"itemsPerPage,omitempty"`
	CurrentPage       int      `json:"currentPage,omitempty"`
}
