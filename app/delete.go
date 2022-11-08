package app

import (
	"net/http"
	"os"
	"path"
	"strings"
)

func (srv *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	newPath := path.Join(srv.diskPath(r), r.FormValue("name"))

	// must be descendant of current path
	if !strings.HasPrefix(newPath, srv.diskPath(r)) {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	var err error
	if r.FormValue("recursive") == "1" {
		err = os.RemoveAll(newPath)
	} else {
		err = os.Remove(newPath)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, srv.fsPath(r), http.StatusFound)
}
