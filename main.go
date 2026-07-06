package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--scan" {
		scanMods()
		return
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TownStory 整合包模组管理",
		Width:  610,
		Height: 520,
		MinWidth:  490,
		MinHeight: 420,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 245, B: 245, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// scanMods prints the SHA256 hash of every jar/jar.disabled file in mods/,
// so a new hash can be copied into the features list in app.go.
func scanMods() {
	attachConsole()

	index := buildHashIndex()

	names := make([]string, 0, len(index))
	byName := make(map[string]string, len(index))

	for hash, path := range index {
		name := filepath.Base(path)
		names = append(names, name)
		byName[name] = hash
	}

	sort.Strings(names)

	fmt.Println("\n===== MOD HASH LIST =====\n")

	for _, name := range names {
		fmt.Printf("FILE: %s\n", name)
		fmt.Printf("SHA256: %s\n\n", byName[name])
	}
}

// attachConsole binds the process to the parent console, switches its
// output code page to UTF-8, and redirects stdout/stderr to it. Wails
// builds as a GUI subsystem binary, which has no console and no inherited
// stdout by default, so fmt.Println would otherwise go nowhere when run
// from a terminal. Without the UTF-8 code page, cmd.exe's default GBK
// codepage (936) garbles the Chinese mod names.
func attachConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	kernel32.NewProc("AttachConsole").Call(^uintptr(0)) // ATTACH_PARENT_PROCESS
	kernel32.NewProc("SetConsoleOutputCP").Call(65001)  // CP_UTF8

	conout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return
	}

	os.Stdout = conout
	os.Stderr = conout
}
