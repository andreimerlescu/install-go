package types

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Responder struct {
	ctx    context.Context
	cancel context.CancelFunc
	app    *Application
	wg     *sync.WaitGroup
}

func NewResponder(ctx context.Context, app *Application, wg *sync.WaitGroup) *Responder {
	rctx, rcancel := context.WithCancel(ctx)
	return &Responder{
		ctx:    rctx,
		cancel: rcancel,
		app:    app,
		wg:     wg,
	}
}

func (r *Responder) StopResponding() {
	r.ctx.Done() // stop responding to future requests
}

func (r *Responder) ShowHelp() {
	r.wg.Add(1)
	defer r.wg.Done()

	defer r.StopResponding()

	if *r.app.Config.Help {
		fmt.Println(r.app.CLI().Usage())
		return
	}
}

func (r *Responder) AddressUsability() {
	r.wg.Add(1)
	defer r.wg.Done()

	// If user sets --version to latest, its a user-error, catch it and correct it
	if *r.app.Config.Version == "latest" {
		r.app.Config.Version = nil
		*r.app.Config.Latest = true
		*r.app.Config.LatestRC = false
	}

	if *r.app.Config.Version == "latest-rc" {
		r.app.Config.Version = nil
		*r.app.Config.Latest = false
		*r.app.Config.LatestRC = true
	}

	// Allow --version to override --latest and --rc
	if len(*r.app.Config.Version) > 0 && (*r.app.Config.Latest || *r.app.Config.LatestRC) {
		*r.app.Config.Latest = false
		*r.app.Config.LatestRC = false
	}
}

func (r *Responder) SetLogging() {
	r.wg.Add(1)
	defer r.wg.Done()

	// Set the --log now
	logFilePath := filepath.Join(".", "tmp.install-go.log")
	if len(*r.app.Config.LogFile) > 0 {
		if _, cfgLogFileErr := os.Stat(*r.app.Config.LogFile); errors.Is(cfgLogFileErr, fs.ErrNotExist) {
			logFilePath = *r.app.Config.LogFile
		}
	}
	logFile, logFileErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_TRUNC|os.O_WRONLY, 0600)
	if logFileErr == nil {
		log.SetOutput(logFile)
	}
}

func (r *Responder) TakeBackup() {
	r.wg.Add(1)
	defer r.wg.Done()

	defer r.StopResponding()

	// Handle --backup requests next
	if *r.app.Config.Backup {
		backupDir := filepath.Join(r.app.GODIR, "backups")
		if len(*r.app.Config.Output) > 0 {
			if outputInfo, outputErr := os.Stat(*r.app.Config.Output); outputErr == nil {
				// We can use --output
				if outputInfo.IsDir() {
					backupDir = *r.app.Config.Output
				} else {
					// --output is a file, so lets take its directory and use that
					backupDir = filepath.Base(*r.app.Config.Output)
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
		r.app.BackupDir = backupDir
		backupFilePath, backupErr := r.app.Backup()
		if backupErr != nil {
			log.Fatal(backupErr)
		}
		log.Printf("SUCCESS! Created %v", backupFilePath)
	}
}

func (r *Responder) PrepareWorkspace() {
	r.wg.Add(1)
	defer r.wg.Done()

	if _, dirErr := os.Stat(r.app.GODIR); dirErr != nil {
		if errors.Is(dirErr, fs.ErrNotExist){
			mkdirErr := os.MkdirAll(r.app.GODIR, 0755)
			if mkdirErr == nil {
				return
			} else {
				log.Printf("Installed workspace at %v! Now use --install --latest", r.app.GODIR)
			}
		}
	}
}
