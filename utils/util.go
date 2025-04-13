package utils

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

type Error struct {
	Message string `json:"error"`
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
