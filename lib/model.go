package paperless

import "time"

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
	tags []Tag
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
