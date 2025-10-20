package util

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"sync"
)

func ParseBody(r *http.Request, o interface{}) error {
	if body, err := io.ReadAll(r.Body); err != nil {
		return errors.New("parsing body failed")
	} else {
		if err = json.Unmarshal(body, o); err != nil {
			return errors.New("parsing body failed")
		}
	}
	return nil
}

type TemplateHandler struct {
	Once     sync.Once
	FileName string
	Template *template.Template
}

func (handler *TemplateHandler) ParseFileOnce() {
	handler.Once.Do(func() {
		handler.Template = template.Must(template.ParseFiles(filepath.Join("templates", handler.FileName)))
	})
}
