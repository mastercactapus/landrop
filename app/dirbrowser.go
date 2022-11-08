package app

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed dirbrowser.html
var html string

var tmpl = template.Must(template.New("dirbrowser").Funcs(template.FuncMap{
	"humanSize": fmtBytes,
	"humanTime": fmtTimeSince,
}).Parse(html))

type fileEntry struct {
	Name  string
	Href  string
	Type  string
	IsDir bool
	Size  string
	Mod   string
}

type breadcrumb struct {
	Name string
	Href string
}
type tmplData struct {
	Files  []fileEntry
	Server Config
	URL    string
}

func (t tmplData) Breadcrumbs() []breadcrumb {
	path := strings.TrimPrefix(t.URL, "/")
	path = strings.TrimSuffix(path, "/")
	if len(path) == 0 {
		return nil
	}

	parts := strings.Split(path, "/")
	var trail []breadcrumb
	for i, p := range parts {
		trail = append(trail, breadcrumb{
			Name: p,
			Href: "/" + strings.Join(parts[:i+1], "/"),
		})
	}
	return trail
}

func (srv *Server) handleDirBrowser(w http.ResponseWriter, r *http.Request, f http.File) {
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortBy := r.FormValue("sort")
	desc := r.FormValue("desc") == "1"
	// sort directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		}
		if !files[i].IsDir() && files[j].IsDir() {
			return false
		}

		if desc {
			i, j = j, i
		}

		switch {
		case sortBy == "size":
			return files[i].Size() < files[j].Size()
		case sortBy == "time":
			return files[i].ModTime().Before(files[j].ModTime())
		default:
			fallthrough
		case sortBy == "name":
			return files[i].Name() < files[j].Name()
		}
	})
	var c tmplData
	c.Server = srv.cfg
	c.URL = strings.TrimSuffix(r.URL.Path, "/")

	for _, f := range files {
		var typeName string
		if f.IsDir() {
			typeName = "directory"
		}
		c.Files = append(c.Files, fileEntry{
			Name:  f.Name(),
			Href:  path.Join(r.URL.Path, f.Name()),
			Size:  fmtBytes(f.Size()),
			Mod:   fmtTimeSince(f.ModTime()),
			IsDir: f.IsDir(),
			Type:  typeName,
		})
	}
	if srv.cfg.Scan && len(c.Files) > 0 {
		args := []string{"-b", "--"}
		for _, f := range c.Files {
			args = append(args, filepath.Join(srv.diskPath(r), f.Name))
		}
		cmd := exec.CommandContext(r.Context(), "file", args...)
		out, err := cmd.Output()
		if err != nil {
			log.Println("file:", err)
		}
		for i, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			c.Files[i].Type = line
		}
	}

	err = tmpl.Execute(w, c)
	if err != nil {
		log.Println("ERROR:", err)
	}
}
