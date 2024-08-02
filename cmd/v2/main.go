package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/andreimerlescu/configurable"
	"github.com/andreimerlescu/install-go/cmd/v2/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, bootErr := types.NewApplication(configurable.New())
	if bootErr != nil {
		log.Fatal(bootErr)
	}
	home := filepath.Join(".", ".go")
	log.Printf("FYI: Fallback GODIR will be %v if your Home Directory cannot be determined.", home)

	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		log.Printf("WARNING: Cannot use User Home Directory. Error: %v", homeErr)
	}
	app.HOME = home
	app.GODIR = filepath.Join(home, "go")
	app.Config = &types.Config{
		Version:   app.CLI().NewString("version", "", "Specify Stable Release or use --latest (or --latest-rc)"),
		Latest:    app.CLI().NewBool("latest", true, "Use latest Stable Release"),
		LatestRC:  app.CLI().NewBool("rc", false, "Use latest Release Candidate"),
		Install:   app.CLI().NewBool("install", false, "Install --version"),
		Uninstall: app.CLI().NewBool("uninstall", false, "Uninstall --version"),
		GODIR:     app.CLI().NewString("godir", app.GODIR, "Path to GODIR (target directory for all actions - must be writable)"),
		Switch:    app.CLI().NewBool("switch", false, "Switch to --version"),
		Backup:    app.CLI().NewBool("backup", false, "Create a backup of GODIR inside --output"),
		Output:    app.CLI().NewString("output", app.GODIR, "Path to store --backup if enabled"),
		GOOS:      app.CLI().NewString("os", "linux", "Override GOOS environment value (one time instant temporary)"),
		GOARCH:    app.CLI().NewString("arch", "amd64", "Override GOARCH environment variable (one time instant temporary)"),
		Debug:     app.CLI().NewBool("debug", false, "Enable debug logging to --log"),
		LogFile:   app.CLI().NewString("log", "STDOUT", "STDOUT or Path to writable log file"),
		Downloads: app.CLI().NewString("downloads", filepath.Join(app.GODIR, "downloads"), "Path to downloads"),
	}

	loadErr := app.Load()
	if loadErr != nil {
		log.Fatal(loadErr)
	}

	wg := &sync.WaitGroup{}
	responder := types.NewResponder(ctx, app, wg)
	app.Responder = responder
	go responder.ShowHelp()
	go responder.AddressUsability()
	go responder.SetLogging()
	go responder.TakeBackup()
	go responder.PrepareWorkspace()
	wg.Wait()

}
