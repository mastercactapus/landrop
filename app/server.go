package app

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
)

type Server struct {
	cfg Config

	fs http.FileSystem

	fsH http.Handler
}

type Config struct {
	Name     string
	Writable bool
	Scan     bool
	Dir      string
}

func NewServer(cfg Config) *Server {
	return &Server{
		cfg: cfg,
		fs:  http.Dir(cfg.Dir),
		fsH: http.FileServer(http.Dir(cfg.Dir)),
	}
}

func (srv *Server) diskPath(r *http.Request) string {
	return path.Join(srv.cfg.Dir, srv.fsPath(r))
}

func (srv *Server) fsPath(r *http.Request) string {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	return path.Clean(upath)
}

func (srv *Server) readOnly(w http.ResponseWriter, r *http.Request) bool {
	if srv.cfg.Writable {
		return false
	}

	http.Error(w, "read-only filesystem", http.StatusForbidden)
	return true
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	switch r.Method {
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	case "POST":
		if srv.readOnly(w, r) {
			return
		}

		// handle separately because we need to use MultipartReader
		if r.URL.Query().Get("upload") == "1" {
			srv.handleUpload(w, r)
			return
		}

		switch r.FormValue("action") {
		case "mkdir":
			srv.handleDirCreate(w, r)
		case "rm":
			srv.handleDelete(w, r)
		case "mv":
			srv.handleRename(w, r)
		default:
			http.Error(w, fmt.Sprintf("unkown action '%s'", r.FormValue("action")), http.StatusBadRequest)
		}
		return
	case "GET", "HEAD":
		// handled below
	}

	path := srv.fsPath(r)
	f, err := srv.fs.Open(path)
	if err != nil {
		// fallback to the fs handler
		srv.fsH.ServeHTTP(w, r)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		// fallback to the fs handler
		srv.fsH.ServeHTTP(w, r)
		return
	}

	if info.IsDir() {
		srv.handleDirBrowser(w, r, f)
		return
	}

	if info.Name() == "index.html" {
		// The fs handler will try to redirect to the directory
		// if we don't do this.
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}

	// all other cases, fallback to the fs handler
	srv.fsH.ServeHTTP(w, r)
}
