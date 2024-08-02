package types

type GoTarFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Checksum string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}
