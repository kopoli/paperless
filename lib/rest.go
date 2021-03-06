package paperless

//go:generate esc -o web-generated.go -pkg paperless -private -prefix ../web/paperless-frontend ../web/paperless-frontend/dist/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gamegos/jsend"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"

	"github.com/kopoli/go-util"
)

type backend struct {
	options util.Options
	db      *db
	imgdir  string

	staticURL string
}

/// JSON responding

func requestJson(r *http.Request, data interface{}) (err error) {
	text, err := ioutil.ReadAll(r.Body)
	if err != nil {
		goto requestError
	}
	err = json.Unmarshal(text, data)
	if err != nil {
		goto requestError
	}

	return
requestError:

	err = util.E.Annotate(err, "Converting HTTP request to JSON failed")
	return
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ item: \"todo\" }"))
}

func (b *backend) respondErr(w http.ResponseWriter, code int, err error) {
	jsend.Wrap(w).Status(code).Message(err.Error()).Send()
}

func getPaging(r *http.Request) (ret *Page) {
	since, err := strconv.Atoi(r.URL.Query().Get("since"))
	if err != nil {
		since = 0
	}
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		count = 20
	}

	if count > 0 {
		ret = &Page{SinceId: since, Count: count}
	}

	return
}

/// Tag handling
func (b *backend) tagHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	annotate := func(arg ...interface{}) {
		err = util.E.Annotate(err, arg...)
	}

	switch r.Method {
	case "POST":
		var t Tag
		err = requestJson(r, &t)
		if err != nil {
			annotate("JSON parsing failed")
			goto requestError
		}
		t, err = b.db.addTag(t)
		if err != nil {
			annotate("Adding tag to db failed")
			goto requestError
		}

		jsend.Wrap(w).Status(http.StatusCreated).Data(t).Send()
	case "GET":
		p := getPaging(r)

		tags, err := b.db.getTags(p)
		if err != nil {
			util.E.Annotate(err)
			annotate("Getting tags from db failed")
			goto requestError
		}

		jsend.Wrap(w).Status(http.StatusOK).Data(tags).Send()
	}

	return

requestError:
	b.respondErr(w, http.StatusBadRequest, err)
	return
}

func (b *backend) singleTagHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	annotate := func(arg ...interface{}) {
		err = util.E.Annotate(err, arg...)
	}

	var t Tag

	tagid, err := strconv.Atoi(chi.URLParam(r, "tagID"))
	if err == nil {
		t, err = b.db.getTag(tagid)
	}
	if err != nil {
		annotate("Invalid tag ID from URL")
		goto requestError
	}

	switch r.Method {
	case "GET":
		jsend.Wrap(w).Status(http.StatusOK).Data(t).Send()
	case "PUT":
		var t2 Tag
		err = requestJson(r, &t)
		if err != nil {
			annotate("JSON parsing failed")
			goto requestError
		}
		t.Comment = t2.Comment
		err = b.db.updateTag(t)
		if err != nil {
			annotate("Updating tag in db failed")
			goto requestError
		}
		jsend.Wrap(w).Status(http.StatusOK).Data(t).Send()
	case "DELETE":
		err = b.db.deleteTag(t)
		if err != nil {
			annotate("Deleting tag from db failed")
			goto requestError
		}
		jsend.Wrap(w).Status(http.StatusOK).Message("Deleted").Send()
	}

	return

requestError:
	b.respondErr(w, http.StatusBadRequest, err)
	return
}

// Image handling

type resultimg struct {
	PageResult

	Images []restimg
}

type restimg struct {
	Image

	OrigImg  string
	CleanImg string
	ThumbImg string
}

func (b *backend) wrapImage(img *Image) (ret restimg) {
	strip := func(s string) string {
		return b.staticURL + "/" + filepath.Base(s)
	}
	ret.Image = *img
	ret.OrigImg = strip(img.OrigFile(""))
	ret.CleanImg = strip(img.CleanFile(""))
	ret.ThumbImg = strip(img.ThumbFile(""))
	return
}

func (b *backend) wrapImages(imgs ImageResult) (ret resultimg) {
	ret.PageResult = imgs.PageResult

	ret.Images = make([]restimg, len(imgs.Images))
	for i := range imgs.Images {
		ret.Images[i] = b.wrapImage(&imgs.Images[i])
	}

	return
}

