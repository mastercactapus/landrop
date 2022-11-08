package app

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (srv *Server) handleRename(w http.ResponseWriter, r *http.Request) {
	oldPath := path.Join(srv.diskPath(r), r.FormValue("name"))
	newPath := path.Join(srv.diskPath(r), r.FormValue("newName"))

	// must be descendant of current path
	if !strings.HasPrefix(oldPath, srv.diskPath(r)) {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(newPath, srv.diskPath(r)) {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	if r.FormValue("overwrite") != "1" {
		_, err := os.Stat(newPath)
		if err == nil {
			http.Error(w, "file already exists.", http.StatusBadRequest)
			return
		}
	}

	err := os.Rename(filepath.FromSlash(oldPath), filepath.FromSlash(newPath))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("MV:", oldPath, newPath)

	http.Redirect(w, r, path.Dir(path.Join(srv.fsPath(r), r.FormValue("newName"))), http.StatusFound)
}
