package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

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
	}

	configFile := ""
	tempConfig := filepath.Join(".", "config.yaml")
	if _, configFileErr := os.Stat(tempConfig); configFileErr != nil {
		if errors.Is(configFileErr, fs.ErrNotExist) {
			configFile = ""
		} else if errors.Is(configFileErr, fs.ErrPermission) {
			log.Printf("WARNING: Permission denied attempting to read %v with error: %v", tempConfig, configFileErr)
			configFile = ""
		} else if errors.Is(configFileErr, fs.ErrInvalid) {
			log.Printf("DANGER: %v", configFileErr)
			configFile = ""
		}
	} else {
		configFile = fmt.Sprintf("%v", tempConfig)
	}

	loadErr := app.Load(configFile)
	if loadErr != nil {
		log.Fatal(loadErr)
	}

	// Respond to help first
	if *app.Config.Help {
		fmt.Println(app.CLI().Usage())
		return
	}

	// If user sets --version to latest, its a user-error, catch it and correct it
	if *app.Config.Version == "latest" {
		app.Config.Version = nil
		*app.Config.Latest = true
		*app.Config.LatestRC = false
	}

	if *app.Config.Version == "latest-rc" {
		app.Config.Version = nil
		*app.Config.Latest = false
		*app.Config.LatestRC = true
	}

	// Allow --version to override --latest and --rc
	if len(*app.Config.Version) > 0 && (*app.Config.Latest || *app.Config.LatestRC) {
		*app.Config.Latest = false
		*app.Config.LatestRC = false
	}

	// Set the --log now
	logFilePath := filepath.Join(".", "tmp.install-go.log")
	if len(*app.Config.LogFile) > 0 {
		if _, cfgLogFileErr := os.Stat(*app.Config.LogFile); errors.Is(cfgLogFileErr, fs.ErrNotExist) {
			logFilePath = *app.Config.LogFile
		}
	}
	logFile, logFileErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_TRUNC|os.O_WRONLY, 0600)
	if logFileErr == nil {
		log.SetOutput(logFile)
	}

	// Handle --backup requests next
	if *app.Config.Backup {
		backupDir := filepath.Join(app.GODIR, "backups")
		if len(*app.Config.Output) > 0 {
			if outputInfo, outputErr := os.Stat(*app.Config.Output); outputErr == nil {
				// We can use --output
				if outputInfo.IsDir() {
					backupDir = *app.Config.Output
				} else {
					backupDir = filepath.Base(*app.Config.Output)
				}
			} else {
				// we cannot use --output
				if errors.Is(outputErr, fs.ErrNotExist) {
					// try creating --output
					mkdirErr := os.MkdirAll(backupDir, 0700)
					if mkdirErr != nil {
						// nothing... probably bad permissions?
						log.Fatal("Failed to write to " + backupDir)
					}
				}
			}
		}
		app.BackupDir = backupDir
		backupFilePath, backupErr := app.Backup()
		if backupErr != nil {
			log.Fatal(backupErr)
		}
		log.Printf("SUCCESS! Created %v", backupFilePath)
	}

	// Then (yes, you can stack commands but only 1 value per run)
	// Setup environment basics
	app.SetupEnvironment()
}
