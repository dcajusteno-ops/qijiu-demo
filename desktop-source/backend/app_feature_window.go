package backend

import (
	"context"
	"math"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) applySavedWindowSettings(ctx context.Context) {
	settings, err := a.loadSettings()
	if err != nil {
		return
	}
	runtime.WindowSetAlwaysOnTop(ctx, settings.AlwaysOnTop)
}

func (a *App) GetWindowBehaviorSettings() (WindowBehaviorSettings, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return WindowBehaviorSettings{}, err
	}
	return WindowBehaviorSettings{AlwaysOnTop: settings.AlwaysOnTop}, nil
}

func (a *App) ToggleAlwaysOnTop(enabled bool) (WindowBehaviorSettings, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return WindowBehaviorSettings{}, err
	}
	settings.AlwaysOnTop = enabled
	if err := a.saveSettings(settings); err != nil {
		return WindowBehaviorSettings{}, err
	}
	if a.ctx != nil {
		runtime.WindowSetAlwaysOnTop(a.ctx, enabled)
	}
	return WindowBehaviorSettings{AlwaysOnTop: enabled}, nil
}

func (a *App) ShowCompactWindow() {
	if a.ctx == nil {
		return
	}
	width, height := a.compactWindowSize(a.ctx)
	runtime.WindowUnfullscreen(a.ctx)
	runtime.WindowUnmaximise(a.ctx)
	runtime.WindowSetMinSize(a.ctx, compactWindowMinWidth, compactWindowMinHeight)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetSize(a.ctx, width, height)
	a.placeWindowBottomRight(a.ctx, width, height)
	a.scheduleCompactWindowSnap(a.ctx)
}

func (a *App) RestoreMainWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowSetMinSize(a.ctx, mainWindowMinWidth, mainWindowMinHeight)
	runtime.WindowSetSize(a.ctx, mainWindowWidth, mainWindowHeight)
	settings, err := a.loadSettings()
	if err == nil {
		runtime.WindowSetAlwaysOnTop(a.ctx, settings.AlwaysOnTop)
	}
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowCenter(a.ctx)
}

func (a *App) placeWindowBottomRight(ctx context.Context, width, height int) {
	if width <= 0 || height <= 0 {
		width, height = a.compactWindowSize(ctx)
	}

	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}

	screen := screens[0]
	for _, item := range screens {
		if item.IsCurrent || item.IsPrimary {
			screen = item
			break
		}
	}

	screenWidth := screen.Size.Width
	screenHeight := screen.Size.Height
	windowWidth := width
	windowHeight := height
	marginX := 22
	marginY := 72
	if screen.PhysicalSize.Width > 0 && screen.PhysicalSize.Height > 0 && screen.Size.Width > 0 && screen.Size.Height > 0 {
		scaleX := float64(screen.PhysicalSize.Width) / float64(screen.Size.Width)
		scaleY := float64(screen.PhysicalSize.Height) / float64(screen.Size.Height)
		screenWidth = screen.PhysicalSize.Width
		screenHeight = screen.PhysicalSize.Height
		windowWidth = int(math.Round(float64(width) * scaleX))
		windowHeight = int(math.Round(float64(height) * scaleY))
		marginX = int(math.Round(float64(marginX) * scaleX))
		marginY = int(math.Round(float64(marginY) * scaleY))
	}

	x := screenWidth - windowWidth - marginX
	y := screenHeight - windowHeight - marginY
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	runtime.WindowSetPosition(ctx, x, y)
}

func (a *App) scheduleCompactWindowSnap(ctx context.Context) {
	go func() {
		time.Sleep(80 * time.Millisecond)
		a.snapCompactWindowBottomRight(ctx)
		time.Sleep(180 * time.Millisecond)
		a.snapCompactWindowBottomRight(ctx)
	}()
}

func (a *App) snapCompactWindowBottomRight(ctx context.Context) {
	width, height := runtime.WindowGetSize(ctx)
	a.placeWindowBottomRight(ctx, width, height)
}

func (a *App) compactWindowSize(ctx context.Context) (int, int) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return 390, 320
	}

	screen := screens[0]
	for _, item := range screens {
		if item.IsCurrent || item.IsPrimary {
			screen = item
			break
		}
	}

	width := screen.Size.Width * 3 / 10
	height := screen.Size.Height / 2
	if width < 440 {
		width = 440
	}
	if width > 560 {
		width = 560
	}
	if height < 520 {
		height = 520
	}
	if height > 680 {
		height = 680
	}
	return width, height
}
