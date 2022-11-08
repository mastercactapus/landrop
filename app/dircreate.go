package app

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (srv *Server) handleDirCreate(w http.ResponseWriter, r *http.Request) {
	newPath := path.Join(srv.diskPath(r), r.FormValue("name"))

	// must be descendant of current path
	if !strings.HasPrefix(newPath, srv.diskPath(r)) {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	err := os.MkdirAll(filepath.FromSlash(newPath), 0o755)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("MKDIR:", filepath.FromSlash(newPath))

	http.Redirect(w, r, path.Join(srv.fsPath(r), r.FormValue("name")), http.StatusFound)
}
