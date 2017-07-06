package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	util "github.com/kopoli/go-util"
	"github.com/kopoliitti/paperless/lib"

	"encoding/json"
	"mime/multipart"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/codegangsta/cli"

	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Supplied information

var version = "development"
var instDatadir string
var instWebdir string

/// Configuration

var (
	uploaddir  = "images"
	dbfilename = "paperless.sqlite3"
)

type progPreferences struct {
	datadir string
	dbfile  string
	webdir  string
	server  string
	verbose bool
}

var preferences = progPreferences{
	datadir: "./target",
	webdir:  "./web/",
	server:  "http://localhost:8078",
	verbose: false,
}

// checks given error and panics if not nil
func check(err error) {
	if err == nil {
		return
	}
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	if preferences.verbose {
		file, line := f.FileLine(pc[0])
		log.Panic(file, ":", line, ": ", f.Name(), ": ", err)
	} else {
		log.Panic(f.Name(), ": ", err)
	}
}

/// Image structure
type img struct {
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

func imagePath(image *img, body string) string {
	return path.Join(uploaddir,
		fmt.Sprintf("%05d-%s", image.Id, body))
}

func createUploadRequest(url string, path string) (req *http.Request) {
	fp, err := os.Open(path)
	check(err)
	defer fp.Close()
	// data, err := ioutil.ReadAll(fp)
	// check(err)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("image", fp.Name())
	check(err)
	_, err = io.Copy(part, fp)
	// _, err = part.Write(data)
	check(err)
	err = writer.Close()
	check(err)

	req, err = http.NewRequest("POST", url, &buf)
	req.Header.Set("Content-type", writer.FormDataContentType())
	check(err)
	return
}

/// Process images and upload to server
func ocrAndUpload(fname string, url string) {

	var image img
	var err error

	// Get the checksum of the image
	image.Checksum, err = paperless.ChecksumFile(fname)
	check(err)

	// fmt.Println(path, image.Checksum)

	// Run through OCR
	fp, err := ioutil.TempFile("", "paperless-ocr")
	check(err)
	defer fp.Close()
	tmpname := fp.Name()

	cmd := exec.Command(path.Join(preferences.webdir, "process.sh"), "ocr", fname, fp.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Command failed, output", output)
	}
	check(err)

	data, err := ioutil.ReadAll(fp)
	check(err)
	fp.Close()

	image.ProcessLog = string(output)
	image.Text = string(data)

	err = os.Remove(tmpname)
	check(err)

	// Marshal to JSON
	data, err = json.Marshal(image)
	check(err)
	fmt.Println(string(data))

	// Upload the data to server
	buf := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url+"/images", buf)
	check(err)
	req.Header.Set("Content-type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
	fmt.Println("Headers:", resp.Header)
	data, err = ioutil.ReadAll(resp.Body)
	check(err)
	fmt.Println("Body:", string(data))

	// Uploading the picture
	err = json.Unmarshal(data, &image)
	check(err)

	uploadURL := fmt.Sprintf(url+"/images/upload/%d", image.Id)
	fmt.Println("Uploadurl on ", uploadURL)
	req = createUploadRequest(uploadURL, fname)
	_, err = client.Do(req)
	check(err)
}

func upload(fname string, url string, tags []string) {

	// Upload the image
	client := &http.Client{}
	uploadUrl := fmt.Sprintf(url + "/images/upload")
	req := createUploadRequest(uploadUrl, fname)
	resp, err := client.Do(req)
	check(err)

	// Update the information
	var image img
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	check(err)

	if resp.StatusCode != http.StatusCreated {
		type jsonerror struct {
			Error string
		}
		var jerr jsonerror
		err = json.Unmarshal(data, &jerr)
		check(err)
		check(fmt.Errorf("Server error: %s", jerr.Error))
	}

	err = json.Unmarshal(data, &image)
	check(err)

	info, err := os.Stat(fname)
	image.TmScanned = info.ModTime().Unix()
	image.Tags = tags

	// Upload the updated information
	data, err = json.Marshal(image)
	check(err)

	buf := bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/images/%d", url, image.Id), buf)
	check(err)

	req.Header.Set("Content-type", "application/json")
	resp, err = client.Do(req)
	check(err)
}

/// Add command
func mainAdd(c *cli.Context) {
	preferences.verbose = c.GlobalBool("verbose")
	preferences.server = c.String("server-url")

	tags := strings.Split(c.String("tags"), ",")

	if len(tags) == 0 {
		log.Panic("Initial tags must be given for images.")
	}

	for _, fname := range c.Args() {
		if preferences.verbose {
			log.Println("Uploading image", fname)
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Uploading image \"%s\" failed due to error: %s", fname, err)
				}
			}()
			upload(fname, preferences.server, tags)
		}()
	}
}

