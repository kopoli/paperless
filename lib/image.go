package paperless

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	util "github.com/kopoli/go-util"
)

// SaveImage saves the image to db and starts to process it
func SaveImage(filename string, data []byte, db *db, destdir string) (ret Image, err error) {

	supportedTypes := map[string]string{
		"image/gif":  "gif",
		"image/png":  "png",
		"image/jpeg": "jpg",
		"image/bmp":  "bmp",
	}
	ft := http.DetectContentType(data)
	var ok bool
	if ret.Fileid, ok = supportedTypes[ft]; !ok {
		err = util.E.New("Unsupported image type:", ft)
		return
	}

	ret.Checksum = Checksum(data)
	ret.AddDate = time.Now()
	ret.ScanDate = time.Now() // TODO
	ret.Filename = filename

	ret, err = db.addImage(ret)
	if err != nil {
		return
	}

	file := bytes.NewReader(data)
	fp, err := os.OpenFile(ret.OrigFile(destdir), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		_ = db.deleteImage(ret)
		return
	}
	defer fp.Close()

	_, err = io.Copy(fp, file)
	return
}
