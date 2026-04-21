package backend

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func isSubPath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	rel = filepath.Clean(rel)
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func normalizeDir(path string) (string, error) {
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return "", fmt.Errorf("path is empty")
	}
	abs, err := filepath.Abs(filepath.Clean(cleaned))
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory")
	}
	return abs, nil
}

func normalizeExistingPath(path string) (string, error) {
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return "", fmt.Errorf("path is empty")
	}
	abs, err := filepath.Abs(filepath.Clean(cleaned))
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
}

func normalizeRelPath(relPath string) string {
	cleaned := strings.TrimSpace(relPath)
	if cleaned == "" {
		return ""
	}
	cleaned = filepath.ToSlash(filepath.Clean(cleaned))
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.Trim(cleaned, "/")
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func isImageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	default:
		return false
	}
}

func (a *App) outputRelPath() string {
	if strings.TrimSpace(a.rootDir) == "" || strings.TrimSpace(a.imageDir) == "" {
		return ""
	}
	rel, err := filepath.Rel(a.rootDir, a.imageDir)
	if err != nil {
		return ""
	}
	return normalizeRelPath(rel)
}

func (a *App) hasDirectoryBinding() bool {
	return strings.TrimSpace(a.rootDir) != "" && strings.TrimSpace(a.imageDir) != ""
}

func (a *App) applyDirectoryBinding(rootDir, outputDir string) error {
	effectiveRoot := strings.TrimSpace(rootDir)
	if effectiveRoot == "" {
		effectiveRoot = a.imageDir
	}

	rootAbs, err := normalizeDir(effectiveRoot)
	if err != nil {
		return fmt.Errorf("failed to normalize root directory: %w", err)
	}

	effectiveOutput := strings.TrimSpace(outputDir)
	if effectiveOutput == "" {
		effectiveOutput = rootAbs
	}

	outputAbs, err := normalizeDir(effectiveOutput)
	if err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}

	if !isSubPath(rootAbs, outputAbs) {
		return fmt.Errorf("output directory must stay inside the selected root directory")
	}

	a.rootDir = rootAbs
	a.imageDir = outputAbs
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) validateDirectoryBinding(rootDir, outputDir string) (string, string, error) {
	effectiveRoot := strings.TrimSpace(rootDir)
	if effectiveRoot == "" {
		effectiveRoot = a.imageDir
	}

	rootAbs, err := normalizeDir(effectiveRoot)
	if err != nil {
		return "", "", fmt.Errorf("invalid root directory: %w", err)
	}

	effectiveOutput := strings.TrimSpace(outputDir)
	if effectiveOutput == "" {
		effectiveOutput = rootAbs
	}

	outputAbs, err := normalizeDir(effectiveOutput)
	if err != nil {
		return "", "", fmt.Errorf("invalid output directory: %w", err)
	}

	if !isSubPath(rootAbs, outputAbs) {
		return "", "", fmt.Errorf("output directory must stay inside the selected root directory")
	}

	return rootAbs, outputAbs, nil
}

func (a *App) restoreDirectoryBinding(rootDir, outputDir string) {
	a.rootDir = rootDir
	a.imageDir = outputDir
	a.restartImageWatcher()
}

func (a *App) resolveRootPath(relPath string) (string, error) {
	if !a.hasDirectoryBinding() {
		return "", fmt.Errorf("directory binding is not configured")
	}
	cleaned := normalizeRelPath(relPath)
	absPath := a.rootDir
	if cleaned != "" {
		absPath = filepath.Join(a.rootDir, filepath.FromSlash(cleaned))
	}

	absPath = filepath.Clean(absPath)
	if !isSubPath(a.rootDir, absPath) {
		return "", fmt.Errorf("path is outside the root directory")
	}
	return absPath, nil
}

func (a *App) resolveProfileAssetPath(relPath string) (string, error) {
	cleaned := normalizeRelPath(strings.TrimPrefix(relPath, profileAssetPrefix))
	if cleaned == "" {
		return "", fmt.Errorf("profile asset path is empty")
	}

	absPath := filepath.Clean(filepath.Join(a.profileImageDir(), filepath.FromSlash(cleaned)))
	if !isSubPath(a.profileImageDir(), absPath) {
		return "", fmt.Errorf("profile asset path is invalid")
	}
	return absPath, nil
}

func imageVariantFilename(kind, relPath string) string {
	sum := md5.Sum([]byte(kind + ":" + normalizeRelPath(relPath)))
	return hex.EncodeToString(sum[:]) + ".png"
}

