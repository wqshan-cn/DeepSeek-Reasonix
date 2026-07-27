package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const desktopPetArg = "--desktop-pet"

type DesktopPetState struct {
	Title   string `json:"title"`
	Status  string `json:"status"`
	Running bool   `json:"running"`
	Waiting bool   `json:"waiting"`
}

type DesktopPetApp struct {
	ctx context.Context
}

func hasDesktopPetArg(args []string) bool {
	for _, arg := range args {
		if arg == desktopPetArg {
			return true
		}
	}
	return false
}

func desktopPetStatePath() string {
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "Reasonix", "desktop-pet-state.json")
}

func desktopPetWebviewDataPath() string {
	base, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "Reasonix", "desktop-pet-webview2")
}

func (a *App) StartDesktopPet() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(executable, desktopPetArg).Start()
}

func (a *App) UpdateDesktopPetState(state DesktopPetState) error {
	path := desktopPetStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}

func (p *DesktopPetApp) startup(ctx context.Context) {
	p.ctx = ctx
}

func (p *DesktopPetApp) domReady(context.Context) {
	if p.ctx == nil {
		return
	}
	if screens, err := wailsruntime.ScreenGetAll(p.ctx); err == nil && len(screens) > 0 {
		screen := screens[0]
		wailsruntime.WindowSetPosition(
			p.ctx,
			desktopPetMax(16, screen.Size.Width-520-24),
			desktopPetMax(16, screen.Size.Height-250-56),
		)
	}
	wailsruntime.WindowShow(p.ctx)
}

func desktopPetMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p *DesktopPetApp) ReadState() DesktopPetState {
	data, err := os.ReadFile(desktopPetStatePath())
	if err != nil {
		return DesktopPetState{Title: "Reasonix", Status: "等待任务"}
	}
	var state DesktopPetState
	if json.Unmarshal(data, &state) != nil {
		return DesktopPetState{Title: "Reasonix", Status: "等待任务"}
	}
	return state
}

func (p *DesktopPetApp) OpenMainWindow() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(executable).Start()
}

func (p *DesktopPetApp) CloseDesktopPet() {
	if p.ctx != nil {
		wailsruntime.Quit(p.ctx)
	}
}

func desktopPetAssetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			r.URL.Path = "/pet.html"
		}
		next.ServeHTTP(w, r)
	})
}

func runDesktopPet() {
	pet := &DesktopPetApp{}
	err := wails.Run(&options.App{
		Title:       "Reasonix Desktop Pet",
		Width:       520,
		Height:      250,
		MinWidth:    520,
		MinHeight:   250,
		MaxWidth:    520,
		MaxHeight:   250,
		Frameless:   true,
		DisableResize: true,
		AlwaysOnTop: true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: desktopPetAssetMiddleware,
		},
		OnStartup:  pet.startup,
		OnDomReady: pet.domReady,
		Bind:       []any{pet},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "7b64737c-a127-4d69-adbc-9f99a2fa1e80",
		},
		Windows: &windows.Options{
			Theme:                              windows.SystemDefault,
			WebviewIsTransparent:               true,
			WindowIsTranslucent:                true,
			DisableFramelessWindowDecorations: true,
			WebviewUserDataPath:               desktopPetWebviewDataPath(),
			WindowClassName:                    "ReasonixDesktopPet",
		},
	})
	if err != nil {
		println("Desktop pet error:", err.Error())
	}
}
