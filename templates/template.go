package templates

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/te6lim/go-chat/database"
)

type TemplateHandler struct {
	Once     sync.Once
	FileName string
	Template *template.Template
}

func (handler *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.Once.Do(func() {
		handler.Template = template.Must(template.ParseFiles(filepath.Join("templates", handler.FileName)))
	})
	me := r.URL.Query().Get("me")
	user := database.GetUser(me)
	if user == nil {
		_ = database.InsertUser(database.User{
			Username: me,
			Chats:    make(map[string]bool),
		})
	}

	data := map[string]interface{}{
		"Host": r.Host,
		"Me":   me,
	}
	handler.Template.Execute(w, data)
}
