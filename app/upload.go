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

		// must be descendant of current path
		if !strings.HasPrefix(path.Join(srv.diskPath(r), part.FileName()), srv.diskPath(r)) {
			http.Error(w, "Invalid filename/path.", http.StatusBadRequest)
			return
		}

		err = srv.saveFile(srv.diskPath(r), part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.URL.Query().Get("redirect") == "1" {
		http.Redirect(w, r, srv.fsPath(r), http.StatusFound)
	}
}

func findNextName(name string) (string, error) {
	if _, err := os.Stat(filepath.FromSlash(name)); err != nil {
		return name, nil
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for n := 1; n < 1000; n++ {
		newName := fmt.Sprintf("%s (%d)%s", base, n, ext)
		if _, err := os.Stat(filepath.FromSlash(newName)); err != nil {
			return newName, nil
		}
	}

	return "", fmt.Errorf("too many files with the same name")
}

func (srv *Server) saveFile(dest string, part *multipart.Part) error {
	newPath, err := findNextName(path.Join(dest, part.FileName()))
	if err != nil {
		return fmt.Errorf("find next name: %v", err)
	}

	fd, err := os.Create(filepath.FromSlash(newPath))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	log.Println("UPLOAD:", newPath)

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
