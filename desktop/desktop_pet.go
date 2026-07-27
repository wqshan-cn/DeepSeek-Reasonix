package main

import "github.com/wailsapp/wails/v2/pkg/runtime"

const (
	desktopPetWidth  = 360
	desktopPetHeight = 84
)

type desktopPetWindowState struct {
	x, y          int
	width, height int
	maximised     bool
}

// EnterDesktopPet compacts the main webview into a small always-on-top task
// surface. Keeping the existing webview alive means streaming state continues
// to update while the pet is visible.
func (a *App) EnterDesktopPet() {
	if a.ctx == nil {
		return
	}
	a.desktopPetMu.Lock()
	defer a.desktopPetMu.Unlock()
	if a.desktopPetActive {
		return
	}
	x, y := runtime.WindowGetPosition(a.ctx)
	width, height := runtime.WindowGetSize(a.ctx)
	a.desktopPetWindow = desktopPetWindowState{
		x: x, y: y, width: width, height: height,
		maximised: runtime.WindowIsMaximised(a.ctx),
	}
	a.desktopPetActive = true

	runtime.WindowUnmaximise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetSize(a.ctx, desktopPetWidth, desktopPetHeight)
	if screens, err := runtime.ScreenGetAll(a.ctx); err == nil && len(screens) > 0 {
		screen := screens[0]
		runtime.WindowSetPosition(
			a.ctx,
			maxInt(16, screen.Size.Width-desktopPetWidth-24),
			maxInt(16, screen.Size.Height-desktopPetHeight-72),
		)
	}
	runtime.WindowShow(a.ctx)
}

// ExitDesktopPet restores the conversation window to the geometry it had
// before entering pet mode.
func (a *App) ExitDesktopPet() {
	if a.ctx == nil {
		return
	}
	a.desktopPetMu.Lock()
	defer a.desktopPetMu.Unlock()
	if !a.desktopPetActive {
		return
	}
	state := a.desktopPetWindow
	a.desktopPetActive = false
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	if state.width > 0 && state.height > 0 {
		runtime.WindowSetSize(a.ctx, state.width, state.height)
	}
	runtime.WindowSetPosition(a.ctx, state.x, state.y)
	if state.maximised {
		runtime.WindowMaximise(a.ctx)
	}
	runtime.WindowShow(a.ctx)
}

func (a *App) DesktopPetActive() bool {
	a.desktopPetMu.Lock()
	defer a.desktopPetMu.Unlock()
	return a.desktopPetActive
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
