//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	goruntime "runtime"
)

const (
	wmHotkey    = 0x0312
	pmRemove    = 0x0001
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000
)

var (
	user32DLL            = syscall.NewLazyDLL("user32.dll")
	registerHotKeyProc   = user32DLL.NewProc("RegisterHotKey")
	unregisterHotKeyProc = user32DLL.NewProc("UnregisterHotKey")
	peekMessageProc      = user32DLL.NewProc("PeekMessageW")
)

type hotkeyPoint struct {
	X int32
	Y int32
}

type hotkeyMessage struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       hotkeyPoint
	LPrivate uint32
}

type hotkeyBindingDef struct {
	ID          int32
	Action      string
	Accelerator string
	Modifiers   uint32
	Key         uint32
}

type hotkeyApplyCommand struct {
	app      *App
	settings ShortcutSettings
	resp     chan error
}

type windowsShortcutManager struct {
	startOnce sync.Once
	cmdCh     chan hotkeyApplyCommand
	stopCh    chan chan error
}

func newShortcutManager() shortcutManager {
	manager := &windowsShortcutManager{
		cmdCh:  make(chan hotkeyApplyCommand),
		stopCh: make(chan chan error),
	}
	manager.start()
	return manager
}

func (m *windowsShortcutManager) start() {
	m.startOnce.Do(func() {
		go m.loop()
	})
}

func (m *windowsShortcutManager) Apply(app *App, settings ShortcutSettings) error {
	resp := make(chan error, 1)
	m.cmdCh <- hotkeyApplyCommand{
		app:      app,
		settings: settings,
		resp:     resp,
	}
	return <-resp
}

func (m *windowsShortcutManager) Close() error {
	resp := make(chan error, 1)
	m.stopCh <- resp
	return <-resp
}

