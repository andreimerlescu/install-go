package types

import (
	"fmt"
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
	PATH                []string
	BackupDir           string
	latestBackup        string
	latestBackupVersion string
}

func (app *Application) CLI() configurable.IConfigurable {
	return app.cfg
}

func (app *Application) Load(configFilePath string) error {
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

	*app.Config.Version = clean(*app.Config.Version)
	*app.Config.Output = clean(*app.Config.Output)
	*app.Config.LogFile = clean(*app.Config.LogFile)
	*app.Config.GOOS = clean(*app.Config.GOOS)
	*app.Config.GOARCH = clean(*app.Config.GOARCH)

	app.GODIR = *app.Config.GODIR
	app.GOARCH = *app.Config.GOARCH
	app.GOOS = *app.Config.GOOS
	app.GOBIN = fmt.Sprintf("%v/bin", app.GODIR)
	app.GOSHIMS = fmt.Sprintf("%v/shims", app.GODIR)
	app.GOSCRIPTS = fmt.Sprintf("%v/scripts", app.GODIR)
	app.GOPATH = fmt.Sprintf("%v/path", app.GODIR)
	app.GOROOT = fmt.Sprintf("%v/root", app.GODIR)

	return app.CLI().Parse(configFilePath)
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
