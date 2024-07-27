package types

type Config struct {
	Version   *string `json:"version" yaml:"Version"`
	Latest    *bool   `json:"latest" yaml:"Latest"`
	LatestRC  *bool   `json:"latest_rc" yaml:"LatestRC"`
	Install   *bool   `json:"install" yaml:"Install"`
	Uninstall *bool   `json:"uninstall" yaml:"Uninstall"`
	GODIR     *string `json:"godir" yaml:"GODIR"`
	Switch    *bool   `json:"switch" yaml:"Switch"`
	Backup    *bool   `json:"backup" yaml:"Backup"`
	Output    *string `json:"output" yaml:"Output"`
	GOOS      *string `json:"goos" yaml:"GOOS"`
	GOARCH    *string `json:"goarch" yaml:"GOARCH"`
	Debug     *bool   `json:"debug" yaml:"Debug"`
	LogFile   *string `json:"log_file" yaml:"LogFile"`
	Help      *bool   `json:"help" yaml:"Help"`
}
