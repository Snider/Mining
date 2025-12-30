package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist/browser
var assets embed.FS

func main() {
	// Create the mining service
	miningService := NewMiningService()

	// Get the sub-filesystem rooted at frontend/dist/browser
	browserFS, err := fs.Sub(assets, "frontend/dist/browser")
	if err != nil {
		log.Fatal("Failed to create sub-filesystem:", err)
	}

	// Create a new Wails application
	app := application.New(application.Options{
		Name:        "Mining Dashboard",
		Description: "Multi-miner management dashboard",
		Services: []application.Service{
			application.NewService(miningService),
		},
		Assets: application.AssetOptions{
			Handler: http.FileServer(http.FS(browserFS)),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create the main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Mining Dashboard",
		Width:  1400,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInsetUnified,
		},
		BackgroundColour: application.NewRGB(10, 10, 18),
		URL:              "/",
	})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		miningService.Shutdown()
		os.Exit(0)
	}()

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
