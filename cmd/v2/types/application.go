package types

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreimerlescu/configurable"
)

type Application struct {
	cfg                 configurable.IConfigurable
	Config              *Config
	Responder           *Responder
	HOME                string
	GODIR               string
	GOBIN               string
	GOPATH              string
	IGOVERSION          string
	GOOS                string
	GOARCH              string
	GOROOT              string
	GOSHIMS             string
	GOSCRIPTS           string
	DownloadsDir        string
	PATH                []string
	BackupDir           string
	latestBackup        string
	latestBackupVersion string
}

func (app *Application) CLI() configurable.IConfigurable {
	return app.cfg
}

func (app *Application) ListVersions() string {
	var sb = strings.Builder{}
	sb.WriteString("Listing available versions: ")
	otherCombos := make(map[string]string, 0)
	versions := app.Responder.FetchAllVersions()
	for _, gov := range versions {
		v := strings.ReplaceAll(gov.Version, `go`, ``)
		sb.WriteString(fmt.Sprintf("--version %v\n", v))
		for _, gtf := range gov.Files {
			if len(gtf.OS) == 0 && len(gtf.Arch) == 0 {
				continue // this is a source tar.gz
			}
			if app.GOOS == gtf.OS && app.GOARCH == gtf.Arch {
				sb.WriteString(fmt.Sprintf("  --version %v --os %v --arch %v\n", v, gtf.OS, gtf.Arch))
			} else {
				otherCombos[gtf.OS] = gtf.Arch
			}
		}
	}
	if len(otherCombos) > 0 {
		sb.WriteString("\nOverride --os and --arch with these possible combinations:\n")
		for os, arch := range otherCombos {
			sb.WriteString(fmt.Sprintf("  --os %v --arch %v --version ''\n", os, arch))
		}
	}
}

func (app *Application) Version() string {
	defaultLatest := "1.22.5"
	defaultRC := "1.23rc2"
	if len(*app.Config.Version) == 0 {
		// --version empty string
		if !*app.Config.Latest && !*app.Config.LatestRC {
			return defaultLatest
		}

		versions := app.Responder.FetchAllVersions()
		if !*app.Config.Latest && *app.Config.LatestRC {
			// --rc enabled
			for _, version := range versions {
				if version.IsRC() {
					defaultRC = strings.ReplaceAll(version.Version, `go`, ``)
				}
			}
		}
	}

	return ""
}

func (app *Application) VersionTarballURL() string {
	return fmt.Sprintf("https://go.dev/dl/go%v.%v-%v.tar.gz", app.Version(), app.GOOS, app.GOARCH)
}

func (app *Application) Load() error {
	removeFirstCharIfQuote := func(s string) string {
		if len(s) > 0 && s[0] == '"' {
			return s[1:]
		}
		return s
	}

	removeLastCharIfQuote := func(s string) string {
		if len(s) > 0 && s[len(s)-1] == '"' {
			return s[:len(s)-1]
		}
		return s
	}

	clean := func(in string) string {
		in = strings.ReplaceAll(in, "\n", " ")
		in = strings.ReplaceAll(in, "\t", " ")
		in = strings.Join(strings.Fields(in), " ")

		in = removeFirstCharIfQuote(in)
		in = removeLastCharIfQuote(in)
		return in
	}

	// Clean String
	*app.Config.Version = clean(*app.Config.Version)
	*app.Config.Output = clean(*app.Config.Output)
	*app.Config.LogFile = clean(*app.Config.LogFile)
	*app.Config.GOOS = clean(*app.Config.GOOS)
	*app.Config.GOARCH = clean(*app.Config.GOARCH)

	// Assign globals
	app.GODIR = *app.Config.GODIR
	app.GOARCH = *app.Config.GOARCH
	app.GOOS = *app.Config.GOOS
	app.GOBIN = fmt.Sprintf("%v/bin", app.GODIR)
	app.GOSHIMS = fmt.Sprintf("%v/shims", app.GODIR)
	app.GOSCRIPTS = fmt.Sprintf("%v/scripts", app.GODIR)
	app.GOPATH = fmt.Sprintf("%v/path", app.GODIR)
	app.GOROOT = fmt.Sprintf("%v/root", app.GODIR)

	// Look for config file
	configFile := ""
	tempConfig := filepath.Join(".", "config.yaml")
	if _, configFileErr := os.Stat(tempConfig); configFileErr != nil {
		// No such ./config.yaml file...
		if errors.Is(configFileErr, fs.ErrNotExist) {
			configFile = ""
			// File ./config.yaml exists, but permissions block the script
		} else if errors.Is(configFileErr, fs.ErrPermission) {
			log.Printf("WARNING: Permission denied attempting to read %v with error: %v", tempConfig, configFileErr)
			configFile = ""
			// File ./config.yaml id invalid, and cannot be read
		} else if errors.Is(configFileErr, fs.ErrInvalid) {
			log.Printf("DANGER: %v", configFileErr)
			configFile = ""
		} else {
			log.Printf("WARNING: Found uncaught configFileErr = %v", configFileErr)
			configFile = ""
		}
	} else {
		configFile = fmt.Sprintf("%v", tempConfig)
	}

	return app.CLI().Parse(configFile)
}

func (app *Application) Backup() (string, error) {
	// TODO: Implement logic needed to create a full backup of the app.GODIR to app.BackupDir
	app.latestBackup = filepath.Join(
		app.BackupDir,
		fmt.Sprintf(
			"backup-go-%v-%d-%d-%d.zip",
			app.Config.Version,
			time.Now().Local().Year(),
			time.Now().Local().Month(),
			time.Now().Local().Day(),
		),
	)
	app.latestBackupVersion = *app.Config.Version
	return app.latestBackup, nil
}

func NewApplication(cfg configurable.IConfigurable) (*Application, error) {
	if cfg != nil {
		return &Application{
			cfg: cfg,
		}, nil
	}
	return nil, fmt.Errorf("cfg is nil")
}
