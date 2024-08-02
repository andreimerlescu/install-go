package types

import (
	"regexp"
	"strings"
)

type GoVersion struct {
	Version string      `json:"version"`
	Stable  bool        `json:"stable"`
	Files   []GoTarFile `json:"files"`
}

var rcRegex *regexp.Regexp
var latestRegex *regexp.Regexp

func (gv *GoVersion) IsLatest() bool {
	if latestRegex == nil {
		var regErr error
		latestRegex, regErr = regexp.Compile(`(\d{1,2})\.(\d{1,2})(?!rc).(\d{1,})`)
		if regErr != nil {
			return false
		}
	}
	v := strings.ReplaceAll(gv.Version, `go`, ``)
	matches := rcRegex.FindAllStringSubmatch(v, -1)
	return len(matches) > 0 && len(matches[0]) > 0
}

func (gv *GoVersion) IsRC() bool {
	if rcRegex == nil {
		var regErr error
		rcRegex, regErr = regexp.Compile(`(\d{1,2})\.\d{1,2}(?=rc)rc\d{1,}`)
		if regErr != nil {
			return false
		}
	}
	v := strings.ReplaceAll(gv.Version, `go`, ``)
	matches := rcRegex.FindAllStringSubmatch(v, -1)
	return len(matches) > 0 && len(matches[0]) > 0
}