func (m *windowsShortcutManager) loop() {
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	var currentApp *App
	currentBindings := make([]hotkeyBindingDef, 0)

	for {
		select {
		case cmd := <-m.cmdCh:
			currentApp = cmd.app
			nextBindings, err := buildHotkeyBindings(cmd.settings)
			if err != nil {
				cmd.resp <- err
				continue
			}

			previousBindings := append([]hotkeyBindingDef(nil), currentBindings...)
			unregisterHotkeyBindings(previousBindings)

			if len(nextBindings) == 0 {
				currentBindings = currentBindings[:0]
				cmd.resp <- nil
				continue
			}

			if err := registerHotkeyBindings(nextBindings); err != nil {
				unregisterHotkeyBindings(nextBindings)
				if restoreErr := registerHotkeyBindings(previousBindings); restoreErr != nil {
					currentBindings = currentBindings[:0]
					cmd.resp <- fmt.Errorf("%v；回滚旧快捷键失败：%v", err, restoreErr)
					continue
				}
				currentBindings = previousBindings
				cmd.resp <- err
				continue
			}

			currentBindings = nextBindings
			cmd.resp <- nil

		case resp := <-m.stopCh:
			unregisterHotkeyBindings(currentBindings)
			resp <- nil
			return

		default:
			var msg hotkeyMessage
			hasMessage, _, _ := peekMessageProc.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
				pmRemove,
			)
			if hasMessage != 0 && msg.Message == wmHotkey && currentApp != nil && currentApp.ctx != nil {
				actionID := findHotkeyAction(currentBindings, int32(msg.WParam))
				if actionID != "" {
					runtime.EventsEmit(currentApp.ctx, "shortcut:triggered", actionID)
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func findHotkeyAction(bindings []hotkeyBindingDef, id int32) string {
	for _, binding := range bindings {
		if binding.ID == id {
			return binding.Action
		}
	}
	return ""
}

func buildHotkeyBindings(settings ShortcutSettings) ([]hotkeyBindingDef, error) {
	if !settings.Enabled {
		return nil, nil
	}

	bindings := make([]hotkeyBindingDef, 0, len(settings.Bindings))
	for _, binding := range settings.Bindings {
		accelerator := normalizeAccelerator(binding.Accelerator)
		if accelerator == "" {
			continue
		}

		modifiers, key, err := parseWindowsAccelerator(accelerator)
		if err != nil {
			action := shortcutCatalogByID[binding.Action]
			return nil, fmt.Errorf("%s 的快捷键无效: %w", action.Label, err)
		}

		bindings = append(bindings, hotkeyBindingDef{
			ID:          int32(len(bindings) + 1),
			Action:      binding.Action,
			Accelerator: accelerator,
			Modifiers:   modifiers | modNoRepeat,
			Key:         key,
		})
	}

	return bindings, nil
}

func registerHotkeyBindings(bindings []hotkeyBindingDef) error {
	for i, binding := range bindings {
		if err := registerHotkey(binding); err != nil {
			unregisterHotkeyBindings(bindings[:i])
			return err
		}
	}
	return nil
}

func unregisterHotkeyBindings(bindings []hotkeyBindingDef) {
	for _, binding := range bindings {
		_, _, _ = unregisterHotKeyProc.Call(0, uintptr(binding.ID))
	}
}

func registerHotkey(binding hotkeyBindingDef) error {
	result, _, callErr := registerHotKeyProc.Call(
		0,
		uintptr(binding.ID),
		uintptr(binding.Modifiers),
		uintptr(binding.Key),
	)
	if result != 0 {
		return nil
	}
	if callErr != syscall.Errno(0) {
		return fmt.Errorf("无法注册 %s: %v", binding.Accelerator, callErr)
	}
	return fmt.Errorf("无法注册 %s: 该快捷键可能已被系统或其他程序占用", binding.Accelerator)
}

func parseWindowsAccelerator(accelerator string) (uint32, uint32, error) {
	parts := strings.Split(normalizeAccelerator(accelerator), "+")
	if len(parts) == 0 {
		return 0, 0, fmt.Errorf("快捷键不能为空")
	}

	var modifiers uint32
	keyPart := ""
	for _, part := range parts {
		switch part {
		case "Ctrl":
			modifiers |= modControl
		case "Alt":
			modifiers |= modAlt
		case "Shift":
			modifiers |= modShift
		case "Win":
			modifiers |= modWin
		default:
			if keyPart != "" {
				return 0, 0, fmt.Errorf("只能包含一个主按键")
			}
			keyPart = part
		}
	}

	if modifiers == 0 {
		return 0, 0, fmt.Errorf("至少需要一个修饰键")
	}
	if keyPart == "" {
		return 0, 0, fmt.Errorf("缺少主按键")
	}

	key, ok := windowsVirtualKeyMap()[keyPart]
	if !ok {
		return 0, 0, fmt.Errorf("暂不支持按键 %s", keyPart)
	}

	return modifiers, key, nil
}

func windowsVirtualKeyMap() map[string]uint32 {
	result := map[string]uint32{
		"SPACE":     0x20,
		"ENTER":     0x0D,
		"TAB":       0x09,
		"ESCAPE":    0x1B,
		"BACKSPACE": 0x08,
		"DELETE":    0x2E,
		"INSERT":    0x2D,
		"HOME":      0x24,
		"END":       0x23,
		"PAGEUP":    0x21,
		"PAGEDOWN":  0x22,
		"UP":        0x26,
		"DOWN":      0x28,
		"LEFT":      0x25,
		"RIGHT":     0x27,
	}

	for code := 'A'; code <= 'Z'; code++ {
		key := string(code)
		result[key] = uint32(code)
	}
	for code := '0'; code <= '9'; code++ {
		key := string(code)
		result[key] = uint32(code)
	}
	for i := 1; i <= 12; i++ {
		key := fmt.Sprintf("F%d", i)
		result[key] = uint32(0x70 + i - 1)
	}

	return result
}
