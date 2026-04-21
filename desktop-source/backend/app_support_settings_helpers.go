package backend

import (
	"sort"
	"strings"
)

func defaultSettings() Settings {
	perf := defaultGalleryPerformanceSettings()
	return Settings{
		TrashRetentionDays:             30,
		ShortcutSettings:               defaultShortcutSettings(),
		UserProfile:                    defaultUserProfile(),
		UtilityMenu:                    defaultUtilityMenuState(),
		GalleryPerformanceMode:         perf.Mode,
		GalleryInitialBatchSize:        perf.InitialBatchSize,
		GalleryPageSize:                perf.PageSize,
		GalleryThumbPreferred:          perf.ThumbPreferred,
		GalleryBackgroundVariantWarmup: perf.BackgroundVariantWarmup,
		GalleryMetadataLazy:            perf.MetadataLazy,
	}
}

func defaultGalleryPerformanceSettings() GalleryPerformanceSettings {
	return GalleryPerformanceSettings{
		Mode:                    "auto",
		InitialBatchSize:        60,
		PageSize:                50,
		ThumbPreferred:          true,
		BackgroundVariantWarmup: true,
		MetadataLazy:            true,
	}
}

func normalizeGalleryPerformanceSettings(settings GalleryPerformanceSettings) GalleryPerformanceSettings {
	defaults := defaultGalleryPerformanceSettings()

	switch strings.TrimSpace(strings.ToLower(settings.Mode)) {
	case "auto", "standard", "performance":
		settings.Mode = strings.TrimSpace(strings.ToLower(settings.Mode))
	default:
		settings.Mode = defaults.Mode
	}

	if settings.InitialBatchSize <= 0 {
		settings.InitialBatchSize = defaults.InitialBatchSize
	}
	if settings.InitialBatchSize > 500 {
		settings.InitialBatchSize = 500
	}

	if settings.PageSize <= 0 {
		settings.PageSize = defaults.PageSize
	}
	if settings.PageSize > 500 {
		settings.PageSize = 500
	}

	if !settings.ThumbPreferred && !settings.BackgroundVariantWarmup && !settings.MetadataLazy &&
		settings.Mode == "" && settings.InitialBatchSize == 0 && settings.PageSize == 0 {
		return defaults
	}

	return settings
}

func settingsToGalleryPerformanceSettings(settings Settings) GalleryPerformanceSettings {
	return normalizeGalleryPerformanceSettings(GalleryPerformanceSettings{
		Mode:                    settings.GalleryPerformanceMode,
		InitialBatchSize:        settings.GalleryInitialBatchSize,
		PageSize:                settings.GalleryPageSize,
		ThumbPreferred:          settings.GalleryThumbPreferred,
		BackgroundVariantWarmup: settings.GalleryBackgroundVariantWarmup,
		MetadataLazy:            settings.GalleryMetadataLazy,
	})
}

func applyGalleryPerformanceSettings(settings *Settings, performance GalleryPerformanceSettings) {
	performance = normalizeGalleryPerformanceSettings(performance)
	settings.GalleryPerformanceMode = performance.Mode
	settings.GalleryInitialBatchSize = performance.InitialBatchSize
	settings.GalleryPageSize = performance.PageSize
	settings.GalleryThumbPreferred = performance.ThumbPreferred
	settings.GalleryBackgroundVariantWarmup = performance.BackgroundVariantWarmup
	settings.GalleryMetadataLazy = performance.MetadataLazy
}

func defaultUserProfile() UserProfile {
	return UserProfile{
		DisplayName:        "\u7075\u52a8\u56fe\u5e93\u7528\u6237",
		Headline:           "\u628a\u4f5c\u54c1\u6574\u7406\u6210\u7a33\u5b9a\u3001\u6e05\u723d\u7684\u56fe\u5e93",
		Bio:                "\u8fd9\u91cc\u4fdd\u5b58\u4f60\u7684\u51fa\u56fe\u8282\u594f\u3001\u504f\u597d\u8bbe\u7f6e\u548c\u5e38\u7528\u5165\u53e3\uff0c\u8ba9\u5de5\u4f5c\u6d41\u4fdd\u6301\u987a\u624b\u3002",
		Location:           "",
		Website:            "",
		DailyGoal:          12,
		PreferredStartPage: "dashboard",
		ImagePath:          "",
	}
}

func normalizeUserProfile(profile UserProfile) UserProfile {
	defaults := defaultUserProfile()

	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.DisplayName = repairLegacyMojibake(profile.DisplayName)
	if profile.DisplayName == "" {
		profile.DisplayName = defaults.DisplayName
	}

	profile.Headline = strings.TrimSpace(profile.Headline)
	profile.Headline = repairLegacyMojibake(profile.Headline)
	if profile.Headline == "" {
		profile.Headline = defaults.Headline
	}

	profile.Bio = strings.TrimSpace(profile.Bio)
	profile.Bio = repairLegacyMojibake(profile.Bio)
	if profile.Bio == "" {
		profile.Bio = defaults.Bio
	}

	profile.Location = strings.TrimSpace(profile.Location)
	profile.Website = strings.TrimSpace(profile.Website)
	profile.ImagePath = normalizeRelPath(strings.TrimSpace(profile.ImagePath))

	if profile.DailyGoal <= 0 {
		profile.DailyGoal = defaults.DailyGoal
	}
	if profile.DailyGoal > 999 {
		profile.DailyGoal = 999
	}

	switch profile.PreferredStartPage {
	case "dashboard", "statistics", "profile", "favorites", "documentation", "output":
	default:
		profile.PreferredStartPage = defaults.PreferredStartPage
	}

	return profile
}

var utilityMenuCatalog = []string{
	"settings",
	"trash",
	"documentation",
	"statistics",
	"launcher",
	"prompt-assistant",
	"prompt-templates",
	"auto-rules",
	"open-output",
	"switch-output",
	"custom-roots",
}

func defaultUtilityMenuState() UtilityMenuState {
	items := make([]UtilityMenuItem, 0, len(utilityMenuCatalog))
	for index, id := range utilityMenuCatalog {
		items = append(items, UtilityMenuItem{
			ID:      id,
			Visible: true,
			Order:   index + 1,
		})
	}
	return UtilityMenuState{Items: items}
}

func normalizeUtilityMenuState(state UtilityMenuState) UtilityMenuState {
	defaults := defaultUtilityMenuState()
	if len(state.Items) == 0 {
		return defaults
	}

	known := make(map[string]UtilityMenuItem, len(state.Items))
	for _, item := range state.Items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		known[id] = UtilityMenuItem{
			ID:      id,
			Visible: item.Visible,
			Order:   item.Order,
		}
	}

	items := make([]UtilityMenuItem, 0, len(defaults.Items))
	for index, fallback := range defaults.Items {
		item, exists := known[fallback.ID]
		if !exists {
			item = fallback
		}
		if item.Order <= 0 {
			item.Order = index + 1
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Order != items[j].Order {
			return items[i].Order < items[j].Order
		}
		return items[i].ID < items[j].ID
	})

	for index := range items {
		items[index].Order = index + 1
	}

	return UtilityMenuState{Items: items}
}