func (a *App) imageVariantURL(kind, relPath string) string {
	cleaned := normalizeRelPath(relPath)
	if cleaned == "" {
		return ""
	}
	switch kind {
	case "thumb":
		return thumbVariantAssetPrefix + cleaned
	case "preview":
		return previewVariantAssetPrefix + cleaned
	default:
		return ""
	}
}

func (a *App) resolveVariantSource(relPath string) (kind string, sourceRelPath string, err error) {
	cleaned := normalizeRelPath(strings.TrimPrefix(relPath, variantAssetPrefix))
	parts := strings.SplitN(cleaned, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("variant path is invalid")
	}
	kind = strings.TrimSpace(parts[0])
	sourceRelPath = normalizeRelPath(parts[1])
	if sourceRelPath == "" {
		return "", "", fmt.Errorf("variant source path is empty")
	}
	switch kind {
	case "thumb", "preview":
		return kind, sourceRelPath, nil
	default:
		return "", "", fmt.Errorf("variant kind is invalid")
	}
}

func (a *App) variantSpec(kind string) (dir string, maxDimension int, err error) {
	switch kind {
	case "thumb":
		return a.thumbVariantsDir(), thumbVariantMaxDimension, nil
	case "preview":
		return a.previewVariantsDir(), previewVariantMaxDimension, nil
	default:
		return "", 0, fmt.Errorf("unsupported variant kind")
	}
}

func (a *App) favoritesFile() string       { return filepath.Join(a.dataDir, "favorites.json") }
func (a *App) tagsFile() string            { return filepath.Join(a.dataDir, "tags.json") }
func (a *App) imageTagsFile() string       { return filepath.Join(a.dataDir, "image-tags.json") }
func (a *App) trashMetadataFile() string   { return filepath.Join(a.dataDir, "trash-metadata.json") }
func (a *App) settingsFile() string        { return filepath.Join(a.dataDir, "settings.json") }
func (a *App) launcherToolsFile() string   { return filepath.Join(a.dataDir, "launcher-tools.json") }
func (a *App) promptToolLinksFile() string { return filepath.Join(a.dataDir, "prompt-tool-links.json") }
func (a *App) promptTemplatesFile() string { return filepath.Join(a.dataDir, "prompt-templates.json") }
func (a *App) promptLibraryDir() string    { return filepath.Join(a.dataDir, "prompt-library") }

func (a *App) promptLibraryFile() string {
	return filepath.Join(a.promptLibraryDir(), "all_prompts_merged.cleaned.json")
}

func (a *App) customPromptEntriesFile() string {
	return filepath.Join(a.dataDir, "custom-prompt-entries.json")
}

func (a *App) promptAssistantStateFile() string {
	return filepath.Join(a.dataDir, "prompt-assistant-state.json")
}

func (a *App) customRootsFile() string { return filepath.Join(a.dataDir, "custom-roots.json") }
func (a *App) imageNotesFile() string  { return filepath.Join(a.dataDir, "image-notes.json") }
func (a *App) autoRulesFile() string   { return filepath.Join(a.dataDir, "auto-rules.json") }

func (a *App) imageMetaCacheFile() string {
	return filepath.Join(a.dataDir, "image-meta-cache.json")
}

func (a *App) profileImageDir() string  { return filepath.Join(a.dataDir, "profile") }
func (a *App) imageVariantsDir() string { return filepath.Join(a.dataDir, "image-variants") }

func (a *App) previewVariantsDir() string {
	return filepath.Join(a.imageVariantsDir(), "preview")
}

func (a *App) thumbVariantsDir() string { return filepath.Join(a.imageVariantsDir(), "thumb") }
func (a *App) iconsDir() string         { return filepath.Join(a.dataDir, "icons") }
func (a *App) trashDir() string         { return filepath.Join(a.appDir, ".trash") }

func (a *App) legacyTrashDirs() []string {
	dirs := []string{}
	if strings.TrimSpace(a.imageDir) != "" {
		dirs = append(dirs, filepath.Join(a.imageDir, ".trash"))
	}
	dirs = append(dirs, filepath.Join(a.appDir, ".trash"))
	return dirs
}

func (a *App) trashRelPath(filename string) string {
	cleaned := normalizeRelPath(filename)
	if cleaned == "" {
		return ""
	}
	return trashAssetPrefix + cleaned
}

func (a *App) resolveTrashAssetPath(relPath string) (string, error) {
	cleaned := normalizeRelPath(strings.TrimPrefix(relPath, trashAssetPrefix))
	if cleaned == "" {
		return "", fmt.Errorf("trash asset path is empty")
	}

	absPath := filepath.Clean(filepath.Join(a.trashDir(), filepath.FromSlash(cleaned)))
	if !isSubPath(a.trashDir(), absPath) {
		return "", fmt.Errorf("trash asset path is invalid")
	}
	return absPath, nil
}
