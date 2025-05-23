package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/te6lim/go-chat/utils"
)

type contextKey string

const ContextKeyUsername contextKey = "username"

func WithJWTMiddleware(actualhanlder http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var response interface{}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			response = utils.Error{
				Message: "Unauthorized - not token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		stringToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(stringToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			response = utils.Error{
				Message: "Unauthorized - invalid token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["username"] == nil {
			response = utils.Error{
				Message: "Unauthorized - bad claims",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		username := claims["username"].(string)
		ctx := context.WithValue(r.Context(), ContextKeyUsername, username)
		actualhanlder.ServeHTTP(w, r.WithContext(ctx))
	}
}

func WithJWTMiddlewareWithData(actualhanlder func(username string) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var response interface{}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			response = utils.Error{
				Message: "Unauthorized - no token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		stringToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(stringToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			response = utils.Error{
				Message: "Unauthorized - invalid token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["username"] == nil {
			response = utils.Error{
				Message: "Unauthorized - bad claims",
			}
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		username := claims["username"].(string)
		ctx := context.WithValue(r.Context(), ContextKeyUsername, username)
		actualhanlder(username).ServeHTTP(w, r.WithContext(ctx))
	}
}
