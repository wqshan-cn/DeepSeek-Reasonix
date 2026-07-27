package main

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const desktopPetArg = "--desktop-pet"

type DesktopPetSession struct {
	TabID   string `json:"tabId"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Phase   string `json:"phase"`
	Running bool   `json:"running"`
	Waiting bool   `json:"waiting"`
}

type DesktopPetState struct {
	Title          string              `json:"title"`
	Status         string              `json:"status"`
	Phase          string              `json:"phase"`
	TabID          string              `json:"tabId"`
	Running        bool                `json:"running"`
	Waiting        bool                `json:"waiting"`
	ActiveCount    int                 `json:"activeCount"`
	AttentionCount int                 `json:"attentionCount"`
	UpdatedAt      int64               `json:"updatedAt"`
	PetID          string              `json:"petId"`
	Sessions       []DesktopPetSession `json:"sessions"`
}

type DesktopPetPack struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	BuiltIn     bool   `json:"builtIn"`
	AssetURL    string `json:"assetUrl,omitempty"`
}

type desktopPetManifest struct {
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	Description     string `json:"description"`
	SpritesheetPath string `json:"spritesheetPath"`
}

type desktopPetPreferences struct {
	PetID   string `json:"petId"`
	Enabled bool   `json:"enabled"`
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

func desktopPetConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "Reasonix")
}

func desktopPetStatePath() string {
	return filepath.Join(desktopPetConfigDir(), "desktop-pet-state.json")
}
func desktopPetRequestPath() string {
	return filepath.Join(desktopPetConfigDir(), "desktop-pet-open-request.json")
}
func desktopPetCommandPath() string {
	return filepath.Join(desktopPetConfigDir(), "desktop-pet-command.json")
}
func desktopPetPreferencesPath() string {
	return filepath.Join(desktopPetConfigDir(), "desktop-pet-preferences.json")
}
func desktopPetPacksDir() string { return filepath.Join(desktopPetConfigDir(), "pets") }

func desktopPetWebviewDataPath() string {
	base, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "Reasonix", "desktop-pet-webview2")
}

func writeDesktopPetJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(value)
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

func (a *App) StartDesktopPet() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(executable, desktopPetArg).Start()
}

func (a *App) SendDesktopPetCommand(command string) error {
	if err := writeDesktopPetJSON(desktopPetCommandPath(), map[string]any{
		"command":   strings.TrimSpace(command),
		"updatedAt": time.Now().UnixMilli(),
	}); err != nil {
		return err
	}
	return a.StartDesktopPet()
}

func (a *App) UpdateDesktopPetState(state DesktopPetState) error {
	if state.UpdatedAt == 0 {
		state.UpdatedAt = time.Now().UnixMilli()
	}
	state.PetID = readDesktopPetPreferences().PetID
	return writeDesktopPetJSON(desktopPetStatePath(), state)
}

func (a *App) ListDesktopPetPacks() []DesktopPetPack { return listDesktopPetPacks() }

func (a *App) DesktopPetPreference() string {
	return readDesktopPetPreferences().PetID
}

func (a *App) DesktopPetEnabled() bool {
	return readDesktopPetPreferences().Enabled
}

func (a *App) SetDesktopPetEnabled(enabled bool) error {
	prefs := readDesktopPetPreferences()
	prefs.Enabled = enabled
	if err := writeDesktopPetJSON(desktopPetPreferencesPath(), prefs); err != nil {
		return err
	}
	if enabled {
		return a.StartDesktopPet()
	}
	return a.SendDesktopPetCommand("close")
}

func (a *App) SetDesktopPetPreference(id string) error {
	id = strings.TrimSpace(id)
	found := false
	for _, pack := range listDesktopPetPacks() {
		if pack.ID == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unknown desktop pet %q", id)
	}
	prefs := readDesktopPetPreferences()
	prefs.PetID = id
	return writeDesktopPetJSON(desktopPetPreferencesPath(), prefs)
}

func (a *App) ConsumeDesktopPetOpenRequest() string {
	data, err := os.ReadFile(desktopPetRequestPath())
	if err != nil {
		return ""
	}
	_ = os.Remove(desktopPetRequestPath())
	var request struct {
		TabID string `json:"tabId"`
	}
	if json.Unmarshal(data, &request) != nil {
		return ""
	}
	return strings.TrimSpace(request.TabID)
}

func readDesktopPetPreferences() desktopPetPreferences {
	prefs := desktopPetPreferences{PetID: "akita", Enabled: true}
	data, err := os.ReadFile(desktopPetPreferencesPath())
	if err == nil {
		_ = json.Unmarshal(data, &prefs)
	}
	if strings.TrimSpace(prefs.PetID) == "" {
		prefs.PetID = "akita"
	}
	return prefs
}

func listDesktopPetPacks() []DesktopPetPack {
	packs := []DesktopPetPack{{
		ID:          "akita",
		DisplayName: "Akita",
		Description: "会随任务状态行动的像素秋田犬",
		Kind:        "gif-set",
		BuiltIn:     true,
	}}
	entries, err := os.ReadDir(desktopPetPacksDir())
	if err != nil {
		return packs
	}
	for _, entry := range entries {
		if !entry.IsDir() || !safeDesktopPetID(entry.Name()) {
			continue
		}
		dir := filepath.Join(desktopPetPacksDir(), entry.Name())
		data, err := os.ReadFile(filepath.Join(dir, "pet.json"))
		if err != nil || len(data) > 128*1024 {
			continue
		}
		var manifest desktopPetManifest
		if json.Unmarshal(data, &manifest) != nil ||
			manifest.ID != entry.Name() ||
			manifest.SpritesheetPath != "spritesheet.webp" ||
			strings.TrimSpace(manifest.DisplayName) == "" {
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, "spritesheet.webp")); err != nil || info.IsDir() {
			continue
		}
		packs = append(packs, DesktopPetPack{
			ID:          manifest.ID,
			DisplayName: manifest.DisplayName,
			Description: manifest.Description,
			Kind:        "codex-spritesheet",
			AssetURL:    "/pet-local/" + manifest.ID + "/spritesheet.webp",
		})
	}
	return packs
}

func safeDesktopPetID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, char := range id {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func (p *DesktopPetApp) startup(ctx context.Context) { p.ctx = ctx }

func (p *DesktopPetApp) domReady(context.Context) {
	if p.ctx == nil {
		return
	}
	if screens, err := wailsruntime.ScreenGetAll(p.ctx); err == nil && len(screens) > 0 {
		screen := screens[0]
		wailsruntime.WindowSetPosition(
			p.ctx,
			desktopPetMax(16, screen.Size.Width-560-24),
			desktopPetMax(16, screen.Size.Height-280-56),
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
		return DesktopPetState{Title: "Reasonix", Status: "等待任务", Phase: "idle", PetID: readDesktopPetPreferences().PetID}
	}
	var state DesktopPetState
	if json.Unmarshal(data, &state) != nil {
		return DesktopPetState{Title: "Reasonix", Status: "等待任务", Phase: "idle", PetID: readDesktopPetPreferences().PetID}
	}
	state.PetID = readDesktopPetPreferences().PetID
	return state
}

func (p *DesktopPetApp) ListDesktopPetPacks() []DesktopPetPack { return listDesktopPetPacks() }

func (p *DesktopPetApp) ReadCommand() string {
	data, err := os.ReadFile(desktopPetCommandPath())
	if err != nil {
		return ""
	}
	_ = os.Remove(desktopPetCommandPath())
	var command struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(data, &command) != nil {
		return ""
	}
	return strings.TrimSpace(command.Command)
}

func (p *DesktopPetApp) OpenMainWindow(tabID string) error {
	if strings.TrimSpace(tabID) != "" {
		if err := writeDesktopPetJSON(desktopPetRequestPath(), map[string]string{"tabId": tabID}); err != nil {
			return err
		}
	}
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
		const prefix = "/pet-local/"
		if strings.HasPrefix(r.URL.Path, prefix) {
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
			if len(parts) != 2 || !safeDesktopPetID(parts[0]) || parts[1] != "spritesheet.webp" {
				http.NotFound(w, r)
				return
			}
			path := filepath.Join(desktopPetPacksDir(), parts[0], "spritesheet.webp")
			w.Header().Set("Content-Type", mime.TypeByExtension(".webp"))
			http.ServeFile(w, r, path)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func runDesktopPet() {
	pet := &DesktopPetApp{}
	err := wails.Run(&options.App{
		Title:            "Reasonix Desktop Pet",
		Width:            560,
		Height:           280,
		MinWidth:         560,
		MinHeight:        280,
		MaxWidth:         560,
		MaxHeight:        280,
		Frameless:        true,
		DisableResize:    true,
		AlwaysOnTop:      true,
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
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				if pet.ctx != nil {
					wailsruntime.WindowShow(pet.ctx)
				}
			},
		},
		Windows: &windows.Options{
			Theme:                             windows.SystemDefault,
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: true,
			WebviewUserDataPath:               desktopPetWebviewDataPath(),
			WindowClassName:                   "ReasonixDesktopPet",
		},
	})
	if err != nil {
		println("Desktop pet error:", err.Error())
	}
}
