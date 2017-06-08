package paperless

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gamegos/jsend"
	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"

	"github.com/kopoli/go-util"
)

type backend struct {
	options util.Options
	db      *db
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

func (b *backend) loadImageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
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
		count = 0
	}

	if count > 0 {
		ret = &Page{SinceId: since, Count: count}
	}

	return
}

/// Tag handling
func (b *backend) tagHandler(w http.ResponseWriter, r *http.Request) {

	errsend := func(err error) {
		b.respondErr(w, http.StatusBadRequest, err)
	}
	switch r.Method {
	case "POST":
		var t Tag
		err := requestJson(r, &t)
		if err != nil {
			errsend(err)
			return
		}
		t, err = b.db.addTag(t)
		if err != nil {
			errsend(err)
			return
		}
		jsend.Wrap(w).Status(http.StatusCreated).Data(t).Send()
	case "GET":
		p := getPaging(r)

		tags, err := b.db.getTags(p)
		if err != nil {
			errsend(err)
			return
		}

		jsend.Wrap(w).Status(http.StatusOK).Data(tags).Send()
	}

}

func (b *backend) loadTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

/// Script handling

func (b *backend) loadScriptCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func (b *backend) versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ \"version\": \"" + b.options.Get("version", "unversioned") + "\" }"))
}

func StartWeb(o util.Options) (err error) {

	db, err := openDbFile(o.Get("database-file", "paperless.sqlite3"))
	if err != nil {
		return
	}

	defer db.Close()
	back := &backend{o, db}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// REST API
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", back.versionHandler)
		r.Route("/image", func(r chi.Router) {
			r.Get("/", todoHandler)
			r.Post("/", todoHandler)
			r.Route("/:imageID", func(r chi.Router) {
				r.Use(back.loadImageCtx)
				r.Get("/", todoHandler)
				r.Put("/", todoHandler)
				r.Delete("/", todoHandler)
			})
		})

		r.Route("/tag", func(r chi.Router) {
			r.Get("/", back.tagHandler)
			r.Post("/", back.tagHandler)
			r.Route("/:tagID", func(r chi.Router) {
				r.Use(back.loadTagCtx)
				r.Get("/", todoHandler)
				r.Put("/", todoHandler)
				r.Delete("/", todoHandler)
			})
		})
		r.Route("/script", func(r chi.Router) {
			r.Get("/", todoHandler)
			r.Post("/", todoHandler)
			r.Route("/:scriptID", func(r chi.Router) {
				r.Use(back.loadScriptCtx)
				r.Get("/", todoHandler)
				r.Put("/", todoHandler)
				r.Delete("/", todoHandler)
			})
		})
	})

	// Web interface
	webdir := o.Get("webdir", "web")
	uploaddir := o.Get("uploaddir", "static")
	r.FileServer("/html", http.Dir(webdir))
	r.FileServer("/static", http.Dir(uploaddir))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(webdir, "paperless.html"))
	})

	if o.IsSet("print-routes") {
		fmt.Println(docgen.JSONRoutesDoc(r))
		return
	}

	http.ListenAndServe(o.Get("address-port", ":8078"), r)

	return
}
