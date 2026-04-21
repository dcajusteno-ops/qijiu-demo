package backend

import (
	"context"
	_ "image/gif"
	_ "image/jpeg"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "golang.org/x/image/webp"
)

type App struct {
	ctx                    context.Context
	rootDir                string
	imageDir               string
	dataDir                string
	appDir                 string
	shortcutManager        shortcutManager
	imageMetaMu            sync.RWMutex
	imageMetaCache         ImageMetaCache
	imageMetaLoaded        bool
	imageMetaWarmupRunning bool
	promptLibraryMu        sync.RWMutex
	promptLibraryCache     []PromptLibraryEntry
	promptLibraryLoaded    bool
	autoRulesMu            sync.Mutex
	autoRulesRunMu         sync.Mutex
	watchMu                sync.Mutex
	imageWatcher           *fsnotify.Watcher
	imageWatchStop         chan struct{}
	imageWatchDebounce     *time.Timer
}
