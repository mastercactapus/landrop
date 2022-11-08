package app

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (srv *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	upl, err := r.MultipartReader()
	if err != nil {
		// invalid request
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for {
		part, err := upl.NextPart()
		if err == io.EOF {
			break
		}
		newPath := path.Join(srv.diskPath(r), part.FileName())
		// must be descendant of current path
		if !strings.HasPrefix(newPath, srv.diskPath(r)) {
			http.Error(w, "Invalid filename/path.", http.StatusBadRequest)
			return
		}

		err = srv.saveFile(srv.diskPath(r), part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (srv *Server) saveFile(dest string, part *multipart.Part) error {
	fullPath := filepath.Join(dest, part.FileName())

	fd, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	log.Println("UPLOAD:", fullPath)

	_, err = io.Copy(fd, part)
	if err != nil {
		fd.Close()
		return fmt.Errorf("save file: %w", err)
	}

	err = fd.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}