/// SQlite functionality
type database struct {
	*sqlx.DB
}
type databaseTx struct {
	*sqlx.Tx
}

func openDbFile(dbfile string) database {
	create := false

	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		create = true
	}

	db, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbfile))
	check(err)

	if create {
		log.Println("Initializing the database file", dbfile)
		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS tag (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE ON CONFLICT IGNORE
);

-- The image data
CREATE TABLE IF NOT EXISTS img (
 id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  checksum TEXT UNIQUE ON CONFLICT ABORT,	-- checksum of the file
  fileid TEXT DEFAULT "",                       -- used to construct the processed image,
						--   thumbnail and text files
  isdiscarded INTEGER DEFAULT 0,                -- Should the image be ignored on ocr
  parentid INTEGER DEFAULT -1,                  -- ID of a parent image
  tmscanned INTEGER DEFAULT 0,                  -- timestamp when it was scanned
  tmprocessed INTEGER DEFAULT 0,                -- timestamp when it was processed
  tmreinterpret INTEGER DEFAULT 0,              -- timestamp when it was reinterpret

  processlog TEXT DEFAULT "",                   -- Log of processing
  filename TEXT DEFAULT "",                     -- The original filename
  isuploaded INTEGER DEFAULT 0,                 -- has a real image been uploaded to paperless yet.

  istobeocrd INTEGER DEFAULT 0			-- Needs to be re-ocr'd
);

CREATE VIRTUAL TABLE IF NOT EXISTS imgtext USING fts4 (
  text DEFAULT "",				-- the OCR'd text
  comment DEFAULT ""				-- freeform comment
);

-- Tags for an image
CREATE TABLE IF NOT EXISTS imgtag (
  tagid INTEGER REFERENCES tag(id) NOT NULL,
  imgid INTEGER REFERENCES img(id) NOT NULL
);
`)
		if err != nil {
			db.Close()
			check(err)
		}

	}
	db.Exec("PRAGMA busy_timeout=2000")
	if err != nil {
		db.Close()
		check(err)
	}

	return database{db}
}

func openDb() database {
	return openDbFile(preferences.dbfile)
}

func (tx databaseTx) updateTags(image *img) {
	_, err := tx.Exec("DELETE FROM imgtag WHERE imgid = $1", image.Id)
	check(err)
	for i := range image.Tags {
		_, err = tx.Exec("INSERT OR IGNORE INTO tag(name) VALUES ($1)",
			image.Tags[i])
		check(err)
		_, err = tx.Exec("INSERT INTO imgtag(tagid,imgid) "+
			"SELECT tag.id, $1 FROM tag WHERE tag.name = $2", image.Id, image.Tags[i])
		check(err)
	}
}

func (db database) addImage(image *img) {
	image.IsUploaded = 1
	image.IsToBeOcrd = 0
	image.IsDiscarded = 0

	sqltx, err := db.Beginx()
	tx := databaseTx{sqltx}
	check(err)

	_, err = tx.NamedExec("INSERT INTO img(checksum,fileid,processlog,filename) "+
		"VALUES (:checksum,:fileid,:processlog,:filename)", image)
	check(err)
	tmp := img{}
	err = tx.Get(&tmp, "SELECT id from img where checksum=$1", image.Checksum)
	check(err)
	image.Id = tmp.Id
	_, err = tx.NamedExec("INSERT INTO imgtext(rowid,text,comment) VALUES (:id,:text,:comment)",
		image)
	check(err)
	tx.updateTags(image)

	err = tx.Commit()
	check(err)
}

func (db database) getTags(image *img) {
	err := db.Select(&image.Tags, "SELECT tag.name FROM imgtag, tag "+
		"WHERE imgtag.imgid = $1 AND tag.id = imgtag.tagid", image.Id)
	check(err)
}

func (db database) getImage(id int) (image img) {
	image.Id = id
	err := db.Get(&image, "SELECT * FROM img, imgtext WHERE img.id = $1 AND img.id = imgtext.rowid", id)
	check(err)

	db.getTags(&image)

	return image
}

func (db database) updateImage(image img) {
	sqltx, err := db.Beginx()
	tx := databaseTx{sqltx}
	check(err)
	_, err = tx.NamedExec(`UPDATE img SET
	fileid = :fileid,
	isdiscarded = :isdiscarded,
	parentid = :parentid,
	tmscanned = :tmscanned,
	tmprocessed = :tmprocessed,
	tmreinterpret = :tmreinterpret,
	processlog = :processlog,
	filename = :filename,
	isuploaded = :isuploaded,
	istobeocrd = :istobeocrd
	WHERE img.id = :id`, image)
	check(err)

	_, err = tx.NamedExec(`UPDATE imgtext SET
	text = :text,
	comment = :comment
	WHERE imgtext.rowid = :id`, image)
	check(err)
	tx.updateTags(&image)
	err = tx.Commit()
	check(err)
}

func (db database) getAllImages(limit int, offset int, search string) (count int, images []img) {
	var err error
	var ids []int

	if search != "" {
		err = db.Select(&ids, "SELECT rowid FROM imgtext WHERE imgtext.text MATCH ?", search)
		check(err)
		count = len(ids)
		if count == 0 {
			return
		}

		err = db.Select(&images, "SELECT * from img, imgtext "+
			"WHERE imgtext.text MATCH ? AND imgtext.rowid = img.id LIMIT ? OFFSET ?",
			search, limit, offset)
	} else {
		err = db.Select(&ids, "SELECT id from img")
		check(err)
		count = len(ids)
		err = db.Select(&images, "SELECT * from img, imgtext "+
			"WHERE imgtext.rowid = img.id LIMIT ? OFFSET ?",
			limit, offset)
	}
	check(err)

	for i := range images {
		db.getTags(&images[i])
		images[i].Fileid = filepath.Base(imagePath(&images[i], ""))
	}

	return
}

/// web handlers
func handleResponse(response interface{}, code int, w http.ResponseWriter,
	r *http.Request) {

	w.WriteHeader(code)
	if response != nil {
		bytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal error.",
				http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
	log.Printf("Responder %s %s %s %d", r.RemoteAddr, r.Method, r.URL, code)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(preferences.webdir, "paperless.html"))
}

type imageResponse struct {
	Limit  int
	Offset int
	Count  int
	Images []img
}

func listImages(w http.ResponseWriter, r *http.Request) {
	var err error
	limit := -1
	offset := 0
	search := ""
	args := r.URL.Query()
	if _, ok := args["limit"]; ok {
		limit, err = strconv.Atoi(args["limit"][0])
		check(err)
	}
	if _, ok := args["offset"]; ok {
		limit, err = strconv.Atoi(args["offset"][0])
		check(err)
	}
	if _, ok := args["q"]; ok {
		search = args["q"][0]
	}

	db := openDb()
	defer db.Close()
	count, images := db.getAllImages(limit, offset, search)

	response := imageResponse{
		Limit:  limit,
		Offset: offset,
		Count:  count,
		Images: images,
	}

	handleResponse(response, http.StatusOK, w, r)
}

func addImage(w http.ResponseWriter, r *http.Request) {
	text, err := ioutil.ReadAll(r.Body)
	check(err)

	var image img
	err = json.Unmarshal(text, &image)
	check(err)

	log.Println(string(text))
	db := openDb()
	defer db.Close()
	db.addImage(&image)
	handleResponse(image, http.StatusCreated, w, r)
}

func receiveFile(image *img, r *http.Request) {
	err := r.ParseMultipartForm(1 * 1024 * 1024)
	check(err)
	file, header, err := r.FormFile("image")
	check(err)

	image.Filename = header.Filename
	extension := strings.ToLower(path.Ext(header.Filename))
	fp, err := os.OpenFile(path.Join(preferences.datadir, imagePath(image, "original"+extension)),
		os.O_WRONLY|os.O_CREATE, 0666)
	check(err)
	defer fp.Close()

	_, err = io.Copy(fp, file)
	check(err)

	// TODO error out if checksums differ

	// KOMENTO kuvan uploadaamiseksi
	// curl -v -F image=@testdata/korkea_prioriteetti_2012-13/SCAN0022.JPG http://localhost:8078/images/upload/1
}

func receiveImage(w http.ResponseWriter, r *http.Request) {

	var err error
	vars := mux.Vars(r)
	strid := vars["id"]
	id, err := strconv.Atoi(strid)
	check(err)

	db := openDb()
	defer db.Close()
	image := db.getImage(id)

	if image.IsUploaded != 0 {
		handleResponse(nil, http.StatusConflict, w, r)
	}

	log.Println("Receiveimage headers", r.Header)

	receiveFile(&image, r)
	image.IsUploaded = 1
	log.Printf("Updating the image %+v", image)

	db.updateImage(image)

	handleResponse(nil, http.StatusAccepted, w, r)
}

func receiveNewImage(w http.ResponseWriter, r *http.Request) {

	// receive the file
	err := r.ParseMultipartForm(1 * 1024 * 1024)
	check(err)
	file, header, err := r.FormFile("image")
	check(err)

	data, err := ioutil.ReadAll(file)
	check(err)

	var image img
	image.Checksum = paperless.Checksum(data)
	image.Filename = header.Filename
	image.TmScanned = time.Now().Unix()

	db := openDb()
	defer db.Close()

	// add to db
	db.addImage(&image)

	// add to directory
	extension := strings.ToLower(path.Ext(header.Filename))
	origpath := path.Join(preferences.datadir, imagePath(&image, "original"+extension))
	tgtimgpath := path.Join(preferences.datadir, imagePath(&image, "processed.jpg"))
	thumbpath := path.Join(preferences.datadir, imagePath(&image, "thumb.jpg"))

	fp, err := os.OpenFile(origpath, os.O_WRONLY|os.O_CREATE, 0666)
	check(err)

	_, err = fp.Write(data)
	fp.Close()
	check(err)

	job := paperless.Job{
		Job: func() {
			// Run through OCR
			log.Println("Processing", origpath)
			cmd := exec.Command(path.Join(preferences.webdir, "process.sh"), origpath, tgtimgpath, thumbpath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("Command failed, output", string(output))
			}
			check(err)

			re, err := regexp.Compile("__LOG_ENDS_HERE__")
			check(err)
			strs := re.Split(string(output), 2)

			db := openDb()
			defer db.Close()
			image = db.getImage(image.Id)
			image.ProcessLog = string(strs[0])
			image.Text = string(strs[1])
			image.TmProcessed = time.Now().Unix()

			db.updateImage(image)
			log.Println("Processing successful", origpath)
		},
		Finalize: func() {},
	}

	paperless.Pool.Do(job)

	handleResponse(image, http.StatusCreated, w, r)
}

func updateImage(w http.ResponseWriter, r *http.Request) {

	var err error
	vars := mux.Vars(r)
	strid := vars["id"]
	id, err := strconv.Atoi(strid)
	check(err)

	log.Println("Updating image", id)

	db := openDb()
	defer db.Close()
	image := db.getImage(id)

	text, err := ioutil.ReadAll(r.Body)
	check(err)

	err = json.Unmarshal(text, &image)
	check(err)

	db.updateImage(image)

	log.Printf("%+v", image)

	handleResponse(nil, http.StatusAccepted, w, r)
}

func httpErrorWrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-type", "application/json")
				fmt.Fprintf(w, `{"error" : "%s"}`, err)
				log.Printf("%s %s %s Error: %s", r.RemoteAddr, r.Method, r.URL, err)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func httpLogWrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

/// Web main
func mainStartWeb(c *cli.Context) {
	var err error

	if c.Bool("syslog") {
		paperless.SetupLogging()
	}

	preferences.webdir = c.String("web-directory")
	preferences.datadir = c.String("data-directory")
	preferences.dbfile = path.Join(preferences.datadir, dbfilename)
	preferences.verbose = c.GlobalBool("verbose")

	err = os.MkdirAll(preferences.datadir, 0755)
	check(err)
	err = os.MkdirAll(path.Join(preferences.datadir, uploaddir), 0755)
	check(err)

	db := openDb()
	db.Close()

	paperless.CreateDefaultPool(c.GlobalInt("jobs"))

	router := mux.NewRouter()
	router.HandleFunc("/", serveIndex)
	router.PathPrefix("/html/").Handler(http.StripPrefix("/html",
		http.FileServer(http.Dir(preferences.webdir))))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(path.Join(preferences.datadir, uploaddir)))))
	router.HandleFunc("/images", listImages).Methods("GET")
	router.HandleFunc("/images/upload", receiveNewImage).Methods("POST")
	router.HandleFunc("/images/{id:[0-9]+}", updateImage).Methods("POST")
	http.Handle("/", httpErrorWrap(httpLogWrap(router)))

	fmt.Println("Starting web!! port:", c.Int("port"))

	err = http.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")), nil)
	check(err)
}

func printErr(err error, message string, arg ...string) {
	msg := ""
	if err != nil {
		msg = fmt.Sprintf(" (error: %s)", err)
	}
	fmt.Fprintf(os.Stderr, "Error: %s%s.%s\n", message, strings.Join(arg, " "), msg)
}

func fault(err error, message string, arg ...string) {
	printErr(err, message, arg...)
	os.Exit(1)
}

func main() {
	opts := util.NewOptions()

	err := paperless.Cli(opts, os.Args)
	if err != nil {
		fault(err, "Command line parsing failed")
	}

	err = paperless.StartWeb(opts)
	if err != nil {
		fault(err, "Starting paperless web server failed")
	}
}

func main2() {
	if instDatadir != "" {
		preferences.datadir = instDatadir
	}
	if instWebdir != "" {
		preferences.webdir = instWebdir
	}

	app := cli.NewApp()
	app.Name = "paperless"
	app.Usage = "Paperless office utility"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "verbose, V",
			Usage:  "Verbose output",
			EnvVar: "PAPERLESS_VERBOSE",
		},
		cli.IntFlag{
			Name:   "jobs, j",
			Value:  runtime.NumCPU(),
			Usage:  "Number of jobs to process the images",
			EnvVar: "PAPERLESS_JOBS",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "add",
			Usage:  "Add a new picture",
			Action: mainAdd,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "preprocess, p",
					Usage:  "Preprocess added images through clipper",
					EnvVar: "PAPERLESS_PREPROCESS",
				},
				cli.StringFlag{
					Name:   "server-url, u",
					Value:  preferences.server,
					Usage:  "URL to upload to",
					EnvVar: "PAPERLESS_SERVER_URL",
				},
				cli.StringFlag{
					Name:   "tags, t",
					Value:  "",
					Usage:  "A comma separated list of tags",
					EnvVar: "PAPERLESS_TAGS",
				},
			},
		},
		{
			Name:   "start-web",
			Usage:  "Start web interface",
			Action: mainStartWeb,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:   "port, p",
					Value:  8078,
					Usage:  "Port to listen to",
					EnvVar: "PAPERLESS_PORT",
				},
				cli.StringFlag{
					Name:   "data-directory, d",
					Value:  preferences.datadir,
					Usage:  "Directory where the images and db are.",
					EnvVar: "PAPERLESS_DATADIR",
				},
				cli.StringFlag{
					Name:   "web-directory, w",
					Value:  preferences.webdir,
					Usage:  "Directory where the static web files are.",
					EnvVar: "PAPERLESS_WEBDIR",
				},
				cli.BoolFlag{
					Name:   "syslog, s",
					Usage:  "Print log through syslog",
					EnvVar: "PAPERLESS_SYSLOG",
				},
			},
		},
	}

	app.Run(os.Args)
}

// Local Variables:
// outline-regexp: "^////*\\|^func\\|^import"
// End:
