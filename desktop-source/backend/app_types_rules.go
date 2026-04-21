package backend

type AutoRuleCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type AutoRuleAction struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type AutoRule struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Enabled        bool                `json:"enabled"`
	MatchMode      string              `json:"matchMode"`
	Conditions     []AutoRuleCondition `json:"conditions"`
	Actions        []AutoRuleAction    `json:"actions"`
	LastRunAt      string              `json:"lastRunAt,omitempty"`
	LastMatchCount int                 `json:"lastMatchCount,omitempty"`
	LastStatus     string              `json:"lastStatus,omitempty"`
	LastError      string              `json:"lastError,omitempty"`
	CreatedAt      string              `json:"createdAt,omitempty"`
	UpdatedAt      string              `json:"updatedAt,omitempty"`
}

type AutoRulesStore struct {
	Enabled bool       `json:"enabled"`
	Rules   []AutoRule `json:"rules"`
}

type AutoRulesRunSummary struct {
	TotalCount     int      `json:"totalCount"`
	ProcessedCount int      `json:"processedCount"`
	MatchedCount   int      `json:"matchedCount"`
	UpdatedCount   int      `json:"updatedCount"`
	ErrorCount     int      `json:"errorCount"`
	RanAt          string   `json:"ranAt"`
	Errors         []string `json:"errors,omitempty"`
}

type AutoRulesRunProgress struct {
	Source          string `json:"source"`
	Stage           string `json:"stage"`
	Running         bool   `json:"running"`
	TotalCount      int    `json:"totalCount"`
	ProcessedCount  int    `json:"processedCount"`
	MatchedCount    int    `json:"matchedCount"`
	UpdatedCount    int    `json:"updatedCount"`
	ErrorCount      int    `json:"errorCount"`
	CurrentRelPath  string `json:"currentRelPath,omitempty"`
	CurrentRuleName string `json:"currentRuleName,omitempty"`
	RanAt           string `json:"ranAt,omitempty"`
	Message         string `json:"message,omitempty"`
}
