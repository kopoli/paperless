package paperless

import (
	"fmt"
	"path"
	"time"
)

type Image struct {
	// in img
	Id            int
	Checksum      string
	Fileid        string
	ScanDate      time.Time
	AddDate       time.Time
	InterpretDate time.Time
	ProcessLog    string
	Filename      string

	// in imgtext
	Text    string
	Comment string

	// in tags
	Tags []Tag
}

func (i Image) OrigFile(basedir string) string {
	return path.Join(basedir, fmt.Sprintf("%05d-original-%s", i.Id, i.Fileid))
}

type Tag struct {
	Id      int
	Name    string
	Comment string
}

type Script struct {
	Id     int
	Name   string
	Script string
}
