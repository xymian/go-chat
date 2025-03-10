package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

func ParseBody(r *http.Request, o interface{}) {
	if body, err := io.ReadAll(r.Body); err == nil {
		if err = json.Unmarshal(body, o); err != nil {
			return
		}
	}
}