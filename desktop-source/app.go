package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "golang.org/x/image/webp"
)

type ImageFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	RelPath string `json:"relPath"`
	ModTime string `json:"modTime"`
	Size    int64  `json:"size"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Prompt  string `json:"prompt,omitempty"`
	Model   string `json:"model,omitempty"`
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
}

type comfyPromptNode struct {
	ClassType string         `json:"class_type"`
	Inputs    map[string]any `json:"inputs"`
}

type Tag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

type ImageTagsMap map[string][]string // relPath -> tag IDs
type TrashMetadataMap map[string]TrashMetadata

type TrashMetadata struct {
	OriginalPath string `json:"originalPath"`
	DeletedAt    string `json:"deletedAt"`
}

type Settings struct {
	TrashRetentionDays int    `json:"trashRetentionDays"`
	RootDir            string `json:"rootDir,omitempty"`
	OutputDir          string `json:"outputDir,omitempty"`
	PathVersion        int    `json:"pathVersion,omitempty"`
}

type LauncherTool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Icon string `json:"icon"`
	Args string `json:"args"`
}

type FavoriteGroup struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

type favoriteGroupsStore struct {
	Groups []FavoriteGroup `json:"groups"`
}

type CacheClearResult struct {
	DeletedFiles int   `json:"deletedFiles"`
	DeletedDirs  int   `json:"deletedDirs"`
	BytesFreed   int64 `json:"bytesFreed"`
}

const defaultFavoriteGroupID = "default"
const defaultFavoriteGroupName = "默认收藏"

type CustomRoot struct {
	ID   string `json:"id"`
	Name string `json:"name"` // 侧边栏显示名称
	Path string `json:"path"` // 相对于 imageDir 的路径
	Icon string `json:"icon"` // Lucide 图标名称
}

type TrashItem struct {
	Filename     string `json:"filename"`
	OriginalPath string `json:"originalPath"`
	DeletedAt    string `json:"deletedAt"`
	Path         string `json:"path"`
}

type Stats struct {
	TotalCount int            `json:"totalCount"`
	TodayCount int            `json:"todayCount"`
	TotalSize  int64          `json:"totalSize"`
	ByDate     map[string]int `json:"byDate"`
	ByTag      map[string]int `json:"byTag"`
}

type imageMetaWarmupTask struct {
	Path  string
	Entry ImageMetaCacheEntry
}

type DirectoryBinding struct {
	RootDir       string `json:"rootDir"`
	OutputDir     string `json:"outputDir"`
	OutputRelPath string `json:"outputRelPath"`
}

var tagMutex sync.Mutex

const pathVersionRootRelative = 2

// App struct
type App struct {
	ctx                    context.Context
	rootDir                string
	imageDir               string
	dataDir                string
	appDir                 string
	imageMetaMu            sync.RWMutex
	imageMetaCache         ImageMetaCache
	imageMetaLoaded        bool
	imageMetaWarmupRunning bool
	watchMu                sync.Mutex
	imageWatcher           *fsnotify.Watcher
	imageWatchStop         chan struct{}
	imageWatchDebounce     *time.Timer
}

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

// NewApp creates a new App application struct
func NewApp() *App {
	// Use os.Executable to get the real path of this exe regardless of where it's launched from.
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exePath, err = filepath.EvalSymlinks(exePath) // resolve any symlinks
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(exePath)

	// Determine the "unified root" (comfy-manager/ folder):
	//
	// Production layouts supported:
	//   A) New layout:  comfy-manager/desktop-app.exe
	//      → exeDir = comfy-manager/  → unifiedRoot = exeDir
	//   B) Old layout:  comfy-manager/desktop-source/build/bin/app.exe
	//      → exeDir = .../build/bin   → unifiedRoot = up 3 levels
	//
	// Dev (wails dev): exe is in system temp; fall back to Getwd()
	//   → Getwd() = comfy-manager/desktop-source/
	//   → unifiedRoot = parent of Getwd() = comfy-manager/

	unifiedRoot := exeDir // default: assume exe is at unified root (layout A)

	exeBase := filepath.Base(exeDir)
	exeParent := filepath.Dir(exeDir)
	if exeBase == "bin" && filepath.Base(exeParent) == "build" {
		// Layout B: build/bin/app.exe inside desktop-source
		unifiedRoot = filepath.Dir(filepath.Dir(exeParent)) // up 3
	} else if strings.Contains(exePath, os.TempDir()) || strings.EqualFold(exeBase, "tmp") {
		// Dev mode — fall back to working directory
		if wd, wdErr := os.Getwd(); wdErr == nil {
			// Getwd() = .../desktop-source/ → parent = comfy-manager/
			unifiedRoot = filepath.Dir(wd)
		}
	}

	// Default layout:
	// appDir    = comfy-manager/
	// outputDir = parent of appDir
	// rootDir   = parent of outputDir
	defaultOutputDir := filepath.Dir(unifiedRoot)
	defaultRootDir := filepath.Dir(defaultOutputDir)
	dataDir := filepath.Join(unifiedRoot, "data")

	app := &App{
		rootDir:  defaultRootDir,
		imageDir: defaultOutputDir,
		dataDir:  dataDir,
		appDir:   unifiedRoot,
	}

	// Ensure data directory exists
	if _, err := os.Stat(app.dataDir); os.IsNotExist(err) {
		os.MkdirAll(app.dataDir, 0755)
	}

	settings, _ := app.loadSettings()
	if err := app.applyDirectoryBinding(settings.RootDir, settings.OutputDir); err != nil {
		log.Printf("failed to apply saved directory binding, using detected defaults: %v", err)
		app.rootDir = defaultOutputDir
		app.imageDir = defaultOutputDir
	}

	if (strings.TrimSpace(settings.RootDir) != "" || strings.TrimSpace(settings.OutputDir) != "") && settings.PathVersion < pathVersionRootRelative {
		if err := app.migrateLegacyPathData(&settings); err != nil {
			log.Printf("failed to migrate legacy paths: %v", err)
		} else {
			if strings.TrimSpace(settings.RootDir) == "" {
				settings.RootDir = app.rootDir
			}
			if strings.TrimSpace(settings.OutputDir) == "" {
				settings.OutputDir = app.imageDir
			}
			settings.PathVersion = pathVersionRootRelative
			_ = app.saveSettings(settings)
		}
	}

	if err := app.migrateLegacyTrash(); err != nil {
		log.Printf("failed to migrate legacy trash: %v", err)
	}

	if err := os.MkdirAll(app.trashDir(), 0755); err != nil {
		log.Printf("failed to ensure trash directory: %v", err)
	}

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		_ = a.cleanExpiredTrash()
		_, _ = a.cleanupTagsSilent()
	}()
	a.restartImageWatcher()
}

func (a *App) shutdown(ctx context.Context) {
	a.stopImageWatcher()
}

func (a *App) outputRelPath() string {
	rel, err := filepath.Rel(a.rootDir, a.imageDir)
	if err != nil {
		return ""
	}
	rel = normalizeRelPath(rel)
	return rel
}

func (a *App) applyDirectoryBinding(rootDir, outputDir string) error {
	effectiveRoot := strings.TrimSpace(rootDir)
	if effectiveRoot == "" {
		effectiveRoot = a.imageDir
	}

	rootAbs, err := normalizeDir(effectiveRoot)
	if err != nil {
		return fmt.Errorf("根目录无效: %w", err)
	}

	effectiveOutput := strings.TrimSpace(outputDir)
	if effectiveOutput == "" {
		effectiveOutput = rootAbs
	}

	outputAbs, err := normalizeDir(effectiveOutput)
	if err != nil {
		return fmt.Errorf("output 目录无效: %w", err)
	}

	if !isSubPath(rootAbs, outputAbs) {
		return fmt.Errorf("output 目录必须位于根目录内")
	}

	a.rootDir = rootAbs
	a.imageDir = outputAbs
	a.restartImageWatcher()
	a.scheduleImagesChangedEvent()
	return nil
}

func (a *App) resolveRootPath(relPath string) (string, error) {
	cleaned := normalizeRelPath(relPath)
	absPath := a.rootDir
	if cleaned != "" {
		absPath = filepath.Join(a.rootDir, filepath.FromSlash(cleaned))
	}

	absPath = filepath.Clean(absPath)
	if !isSubPath(a.rootDir, absPath) {
		return "", fmt.Errorf("路径不在根目录内")
	}
	return absPath, nil
}

func (a *App) shouldSkipDir(path, name string) bool {
	switch name {
	case "node_modules", ".git", ".trash":
		return true
	}

	lowerName := strings.ToLower(name)
	if strings.HasPrefix(lowerName, "comfy-manager") {
		return true
	}

	return a.appDir != "" && (samePath(path, a.appDir) || isSubPath(a.appDir, path))
}

func (a *App) managedImageRoots() []string {
	candidates := []string{a.imageDir}

	customRoots, err := a.loadCustomRoots()
	if err == nil {
		for _, root := range customRoots {
			absPath, resolveErr := a.resolveRootPath(root.Path)
			if resolveErr != nil {
				continue
			}
			info, statErr := os.Stat(absPath)
			if statErr != nil || !info.IsDir() {
				continue
			}
			candidates = append(candidates, absPath)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return len(filepath.Clean(candidates[i])) < len(filepath.Clean(candidates[j]))
	})

	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		normalized := filepath.Clean(candidate)
		covered := false
		for _, existing := range roots {
			if isSubPath(existing, normalized) {
				covered = true
				break
			}
		}
		if !covered {
			roots = append(roots, normalized)
		}
	}

	return roots
}

func (a *App) stopImageWatcher() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.imageWatchDebounce != nil {
		a.imageWatchDebounce.Stop()
		a.imageWatchDebounce = nil
	}
	if a.imageWatchStop != nil {
		close(a.imageWatchStop)
		a.imageWatchStop = nil
	}
	if a.imageWatcher != nil {
		_ = a.imageWatcher.Close()
		a.imageWatcher = nil
	}
}

func (a *App) scheduleImagesChangedEvent() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()

	if a.ctx == nil {
		return
	}

	if a.imageWatchDebounce != nil {
		a.imageWatchDebounce.Stop()
	}
	a.imageWatchDebounce = time.AfterFunc(350*time.Millisecond, func() {
		runtime.EventsEmit(a.ctx, "images:changed")
	})
}

func (a *App) addWatchTree(watcher *fsnotify.Watcher, root string, seen map[string]struct{}) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if a.shouldSkipDir(path, d.Name()) {
			if samePath(path, root) {
				return nil
			}
			return fs.SkipDir
		}

		cleaned := filepath.Clean(path)
		if _, ok := seen[cleaned]; ok {
			return nil
		}
		if err := watcher.Add(cleaned); err != nil {
			return nil
		}
		seen[cleaned] = struct{}{}
		return nil
	})
}

func shouldReactToWatchEvent(event fsnotify.Event) bool {
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0
}

func (a *App) restartImageWatcher() {
	a.stopImageWatcher()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("failed to create image watcher: %v", err)
		return
	}

	seen := make(map[string]struct{})
	for _, root := range a.managedImageRoots() {
		if err := a.addWatchTree(watcher, root, seen); err != nil {
			log.Printf("failed to watch root %s: %v", root, err)
		}
	}

	stop := make(chan struct{})

	a.watchMu.Lock()
	a.imageWatcher = watcher
	a.imageWatchStop = stop
	a.watchMu.Unlock()

	go func(localWatcher *fsnotify.Watcher, stopCh chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case event, ok := <-localWatcher.Events:
				if !ok {
					return
				}
				if !shouldReactToWatchEvent(event) {
					continue
				}

				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						a.restartImageWatcher()
						a.scheduleImagesChangedEvent()
						continue
					}
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					a.restartImageWatcher()
				}

				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					a.scheduleImagesChangedEvent()
					continue
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 || isImageExt(filepath.Ext(event.Name)) {
					a.scheduleImagesChangedEvent()
				}
			case err, ok := <-localWatcher.Errors:
				if !ok {
					return
				}
				log.Printf("image watcher error: %v", err)
			}
		}
	}(watcher, stop)
}

func (a *App) walkManagedImages(visitor func(absPath, relPath string, info fs.FileInfo) error) error {
	seen := make(map[string]bool)

	for _, root := range a.managedImageRoots() {
		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				if a.shouldSkipDir(path, d.Name()) {
					return fs.SkipDir
				}
				return nil
			}

			if !isImageExt(filepath.Ext(path)) {
				return nil
			}

			relPath, relErr := filepath.Rel(a.rootDir, path)
			if relErr != nil {
				return nil
			}
			relPath = normalizeRelPath(relPath)

			if seen[relPath] {
				return nil
			}
			seen[relPath] = true

			info, infoErr := d.Info()
			if infoErr != nil {
				return nil
			}

			return visitor(path, relPath, info)
		})
		if walkErr != nil {
			return walkErr
		}
	}

	return nil
}

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

// Custom Asset Handler for serving images from the filesystem
func (a *App) serveImage(w http.ResponseWriter, r *http.Request) {
	path := normalizeRelPath(strings.TrimPrefix(r.URL.Path, "/"))

	absPath, err := a.resolveRootPath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Path"))
		return
	}

	// Open file
	http.ServeFile(w, r, absPath)
}

// --- Internal Data Helpers ---
func (a *App) favoritesFile() string     { return filepath.Join(a.dataDir, "favorites.json") }
func (a *App) tagsFile() string          { return filepath.Join(a.dataDir, "tags.json") }
func (a *App) imageTagsFile() string     { return filepath.Join(a.dataDir, "image-tags.json") }
func (a *App) trashMetadataFile() string { return filepath.Join(a.dataDir, "trash-metadata.json") }
func (a *App) settingsFile() string      { return filepath.Join(a.dataDir, "settings.json") }
func (a *App) launcherToolsFile() string { return filepath.Join(a.dataDir, "launcher-tools.json") }
func (a *App) customRootsFile() string   { return filepath.Join(a.dataDir, "custom-roots.json") }
func (a *App) imageMetaCacheFile() string {
	return filepath.Join(a.dataDir, "image-meta-cache.json")
}
func (a *App) imageVariantsDir() string { return filepath.Join(a.dataDir, "image-variants") }
func (a *App) previewVariantsDir() string {
	return filepath.Join(a.imageVariantsDir(), "preview")
}
func (a *App) thumbVariantsDir() string { return filepath.Join(a.imageVariantsDir(), "thumb") }
func (a *App) iconsDir() string         { return filepath.Join(a.dataDir, "icons") }
func (a *App) trashDir() string         { return filepath.Join(a.appDir, ".trash") }
func (a *App) legacyTrashDir() string {
	return filepath.Join(a.imageDir, ".trash")
}

func (a *App) trashRelPath(filename string) string {
	trashPath := filepath.Join(a.trashDir(), filename)
	rel, err := filepath.Rel(a.rootDir, trashPath)
	if err != nil {
		return ""
	}
	return normalizeRelPath(rel)
}

func moveFile(sourcePath, destPath string) error {
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}
	input, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	input.Close()
	output.Close()
	return os.Remove(sourcePath)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func defaultFavoriteGroup() FavoriteGroup {
	return FavoriteGroup{
		ID:    defaultFavoriteGroupID,
		Name:  defaultFavoriteGroupName,
		Paths: []string{},
	}
}

func uniqueNonEmptyStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
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

		name := strings.TrimSpace(group.Name)
		if name == "" {
			if id == defaultFavoriteGroupID {
				name = defaultFavoriteGroupName
			} else {
				name = "未命名分组"
			}
		}

		paths := uniqueNonEmptyStrings(group.Paths)
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

// --- Images ---

func (a *App) GetImages(sortBy, sortOrder string) ([]ImageFile, error) {
	a.ensureImageMetaCacheLoaded()
	cachedMeta := a.snapshotImageMetaCache()

	images := []ImageFile{}
	newCache := make(ImageMetaCache, len(cachedMeta))
	warmupTasks := make([]imageMetaWarmupTask, 0)
	cacheChanged := len(cachedMeta) == 0

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
		name := filepath.Base(path)
		width, height := 0, 0

		if cached, ok := cachedMeta[relPath]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
			if cached.Width > 0 || cached.Height > 0 {
				width = cached.Width
				height = cached.Height
			} else if sortBy == "dimensions" {
				width, height = readImageDimensions(path)
				cacheChanged = true
			} else {
				warmupTasks = append(warmupTasks, imageMetaWarmupTask{
					Path: path,
					Entry: ImageMetaCacheEntry{
						Name:    name,
						RelPath: relPath,
						ModTime: modTime,
						Size:    info.Size(),
					},
				})
			}
		} else if sortBy == "dimensions" {
			width, height = readImageDimensions(path)
			cacheChanged = true
		} else {
			warmupTasks = append(warmupTasks, imageMetaWarmupTask{
				Path: path,
				Entry: ImageMetaCacheEntry{
					Name:    name,
					RelPath: relPath,
					ModTime: modTime,
					Size:    info.Size(),
				},
			})
			cacheChanged = true
		}

		entry := ImageMetaCacheEntry{
			Name:    name,
			RelPath: relPath,
			ModTime: modTime,
			Size:    info.Size(),
			Width:   width,
			Height:  height,
		}
		if cached, ok := cachedMeta[relPath]; ok {
			entry.MetadataScanned = cached.MetadataScanned
			entry.HasMetadata = cached.HasMetadata
			entry.HasWorkflow = cached.HasWorkflow
			entry.Positive = cached.Positive
			entry.Negative = cached.Negative
			entry.Model = cached.Model
			entry.Sampler = cached.Sampler
			if len(cached.Loras) > 0 {
				entry.Loras = append([]string(nil), cached.Loras...)
			}
			entry.SearchText = cached.SearchText

			if cached.Name != entry.Name ||
				cached.RelPath != entry.RelPath ||
				cached.ModTime != entry.ModTime ||
				cached.Size != entry.Size ||
				cached.Width != entry.Width ||
				cached.Height != entry.Height {
				cacheChanged = true
			}
		} else {
			cacheChanged = true
		}
		newCache[relPath] = entry

		images = append(images, ImageFile{
			Name:    name,
			Path:    relPath,
			RelPath: relPath,
			ModTime: info.ModTime().Format(time.RFC3339),
			Size:    info.Size(),
			Width:   width,
			Height:  height,
			Prompt:  entry.Positive,
			Model:   entry.Model,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(newCache) != len(cachedMeta) {
		cacheChanged = true
	}

	a.replaceImageMetaCache(newCache)
	if cacheChanged {
		if err := a.saveImageMetaCache(newCache); err != nil {
			log.Printf("failed to save image metadata cache: %v", err)
		}
	}
	if sortBy != "dimensions" {
		a.scheduleImageMetaWarmup(warmupTasks)
	}

	sort.Slice(images, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "time":
			less = images[i].ModTime < images[j].ModTime
		case "size":
			less = images[i].Size < images[j].Size
		case "name":
			less = images[i].Name < images[j].Name
		case "dimensions":
			less = (images[i].Width * images[i].Height) < (images[j].Width * images[j].Height)
		default:
			less = images[i].ModTime < images[j].ModTime
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})

	return images, nil
}

func (a *App) DeleteImage(relPath string) error {
	relPath = normalizeRelPath(relPath)
	targetPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid filename")
	}
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	fileName := filepath.Base(targetPath)
	trashPath := filepath.Join(a.trashDir(), fileName)

	if _, err := os.Stat(trashPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		trashPath = filepath.Join(a.trashDir(), fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	if err := moveFile(targetPath, trashPath); err != nil {
		return fmt.Errorf("failed to move to trash: %v", err)
	}

	meta, _ := a.loadTrashMetadata()
	if meta == nil {
		meta = make(TrashMetadataMap)
	}
	trashFilename := filepath.Base(trashPath)
	meta[trashFilename] = TrashMetadata{
		OriginalPath: relPath,
		DeletedAt:    time.Now().Format(time.RFC3339),
	}
	a.saveTrashMetadata(meta)

	return nil
}

// --- Favorites ---

func (a *App) loadFavoriteGroups() ([]FavoriteGroup, error) {
	data, err := os.ReadFile(a.favoritesFile())
	if err != nil {
		if os.IsNotExist(err) {
			return []FavoriteGroup{defaultFavoriteGroup()}, nil
		}
		return nil, err
	}

	store := favoriteGroupsStore{}
	if err := json.Unmarshal(data, &store); err == nil && len(store.Groups) > 0 {
		return normalizeFavoriteGroups(store.Groups), nil
	}

	var legacyGroups []FavoriteGroup
	if err := json.Unmarshal(data, &legacyGroups); err == nil && len(legacyGroups) > 0 {
		return normalizeFavoriteGroups(legacyGroups), nil
	}

	var legacyPaths []string
	if err := json.Unmarshal(data, &legacyPaths); err == nil {
		return []FavoriteGroup{
			{
				ID:    defaultFavoriteGroupID,
				Name:  defaultFavoriteGroupName,
				Paths: uniqueNonEmptyStrings(legacyPaths),
			},
		}, nil
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
	} else if action == "remove" {
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
	} else {
		return 0, fmt.Errorf("invalid action")
	}
}

// --- Tags ---

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

	// Remove tag from all images
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

// --- Image Tags ---

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
	return imageTags, nil
}

func (a *App) saveImageTags(imageTags ImageTagsMap) error {
	data, _ := json.MarshalIndent(imageTags, "", "  ")
	return os.WriteFile(a.imageTagsFile(), data, 0644)
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

// --- Trash Metadata ---

func (a *App) loadTrashMetadataRaw() (TrashMetadataMap, error) {
	var meta TrashMetadataMap
	data, err := os.ReadFile(a.trashMetadataFile())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		_ = json.Unmarshal(data, &meta)
	}
	if meta == nil {
		meta = make(TrashMetadataMap)
	}
	return meta, nil
}

func (a *App) loadTrashMetadata() (TrashMetadataMap, error) {
	meta, err := a.loadTrashMetadataRaw()
	if err != nil {
		return nil, err
	}

	// Sync with actual .trash folder
	entries, err := os.ReadDir(a.trashDir())
	if err == nil {
		changed := false
		existingFiles := make(map[string]bool)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			existingFiles[name] = true

			if _, exists := meta[name]; !exists {
				info, _ := entry.Info()
				deletedAt := time.Now()
				if info != nil {
					deletedAt = info.ModTime()
				}
				dateFolder := deletedAt.Format("2006-01-02")
				meta[name] = TrashMetadata{
					OriginalPath: filepath.ToSlash(filepath.Join(dateFolder, name)),
					DeletedAt:    deletedAt.Format(time.RFC3339),
				}
				changed = true
			}
		}

		for name := range meta {
			if !existingFiles[name] {
				delete(meta, name)
				changed = true
			}
		}
		if changed {
			a.saveTrashMetadata(meta)
		}
	}

	return meta, nil
}

func uniqueTrashFilename(name string, existing map[string]bool) string {
	if !existing[name] {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	timestamp := time.Now().Format("20060102_150405")
	candidate := fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	index := 1
	for existing[candidate] {
		candidate = fmt.Sprintf("%s_%s_%d%s", base, timestamp, index, ext)
		index++
	}
	return candidate
}

func (a *App) migrateLegacyTrash() error {
	legacyDir := a.legacyTrashDir()
	currentDir := a.trashDir()
	if legacyDir == "" || currentDir == "" || samePath(legacyDir, currentDir) {
		return nil
	}

	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return err
	}

	meta, err := a.loadTrashMetadataRaw()
	if err != nil {
		return err
	}

	existing := make(map[string]bool)
	if currentEntries, readErr := os.ReadDir(currentDir); readErr == nil {
		for _, entry := range currentEntries {
			if entry.IsDir() {
				continue
			}
			existing[entry.Name()] = true
		}
	}

	metaChanged := false
	movedAny := false

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		sourceName := entry.Name()
		sourcePath := filepath.Join(legacyDir, sourceName)
		targetName := uniqueTrashFilename(sourceName, existing)
		targetPath := filepath.Join(currentDir, targetName)

		if err := moveFile(sourcePath, targetPath); err != nil {
			log.Printf("failed to migrate trash file %s: %v", sourceName, err)
			continue
		}

		existing[targetName] = true
		movedAny = true

		if targetName != sourceName {
			if item, ok := meta[sourceName]; ok {
				delete(meta, sourceName)
				meta[targetName] = item
			} else {
				info, _ := os.Stat(targetPath)
				deletedAt := time.Now()
				if info != nil {
					deletedAt = info.ModTime()
				}
				dateFolder := deletedAt.Format("2006-01-02")
				meta[targetName] = TrashMetadata{
					OriginalPath: filepath.ToSlash(filepath.Join(dateFolder, targetName)),
					DeletedAt:    deletedAt.Format(time.RFC3339),
				}
			}
			metaChanged = true
		}
	}

	if metaChanged {
		if err := a.saveTrashMetadata(meta); err != nil {
			return err
		}
	}

	if movedAny {
		if remaining, readErr := os.ReadDir(legacyDir); readErr == nil && len(remaining) == 0 {
			_ = os.Remove(legacyDir)
		}
	}

	return nil
}

func (a *App) saveTrashMetadata(meta TrashMetadataMap) error {
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(a.trashMetadataFile(), data, 0644)
}

func (a *App) loadSettings() (Settings, error) {
	var settings Settings
	data, err := os.ReadFile(a.settingsFile())
	if err != nil {
		return Settings{TrashRetentionDays: 30}, nil
	}
	json.Unmarshal(data, &settings)
	if settings.TrashRetentionDays <= 0 {
		settings.TrashRetentionDays = 30
	}
	return settings, nil
}

func (a *App) saveSettings(settings Settings) error {
	data, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(a.settingsFile(), data, 0644)
}

func (a *App) loadImageMetaCache() (ImageMetaCache, error) {
	var cache ImageMetaCache
	data, err := os.ReadFile(a.imageMetaCacheFile())
	if err != nil {
		if os.IsNotExist(err) {
			return ImageMetaCache{}, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		return ImageMetaCache{}, nil
	}
	if cache == nil {
		cache = ImageMetaCache{}
	}
	return cache, nil
}

func (a *App) saveImageMetaCache(cache ImageMetaCache) error {
	data, _ := json.MarshalIndent(cache, "", "  ")
	return os.WriteFile(a.imageMetaCacheFile(), data, 0644)
}

func measureDirectoryUsage(root string) (int, int, int64, error) {
	files := 0
	dirs := 0
	var bytesFreed int64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			dirs++
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		files++
		bytesFreed += info.Size()
		return nil
	})

	return files, dirs, bytesFreed, err
}

func (a *App) ClearPreviewCache() (CacheClearResult, error) {
	result := CacheClearResult{}
	var firstErr error

	if _, err := os.Stat(a.imageVariantsDir()); err == nil {
		files, dirs, bytesFreed, measureErr := measureDirectoryUsage(a.imageVariantsDir())
		if measureErr != nil {
			firstErr = measureErr
		} else {
			result.DeletedFiles += files
			result.DeletedDirs += dirs
			result.BytesFreed += bytesFreed
		}

		if err := os.RemoveAll(a.imageVariantsDir()); err != nil && firstErr == nil {
			firstErr = err
		}
	} else if err != nil && !os.IsNotExist(err) {
		firstErr = err
	}

	if err := os.MkdirAll(a.previewVariantsDir(), 0755); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := os.MkdirAll(a.thumbVariantsDir(), 0755); err != nil && firstErr == nil {
		firstErr = err
	}

	return result, firstErr
}

func (a *App) ensureImageMetaCacheLoaded() {
	a.imageMetaMu.Lock()
	defer a.imageMetaMu.Unlock()

	if a.imageMetaLoaded {
		return
	}

	cache, err := a.loadImageMetaCache()
	if err != nil {
		log.Printf("failed to load image metadata cache: %v", err)
		cache = ImageMetaCache{}
	}

	a.imageMetaCache = cache
	a.imageMetaLoaded = true
}

func (a *App) snapshotImageMetaCache() ImageMetaCache {
	a.imageMetaMu.RLock()
	defer a.imageMetaMu.RUnlock()

	cache := make(ImageMetaCache, len(a.imageMetaCache))
	for key, value := range a.imageMetaCache {
		cache[key] = value
	}
	return cache
}

func (a *App) replaceImageMetaCache(cache ImageMetaCache) {
	a.imageMetaMu.Lock()
	defer a.imageMetaMu.Unlock()

	a.imageMetaCache = cache
	a.imageMetaLoaded = true
}

func readImageDimensions(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func parsePNGTextChunks(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	signature := make([]byte, 8)
	if _, err := io.ReadFull(file, signature); err != nil {
		return nil, err
	}
	if !bytes.Equal(signature, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return nil, fmt.Errorf("not a png file")
	}

	chunks := make(map[string]string)
	for {
		var length uint32
		if err := binary.Read(file, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				return chunks, nil
			}
			return nil, err
		}

		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(file, chunkType); err != nil {
			return nil, err
		}

		if length > 32<<20 {
			return nil, fmt.Errorf("png chunk too large")
		}

		chunkData := make([]byte, length)
		if _, err := io.ReadFull(file, chunkData); err != nil {
			return nil, err
		}
		if _, err := io.CopyN(io.Discard, file, 4); err != nil {
			return nil, err
		}

		switch string(chunkType) {
		case "tEXt":
			key, value, ok := parsePNGTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "zTXt":
			key, value, ok := parsePNGCompressedTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "iTXt":
			key, value, ok := parsePNGInternationalTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "IEND":
			return chunks, nil
		}
	}
}

func parsePNGTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(string(data[:separator]))
	value := string(data[separator+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func parsePNGCompressedTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 || separator+2 > len(data) {
		return "", "", false
	}
	if data[separator+1] != 0 {
		return "", "", false
	}

	reader, err := zlib.NewReader(bytes.NewReader(data[separator+2:]))
	if err != nil {
		return "", "", false
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", "", false
	}

	key := strings.TrimSpace(string(data[:separator]))
	if key == "" {
		return "", "", false
	}
	return key, string(decoded), true
}

func parsePNGInternationalTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 || separator+5 > len(data) {
		return "", "", false
	}

	key := strings.TrimSpace(string(data[:separator]))
	if key == "" {
		return "", "", false
	}

	compressionFlag := data[separator+1]
	compressionMethod := data[separator+2]
	rest := data[separator+3:]

	languageEnd := bytes.IndexByte(rest, 0)
	if languageEnd < 0 {
		return "", "", false
	}
	rest = rest[languageEnd+1:]

	translatedEnd := bytes.IndexByte(rest, 0)
	if translatedEnd < 0 {
		return "", "", false
	}
	textData := rest[translatedEnd+1:]

	if compressionFlag == 1 {
		if compressionMethod != 0 {
			return "", "", false
		}
		reader, err := zlib.NewReader(bytes.NewReader(textData))
		if err != nil {
			return "", "", false
		}
		defer reader.Close()

		decoded, err := io.ReadAll(reader)
		if err != nil {
			return "", "", false
		}
		return key, string(decoded), true
	}

	return key, string(textData), true
}

func decodeJSONUseNumber(raw string, target any) error {
	toDecode := raw
	probeDecoder := json.NewDecoder(strings.NewReader(raw))
	probeDecoder.UseNumber()
	var probe any
	if err := probeDecoder.Decode(&probe); err != nil {
		if sanitized, changed := sanitizeJSONSpecialNumbers(raw); changed {
			toDecode = sanitized
		}
	}

	decoder := json.NewDecoder(strings.NewReader(toDecode))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func sanitizeJSONSpecialNumbers(raw string) (string, bool) {
	var builder strings.Builder
	builder.Grow(len(raw))

	inString := false
	escaped := false
	changed := false

	for i := 0; i < len(raw); {
		ch := raw[i]

		if inString {
			builder.WriteByte(ch)
			if escaped {
				escaped = false
			} else {
				if ch == '\\' {
					escaped = true
				} else if ch == '"' {
					inString = false
				}
			}
			i++
			continue
		}

		if ch == '"' {
			inString = true
			builder.WriteByte(ch)
			i++
			continue
		}

		if strings.HasPrefix(raw[i:], "-Infinity") && isJSONSpecialTokenBoundary(raw, i, 9) {
			builder.WriteString("null")
			i += 9
			changed = true
			continue
		}
		if strings.HasPrefix(raw[i:], "Infinity") && isJSONSpecialTokenBoundary(raw, i, 8) {
			builder.WriteString("null")
			i += 8
			changed = true
			continue
		}
		if strings.HasPrefix(raw[i:], "NaN") && isJSONSpecialTokenBoundary(raw, i, 3) {
			builder.WriteString("null")
			i += 3
			changed = true
			continue
		}

		builder.WriteByte(ch)
		i++
	}

	if !changed {
		return raw, false
	}
	return builder.String(), true
}

func isJSONSpecialTokenBoundary(raw string, start, tokenLen int) bool {
	if start > 0 {
		prev := raw[start-1]
		if !isJSONSpecialTokenDelimiter(prev) {
			return false
		}
	}

	end := start + tokenLen
	if end < len(raw) {
		next := raw[end]
		if !isJSONSpecialTokenDelimiter(next) {
			return false
		}
	}

	return true
}

func isJSONSpecialTokenDelimiter(ch byte) bool {
	switch ch {
	case ' ', '\n', '\r', '\t', ':', ',', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func stringifyMetadataValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		if data, err := json.Marshal(v); err == nil {
			return strings.TrimSpace(string(data))
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func directTextInput(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

type multiCharacterEditorConfig struct {
	BasePrompt   string                          `json:"base_prompt"`
	GlobalPrompt string                          `json:"global_prompt"`
	Characters   []multiCharacterEditorCharacter `json:"characters"`
}

type multiCharacterEditorCharacter struct {
	Prompt  string `json:"prompt"`
	Enabled *bool  `json:"enabled"`
}

func extractMultiCharacterEditorPrompt(value any) string {
	raw := directTextInput(value)
	if raw == "" {
		return ""
	}

	var config multiCharacterEditorConfig
	if err := decodeJSONUseNumber(raw, &config); err != nil {
		return ""
	}

	parts := make([]string, 0, len(config.Characters)+2)
	parts = append(parts, config.BasePrompt, config.GlobalPrompt)
	for _, character := range config.Characters {
		if character.Enabled != nil && !*character.Enabled {
			continue
		}
		parts = append(parts, character.Prompt)
	}

	return joinMetadataTexts(parts...)
}

func joinMetadataTexts(parts ...string) string {
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return strings.Join(result, "\n\n")
}

func appendUniqueTexts(target []string, values ...string) []string {
	seen := make(map[string]struct{}, len(target))
	for _, item := range target {
		seen[item] = struct{}{}
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		target = append(target, value)
	}
	return target
}

func extractNodePromptTexts(node comfyPromptNode) (string, string) {
	positive := joinMetadataTexts(
		directTextInput(node.Inputs["text"]),
		directTextInput(node.Inputs["text_g"]),
		directTextInput(node.Inputs["text_l"]),
		directTextInput(node.Inputs["string"]),
		directTextInput(node.Inputs["prompt"]),
		directTextInput(node.Inputs["positive"]),
		directTextInput(node.Inputs["positive_prompt"]),
		directTextInput(node.Inputs["text_positive"]),
		extractMultiCharacterEditorPrompt(node.Inputs["mce_config"]),
	)
	negative := joinMetadataTexts(
		directTextInput(node.Inputs["negative"]),
		directTextInput(node.Inputs["negative_prompt"]),
		directTextInput(node.Inputs["text_negative"]),
	)

	lowerClass := strings.ToLower(node.ClassType)
	if negative == "" && positive != "" && strings.Contains(lowerClass, "negative") {
		negative = positive
		positive = ""
	}
	if positive == "" && negative != "" && strings.Contains(lowerClass, "positive") {
		positive = negative
		negative = ""
	}
	return positive, negative
}

func collectFallbackPromptTexts(nodes map[string]comfyPromptNode) ([]string, []string) {
	positiveTexts := make([]string, 0, 2)
	negativeTexts := make([]string, 0, 2)

	for _, id := range sortedPromptNodeIDs(nodes) {
		node := nodes[id]
		positive, negative := extractNodePromptTexts(node)
		lowerClass := strings.ToLower(node.ClassType)

		if positive != "" {
			if strings.Contains(lowerClass, "negative") && !strings.Contains(lowerClass, "positive") {
				negativeTexts = appendUniqueTexts(negativeTexts, positive)
			} else {
				positiveTexts = appendUniqueTexts(positiveTexts, positive)
			}
		}
		if negative != "" {
			negativeTexts = appendUniqueTexts(negativeTexts, negative)
		}
	}

	return positiveTexts, negativeTexts
}

func sortedPromptNodeIDs(nodes map[string]comfyPromptNode) []string {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		left, leftErr := strconv.Atoi(ids[i])
		right, rightErr := strconv.Atoi(ids[j])
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return ids[i] < ids[j]
	})
	return ids
}

func connectedPromptNodeID(value any) (string, bool) {
	connection, ok := value.([]any)
	if !ok || len(connection) == 0 {
		return "", false
	}
	id := strings.TrimSpace(fmt.Sprint(connection[0]))
	return id, id != ""
}

func resolvePromptTextForKeys(nodes map[string]comfyPromptNode, value any, preferredKeys []string, visited map[string]bool) string {
	if text := directTextInput(value); text != "" {
		return text
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return ""
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return ""
	}

	parts := make([]string, 0, len(preferredKeys))
	for _, key := range preferredKeys {
		parts = append(parts, directTextInput(node.Inputs[key]))
	}
	parts = append(parts, extractMultiCharacterEditorPrompt(node.Inputs["mce_config"]))
	if combined := joinMetadataTexts(parts...); combined != "" {
		return combined
	}

	for _, key := range []string{"text", "conditioning", "positive", "negative", "clip", "sdxl_tuple", "text_positive", "text_negative", "prompt"} {
		if next, exists := node.Inputs[key]; exists {
			if text := resolvePromptTextForKeys(nodes, next, preferredKeys, visited); text != "" {
				return text
			}
		}
	}

	return ""
}

func resolvePromptText(nodes map[string]comfyPromptNode, value any, visited map[string]bool) string {
	return resolvePromptTextForKeys(
		nodes,
		value,
		[]string{
			"text",
			"text_g",
			"text_l",
			"string",
			"prompt",
			"positive",
			"positive_prompt",
			"text_positive",
			"negative",
			"negative_prompt",
			"text_negative",
		},
		visited,
	)
}

func collectPromptModel(nodes map[string]comfyPromptNode, value any, visited map[string]bool, loras map[string]struct{}) string {
	if model := stringifyMetadataValue(value); model != "" && strings.HasSuffix(strings.ToLower(model), ".safetensors") {
		return model
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return ""
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return ""
	}

	switch node.ClassType {
	case "CheckpointLoaderSimple", "CheckpointLoader":
		return stringifyMetadataValue(node.Inputs["ckpt_name"])
	case "Efficient Loader", "Eff. Loader SDXL":
		if model := stringifyMetadataValue(node.Inputs["ckpt_name"]); model != "" {
			return model
		}
		if model := stringifyMetadataValue(node.Inputs["base_ckpt_name"]); model != "" {
			return model
		}
		if model := stringifyMetadataValue(node.Inputs["refiner_ckpt_name"]); model != "" {
			return model
		}
	case "LoraLoader", "LoraLoaderModelOnly":
		if lora := stringifyMetadataValue(node.Inputs["lora_name"]); lora != "" {
			loras[lora] = struct{}{}
		}
	}

	for _, index := range []string{
		"lora_name", "lora_name_1", "lora_name_2", "lora_name_3", "lora_name_4", "lora_name_5",
		"lora_name_6", "lora_name_7", "lora_name_8", "lora_name_9", "lora_name_10",
	} {
		if lora := stringifyMetadataValue(node.Inputs[index]); lora != "" && strings.ToLower(lora) != "none" {
			loras[lora] = struct{}{}
		}
	}

	for _, key := range []string{"model", "clip", "base_model", "sdxl_tuple"} {
		if next, exists := node.Inputs[key]; exists {
			if model := collectPromptModel(nodes, next, visited, loras); model != "" {
				return model
			}
		}
	}

	return ""
}

func collectPromptLoras(nodes map[string]comfyPromptNode, value any, visited map[string]bool, loras map[string]struct{}) {
	if lora := stringifyMetadataValue(value); lora != "" && strings.HasSuffix(strings.ToLower(lora), ".safetensors") {
		loras[lora] = struct{}{}
		return
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return
	}

	if lora := stringifyMetadataValue(node.Inputs["lora_name"]); lora != "" {
		loras[lora] = struct{}{}
	}
	for _, index := range []string{
		"lora_name_1", "lora_name_2", "lora_name_3", "lora_name_4", "lora_name_5",
		"lora_name_6", "lora_name_7", "lora_name_8", "lora_name_9", "lora_name_10",
		"lora_name_11", "lora_name_12", "lora_name_13", "lora_name_14", "lora_name_15",
		"lora_name_16", "lora_name_17", "lora_name_18", "lora_name_19", "lora_name_20",
		"lora_name_21", "lora_name_22", "lora_name_23", "lora_name_24", "lora_name_25",
		"lora_name_26", "lora_name_27", "lora_name_28", "lora_name_29", "lora_name_30",
		"lora_name_31", "lora_name_32", "lora_name_33", "lora_name_34", "lora_name_35",
		"lora_name_36", "lora_name_37", "lora_name_38", "lora_name_39", "lora_name_40",
		"lora_name_41", "lora_name_42", "lora_name_43", "lora_name_44", "lora_name_45",
		"lora_name_46", "lora_name_47", "lora_name_48", "lora_name_49", "lora_name_50",
	} {
		if lora := stringifyMetadataValue(node.Inputs[index]); lora != "" && strings.ToLower(lora) != "none" {
			loras[lora] = struct{}{}
		}
	}

	for _, key := range []string{"model", "clip", "conditioning", "positive", "negative", "sdxl_tuple"} {
		if next, exists := node.Inputs[key]; exists {
			collectPromptLoras(nodes, next, visited, loras)
		}
	}
}

func extractComfyPromptSummary(metadata *ImageMetadata, promptRaw string) {
	var nodes map[string]comfyPromptNode
	if err := decodeJSONUseNumber(promptRaw, &nodes); err != nil || len(nodes) == 0 {
		return
	}

	ids := sortedPromptNodeIDs(nodes)
	var samplerNode comfyPromptNode
	foundSampler := false
	preferredSamplerClasses := []string{
		"KSampler",
		"KSamplerAdvanced",
		"KSampler (Efficient)",
		"KSampler (Eff.)",
		"KSampler SDXL (Eff.)",
		"LanPaint_KSampler",
		"SamplerCustom",
		"SamplerCustomAdvanced",
	}

	for _, classType := range preferredSamplerClasses {
		for _, id := range ids {
			if nodes[id].ClassType == classType {
				samplerNode = nodes[id]
				foundSampler = true
				break
			}
		}
		if foundSampler {
			break
		}
	}

	if foundSampler {
		metadata.Seed = stringifyMetadataValue(samplerNode.Inputs["seed"])
		if metadata.Seed == "" {
			metadata.Seed = stringifyMetadataValue(samplerNode.Inputs["noise_seed"])
		}
		metadata.Steps = stringifyMetadataValue(samplerNode.Inputs["steps"])
		metadata.CFG = stringifyMetadataValue(samplerNode.Inputs["cfg"])
		metadata.Sampler = stringifyMetadataValue(samplerNode.Inputs["sampler_name"])
		metadata.Scheduler = stringifyMetadataValue(samplerNode.Inputs["scheduler"])
		metadata.Positive = resolvePromptTextForKeys(nodes, samplerNode.Inputs["positive"], []string{"text", "text_g", "text_l", "string", "prompt", "positive", "positive_prompt", "text_positive"}, map[string]bool{})
		metadata.Negative = resolvePromptTextForKeys(nodes, samplerNode.Inputs["negative"], []string{"negative", "negative_prompt", "text_negative", "text", "text_g", "text_l", "string"}, map[string]bool{})

		loras := make(map[string]struct{})
		metadata.Model = collectPromptModel(nodes, samplerNode.Inputs["model"], map[string]bool{}, loras)
		if metadata.Model == "" {
			metadata.Model = collectPromptModel(nodes, samplerNode.Inputs["sdxl_tuple"], map[string]bool{}, loras)
		}
		if metadata.Positive == "" {
			metadata.Positive = resolvePromptTextForKeys(nodes, samplerNode.Inputs["sdxl_tuple"], []string{"positive", "positive_prompt", "text_positive", "text", "text_g", "text_l", "string", "prompt"}, map[string]bool{})
		}
		if metadata.Negative == "" {
			if text := resolvePromptTextForKeys(nodes, samplerNode.Inputs["sdxl_tuple"], []string{"negative", "negative_prompt", "text_negative", "text", "text_g", "text_l", "string"}, map[string]bool{}); text != metadata.Positive {
				metadata.Negative = text
			}
		}
		collectPromptLoras(nodes, samplerNode.Inputs["positive"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["negative"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["model"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["sdxl_tuple"], map[string]bool{}, loras)
		if len(loras) > 0 {
			metadata.Loras = metadata.Loras[:0]
			for lora := range loras {
				metadata.Loras = append(metadata.Loras, lora)
			}
			sort.Strings(metadata.Loras)
		}
	}

	if metadata.Positive == "" || metadata.Negative == "" {
		fallbackPositive, fallbackNegative := collectFallbackPromptTexts(nodes)
		if metadata.Positive == "" && len(fallbackPositive) > 0 {
			metadata.Positive = fallbackPositive[0]
		}
		if metadata.Negative == "" && len(fallbackNegative) > 0 {
			metadata.Negative = fallbackNegative[0]
		}
		if metadata.Positive == "" && len(fallbackPositive) == 0 && len(fallbackNegative) == 0 {
			textNodes := make([]string, 0, 2)
			for _, id := range ids {
				node := nodes[id]
				if !strings.Contains(strings.ToLower(node.ClassType), "textencode") {
					continue
				}
				positive, _ := extractNodePromptTexts(node)
				if positive != "" {
					textNodes = appendUniqueTexts(textNodes, positive)
				}
			}
			if len(textNodes) > 0 {
				metadata.Positive = textNodes[0]
			}
			if metadata.Negative == "" && len(textNodes) > 1 {
				metadata.Negative = textNodes[1]
			}
		}
	}

	if metadata.Model == "" {
		for _, id := range ids {
			node := nodes[id]
			switch node.ClassType {
			case "CheckpointLoaderSimple", "CheckpointLoader", "CheckpointLoaderNF4", "UNETLoader":
				metadata.Model = stringifyMetadataValue(node.Inputs["ckpt_name"])
				if metadata.Model == "" {
					metadata.Model = stringifyMetadataValue(node.Inputs["unet_name"])
				}
			}
			if metadata.Model != "" {
				break
			}
		}
	}

	if len(metadata.Loras) == 0 {
		loras := make([]string, 0, 4)
		for _, id := range ids {
			if lora := stringifyMetadataValue(nodes[id].Inputs["lora_name"]); lora != "" {
				loras = appendUniqueTexts(loras, lora)
			}
		}
		if len(loras) > 0 {
			sort.Strings(loras)
			metadata.Loras = loras
		}
	}
}

func parseAutomatic1111Parameters(metadata *ImageMetadata, parameters string) {
	normalized := strings.ReplaceAll(parameters, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return
	}

	paramsLine := strings.TrimSpace(lines[len(lines)-1])
	contentLines := lines[:len(lines)-1]
	negativeIndex := -1
	for index, line := range contentLines {
		if strings.HasPrefix(line, "Negative prompt:") {
			negativeIndex = index
			break
		}
	}

	if negativeIndex >= 0 {
		metadata.Positive = strings.TrimSpace(strings.Join(contentLines[:negativeIndex], "\n"))
		negativeLines := append([]string{strings.TrimSpace(strings.TrimPrefix(contentLines[negativeIndex], "Negative prompt:"))}, contentLines[negativeIndex+1:]...)
		metadata.Negative = strings.TrimSpace(strings.Join(negativeLines, "\n"))
	} else {
		metadata.Positive = strings.TrimSpace(strings.Join(contentLines, "\n"))
	}

	for _, part := range strings.Split(paramsLine, ",") {
		pair := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(pair[0]))
		value := strings.TrimSpace(pair[1])
		switch key {
		case "steps":
			metadata.Steps = value
		case "sampler":
			metadata.Sampler = value
		case "cfg scale":
			metadata.CFG = value
		case "seed":
			metadata.Seed = value
		case "model":
			metadata.Model = value
		}
	}
}

func buildImageMetadata(relPath string, width, height int, textChunks map[string]string) ImageMetadata {
	metadata := ImageMetadata{
		RelPath:     relPath,
		Format:      strings.TrimPrefix(strings.ToLower(filepath.Ext(relPath)), "."),
		Width:       width,
		Height:      height,
		HasMetadata: len(textChunks) > 0,
		ExtraFields: make(map[string]string),
	}

	for key, value := range textChunks {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		switch {
		case lowerKey == "prompt":
			metadata.Prompt = value
		case lowerKey == "workflow":
			metadata.Workflow = value
		case lowerKey == "parameters":
			metadata.ExtraFields[key] = value
			if metadata.Positive == "" && metadata.Negative == "" {
				parseAutomatic1111Parameters(&metadata, value)
			}
		case strings.Contains(lowerKey, "workflow") && metadata.Workflow == "":
			metadata.Workflow = value
			metadata.ExtraFields[key] = value
		case strings.Contains(lowerKey, "prompt") && metadata.Prompt == "" && strings.HasPrefix(strings.TrimSpace(value), "{"):
			metadata.Prompt = value
			metadata.ExtraFields[key] = value
		default:
			metadata.ExtraFields[key] = value
		}
	}

	if metadata.Prompt != "" {
		extractComfyPromptSummary(&metadata, metadata.Prompt)
	}

	if metadata.Workflow != "" {
		var workflow struct {
			Nodes []any `json:"nodes"`
		}
		if err := decodeJSONUseNumber(metadata.Workflow, &workflow); err == nil {
			metadata.NodeCount = len(workflow.Nodes)
		}
	}

	if len(metadata.ExtraFields) == 0 {
		metadata.ExtraFields = nil
	}

	return metadata
}

func (a *App) GetImageMetadata(relPath string) (ImageMetadata, error) {
	normalized := normalizeRelPath(relPath)
	if normalized == "" {
		return ImageMetadata{}, fmt.Errorf("invalid path")
	}

	absPath, err := a.resolveRootPath(normalized)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("invalid path")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ImageMetadata{}, fmt.Errorf("file not found")
		}
		return ImageMetadata{}, err
	}

	a.ensureImageMetaCacheLoaded()
	width, height := 0, 0
	modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
	if cached, ok := a.snapshotImageMetaCache()[normalized]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
		width = cached.Width
		height = cached.Height
	}
	if width == 0 && height == 0 {
		width, height = readImageDimensions(absPath)
	}

	metadata := ImageMetadata{
		RelPath: normalized,
		Format:  strings.TrimPrefix(strings.ToLower(filepath.Ext(absPath)), "."),
		Width:   width,
		Height:  height,
	}

	if strings.ToLower(filepath.Ext(absPath)) != ".png" {
		return metadata, nil
	}

	textChunks, err := parsePNGTextChunks(absPath)
	if err != nil {
		return metadata, err
	}

	return buildImageMetadata(normalized, width, height, textChunks), nil
}

func (a *App) CopyText(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}

func (a *App) scheduleImageMetaWarmup(tasks []imageMetaWarmupTask) {
	if len(tasks) == 0 {
		return
	}

	a.imageMetaMu.Lock()
	if a.imageMetaWarmupRunning {
		a.imageMetaMu.Unlock()
		return
	}
	a.imageMetaWarmupRunning = true
	a.imageMetaMu.Unlock()

	go func(pending []imageMetaWarmupTask) {
		defer func() {
			a.imageMetaMu.Lock()
			a.imageMetaWarmupRunning = false
			a.imageMetaMu.Unlock()
		}()

		updated := false
		for _, task := range pending {
			width, height := readImageDimensions(task.Path)
			if width == 0 && height == 0 {
				continue
			}

			a.imageMetaMu.Lock()
			entry, ok := a.imageMetaCache[task.Entry.RelPath]
			if ok && entry.ModTime == task.Entry.ModTime && entry.Size == task.Entry.Size {
				if entry.Width != width || entry.Height != height {
					entry.Width = width
					entry.Height = height
					a.imageMetaCache[task.Entry.RelPath] = entry
					updated = true
				}
			}
			a.imageMetaMu.Unlock()
		}

		if updated {
			if err := a.saveImageMetaCache(a.snapshotImageMetaCache()); err != nil {
				log.Printf("failed to save warmed image metadata cache: %v", err)
			}
		}
	}(tasks)
}

func (a *App) cleanExpiredTrash() error {
	settings, _ := a.loadSettings()
	meta, _ := a.loadTrashMetadata()

	cutoffTime := time.Now().AddDate(0, 0, -settings.TrashRetentionDays)
	deletedCount := 0

	for trashFilename, item := range meta {
		itemTime, _ := time.Parse(time.RFC3339, item.DeletedAt)
		if itemTime.Before(cutoffTime) {
			trashPath := filepath.Join(a.trashDir(), trashFilename)
			if err := os.Remove(trashPath); err == nil || os.IsNotExist(err) {
				delete(meta, trashFilename)
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		return a.saveTrashMetadata(meta)
	}
	return nil
}

func (a *App) GetTrashList() ([]TrashItem, error) {
	meta, _ := a.loadTrashMetadata()
	items := []TrashItem{}

	for filename, itemMeta := range meta {
		items = append(items, TrashItem{
			Filename:     filename,
			OriginalPath: itemMeta.OriginalPath,
			DeletedAt:    itemMeta.DeletedAt,
			Path:         a.trashRelPath(filename),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DeletedAt > items[j].DeletedAt
	})

	return items, nil
}

func (a *App) RestoreTrash(filename string) (string, error) {
	meta, _ := a.loadTrashMetadata()
	itemMeta, exists := meta[filename]
	if !exists {
		return "", fmt.Errorf("file not found in trash")
	}

	trashPath := filepath.Join(a.trashDir(), filename)
	restorePath := filepath.Join(a.imageDir, itemMeta.OriginalPath)

	os.MkdirAll(filepath.Dir(restorePath), 0755)

	if _, err := os.Stat(restorePath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(restorePath)
		base := strings.TrimSuffix(restorePath, ext)
		restorePath = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	}

	if err := moveFile(trashPath, restorePath); err != nil {
		return "", err
	}

	delete(meta, filename)
	a.saveTrashMetadata(meta)

	return restorePath, nil
}

func (a *App) BatchRestoreTrash(filenames []string) (int, error) {
	meta, _ := a.loadTrashMetadata()
	successCount := 0

	for _, filename := range filenames {
		itemMeta, exists := meta[filename]
		if !exists {
			continue
		}

		trashPath := filepath.Join(a.trashDir(), filename)
		restorePath := filepath.Join(a.imageDir, itemMeta.OriginalPath)
		os.MkdirAll(filepath.Dir(restorePath), 0755)

		if _, err := os.Stat(restorePath); err == nil {
			timestamp := time.Now().Format("20060102_150405")
			ext := filepath.Ext(restorePath)
			name := strings.TrimSuffix(restorePath, ext)
			restorePath = fmt.Sprintf("%s_%s%s", name, timestamp, ext)
		}

		if err := moveFile(trashPath, restorePath); err == nil {
			delete(meta, filename)
			successCount++
		}
	}

	a.saveTrashMetadata(meta)
	return successCount, nil
}

func (a *App) BatchDeleteTrash(filenames []string) (int, error) {
	meta, _ := a.loadTrashMetadata()
	deletedCount := 0

	for _, filename := range filenames {
		trashPath := filepath.Join(a.trashDir(), filename)
		if err := os.RemoveAll(trashPath); err == nil || os.IsNotExist(err) {
			deletedCount++
			delete(meta, filename)
		}
	}

	a.saveTrashMetadata(meta)
	return deletedCount, nil
}

func (a *App) EmptyTrash() (int, error) {
	deletedCount := 0
	entries, err := os.ReadDir(a.trashDir())
	if err == nil {
		for _, entry := range entries {
			path := filepath.Join(a.trashDir(), entry.Name())
			if err := os.RemoveAll(path); err == nil {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		a.saveTrashMetadata(TrashMetadataMap{})
	}

	return deletedCount, nil
}

func (a *App) GetTrashSettings() (Settings, error) {
	return a.loadSettings()
}

func (a *App) SaveTrashSettings(settings Settings) error {
	return a.saveSettings(settings)
}

func (a *App) GetDirectoryBinding() (DirectoryBinding, error) {
	return DirectoryBinding{
		RootDir:       a.rootDir,
		OutputDir:     a.imageDir,
		OutputRelPath: a.outputRelPath(),
	}, nil
}

func (a *App) SaveDirectoryBinding(rootDir, outputDir string) (DirectoryBinding, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return DirectoryBinding{}, err
	}

	previousRoot := a.rootDir
	previousOutput := a.imageDir

	if err := a.applyDirectoryBinding(rootDir, outputDir); err != nil {
		a.rootDir = previousRoot
		a.imageDir = previousOutput
		return DirectoryBinding{}, err
	}

	settings.RootDir = a.rootDir
	settings.OutputDir = a.imageDir
	settings.PathVersion = pathVersionRootRelative
	if settings.TrashRetentionDays <= 0 {
		settings.TrashRetentionDays = 30
	}

	if err := a.saveSettings(settings); err != nil {
		a.rootDir = previousRoot
		a.imageDir = previousOutput
		return DirectoryBinding{}, err
	}

	if err := os.MkdirAll(a.trashDir(), 0755); err != nil {
		return DirectoryBinding{}, err
	}

	if err := a.migrateLegacyTrash(); err != nil {
		log.Printf("failed to migrate legacy trash after rebinding: %v", err)
	}

	return a.GetDirectoryBinding()
}

// --- Utilities ---

func (a *App) BatchMove(paths []string, targetFolder string) (int, error) {
	targetFolder = filepath.Clean(targetFolder)

	var targetPath string
	isAbs := filepath.IsAbs(targetFolder)
	hasDrive := len(targetFolder) >= 2 && targetFolder[1] == ':'

	if isAbs || hasDrive {
		if strings.ContainsAny(targetFolder, "<>\"|?*") {
			return 0, fmt.Errorf("folder path contains invalid characters")
		}
		targetPath = targetFolder
	} else {
		targetFolder = normalizeRelPath(targetFolder)
		if strings.ContainsAny(targetFolder, "<>:\"|?*") {
			return 0, fmt.Errorf("folder name contains invalid characters")
		}
		var err error
		targetPath, err = a.resolveRootPath(targetFolder)
		if err != nil {
			return 0, err
		}
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return 0, err
	}

	successCount := 0
	for _, relPath := range paths {
		sourcePath, err := a.resolveRootPath(relPath)
		if err != nil {
			continue
		}
		fileName := filepath.Base(sourcePath)
		destPath := filepath.Join(targetPath, fileName)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue
		}

		if _, err := os.Stat(destPath); err == nil {
			ext := filepath.Ext(fileName)
			name := strings.TrimSuffix(fileName, ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(targetPath, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		}

		if err := moveFile(sourcePath, destPath); err == nil {
			successCount++
		}
	}

	return successCount, nil
}

func (a *App) CleanEmptyFolders() (int, error) {
	var emptyDirs []string
	filepath.WalkDir(a.imageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".trash" || d.Name() == "desktop-source" {
				return fs.SkipDir
			}
			if path != a.imageDir {
				emptyDirs = append(emptyDirs, path)
			}
		}
		return nil
	})

	sort.Slice(emptyDirs, func(i, j int) bool {
		return len(emptyDirs[i]) > len(emptyDirs[j])
	})

	removedCount := 0
	for _, dirPath := range emptyDirs {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		isEmpty := true
		for _, entry := range entries {
			if entry.IsDir() {
				isEmpty = false
				break
			}
			nameUpper := strings.ToUpper(entry.Name())
			if nameUpper != "THUMBS.DB" && nameUpper != ".DS_STORE" && nameUpper != "DESKTOP.INI" {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			if os.RemoveAll(dirPath) == nil {
				removedCount++
			}
		}
	}

	return removedCount, nil
}

func (a *App) CleanupTags() (int, error) {
	return a.cleanupTagsSilent()
}

func (a *App) cleanupTagsSilent() (int, error) {
	imageTags, _ := a.loadImageTags()

	validPaths := make(map[string]bool)
	_ = a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		validPaths[relPath] = true
		return nil
	})

	removedCount := 0
	newImageTags := make(ImageTagsMap)
	for relPath, tags := range imageTags {
		if validPaths[relPath] {
			newImageTags[relPath] = tags
		} else {
			removedCount++
		}
	}

	if removedCount > 0 {
		a.saveImageTags(newImageTags)
	}

	return removedCount, nil
}

// --- Export & Folder Dialogs ---

func (a *App) ExportImages(paths []string, targetDir string, move bool) (int, error) {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return 0, err
		}
	}

	successCount := 0
	for _, relPath := range paths {
		sourcePath, err := a.resolveRootPath(relPath)
		if err != nil {
			continue
		}
		fileName := filepath.Base(sourcePath)
		destPath := filepath.Join(targetDir, fileName)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue
		}

		srcFile, err := os.Open(sourcePath)
		if err != nil {
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			continue
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			continue
		}

		if move {
			os.Remove(sourcePath)
		}
		successCount++
	}

	return successCount, nil
}

func (a *App) SelectFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Folder",
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, options)
	return dir, err
}

func (a *App) OpenImageLocation(relPath string) error {
	absPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid path")
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	query := fmt.Sprintf("/select,%s", absPath)
	cmd := exec.Command("explorer", query)
	return cmd.Start()
}

func (a *App) OpenFile(relPath string) error {
	absPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid path")
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	var cmd *exec.Cmd
	switch {
	case os.Getenv("OS") == "Windows_NT":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", absPath)
	default:
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", absPath)
	}

	return cmd.Start()
}

func (a *App) GetStatistics(period string) (*Stats, error) {
	stats := &Stats{
		ByDate: make(map[string]int),
		ByTag:  make(map[string]int),
	}

	imageTags, _ := a.loadImageTags()
	tagIDsToNames := make(map[string]string)
	tags, _ := a.loadTags()
	for _, t := range tags {
		tagIDsToNames[t.ID] = t.Name
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	err := a.walkManagedImages(func(path, relPath string, info fs.FileInfo) error {
		stats.TotalCount++
		stats.TotalSize += info.Size()

		modTime := info.ModTime()
		dateStr := modTime.Format("2006-01-02")
		if dateStr == today {
			stats.TodayCount++
		}

		dateKey := dateStr
		if period == "month" {
			dateKey = modTime.Format("2006-01")
		} else if period == "year" {
			dateKey = modTime.Format("2006")
		}
		stats.ByDate[dateKey]++

		if tids, ok := imageTags[relPath]; ok {
			for _, tid := range tids {
				if name, ok := tagIDsToNames[tid]; ok {
					stats.ByTag[name]++
				}
			}
		}
		return nil
	})

	return stats, err
}

func (a *App) OrganizeFiles(mode string) (int, error) {
	organizedCount := 0
	entries, err := os.ReadDir(a.imageDir)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		year := info.ModTime().Format("2006")
		month := info.ModTime().Format("01")

		targetRelDir := filepath.Join(year, month)
		if mode == "year" {
			targetRelDir = year
		}

		targetDir := filepath.Join(a.imageDir, targetRelDir)
		os.MkdirAll(targetDir, 0755)

		sourcePath := filepath.Join(a.imageDir, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		if _, err := os.Stat(destPath); err == nil {
			name := strings.TrimSuffix(entry.Name(), ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(targetDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		}

		if err := moveFile(sourcePath, destPath); err == nil {
			organizedCount++
		}
	}

	return organizedCount, nil
}

// --- Launcher Tools ---

func (a *App) loadLauncherTools() ([]LauncherTool, error) {
	var tools []LauncherTool
	data, err := os.ReadFile(a.launcherToolsFile())
	if err != nil {
		return []LauncherTool{}, nil
	}
	json.Unmarshal(data, &tools)
	return tools, nil
}

func (a *App) saveLauncherTools(tools []LauncherTool) error {
	data, _ := json.MarshalIndent(tools, "", "  ")
	return os.WriteFile(a.launcherToolsFile(), data, 0644)
}

func (a *App) GetLauncherTools() ([]LauncherTool, error) {
	return a.loadLauncherTools()
}

func (a *App) AddLauncherTool(tool LauncherTool) (LauncherTool, error) {
	tools, _ := a.loadLauncherTools()
	tool.ID = uuid.New().String()
	tools = append(tools, tool)
	err := a.saveLauncherTools(tools)
	return tool, err
}

func (a *App) UpdateLauncherTool(id string, tool LauncherTool) error {
	tools, _ := a.loadLauncherTools()
	updated := false
	for i, t := range tools {
		if t.ID == id {
			tools[i].Name = tool.Name
			tools[i].Path = tool.Path
			tools[i].Icon = tool.Icon
			tools[i].Args = tool.Args
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("tool not found")
	}
	return a.saveLauncherTools(tools)
}

// --- Custom Roots ---

func (a *App) loadCustomRoots() ([]CustomRoot, error) {
	var roots []CustomRoot
	data, err := os.ReadFile(a.customRootsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return []CustomRoot{}, nil
		}
		return nil, err
	}
	json.Unmarshal(data, &roots)
	if roots == nil {
		roots = []CustomRoot{}
	}
	return roots, nil
}

func (a *App) saveCustomRoots(roots []CustomRoot) error {
	data, _ := json.MarshalIndent(roots, "", "  ")
	return os.WriteFile(a.customRootsFile(), data, 0644)
}

func (a *App) GetCustomRoots() ([]CustomRoot, error) {
	return a.loadCustomRoots()
}

func (a *App) AddCustomRoot(name, relPath, icon string) (CustomRoot, error) {
	normalizedPath := normalizeRelPath(relPath)
	abs, err := a.resolveRootPath(normalizedPath)
	if err != nil {
		return CustomRoot{}, fmt.Errorf("路径不合法")
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return CustomRoot{}, fmt.Errorf("文件夹不存在: %s", relPath)
	}

	roots, _ := a.loadCustomRoots()
	// Check duplicate path
	for _, r := range roots {
		if normalizeRelPath(r.Path) == normalizedPath {
			return CustomRoot{}, fmt.Errorf("该文件夹已添加")
		}
	}

	if strings.TrimSpace(name) == "" {
		parts := strings.Split(normalizedPath, "/")
		name = parts[len(parts)-1]
	}

	if icon == "" {
		icon = "FolderSymlink"
	}

	newRoot := CustomRoot{
		ID:   uuid.New().String(),
		Name: name,
		Path: normalizedPath,
		Icon: icon,
	}
	roots = append(roots, newRoot)
	err = a.saveCustomRoots(roots)
	if err == nil {
		a.restartImageWatcher()
		a.scheduleImagesChangedEvent()
	}
	return newRoot, err
}

func (a *App) UpdateCustomRoot(id, name, icon string) error {
	roots, _ := a.loadCustomRoots()
	updated := false

	for i, root := range roots {
		if root.ID != id {
			continue
		}

		displayName := strings.TrimSpace(name)
		if displayName == "" {
			parts := strings.Split(filepath.ToSlash(root.Path), "/")
			displayName = parts[len(parts)-1]
		}
		if icon == "" {
			icon = "FolderSymlink"
		}

		roots[i].Name = displayName
		roots[i].Icon = icon
		updated = true
		break
	}

	if !updated {
		return fmt.Errorf("鑷畾涔夌洰褰曚笉瀛樺湪")
	}

	err := a.saveCustomRoots(roots)
	if err == nil {
		a.restartImageWatcher()
		a.scheduleImagesChangedEvent()
	}
	return err
}

func (a *App) DeleteCustomRoot(id string) error {
	roots, _ := a.loadCustomRoots()
	newRoots := []CustomRoot{}
	for _, r := range roots {
		if r.ID != id {
			newRoots = append(newRoots, r)
		}
	}
	err := a.saveCustomRoots(newRoots)
	if err == nil {
		a.restartImageWatcher()
		a.scheduleImagesChangedEvent()
	}
	return err
}

// GetRelativePath converts an absolute path to a path relative to rootDir.
// Returns error if the path is not inside rootDir.
func (a *App) GetRelativePath(absPath string) (string, error) {
	abs, err := normalizeExistingPath(absPath)
	if err != nil {
		return "", err
	}
	if !isSubPath(a.rootDir, abs) {
		return "", fmt.Errorf("路径不在图片目录内")
	}
	rel, err := filepath.Rel(a.rootDir, abs)
	if err != nil {
		return "", err
	}
	return normalizeRelPath(rel), nil
}

// GetSubFolders returns immediate sub-folders of a given relative path (or rootDir if empty).
// Used by the UI folder picker to let users choose a custom root directory.
func (a *App) GetSubFolders(relPath string) ([]string, error) {
	var base string
	if relPath == "" {
		base = a.rootDir
	} else {
		resolved, err := a.resolveRootPath(relPath)
		if err != nil {
			return nil, fmt.Errorf("路径不合法")
		}
		base = resolved
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}

	var folders []string
	skipNames := map[string]bool{
		"node_modules": true, ".git": true, ".trash": true,
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if skipNames[name] || name == "desktop-source" {
			continue
		}
		var folderRel string
		if relPath == "" {
			folderRel = name
		} else {
			folderRel = filepath.ToSlash(filepath.Join(normalizeRelPath(relPath), name))
		}
		folders = append(folders, folderRel)
	}
	return folders, nil
}

func (a *App) DeleteLauncherTool(id string) error {
	tools, _ := a.loadLauncherTools()
	newTools := []LauncherTool{}
	for _, t := range tools {
		if t.ID != id {
			newTools = append(newTools, t)
		}
	}
	return a.saveLauncherTools(newTools)
}

func (a *App) RunLauncherTool(id string) error {
	tools, _ := a.loadLauncherTools()
	var tool *LauncherTool
	for _, t := range tools {
		if t.ID == id {
			tool = &t
			break
		}
	}

	if tool == nil {
		return fmt.Errorf("tool not found")
	}

	targetPath := strings.TrimSpace(tool.Path)
	if targetPath == "" {
		return fmt.Errorf("tool path is empty")
	}
	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tool file not found")
		}
		return err
	}

	args := []string{}
	if strings.TrimSpace(tool.Args) != "" {
		args = strings.Fields(tool.Args)
	}

	ext := strings.ToLower(filepath.Ext(targetPath))
	var cmd *exec.Cmd
	switch ext {
	case ".bat", ".cmd", ".html", ".htm", ".url":
		cmdArgs := append([]string{"/c", "start", "", targetPath}, args...)
		cmd = exec.Command("cmd.exe", cmdArgs...)
	default:
		cmd = exec.Command(targetPath, args...)
	}

	cmd.Dir = filepath.Dir(targetPath)
	return cmd.Start()
}

func (a *App) ExtractIcon(path string) (string, error) {
	if _, err := os.Stat(a.iconsDir()); os.IsNotExist(err) {
		os.MkdirAll(a.iconsDir(), 0755)
	}

	hash := md5.Sum([]byte(path))
	iconFilename := hex.EncodeToString(hash[:]) + ".png"
	iconPath := filepath.Join(a.iconsDir(), iconFilename)

	// Since we use the Wails asset server for the frontend, we could use an absolute path URI or simple base64
	// Let's just return the absolute path to be loaded via frontend Custom URL or base64 the file
	// Wails standard is usually exposing via AssetServer or passing base64. Because the icon is tiny,
	// we'll extract it and read it as base64 string to keep it simple!

	generateBase64 := func(ip string) (string, error) {
		bytes, err := os.ReadFile(ip)
		if err != nil {
			return "", err
		}
		// Data URI scheme
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes), nil
	}

	if _, err := os.Stat(iconPath); err == nil {
		return generateBase64(iconPath)
	}

	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Drawing
		$icon = [System.Drawing.Icon]::ExtractAssociatedIcon('%s')
		if ($icon) {
			$bitmap = $icon.ToBitmap()
			$bitmap.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png)
			$bitmap.Dispose()
			$icon.Dispose()
		}
	`, path, iconPath)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", "& {"+psScript+"}")
	if _, err := cmd.CombinedOutput(); err != nil {
		return "", err
	}

	return generateBase64(iconPath)
}
