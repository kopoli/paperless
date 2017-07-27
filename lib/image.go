package paperless

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	util "github.com/kopoli/go-util"
)

// SaveImage saves the image to db and starts to process it
func SaveImage(filename string, data []byte, db *db, destdir string, tags string) (ret Image, err error) {

	supportedTypes := map[string]string{
		"image/gif":  "gif",
		"image/png":  "png",
		"image/jpeg": "jpg",
		"image/bmp":  "bmp",
	}
	ft := http.DetectContentType(data)
	var ok bool
	if ret.Fileid, ok = supportedTypes[ft]; !ok {
		err = util.E.New("Unsupported image type: %s", ft)
		return
	}

	ret.Checksum = Checksum(data)
	ret.AddDate = time.Now()
	ret.ScanDate = time.Now() // TODO
	ret.Filename = filename

	taglist := strings.Split(tags, ",")
	if len(taglist) > 0 {
		ret.Tags = make([]Tag, len(taglist))

		for i := range taglist {
			ret.Tags[i].Name = strings.Trim(taglist[i], " \t\n\r")

			// Add a tag, ignore errors
			_, _ = db.addTag(ret.Tags[i])
		}
	}

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

func ProcessImage(img *Image, scriptname string, db *db, destdir string) (err error) {
	script := `
unpaper --version
convert -version
tesseract --version

convert -depth 8 $input pnm:$tmpUnpaper.pnm

unpaper -vv -s a4 -l single -dv 3.0 -dr 80.0 --overwrite $tmpUnpaper.pnm $tmpConvert

convert -normalize -colorspace Gray pnm:$tmpConvert pnm:$tmpTesseract

tesseract -l fin -psm 1 $tmpTesseract stdout > $contents

convert -trim -quality 80% +repage -type optimize pnm:$tmpConvert $cleanout

convert -trim -quality 80% +repage -type optimize -thumbnail 200x200> pnm:$tmpConvert $thumbout

`

	ch, err := NewCmdChainScript(script)
	if err != nil {
		return
	}

	buf := &bytes.Buffer{}
	s := Status{
		Environment: ch.Environment,
		Log:         buf,
	}
	s.Constants = map[string]string{
		"input":    img.OrigFile(destdir),
		"contents": img.TxtFile(destdir),
		"cleanout": img.CleanFile(destdir),
		"thumbout": img.ThumbFile(destdir),
	}
	s.AllowedCommands = map[string]bool{
		"convert":   true,
		"unpaper":   true,
		"tesseract": true,
		"file":      true,
		"cat":       true,
	}

	fmt.Fprintln(s.Log, "# Running the script named:", scriptname)

	err = RunCmdChain(ch, &s)
	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(img.TxtFile(destdir))
	// Ignore the error if the text-file was not generated
	if err != nil {
		data = []byte{}
	}

	img.InterpretDate = time.Now()
	img.ProcessLog = buf.String()
	img.Text = string(data)

	err = db.updateImage(*img)
	return
}

// DeleteImage deletes the image's files and data from the database
func DeleteImage(img *Image, db *db, destdir string) error {
	var err error
	ret := util.NewErrorList("Deleting image data failed")

	err = db.deleteImage(*img)
	if err != nil {
		ret.Append(err)
	}

	files := []string{
		img.OrigFile(destdir),
		img.TxtFile(destdir),
		img.CleanFile(destdir),
		img.ThumbFile(destdir),
	}

	for i := range files {
		err = os.Remove(files[i])
		if err != nil {
			ret.Append(util.E.Annotate(err, "Removing file ", files[i], "failed"))
		}
	}

	if ret.IsEmpty() {
		return nil
	}

	return ret
}
