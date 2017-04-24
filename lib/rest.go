package paperless

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"

	"github.com/kopoli/go-util"
)

func todoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ item: \"todo\" }"))
}

func loadImageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func loadTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func StartWeb(o util.Options) (err error) {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// REST API
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/image", func(r chi.Router) {
			r.Get("/", todoHandler)
			r.Post("/", todoHandler)
			r.Route("/:imageID", func(r chi.Router) {
				r.Use(loadImageCtx)
				r.Get("/", todoHandler)
				r.Put("/", todoHandler)
				r.Delete("/", todoHandler)
			})
		})

		r.Route("/tag", func(r chi.Router) {
			r.Get("/", todoHandler)
			r.Post("/", todoHandler)
			r.Route("/:tagID", func(r chi.Router) {
				r.Use(loadTagCtx)
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