func (b *backend) imageHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	annotate := func(arg ...interface{}) {
		err = util.E.Annotate(err, arg...)
	}

	switch r.Method {
	case "POST":
		err = r.ParseMultipartForm(20 * 1024 * 1024)
		if err != nil {
			annotate("Parsing multipartform failed")
			goto requestError
		}
		file, header, e2 := r.FormFile("image")
		if e2 != nil {
			err = e2
			annotate("Could not find image from POST data")
			goto requestError
		}
		tags := r.FormValue("tags")
		if tags == "" {
			err = util.E.New("Tags are required when uploading.")
			goto requestError
		}
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, file)
		if err != nil {
			annotate("Could not copy image data to buffer")
			goto requestError
		}

		img, e2 := SaveImage(header.Filename, buf.Bytes(), b.db, b.imgdir, tags)
		if e2 != nil {
			err = e2
			annotate("Could not save image")
			goto requestError
		}

		err = ProcessImage(&img, "default", b.db, b.imgdir)
		if err != nil {
			annotate("Could not process image")
			// Ignore errors with this as the data could be
			// incomplete before deletion
			_ = DeleteImage(&img, b.db, b.imgdir)
			goto requestError
		}

		jsend.Wrap(w).Status(http.StatusCreated).Data(img).Send()
	case "GET":
		p := getPaging(r)
		query := r.URL.Query().Get("q")
		tag := r.URL.Query().Get("t")

		s := &Search{
			Match: query,
			Tag: tag,
		}

		images, e2 := b.db.getImages(p, s)
		if e2 != nil {
			err = e2
			annotate("Getting images from db failed")
			goto requestError
		}

		jsend.Wrap(w).Status(http.StatusOK).Data(b.wrapImages(images)).Send()
	}

	return

requestError:
	b.respondErr(w, http.StatusBadRequest, err)
	return
}

func (b *backend) singleImageHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	annotate := func(arg ...interface{}) {
		err = util.E.Annotate(err, arg...)
	}

	var img Image

	id, err := strconv.Atoi(chi.URLParam(r, "imageID"))
	if err == nil {
		img, err = b.db.getImage(id)
	}
	if err != nil {
		annotate("Invalid image ID from URL")
		goto requestError
	}

	switch r.Method {
	case "GET":
		jsend.Wrap(w).Status(http.StatusOK).Data(b.wrapImage(&img)).Send()
	case "PUT":
		var img2 Image
		err = requestJson(r, &img2)
		if err != nil {
			annotate("JSON parsing failed")
			goto requestError
		}
		img.Text = img2.Text
		img.Comment = img2.Comment
		err = b.db.updateImage(img)
		if err != nil {
			annotate("Updating image in db failed")
			goto requestError
		}
		jsend.Wrap(w).Status(http.StatusOK).Data(b.wrapImage(&img)).Send()
	case "DELETE":
		err = DeleteImage(&img, b.db, b.imgdir)
		if err != nil {
			annotate("Deleting image failed")
			goto requestError
		}
		jsend.Wrap(w).Status(http.StatusOK).Message("Deleted").Send()
	}

	return

requestError:
	b.respondErr(w, http.StatusBadRequest, err)
	return
}

/// Script handling

func (b *backend) loadScriptCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func (b *backend) versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ \"version\": \"" + b.options.Get("version", "unversioned") + "\" }"))
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

// FileServer conveniently sets up a http.FileServer handler to serve static
// files from a http.FileSystem.  As chi updated to 3.x, the equivalent
// function was removed. This one is copied from the example in:
// https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func StartWeb(o util.Options) (err error) {

	db, err := openDbFile(o.Get("database-file", "paperless.sqlite3"))
	if err != nil {
		return
	}
	defer db.Close()

	imgdir := o.Get("image-directory", "images")
	err = os.MkdirAll(imgdir, 0755)
	if err != nil {
		return
	}

	back := &backend{o, db, imgdir, "/static"}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(600 * time.Second))
	r.Use(corsHandler)

	// REST API
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", back.versionHandler)
		r.Route("/image", func(r chi.Router) {
			r.Get("/", back.imageHandler)
			r.Post("/", back.imageHandler)
			r.Route("/{imageID}", func(r chi.Router) {
				r.Get("/", back.singleImageHandler)
				r.Put("/", back.singleImageHandler)
				r.Delete("/", back.singleImageHandler)
			})
		})

		r.Route("/tag", func(r chi.Router) {
			r.Get("/", back.tagHandler)
			r.Post("/", back.tagHandler)
			r.Route("/{tagID}", func(r chi.Router) {
				r.Get("/", back.singleTagHandler)
				r.Put("/", back.singleTagHandler)
				r.Delete("/", back.singleTagHandler)
			})
		})
		r.Route("/script", func(r chi.Router) {
			r.Get("/", todoHandler)
			r.Post("/", todoHandler)
			r.Route("/{scriptID}", func(r chi.Router) {
				r.Use(back.loadScriptCtx)
				r.Get("/", todoHandler)
				r.Put("/", todoHandler)
				r.Delete("/", todoHandler)
			})
		})
	})

	FileServer(r, back.staticURL, http.Dir(imgdir))
	FileServer(r, "/dist", _escDir(false, "/dist/"))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		fs := _escFS(false)
		httpfile, _ := fs.Open("/dist/index.html")
		st, _ := httpfile.Stat()
		http.ServeContent(w, r, "index.html", st.ModTime(), httpfile)
	})

	if o.IsSet("print-routes") {
		fmt.Println(docgen.JSONRoutesDoc(r))
		return
	}

	http.ListenAndServe(o.Get("listen-address", ":8078"), r)

	return
}
