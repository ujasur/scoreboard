package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type page struct {
	Name    string
	Title   string
	Data    interface{}
	Version string
	Team    string
}

type templateMgr struct {
	templates map[string]*template.Template
	defaults  *page
}

func newTemplateMgr(templateDir string, defaults *page) *templateMgr {
	m := new(templateMgr)
	m.defaults = defaults

	root, err := template.New("root").Delims("[[", "]]").Parse(`[[define "root" ]] [[ template "base" . ]] [[ end ]]`)
	if err != nil {
		log.Fatal(err)
	}

	layoutFiles, err := filepath.Glob(filepath.Join(templateDir, "*.tpl"))
	if err != nil {
		log.Fatal(err)
	}

	htmlFiles, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		log.Fatal(err)
	}

	m.templates = make(map[string]*template.Template)
	for _, file := range htmlFiles {
		name := filepath.Base(file)
		files := append(layoutFiles[:], file)
		tpl, err := template.Must(root.Clone()).Delims("[[", "]]").ParseFiles(files...)
		if err != nil {
			log.Fatal(err)
		}
		m.templates[name] = tpl
	}
	return m
}

func (m *templateMgr) render(w http.ResponseWriter, p *page, etag string) error {
	t, ok := m.templates[p.Name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if len(p.Team) == 0 && m.defaults != nil {
		p.Team = m.defaults.Team
	}
	if len(p.Version) == 0 && m.defaults != nil {
		p.Version = m.defaults.Version
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	if err := t.Execute(buf, p); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	tag := fmt.Sprintf("%x", sha1.Sum(buf.Bytes()))
	if tag == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Header().Set("Cache-Control", "max-age=60")
	w.Header().Set("Etag", tag)
	_, err := buf.WriteTo(w)
	return err
}
