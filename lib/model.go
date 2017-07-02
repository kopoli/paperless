package paperless

import (
	"fmt"
	"path/filepath"
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

func (i *Image) imgFile(basedir, kind, extension string) string {
	ret, _ := filepath.Abs(filepath.Join(basedir,
		fmt.Sprintf("%05d-%s.%s", i.Id, kind, extension)))
	return ret
}
func (i *Image) OrigFile(basedir string) string {
	return i.imgFile(basedir, "original", i.Fileid)
}

func (i *Image) TxtFile(basedir string) string {
	return i.imgFile(basedir, "contents", "txt")
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
