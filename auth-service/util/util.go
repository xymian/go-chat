package util

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

type JWTData struct {
	Token  string
	Expiry string
}

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

func GenerateJWT(username string) (*JWTData, error) {
	exp := jwt.NewNumericDate(time.Now().Add(time.Hour * 72))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      exp,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}
	data := JWTData{
		Token:  tokenString,
		Expiry: exp.String(),
	}
	return &data, nil
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
