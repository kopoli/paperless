package paperless

type Image struct {
	// in img
	Id            int
	Checksum      string
	Fileid        string
	IsDiscarded   int
	ParentId      int
	TmScanned     int64
	TmProcessed   int64
	TmReinterpret int64
	ProcessLog    string
	Filename      string
	IsUploaded    int
	IsToBeOcrd    int

	// in imgtext
	Text    string
	Comment string

	// in tags
	Tags []string
}

type Tag struct {
	Name string
	Comment string
}

type Script struct {
	Name string
	Script string
}
